<template>
  <div class="doc-editor-page">
    <!-- Hide default header for slide and sheet types (they have their own toolbar) -->
    <header v-if="!['slide', 'sheet'].includes(doc.docType.value)" class="editor-header">
      <div class="editor-header-left">
        <el-button text @click="$router.back()"><el-icon><ArrowLeft /></el-icon></el-button>
        <!-- File type: show file icon and name -->
        <template v-if="doc.docType.value === 'file' && filePreviewRef?.fileType">
          <el-icon :size="20" :color="filePreviewRef.fileTypeColor"><component :is="filePreviewRef.fileTypeIcon" /></el-icon>
          <span class="file-name-header">{{ filePreviewRef.fileName || doc.title.value || '未命名文件' }}</span>
          <el-tag size="small" :type="getFileTagType(filePreviewRef.fileType)">{{ filePreviewRef.fileTypeLabel }}</el-tag>
        </template>
        <!-- Other document types: show title input -->
        <template v-else>
          <input v-model="doc.title.value" class="title-input" placeholder="无标题文档" @blur="doc.saveTitle" />
        </template>
      </div>
      <div class="editor-header-center">
        <div class="collab-avatars" v-if="collaborators.length">
          <el-tooltip v-for="c in collaborators" :key="c.userId" :content="c.userName">
            <el-avatar :size="24" :style="{ border: `2px solid ${c.color}` }" :src="c.userAvatar">{{ c.userName[0] }}</el-avatar>
          </el-tooltip>
        </div>
      </div>
      <div class="editor-header-right">
        <!-- File type: show download and edit mode buttons before comment button -->
        <template v-if="doc.docType.value === 'file'">
          <el-button size="small" @click="filePreviewRef?.downloadFile()"><el-icon><Download /></el-icon>下载</el-button>
          <el-button v-if="filePreviewRef?.canEdit" size="small" type="primary" @click="filePreviewRef?.toggleEditMode()">
            <el-icon><Edit /></el-icon>{{ filePreviewRef?.isEditMode ? '预览模式' : '编辑模式' }}
          </el-button>
        </template>
        <el-button size="small" @click="showComments = !showComments"><el-icon><ChatDotRound /></el-icon>评论</el-button>
        <el-button size="small" @click="showPermDialog = true"><el-icon><Lock /></el-icon>权限</el-button>
        <el-button size="small" @click="showShareDialog = true"><el-icon><Share /></el-icon>分享</el-button>
        <el-dropdown trigger="click" @command="handleMore">
          <el-button size="small"><el-icon><MoreFilled /></el-icon></el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="history">版本历史</el-dropdown-item>
              <el-dropdown-item command="export">导出</el-dropdown-item>
              <el-dropdown-item command="pin">{{ doc.isPinned.value ? '取消置顶' : '置顶' }}</el-dropdown-item>
              <el-dropdown-item command="favorite">{{ doc.isFavorite.value ? '取消收藏' : '收藏' }}</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <span class="save-status">{{ doc.saveStatus.value }}</span>
      </div>
    </header>

    <EditorToolbar v-if="editor && doc.docType.value === 'doc'" :editor="editor" />

    <div class="editor-body" v-loading="doc.loading.value">
      <template v-if="doc.docType.value === 'doc'">
        <div class="editor-content">
          <EditorContent :editor="editor" />
        </div>
      </template>
      <template v-else-if="doc.docType.value === 'sheet'">
        <SheetEditor
          :document-id="docId"
          :content="doc.content.value"
          :doc-title="doc.title.value"
          @save="onSubEditorSave"
          @update:title="doc.title.value = $event"
          @save-title="doc.saveTitle"
        />
      </template>
      <template v-else-if="doc.docType.value === 'slide'">
        <SlideEditor
          :document-id="docId"
          :content="doc.content.value"
          :doc-title="doc.title.value"
          @save="onSubEditorSave"
          @update:title="doc.title.value = $event"
          @save-title="doc.saveTitle"
        />
      </template>
      <template v-else-if="doc.docType.value === 'mindnote'">
        <MindEditor :document-id="docId" :content="doc.content.value" @save="onSubEditorSave" />
      </template>
      <template v-else-if="doc.docType.value === 'code'">
        <CodeEditor :document-id="docId" :content="doc.content.value" @save="onSubEditorSave" />
      </template>
      <template v-else-if="doc.docType.value === 'survey'">
        <SurveyEditor :document-id="docId" :content="doc.content.value" @save="onSubEditorSave" />
      </template>
      <template v-else-if="doc.docType.value === 'bitable'">
        <BitableEditor :document-id="docId" :content="doc.content.value" @save="onSubEditorSave" />
      </template>
      <template v-else-if="doc.docType.value === 'image'">
        <ImageEditor :document-id="docId" :content="doc.content.value" @save="onSubEditorSave" />
      </template>
      <template v-else-if="doc.docType.value === 'file'">
        <FilePreviewEditor ref="filePreviewRef" :document-id="docId" :content="doc.content.value" />
      </template>
    </div>

    <CommentPanel
      v-if="showComments"
      :document-id="docId"
      :editor="editor"
      @close="showComments = false"
    />

    <PermissionDialog v-if="showPermDialog" :document-id="docId" @close="showPermDialog = false" />
    <ShareDialog v-if="showShareDialog" :document-id="docId" @close="showShareDialog = false" />

    <el-drawer v-model="showHistory" title="版本历史" direction="rtl" size="400px">
      <div v-for="v in versions" :key="v.id" class="version-item" @click="handleRollback(v.version)">
        <div class="version-meta">{{ v.editorName }} - v{{ v.version }}</div>
        <div class="version-time">{{ new Date(v.createdAt).toLocaleString('zh-CN') }}</div>
      </div>
    </el-drawer>

    <el-dialog v-model="showExportDialog" title="导出文档" width="400px">
      <div class="export-options">
        <el-button @click="doExport('json')">导出为 JSON</el-button>
        <el-button @click="doExport('html')">导出为 HTML</el-button>
        <el-button @click="doExport('markdown')">导出为 Markdown</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch, defineAsyncComponent } from 'vue'
