<template>
  <div class="knowledge-home">
    <!-- Sidebar Toggle (mobile) -->
    <div v-if="!sidebarVisible" class="sidebar-toggle-btn" @click="sidebarVisible = true">
      <el-icon><Expand /></el-icon>
    </div>

    <!-- Left Sidebar: Conversation List -->
    <aside class="kb-sidebar" :class="{ collapsed: !sidebarVisible }">
      <div class="sidebar-header">
        <span class="sidebar-title">{{ $t('knowledge.conversationHistory') }}</span>
        <el-icon class="sidebar-collapse-btn" @click="sidebarVisible = false"><Fold /></el-icon>
      </div>
      <!-- Sidebar Skeleton -->
      <div v-if="pageLoading" class="sidebar-skeleton">
        <el-skeleton v-for="i in 6" :key="i" animated :loading="true" style="padding: 8px 14px">
          <template #template>
            <div style="display: flex; align-items: center; gap: 10px">
              <el-skeleton-item variant="circle" style="width: 28px; height: 28px; flex-shrink: 0" />
              <div style="flex: 1">
                <el-skeleton-item variant="text" style="width: 80%; height: 14px" />
                <el-skeleton-item variant="text" style="width: 50%; height: 12px; margin-top: 6px" />
              </div>
            </div>
          </template>
        </el-skeleton>
      </div>
      <ConversationList
        v-else
        ref="convListRef"
        :current-id="currentConversationId"
        :knowledge-base-id="selectedKbId"
        @select="handleConvSelect"
        @create="handleNewConversation"
      />
    </aside>

    <!-- Right Main Area -->
    <div class="kb-main" :class="{ 'sidebar-hidden': !sidebarVisible }">
      <div class="kb-home-center" :class="{ 'welcome-mode': isWelcomeMode }">

        <!-- Mode A: Welcome/Empty State - All centered -->
        <div v-if="isWelcomeMode" class="kb-welcome-centered">
          <div class="kb-welcome-content">
            <div class="kb-welcome">
              <div class="kb-logo">
                <span class="kb-logo-icon">
                  <el-icon :size="32" color="#3370ff"><ChatDotRound /></el-icon>
                </span>
                <div class="kb-logo-text">
                  <span class="kb-title">{{ $t('knowledge.assistant') }}</span>
                  <span class="kb-subtitle">{{ $t('knowledge.assistantDesc') }}</span>
                </div>
              </div>
            </div>

            <!-- Quick Actions -->
            <div class="kb-quick-section">
              <div class="quick-title">{{ $t('knowledge.quickStart') }}</div>
              <div class="kb-quick-actions">
                <div class="kb-quick-card" @click="$router.push('/knowledge/list')">
                  <el-icon :size="28" color="#3370ff"><Collection /></el-icon>
                  <span>{{ $t('knowledge.kbSquare') }}</span>
                  <span class="quick-desc">{{ $t('knowledge.kbSquareDesc') }}</span>
                </div>
                <div class="kb-quick-card" @click="askQuestion(t('knowledge.contentQuestion'))">
                  <el-icon :size="28" color="#36b37e"><Document /></el-icon>
                  <span>{{ $t('knowledge.contentOverview') }}</span>
                  <span class="quick-desc">{{ $t('knowledge.contentOverviewDesc') }}</span>
                </div>
                <div class="kb-quick-card" @click="askQuestion(t('knowledge.summaryQuestion'))">
                  <el-icon :size="28" color="#ff7d00"><Clock /></el-icon>
                  <span>{{ $t('knowledge.recentUpdates') }}</span>
                  <span class="quick-desc">{{ $t('knowledge.recentUpdatesDesc') }}</span>
                </div>
              </div>
            </div>

            <!-- Input area inside welcome (centered) -->
            <div class="kb-chat-input-area kb-chat-input-welcome">
              <div class="kb-chat-options">
                <el-dropdown trigger="click" @command="handleModelChange">
                  <span class="kb-option-btn">
                    <el-icon><Cpu /></el-icon>
                    {{ selectedModelName || $t('knowledge.selectModel') }}
                    <el-icon class="el-icon--right"><ArrowDown /></el-icon>
                  </span>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item v-if="!availableModels.length" disabled>{{ $t('knowledge.noModelAvailable') }}</el-dropdown-item>
                      <el-option-group v-for="p in activeProviders" :key="p.id" :label="p.name">
                        <el-dropdown-item
                          v-for="m in getChatModels(p.id)"
                          :key="`${p.id}_${m.id}`"
                          :command="`${p.id}_${m.id}`"
                        >
                          {{ m.name }}
                          <el-tag v-if="m.isDefault" size="small" type="success" style="margin-left: 8px">{{ $t('knowledge.default') }}</el-tag>
                        </el-dropdown-item>
                      </el-option-group>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>

                <el-dropdown trigger="click" @command="handleKbChange">
                  <span class="kb-option-btn">
                    <el-icon><Collection /></el-icon>
                    {{ selectedKbName || $t('knowledge.selectKb') }}
                    <el-icon class="el-icon--right"><ArrowDown /></el-icon>
                  </span>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="__all__">
                        <el-icon><FolderOpened /></el-icon>
                        {{ $t('knowledge.allKb') }}
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
                        <el-icon><Promotion /></el-icon>{{ $t('knowledge.fastAnswer') }}
                      </el-dropdown-item>
                      <el-dropdown-item command="deep">
                        <el-icon><MagicStick /></el-icon>{{ $t('knowledge.deepThinking') }}
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
                  :placeholder="$t('knowledge.inputPlaceholder')"
                  class="kb-chat-textarea"
                  @keydown.enter.exact.prevent="handleSend"
                />
                <div class="kb-chat-bar-actions">
                  <el-tooltip :content="$t('knowledge.clearChat')" placement="top">
                    <el-button circle size="small" @click="clearMessages" :disabled="!messages.length">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </el-tooltip>
                  <el-button
                    type="primary"
                    circle
                    :icon="Promotion"
                    :disabled="!chatInput.trim() || isStreaming"
                    :loading="sending"
                    @click="handleSend"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Mode B: Conversation State -->
        <template v-else>
          <!-- Conversation Toolbar (top) -->
          <div v-if="currentConversationId && messages.length" class="kb-conv-toolbar">
            <el-dropdown trigger="click" @command="handleExport">
              <el-button size="small" text>
                <el-icon><Download /></el-icon> {{ $t('knowledge.exportChat') }}
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="markdown">
                    <el-icon><Document /></el-icon> Markdown
                  </el-dropdown-item>
                  <el-dropdown-item command="pdf" disabled>
                    <el-icon><Document /></el-icon> PDF
                    <el-tag size="small" type="info" style="margin-left:8px">{{ $t('knowledge.comingSoon') }}</el-tag>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
            <el-button size="small" text @click="shareDialogVisible = true">
              <el-icon><Share /></el-icon> {{ $t('common.share') }}
            </el-button>
          </div>

          <!-- Messages Area (middle, scrollable) -->
          <div class="kb-messages-area" ref="messagesContainer">
            <!-- Loading conversation messages -->
            <div v-if="loadingConversation" class="conv-detail-loading">
              <el-icon class="is-loading" :size="24"><Loading /></el-icon>
              <span>{{ $t('knowledge.loadingConversation') }}</span>
            </div>

            <!-- Chat Messages -->
            <template v-if="messages.length">
              <ChatMessage
                v-for="(msg, i) in messages"
                :key="i"
                :message="msg"
                @regenerate="handleRegenerate(i)"
                @feedback="(rating) => handleFeedback(msg, rating)"
                @suggestion-click="handleSuggestionClick"
                @copy="() => {}"
              />
            </template>

            <!-- Stop generating button -->
            <div v-if="isStreaming" class="stop-generate-bar">
              <el-button type="danger" plain size="small" @click="stopGenerate">
                <el-icon><VideoPause /></el-icon> {{ $t('knowledge.stopGenerate') }}
              </el-button>
            </div>
          </div>

          <!-- Chat Input Area (bottom, fixed) -->
          <div class="kb-chat-input-area">
            <div class="kb-chat-options">
              <el-dropdown trigger="click" @command="handleModelChange">
                <span class="kb-option-btn">
                  <el-icon><Cpu /></el-icon>
                  {{ selectedModelName || $t('knowledge.selectModel') }}
                  <el-icon class="el-icon--right"><ArrowDown /></el-icon>
                </span>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item v-if="!availableModels.length" disabled>{{ $t('knowledge.noModelAvailable') }}</el-dropdown-item>
                    <el-option-group v-for="p in activeProviders" :key="p.id" :label="p.name">
                      <el-dropdown-item
                        v-for="m in getChatModels(p.id)"
                        :key="`${p.id}_${m.id}`"
                        :command="`${p.id}_${m.id}`"
                      >
                        {{ m.name }}
                        <el-tag v-if="m.isDefault" size="small" type="success" style="margin-left: 8px">{{ $t('knowledge.default') }}</el-tag>
                      </el-dropdown-item>
                    </el-option-group>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>

              <el-dropdown trigger="click" @command="handleKbChange">
                <span class="kb-option-btn">
                  <el-icon><Collection /></el-icon>
                  {{ selectedKbName || $t('knowledge.selectKb') }}
                  <el-icon class="el-icon--right"><ArrowDown /></el-icon>
                </span>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="__all__">
                      <el-icon><FolderOpened /></el-icon>
                      {{ $t('knowledge.allKb') }}
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
                      <el-icon><Promotion /></el-icon>{{ $t('knowledge.fastAnswer') }}
                    </el-dropdown-item>
                    <el-dropdown-item command="deep">
                      <el-icon><MagicStick /></el-icon>{{ $t('knowledge.deepThinking') }}
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
                :placeholder="$t('knowledge.inputPlaceholder')"
                class="kb-chat-textarea"
                @keydown.enter.exact.prevent="handleSend"
              />
              <div class="kb-chat-bar-actions">
                <el-tooltip :content="$t('knowledge.clearChat')" placement="top">
                  <el-button circle size="small" @click="clearMessages" :disabled="!messages.length">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </el-tooltip>
                <el-button
                  type="primary"
                  circle
                  :icon="Promotion"
                  :disabled="!chatInput.trim() || isStreaming"
                  :loading="sending"
                  @click="handleSend"
                />
              </div>
            </div>
          </div>
        </template>
      </div>
    </div>

    <!-- Feedback Dialog -->
    <FeedbackDialog
      v-model:visible="feedbackDialogVisible"
      :message-id="feedbackTargetMsgId"
      @submit="handleFeedbackSubmit"
    />

    <!-- Share Dialog -->
    <ShareDialog
      v-model:visible="shareDialogVisible"
      :conversation-id="currentConversationId"
      :conversation-title="currentConversationTitle"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, computed, reactive } from 'vue'
