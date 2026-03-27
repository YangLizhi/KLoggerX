<template>
  <div class="survey-toolbar">
    <div class="toolbar-group">
      <el-dropdown trigger="click" @command="addQuestion">
        <el-button size="small" type="primary">
          <el-icon><Plus /></el-icon>添加题目
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="single">单选题</el-dropdown-item>
            <el-dropdown-item command="multiple">多选题</el-dropdown-item>
            <el-dropdown-item command="text">填空题</el-dropdown-item>
            <el-dropdown-item command="rating">评分题</el-dropdown-item>
            <el-dropdown-item command="matrix">矩阵题</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-button size="small" @click="moveUp" :disabled="!canMoveUp"><el-icon><Top /></el-icon>上移</el-button>
      <el-button size="small" @click="moveDown" :disabled="!canMoveDown"><el-icon><Bottom /></el-icon>下移</el-button>
      <el-button size="small" @click="duplicateQuestion" :disabled="!hasSelection"><el-icon><CopyDocument /></el-icon>复制</el-button>
      <el-button size="small" @click="deleteQuestion" :disabled="!hasSelection"><el-icon><Delete /></el-icon>删除</el-button>
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-button size="small" @click="previewSurvey"><el-icon><View /></el-icon>预览</el-button>
      <el-button size="small" @click="showSettings"><el-icon><Setting /></el-icon>设置</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  questionCount: number
  selectedIndex: number
}>()

const emit = defineEmits(['action'])

const hasSelection = computed(() => props.selectedIndex >= 0)
const canMoveUp = computed(() => props.selectedIndex > 0)
const canMoveDown = computed(() => props.selectedIndex >= 0 && props.selectedIndex < props.questionCount - 1)

function emitAction(action: string, params?: any) {
  emit('action', { action, params })
}

function addQuestion(type: string) { emitAction('addQuestion', type) }
function moveUp() { emitAction('moveUp') }
function moveDown() { emitAction('moveDown') }
function duplicateQuestion() { emitAction('duplicateQuestion') }
function deleteQuestion() { emitAction('deleteQuestion') }
function previewSurvey() { emitAction('preview') }
function showSettings() { emitAction('settings') }
</script>

<style scoped>
.survey-toolbar {
  display: flex;
  align-items: center;
  padding: 6px 16px;
  border-bottom: 1px solid var(--kx-border, #e5e6eb);
  background: #fafafa;
  gap: 4px;
  flex-wrap: wrap;
}
.toolbar-group {
  display: flex;
  align-items: center;
  gap: 4px;
}
.toolbar-divider {
  width: 1px;
  height: 20px;
  background: var(--kx-border, #e5e6eb);
  margin: 0 8px;
}
</style>