import { useRoute } from 'vue-router'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Placeholder from '@tiptap/extension-placeholder'
import Underline from '@tiptap/extension-underline'
import TextAlign from '@tiptap/extension-text-align'
import Highlight from '@tiptap/extension-highlight'
import TaskList from '@tiptap/extension-task-list'
import TaskItem from '@tiptap/extension-task-item'
import Image from '@tiptap/extension-image'
import Link from '@tiptap/extension-link'
import { Table } from '@tiptap/extension-table'
import TableRow from '@tiptap/extension-table-row'
import TableCell from '@tiptap/extension-table-cell'
import TableHeader from '@tiptap/extension-table-header'
import { getDocumentVersions, rollbackVersion } from '@/api/modules/document'
import { getDocWs, closeDocWs } from '@/api/websocket'
import { useCollaborateStore } from '@/store/modules/collaborate'
import { useDocument } from '@/hooks/useDocument'
import EditorToolbar from '@/components/editor/rich-text/EditorToolbar.vue'
import PermissionDialog from '@/components/permission/PermissionDialog.vue'
import ShareDialog from '@/components/share/ShareDialog.vue'
import '@/components/editor/rich-text/editor-styles.css'
import type { DocumentVersion, Collaborator } from '@/types'
import { ElMessage, ElMessageBox } from 'element-plus'

const CommentPanel = defineAsyncComponent(() => import('@/components/editor/rich-text/CommentPanel.vue'))
const SheetEditor = defineAsyncComponent(() => import('@/components/editor/sheet/SheetEditor.vue'))
const SlideEditor = defineAsyncComponent(() => import('@/components/editor/slide/SlideEditor.vue'))
const MindEditor = defineAsyncComponent(() => import('@/components/editor/mind/MindEditor.vue'))
const CodeEditor = defineAsyncComponent(() => import('@/components/editor/code/CodeEditor.vue'))
const SurveyEditor = defineAsyncComponent(() => import('@/components/editor/survey/SurveyEditor.vue'))
const BitableEditor = defineAsyncComponent(() => import('@/components/editor/bitable/BitableEditor.vue'))
const ImageEditor = defineAsyncComponent(() => import('@/components/editor/image/ImageEditor.vue'))
const FilePreviewEditor = defineAsyncComponent(() => import('@/components/editor/file/FilePreviewEditor.vue'))

const route = useRoute()
const collabStore = useCollaborateStore()
const docId = ref(Number(route.params.id))
const showPermDialog = ref(false)
const showShareDialog = ref(false)
const showComments = ref(false)
const showHistory = ref(false)
const showExportDialog = ref(false)
const versions = ref<DocumentVersion[]>([])
const collaborators = ref<Collaborator[]>([])