import { Promotion, Delete, ArrowDown, FolderOpened, ChatDotRound, User, Clock, Document, Collection, Expand, Fold, Loading, Cpu, Operation, MagicStick, VideoPause, Download, Share } from '@element-plus/icons-vue'
import { getKnowledgeBaseList, streamChat, createConversation, getConversationDetail, submitFeedback, exportConversation } from '@/api/modules/knowledge'
import type { Conversation } from '@/api/modules/knowledge'
import { getAIModelSettings } from '@/api/modules/admin'
import { ElMessage } from 'element-plus'
import ConversationList from '@/components/knowledge/ConversationList.vue'
import ChatMessage from '@/components/knowledge/ChatMessage.vue'
import type { ChatMessageData } from '@/components/knowledge/ChatMessage.vue'
import FeedbackDialog from '@/components/knowledge/FeedbackDialog.vue'
import ShareDialog from '@/components/knowledge/ShareDialog.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

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

const chatInput = ref('')
const chatMode = ref<'fast' | 'deep'>('fast')
const selectedModel = ref('')
const selectedKb = ref('__all__')
const messages = ref<ChatMessageData[]>([])
const sending = ref(false)
const isStreaming = ref(false)
const messagesContainer = ref<HTMLElement>()
const sidebarVisible = ref(true)
const convListRef = ref<InstanceType<typeof ConversationList>>()
const pageLoading = ref(true)

