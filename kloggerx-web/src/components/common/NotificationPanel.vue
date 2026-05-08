<template>
  <el-popover placement="bottom-end" :width="380" trigger="click" @show="onPopoverShow">
    <template #reference>
      <el-badge :value="unreadCount" :hidden="unreadCount === 0" :max="99" class="notification-badge">
        <el-icon class="topbar-icon"><Bell /></el-icon>
      </el-badge>
    </template>

    <div class="notification-panel">
      <div class="panel-header">
        <span class="panel-title">{{ $t('notification.title') }}</span>
        <el-button type="primary" link size="small" @click="handleMarkAllRead" :disabled="unreadCount === 0">
          {{ $t('notification.markAllRead') }}
        </el-button>
      </div>

      <div class="panel-tabs">
        <span class="panel-tab" :class="{ active: activeTab === 'all' }" @click="activeTab = 'all'">{{ $t('common.all') }}</span>
        <span class="panel-tab" :class="{ active: activeTab === 'unread' }" @click="activeTab = 'unread'">{{ $t('notification.unread', { count: unreadCount }) }}</span>
      </div>

      <div class="notification-list" v-loading="loading">
        <template v-if="filteredNotifications.length">
          <div
            v-for="item in filteredNotifications"
            :key="item.id"
            class="notification-item"
            :class="{ unread: !item.isRead }"
            @click="handleClick(item)"
          >
            <div class="notif-icon" :class="item.type">
              <el-icon><component :is="getNotifIcon(item.type)" /></el-icon>
            </div>
            <div class="notif-content">
              <p class="notif-title">{{ item.title }}</p>
              <p class="notif-message">{{ item.content }}</p>
              <div class="notif-meta">
                <span class="notif-time">{{ formatTime(item.createdAt) }}</span>
                <span v-if="item.docTitle" class="notif-doc">{{ item.docTitle }}</span>
              </div>
            </div>
          </div>
        </template>
        <el-empty v-else :description="$t('notification.empty')" :image-size="60" />
      </div>
    </div>
  </el-popover>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { getNotifications, markNotificationRead, markAllNotificationsRead } from '@/api/modules/collaborate'
import { Bell, ChatDotRound, Share, UserFilled, InfoFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import type { Notification } from '@/types'

const router = useRouter()
const { t } = useI18n()

const notifications = ref<Notification[]>([])
const loading = ref(false)
const activeTab = ref<'all' | 'unread'>('all')
let refreshTimer: ReturnType<typeof setInterval> | null = null

const unreadCount = computed(() => notifications.value.filter(n => !n.isRead).length)

const filteredNotifications = computed(() => {
  if (activeTab.value === 'unread') {
    return notifications.value.filter(n => !n.isRead)
  }
  return notifications.value
})

function getNotifIcon(type: string): any {
  const icons: Record<string, any> = {
    comment: ChatDotRound,
    mention: ChatDotRound,
    share: Share,
    permission: UserFilled,
    approval: UserFilled,
    system: InfoFilled,
  }
  return icons[type] || Bell
}

function formatTime(time: string): string {
  if (!time) return ''
  const date = new Date(time)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return t('notification.justNow')
  if (minutes < 60) return t('notification.minutesAgo', { count: minutes })
  if (hours < 24) return t('notification.hoursAgo', { count: hours })
  if (days === 1) return t('notification.yesterday')
  if (days < 7) return t('notification.daysAgo', { count: days })
  return date.toLocaleDateString()
}

async function fetchNotifications() {
  loading.value = true
  try {
    const res: any = await getNotifications({ page: 1, pageSize: 50 })
    notifications.value = res.data?.list || []
  } catch { /* ignore */ }
  finally { loading.value = false }
}

function onPopoverShow() {
  fetchNotifications()
}

async function handleClick(item: Notification) {
  if (!item.isRead) {
    try {
      await markNotificationRead(item.id)
      item.isRead = true
    } catch { /* ignore */ }
  }
  if (item.documentId) {
    window.open(`/doc/${item.documentId}`, '_blank')
  }
}

async function handleMarkAllRead() {
  try {
    await markAllNotificationsRead()
    notifications.value.forEach(n => n.isRead = true)
    ElMessage.success(t('notification.allMarkedRead'))
  } catch {
    ElMessage.error(t('common.failed'))
  }
}

onMounted(() => {
  fetchNotifications()
  refreshTimer = setInterval(fetchNotifications, 30000)
})

onBeforeUnmount(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
})
</script>

<style scoped>
.notification-badge {
  line-height: 1;
}
.topbar-icon {
  font-size: 20px;
  cursor: pointer;
  color: var(--kx-text-secondary);
}
.topbar-icon:hover {
  color: var(--kx-primary);
}

.notification-panel {
  margin: -12px;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px 8px;
  border-bottom: 1px solid var(--kx-border);
}
.panel-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--kx-text-primary);
}

.panel-tabs {
  display: flex;
  gap: 4px;
  padding: 8px 16px;
  border-bottom: 1px solid var(--kx-border);
}
.panel-tab {
  padding: 4px 12px;
  border-radius: 14px;
  font-size: 13px;
  cursor: pointer;
  color: var(--kx-text-secondary);
  transition: all 0.2s;
  user-select: none;
}
.panel-tab:hover {
  background: rgba(0, 0, 0, 0.04);
}
.panel-tab.active {
  background: rgba(51, 112, 255, 0.1);
  color: var(--kx-primary);
  font-weight: 500;
}

.notification-list {
  max-height: 400px;
  overflow-y: auto;
  min-height: 100px;
}

.notification-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px 16px;
  cursor: pointer;
  transition: background 0.15s;
  position: relative;
}
.notification-item:hover {
  background: rgba(0, 0, 0, 0.03);
}
.notification-item.unread {
  background: rgba(51, 112, 255, 0.04);
}
.notification-item.unread::before {
  content: '';
  position: absolute;
  left: 0;
  top: 8px;
  bottom: 8px;
  width: 3px;
  border-radius: 0 3px 3px 0;
  background: var(--kx-primary, #3370ff);
}
.notification-item.unread:hover {
  background: rgba(51, 112, 255, 0.08);
}

.notif-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  flex-shrink: 0;
  font-size: 16px;
  color: #fff;
  background: #8592a6;
}
.notif-icon.comment,
.notif-icon.mention {
  background: #3370ff;
}
.notif-icon.share {
  background: #36b37e;
}
.notif-icon.permission,
.notif-icon.approval {
  background: #9254de;
}
.notif-icon.system {
  background: #8592a6;
}

.notif-content {
  flex: 1;
  min-width: 0;
}
.notif-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--kx-text-primary);
  margin: 0 0 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.notif-message {
  font-size: 12px;
  color: var(--kx-text-secondary);
  margin: 0 0 4px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.4;
}
.notif-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--kx-text-placeholder);
}
.notif-time {
  white-space: nowrap;
}
.notif-doc {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--kx-primary);
  max-width: 160px;
}
.notif-doc::before {
  content: '·';
  margin-right: 6px;
  color: var(--kx-text-placeholder);
}
</style>
