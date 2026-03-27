<template>
  <div class="kb-chat-panel" :class="{ open: isOpen }">
    <div class="chat-header">
      <div class="chat-title">
        <el-icon><ChatDotRound /></el-icon>
        <span>AI 助手</span>
      </div>
      <div class="chat-actions">
        <el-tooltip content="清空对话" placement="bottom">
          <el-icon class="action-btn" @click="clearChat"><Delete /></el-icon>
        </el-tooltip>
        <el-icon class="action-btn close-btn" @click="$emit('close')"><Close /></el-icon>
      </div>
    </div>
    <div class="chat-messages" ref="messagesRef">
      <div v-if="messages.length === 0" class="chat-welcome">
        <el-icon :size="48" color="#3370ff"><ChatDotRound /></el-icon>
        <h3>知识库 AI 助手</h3>
        <p>我可以帮您解答关于知识库内容的问题</p>
        <div class="quick-actions">
          <div class="quick-action" @click="sendQuickMessage('这个知识库包含哪些内容？')">
            <el-icon><QuestionFilled /></el-icon>
            <span>这个知识库包含哪些内容？</span>
          </div>
          <div class="quick-action" @click="sendQuickMessage('帮我总结最近的文档')">
            <el-icon><Document /></el-icon>
            <span>帮我总结最近的文档</span>
          </div>
          <div class="quick-action" @click="sendQuickMessage('查找相关的技术文档')">
            <el-icon><Search /></el-icon>
            <span>查找相关的技术文档</span>
          </div>
        </div>
      </div>
      <div v-for="(msg, idx) in messages" :key="idx" class="chat-message" :class="msg.role">
        <div class="message-avatar">
          <el-avatar v-if="msg.role === 'user'" :size="32">{{ userInitial }}</el-avatar>
          <el-avatar v-else :size="32" style="background: #3370ff"><el-icon><ChatDotRound /></el-icon></el-avatar>
        </div>
        <div class="message-content">
          <div class="message-text" v-html="formatMessage(msg.content)"></div>
          <div v-if="msg.role === 'assistant' && msg.sources?.length" class="message-sources">
            <div class="sources-label">参考来源:</div>
            <div v-for="src in msg.sources" :key="src.id" class="source-item" @click="$emit('open-doc', src.id)">
              <el-icon><Document /></el-icon>
              <span>{{ src.title }}</span>
            </div>
          </div>
        </div>
      </div>
      <div v-if="loading" class="chat-message assistant">
        <div class="message-avatar">
          <el-avatar :size="32" style="background: #3370ff"><el-icon><ChatDotRound /></el-icon></el-avatar>
        </div>
        <div class="message-content">
          <div class="typing-indicator">
            <span></span><span></span><span></span>
          </div>
        </div>
      </div>
    </div>
    <div class="chat-input-area">
      <el-input
        v-model="inputText"
        type="textarea"
        :rows="2"
        placeholder="输入您的问题..."
        @keydown.enter.ctrl="sendMessage"
        :disabled="loading"
      />
      <el-button type="primary" :loading="loading" :disabled="!inputText.trim()" @click="sendMessage">
        <el-icon><Promotion /></el-icon>
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

interface Message {
  role: 'user' | 'assistant'
  content: string
  sources?: { id: number; title: string }[]
}

const props = defineProps<{
  isOpen: boolean
  kbId: number
  kbName: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'open-doc', docId: number): void
}>()

const messages = ref<Message[]>([])
const inputText = ref('')
const loading = ref(false)
const messagesRef = ref<HTMLElement>()

const userInitial = computed(() => {
  const name = localStorage.getItem('nickname') || localStorage.getItem('username') || 'U'
  return name[0].toUpperCase()
})

function formatMessage(content: string): string {
  // Basic markdown-like formatting
  return content
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
    .replace(/`(.*?)`/g, '<code>$1</code>')
    .replace(/\n/g, '<br>')
}

function scrollToBottom() {
  nextTick(() => {
    if (messagesRef.value) {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight
    }
  })
}

async function sendMessage() {
  const text = inputText.value.trim()
  if (!text || loading.value) return

  messages.value.push({ role: 'user', content: text })
  inputText.value = ''
  scrollToBottom()

  loading.value = true
  try {
    // Simulate AI response (in production, call actual API)
    await new Promise(r => setTimeout(r, 1500))
    
    // Generate a contextual response
    const response = generateResponse(text)
    messages.value.push(response)
  } catch (e: any) {
    ElMessage.error(e.message || '发送失败')
  } finally {
    loading.value = false
    scrollToBottom()
  }
}

function sendQuickMessage(text: string) {
  inputText.value = text
  sendMessage()
}