// Conversation state
const currentConversationId = ref<number | null>(null)
const loadingConversation = ref(false)

// Feedback dialog
const feedbackDialogVisible = ref(false)
const feedbackTargetMsgId = ref(0)
let feedbackTargetMsg: ChatMessageData | null = null

// Share dialog
const shareDialogVisible = ref(false)
const currentConversationTitle = computed(() => {
  // Use conversation title from messages or fallback
  return messages.value[0]?.content?.slice(0, 30) || t('knowledge.chat')
})

// Abort controller for streaming
let abortController: AbortController | null = null

const kbList = ref<{ id: number; name: string }[]>([])
const providers = ref<Provider[]>([])

const selectedKbId = computed<number | null>(() => {
  if (selectedKb.value === '__all__') return null
  return parseInt(selectedKb.value) || null
})

// Welcome mode: no messages and no conversation selected and not loading
const isWelcomeMode = computed(() => messages.value.length === 0 && !currentConversationId.value && !loadingConversation.value)

const chatModeLabel = computed(() => chatMode.value === 'fast' ? t('knowledge.fastAnswer') : t('knowledge.deepThinking'))
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
  return m?.name || (availableModels.value[0]?.name || t('knowledge.selectModel'))
})

const selectedKbName = computed(() => {
  if (selectedKb.value === '__all__') return t('knowledge.allKb')
  const kb = kbList.value.find(x => String(x.id) === selectedKb.value)
  return kb?.name || t('knowledge.selectKb')
})

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
}

