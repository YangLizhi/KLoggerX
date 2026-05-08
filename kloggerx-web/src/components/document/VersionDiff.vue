<template>
  <div class="version-diff">
    <div class="diff-header">
      <div class="diff-stats">
        <span class="stat-added">+ {{ stats.added }}</span>
        <span class="stat-deleted">- {{ stats.deleted }}</span>
      </div>
      <div class="version-labels">
        <span class="label-old">版本 {{ oldVersion }}</span>
        <span class="arrow">→</span>
        <span class="label-new">版本 {{ newVersion }}</span>
      </div>
    </div>

    <div class="diff-body" ref="diffBody">
      <div
        v-for="(line, idx) in displayLines"
        :key="idx"
        :class="['diff-line', `diff-${line.type}`]"
      >
        <span class="line-number old-num">{{ line.old_line || '' }}</span>
        <span class="line-number new-num">{{ line.new_line || '' }}</span>
        <span class="line-prefix">{{ getPrefix(line.type) }}</span>
        <span class="line-content">{{ line.content }}</span>
      </div>
      <div v-if="lines.length > maxLines" class="diff-truncated">
        <el-button size="small" @click="showAll = true" v-if="!showAll">
          显示全部 {{ lines.length }} 行（当前显示 {{ maxLines }} 行）
        </el-button>
      </div>
    </div>

    <div v-if="!lines.length" class="diff-empty">两个版本内容相同</div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { DiffLine } from '@/api/modules/document'

const props = defineProps<{
  oldVersion: string
  newVersion: string
  lines: DiffLine[]
  stats: { added: number; deleted: number; changed: number }
}>()

const maxLines = 1000
const showAll = ref(false)

const displayLines = computed(() => {
  if (showAll.value || props.lines.length <= maxLines) {
    return props.lines
  }
  return props.lines.slice(0, maxLines)
})

function getPrefix(type: string): string {
  switch (type) {
    case 'add': return '+'
    case 'delete': return '-'
    default: return ' '
  }
}
</script>

<style scoped>
.version-diff {
  font-family: 'Menlo', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  overflow: hidden;
}

.diff-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
}

.diff-stats {
  display: flex;
  gap: 12px;
}

.stat-added {
  color: #67c23a;
  font-weight: 600;
}

.stat-deleted {
  color: #f56c6c;
  font-weight: 600;
}

.version-labels {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #606266;
}

.arrow {
  color: #909399;
}

.diff-body {
  max-height: 500px;
  overflow-y: auto;
}

.diff-line {
  display: flex;
  align-items: stretch;
  min-height: 22px;
  line-height: 22px;
}

.diff-add {
  background-color: #e6ffec;
}

.diff-delete {
  background-color: #ffebe9;
}

.diff-equal {
  background-color: transparent;
}

.line-number {
  display: inline-block;
  width: 40px;
  min-width: 40px;
  text-align: right;
  padding-right: 8px;
  color: #909399;
  user-select: none;
  border-right: 1px solid #e4e7ed;
}

.old-num {
  border-right: none;
}

.line-prefix {
  display: inline-block;
  width: 16px;
  min-width: 16px;
  text-align: center;
  color: #606266;
  user-select: none;
}

.line-content {
  flex: 1;
  padding-left: 4px;
  white-space: pre-wrap;
  word-break: break-all;
}

.diff-truncated {
  padding: 12px;
  text-align: center;
  background: #fafafa;
  border-top: 1px solid #e4e7ed;
}

.diff-empty {
  padding: 40px;
  text-align: center;
  color: #909399;
}
</style>
