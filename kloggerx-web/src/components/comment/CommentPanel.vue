<template>
  <el-drawer
    :model-value="true"
    :title="t('comment.title')"
    direction="rtl"
    size="380px"
    :show-close="true"
    @close="$emit('close')"
  >
    <template #header>
      <div class="panel-header">
        <span class="panel-title">{{ t('comment.title') }} ({{ totalCount }})</span>
      </div>
    </template>

    <div ref="commentListRef" class="comment-list" v-loading="loading">
      <div v-if="!comments.length && !loading" class="comment-empty">
        <el-icon :size="40" color="#c0c4cc"><ChatDotRound /></el-icon>
        <p>{{ t('comment.empty') }}</p>
      </div>

      <div
        v-for="comment in comments"
        :key="comment.id"
        class="comment-item"
        :class="{ resolved: comment.resolved, highlight: highlightId === comment.id }"
        :data-comment-id="comment.id"
      >
        <div class="comment-header">
          <el-avatar :size="28" :src="comment.user?.avatar || undefined">
            {{ comment.user?.username?.[0]?.toUpperCase() || 'U' }}
          </el-avatar>
          <div class="comment-meta">
            <span class="comment-author">{{ comment.user?.username || t('comment.unknownUser') }}</span>
            <span class="comment-time">{{ formatTime(comment.created_at) }}</span>
          </div>
          <el-tag v-if="comment.resolved" size="small" type="success">{{ t('comment.resolved') }}</el-tag>
        </div>

        <!-- Quoted text -->
        <div v-if="comment.quoted_text" class="comment-quoted">
          <span class="quoted-mark">"</span>{{ comment.quoted_text }}<span class="quoted-mark">"</span>
        </div>

        <div class="comment-content">{{ comment.content }}</div>

        <div class="comment-actions">
          <el-button
            v-if="!comment.resolved"
            text size="small"
            @click="handleResolve(comment.id)"
          >​{{ t('comment.resolve') }}</el-button>
          <el-button text size="small" @click="toggleReply(comment.id)">{{ t('comment.reply') }}</el-button>
          <el-button
            v-if="canDelete(comment)"
            text size="small" type="danger"
            @click="handleDelete(comment.id)"
          >{{ t('common.delete') }}</el-button>
        </div>

        <!-- Replies -->
        <div v-if="comment.replies?.length" class="comment-replies">
          <div v-for="reply in comment.replies" :key="reply.id" class="reply-item">
            <div class="comment-header">
              <el-avatar :size="22" :src="reply.user?.avatar || undefined">
                {{ reply.user?.username?.[0]?.toUpperCase() || 'U' }}
              </el-avatar>
              <div class="comment-meta">
                <span class="comment-author">{{ reply.user?.username || t('comment.unknownUser') }}</span>
                <span class="comment-time">{{ formatTime(reply.created_at) }}</span>
              </div>
            </div>
            <div class="comment-content reply-content">{{ reply.content }}</div>
            <div class="comment-actions" v-if="canDelete(reply)">
              <el-button text size="small" type="danger" @click="handleDelete(reply.id)">{{ t('common.delete') }}</el-button>
            </div>
          </div>
        </div>

        <!-- Reply input -->
        <div v-if="replyingTo === comment.id" class="reply-input">
          <el-input
            v-model="replyText"
            size="small"
            :placeholder="t('comment.replyPlaceholder')"
            @keyup.enter="submitReply(comment.id)"
          />
          <div class="reply-actions">
            <el-button size="small" @click="replyingTo = null">{{ t('common.cancel') }}</el-button>
            <el-button size="small" type="primary" :disabled="!replyText.trim()" @click="submitReply(comment.id)">{{ t('comment.send') }}</el-button>
          </div>
        </div>
      </div>
    </div>

    <div class="comment-input-area">
      <el-input
        v-model="newComment"
        type="textarea"
        :rows="3"
        :placeholder="t('comment.placeholder')"
        resize="none"
      />
      <el-button
        type="primary"
        :disabled="!newComment.trim()"
        @click="handleAdd"
      >{{ t('comment.submit') }}</el-button>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, nextTick } from 'vue'
import { createComment, listComments, deleteComment, resolveComment, type Comment } from '@/api/modules/comment'
import { useUserStore } from '@/store/modules/user'
import { ElMessage } from 'element-plus'
import { ChatDotRound } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  documentId: number
  activeCommentId?: string | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const { t } = useI18n()