function handleKbChange(cmd: string) {
  selectedKb.value = cmd
  convListRef.value?.refresh()
}

function askQuestion(question: string) {
  chatInput.value = question
  handleSend()
}

function buildHistory(): { role: string; content: string }[] {
  const history: { role: string; content: string }[] = []
  const recentMessages = messages.value.slice(-10)
  for (const msg of recentMessages) {
    history.push({ role: msg.role, content: msg.content })
  }
  return history
}

// Handle conversation selection
async function handleConvSelect(conv: Conversation | null) {
  if (!conv) {
    currentConversationId.value = null
    messages.value = []
    return
  }
  if (conv.id === currentConversationId.value) return
  currentConversationId.value = conv.id
  loadingConversation.value = true
  messages.value = []
  try {
    const res: any = await getConversationDetail(conv.id, { msgPage: 1, msgPageSize: 50 })
    const detail = res.data
    if (detail?.messages?.length) {
      messages.value = detail.messages.map((m: any) => ({
        id: m.id,
        role: m.role,
        content: m.content,
        sources: m.sources || [],
        isStreaming: false,
        suggestions: [],
        feedback: null,
      }))
    }
    if (conv.model) {
      const found = availableModels.value.find(am => am.id.includes(conv.model))
      if (found) selectedModel.value = found.id
    }
    await nextTick()
    scrollToBottom()
  } catch (e) {
    ElMessage.error(t('knowledge.loadConvFailed'))
  } finally {
    loadingConversation.value = false
  }
}

function handleNewConversation() {
  currentConversationId.value = null
  messages.value = []
}

