<template>
  <div class="chat-message" :class="[message.role]">
    <div class="msg-avatar">
      <el-avatar v-if="message.role === 'user'" :size="36" style="background: var(--kx-primary);">
        <el-icon><User /></el-icon>
      </el-avatar>
      <el-avatar v-else :size="36" style="background: linear-gradient(135deg, #36b37e 0%, #00b8d9 100%);">
        <el-icon><ChatDotRound /></el-icon>
      </el-avatar>
    </div>
    <div class="msg-body">
      <div class="msg-header">
        <span class="msg-role">{{ message.role === 'user' ? $t('knowledge.chat.me') : $t('knowledge.chat.aiAssistant') }}</span>
      </div>

      <!-- User message -->
      <div v-if="message.role === 'user'" class="msg-content user-bubble">
        {{ message.content }}
      </div>

      <!-- Assistant message -->
      <template v-else>
        <!-- Loading state -->
        <div v-if="!message.content && message.isStreaming" class="msg-content typing-indicator">
          <span></span><span></span><span></span>
        </div>

        <!-- Content with markdown -->
        <div v-else class="msg-content assistant-card">
          <div class="markdown-body" v-html="renderedContent"></div>
          <span v-if="message.isStreaming" class="streaming-cursor">|</span>
        </div>

        <!-- Sources -->
        <div v-if="message.sources?.length" class="msg-sources">
          <div class="sources-label">
            <el-icon><FolderOpened /></el-icon>
            <span>{{ $t('knowledge.chat.referenceSources', { count: message.sources.length }) }}</span>
          </div>
          <div class="sources-list">
            <div
              v-for="(s, i) in message.sources"
              :key="i"
              class="source-chip"
              @click="openSource(s)"
            >
              <el-icon><Document /></el-icon>
              <span>{{ s.title || $t('knowledge.chat.unknownDoc') }}</span>
            </div>
          </div>
        </div>

        <!-- Actions bar -->
        <div v-if="!message.isStreaming && message.content" class="msg-actions">
          <span
            class="feedback-btn action-copy"
            :class="{ 'is-copied': copySuccess }"
            @click="handleCopy"
            :title="$t('common.copy')"
          >
            <svg v-if="!copySuccess" viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
              <path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/>
            </svg>
            <svg v-else viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
              <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/>
            </svg>
          </span>
          <span
            class="feedback-btn action-regenerate"
            @click="$emit('regenerate')"
            :title="$t('knowledge.chat.regenerate')"
          >
            <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
              <path d="M17.65 6.35C16.2 4.9 14.21 4 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08c-.82 2.33-3.04 4-5.65 4-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"/>
            </svg>
          </span>
          <span
            class="feedback-btn thumb-up"
            :class="{ 'is-active': message.feedback?.rating === 1 }"
            @click="$emit('feedback', 1)"
            :title="$t('knowledge.chat.helpful')"
          >
            <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
              <path d="M2 20h2c.55 0 1-.45 1-1v-9c0-.55-.45-1-1-1H2v11zm19.83-7.12c.11-.25.17-.52.17-.8V11c0-1.1-.9-2-2-2h-5.5l.92-4.65c.05-.22.02-.46-.08-.66-.23-.45-.52-.86-.88-1.22L14 2 7.59 8.41C7.21 8.79 7 9.3 7 9.83v7.84C7 18.95 8.05 20 9.34 20h8.11c.7 0 1.36-.37 1.72-.97l2.66-6.15z"/>
            </svg>
          </span>
          <span
            class="feedback-btn thumb-down"
            :class="{ 'is-active': message.feedback?.rating === -1 }"
            @click="$emit('feedback', -1)"
            :title="$t('knowledge.chat.notHelpful')"
          >
            <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
              <path d="M22 4h-2c-.55 0-1 .45-1 1v9c0 .55.45 1 1 1h2V4zM2.17 11.12c-.11.25-.17.52-.17.8V13c0 1.1.9 2 2 2h5.5l-.92 4.65c-.05.22-.02.46.08.66.23.45.52.86.88 1.22L10 22l6.41-6.41c.38-.38.59-.89.59-1.42V6.34C17 5.05 15.95 4 14.66 4h-8.1c-.71 0-1.37.37-1.72.97l-2.67 6.15z"/>
            </svg>
          </span>
        </div>

        <!-- Suggestions -->
        <div v-if="message.suggestions?.length && !message.isStreaming" class="msg-suggestions">
          <span
            v-for="(s, i) in message.suggestions"
            :key="i"
            class="suggestion-tag"
            @click="$emit('suggestion-click', s)"
          >
            {{ s }}
          </span>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { User, ChatDotRound, FolderOpened, Document } from '@element-plus/icons-vue'
import { marked } from 'marked'
import { ElMessage } from 'element-plus'

const { t } = useI18n()

export interface ChatMessageData {
  id?: number
  role: 'user' | 'assistant'
  content: string
  sources?: Array<{ docId: number; title: string; content: string }>
  isStreaming?: boolean
  suggestions?: string[]
  feedback?: { rating: number } | null
}

const props = defineProps<{
  message: ChatMessageData
}>()

const emit = defineEmits<{
  regenerate: []
  feedback: [rating: number]
  'suggestion-click': [question: string]
  copy: []
}>()

// Configure marked
marked.setOptions({ breaks: true, gfm: true })

const renderedContent = computed(() => {
  if (!props.message.content) return ''
  try {
    return marked.parse(props.message.content) as string
  } catch {
    return props.message.content.replace(/\n/g, '<br>')
  }
})

const copySuccess = ref(false)

function handleCopy() {
  navigator.clipboard.writeText(props.message.content)
  ElMessage.success(t('common.copiedToClipboard'))
  emit('copy')
  copySuccess.value = true
  setTimeout(() => { copySuccess.value = false }, 1500)
}

