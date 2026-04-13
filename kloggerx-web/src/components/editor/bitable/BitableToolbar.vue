<template>
  <div class="bitable-toolbar">
    <div class="toolbar-group">
      <el-button size="small" @click="addRecord"><el-icon><Plus /></el-icon>添加记录</el-button>
      <el-button size="small" @click="deleteRecord" :disabled="!hasSelection"><el-icon><Delete /></el-icon>删除记录</el-button>
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-dropdown trigger="click" @command="addField">
        <el-button size="small">
          <el-icon><Grid /></el-icon>添加字段<el-icon class="el-icon--right"><ArrowDown /></el-icon>
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="text">文本</el-dropdown-item>
            <el-dropdown-item command="number">数字</el-dropdown-item>
            <el-dropdown-item command="select">单选</el-dropdown-item>
            <el-dropdown-item command="multiSelect">多选</el-dropdown-item>
            <el-dropdown-item command="date">日期</el-dropdown-item>
            <el-dropdown-item command="person">人员</el-dropdown-item>
            <el-dropdown-item command="attachment">附件</el-dropdown-item>
            <el-dropdown-item command="link">链接</el-dropdown-item>
            <el-dropdown-item command="checkbox">复选框</el-dropdown-item>
            <el-dropdown-item command="rating">评分</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
      <el-button size="small" @click="editField" :disabled="!selectedField"><el-icon><Edit /></el-icon>编辑字段</el-button>
      <el-button size="small" @click="deleteField" :disabled="!selectedField"><el-icon><Delete /></el-icon>删除字段</el-button>
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-button size="small" @click="addView"><el-icon><View /></el-icon>新建视图</el-button>
      <el-button size="small" @click="showFilter"><el-icon><Filter /></el-icon>筛选</el-button>
      <el-button size="small" @click="showSort"><el-icon><Sort /></el-icon>排序</el-button>
      <el-button size="small" @click="showGroup"><el-icon><Menu /></el-icon>分组</el-button>
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-button size="small" @click="exportData"><el-icon><Download /></el-icon>导出</el-button>
      <el-button size="small" @click="importData"><el-icon><Upload /></el-icon>导入</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
// import { computed } from 'vue'

const props = defineProps<{
  hasSelection: boolean
  selectedField: string | null
}>()

const emit = defineEmits(['action'])

function emitAction(action: string, params?: any) {
  emit('action', { action, params })
}

function addRecord() { emitAction('addRecord') }
function deleteRecord() { emitAction('deleteRecord') }
function addField(type: string) { emitAction('addField', type) }
function editField() { emitAction('editField') }
function deleteField() { emitAction('deleteField') }
function addView() { emitAction('addView') }
function showFilter() { emitAction('showFilter') }
function showSort() { emitAction('showSort') }
function showGroup() { emitAction('showGroup') }
function exportData() { emitAction('export') }
function importData() { emitAction('import') }
</script>

<style scoped>
.bitable-toolbar {
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
