<template>
  <div class="onlyoffice-editor">
    <!-- OnlyOffice iframe container should stay mounted/visible for DocsAPI init -->
    <div :id="editorContainerId" ref="editorContainer" class="editor-container"></div>

    <!-- Loading skeleton overlay -->
    <div v-if="loading" class="editor-overlay editor-skeleton">
      <div class="skeleton-toolbar">
        <div class="skeleton-item" style="width: 60px"></div>
        <div class="skeleton-item" style="width: 80px"></div>
        <div class="skeleton-item" style="width: 70px"></div>
        <div class="skeleton-item" style="width: 90px"></div>
      </div>
      <div class="skeleton-content">
        <div class="skeleton-line" v-for="i in 15" :key="i" :style="{ width: `${60 + Math.random() * 40}%` }"></div>
      </div>
    </div>

    <!-- Error state overlay -->
    <div v-else-if="error" class="editor-overlay editor-error">
      <el-icon :size="48" color="#f54a45"><Warning /></el-icon>
      <p>{{ error }}</p>
      <el-button type="primary" @click="initEditor">{{ $t('editor.onlyoffice.retry') }}</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { getOnlyOfficeConfig, getOnlyOfficeServerURL, type OnlyOfficeConfig } from '@/api/modules/onlyoffice'
import { ElMessage } from 'element-plus'
import { Warning } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps<{
  documentId: number
  mode?: 'edit' | 'view'
}>()

const emit = defineEmits<{
  (e: 'save', data: any): void
  (e: 'error', error: string): void
}>()

const loading = ref(true)
const error = ref('')
const editorContainer = ref<HTMLDivElement>()
const editorContainerId = `onlyoffice-editor-${Math.random().toString(36).slice(2, 10)}`
const editorConfig = ref<OnlyOfficeConfig | null>(null)
const serverUrl = ref('')

let docEditor: any = null
let readyTimeout: ReturnType<typeof setTimeout> | null = null

const ONLYOFFICE_ERROR_MAP: Record<number, string> = {
  '-1': t('editor.onlyoffice.unknownError'),
  0: t('editor.onlyoffice.uncategorized'),
  1: t('editor.onlyoffice.invalidKey'),
  2: t('editor.onlyoffice.downloadFailed'),
  3: t('editor.onlyoffice.noPermission'),
  4: t('editor.onlyoffice.internalError'),
  5: t('editor.onlyoffice.typeMismatch'),
  6: t('editor.onlyoffice.callbackFailed'),
  7: t('editor.onlyoffice.jwtFailed')
}

function extractOnlyOfficeErrorCode(event: any): number {
  const code = event?.data?.errorCode ?? event?.data?.code ?? event?.errorCode ?? event?.code
  return typeof code === 'number' ? code : -1
}

function toReadableOnlyOfficeError(event: any): string {
  const code = extractOnlyOfficeErrorCode(event)
  const message = event?.data?.message || event?.message || ''
  const mapped = ONLYOFFICE_ERROR_MAP[code] || ONLYOFFICE_ERROR_MAP[-1]
  return `OnlyOffice error (code: ${code}): ${mapped}${message ? ` - ${message}` : ''}`
}

function clearReadyTimeout() {
  if (readyTimeout) {
    clearTimeout(readyTimeout)
    readyTimeout = null
  }
}

// Load OnlyOffice API script
function loadOnlyOfficeScript(): Promise<void> {
  return new Promise((resolve, reject) => {
    if ((window as any).DocsAPI) {
      console.log('DocsAPI already loaded')
      resolve()
      return
    }

    const script = document.createElement('script')
    script.src = `${serverUrl.value}/web-apps/apps/api/documents/api.js`
    script.async = true
    // Note: Do NOT set crossOrigin as OnlyOffice nginx doesn't support CORS preflight
    script.onload = () => {
      console.log('OnlyOffice API script loaded successfully')
      if ((window as any).DocsAPI) {
        resolve()
      } else {
        reject(new Error('DocsAPI not available after script load'))
      }
    }
    script.onerror = (e) => {
      console.error('Failed to load OnlyOffice API script:', e, 'URL:', script.src)
      reject(new Error(`Failed to load OnlyOffice API from ${script.src}`))
    }
    document.head.appendChild(script)
  })
}

