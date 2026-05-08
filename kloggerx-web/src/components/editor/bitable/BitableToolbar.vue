<template>
  <div class="bitable-toolbar">
    <div class="toolbar-group">
      <el-button size="small" @click="addRecord"><el-icon><Plus /></el-icon>{{ $t('editor.bitable.addRecord') }}</el-button>
      <el-button size="small" @click="deleteRecord" :disabled="!hasSelection"><el-icon><Delete /></el-icon>{{ $t('editor.bitable.deleteRecord') }}</el-button>
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-dropdown trigger="click" @command="addField">
        <el-button size="small">
          <el-icon><Grid /></el-icon>{{ $t('editor.bitable.addField') }}<el-icon class="el-icon--right"><ArrowDown /></el-icon>
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="text">{{ $t('editor.bitable.text') }}</el-dropdown-item>
            <el-dropdown-item command="number">{{ $t('editor.bitable.number') }}</el-dropdown-item>
            <el-dropdown-item command="select">{{ $t('editor.bitable.select') }}</el-dropdown-item>
            <el-dropdown-item command="multiSelect">{{ $t('editor.bitable.multiSelect') }}</el-dropdown-item>
            <el-dropdown-item command="date">{{ $t('editor.bitable.date') }}</el-dropdown-item>
            <el-dropdown-item command="person">{{ $t('editor.bitable.person') }}</el-dropdown-item>
            <el-dropdown-item command="attachment">{{ $t('editor.bitable.attachment') }}</el-dropdown-item>
            <el-dropdown-item command="link">{{ $t('editor.bitable.link') }}</el-dropdown-item>
            <el-dropdown-item command="checkbox">{{ $t('editor.bitable.checkbox') }}</el-dropdown-item>
            <el-dropdown-item command="rating">{{ $t('editor.bitable.rating') }}</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
      <el-button size="small" @click="editField" :disabled="!selectedField"><el-icon><Edit /></el-icon>{{ $t('editor.bitable.editField') }}</el-button>
      <el-button size="small" @click="deleteField" :disabled="!selectedField"><el-icon><Delete /></el-icon>{{ $t('editor.bitable.deleteField') }}</el-button>
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-button size="small" @click="addView"><el-icon><View /></el-icon>{{ $t('editor.bitable.newView') }}</el-button>
      <el-button size="small" @click="showFilter"><el-icon><Filter /></el-icon>{{ $t('editor.bitable.filter') }}</el-button>
      <el-button size="small" @click="showSort"><el-icon><Sort /></el-icon>{{ $t('editor.bitable.sort') }}</el-button>
      <el-button size="small" @click="showGroup"><el-icon><Menu /></el-icon>{{ $t('editor.bitable.group') }}</el-button>
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-button size="small" @click="exportData"><el-icon><Download /></el-icon>{{ $t('editor.bitable.export') }}</el-button>
      <el-button size="small" @click="importData"><el-icon><Upload /></el-icon>{{ $t('editor.bitable.import') }}</el-button>
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
