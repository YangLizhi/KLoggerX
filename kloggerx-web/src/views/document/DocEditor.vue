<template>
  <div class="doc-editor-page">
    <header class="editor-header">
      <div class="editor-header-left">
        <el-button text @click="$router.back()"><el-icon><ArrowLeft /></el-icon></el-button>
        <input v-model="doc.title.value" class="title-input" placeholder="无标题文档" @blur="doc.saveTitle" />
      </div>
      <div class="editor-header-center">
        <div class="collab-avatars" v-if="collaborators.length">
          <el-tooltip v-for="c in collaborators" :key="c.userId" :content="c.userName">
            <el-avatar :size="24" :style="{ border: `2px solid ${c.color}` }" :src="c.userAvatar">{{ c.userName[0] }}</el-avatar>
          </el-tooltip>
        </div>
      </div>
      <div class="editor-header-right">
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
        <SheetEditor :document-id="docId" :content="doc.content.value" @save="onSubEditorSave" />
      </template>
      <template v-else-if="doc.docType.value === 'slide'">
        <SlideEditor :document-id="docId" :content="doc.content.value" @save="onSubEditorSave" />
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
        <FilePreviewEditor :document-id="docId" :content="doc.content.value" />
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
})

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
