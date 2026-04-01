<template>
  <div class="knowledge-home">
    <div class="kb-home-center">
      <!-- Welcome Area -->
      <div class="kb-welcome">
        <el-tag type="success" size="small" effect="plain" class="kb-welcome-tag">一站式对接</el-tag>
        <div class="kb-logo">
          <span class="kb-logo-text">IMA</span>
          <span class="kb-logo-sub">copilot</span>
        </div>
        <p class="kb-welcome-hint">有问题尽管问IMA</p>
      </div>

      <!-- Chat Input Area -->
      <div class="kb-chat-input-area">
        <div class="kb-chat-options">
          <el-dropdown trigger="click" @command="handleModeChange">
            <span class="kb-option-btn">
              <el-icon><ChatDotSquare /></el-icon>
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

          <el-dropdown trigger="click" @command="handleModelChange">
            <span class="kb-option-btn">
              <el-icon><Cpu /></el-icon>
              {{ selectedModelName || '选择模型' }}
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-if="!availableModels.length" disabled>暂无可用模型，请在系统管理中配置</el-dropdown-item>
                <el-dropdown-item
                  v-for="m in availableModels"
                  :key="m.id"
                  :command="m.id"
                >
                  {{ m.name }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>

          <el-dropdown trigger="click" @command="handleKbChange">
            <span class="kb-option-btn">
              <el-icon><Collection /></el-icon>
              {{ selectedKbName || '知识库1' }}
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="__all__">全部知识库</el-dropdown-item>
                <el-dropdown-item
                  v-for="kb in kbList"
                  :key="kb.id"
                  :command="String(kb.id)"
                >
                  {{ kb.name }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>

        <div class="kb-chat-bar">
          <el-input
            v-model="chatInput"
            type="textarea"
            :autosize="{ minRows: 1, maxRows: 4 }"
            placeholder="输入你的问题..."
            class="kb-chat-textarea"
            @keydown.enter.exact.prevent="handleSend"
          />
          <div class="kb-chat-bar-actions">
            <el-button circle size="small" :icon="Microphone" />
            <el-button
              type="primary"
              circle
              :icon="Promotion"
              :disabled="!chatInput.trim()"
              @click="handleSend"
            />
          </div>
        </div>
      </div>

      <!-- Chat Messages -->
      <div v-if="messages.length" class="kb-chat-messages" ref="messagesContainer">
        <div v-for="(msg, i) in messages" :key="i" class="kb-chat-msg" :class="msg.role">
          <div class="msg-avatar">
            <el-avatar v-if="msg.role === 'user'" :size="32" style="background: var(--kx-primary);">我</el-avatar>
            <el-avatar v-else :size="32" style="background: #36b37e;">AI</el-avatar>
          </div>
          <div class="msg-body">
            <div class="msg-content" v-html="msg.content" />
            <div v-if="msg.sources?.length" class="msg-sources">
              <span class="source-label">参考来源：</span>
              <el-tag v-for="(s, si) in msg.sources" :key="si" size="small" type="info" class="source-tag">{{ s }}</el-tag>
            </div>
          </div>
        </div>
        <div v-if="sending" class="kb-chat-msg assistant">
          <div class="msg-avatar">
            <el-avatar :size="32" style="background: #36b37e;">AI</el-avatar>
          </div>
          <div class="msg-body">
            <div class="msg-content typing-indicator">
              <span></span><span></span><span></span>
            </div>
          </div>
        </div>
      </div>

      <!-- Quick Actions -->
      <div class="kb-quick-actions">
        <div class="kb-quick-card" @click="$router.push('/knowledge/list')">
          <el-icon :size="28" color="#3370ff"><Opportunity /></el-icon>
          <span>知识库广场</span>
        </div>
        <div class="kb-quick-card" @click="$router.push('/knowledge/list')">
          <el-icon :size="28" color="#36b37e"><Document /></el-icon>
          <span>文档解读</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, computed } from 'vue'
import { Promotion, Microphone } from '@element-plus/icons-vue'
import { getKnowledgeBaseList, chatWithKnowledge, chatWithKnowledgeGlobal } from '@/api/modules/knowledge'
import { getAIModelSettings } from '@/api/modules/admin'
import { ElMessage } from 'element-plus'

interface ChatMessage {
  role: 'user' | 'assistant'
  content: string
  sources?: string[]
}

const chatInput = ref('')
const chatMode = ref<'fast' | 'deep'>('fast')
const selectedModel = ref('')
const selectedKb = ref('__all__')
const messages = ref<ChatMessage[]>([])
const sending = ref(false)
const messagesContainer = ref<HTMLElement>()

const kbList = ref<{ id: number; name: string }[]>([])
const availableModels = ref<{ id: string; name: string }[]>([])

const chatModeLabel = computed(() => chatMode.value === 'fast' ? '快速回答' : '深度思考')
const selectedModelName = computed(() => {
  const m = availableModels.value.find(x => x.id === selectedModel.value)
  return m?.name || (availableModels.value[0]?.name || '选择模型')
})
const selectedKbName = computed(() => {
  if (selectedKb.value === '__all__') return '全部知识库'
  const kb = kbList.value.find(x => String(x.id) === selectedKb.value)
  return kb?.name || '知识库1'
})

function handleModeChange(cmd: string) {
  chatMode.value = cmd as 'fast' | 'deep'
}
function handleModelChange(cmd: string) {
  selectedModel.value = cmd
}
function handleKbChange(cmd: string) {
  selectedKb.value = cmd
}

async function handleSend() {
  const text = chatInput.value.trim()
  if (!text || sending.value) return
  messages.value.push({ role: 'user', content: text })
  chatInput.value = ''
  sending.value = true
  await nextTick()
  scrollToBottom()

  try {
    let res: any
    if (selectedKb.value === '__all__') {
      res = await chatWithKnowledgeGlobal({ question: text })
    } else {
      const kbId = parseInt(selectedKb.value)
      res = await chatWithKnowledge(kbId, { question: text })
    }

    if (res.data?.answer) {
      const response: ChatMessage = {
        role: 'assistant',
        content: res.data.answer,
        sources: res.data.sources?.map((s: any) => s.documentTitle) || [],
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
      content: `错误: ${e.message || '请求失败，请检查AI模型配置'}`,
    })
  } finally {
    sending.value = false
    nextTick(() => scrollToBottom())
  }
}

function scrollToBottom() {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

async function loadModels() {
  try {
    const res: any = await getAIModelSettings()
    const providers = res.data?.providers || []
    const models: { id: string; name: string }[] = []
    for (const p of providers) {
      if (p.isActive && p.models?.length) {
        for (const m of p.models) {
          if (m.type === 'chat') {
            models.push({ id: `${p.id}_${m.id}`, name: m.name || m.id })
          }
        }
      }
    }
    availableModels.value = models
    if (models.length) selectedModel.value = models[0].id
  } catch { /* ignore */ }
}

async function loadKbList() {
  try {
    const res: any = await getKnowledgeBaseList({ page: 1, pageSize: 100 })
    kbList.value = (res.data?.list || []).map((kb: any) => ({ id: kb.id, name: kb.name }))
  } catch { /* ignore */ }
}

onMounted(() => {
  loadModels()
  loadKbList()
})
</script>

<style scoped>
.knowledge-home {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  min-height: 100%;
  padding: 40px 20px 20px;
  background: #fafbfc;
}
.kb-home-center {
  width: 100%;
  max-width: 680px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

/* Welcome */
.kb-welcome {
  text-align: center;
  margin-bottom: 32px;
}
.kb-welcome-tag {
  margin-bottom: 16px;
}
.kb-logo {
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: 4px;
  margin-bottom: 12px;
}
.kb-logo-text {
  font-size: 48px;
  font-weight: 700;
  color: var(--kx-text-primary);
  font-family: 'Arial Black', 'Helvetica', sans-serif;
  letter-spacing: -1px;
}
.kb-logo-sub {
  font-size: 18px;
  color: var(--kx-text-secondary);
  font-style: italic;
}
.kb-welcome-hint {
  font-size: 15px;
  color: var(--kx-text-secondary);
}

/* Chat input */
.kb-chat-input-area {
  width: 100%;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 16px;
  padding: 12px 16px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.06);
  margin-bottom: 24px;
}
.kb-chat-options {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}
.kb-option-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 12px;
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
}
.kb-chat-bar {
  display: flex;
  align-items: flex-end;
  gap: 8px;
}
.kb-chat-textarea {
  flex: 1;
}
.kb-chat-textarea :deep(.el-textarea__inner) {
  border: none;
  box-shadow: none;
  padding: 4px 0;
  resize: none;
}
.kb-chat-bar-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
  padding-bottom: 2px;
}

