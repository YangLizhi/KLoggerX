<template>
  <div class="shared-conversation">
    <!-- Header -->
    <div class="shared-header">
      <div class="shared-logo">
        <el-icon :size="24" color="#3370ff"><ChatDotRound /></el-icon>
        <span class="shared-brand">KLoggerX 知识库助手</span>
      </div>
      <div v-if="conversation" class="shared-title">{{ conversation.title }}</div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="shared-loading">
      <el-icon class="is-loading" :size="32"><Loading /></el-icon>
      <span>加载中...</span>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="shared-error">
      <el-icon :size="48" color="#f56c6c"><CircleClose /></el-icon>
      <div class="error-title">{{ errorTitle }}</div>
      <div class="error-desc">{{ errorDesc }}</div>
      <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
    </div>

    <!-- Messages -->
    <div v-else-if="messages.length" class="shared-messages">
      <ChatMessage
        v-for="(msg, i) in messages"
        :key="i"
        :message="msg"
      />
    </div>

    <!-- Footer -->
    <div v-if="!loading && !error" class="shared-footer">
      <span>此对话由 KLoggerX 知识库助手生成</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ChatDotRound, Loading, CircleClose } from '@element-plus/icons-vue'
import { getSharedConversation } from '@/api/modules/knowledge'
import type { Conversation, ConversationMessage } from '@/api/modules/knowledge'
import ChatMessage from '@/components/knowledge/ChatMessage.vue'
import type { ChatMessageData } from '@/components/knowledge/ChatMessage.vue'

const route = useRoute()
const loading = ref(true)
const error = ref(false)
const errorTitle = ref('')
const errorDesc = ref('')
const conversation = ref<Conversation | null>(null)
const messages = ref<ChatMessageData[]>([])

onMounted(async () => {
  const token = route.params.token as string
  if (!token) {
    error.value = true
    errorTitle.value = '无效链接'
    errorDesc.value = '分享链接格式不正确'
    loading.value = false
    return
  }

  try {
    const res: any = await getSharedConversation(token)
    if (res.data) {
      conversation.value = res.data.conversation
      messages.value = (res.data.messages || []).map((m: ConversationMessage) => ({
        id: m.id,
        role: m.role,
        content: m.content,
        sources: m.sources || [],
        isStreaming: false,
        suggestions: [],
        feedback: null
      }))
    }
  } catch (e: any) {
    error.value = true
    const status = e?.response?.status
    if (status === 404) {
      errorTitle.value = '链接不存在'
      errorDesc.value = '该分享链接不存在或已被取消'
    } else if (status === 410) {
      errorTitle.value = '链接已过期'
      errorDesc.value = '该分享链接已过期，请联系分享者重新分享'
    } else {
      errorTitle.value = '加载失败'
      errorDesc.value = '无法加载分享内容，请稍后重试'
    }
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.shared-conversation {
  min-height: 100vh;
  background: linear-gradient(180deg, #f0f5ff 0%, #fafbfc 100%);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px 20px;
}

.shared-header {
  text-align: center;
  margin-bottom: 32px;
}

.shared-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  margin-bottom: 12px;
}

.shared-brand {
  font-size: 18px;
  font-weight: 600;
  color: var(--kx-text-primary, #303133);
}

.shared-title {
  font-size: 20px;
  font-weight: 700;
  color: var(--kx-text-primary, #303133);
}

.shared-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 60px 0;
  color: #909399;
  font-size: 14px;
}

.shared-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 60px 0;
  text-align: center;
}

.error-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--kx-text-primary, #303133);
}

.error-desc {
  font-size: 14px;
  color: #909399;
  margin-bottom: 12px;
}

.shared-messages {
  width: 100%;
  max-width: 800px;
}

.shared-footer {
  margin-top: 40px;
  padding: 16px;
  text-align: center;
  font-size: 13px;
  color: #909399;
  border-top: 1px solid #ebeef5;
  width: 100%;
  max-width: 800px;
}
</style>