// Reference to FilePreviewEditor component for accessing file info
const filePreviewRef = ref<{
  fileName: string
  fileType: string
  fileTypeIcon: string
  fileTypeColor: string
  fileTypeLabel: string
  canEdit: boolean
  isEditMode: boolean
  toggleEditMode: () => void
  downloadFile: () => void
} | null>(null)

// Helper function to get tag type for file type
function getFileTagType(fileType: string): string {
  const map: Record<string, string> = { pdf: 'danger', word: '', excel: 'success', ppt: 'warning' }
  return map[fileType] || 'info'
}

const doc = useDocument(() => docId.value)

const editor = useEditor({
  extensions: [
    StarterKit.configure({ history: true }),
    Placeholder.configure({ placeholder: '输入 / 唤起菜单...' }),
    Underline,
    TextAlign.configure({ types: ['heading', 'paragraph'] }),
    Highlight.configure({ multicolor: true }),
    TaskList,
    TaskItem.configure({ nested: true }),
    Image,
    Link.configure({ openOnClick: false }),
    Table.configure({ resizable: true }),
    TableRow,
    TableCell,
    TableHeader,
  ],
  onUpdate: () => {
    if (!editor.value) return
    doc.scheduleSave(() => JSON.stringify(editor.value!.getJSON()))
  },
})

async function loadDocument() {
  const data = await doc.load()
  if (editor.value && data.content && data.type === 'doc') {
    try {
      editor.value.commands.setContent(JSON.parse(data.content))
    } catch {
      editor.value.commands.setContent(data.content)
    }
  }
  // Update favicon based on document type
  updateFavicon(data.type, data.fileExt)
}

// Store original favicon info for restoration
let originalFaviconHref: string | null = null
let originalFaviconType: string | null = null

