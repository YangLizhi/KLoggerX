<template>
  <div class="survey-toolbar">
    <div class="toolbar-group">
      <el-dropdown trigger="click" @command="addQuestion">
        <el-button size="small" type="primary">
          <el-icon><Plus /></el-icon>{{ $t('editor.survey.addQuestion') }}
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="single">{{ $t('editor.survey.singleChoice') }}</el-dropdown-item>
            <el-dropdown-item command="multiple">{{ $t('editor.survey.multipleChoice') }}</el-dropdown-item>
            <el-dropdown-item command="text">{{ $t('editor.survey.textQuestion') }}</el-dropdown-item>
            <el-dropdown-item command="rating">{{ $t('editor.survey.ratingQuestion') }}</el-dropdown-item>
            <el-dropdown-item command="matrix">{{ $t('editor.survey.matrixQuestion') }}</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-button size="small" @click="moveUp" :disabled="!canMoveUp"><el-icon><Top /></el-icon>{{ $t('editor.survey.moveUp') }}</el-button>
      <el-button size="small" @click="moveDown" :disabled="!canMoveDown"><el-icon><Bottom /></el-icon>{{ $t('editor.survey.moveDown') }}</el-button>
      <el-button size="small" @click="duplicateQuestion" :disabled="!hasSelection"><el-icon><CopyDocument /></el-icon>{{ $t('editor.survey.duplicate') }}</el-button>
      <el-button size="small" @click="deleteQuestion" :disabled="!hasSelection"><el-icon><Delete /></el-icon>{{ $t('editor.survey.deleteQuestion') }}</el-button>
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-button size="small" @click="previewSurvey"><el-icon><View /></el-icon>{{ $t('editor.survey.preview') }}</el-button>
      <el-button size="small" @click="showSettings"><el-icon><Setting /></el-icon>{{ $t('editor.survey.settings') }}</el-button>
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
