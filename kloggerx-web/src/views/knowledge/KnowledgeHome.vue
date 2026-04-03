<template>
  <div class="knowledge-home">
    <div class="kb-home-center">
      <!-- Welcome Area -->
      <div class="kb-welcome">
        <div class="kb-logo">
          <span class="kb-logo-icon">
            <el-icon :size="32" color="#3370ff"><ChatDotRound /></el-icon>
          </span>
          <div class="kb-logo-text">
            <span class="kb-title">知识库助手</span>
            <span class="kb-subtitle">基于您的知识库内容智能问答</span>
          </div>
        </div>
      </div>

      <!-- Chat Input Area -->
      <div class="kb-chat-input-area">
        <div class="kb-chat-options">
          <el-dropdown trigger="click" @command="handleModelChange">
            <span class="kb-option-btn">
              <el-icon><Cpu /></el-icon>
              {{ selectedModelName || '选择模型' }}
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-if="!availableModels.length" disabled>暂无可用模型，请在系统管理中配置</el-dropdown-item>
                <el-option-group v-for="p in activeProviders" :key="p.id" :label="p.name">
                  <el-dropdown-item
                    v-for="m in getChatModels(p.id)"
                    :key="`${p.id}_${m.id}`"
                    :command="`${p.id}_${m.id}`"
                  >
                    {{ m.name }}
                    <el-tag v-if="m.isDefault" size="small" type="success" style="margin-left: 8px">默认</el-tag>
                  </el-dropdown-item>
                </el-option-group>
              </el-dropdown-menu>
            </template>
          </el-dropdown>

          <el-dropdown trigger="click" @command="handleKbChange">
            <span class="kb-option-btn">
              <el-icon><Collection /></el-icon>
              {{ selectedKbName || '选择知识库' }}
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="__all__">
                  <el-icon><FolderOpened /></el-icon>
                  全部知识库
                </el-dropdown-item>
                <el-dropdown-item
                  v-for="kb in kbList"
                  :key="kb.id"
                  :command="String(kb.id)"
                >
                  <el-icon><Collection /></el-icon>
                  {{ kb.name }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>

          <el-dropdown trigger="click" @command="handleModeChange">
            <span class="kb-option-btn">
              <el-icon><Operation /></el-icon>
              {{ chatModeLabel }}
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="fast">
                  <el-icon><Promotion /></el-icon>快速回答
                </el-dropdown-item>
                <el-dropdown-item command="deep">
                  <el-icon><MagicStick /></el-icon>深度思考
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>

        <div class="kb-chat-bar">
          <el-input
            v-model="chatInput"
            type="textarea"
            :autosize="{ minRows: 1, maxRows: 6 }"
            placeholder="输入你的问题，按 Enter 发送..."
            class="kb-chat-textarea"
            @keydown.enter.exact.prevent="handleSend"
          />
          <div class="kb-chat-bar-actions">
            <el-tooltip content="清空对话" placement="top">
              <el-button circle size="small" @click="clearMessages" :disabled="!messages.length">
                <el-icon><Delete /></el-icon>
              </el-button>
            </el-tooltip>
            <el-button
              type="primary"
              circle
              :icon="Promotion"
              :disabled="!chatInput.trim()"
              :loading="sending"
              @click="handleSend"
            />
          </div>
        </div>
      </div>

      <!-- Chat Messages -->
      <div v-if="messages.length" class="kb-chat-messages" ref="messagesContainer">
        <div v-for="(msg, i) in messages" :key="i" class="kb-chat-msg" :class="msg.role">
          <div class="msg-avatar">
            <el-avatar v-if="msg.role === 'user'" :size="36" style="background: var(--kx-primary);">
              <el-icon><User /></el-icon>
            </el-avatar>
            <el-avatar v-else :size="36" style="background: linear-gradient(135deg, #36b37e 0%, #00b8d9 100%);">
              <el-icon><ChatDotRound /></el-icon>
            </el-avatar>
          </div>
          <div class="msg-body">
            <div class="msg-header">
              <span class="msg-role">{{ msg.role === 'user' ? '我' : 'AI 助手' }}</span>
              <span v-if="msg.model" class="msg-model">{{ msg.model }}</span>
            </div>
            <div class="msg-content" v-html="formatMarkdown(msg.content)" />
            <div v-if="msg.sources?.length" class="msg-sources">
              <div class="sources-header" @click="toggleSources(msg)">
                <el-icon><FolderOpened /></el-icon>
                <span>参考来源 ({{ msg.sources.length }})</span>
                <el-icon class="sources-toggle" :class="{ expanded: msg.showSources }"><ArrowDown /></el-icon>
              </div>
              <transition name="slide">
                <div v-show="msg.showSources" class="sources-detail">
                  <div
                    v-for="(s, si) in msg.sources"
                    :key="si"
                    class="source-item"
                    @click="openSource(s)"
                  >
                    <span class="source-index">[{{ si + 1 }}]</span>
                    <span class="source-title">{{ s.documentTitle || '未知文档' }}</span>
                    <span v-if="s.chunkIndex !== undefined" class="source-chunk">第 {{ s.chunkIndex + 1 }} 段</span>
                    <el-icon class="source-arrow"><ArrowRight /></el-icon>
                  </div>
                </div>
              </transition>
              <div v-if="!msg.showSources" class="sources-preview">
                <el-tag
                  v-for="(s, si) in msg.sources.slice(0, 3)"
                  :key="si"
                  size="small"
                  type="info"
                  class="source-tag"
                  @click="openSource(s)"
                >
                  [{{ si + 1 }}] {{ s.documentTitle || '文档' }}
                </el-tag>
                <span v-if="msg.sources.length > 3" class="source-more" @click="msg.showSources = true">
                  +{{ msg.sources.length - 3 }} 更多
                </span>
              </div>
            </div>
            <div v-if="msg.role === 'assistant'" class="msg-actions">
              <el-button link size="small" @click="copyMessage(msg.content)">
                <el-icon><DocumentCopy /></el-icon> 复制
              </el-button>
              <el-button link size="small" @click="regenerate(msg)">
                <el-icon><Refresh /></el-icon> 重新生成
              </el-button>
            </div>
          </div>
        </div>
        <div v-if="sending" class="kb-chat-msg assistant">
          <div class="msg-avatar">
            <el-avatar :size="36" style="background: linear-gradient(135deg, #36b37e 0%, #00b8d9 100%);">
              <el-icon><ChatDotRound /></el-icon>
            </el-avatar>
          </div>
          <div class="msg-body">
            <div class="msg-header">
              <span class="msg-role">AI 助手</span>
            </div>
            <div class="msg-content typing-indicator">
              <span></span><span></span><span></span>
            </div>
          </div>
        </div>
      </div>

      <!-- Quick Actions -->
      <div v-if="!messages.length" class="kb-quick-section">
        <div class="quick-title">快速开始</div>
        <div class="kb-quick-actions">
          <div class="kb-quick-card" @click="$router.push('/knowledge/list')">
            <el-icon :size="28" color="#3370ff"><Collection /></el-icon>
            <span>知识库广场</span>
            <span class="quick-desc">浏览和管理知识库</span>
          </div>
          <div class="kb-quick-card" @click="askQuestion('这个知识库包含哪些内容？')">
            <el-icon :size="28" color="#36b37e"><Document /></el-icon>
            <span>内容概览</span>
            <span class="quick-desc">了解知识库内容</span>
          </div>
          <div class="kb-quick-card" @click="askQuestion('帮我总结最近的更新内容')">
            <el-icon :size="28" color="#ff7d00"><Clock /></el-icon>
            <span>最近更新</span>
            <span class="quick-desc">查看最新变更</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, computed } from 'vue'
