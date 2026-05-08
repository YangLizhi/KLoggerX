<template>
  <div class="conversation-list">
    <!-- New Conversation Button -->
    <div class="conv-header">
      <el-button type="primary" class="new-conv-btn" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        {{ $t('knowledge.chat.newConversation') }}
      </el-button>
    </div>

    <!-- Search -->
    <div class="conv-search">
      <el-input
        v-model="searchQuery"
                :placeholder="$t('knowledge.chat.searchConv')"
        prefix-icon="Search"
        clearable
        size="small"
      />
    </div>

    <!-- Conversation List -->
    <div class="conv-list-body" ref="listRef" @scroll="handleScroll">
      <template v-if="filteredGroups.length">
        <div v-for="group in filteredGroups" :key="group.label" class="conv-group">
          <div class="conv-group-label">{{ group.label }}</div>
          <div
            v-for="conv in group.items"
            :key="conv.id"
            class="conv-item"
            :class="{ active: conv.id === currentId, pinned: conv.isPinned }"
            @click="handleSelect(conv)"
          >
            <div class="conv-item-content">
              <el-icon v-if="conv.isPinned" class="pin-icon"><Top /></el-icon>
              <template v-if="editingId === conv.id">
                <el-input
                  v-model="editTitle"
                  size="small"
                  class="conv-edit-input"
                  @keydown.enter="confirmRename(conv)"
                  @keydown.escape="cancelRename"
                  @blur="confirmRename(conv)"
                  ref="editInputRef"
                />
              </template>
              <span v-else class="conv-title">{{ conv.title || $t('knowledge.chat.newChat') }}</span>
            </div>
            <div class="conv-item-actions" v-show="editingId !== conv.id">
              <el-icon class="action-icon" @click.stop="startRename(conv)"><Edit /></el-icon>
              <el-dropdown trigger="click" @command="(cmd: string) => handleAction(cmd, conv)" @click.stop>
                <el-icon class="action-icon"><MoreFilled /></el-icon>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item :command="conv.isPinned ? 'unpin' : 'pin'">
                      <el-icon><Top /></el-icon>
                      {{ conv.isPinned ? $t('knowledge.chat.unpin') : $t('knowledge.chat.pinned') }}
                    </el-dropdown-item>
                    <el-dropdown-item command="delete" divided>
                      <el-icon color="#f56c6c"><Delete /></el-icon>
                      <span style="color: #f56c6c">{{ $t('common.delete') }}</span>
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
        </div>
      </template>

      <!-- Empty State -->
      <div v-else-if="!loading" class="conv-empty">
        <el-icon :size="48" color="#c0c4cc"><ChatDotRound /></el-icon>
        <p>{{ $t('knowledge.chat.noConversations') }}</p>
        <p class="conv-empty-hint">{{ $t('knowledge.chat.startNewConv') }}</p>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="conv-loading">
        <el-icon class="is-loading"><Loading /></el-icon>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Plus, Edit, Delete, MoreFilled, Top, ChatDotRound, Loading } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

const { t } = useI18n()
import {
  getConversations,
  updateConversation,
  deleteConversation,
  type Conversation
} from '@/api/modules/knowledge'

const props = defineProps<{
  currentId: number | null
  knowledgeBaseId: number | null
}>()

const emit = defineEmits<{
  select: [conversation: Conversation]
  create: []
}>()

const conversations = ref<Conversation[]>([])
const loading = ref(false)
const searchQuery = ref('')
const editingId = ref<number | null>(null)
const editTitle = ref('')
const editInputRef = ref()
const listRef = ref<HTMLElement>()

const page = ref(1)
const pageSize = 20
const total = ref(0)
const hasMore = computed(() => conversations.value.length < total.value)

// Group conversations by time
interface ConvGroup {
  label: string
  items: Conversation[]
}

const filteredGroups = computed<ConvGroup[]>(() => {
  let list = conversations.value
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter(c => (c.title || t('knowledge.chat.newChat')).toLowerCase().includes(q))
  }

  const pinned: Conversation[] = []
  const today: Conversation[] = []
  const yesterday: Conversation[] = []
  const earlier: Conversation[] = []

  const now = new Date()
  const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime()
  const yesterdayStart = todayStart - 86400000

  for (const c of list) {
    if (c.isPinned) {
      pinned.push(c)
      continue
    }
    const t = new Date(c.updatedAt).getTime()
    if (t >= todayStart) today.push(c)
    else if (t >= yesterdayStart) yesterday.push(c)
    else earlier.push(c)
  }

  const groups: ConvGroup[] = []
  if (pinned.length) groups.push({ label: t('knowledge.chat.pinned'), items: pinned })
  if (today.length) groups.push({ label: t('knowledge.chat.today'), items: today })
  if (yesterday.length) groups.push({ label: t('knowledge.chat.yesterday'), items: yesterday })
  if (earlier.length) groups.push({ label: t('knowledge.chat.earlier'), items: earlier })
  return groups
})

