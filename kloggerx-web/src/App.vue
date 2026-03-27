<template>
  <router-view />
</template>

<script setup lang="ts">
import { onMounted, onBeforeUnmount } from 'vue'

// 全局屏蔽浏览器默认右键菜单
function preventDefaultContextMenu(e: MouseEvent) {
  e.preventDefault()
}

onMounted(() => {
  document.addEventListener('contextmenu', preventDefaultContextMenu)
})

onBeforeUnmount(() => {
  document.removeEventListener('contextmenu', preventDefaultContextMenu)
})
</script>