async function handleExport(format: string) {
  if (!currentConversationId.value) return
  try {
    const res = await exportConversation(currentConversationId.value, format as 'markdown' | 'pdf')
    const blob = new Blob([res.data])
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${currentConversationTitle.value}.md`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success(t('knowledge.exportSuccess'))
  } catch {
    ElMessage.error(t('knowledge.exportFailed'))
  }
}

async function handleSend() {
  const text = chatInput.value.trim()
  if (!text || sending.value || isStreaming.value) return

  // If no current conversation, create one first
  if (!currentConversationId.value) {
    try {
      const kbId = selectedKbId.value || undefined
      const modelId = getModelId() || undefined
      const res: any = await createConversation({ knowledgeBaseId: kbId, model: modelId })
      if (res.data) {
        currentConversationId.value = res.data.id
        convListRef.value?.addConversation(res.data)
      }
    } catch (e) {
      ElMessage.error(t('knowledge.createConvFailed'))
      return
    }
  }

  // Add user message
  messages.value.push({ role: 'user', content: text })
  chatInput.value = ''
  sending.value = true
  await nextTick()
  scrollToBottom()

  // Add empty AI message for streaming
  const aiMessage: ChatMessageData = reactive({
    role: 'assistant',
    content: '',
    isStreaming: true,
    sources: [],
    suggestions: [],
    feedback: null
  })
  messages.value.push(aiMessage)
  isStreaming.value = true
  sending.value = false

  // Create abort controller
  abortController = new AbortController()

  try {
    const history = buildHistory().slice(0, -1) // Exclude the empty AI message
    const response = await streamChat({
      conversationId: currentConversationId.value!,
      knowledgeBaseId: selectedKbId.value || undefined,
      question: text,
      history,
      model: getModelId() || undefined
    }, abortController.signal)

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`)
    }

    const reader = response.body!.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (line.startsWith('data: ')) {
          try {
            const data = JSON.parse(line.slice(6))
            switch (data.type) {
              case 'chunk':
                aiMessage.content += data.content
                await nextTick()
                scrollToBottom()
                break
              case 'sources':
                aiMessage.sources = data.data
                break
              case 'done':
                aiMessage.isStreaming = false
                aiMessage.id = data.data.messageId
                if (data.data.conversationId) {
                  currentConversationId.value = data.data.conversationId
                }
                break
              case 'suggestions':
                aiMessage.suggestions = data.data
                break
            }
          } catch {
            // Ignore malformed JSON lines
          }
        }
      }
    }
  } catch (e: any) {
    if (e.name === 'AbortError') {
      // User stopped generation
      if (!aiMessage.content) {
        aiMessage.content = t('knowledge.stoppedGenerate')
      }
    } else {
      aiMessage.content = `${t('knowledge.requestError', { msg: e.message || t('knowledge.requestFailedHint') })}`
    }
  } finally {
    aiMessage.isStreaming = false
    isStreaming.value = false
    abortController = null
    await nextTick()
    scrollToBottom()
  }
}

function stopGenerate() {
  if (abortController) {
    abortController.abort()
  }
}

function clearMessages() {
  messages.value = []
  currentConversationId.value = null
}

function handleRegenerate(index: number) {
  // Find the user message before this assistant message
  if (index > 0 && messages.value[index - 1]?.role === 'user') {
    const question = messages.value[index - 1].content
    // Remove the AI message
    messages.value.splice(index, 1)
    // Re-send
    chatInput.value = question
    handleSend()
  }
}

function handleSuggestionClick(question: string) {
  chatInput.value = question
  handleSend()
}

async function handleFeedback(msg: ChatMessageData, rating: number) {
  if (rating === 1) {
    // Direct thumbs up
    if (msg.id) {
      try {
        await submitFeedback(msg.id, { rating: 1 })
        msg.feedback = { rating: 1 }
        ElMessage.success(t('knowledge.thanksFeedback'))
      } catch {
        ElMessage.error(t('knowledge.feedbackFailed'))
      }
    }
  } else {
    // Open feedback dialog for thumbs down
    feedbackTargetMsg = msg
    feedbackTargetMsgId.value = msg.id || 0
    feedbackDialogVisible.value = true
  }
}

async function handleFeedbackSubmit(data: { feedbackType: string; comment: string; correctAnswer: string }) {
  if (!feedbackTargetMsg?.id) return
  try {
    await submitFeedback(feedbackTargetMsg.id, {
      rating: -1,
      feedbackType: data.feedbackType,
      comment: data.comment,
      correctAnswer: data.correctAnswer
    })
    feedbackTargetMsg.feedback = { rating: -1 }
    ElMessage.success(t('knowledge.feedbackSubmitted'))
  } catch {
    ElMessage.error(t('knowledge.feedbackFailed'))
  } finally {
    feedbackDialogVisible.value = false
    feedbackTargetMsg = null
  }
}

function scrollToBottom() {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTo({
      top: messagesContainer.value.scrollHeight,
      behavior: 'smooth'
    })
  }
}

