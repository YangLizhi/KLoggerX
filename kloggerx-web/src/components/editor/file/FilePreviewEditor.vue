<template>
  <div class="file-preview-editor">
    <div v-if="loading" class="file-preview-loading">
      <el-icon class="is-loading" :size="32"><Loading /></el-icon>
      <p>{{ $t('editor.filePreview.loading') }}</p>
    </div>
    <div v-else-if="error" class="file-preview-error">
      <el-icon :size="48" color="#f54a45"><WarningFilled /></el-icon>
      <p>{{ error }}</p>
      <el-button type="primary" @click="loadPreview">{{ $t('editor.filePreview.retry') }}</el-button>
      <el-button @click="downloadFile">{{ $t('editor.filePreview.downloadOriginal') }}</el-button>
    </div>
    <div v-else class="file-preview-content">
      <div class="file-preview-frame">
        <!-- Use OnlyOffice for Word/Excel/PPT -->
        <OnlyOfficeEditor
          v-if="showOnlyOffice"
          :document-id="documentId"
          :mode="isEditMode ? 'edit' : 'view'"
          @error="handleEditorError"
        />
        <!-- Fallback for PDF: use iframe -->
        <iframe
          v-else-if="fileType === 'pdf' && previewUrl"
          :src="previewUrl"
          class="preview-iframe"
        />
        <!-- Fallback for unsupported types -->
        <div v-else class="unsupported-file">
          <el-icon :size="64" color="#c0c4cc"><Document /></el-icon>
          <p>{{ $t('editor.filePreview.unsupported') }}</p>
          <el-button type="primary" @click="downloadFile">{{ $t('editor.filePreview.downloadFile') }}</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getDocumentFilePreview } from '@/api/modules/document'
import OnlyOfficeEditor from '@/components/onlyoffice/OnlyOfficeEditor.vue'

const props = defineProps<{
  documentId: number
  content: string
}>()

const { t } = useI18n()

const loading = ref(true)
const error = ref('')
const previewUrl = ref('')
const fileName = ref('')
const fileType = ref('')
const isEditMode = ref(false)
const showOnlyOffice = ref(false)

const fileTypeIcon = computed(() => {
  const map: Record<string, string> = { pdf: 'Document', word: 'Document', excel: 'Grid', ppt: 'Monitor' }
  return map[fileType.value] || 'Document'
})

const fileTypeColor = computed(() => {
  const map: Record<string, string> = { pdf: '#f54a45', word: '#3370ff', excel: '#36b37e', ppt: '#ff7d00' }
  return map[fileType.value] || '#999'
})

const fileTypeLabel = computed(() => {
  const map: Record<string, string> = { pdf: 'PDF', word: 'Word', excel: 'Excel', ppt: 'PPT' }
  return map[fileType.value] || t('editor.filePreview.file')
})

// Check if file type can be edited with OnlyOffice
const canEdit = computed(() => {
  return ['word', 'excel', 'ppt'].includes(fileType.value)
})

/** Build a direct file-content streaming URL with token for iframe embedding */
function buildFileContentUrl(docId: number): string {
  const token = localStorage.getItem('kx_token') || ''
  const base = import.meta.env.VITE_API_BASE_URL || ''
  return `${base}/api/v1/document/${docId}/file-content?token=${encodeURIComponent(token)}`
}

function toggleEditMode() {
  isEditMode.value = !isEditMode.value
}

function downloadFile() {
  if (previewUrl.value) {
    const a = document.createElement('a')
    a.href = previewUrl.value
    a.download = fileName.value
    a.target = '_blank'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
  }
}

function handleEditorError(err: string) {
  console.error('OnlyOffice editor error:', err)
  error.value = t('editor.filePreview.editorLoadFailed', { error: err })
}

async function loadPreview() {
  loading.value = true
  error.value = ''
  showOnlyOffice.value = false

  try {
    // Parse content for file info
    if (props.content) {
      try {
        const parsed = JSON.parse(props.content)
        if (parsed.fileName) fileName.value = parsed.fileName
        if (parsed.fileType) fileType.value = parsed.fileType
      } catch { /* ignore */ }
    }

    const res: any = await getDocumentFilePreview(props.documentId)
    if (res.data?.url) {
      previewUrl.value = res.data.url
      if (res.data.fileName) fileName.value = res.data.fileName
      if (res.data.fileType) fileType.value = res.data.fileType

      // For PDF: use the backend streaming endpoint instead of the presigned URL
      // This avoids issues where MinIO/presigned URLs are not reachable from the client
      if (fileType.value === 'pdf') {
        previewUrl.value = buildFileContentUrl(props.documentId)
      }

      // Use OnlyOffice for Word/Excel/PPT files
      if (['word', 'excel', 'ppt'].includes(fileType.value)) {
        showOnlyOffice.value = true
      }
    } else {
      // If file-preview API didn't return a url but we know the file type, try streaming directly
      if (fileType.value === 'pdf') {
        previewUrl.value = buildFileContentUrl(props.documentId)
      } else {
        error.value = t('editor.filePreview.cannotGetPreview')
      }
    }
  } catch (e: any) {
    // If the preview API fails but we know it's a PDF, still try the streaming endpoint
    if (fileType.value === 'pdf') {
      previewUrl.value = buildFileContentUrl(props.documentId)
    } else {
      error.value = e.message || t('editor.filePreview.loadPreviewFailed')
    }
  } finally {
    loading.value = false
  }
}

// Watch for document ID changes
watch(() => props.documentId, () => {
  loadPreview()
})

onMounted(() => {
  loadPreview()
})

// Expose file info and functions to parent component
defineExpose({
  fileName,
  fileType,
  fileTypeIcon,
  fileTypeColor,
  fileTypeLabel,
  canEdit,
  isEditMode,
  toggleEditMode,
  downloadFile,
  loadPreview,
})
</script>

<style scoped>
.file-preview-editor {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #f7f8fa;
}
.file-preview-loading {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  color: var(--kx-text-secondary);
}
.file-preview-error {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  color: var(--kx-text-secondary);
}
.file-preview-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.file-preview-frame {
  flex: 1;
  overflow: hidden;
  position: relative;
}
.preview-iframe {
  width: 100%;
  height: 100%;
  border: none;
  background: #fff;
}
.unsupported-file {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  color: var(--kx-text-secondary);
}
</style>