import { Promotion, Delete, DocumentCopy, Refresh, ArrowDown, ArrowRight, FolderOpened, ChatDotRound, User, Clock, Document, Collection } from '@element-plus/icons-vue'
import { getKnowledgeBaseList, chatWithKnowledge, chatWithKnowledgeGlobal } from '@/api/modules/knowledge'
import { getAIModelSettings } from '@/api/modules/admin'
import { ElMessage } from 'element-plus'
import { marked } from 'marked'

interface ChatMessage {
  role: 'user' | 'assistant'
  content: string
  sources?: any[]
  model?: string
  showSources?: boolean
}

interface Model {
  id: string
  name: string
  vendor: string
  category: string
  type: 'chat' | 'embedding' | 'image'
  isDefault: boolean
}

interface Provider {
  id: number
  name: string
  baseUrl: string
  apiKey: string
  color: string
  isActive: boolean
  models: Model[]
}

const CHAT_HISTORY_KEY = 'kb_chat_history'

const chatInput = ref('')
const chatMode = ref<'fast' | 'deep'>('fast')
const selectedModel = ref('')
const selectedKb = ref('__all__')
const messages = ref<ChatMessage[]>([])
const sending = ref(false)
const messagesContainer = ref<HTMLElement>()

const kbList = ref<{ id: number; name: string }[]>([])
const providers = ref<Provider[]>([])