// Initialize the editor
async function initEditor() {
  loading.value = true
  error.value = ''
  clearReadyTimeout()

  try {
    // Get server URL first
    const serverRes: any = await getOnlyOfficeServerURL()
    console.log('[OnlyOffice] server response:', serverRes)
    serverUrl.value = serverRes.data?.serverUrl || 'http://127.0.0.1:8082'
    console.log('[OnlyOffice] using server URL:', serverUrl.value)

    // Get editor config
    const res: any = await getOnlyOfficeConfig(props.documentId, props.mode || 'edit')
    console.log('[OnlyOffice] config response:', res)
    editorConfig.value = res.data

    if (!editorConfig.value) {
      throw new Error(t('editor.onlyoffice.noConfig'))
    }

    // Load OnlyOffice API
    await loadOnlyOfficeScript()

    // Guard timeout: avoid endless loading skeleton
    readyTimeout = setTimeout(() => {
      if (loading.value && !error.value) {
        error.value = t('editor.onlyoffice.initTimeout')
        loading.value = false
        emit('error', error.value)
        ElMessage.error(error.value)
      }
    }, 20000)

    // Initialize editor (DocsAPI expects container id string)
    if (editorContainer.value && (window as any).DocsAPI) {
      console.log('[OnlyOffice] creating DocEditor with container id:', editorContainerId)
      docEditor = new (window as any).DocsAPI.DocEditor(editorContainerId, {
        ...editorConfig.value,
        width: '100%',
        height: '100%',
        events: {
          onAppReady: () => {
            console.log('[OnlyOffice] app ready')
          },
          onDocumentReady: () => {
            console.log('[OnlyOffice] document ready')
            clearReadyTimeout()
            loading.value = false
          },
          onDocumentStateChange: (event: any) => {
            console.log('[OnlyOffice] document state changed:', event)
            emit('save', event?.data)
          },
          onRequestSaveAs: (event: any) => {
            console.log('[OnlyOffice] request save as:', event)
          },
          onError: (event: any) => {
            console.error('[OnlyOffice] error:', event)
            clearReadyTimeout()
            error.value = toReadableOnlyOfficeError(event)
            loading.value = false
            emit('error', error.value)
            ElMessage.error(error.value)
          },
          onWarning: (event: any) => {
            console.warn('[OnlyOffice] warning:', event)
          },
          onInfo: (event: any) => {
            console.log('[OnlyOffice] info:', event)
          },
          onMetaChange: (event: any) => {
            console.log('[OnlyOffice] document meta changed:', event)
          }
        }
      })
    }
  } catch (err: any) {
    console.error('[OnlyOffice] failed to initialize editor:', err)
    clearReadyTimeout()
    error.value = err?.message || t('editor.onlyoffice.initFailed')
    loading.value = false
    emit('error', error.value)
    ElMessage.error(error.value)
  }
}

// Destroy editor
function destroyEditor() {
  clearReadyTimeout()
  if (docEditor) {
    try {
      docEditor.destroyEditor()
    } catch (e) {
      // Ignore destroy errors
    }
    docEditor = null
  }
}

// Watch for document ID changes
watch(() => props.documentId, () => {
  destroyEditor()
  initEditor()
})

// Watch for mode changes
watch(() => props.mode, () => {
  destroyEditor()
  initEditor()
})

onMounted(() => {
  initEditor()
})

onBeforeUnmount(() => {
  destroyEditor()
})

// Expose methods
defineExpose({
  initEditor,
  destroyEditor
})
</script>

<style scoped>
.onlyoffice-editor {
  width: 100%;
  height: 100%;
  min-height: 600px;
  position: relative;
  background: #fff;
}

.editor-container {
  width: 100%;
  height: 100%;
  min-height: 600px;
}

.editor-overlay {
  position: absolute;
  inset: 0;
  z-index: 2;
}

.editor-skeleton {
  padding: 16px;
  height: 100%;
  background: #f5f7fa;
}

.skeleton-toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  padding: 12px;
  background: #fff;
  border-radius: 4px;
}

.skeleton-item {
  height: 32px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s infinite;
  border-radius: 4px;
}

.skeleton-content {
  padding: 16px;
  background: #fff;
  border-radius: 4px;
}

.skeleton-line {
  height: 16px;
  margin-bottom: 12px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s infinite;
  border-radius: 2px;
}

@keyframes skeleton-loading {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

.editor-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  min-height: 400px;
  color: var(--kx-text-secondary);
  gap: 16px;
}

.editor-error p {
  font-size: 14px;
  margin: 0;
}
</style>