// Update page favicon based on document type
function updateFavicon(docType: string, fileExt?: string) {
  // SVG icons for different file types (colored squares with symbols)
  const svgIcons: Record<string, string> = {
    pdf: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#ef4444"/><text x="16" y="22" font-size="12" font-weight="bold" fill="white" text-anchor="middle">PDF</text></svg>',
    doc: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#3b82f6"/><text x="16" y="22" font-size="12" font-weight="bold" fill="white" text-anchor="middle">DOC</text></svg>',
    docx: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#3b82f6"/><text x="16" y="22" font-size="12" font-weight="bold" fill="white" text-anchor="middle">DOC</text></svg>',
    xls: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#22c55e"/><text x="16" y="22" font-size="12" font-weight="bold" fill="white" text-anchor="middle">XLS</text></svg>',
    xlsx: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#22c55e"/><text x="16" y="22" font-size="12" font-weight="bold" fill="white" text-anchor="middle">XLS</text></svg>',
    ppt: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#f97316"/><text x="16" y="22" font-size="12" font-weight="bold" fill="white" text-anchor="middle">PPT</text></svg>',
    pptx: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#f97316"/><text x="16" y="22" font-size="12" font-weight="bold" fill="white" text-anchor="middle">PPT</text></svg>',
    png: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#8b5cf6"/><text x="16" y="22" font-size="10" font-weight="bold" fill="white" text-anchor="middle">IMG</text></svg>',
    jpg: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#8b5cf6"/><text x="16" y="22" font-size="10" font-weight="bold" fill="white" text-anchor="middle">IMG</text></svg>',
    jpeg: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#8b5cf6"/><text x="16" y="22" font-size="10" font-weight="bold" fill="white" text-anchor="middle">IMG</text></svg>',
    gif: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#8b5cf6"/><text x="16" y="22" font-size="10" font-weight="bold" fill="white" text-anchor="middle">GIF</text></svg>',
    svg: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#8b5cf6"/><text x="16" y="22" font-size="10" font-weight="bold" fill="white" text-anchor="middle">SVG</text></svg>',
    mp3: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#ec4899"/><text x="16" y="22" font-size="10" font-weight="bold" fill="white" text-anchor="middle">MP3</text></svg>',
    wav: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#ec4899"/><text x="16" y="22" font-size="10" font-weight="bold" fill="white" text-anchor="middle">WAV</text></svg>',
    mp4: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#eab308"/><text x="16" y="22" font-size="10" font-weight="bold" fill="white" text-anchor="middle">MP4</text></svg>',
    mkv: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#eab308"/><text x="16" y="22" font-size="10" font-weight="bold" fill="white" text-anchor="middle">MKV</text></svg>',
    zip: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#64748b"/><text x="16" y="22" font-size="10" font-weight="bold" fill="white" text-anchor="middle">ZIP</text></svg>',
    rar: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#64748b"/><text x="16" y="22" font-size="10" font-weight="bold" fill="white" text-anchor="middle">RAR</text></svg>',
    md: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#6366f1"/><text x="16" y="22" font-size="10" font-weight="bold" fill="white" text-anchor="middle">MD</text></svg>',
    json: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#f59e0b"/><text x="16" y="22" font-size="10" font-weight="bold" fill="white" text-anchor="middle">JSON</text></svg>',
    html: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#14b8a6"/><text x="16" y="22" font-size="10" font-weight="bold" fill="white" text-anchor="middle">HTML</text></svg>',
    py: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#3b82f6"/><text x="16" y="22" font-size="10" font-weight="bold" fill="white" text-anchor="middle">PY</text></svg>',
    go: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#06b6d4"/><text x="16" y="22" font-size="12" font-weight="bold" fill="white" text-anchor="middle">GO</text></svg>',
    js: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#eab308"/><text x="16" y="22" font-size="10" font-weight="bold" fill="white" text-anchor="middle">JS</text></svg>',
    ts: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#3b82f6"/><text x="16" y="22" font-size="10" font-weight="bold" fill="white" text-anchor="middle">TS</text></svg>',
  }

  // Default icons by document type
  const defaultIcons: Record<string, string> = {
    doc: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#3b82f6"/><text x="16" y="22" font-size="10" font-weight="bold" fill="white" text-anchor="middle">DOC</text></svg>',
    sheet: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#22c55e"><rect x="6" y="6" width="20" height="20" fill="none" stroke="white" stroke-width="2"/><line x1="6" y1="12" x2="26" y2="12" stroke="white" stroke-width="1"/><line x1="6" y1="18" x2="26" y2="18" stroke="white" stroke-width="1"/><line x1="13" y1="6" x2="13" y2="26" stroke="white" stroke-width="1"/><line x1="20" y1="6" x2="20" y2="26" stroke="white" stroke-width="1"/></rect></svg>',
    slide: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#f97316"/><rect x="6" y="8" width="20" height="16" rx="2" fill="white"/></svg>',
    mindnote: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#8b5cf6"/><circle cx="16" cy="16" r="4" fill="white"/><circle cx="8" cy="10" r="2" fill="white"/><circle cx="24" cy="10" r="2" fill="white"/><circle cx="8" cy="22" r="2" fill="white"/><circle cx="24" cy="22" r="2" fill="white"/><line x1="12" y1="14" x2="10" y2="11" stroke="white" stroke-width="1"/><line x1="20" y1="14" x2="22" y2="11" stroke="white" stroke-width="1"/><line x1="12" y1="18" x2="10" y2="21" stroke="white" stroke-width="1"/><line x1="20" y1="18" x2="22" y2="21" stroke="white" stroke-width="1"/></rect></svg>',
    bitable: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#06b6d4"/><rect x="5" y="5" width="22" height="22" fill="none" stroke="white" stroke-width="2"/><line x1="5" y1="11" x2="27" y2="11" stroke="white" stroke-width="1"/><line x1="5" y1="17" x2="27" y2="17" stroke="white" stroke-width="1"/><line x1="12" y1="5" x2="12" y2="27" stroke="white" stroke-width="1"/><line x1="20" y1="5" x2="20" y2="27" stroke="white" stroke-width="1"/></svg>',
    survey: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#ec4899"/><rect x="6" y="6" width="20" height="20" fill="none" stroke="white" stroke-width="2"/><line x1="10" y1="12" x2="22" y2="12" stroke="white" stroke-width="2"/><line x1="10" y1="16" x2="22" y2="16" stroke="white" stroke-width="2"/><line x1="10" y1="20" x2="18" y2="20" stroke="white" stroke-width="2"/></svg>',
    image: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#8b5cf6"/><rect x="5" y="7" width="22" height="18" rx="2" fill="none" stroke="white" stroke-width="2"/><circle cx="11" cy="13" r="2" fill="white"/><path d="M7 23 L13 17 L17 21 L22 14 L25 23 Z" fill="white"/></svg>',
    code: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#6366f1"/><text x="16" y="22" font-size="14" font-weight="bold" fill="white" text-anchor="middle" font-family="monospace">&lt;/&gt;</text></svg>',
    file: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#64748b"/><path d="M10 6 L18 6 L22 10 L22 26 L10 26 Z" fill="white"/><path d="M18 6 L18 10 L22 10" fill="none" stroke="#64748b" stroke-width="1"/></svg>',
    folder: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="4" fill="#f59e0b"/><path d="M4 10 L4 24 Q4 26 6 26 L26 26 Q28 26 28 24 L28 12 Q28 10 26 10 L16 10 L14 8 L6 8 Q4 8 4 10 Z" fill="white"/></svg>',
  }

  let svgContent: string

  // Determine which icon to use
  if (fileExt) {
    const ext = fileExt.replace('.', '').toLowerCase()
    if (svgIcons[ext]) {
      svgContent = svgIcons[ext]
    } else {
      svgContent = defaultIcons[docType] || defaultIcons['file']
    }
  } else {
    svgContent = defaultIcons[docType] || defaultIcons['file']
  }

  // Create data URL from SVG
  const faviconUrl = 'data:image/svg+xml,' + encodeURIComponent(svgContent)

  // Remove all existing icon link elements to prevent duplicates
  const existingLinks = document.querySelectorAll("link[rel*='icon']")
  if (existingLinks.length > 0) {
    // Store original favicon info from the first link
    if (!originalFaviconHref) {
      originalFaviconHref = existingLinks[0].getAttribute('href')
      originalFaviconType = existingLinks[0].getAttribute('type')
    }
    // Remove all existing icon links
    existingLinks.forEach(link => link.remove())
  }

  // Create new favicon link element
  const newLink = document.createElement('link')
  newLink.rel = 'icon'
  newLink.type = 'image/svg+xml'
  newLink.href = faviconUrl
  document.head.appendChild(newLink)

  // Update page title with document name
  const title = doc.title.value || '未命名文档'
  document.title = title
}

