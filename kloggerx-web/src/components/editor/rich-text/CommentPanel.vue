<template>
  <div class="comment-panel">
    <div class="comment-panel-header">
      <span class="panel-title">评论 ({{ comments.length }})</span>
      <el-icon class="panel-close" @click="$emit('close')"><Close /></el-icon>
    </div>
    <div ref="commentListRef" class="comment-list" v-loading="loading">
      <div v-if="!comments.length && !loading" class="comment-empty">暂无评论</div>
      <div v-for="comment in comments" :key="comment.id" class="comment-item" :class="{ resolved: comment.resolved }" :data-comment-id="comment.id">
        <div class="comment-header">
          <el-avatar :size="24">{{ comment.userName?.[0] || 'U' }}</el-avatar>
          <span class="comment-author">{{ comment.userName }}</span>
          <span class="comment-time">{{ formatCommentTime(comment.createdAt) }}</span>
        </div>
        <div class="comment-body" v-if="comment.selection">
          <div class="comment-selection">"{{ comment.selection }}"</div>
        </div>
        <div class="comment-content" v-html="renderMentions(comment.content)"></div>
        <div class="comment-actions">
          <el-button v-if="!comment.resolved && canResolveComment(comment)" text size="small" @click="handleResolve(comment.id)">标记已解决</el-button>
          <el-tag v-else-if="comment.resolved" size="small" type="success">已解决</el-tag>
          <el-button v-if="canDeleteComment(comment)" text size="small" type="danger" @click="handleDelete(comment.id)">删除</el-button>
        </div>
        <!-- Replies -->
        <div v-if="comment.replies?.length" class="comment-replies">
          <div v-for="reply in comment.replies" :key="reply.id" class="reply-item">
            <div class="comment-header">
              <el-avatar :size="20">{{ reply.userName?.[0] || 'U' }}</el-avatar>
              <span class="comment-author">{{ reply.userName }}</span>
              <span class="comment-time">{{ formatCommentTime(reply.createdAt) }}</span>
            </div>
            <div class="comment-content" v-html="renderMentions(reply.content)"></div>
          </div>
        </div>
        <div class="reply-input" v-if="!comment.resolved">
          <el-input v-model="replyTexts[comment.id]" size="small" placeholder="回复..." @keyup.enter="handleReply(comment.id)">
            <template #append>
              <el-button size="small" @click="handleReply(comment.id)">发送</el-button>
            </template>
          </el-input>
        </div>
      </div>
    </div>
    <div class="comment-input-area">
      <div class="mention-input-wrapper">
        <el-input
          ref="commentInputRef"
          v-model="newComment"
          type="textarea"
          :rows="2"
          placeholder="添加评论... 使用 @ 提及他人"
          resize="none"
          @input="handleInput"
          @keydown="handleKeydown"
        />
        <!-- Mention dropdown -->
        <div v-if="showMentionDropdown" class="mention-dropdown" ref="mentionDropdownRef">
          <div v-if="mentionLoading" class="mention-loading">
            <el-icon class="is-loading"><Loading /></el-icon>
            <span>搜索中...</span>
          </div>
          <div v-else-if="mentionUsers.length === 0" class="mention-empty">
            未找到用户
          </div>
          <div
            v-for="(user, index) in mentionUsers"
            :key="user.id"
            class="mention-item"
            :class="{ active: index === mentionActiveIndex }"
            @click="selectMention(user)"
          >
            <el-avatar :size="28" :src="user.avatar">
              {{ user.username?.[0] || 'U' }}
            </el-avatar>
            <span class="mention-username">{{ user.username }}</span>
          </div>
        </div>
      </div>
      <el-button type="primary" size="small" :disabled="!newComment.trim()" @click="handleAdd">发送</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch, nextTick, onUnmounted, computed } from 'vue'
import { addComment, getComments, resolveComment, deleteComment, searchUsers, type UserSearchResult } from '@/api/modules/collaborate'
import { useUserStore } from '@/store/modules/user'
import { ElMessage } from 'element-plus'
import type { Editor } from '@tiptap/vue-3'