function generateResponse(query: string): Message {
  const lowerQuery = query.toLowerCase()
  
  // Generate contextual response based on query type
  let content = ''
  let sources: { id: number; title: string }[] = []

  if (lowerQuery.includes('包含') || lowerQuery.includes('内容')) {
    content = `**${props.kbName}** 知识库目前包含多个文档分类，涵盖了技术文档、产品说明、操作指南等内容。\n\n您可以通过左侧目录浏览具体文档，或者直接向我提问您感兴趣的话题。`
  } else if (lowerQuery.includes('总结') || lowerQuery.includes('摘要')) {
    content = `根据知识库内容，最近的文档主要涉及以下主题：\n\n1. **技术规范** - 系统架构和开发规范\n2. **操作指南** - 用户操作手册\n3. **更新日志** - 版本更新记录\n\n如需了解具体内容，请点击下方参考来源查看原文。`
    sources = [
      { id: 1, title: '系统架构文档' },
      { id: 2, title: '用户操作手册' },
    ]
  } else if (lowerQuery.includes('查找') || lowerQuery.includes('搜索') || lowerQuery.includes('技术')) {
    content = `我为您找到了以下相关技术文档：\n\n- **API 接口文档** - 详细的接口说明\n- **数据库设计** - 数据模型和表结构\n- **部署指南** - 系统部署步骤\n\n您可以点击参考来源直接跳转到对应文档。`
    sources = [
      { id: 3, title: 'API接口文档' },
      { id: 4, title: '数据库设计说明' },
    ]
  } else {
    content = `感谢您的提问！关于"${query}"，我正在从知识库中检索相关信息。\n\n目前知识库中有多个相关文档，您可以通过以下方式获取更多信息：\n\n1. 在搜索框中输入关键词查找\n2. 浏览左侧目录分类\n3. 继续向我提问具体问题\n\n我会尽力为您提供准确的答案。`
  }

  return { role: 'assistant', content, sources }
}

function clearChat() {
  messages.value = []
}

onMounted(() => {
  // Load chat history from localStorage if needed
  const saved = localStorage.getItem(`kb-chat-${props.kbId}`)
  if (saved) {
    try {
      messages.value = JSON.parse(saved)
    } catch {}
  }
})
</script>

<style scoped>
.kb-chat-panel {
  position: fixed;
  right: -420px;
  top: 0;
  width: 400px;
  height: 100vh;
  background: #fff;
  box-shadow: -2px 0 12px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  z-index: 1000;
  transition: right 0.3s ease;
}
.kb-chat-panel.open {
  right: 0;
}
.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid var(--kx-border);
  background: #fafafa;
}
.chat-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  font-size: 16px;
}
.chat-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
.action-btn {
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  color: var(--kx-text-secondary);
  transition: all 0.15s;
}
.action-btn:hover {
  background: rgba(0, 0, 0, 0.06);
  color: var(--kx-text-primary);
}
.close-btn:hover {
  color: #f54a45;
}
.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}
.chat-welcome {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  text-align: center;
  color: var(--kx-text-secondary);
}
.chat-welcome h3 {
  margin: 16px 0 8px;
  color: var(--kx-text-primary);
}
.chat-welcome p {
  margin: 0 0 24px;
}
.quick-actions {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
  max-width: 280px;
}
.quick-action {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  text-align: left;
  transition: all 0.15s;
}
.quick-action:hover {
  border-color: var(--kx-primary);
  background: rgba(51, 112, 255, 0.04);
}
.chat-message {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}
.chat-message.user {
  flex-direction: row-reverse;
}
.message-content {
  max-width: 80%;
}
.message-text {
  padding: 10px 14px;
  border-radius: 12px;
  font-size: 14px;
  line-height: 1.6;
}
.chat-message.user .message-text {
  background: var(--kx-primary);
  color: #fff;
  border-bottom-right-radius: 4px;
}
.chat-message.assistant .message-text {
  background: #f5f6f7;
  color: var(--kx-text-primary);
  border-bottom-left-radius: 4px;
}
.message-text code {
  background: rgba(0, 0, 0, 0.06);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: monospace;
}
.chat-message.user .message-text code {
  background: rgba(255, 255, 255, 0.2);
}
.message-sources {
  margin-top: 8px;
  padding: 8px;
  background: #f9fafb;
  border-radius: 6px;
}
.sources-label {
  font-size: 12px;
  color: var(--kx-text-secondary);
  margin-bottom: 6px;
}
.source-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  font-size: 13px;
  color: var(--kx-primary);
  cursor: pointer;
  border-radius: 4px;
}
.source-item:hover {
  background: rgba(51, 112, 255, 0.08);
}
.typing-indicator {
  display: flex;
  gap: 4px;
  padding: 12px 16px;
}
.typing-indicator span {
  width: 8px;
  height: 8px;
  background: #bbb;
  border-radius: 50%;
  animation: typing 1.4s infinite ease-in-out;
}
.typing-indicator span:nth-child(2) {
  animation-delay: 0.2s;
}
.typing-indicator span:nth-child(3) {
  animation-delay: 0.4s;
}
@keyframes typing {
  0%, 60%, 100% { transform: translateY(0); }
  30% { transform: translateY(-6px); }
}
.chat-input-area {
  display: flex;
  gap: 8px;
  padding: 16px;
  border-top: 1px solid var(--kx-border);
  background: #fafafa;
}
.chat-input-area .el-textarea {
  flex: 1;
}
</style>