function initWebSocket() {
  const ws = getDocWs(docId.value)
  ws.on('collaborators', (data: Collaborator[]) => {
    collaborators.value = data
    collabStore.setCollaborators(data)
  })
  ws.on('user_join', (data: Collaborator) => {
    collabStore.addCollaborator(data)
    collaborators.value = [...collabStore.collaborators]
  })
  ws.on('user_leave', (data: { userId: number }) => {
    collabStore.removeCollaborator(data.userId)
    collaborators.value = [...collabStore.collaborators]
  })
}

async function handleMore(cmd: string) {
  if (cmd === 'history') {
    const res: any = await getDocumentVersions(docId.value)
    versions.value = res.data || []
    showHistory.value = true
  } else if (cmd === 'export') {
    showExportDialog.value = true
  } else if (cmd === 'pin') {
    await doc.togglePin()
  } else if (cmd === 'favorite') {
    await doc.toggleFavorite()
  }
}

function doExport(format: string) {
  if (!editor.value) return
  let content: string
  let filename: string
  let mimeType: string

  if (format === 'json') {
    content = JSON.stringify(editor.value.getJSON(), null, 2)
    filename = `${doc.title.value || 'document'}.json`
    mimeType = 'application/json'
  } else if (format === 'html') {
    content = `<!DOCTYPE html><html><head><meta charset="utf-8"><title>${doc.title.value}</title></head><body>${editor.value.getHTML()}</body></html>`
    filename = `${doc.title.value || 'document'}.html`
    mimeType = 'text/html'
  } else {
    // markdown - simple html to markdown conversion
    const html = editor.value.getHTML()
    content = htmlToMarkdown(html)
    filename = `${doc.title.value || 'document'}.md`
    mimeType = 'text/markdown'
  }

  const blob = new Blob([content], { type: mimeType })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
  showExportDialog.value = false
  ElMessage.success('导出成功')
}

