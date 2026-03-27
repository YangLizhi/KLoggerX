<template>
  <div class="sheet-toolbar">
    <div class="toolbar-group">
      <el-button size="small" @click="addRowAbove"><el-icon><Top /></el-icon>上方插入行</el-button>
      <el-button size="small" @click="addRowBelow"><el-icon><Bottom /></el-icon>下方插入行</el-button>
      <el-button size="small" @click="addColLeft"><el-icon><Back /></el-icon>左侧插入列</el-button>
      <el-button size="small" @click="addColRight"><el-icon><Right /></el-icon>右侧插入列</el-button>
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-button size="small" @click="deleteRow"><el-icon><Minus /></el-icon>删除行</el-button>
      <el-button size="small" @click="deleteCol"><el-icon><Minus /></el-icon>删除列</el-button>
      <el-button size="small" @click="clearCells"><el-icon><Delete /></el-icon>清空单元格</el-button>
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-select v-model="cellFormat" size="small" placeholder="格式" style="width: 100px">
        <el-option label="常规" value="normal" />
        <el-option label="数字" value="number" />
        <el-option label="货币" value="currency" />
        <el-option label="百分比" value="percent" />
        <el-option label="日期" value="date" />
      </el-select>
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-button size="small" :class="{ active: align === 'left' }" @click="align = 'left'">
        <el-icon><align-left /></el-icon>
      </el-button>
      <el-button size="small" :class="{ active: align === 'center' }" @click="align = 'center'">
        <el-icon><align-center /></el-icon>
      </el-button>
      <el-button size="small" :class="{ active: align === 'right' }" @click="align = 'right'">
        <el-icon><align-right /></el-icon>
      </el-button>
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-color-picker v-model="bgColor" size="small" title="背景色" />
      <el-color-picker v-model="textColor" size="small" title="文字颜色" />
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-button size="small" @click="mergeCells"><el-icon><Grid /></el-icon>合并单元格</el-button>
      <el-button size="small" @click="freezeRow"><el-icon><Lock /></el-icon>冻结行</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const emit = defineEmits(['action'])

const cellFormat = ref('normal')
const align = ref('left')
const bgColor = ref('#ffffff')
const textColor = ref('#000000')

function emitAction(action: string, params?: any) {
  emit('action', { action, params })
}

function addRowAbove() { emitAction('addRowAbove') }
function addRowBelow() { emitAction('addRowBelow') }
function addColLeft() { emitAction('addColLeft') }
function addColRight() { emitAction('addColRight') }
function deleteRow() { emitAction('deleteRow') }
function deleteCol() { emitAction('deleteCol') }
function clearCells() { emitAction('clearCells') }
function mergeCells() { emitAction('mergeCells') }
function freezeRow() { emitAction('freezeRow') }
</script>

<style scoped>
.sheet-toolbar {
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
.sheet-toolbar .el-button.active {
  background: #dee0e3;
  color: #3370ff;
}
</style>
