<template>
  <div class="comment-panel">
    <div class="comment-panel-header">
      <span class="panel-title">评论 ({{ comments.length }})</span>
      <el-icon class="panel-close" @click="$emit('close')"><Close /></el-icon>
    </div>
    <div class="comment-list" v-loading="loading">
      <div v-if="!comments.length && !loading" class="comment-empty">暂无评论</div>
      <div v-for="comment in comments" :key="comment.id" class="comment-item" :class="{ resolved: comment.resolved }">
        <div class="comment-header">
          <el-avatar :size="24">{{ comment.userName?.[0] || 'U' }}</el-avatar>
          <span class="comment-author">{{ comment.userName }}</span>
          <span class="comment-time">{{ formatCommentTime(comment.createdAt) }}</span>
        </div>
        <div class="comment-body" v-if="comment.selection">
          <div class="comment-selection">"{{ comment.selection }}"</div>
        </div>
        <div class="comment-content">{{ comment.content }}</div>
        <div class="comment-actions">
          <el-button v-if="!comment.resolved" text size="small" @click="handleResolve(comment.id)">标记已解决</el-button>
          <el-tag v-else size="small" type="success">已解决</el-tag>
          <el-button text size="small" type="danger" @click="handleDelete(comment.id)">删除</el-button>
        </div>
        <!-- Replies -->
        <div v-if="comment.replies?.length" class="comment-replies">
          <div v-for="reply in comment.replies" :key="reply.id" class="reply-item">
            <div class="comment-header">
              <el-avatar :size="20">{{ reply.userName?.[0] || 'U' }}</el-avatar>
              <span class="comment-author">{{ reply.userName }}</span>
              <span class="comment-time">{{ formatCommentTime(reply.createdAt) }}</span>
            </div>
            <div class="comment-content">{{ reply.content }}</div>
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
      <el-input v-model="newComment" type="textarea" :rows="2" placeholder="添加评论..." resize="none" />
      <el-button type="primary" size="small" :disabled="!newComment.trim()" @click="handleAdd">发送</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { addComment, getComments, resolveComment, deleteComment } from '@/api/modules/collaborate'
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
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const loading = ref(false)
const comments = ref<CommentItem[]>([])
const newComment = ref('')
const replyTexts = reactive<Record<number, string>>({})

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
  if (props.editor) {
    const { from, to } = props.editor.state.selection
    if (from !== to) {
      selection = props.editor.state.doc.textBetween(from, to, ' ')
    }
  }
  try {
    await addComment({
      documentId: props.documentId,
      content: newComment.value.trim(),
      selection,
    })
    newComment.value = ''
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
    fetchComments()
    ElMessage.success('已标记为已解决')
  } catch {
    // handled
  }
}

async function handleDelete(id: number) {
  try {
    await deleteComment(id)
    fetchComments()
  } catch {
    // handled
  }
}

onMounted(() => {
  fetchComments()
})
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
</style>
