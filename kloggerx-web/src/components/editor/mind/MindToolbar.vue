<template>
  <div class="mind-toolbar">
    <div class="toolbar-group">
      <el-button size="small" @click="addChild" :disabled="!hasSelection"><el-icon><Plus /></el-icon>{{ $t('editor.mind.addChild') }}</el-button>
      <el-button size="small" @click="addSibling" :disabled="!canAddSibling"><el-icon><Right /></el-icon>{{ $t('editor.mind.addSibling') }}</el-button>
      <el-button size="small" @click="removeNode" :disabled="!canDelete"><el-icon><Delete /></el-icon>{{ $t('editor.mind.deleteNode') }}</el-button>
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-button size="small" @click="editNode" :disabled="!hasSelection"><el-icon><Edit /></el-icon>{{ $t('editor.mind.edit') }}</el-button>
      <el-color-picker v-model="nodeColor" size="small" :title="$t('editor.mind.nodeColor')" @change="onColorChange" />
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-select v-model="layoutType" size="small" :placeholder="$t('editor.mind.layout')" style="width: 100px" @change="onLayoutChange">
        <el-option :label="$t('editor.mind.expandRight')" value="right" />
        <el-option :label="$t('editor.mind.expandLeft')" value="left" />
        <el-option :label="$t('editor.mind.expandBoth')" value="both" />
      </el-select>
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-button size="small" @click="expandAll"><el-icon><FullScreen /></el-icon>{{ $t('editor.mind.expandAll') }}</el-button>
      <el-button size="small" @click="collapseAll"><el-icon><Minus /></el-icon>{{ $t('editor.mind.collapseAll') }}</el-button>
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-button size="small" @click="exportImage"><el-icon><Download /></el-icon>{{ $t('editor.mind.exportImage') }}</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const props = defineProps<{
  hasSelection: boolean
  isRoot: boolean
}>()

const emit = defineEmits(['action'])

const nodeColor = ref('#3370ff')
const layoutType = ref('right')

const canAddSibling = computed(() => props.hasSelection && !props.isRoot)
const canDelete = computed(() => props.hasSelection && !props.isRoot)

function emitAction(action: string, params?: any) {
  emit('action', { action, params })
}

function addChild() { emitAction('addChild') }
function addSibling() { emitAction('addSibling') }
function removeNode() { emitAction('removeNode') }
function editNode() { emitAction('editNode') }
function onColorChange(val: string) { emitAction('setColor', val) }
function onLayoutChange(val: string) { emitAction('setLayout', val) }
function expandAll() { emitAction('expandAll') }
function collapseAll() { emitAction('collapseAll') }
function exportImage() { emitAction('exportImage') }
</script>

<style scoped>
.mind-toolbar {
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