async function loadConversations(reset = false) {
  if (reset) {
    page.value = 1
    conversations.value = []
  }
  loading.value = true
  try {
    const params: any = { page: page.value, pageSize }
    if (props.knowledgeBaseId) {
      params.knowledgeBaseId = props.knowledgeBaseId
    }
    const res: any = await getConversations(params)
    const list = res.data?.list || []
    total.value = res.data?.total || 0
    if (reset) {
      conversations.value = list
    } else {
      conversations.value.push(...list)
    }
  } catch (e) {
    console.error('Failed to load conversations:', e)
  } finally {
    loading.value = false
  }
}

function handleScroll() {
  if (!listRef.value || loading.value || !hasMore.value) return
  const { scrollTop, scrollHeight, clientHeight } = listRef.value
  if (scrollHeight - scrollTop - clientHeight < 60) {
    page.value++
    loadConversations()
  }
}

function handleCreate() {
  emit('create')
}

function handleSelect(conv: Conversation) {
  if (editingId.value === conv.id) return
  emit('select', conv)
}

function startRename(conv: Conversation) {
  editingId.value = conv.id
  editTitle.value = conv.title || ''
  nextTick(() => {
    const inputs = editInputRef.value
    if (Array.isArray(inputs) && inputs.length) {
      inputs[0].focus()
    }
  })
}

async function confirmRename(conv: Conversation) {
  if (editingId.value !== conv.id) return
  const newTitle = editTitle.value.trim()
  editingId.value = null
  if (!newTitle || newTitle === conv.title) return
  try {
    await updateConversation(conv.id, { title: newTitle })
    conv.title = newTitle
  } catch (e) {
    ElMessage.error(t('knowledge.chat.renameFailed'))
  }
}

function cancelRename() {
  editingId.value = null
}

async function handleAction(cmd: string, conv: Conversation) {
  if (cmd === 'pin' || cmd === 'unpin') {
    try {
      await updateConversation(conv.id, { isPinned: cmd === 'pin' })
      conv.isPinned = cmd === 'pin'
    } catch (e) {
      ElMessage.error(t('common.operationFailed'))
    }
  } else if (cmd === 'delete') {
    try {
      await ElMessageBox.confirm(t('knowledge.chat.deleteConvConfirm'), t('knowledge.chat.deleteConvTitle'), {
        confirmButtonText: t('common.delete'),
        cancelButtonText: t('common.cancel'),
        type: 'warning',
      })
      await deleteConversation(conv.id)
      conversations.value = conversations.value.filter(c => c.id !== conv.id)
      total.value--
      if (conv.id === props.currentId) {
        emit('select', null as any)
      }
      ElMessage.success('已删除')
    } catch (e: any) {
      if (e !== 'cancel' && e?.message !== 'cancel') {
        ElMessage.error('删除失败')
      }
    }
  }
}

// Expose refresh method
function refresh() {
  loadConversations(true)
}

// Add new conversation to top of list
function addConversation(conv: Conversation) {
  conversations.value.unshift(conv)
  total.value++
}

defineExpose({ refresh, addConversation })

watch(() => props.knowledgeBaseId, () => {
  loadConversations(true)
})

onMounted(() => {
  loadConversations(true)
})
</script>

<style scoped>
.conversation-list {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #f7f8fa;
}

.conv-header {
  padding: 16px 12px 8px;
}

.new-conv-btn {
  width: 100%;
  border-radius: 8px;
  font-weight: 500;
}

.conv-search {
  padding: 4px 12px 8px;
}

.conv-list-body {
  flex: 1;
  overflow-y: auto;
  padding: 0 8px 12px;
}

.conv-group {
  margin-bottom: 4px;
}

.conv-group-label {
  font-size: 12px;
  font-weight: 500;
  color: #909399;
  padding: 8px 8px 4px;
  user-select: none;
}

.conv-item {
  display: flex;
  align-items: center;
  padding: 10px 10px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.15s;
  position: relative;
  gap: 4px;
}

.conv-item:hover {
  background: #edf2fc;
}

.conv-item.active {
  background: #e1ecff;
}

.conv-item.active .conv-title {
  font-weight: 500;
  color: #3370ff;
}

.conv-item-content {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 4px;
}

.pin-icon {
  font-size: 12px;
  color: #3370ff;
  flex-shrink: 0;
}

.conv-title {
  font-size: 13px;
  color: #303133;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.conv-edit-input {
  flex: 1;
}

.conv-item-actions {
  display: none;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.conv-item:hover .conv-item-actions {
  display: flex;
}

.action-icon {
  font-size: 14px;
  color: #909399;
  padding: 4px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s;
}

.action-icon:hover {
  color: #3370ff;
  background: #e1ecff;
}

.conv-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #909399;
}

.conv-empty p {
  margin: 8px 0 0;
  font-size: 14px;
}

.conv-empty-hint {
  font-size: 12px !important;
  color: #c0c4cc;
}

.conv-loading {
  display: flex;
  justify-content: center;
  padding: 12px;
  color: #909399;
}
</style>