async function loadProviders() {
  try {
    const res: any = await getAIModelSettings()
    if (res.data?.providers) {
      providers.value = res.data.providers
    }
    if (res.data?.kbSettings?.chatModel) {
      const defaultModelId = res.data.kbSettings.chatModel
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
  pageLoading.value = false
})
</script>

<style scoped>
.knowledge-home {
  display: flex;
  height: 100%;
  background: linear-gradient(180deg, #f0f5ff 0%, #fafbfc 100%);
  position: relative;
}

/* Sidebar */
.kb-sidebar {
  width: 260px;
  min-width: 260px;
  border-right: 1px solid var(--kx-border);
  display: flex;
  flex-direction: column;
  background: #f7f8fa;
  transition: all 0.3s ease;
  overflow: hidden;
}

.kb-sidebar.collapsed {
  width: 0;
  min-width: 0;
  border-right: none;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 14px 8px;
}

.sidebar-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--kx-text-primary);
}

.sidebar-collapse-btn {
  font-size: 18px;
  color: #909399;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.15s;
}

.sidebar-collapse-btn:hover {
  color: #3370ff;
  background: #e1ecff;
}

.sidebar-toggle-btn {
  position: absolute;
  left: 12px;
  top: 16px;
  z-index: 10;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  background: #fff;
  border: 1px solid var(--kx-border);
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  transition: all 0.15s;
}

.sidebar-toggle-btn:hover {
  border-color: #3370ff;
  color: #3370ff;
}

/* Main Content */
.kb-main {
  flex: 1;
  display: flex;
  justify-content: center;
  min-width: 0;
  height: 100%;
  overflow: hidden;
}

.kb-home-center {
  width: 100%;
  max-width: 800px;
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 0 20px;
}

/* Welcome mode: center everything */
.kb-home-center.welcome-mode {
  justify-content: center;
  align-items: center;
}

.kb-welcome-centered {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
  width: 100%;
  margin-bottom: 10vh;
}

.kb-welcome-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  max-width: 680px;
}

.kb-chat-input-welcome {
  margin-top: 36px;
  margin-bottom: 0;
  padding-bottom: 16px;
}

/* Messages area (middle, scrollable) */
.kb-messages-area {
  flex: 1;
  overflow-y: auto;
  padding: 20px 0;
  scroll-behavior: smooth;
}

/* Welcome */
.kb-welcome {
  text-align: center;
  margin-bottom: 0;
  padding: 0;
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

/* Chat input (fixed at bottom) */
.kb-chat-input-area {
  width: 100%;
  flex-shrink: 0;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 16px;
  padding: 16px 16px 20px;
  box-shadow: 0 -2px 16px rgba(0, 0, 0, 0.06);
  margin: 12px 0 60px;
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

/* Loading conversation */
.conv-detail-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 40px;
  color: #909399;
  font-size: 14px;
}

/* Conversation toolbar */
.kb-conv-toolbar {
  width: 100%;
  flex-shrink: 0;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 12px 4px 0;
}

.stop-generate-bar {
  display: flex;
  justify-content: center;
  margin-bottom: 16px;
}



/* Quick Actions */
.kb-quick-section {
  width: 100%;
  margin-top: 36px;
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

/* Responsive */
@media (max-width: 768px) {
  .kb-sidebar {
    position: absolute;
    left: 0;
    top: 0;
    bottom: 0;
    z-index: 100;
    box-shadow: 4px 0 16px rgba(0, 0, 0, 0.1);
  }

  .kb-sidebar.collapsed {
    width: 0;
  }

  .kb-grid {
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  }
}

/* 640px - Large Phone */
@media (max-width: 640px) {
  .kb-sidebar {
    width: 0;
    min-width: 0;
  }

  .kb-grid {
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 12px;
  }

  .kb-main-header {
    flex-wrap: wrap;
    gap: 8px;
  }

  .quick-actions {
    flex-wrap: wrap;
    gap: 8px;
  }
}

/* 480px - Small Phone */
@media (max-width: 480px) {
  .kb-grid {
    grid-template-columns: 1fr 1fr;
    gap: 8px;
  }

  .kb-main-header {
    padding: 8px 12px;
  }

  .quick-actions {
    gap: 6px;
  }

  .quick-action-card {
    padding: 10px;
    min-width: 0;
  }

  .quick-desc {
    display: none;
  }
}
</style>
