<template>
  <transition name="slide-down">
    <div v-if="!isOnline" class="network-offline-bar">
      <el-icon><Warning /></el-icon>
      <span>{{ $t('network.offline') }}</span>
      <el-button size="small" text @click="retry">{{ $t('common.retry') }}</el-button>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { Warning } from '@element-plus/icons-vue'

const isOnline = ref(navigator.onLine)

function handleOnline() { isOnline.value = true }
function handleOffline() { isOnline.value = false }
function retry() { window.location.reload() }

onMounted(() => {
  window.addEventListener('online', handleOnline)
  window.addEventListener('offline', handleOffline)
})
onBeforeUnmount(() => {
  window.removeEventListener('online', handleOnline)
  window.removeEventListener('offline', handleOffline)
})
</script>

<style scoped>
.network-offline-bar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px 16px;
  background: #fef0f0;
  border-bottom: 1px solid #fde2e2;
  color: #f56c6c;
  font-size: 14px;
}

.slide-down-enter-active,
.slide-down-leave-active {
  transition: all 0.3s ease;
}
.slide-down-enter-from,
.slide-down-leave-to {
  transform: translateY(-100%);
  opacity: 0;
}
</style>