interface CommentItem {
  id: number
  documentId: number
  userId: number
  userName: string
  content: string
  selection: string
  parentId: number | null
  resolved: boolean
  createdAt: string
  replies?: CommentItem[]
}

const props = defineProps<{
  documentId: number
  editor?: Editor
  activeCommentId?: string | null
  ownerId?: number  // Document owner ID for permission check
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

// 评论列表容器引用
const commentListRef = ref<HTMLElement | null>(null)

// 监听 activeCommentId 变化，滚动到对应评论
watch(() => props.activeCommentId, (commentId) => {
  if (commentId) {
    scrollToComment(Number(commentId))
  }
})

// 滚动到指定评论
function scrollToComment(commentId: number) {
  nextTick(() => {
    const commentEl = commentListRef.value?.querySelector(`[data-comment-id="${commentId}"]`)
    if (commentEl) {
      commentEl.scrollIntoView({ behavior: 'smooth', block: 'center' })
      // 添加高亮效果
      commentEl.classList.add('highlight')
      setTimeout(() => {
        commentEl.classList.remove('highlight')
      }, 2000)
    }
  })
}

const loading = ref(false)
const comments = ref<CommentItem[]>([])
const newComment = ref('')
const replyTexts = reactive<Record<number, string>>({})

// Get current user info for permission check
const userStore = useUserStore()
const currentUser = computed(() => userStore.user)
const currentUserId = computed(() => currentUser.value?.id || 0)
const currentUserRole = computed(() => currentUser.value?.role || 'member')
const isAdmin = computed(() => currentUserRole.value === 'admin')

// Permission check helpers
function canDeleteComment(comment: CommentItem): boolean {
  // Only comment author or admin can delete
  return comment.userId === currentUserId.value || isAdmin.value
}

function canResolveComment(comment: CommentItem): boolean {
  // Comment author, document owner, or admin can resolve
  return comment.userId === currentUserId.value || 
         props.ownerId === currentUserId.value || 
         isAdmin.value
}

// Mention functionality
const showMentionDropdown = ref(false)
const mentionUsers = ref<UserSearchResult[]>([])
const mentionLoading = ref(false)
const mentionActiveIndex = ref(0)
const mentionQuery = ref('')
const mentionStartPos = ref(-1)
const commentInputRef = ref<any>(null)
const mentionDropdownRef = ref<HTMLElement | null>(null)
let mentionDebounceTimer: ReturnType<typeof setTimeout> | null = null

// Render mentions in comment content with blue highlight
function renderMentions(content: string): string {
  if (!content) return ''
  // Escape HTML to prevent XSS
  const escaped = content
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#x27;')
  // Replace @username with highlighted span
  return escaped.replace(/@([a-zA-Z0-9_\-\u4e00-\u9fa5]+)/g, '<span class="mention-highlight">@$1</span>')
}

// Handle input for mention detection
function handleInput() {
  const text = newComment.value
  const cursorPos = commentInputRef.value?.input?.selectionStart || text.length

  // Find the last @ before cursor
  const textBeforeCursor = text.substring(0, cursorPos)
  const lastAtIndex = textBeforeCursor.lastIndexOf('@')

  if (lastAtIndex !== -1) {
    const textAfterAt = textBeforeCursor.substring(lastAtIndex + 1)
    // Check if there's a space after @ (which means @ is not for mention)
    if (!textAfterAt.includes(' ') && !textAfterAt.includes('\n')) {
      mentionQuery.value = textAfterAt
      mentionStartPos.value = lastAtIndex
      if (mentionQuery.value.length > 0) {
        searchMentionUsers(mentionQuery.value)
      } else {
        showMentionDropdown.value = false
      }
      return
    }
  }

  showMentionDropdown.value = false
}

// Search users with debounce
function searchMentionUsers(query: string) {
  if (mentionDebounceTimer) {
    clearTimeout(mentionDebounceTimer)
  }

  mentionLoading.value = true
  showMentionDropdown.value = true
  mentionActiveIndex.value = 0

  mentionDebounceTimer = setTimeout(async () => {
    try {
      const res: any = await searchUsers(query)
      mentionUsers.value = res.data || []
    } catch {
      mentionUsers.value = []
    } finally {
      mentionLoading.value = false
    }
  }, 300)
}

// Handle keyboard navigation in mention dropdown
function handleKeydown(event: KeyboardEvent) {
  if (!showMentionDropdown.value) return

  switch (event.key) {
    case 'ArrowDown':
      event.preventDefault()
      mentionActiveIndex.value = (mentionActiveIndex.value + 1) % mentionUsers.value.length
      break
    case 'ArrowUp':
      event.preventDefault()
      mentionActiveIndex.value = (mentionActiveIndex.value - 1 + mentionUsers.value.length) % mentionUsers.value.length
      break
    case 'Enter':
      event.preventDefault()
      if (mentionUsers.value.length > 0) {
        selectMention(mentionUsers.value[mentionActiveIndex.value])
      }
      break
    case 'Escape':
      showMentionDropdown.value = false
      break
  }
}

// Select a user from mention dropdown
function selectMention(user: UserSearchResult) {
  if (mentionStartPos.value === -1) return

  const beforeMention = newComment.value.substring(0, mentionStartPos.value)
  const afterMention = newComment.value.substring(mentionStartPos.value + mentionQuery.value.length + 1)
  newComment.value = beforeMention + '@' + user.username + ' ' + afterMention
  showMentionDropdown.value = false

  // Focus back to input
  nextTick(() => {
    const input = commentInputRef.value?.input
    if (input) {
      const newCursorPos = mentionStartPos.value + user.username.length + 2 // +2 for @ and space
      input.setSelectionRange(newCursorPos, newCursorPos)
      input.focus()
    }
  })
}

// Close mention dropdown when clicking outside
function handleClickOutside(event: MouseEvent) {
  const dropdown = mentionDropdownRef.value
  const input = commentInputRef.value?.$el
  if (dropdown && !dropdown.contains(event.target as Node) && !input?.contains(event.target as Node)) {
    showMentionDropdown.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  fetchComments()
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  if (mentionDebounceTimer) {
    clearTimeout(mentionDebounceTimer)
  }
})

function formatCommentTime(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return '刚刚'
  if (mins < 60) return `${mins}分钟前`
  const hours = Math.floor(diff / 3600000)
  if (hours < 24) return `${hours}小时前`
  return d.toLocaleDateString('zh-CN')
}

async function fetchComments() {
  loading.value = true
  try {
    const res: any = await getComments(props.documentId)
    const all: CommentItem[] = res.data || []
    // Group replies under parent comments
    const topLevel: CommentItem[] = []
    const replyMap: Record<number, CommentItem[]> = {}
    for (const c of all) {
      if (c.parentId) {
        if (!replyMap[c.parentId]) replyMap[c.parentId] = []
        replyMap[c.parentId].push(c)
      } else {
        topLevel.push(c)
      }
    }
    for (const c of topLevel) {
      c.replies = replyMap[c.id] || []
    }
    comments.value = topLevel
  } catch {
    // handled
  } finally {
    loading.value = false
  }
}

async function handleAdd() {
  if (!newComment.value.trim()) return
  let selection = ''
  let hasSelection = false
  if (props.editor) {
    const { from, to } = props.editor.state.selection
    if (from !== to) {
      selection = props.editor.state.doc.textBetween(from, to, ' ')
      hasSelection = true
    }
  }
  try {
    const res: any = await addComment({
      documentId: props.documentId,
      content: newComment.value.trim(),
      selection,
    })
    newComment.value = ''
    // 如果有选区，应用评论标记
    if (hasSelection && props.editor && res?.data?.id) {
      const commentId = String(res.data.id)
      props.editor.commands.setCommentMark(commentId)
    }
    fetchComments()
  } catch {
    // handled
  }
}

async function handleReply(parentId: number) {
  const text = replyTexts[parentId]?.trim()
  if (!text) return
  try {
    await addComment({
      documentId: props.documentId,
      content: text,
      parentId,
    })
    replyTexts[parentId] = ''
    fetchComments()
  } catch {
    // handled
  }
}

async function handleResolve(id: number) {
  try {
    await resolveComment(id)
    // 更新编辑器中的评论标记为已解决状态
    if (props.editor) {
      props.editor.commands.updateCommentMarkResolved(String(id), true)
    }
    fetchComments()
    ElMessage.success('已标记为已解决')
  } catch {
    // handled
  }
}

async function handleDelete(id: number) {
  try {
    await deleteComment(id)
    // 移除编辑器中的评论标记
    if (props.editor) {
      props.editor.commands.unsetCommentMark(String(id))
    }
    fetchComments()
  } catch {
    // handled
  }
}


</script>

<style scoped>
.comment-panel {
  position: fixed;
  right: 0;
  top: 0;
  bottom: 0;
  width: 360px;
  background: #fff;
  border-left: 1px solid var(--kx-border, #e5e6eb);
  display: flex;
  flex-direction: column;
  z-index: 200;
  box-shadow: -4px 0 12px rgba(0, 0, 0, 0.06);
}
.comment-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--kx-border, #e5e6eb);
  flex-shrink: 0;
}
.panel-title {
  font-weight: 600;
  font-size: 15px;
}
.panel-close {
  cursor: pointer;
  font-size: 18px;
  color: var(--kx-text-secondary, #646a73);
}
.panel-close:hover {
  color: var(--kx-text-primary, #1f2329);
}
.comment-list {
  flex: 1;
  overflow-y: auto;
  padding: 12px 16px;
}
.comment-empty {
  text-align: center;
  padding: 40px 0;
  color: var(--kx-text-placeholder, #bbbfc4);
  font-size: 14px;
}
.comment-item {
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}
.comment-item.resolved {
  opacity: 0.6;
}
.comment-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}
.comment-author {
  font-size: 13px;
  font-weight: 500;
  color: var(--kx-text-primary, #1f2329);
}
.comment-time {
  font-size: 11px;
  color: var(--kx-text-placeholder, #bbbfc4);
  margin-left: auto;
}
.comment-selection {
  font-size: 12px;
  color: var(--kx-text-secondary, #646a73);
  background: #f5f6f7;
  padding: 4px 8px;
  border-radius: 4px;
  border-left: 3px solid #3370ff;
  margin-bottom: 6px;
  line-height: 1.5;
}
.comment-content {
  font-size: 13px;
  line-height: 1.6;
  color: var(--kx-text-primary, #1f2329);
}
.comment-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 6px;
}
.comment-replies {
  margin-top: 8px;
  margin-left: 16px;
  padding-left: 12px;
  border-left: 2px solid #f0f0f0;
}
.reply-item {
  padding: 6px 0;
}
.reply-input {
  margin-top: 8px;
}
.comment-input-area {
  padding: 12px 16px;
  border-top: 1px solid var(--kx-border, #e5e6eb);
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex-shrink: 0;
}
.comment-input-area .el-button {
  align-self: flex-end;
}

/* 评论高亮动画 */
@keyframes commentHighlight {
  0% {
    background-color: transparent;
  }
  50% {
    background-color: #fef3cd;
  }
  100% {
    background-color: transparent;
  }
}

.comment-item.highlight {
  animation: commentHighlight 2s ease;
}

/* Mention highlight in comment content */
:deep(.mention-highlight) {
  color: #3370ff;
  font-weight: 600;
  cursor: pointer;
}

/* Mention input wrapper */
.mention-input-wrapper {
  position: relative;
  flex: 1;
}

/* Mention dropdown */
.mention-dropdown {
  position: absolute;
  bottom: 100%;
  left: 0;
  right: 0;
  background: #fff;
  border: 1px solid var(--kx-border, #e5e6eb);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  max-height: 240px;
  overflow-y: auto;
  z-index: 1000;
  margin-bottom: 4px;
}

.mention-loading,
.mention-empty {
  padding: 16px;
  text-align: center;
  color: var(--kx-text-secondary, #646a73);
  font-size: 13px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.mention-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.mention-item:hover,
.mention-item.active {
  background-color: #f5f6f7;
}

.mention-username {
  font-size: 14px;
  color: var(--kx-text-primary, #1f2329);
  font-weight: 500;
}
</style>
