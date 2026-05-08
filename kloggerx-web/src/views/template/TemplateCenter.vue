<template>
  <div class="template-center">
    <TemplateLibrary mode="page" @use="handleUseTemplate" />
  </div>
</template>

<script setup lang="ts">
import TemplateLibrary from '@/components/common/TemplateLibrary.vue'
import type { Template } from '@/api/modules/template'

/**
 * 处理模板使用事件
 * TemplateLibrary 组件内部已完成 API 调用和页面跳转，
 * 此处可用于额外的页面级逻辑（如埋点、刷新侧边栏等）
 */
function handleUseTemplate(_payload: { template: Template; title: string; parentId?: number }) {
  // 组件内部已处理创建和跳转，此处预留扩展
}
</script>

<style scoped>
.template-center {
  height: 100%;
  display: flex;
  flex-direction: column;
}
</style>