const chatModeLabel = computed(() => chatMode.value === 'fast' ? '快速回答' : '深度思考')
const activeProviders = computed(() => providers.value.filter(p => p.isActive && p.models?.length))

const availableModels = computed(() => {
  const models: { id: string; name: string; providerId: number; isDefault: boolean }[] = []
  for (const p of providers.value) {
    if (p.isActive && p.models?.length) {
      for (const m of p.models) {
        if (m.type === 'chat') {
          models.push({ id: `${p.id}_${m.id}`, name: m.name || m.id, providerId: p.id, isDefault: m.isDefault })
        }
      }
    }
  }
  return models
})

const selectedModelName = computed(() => {
  const m = availableModels.value.find(x => x.id === selectedModel.value)
  return m?.name || (availableModels.value[0]?.name || '选择模型')
})

const selectedKbName = computed(() => {
  if (selectedKb.value === '__all__') return '全部知识库'
  const kb = kbList.value.find(x => String(x.id) === selectedKb.value)
  return kb?.name || '选择知识库'
})

// Get the actual model ID from the composite key
function getModelId(): string {
  if (!selectedModel.value) return ''
  const parts = selectedModel.value.split('_')
  return parts.length > 1 ? parts.slice(1).join('_') : selectedModel.value
}

function getChatModels(providerId: number): Model[] {
  const provider = providers.value.find(p => p.id === providerId)
  return provider?.models.filter(m => m.type === 'chat') || []
}

function handleModeChange(cmd: string) {
  chatMode.value = cmd as 'fast' | 'deep'
}

function handleModelChange(cmd: string) {
  selectedModel.value = cmd
  saveChatHistory()
}

function handleKbChange(cmd: string) {
  selectedKb.value = cmd
  // Load chat history for this KB
  loadChatHistory()
}

function askQuestion(question: string) {
  chatInput.value = question
  handleSend()
}

// Build history for API request (last 10 messages)
function buildHistory(): { role: string; content: string }[] {
  const history: { role: string; content: string }[] = []
  const recentMessages = messages.value.slice(-10)
  for (const msg of recentMessages) {
    history.push({ role: msg.role, content: msg.content })
  }
  return history
}

async function handleSend() {
  const text = chatInput.value.trim()
  if (!text || sending.value) return

  messages.value.push({ role: 'user', content: text })
  chatInput.value = ''
  sending.value = true
  await nextTick()
  scrollToBottom()
  saveChatHistory()

  try {
    let res: any
    const modelId = getModelId()
    const history = buildHistory()

    if (selectedKb.value === '__all__') {
      res = await chatWithKnowledgeGlobal({ question: text, history, model: modelId })
    } else {
      const kbId = parseInt(selectedKb.value)
      res = await chatWithKnowledge(kbId, { question: text, history, model: modelId })
    }

    if (res.data?.answer) {
      const response: ChatMessage = {
        role: 'assistant',
        content: res.data.answer,
        sources: res.data.sources || [],
        model: selectedModelName.value,
        showSources: false,
      }
      messages.value.push(response)
    } else {
      messages.value.push({
        role: 'assistant',
        content: '抱歉，AI未能返回有效回答，请稍后重试。',
      })
    }
  } catch (e: any) {
    messages.value.push({
      role: 'assistant',
      content: `错误: ${e.response?.data?.message || e.message || '请求失败，请检查AI模型配置'}`,
    })
  } finally {
    sending.value = false
    nextTick(() => scrollToBottom())
    saveChatHistory()
  }
}