const userStore = useUserStore()
const loading = ref(false)
const comments = ref<Comment[]>([])
const newComment = ref('')
const replyingTo = ref<number | null>(null)
const replyText = ref('')
const highlightId = ref<number | null>(null)
const commentListRef = ref<HTMLElement | null>(null)
const totalCount = ref(0)

// Permission check
function canDelete(comment: Comment): boolean {
  const userId = userStore.user?.id
  return comment.user_id === userId
}

function formatTime(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return t('comment.justNow')
  if (mins < 60) return t('comment.minutesAgo', { count: mins })
  const hours = Math.floor(diff / 3600000)
  if (hours < 24) return t('comment.hoursAgo', { count: hours })
  return d.toLocaleDateString()
}

function toggleReply(commentId: number) {
  if (replyingTo.value === commentId) {
    replyingTo.value = null
  } else {
    replyingTo.value = commentId
    replyText.value = ''
  }
}

async function fetchComments() {
  loading.value = true
  try {
    const res: any = await listComments(props.documentId)
    comments.value = res.data || []
    // Count total including replies
    let count = 0
    for (const c of comments.value) {
      count++
      if (c.replies) count += c.replies.length
    }
    totalCount.value = count
  } catch {
    // handled
  } finally {
    loading.value = false
  }
}

async function handleAdd() {
  if (!newComment.value.trim()) return
  try {
    await createComment(props.documentId, { content: newComment.value.trim() })
    newComment.value = ''
    fetchComments()
    ElMessage.success(t('comment.added'))
  } catch {
    ElMessage.error(t('comment.addFailed'))
  }
}

async function submitReply(parentId: number) {
  if (!replyText.value.trim()) return
  try {
    await createComment(props.documentId, {
      content: replyText.value.trim(),
      parent_id: parentId,
    })
    replyText.value = ''
    replyingTo.value = null
    fetchComments()
  } catch {
    ElMessage.error(t('comment.replyFailed'))
  }
}

async function handleResolve(id: number) {
  try {
    await resolveComment(id)
    fetchComments()
    ElMessage.success(t('comment.markedResolved'))
  } catch {
    ElMessage.error(t('common.failed'))
  }
}

async function handleDelete(id: number) {
  try {
    await deleteComment(id)
    fetchComments()
    ElMessage.success(t('comment.deleted'))
  } catch {
    ElMessage.error(t('comment.deleteFailed'))
  }
}

// Scroll to active comment
function scrollToComment(commentId: number) {
  nextTick(() => {
    highlightId.value = commentId
    const el = commentListRef.value?.querySelector(`[data-comment-id="${commentId}"]`)
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    }
    setTimeout(() => { highlightId.value = null }, 2000)
  })
}

watch(() => props.activeCommentId, (id) => {
  if (id) scrollToComment(Number(id))
})

onMounted(() => {
  fetchComments()
})
</script>

<style scoped>
.panel-header {
  display: flex;
  align-items: center;
}
.panel-title {
  font-weight: 600;
  font-size: 16px;
}
.comment-list {
  min-height: 200px;
}
.comment-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  color: #909399;
  gap: 12px;
}
.comment-empty p {
  font-size: 14px;
  margin: 0;
}
.comment-item {
  padding: 14px 0;
  border-bottom: 1px solid #f0f0f0;
}
.comment-item.resolved {
  opacity: 0.6;
}
.comment-item.highlight {
  background-color: #fef3cd;
  border-radius: 6px;
  padding: 14px 8px;
  transition: background-color 2s ease;
}
.comment-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.comment-meta {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.comment-author {
  font-size: 13px;
  font-weight: 500;
  color: #1f2329;
}
.comment-time {
  font-size: 11px;
  color: #bbbfc4;
}
.comment-quoted {
  font-size: 12px;
  color: #646a73;
  background: #f5f6f7;
  padding: 6px 10px;
  border-radius: 4px;
  border-left: 3px solid #409eff;
  margin-bottom: 8px;
  line-height: 1.5;
}
.quoted-mark {
  color: #409eff;
  font-weight: 600;
}
.comment-content {
  font-size: 14px;
  line-height: 1.6;
  color: #1f2329;
  white-space: pre-wrap;
  word-break: break-word;
}
.reply-content {
  margin-left: 30px;
}
.comment-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 6px;
}
.comment-replies {
  margin-top: 10px;
  margin-left: 20px;
  padding-left: 12px;
  border-left: 2px solid #f0f0f0;
}
.reply-item {
  padding: 8px 0;
}
.reply-input {
  margin-top: 10px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.reply-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
.comment-input-area {
  padding: 16px 0 0;
  border-top: 1px solid #f0f0f0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.comment-input-area .el-button {
  align-self: flex-end;
}
</style>