function htmlToMarkdown(html: string): string {
  let md = html
  md = md.replace(/<h1[^>]*>(.*?)<\/h1>/gi, '# $1\n\n')
  md = md.replace(/<h2[^>]*>(.*?)<\/h2>/gi, '## $1\n\n')
  md = md.replace(/<h3[^>]*>(.*?)<\/h3>/gi, '### $1\n\n')
  md = md.replace(/<strong>(.*?)<\/strong>/gi, '**$1**')
  md = md.replace(/<em>(.*?)<\/em>/gi, '*$1*')
  md = md.replace(/<u>(.*?)<\/u>/gi, '$1')
  md = md.replace(/<s>(.*?)<\/s>/gi, '~~$1~~')
  md = md.replace(/<a[^>]*href="([^"]*)"[^>]*>(.*?)<\/a>/gi, '[$2]($1)')
  md = md.replace(/<img[^>]*src="([^"]*)"[^>]*\/?>/gi, '![]($1)')
  md = md.replace(/<code>(.*?)<\/code>/gi, '`$1`')
  md = md.replace(/<blockquote[^>]*>(.*?)<\/blockquote>/gi, '> $1\n\n')
  md = md.replace(/<li[^>]*>(.*?)<\/li>/gi, '- $1\n')
  md = md.replace(/<hr\s*\/?>/gi, '---\n\n')
  md = md.replace(/<p[^>]*>(.*?)<\/p>/gi, '$1\n\n')
  md = md.replace(/<br\s*\/?>/gi, '\n')
  md = md.replace(/<[^>]+>/g, '')
  md = md.replace(/\n{3,}/g, '\n\n')
  return md.trim()
}

async function handleRollback(version: number) {
  await ElMessageBox.confirm(`确定回滚到v${version}？`, '版本回滚')
  await rollbackVersion(docId.value, version)
  ElMessage.success('已回滚')
  showHistory.value = false
  loadDocument()
}

function onSubEditorSave(content: string) {
  doc.saveContent(content)
}

onMounted(() => {
  loadDocument()
  initWebSocket()
})

onBeforeUnmount(() => {
  doc.cleanup()
  if (editor.value && doc.docType.value === 'doc') {
    doc.saveContent(JSON.stringify(editor.value.getJSON()))
  }
  closeDocWs()
  // Restore default favicon
  restoreFavicon()
})

// Restore the default favicon
function restoreFavicon() {
  // Remove all existing icon links
  const existingLinks = document.querySelectorAll("link[rel*='icon']")
  existingLinks.forEach(link => link.remove())

  // Restore original favicon
  if (originalFaviconHref) {
    const link = document.createElement('link')
    link.rel = 'icon'
    if (originalFaviconType) {
      link.type = originalFaviconType
    }
    link.href = originalFaviconHref
    document.head.appendChild(link)
  } else {
    // Fallback to default favicon
    const link = document.createElement('link')
    link.rel = 'icon'
    link.type = 'image/svg+xml'
    link.href = '/favicon.svg'
    document.head.appendChild(link)
  }

  document.title = 'iKloggerX'
}

watch(() => route.params.id, (newId) => {
  if (newId) {
    docId.value = Number(newId)
    loadDocument()
    closeDocWs()
    initWebSocket()
  }
})
</script>

<style scoped>
.doc-editor-page {
  position: fixed;
  inset: 0;
  display: flex;
  flex-direction: column;
  background: #fff;
  z-index: 100;
}
.editor-header {
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  border-bottom: 1px solid var(--kx-border);
  flex-shrink: 0;
}
.editor-header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}
.title-input {
  border: none;
  outline: none;
  font-size: 16px;
  font-weight: 600;
  color: var(--kx-text-primary);
  background: transparent;
  width: 300px;
}
.title-input::placeholder {
  color: var(--kx-text-placeholder);
}
.file-name-header {
  font-size: 16px;
  font-weight: 600;
  color: var(--kx-text-primary);
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.editor-header-center {
  flex: 1;
  display: flex;
  justify-content: center;
}
.collab-avatars {
  display: flex;
}
.collab-avatars .el-avatar {
  margin-left: -4px;
}
.editor-header-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  justify-content: flex-end;
}
.save-status {
  font-size: 12px;
  color: var(--kx-text-placeholder);
  margin-left: 8px;
}
.editor-body {
  flex: 1;
  overflow-y: auto;
  display: flex;
  justify-content: center;
}
.editor-content {
  width: 100%;
  max-width: 800px;
  padding: 40px 24px;
  min-height: calc(100vh - 96px);
}
.version-item {
  padding: 12px 0;
  border-bottom: 1px solid var(--kx-border);
  cursor: pointer;
}
.version-item:hover {
  background: var(--kx-sidebar-bg);
}
.version-meta {
  font-weight: 500;
}
.version-time {
  font-size: 12px;
  color: var(--kx-text-placeholder);
  margin-top: 4px;
}
.export-options {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.export-options .el-button {
  width: 100%;
}
</style>