function clearMessages() {
  messages.value = []
  saveChatHistory()
}

function copyMessage(content: string) {
  navigator.clipboard.writeText(content)
  ElMessage.success('已复制到剪贴板')
}

async function regenerate(msg: ChatMessage) {
  // Find the previous user message
  const idx = messages.value.indexOf(msg)
  if (idx > 0) {
    const userMsg = messages.value[idx - 1]
    if (userMsg.role === 'user') {
      // Remove the assistant message and resend
      messages.value = messages.value.slice(0, idx)
      chatInput.value = userMsg.content
      await handleSend()
    }
  }
}

function openSource(source: any) {
  if (source.documentId) {
    // Open document with chunk index as hash for scroll position
    const chunkParam = source.chunkIndex !== undefined ? `?chunk=${source.chunkIndex}` : ''
    window.open(`/doc/${source.documentId}${chunkParam}`, '_blank')
  }
}

function toggleSources(msg: ChatMessage) {
  msg.showSources = !msg.showSources
}

// Configure marked options
marked.setOptions({
  breaks: true,
  gfm: true
})

function formatMarkdown(content: string): string {
  try {
    return marked.parse(content) as string
  } catch {
    // Fallback to basic formatting if marked fails
    return content
      .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
      .replace(/\*(.*?)\*/g, '<em>$1</em>')
      .replace(/`(.*?)`/g, '<code>$1</code>')
      .replace(/\n/g, '<br>')
  }
}

