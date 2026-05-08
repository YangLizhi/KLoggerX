<template>
  <div class="doc-editor-page">
    <!-- Hide default header for slide and sheet types (they have their own toolbar) -->
    <header v-if="!['slide', 'sheet'].includes(doc.docType.value)" class="editor-header">
      <div class="editor-header-left">
        <el-button text @click="$router.back()"><el-icon><ArrowLeft /></el-icon></el-button>
        <!-- File type: show file icon and name -->
        <template v-if="doc.docType.value === 'file' && filePreviewRef?.fileType">
          <el-icon :size="20" :color="filePreviewRef.fileTypeColor"><component :is="filePreviewRef.fileTypeIcon" /></el-icon>
          <span class="file-name-header">{{ filePreviewRef.fileName || doc.title.value || $t('editor.untitledFile') }}</span>
          <el-tag size="small" :type="getFileTagType(filePreviewRef.fileType)">{{ filePreviewRef.fileTypeLabel }}</el-tag>
        </template>
        <!-- Other document types: show title input -->
        <template v-else>
          <input v-model="doc.title.value" class="title-input" :placeholder="$t('editor.untitledDoc')" @blur="doc.saveTitle" />
        </template>
      </div>
      <div class="editor-header-center">
        <div class="collab-container">
          <transition-group name="collab-fade" tag="div" class="collab-avatars" v-if="displayCollaborators.length">
            <el-tooltip
              v-for="c in displayCollaborators"
              :key="c.userId"
              :content="c.userName"
              placement="bottom"
            >
              <div class="collab-avatar-wrapper">
                <el-avatar
                  :size="32"
                  :style="{ borderColor: c.color }"
                  :src="c.userAvatar || undefined"
                  class="collab-avatar"
                >
                  {{ c.userName?.[0]?.toUpperCase() || '?' }}
                </el-avatar>
              </div>
            </el-tooltip>
          </transition-group>
          <el-badge :value="collaborators.length" :hidden="collaborators.length <= maxDisplayAvatars" class="collab-badge" v-if="collaborators.length">
            <span class="collab-count">{{ $t('editor.onlineCount', { count: collaborators.length }) }}</span>
          </el-badge>
          <!-- Yjs 连接状态指示器（仅当 Yjs 启用时显示） -->
          <template v-if="enableYjs">
            <span v-if="yjsConnectionStatus === 'connecting'" class="yjs-status connecting">
              <el-icon class="is-loading"><Loading /></el-icon> {{ $t('editor.connecting') }}
            </span>
            <span v-else-if="yjsConnectionStatus === 'connected'" class="yjs-status connected">
              <span class="status-dot green"></span> {{ $t('editor.connected') }}
            </span>
            <span v-else class="yjs-status disconnected">
              <span class="status-dot red"></span> {{ $t('editor.disconnected') }}
            </span>
          </template>
        </div>
      </div>
      <div class="editor-header-right">
        <!-- File type: show download and edit mode buttons before comment button -->
        <template v-if="doc.docType.value === 'file'">
          <el-button size="small" @click="filePreviewRef?.downloadFile()"><el-icon><Download /></el-icon>{{ $t('editor.download') }}</el-button>
          <el-button v-if="filePreviewRef?.canEdit" size="small" type="primary" @click="filePreviewRef?.toggleEditMode()">
            <el-icon><Edit /></el-icon>{{ filePreviewRef?.isEditMode ? $t('editor.previewMode') : $t('editor.editMode') }}
          </el-button>
        </template>
        <el-button size="small" @click="showComments = !showComments"><el-icon><ChatDotRound /></el-icon>{{ $t('editor.comment') }}</el-button>
        <el-button size="small" @click="showPermDialog = true"><el-icon><Lock /></el-icon>{{ $t('editor.permission') }}</el-button>
        <el-button size="small" @click="showShareDialog = true"><el-icon><Share /></el-icon>{{ $t('editor.share') }}</el-button>
        <el-dropdown trigger="click" @command="handleMore">
          <el-button size="small"><el-icon><MoreFilled /></el-icon></el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="history">{{ $t('editor.versionHistory') }}</el-dropdown-item>
              <el-dropdown-item command="export">{{ $t('editor.export') }}</el-dropdown-item>
              <el-dropdown-item command="pin">{{ doc.isPinned.value ? $t('editor.unpin') : $t('editor.pin') }}</el-dropdown-item>
              <el-dropdown-item command="favorite">{{ doc.isFavorite.value ? $t('editor.unfavorite') : $t('editor.favorite') }}</el-dropdown-item>
              <el-dropdown-item command="saveAsTemplate">{{ $t('editor.saveAsTemplate') }}</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <span class="save-status">{{ doc.saveStatus.value }}</span>
      </div>
    </header>

    <EditorToolbar v-if="editor && doc.docType.value === 'doc'" :editor="editor as any" />

    <!-- 协作状态条（仅当 Yjs 启用时显示） -->
    <transition name="slide-down">
      <div v-if="enableYjs && showCollabStatus" class="collab-status-bar" :class="collabStatusClass">
        <!-- 断线状态 -->
        <template v-if="yjsConnectionStatus === 'disconnected'">
          <el-icon><WarningFilled /></el-icon>
          <span>{{ $t('document.collab.disconnectedRetrying') }}</span>
          <el-button size="small" @click="yjsCollaboration?.retryConnection()">{{ $t('document.collab.manualReconnect') }}</el-button>
        </template>

        <!-- 正在连接 -->
        <template v-else-if="yjsConnectionStatus === 'connecting'">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>{{ $t('document.collab.connectingService') }}</span>
        </template>

        <!-- 冲突检测 -->
        <template v-else-if="yjsCollaboration?.conflictDetected.value">
          <el-icon><WarningFilled /></el-icon>
          <span>{{ $t('document.collab.conflictMerged') }}</span>
          <el-button size="small" type="primary" @click="handleDismissConflict">{{ $t('document.collab.gotIt') }}</el-button>
        </template>

        <!-- 同步完成（短暂显示后自动隐藏） -->
        <template v-else-if="justSynced">
          <el-icon><CircleCheckFilled /></el-icon>
          <span>{{ $t('document.collab.allSynced') }}</span>
        </template>
      </div>
    </transition>

    <div class="editor-body">
      <!-- Editor Skeleton Loading -->
      <div v-if="doc.loading.value && !loadError" class="editor-skeleton">
        <el-skeleton animated :loading="true">
          <template #template>
            <div class="editor-skeleton-content">
              <el-skeleton-item variant="h1" style="width: 50%; height: 28px; margin-bottom: 24px" />
              <el-skeleton-item variant="text" style="width: 100%; height: 16px; margin-bottom: 12px" />
              <el-skeleton-item variant="text" style="width: 92%; height: 16px; margin-bottom: 12px" />
              <el-skeleton-item variant="text" style="width: 85%; height: 16px; margin-bottom: 24px" />
              <el-skeleton-item variant="text" style="width: 35%; height: 22px; margin-bottom: 16px" />
              <el-skeleton-item variant="text" style="width: 100%; height: 16px; margin-bottom: 12px" />
              <el-skeleton-item variant="text" style="width: 78%; height: 16px; margin-bottom: 12px" />
              <el-skeleton-item variant="text" style="width: 95%; height: 16px; margin-bottom: 12px" />
              <el-skeleton-item variant="text" style="width: 60%; height: 16px; margin-bottom: 24px" />
              <el-skeleton-item variant="text" style="width: 100%; height: 16px; margin-bottom: 12px" />
              <el-skeleton-item variant="text" style="width: 88%; height: 16px; margin-bottom: 12px" />
            </div>
          </template>
        </el-skeleton>
      </div>

      <!-- Error State -->
      <div v-else-if="loadError" class="error-state">
        <el-icon :size="48" color="#f56c6c"><WarningFilled /></el-icon>
        <p class="error-message">{{ loadError }}</p>
        <el-button type="primary" @click="retry">{{ $t('editor.reload') }}</el-button>
      </div>
      <template v-else-if="doc.docType.value === 'doc'">
        <div class="editor-content-wrapper">
          <div ref="editorContentRef" class="editor-content" :class="{ 'has-comment-bubbles': showComments }">
            <EditorContent :editor="(editor as any)" />
            <CommentBubble
              v-if="editor"
              :editor="(editor as any)"
              :document-id="docId"
              :editor-container="editorContentRef"
              @click-comment="handleCommentBubbleClick"
            />
          </div>
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

    <!-- 协作通知提示 -->
    <div class="collaborate-notices">
      <TransitionGroup name="notice-fade">
        <div v-for="notice in collaborateNotices" :key="notice.id" class="collaborate-notice">
          <span class="notice-avatar" :style="{ borderColor: notice.color }">
            {{ notice.username?.charAt(0)?.toUpperCase() || '?' }}
          </span>
          <span>{{ notice.username }} {{ notice.action === 'join' ? $t('editor.joinedEditing') : $t('editor.leftEditing') }}</span>
        </div>
      </TransitionGroup>
    </div>

    <CommentPanel
      v-if="showComments"
      :document-id="docId"
      :editor="(editor as any)"
      :active-comment-id="activeCommentId"
      :owner-id="documentOwnerId"
      @close="showComments = false"
    />

    <PermissionDialog v-if="showPermDialog" :document-id="docId" @close="showPermDialog = false" />
    <ShareDialog v-if="showShareDialog" :document-id="docId" @close="showShareDialog = false" />

    <el-drawer v-model="showHistory" :title="$t('editor.versionHistory')" direction="rtl" size="400px">
      <div v-for="v in versions" :key="v.id" class="version-item" @click="handleRollback(v.version)">
        <div class="version-meta">{{ v.editorName }} - v{{ v.version }}</div>
        <div class="version-time">{{ new Date(v.createdAt).toLocaleString('zh-CN') }}</div>
      </div>
    </el-drawer>

    <el-dialog v-model="showExportDialog" :title="$t('editor.exportDocument')" width="400px">
      <div class="export-options">
        <el-button @click="doExport('json')">{{ $t('editor.exportJSON') }}</el-button>
        <el-button @click="doExport('html')">{{ $t('editor.exportHTML') }}</el-button>
        <el-button @click="doExport('markdown')">{{ $t('editor.exportMarkdown') }}</el-button>
      </div>
    </el-dialog>

    <!-- 冲突解决弹窗 -->
    <el-dialog v-model="showConflictDialog" :title="$t('document.collab.conflictTitle')" :close-on-click-modal="false">
      <p>{{ $t('document.collab.conflictDesc') }}</p>
      <p>{{ $t('document.collab.conflictChoose') }}</p>
      <template #footer>
        <el-button @click="keepMine">{{ $t('document.collab.keepMine') }}</el-button>
        <el-button @click="keepTheirs">{{ $t('document.collab.keepTheirs') }}</el-button>
        <el-button type="primary" @click="reloadDocument">{{ $t('document.collab.refreshPage') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showSaveAsTemplateDialog" :title="$t('editor.saveAsTemplate')" width="480px">
      <el-form :model="templateForm" label-width="80px">
        <el-form-item :label="$t('editor.templateName')" required>
          <el-input v-model="templateForm.name" :placeholder="$t('editor.templateNamePlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('editor.templateDesc')">
          <el-input v-model="templateForm.description" type="textarea" :rows="3" :placeholder="$t('editor.templateDescPlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('editor.templateCategory')">
          <el-input v-model="templateForm.category" :placeholder="$t('editor.templateCategoryPlaceholder')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showSaveAsTemplateDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="templateSaving" @click="handleSaveAsTemplate">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch, defineAsyncComponent, computed } from 'vue'
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
import Collaboration from '@tiptap/extension-collaboration'
import CollaborationCursor from '@tiptap/extension-collaboration-cursor'
import { CommentMark } from '@/components/editor/rich-text/extensions/CommentMark'
import CommentBubble from '@/components/editor/rich-text/CommentBubble.vue'
import { getDocumentVersions, rollbackVersion, saveDocumentAsTemplate } from '@/api/modules/document'
import { getDocWs, closeDocWs } from '@/api/websocket'
import { useCollaborateStore } from '@/store/modules/collaborate'
import { useUserStore } from '@/store/modules/user'
import { useDocument } from '@/hooks/useDocument'
import { useYjsCollaboration, type ConnectionStatus, type SyncStatus } from '@/composables/useYjsCollaboration'
import EditorToolbar from '@/components/editor/rich-text/EditorToolbar.vue'
import PermissionDialog from '@/components/permission/PermissionDialog.vue'
import ShareDialog from '@/components/share/ShareDialog.vue'
import '@/components/editor/rich-text/editor-styles.css'
import type { DocumentVersion, Collaborator } from '@/types'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { WarningFilled, Loading, CircleCheckFilled } from '@element-plus/icons-vue'

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
const { t } = useI18n()
const collabStore = useCollaborateStore()
const userStore = useUserStore()
const docId = ref(Number(route.params.id))

// 环境变量控制 Yjs 协作功能（默认禁用）
const enableYjs = import.meta.env.VITE_ENABLE_YJS === 'true'
const showPermDialog = ref(false)
const showShareDialog = ref(false)
const showComments = ref(false)
const activeCommentId = ref<string | null>(null)
const editorContentRef = ref<HTMLElement | null>(null)
const showHistory = ref(false)
const showExportDialog = ref(false)
const showSaveAsTemplateDialog = ref(false)
const templateForm = ref({
  name: '',
  description: '',
  category: ''
})
const templateSaving = ref(false)
const versions = ref<DocumentVersion[]>([])
const collaborators = ref<Collaborator[]>([])
const maxDisplayAvatars = 5
const documentOwnerId = ref<number | undefined>(undefined)
const loadError = ref<string | null>(null)

// 协作通知提示
interface CollaborateNotice {
  id: number
  username: string
  avatar: string
  color: string
  action: 'join' | 'leave'
  timestamp: number
}
const collaborateNotices = ref<CollaborateNotice[]>([])
let noticeIdCounter = 0

// 添加协作通知
function addCollaborateNotice(username: string, avatar: string, color: string, action: 'join' | 'leave') {
  // 最多同时显示3条，多余的排队（通过移除旧的实现）
  if (collaborateNotices.value.length >= 3) {
    collaborateNotices.value.shift()
  }
  
  const notice: CollaborateNotice = {
    id: ++noticeIdCounter,
    username,
    avatar,
    color,
    action,
    timestamp: Date.now()
  }
  
  collaborateNotices.value.push(notice)
  
  // 3秒后自动移除
  setTimeout(() => {
    const index = collaborateNotices.value.findIndex(n => n.id === notice.id)
    if (index > -1) {
      collaborateNotices.value.splice(index, 1)
    }
  }, 3000)
}

// Yjs 协作
const currentUserColor = ref('#3b82f6')
const yjsConnectionStatus = ref<ConnectionStatus>(enableYjs ? 'connecting' : 'disconnected')

// 获取当前用户信息用于 Yjs
function getCurrentUserName(): string {
  const userStr = localStorage.getItem('kx_user')
  if (userStr) {
    try { return JSON.parse(userStr).username || 'Anonymous' } catch { return 'Anonymous' }
  }
  return 'Anonymous'
}

function getCurrentUserAvatar(): string {
  const userStr = localStorage.getItem('kx_user')
  if (userStr) {
    try { return JSON.parse(userStr).avatar || '' } catch { return '' }
  }
  return ''
}

// 初始化 Yjs 协作（仅当 enableYjs 为 true 时）
const yjsCollaboration = enableYjs
  ? useYjsCollaboration({
      documentId: docId.value,
      user: {
        userId: getCurrentUserId(),
        userName: getCurrentUserName(),
        userAvatar: getCurrentUserAvatar(),
        color: currentUserColor.value,
      },
      onCollaboratorsUpdate: (collabs) => {
        // 合并 Yjs awareness 协作者到列表
        collabStore.setCollaborators(collabs)
      },
      onConcurrentEdit: (info) => {
        ElMessage.warning({
          message: t('editor.editingNearby', { user: info.userName }),
          duration: 3000,
        })
      },
      onSync: (isSynced) => {
        if (isSynced) {
          console.log('[DocEditor] Yjs initial sync complete')
        }
      },
    })
  : null

// 协作状态条相关
const showConflictDialog = ref(false)
const justSynced = ref(false)
let justSyncedTimer: ReturnType<typeof setTimeout> | null = null

// 是否显示协作状态条
const showCollabStatus = computed(() => {
  if (!enableYjs) return false
  if (yjsConnectionStatus.value === 'disconnected') return true
  if (yjsConnectionStatus.value === 'connecting') return true
  if (yjsCollaboration?.conflictDetected.value) return true
  if (justSynced.value) return true
  return false
})

// 协作状态条样式类
const collabStatusClass = computed(() => {
  if (yjsConnectionStatus.value === 'disconnected') return 'error'
  if (yjsConnectionStatus.value === 'connecting') return 'info'
  if (yjsCollaboration?.conflictDetected.value) return 'warning'
  if (justSynced.value) return 'success'
  return ''
})

// 消除冲突提示
function handleDismissConflict() {
  yjsCollaboration?.dismissConflict()
}

// 冲突弹窗操作
function keepMine() {
  // Yjs CRDT 中本地版本已自动保留，只需关闭弹窗
  showConflictDialog.value = false
  yjsCollaboration?.dismissConflict()
}

function keepTheirs() {
  // 重新从服务端同步
  showConflictDialog.value = false
  yjsCollaboration?.dismissConflict()
  loadDocument()
}

function reloadDocument() {
  showConflictDialog.value = false
  window.location.reload()
}

// 监听连接状态（仅当 Yjs 启用时）
if (enableYjs && yjsCollaboration) {
  watch(() => yjsCollaboration.connectionStatus.value, (status) => {
    yjsConnectionStatus.value = status
  })

  // 监听同步状态，短暂显示同步完成提示
  watch(() => yjsCollaboration.syncStatus.value, (status: SyncStatus) => {
    if (status === 'synced' && !yjsCollaboration.conflictDetected.value) {
      justSynced.value = true
      if (justSyncedTimer) clearTimeout(justSyncedTimer)
      justSyncedTimer = setTimeout(() => {
        justSynced.value = false
      }, 3000)
    }
  })
}

// Computed property to limit displayed collaborators
const displayCollaborators = computed(() => {
  return collaborators.value.slice(0, maxDisplayAvatars)
})

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

// 获取当前用户名（Yjs 使用上方定义的 getCurrentUserName）
// 获取当前用户头像（Yjs 使用上方定义的 getCurrentUserAvatar）
function getCurrentUserId(): number {
  const userStr = localStorage.getItem('kx_user')
  if (userStr) {
    try {
      const user = JSON.parse(userStr)
      return user.id || 0
    } catch {
      return 0
    }
  }
  return 0
}

// 获取当前用户头像（Yjs 已禁用，保留以备后续启用）
// function getCurrentUserAvatar(): string { ... }

// 编辑器扩展配置（根据 enableYjs 条件加载协作扩展）
const createEditorExtensions = () => {
  const extensions: any[] = [
    StarterKit.configure({
      // Collaboration 自带 history 管理，禁用 StarterKit 的 undoRedo；当 Yjs 禁用时恢复默认
      ...(enableYjs ? { undoRedo: false } : {}),
    }),
    Placeholder.configure({ placeholder: t('editor.placeholderSlash') }),
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
    CommentMark,
  ]

  // 仅当 Yjs 启用时加载协作扩展
  if (enableYjs && yjsCollaboration) {
    extensions.push(
      Collaboration.configure({
        document: yjsCollaboration.ydoc,
      }),
      CollaborationCursor.configure({
        provider: yjsCollaboration.provider,
      }),
    )
  }

  return extensions
}

// 处理编辑器点击事件（用于评论标记点击）
const handleEditorClick = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  const commentMark = target.closest('.comment-mark') as HTMLElement
  if (commentMark) {
    const commentId = commentMark.getAttribute('data-comment-id')
    if (commentId) {
      // 打开评论面板并激活对应评论
      showComments.value = true
      activeCommentId.value = commentId
      // 重置 activeCommentId 以便下次点击可以再次触发
      setTimeout(() => {
        activeCommentId.value = null
      }, 100)
    }
  }
}

// 处理评论气泡点击事件
const handleCommentBubbleClick = (commentId: string) => {
  // 打开评论面板
  showComments.value = true
  // 设置激活的评论ID
  activeCommentId.value = commentId
  // 重置 activeCommentId 以便下次点击可以再次触发
  setTimeout(() => {
    activeCommentId.value = null
  }, 100)
}

// 编辑器实例 - 在 script setup 顶层同步创建，确保 onMounted 钩子正确注册
const editor = useEditor({
  extensions: createEditorExtensions(),
  onUpdate: () => {
    if (!editor.value) return
    if (doc.docType.value !== 'doc') return
    doc.scheduleSave(() => JSON.stringify(editor.value!.getJSON()))
  },
  editorProps: {
    handleClick: (_view: any, _pos: any, event: any) => {
      handleEditorClick(event as MouseEvent)
      return false
    },
  },
})

async function loadDocument() {
  loadError.value = null
  try {
    const data = await doc.load()
    documentOwnerId.value = data.ownerId
    
    updateFavicon(data.type, data.fileExt)
    
    // doc 类型：编辑器已在 setup 顶层通过 useEditor 创建，这里只需设置内容
    if (data.type === 'doc' && editor.value && data.content) {
      try {
        editor.value.commands.setContent(JSON.parse(data.content))
      } catch {
        editor.value.commands.setContent(data.content)
      }
    }
  } catch (e: any) {
    loadError.value = e?.response?.data?.message || e?.message || t('editor.loadFailed')
    console.error('[DocEditor] loadDocument failed:', e)
  }
}

function retry() {
  loadDocument()
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
  const title = doc.title.value || t('editor.untitledDoc')
  document.title = title
}

function initWebSocket() {
  const ws = getDocWs(docId.value)
  ws.on('collaborators_update', (data: Collaborator[]) => {
    collaborators.value = data
    collabStore.setCollaborators(data)
    
    const currentUserId = getCurrentUserId()
    const currentUser = data.find(c => c.userId === currentUserId)
    if (currentUser && currentUser.color) {
      currentUserColor.value = currentUser.color
    }
    
  })
  ws.on('user_join', (data: Collaborator) => {
    // Add new collaborator if not already in list
    const exists = collaborators.value.some(c => c.userId === data.userId)
    if (!exists) {
      collaborators.value = [...collaborators.value, data]
      collabStore.setCollaborators(collaborators.value)
    }
    // 显示加入通知（不显示自己的加入）
    const currentUserId = getCurrentUserId()
    if (data.userId !== currentUserId) {
      addCollaborateNotice(data.userName, data.userAvatar || '', data.color || '#3b82f6', 'join')
    }
  })
  ws.on('user_leave', (data: { userId: number; userName?: string; userAvatar?: string; color?: string }) => {
    // 找到离开的用户信息，用于显示通知
    const leavingUser = collaborators.value.find(c => c.userId === data.userId)
    collaborators.value = collaborators.value.filter(c => c.userId !== data.userId)
    collabStore.setCollaborators(collaborators.value)
    // 显示离开通知（使用已有的用户信息或消息中的信息）
    const currentUserId = getCurrentUserId()
    if (data.userId !== currentUserId) {
      const username = leavingUser?.userName || data.userName || t('editor.unknownUser')
      const avatar = leavingUser?.userAvatar || data.userAvatar || ''
      const color = leavingUser?.color || data.color || '#3b82f6'
      addCollaborateNotice(username, avatar, color, 'leave')
    }
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
  } else if (cmd === 'saveAsTemplate') {
    // Pre-fill form with document title
    templateForm.value.name = (doc.title.value || t('editor.untitledDoc')) + t('editor.templateSuffix')
    templateForm.value.description = ''
    templateForm.value.category = ''
    showSaveAsTemplateDialog.value = true
  }
}

async function handleSaveAsTemplate() {
  if (!templateForm.value.name.trim()) {
    ElMessage.warning(t('editor.templateNameRequired'))
    return
  }
  templateSaving.value = true
  try {
    await saveDocumentAsTemplate(docId.value, {
      name: templateForm.value.name,
      description: templateForm.value.description,
      category: templateForm.value.category
    })
    ElMessage.success(t('editor.savedAsTemplate'))
    showSaveAsTemplateDialog.value = false
  } catch (error) {
    ElMessage.error(t('editor.saveTemplateFailed'))
  } finally {
    templateSaving.value = false
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
  ElMessage.success(t('editor.exportSuccess'))
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
  await ElMessageBox.confirm(t('editor.rollbackConfirm', { version }), t('editor.rollbackTitle'))
  await rollbackVersion(docId.value, version)
  ElMessage.success(t('editor.rolledBack'))
  showHistory.value = false
  loadDocument()
}

function onSubEditorSave(content: string) {
  doc.saveContent(content)
}

onMounted(async () => {
  userStore.fetchUserInfo()
  initWebSocket()   // WebSocket 仅用于协作者在线状态
  await loadDocument()  // 加载文档数据并初始化编辑器
})

onBeforeUnmount(() => {
  doc.cleanup()
  if (editor.value && doc.docType.value === 'doc') {
    doc.saveContent(JSON.stringify(editor.value.getJSON()))
  }
  // 清理 Yjs 协作
  if (yjsCollaboration) {
    yjsCollaboration.destroy()
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
    // 清理旧的 Yjs 协作
    if (yjsCollaboration) {
      yjsCollaboration.destroy()
    }
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
.collab-container {
  display: flex;
  align-items: center;
  gap: 12px;
}
.collab-avatars {
  display: flex;
  align-items: center;
}
.collab-avatar-wrapper {
  margin-left: -8px;
}
.collab-avatar-wrapper:first-child {
  margin-left: 0;
}
.collab-avatar {
  border: 2px solid;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}
.collab-avatar:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
  z-index: 10;
  position: relative;
}
.collab-badge {
  display: flex;
  align-items: center;
}
.collab-count {
  font-size: 12px;
  color: var(--kx-text-secondary);
  background: var(--kx-bg-secondary);
  padding: 4px 8px;
  border-radius: 12px;
  white-space: nowrap;
}
.yjs-status {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  padding: 4px 8px;
  border-radius: 12px;
  white-space: nowrap;
}
.yjs-status.connecting {
  color: #e6a23c;
}
.yjs-status.connected {
  color: #67c23a;
}
.yjs-status.disconnected {
  color: #f56c6c;
}
.status-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
}
.status-dot.green {
  background-color: #67c23a;
}
.status-dot.red {
  background-color: #f56c6c;
}
/* Fade animation for collaborators */
.collab-fade-enter-active,
.collab-fade-leave-active {
  transition: all 0.3s ease;
}
.collab-fade-enter-from,
.collab-fade-leave-to {
  opacity: 0;
  transform: scale(0.8);
}
.collab-fade-move {
  transition: transform 0.3s ease;
}
/* Responsive styles */
@media (max-width: 768px) {
  .collab-avatar {
    width: 24px !important;
    height: 24px !important;
    line-height: 24px !important;
  }
  .collab-avatar-wrapper {
    margin-left: -6px;
  }
  .collab-count {
    font-size: 11px;
    padding: 2px 6px;
  }
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
.editor-skeleton {
  width: 100%;
  max-width: 800px;
  padding: 40px 24px;
  margin: 0 auto;
}
.editor-skeleton-content {
  display: flex;
  flex-direction: column;
}
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 60px 20px;
  width: 100%;
}
.error-message {
  font-size: 15px;
  color: var(--kx-text-secondary);
  text-align: center;
  max-width: 400px;
}
.editor-content-wrapper {
  position: relative;
  display: flex;
  justify-content: center;
  width: 100%;
  max-width: 880px;
}
.editor-content {
  position: relative;
  width: 100%;
  max-width: 800px;
  padding: 40px 24px;
  min-height: calc(100vh - 96px);
}
.editor-content.has-comment-bubbles {
  padding-right: 48px;
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

/* 协作状态条样式 */
.collab-status-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 16px;
  font-size: 13px;
  transition: all 0.3s;
  flex-shrink: 0;
}
.collab-status-bar.warning { background: #fef0e7; color: #e6a23c; }
.collab-status-bar.error { background: #fef0f0; color: #f56c6c; }
.collab-status-bar.success { background: #f0f9eb; color: #67c23a; }
.collab-status-bar.info { background: #ecf5ff; color: #409eff; }

/* slide-down transition */
.slide-down-enter-active,
.slide-down-leave-active {
  transition: all 0.3s ease;
}
.slide-down-enter-from,
.slide-down-leave-to {
  opacity: 0;
  transform: translateY(-100%);
  max-height: 0;
  padding-top: 0;
  padding-bottom: 0;
}
.slide-down-enter-to,
.slide-down-leave-from {
  max-height: 40px;
}

/* 协作者光标样式 */
.collaboration-cursor__caret {
  border-left: 2px solid;
  border-right: none;
  margin-left: -1px;
  position: relative;
}
.collaboration-cursor__label {
  font-size: 11px;
  padding: 1px 4px;
  border-radius: 3px;
  position: absolute;
  top: -1.2em;
  left: -1px;
  white-space: nowrap;
  color: white;
}

/* 远程选区高亮 */
.collaboration-cursor__selection {
  opacity: 0.3;
}

/* 协作通知提示样式 */
.collaborate-notices {
  position: absolute;
  top: 60px;
  right: 16px;
  z-index: 100;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.collaborate-notice {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  font-size: 13px;
  color: #333;
  backdrop-filter: blur(4px);
}

.notice-avatar {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  color: #fff;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: 2px solid;
}

.notice-fade-enter-active {
  animation: noticeIn 0.3s ease;
}

.notice-fade-leave-active {
  animation: noticeOut 0.3s ease;
}

@keyframes noticeIn {
  from {
    opacity: 0;
    transform: translateX(20px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

@keyframes noticeOut {
  from {
    opacity: 1;
    transform: translateX(0);
  }
  to {
    opacity: 0;
    transform: translateX(20px);
  }
}

/* ===== Responsive: 768px - Tablet ===== */
@media (max-width: 768px) {
  .editor-header {
    padding: 0 12px;
  }
  .title-input {
    width: 200px;
    font-size: 15px;
  }
  .file-name-header {
    max-width: 200px;
    font-size: 15px;
  }
  .editor-content {
    padding: 24px 16px;
    max-width: 100%;
  }
  .editor-content-wrapper {
    max-width: 100%;
  }
  .editor-header-center {
    display: none;
  }
}

/* ===== Responsive: 640px - Large Phone ===== */
@media (max-width: 640px) {
  .editor-header {
    height: 44px;
    padding: 0 8px;
  }
  .title-input {
    width: 140px;
    font-size: 14px;
  }
  .file-name-header {
    max-width: 140px;
    font-size: 14px;
  }
  .editor-header-right {
    gap: 4px;
  }
  .editor-header-right .el-button span {
    display: none;
  }
  .save-status {
    display: none;
  }
  .editor-content {
    padding: 16px 12px;
  }
  .collaborate-notices {
    right: 8px;
    top: 50px;
  }
}

/* ===== Responsive: 480px - Small Phone ===== */
@media (max-width: 480px) {
  .editor-header {
    height: 40px;
    padding: 0 6px;
  }
  .title-input {
    width: 100px;
    font-size: 13px;
  }
  .file-name-header {
    max-width: 100px;
    font-size: 13px;
  }
  .editor-content {
    padding: 12px 8px;
    min-height: calc(100vh - 80px);
  }
  .editor-content.has-comment-bubbles {
    padding-right: 8px;
  }
}
</style>