/* Chat messages */
.kb-chat-messages {
  width: 100%;
  max-height: 400px;
  overflow-y: auto;
  margin-bottom: 24px;
}
.kb-chat-msg {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}
.kb-chat-msg.user {
  flex-direction: row-reverse;
}
.kb-chat-msg.user .msg-body {
  align-items: flex-end;
}
.msg-body {
  display: flex;
  flex-direction: column;
  max-width: 80%;
}
.msg-content {
  padding: 10px 14px;
  border-radius: 12px;
  font-size: 14px;
  line-height: 1.6;
}
.kb-chat-msg.user .msg-content {
  background: var(--kx-primary);
  color: #fff;
  border-top-right-radius: 4px;
}
.kb-chat-msg.assistant .msg-content {
  background: #fff;
  border: 1px solid var(--kx-border);
  border-top-left-radius: 4px;
}
.msg-sources {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 6px;
  flex-wrap: wrap;
}
.source-label {
  font-size: 12px;
  color: var(--kx-text-placeholder);
}
.source-tag {
  cursor: pointer;
}
.typing-indicator {
  display: flex;
  gap: 4px;
  padding: 12px 18px;
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
.kb-quick-actions {
  display: flex;
  gap: 16px;
  width: 100%;
  justify-content: center;
}
.kb-quick-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 24px 40px;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 14px;
  color: var(--kx-text-secondary);
}
.kb-quick-card:hover {
  border-color: var(--kx-primary);
  box-shadow: 0 2px 12px rgba(51,112,255,0.1);
  color: var(--kx-primary);
}
</style>