function scrollToBottom() {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

// Save chat history to localStorage
function saveChatHistory() {
  const data = {
    kbId: selectedKb.value,
    model: selectedModel.value,
    messages: messages.value,
  }
  localStorage.setItem(CHAT_HISTORY_KEY, JSON.stringify(data))
}

// Load chat history from localStorage
function loadChatHistory() {
  try {
    const saved = localStorage.getItem(CHAT_HISTORY_KEY)
    if (saved) {
      const data = JSON.parse(saved)
      // Only restore if KB matches
      if (data.kbId === selectedKb.value) {
        messages.value = data.messages || []
        if (data.model && availableModels.value.find(m => m.id === data.model)) {
          selectedModel.value = data.model
        }
      } else {
        // Clear messages if KB changed
        messages.value = []
      }
    }
  } catch (e) {
    console.error('Failed to load chat history:', e)
  }
}

async function loadProviders() {
  try {
    const res: any = await getAIModelSettings()
    if (res.data?.providers) {
      providers.value = res.data.providers
    }
    // Set default model from kbSettings or first available
    if (res.data?.kbSettings?.chatModel) {
      const defaultModelId = res.data.kbSettings.chatModel
      // Find the model in providers
      for (const p of providers.value) {
        if (p.isActive) {
          for (const m of p.models) {
            if (m.id === defaultModelId) {
              selectedModel.value = `${p.id}_${m.id}`
              return
            }
          }
        }
      }
    }
    // Fallback to first default or first available
    if (availableModels.value.length) {
      const defaultModel = availableModels.value.find(m => m.isDefault)
      selectedModel.value = defaultModel?.id || availableModels.value[0].id
    }
  } catch (e) {
    console.error('Failed to load AI settings:', e)
  }
}

async function loadKbList() {
  try {
    const res: any = await getKnowledgeBaseList({ page: 1, pageSize: 100 })
    kbList.value = (res.data?.list || []).map((kb: any) => ({ id: kb.id, name: kb.name }))
  } catch (e) {
    console.error('Failed to load KB list:', e)
  }
}

onMounted(async () => {
  await loadProviders()
  await loadKbList()
  loadChatHistory()
})
</script>

<style scoped>
.knowledge-home {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  min-height: 100%;
  padding: 20px;
  background: linear-gradient(180deg, #f0f5ff 0%, #fafbfc 100%);
}

.kb-home-center {
  width: 100%;
  max-width: 800px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

/* Welcome */
.kb-welcome {
  text-align: center;
  margin-bottom: 24px;
  padding: 20px 0;
}

.kb-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
}

.kb-logo-icon {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  background: linear-gradient(135deg, #e8f3ff 0%, #d4e8ff 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(51, 112, 255, 0.2);
}

.kb-logo-text {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.kb-title {
  font-size: 28px;
  font-weight: 700;
  color: var(--kx-text-primary);
  letter-spacing: -0.5px;
}

.kb-subtitle {
  font-size: 14px;
  color: var(--kx-text-secondary);
  margin-top: 2px;
}

/* Chat input */
.kb-chat-input-area {
  width: 100%;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 16px;
  padding: 16px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.08);
  margin-bottom: 20px;
}

.kb-chat-options {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.kb-option-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border: 1px solid var(--kx-border);
  border-radius: 20px;
  font-size: 13px;
  color: var(--kx-text-secondary);
  cursor: pointer;
  transition: all 0.15s;
  background: #fff;
}

.kb-option-btn:hover {
  border-color: var(--kx-primary);
  color: var(--kx-primary);
  background: #f5f8ff;
}

.kb-chat-bar {
  display: flex;
  align-items: flex-end;
  gap: 12px;
}

.kb-chat-textarea {
  flex: 1;
}

.kb-chat-textarea :deep(.el-textarea__inner) {
  border: none;
  box-shadow: none;
  padding: 8px 0;
  resize: none;
  font-size: 15px;
  line-height: 1.6;
}

.kb-chat-bar-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
  padding-bottom: 4px;
}

/* Chat messages */
.kb-chat-messages {
  width: 100%;
  max-height: calc(100vh - 380px);
  overflow-y: auto;
  margin-bottom: 20px;
  padding-right: 8px;
}

.kb-chat-msg {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.kb-chat-msg.user {
  flex-direction: row-reverse;
}

.kb-chat-msg.user .msg-body {
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

.msg-model {
  font-size: 11px;
  color: var(--kx-text-placeholder);
  background: var(--kx-bg-gray);
  padding: 2px 8px;
  border-radius: 10px;
}

.msg-content {
  padding: 12px 16px;
  border-radius: 12px;
  font-size: 14px;
  line-height: 1.7;
}

.kb-chat-msg.user .msg-content {
  background: linear-gradient(135deg, var(--kx-primary) 0%, #5080ff 100%);
  color: #fff;
  border-top-right-radius: 4px;
}

.kb-chat-msg.assistant .msg-content {
  background: #fff;
  border: 1px solid var(--kx-border);
  border-top-left-radius: 4px;
}

/* Markdown content styles */
.msg-content :deep(h1),
.msg-content :deep(h2),
.msg-content :deep(h3),
.msg-content :deep(h4) {
  margin: 16px 0 8px 0;
  font-weight: 600;
  color: var(--kx-text-primary);
}

.msg-content :deep(h1) { font-size: 20px; }
.msg-content :deep(h2) { font-size: 18px; }
.msg-content :deep(h3) { font-size: 16px; }
.msg-content :deep(h4) { font-size: 15px; }

.msg-content :deep(p) {
  margin: 8px 0;
  line-height: 1.7;
}

.msg-content :deep(ul),
.msg-content :deep(ol) {
  margin: 8px 0;
  padding-left: 24px;
}

.msg-content :deep(li) {
  margin: 4px 0;
  line-height: 1.6;
}

.msg-content :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 12px 0;
  font-size: 13px;
}

.msg-content :deep(th),
.msg-content :deep(td) {
  border: 1px solid var(--kx-border);
  padding: 8px 12px;
  text-align: left;
}

.msg-content :deep(th) {
  background: #f5f7fa;
  font-weight: 600;
  color: var(--kx-text-primary);
}

.msg-content :deep(tr:nth-child(even)) {
  background: #fafbfc;
}

.msg-content :deep(code) {
  background: #f5f7fa;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'SF Mono', Monaco, Consolas, monospace;
  font-size: 13px;
  color: #e83e8c;
}

.msg-content :deep(pre) {
  background: #1e1e1e;
  border-radius: 8px;
  padding: 12px 16px;
  margin: 12px 0;
  overflow-x: auto;
}

.msg-content :deep(pre code) {
  background: transparent;
  padding: 0;
  color: #d4d4d4;
  font-size: 13px;
  line-height: 1.5;
}

.msg-content :deep(blockquote) {
  border-left: 4px solid var(--kx-primary);
  margin: 12px 0;
  padding: 8px 16px;
  background: #f5f8ff;
  color: var(--kx-text-secondary);
}

.msg-content :deep(img) {
  max-width: 100%;
  border-radius: 8px;
  margin: 8px 0;
  cursor: pointer;
}

.msg-content :deep(a) {
  color: var(--kx-primary);
  text-decoration: none;
}

.msg-content :deep(a:hover) {
  text-decoration: underline;
}

.msg-content :deep(hr) {
  border: none;
  border-top: 1px solid var(--kx-border);
  margin: 16px 0;
}

.msg-sources {
  margin-top: 10px;
  padding: 10px 12px;
  background: #f7f8fa;
  border-radius: 8px;
}

.sources-header {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--kx-text-secondary);
  cursor: pointer;
  user-select: none;
  padding: 4px 0;
}

.sources-header:hover {
  color: var(--kx-primary);
}

.sources-toggle {
  margin-left: auto;
  transition: transform 0.2s;
}

.sources-toggle.expanded {
  transform: rotate(180deg);
}

.sources-preview {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
  align-items: center;
}

.source-tag {
  cursor: pointer;
  transition: all 0.15s;
}

.source-tag:hover {
  background: var(--kx-primary);
  color: #fff;
  border-color: var(--kx-primary);
}

.source-more {
  font-size: 12px;
  color: var(--kx-primary);
  cursor: pointer;
  margin-left: 4px;
}

.source-more:hover {
  text-decoration: underline;
}

.sources-detail {
  margin-top: 8px;
  border-radius: 6px;
  overflow: hidden;
  background: #fff;
  border: 1px solid var(--kx-border);
}

.source-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  cursor: pointer;
  transition: background 0.15s;
  border-bottom: 1px solid var(--kx-border);
}

.source-item:last-child {
  border-bottom: none;
}

.source-item:hover {
  background: #f5f8ff;
}

.source-index {
  font-size: 12px;
  font-weight: 600;
  color: var(--kx-primary);
  min-width: 24px;
}

.source-title {
  flex: 1;
  font-size: 13px;
  color: var(--kx-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.source-chunk {
  font-size: 11px;
  color: var(--kx-text-secondary);
  background: #f0f2f5;
  padding: 2px 6px;
  border-radius: 4px;
}

.source-arrow {
  color: var(--kx-text-placeholder);
  font-size: 12px;
}

.slide-enter-active,
.slide-leave-active {
  transition: all 0.2s ease;
}

.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  max-height: 0;
}

.slide-enter-to,
.slide-leave-from {
  opacity: 1;
  max-height: 500px;
}

.msg-actions {
  display: flex;
  gap: 12px;
  margin-top: 8px;
  padding-left: 4px;
}

.typing-indicator {
  display: flex;
  gap: 4px;
  padding: 14px 20px;
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

/* Quick Actions */
.kb-quick-section {
  width: 100%;
  margin-top: 20px;
}

.quick-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--kx-text-secondary);
  margin-bottom: 12px;
  text-align: center;
}

.kb-quick-actions {
  display: flex;
  gap: 12px;
  width: 100%;
  justify-content: center;
  flex-wrap: wrap;
}

.kb-quick-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 20px 24px;
  min-width: 140px;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 14px;
  color: var(--kx-text-primary);
}

.kb-quick-card:hover {
  border-color: var(--kx-primary);
  box-shadow: 0 4px 16px rgba(51, 112, 255, 0.15);
  transform: translateY(-2px);
}

.quick-desc {
  font-size: 12px;
  color: var(--kx-text-placeholder);
}
</style>