function openSource(source: { docId: number; title: string; content: string }) {
  if (source.docId) {
    window.open(`/doc/${source.docId}`, '_blank')
  }
}
</script>

<style scoped>
.chat-message {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.chat-message.user {
  flex-direction: row-reverse;
}

.chat-message.user .msg-body {
  align-items: flex-end;
}

.msg-avatar {
  flex-shrink: 0;
}

.msg-body {
  display: flex;
  flex-direction: column;
  max-width: 80%;
}

.msg-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.msg-role {
  font-size: 13px;
  font-weight: 600;
  color: var(--kx-text-primary);
}

.msg-content {
  padding: 12px 16px;
  border-radius: 12px;
  font-size: 14px;
  line-height: 1.7;
}

.user-bubble {
  background: linear-gradient(135deg, var(--kx-primary) 0%, #5080ff 100%);
  color: #fff;
  border-top-right-radius: 4px;
  white-space: pre-wrap;
  word-break: break-word;
}

.assistant-card {
  background: #fff;
  border: 1px solid var(--kx-border);
  border-top-left-radius: 4px;
  position: relative;
}

/* Streaming cursor */
.streaming-cursor {
  display: inline;
  font-weight: 700;
  color: var(--kx-primary);
  animation: blink 0.8s infinite;
}

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}

/* Typing indicator */
.typing-indicator {
  display: flex;
  gap: 4px;
  padding: 14px 20px;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-top-left-radius: 4px;
}

.typing-indicator span {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--kx-text-placeholder);
  animation: typing 1.2s infinite ease-in-out;
}

.typing-indicator span:nth-child(2) { animation-delay: 0.2s; }
.typing-indicator span:nth-child(3) { animation-delay: 0.4s; }

@keyframes typing {
  0%, 60%, 100% { transform: translateY(0); opacity: 0.4; }
  30% { transform: translateY(-6px); opacity: 1; }
}

/* Markdown styles */
.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3),
.markdown-body :deep(h4) {
  margin: 16px 0 8px 0;
  font-weight: 600;
  color: var(--kx-text-primary);
}

.markdown-body :deep(h1) { font-size: 20px; }
.markdown-body :deep(h2) { font-size: 18px; }
.markdown-body :deep(h3) { font-size: 16px; }

.markdown-body :deep(p) {
  margin: 8px 0;
  line-height: 1.7;
}

.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  margin: 8px 0;
  padding-left: 24px;
}

.markdown-body :deep(li) {
  margin: 4px 0;
  line-height: 1.6;
}

.markdown-body :deep(code) {
  background: #f5f7fa;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'SF Mono', Monaco, Consolas, monospace;
  font-size: 13px;
  color: #e83e8c;
}

.markdown-body :deep(pre) {
  background: #1e1e1e;
  border-radius: 8px;
  padding: 12px 16px;
  margin: 12px 0;
  overflow-x: auto;
}

.markdown-body :deep(pre code) {
  background: transparent;
  padding: 0;
  color: #d4d4d4;
  font-size: 13px;
  line-height: 1.5;
}

.markdown-body :deep(blockquote) {
  border-left: 4px solid var(--kx-primary);
  margin: 12px 0;
  padding: 8px 16px;
  background: #f5f8ff;
  color: var(--kx-text-secondary);
}

.markdown-body :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 12px 0;
  font-size: 13px;
}

.markdown-body :deep(th),
.markdown-body :deep(td) {
  border: 1px solid var(--kx-border);
  padding: 8px 12px;
  text-align: left;
}

.markdown-body :deep(th) {
  background: #f5f7fa;
  font-weight: 600;
}

.markdown-body :deep(a) {
  color: var(--kx-primary);
  text-decoration: none;
}

.markdown-body :deep(a:hover) {
  text-decoration: underline;
}

/* Sources */
.msg-sources {
  margin-top: 10px;
  padding: 10px 12px;
  background: #f7f8fa;
  border-radius: 8px;
}

.sources-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--kx-text-secondary);
  margin-bottom: 8px;
}

.sources-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.source-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 6px;
  font-size: 12px;
  color: var(--kx-text-primary);
  cursor: pointer;
  transition: all 0.15s;
}

.source-chip:hover {
  border-color: var(--kx-primary);
  color: var(--kx-primary);
  background: #f5f8ff;
}

/* Actions */
.msg-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  padding-left: 4px;
}

.feedback-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  color: #999;
  cursor: pointer;
  transition: all 0.2s;
}

.feedback-btn:hover {
  background: #ecf5ff;
  color: #409eff;
}

.feedback-btn.thumb-up:hover {
  color: #409eff;
}

.feedback-btn.thumb-down:hover {
  color: #f56c6c;
  background: #fef0f0;
}

.feedback-btn.thumb-up.is-active {
  color: #409eff;
  background: #ecf5ff;
}

.feedback-btn.thumb-down.is-active {
  color: #f56c6c;
  background: #fef0f0;
}

.feedback-btn.action-copy:hover {
  color: #409eff;
  background: #ecf5ff;
}

.feedback-btn.action-copy.is-copied {
  color: #67c23a;
  background: #f0f9eb;
}

.feedback-btn.action-regenerate:hover {
  color: #409eff;
  background: #ecf5ff;
}

/* Suggestions */
.msg-suggestions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}

.suggestion-tag {
  display: inline-block;
  padding: 6px 14px;
  background: #f0f5ff;
  border: 1px solid #d4e8ff;
  border-radius: 16px;
  font-size: 13px;
  color: var(--kx-primary);
  cursor: pointer;
  transition: all 0.15s;
}

.suggestion-tag:hover {
  background: var(--kx-primary);
  color: #fff;
  border-color: var(--kx-primary);
}
</style>
