<template>
  <div class="sheet-editor">
    <SheetToolbar @action="handleToolbarAction" />
    <div class="sheet-container">
      <table class="sheet-table">
        <thead>
          <tr>
            <th class="row-header"></th>
            <th v-for="(_, ci) in (data[0] || [])" :key="ci" class="col-header">
              {{ getColLabel(ci) }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, ri) in data" :key="ri">
            <td class="row-header">{{ ri + 1 }}</td>
            <td
              v-for="(cell, ci) in row"
              :key="ci"
              class="sheet-cell"
              :class="{ active: activeCell?.r === ri && activeCell?.c === ci }"
              :style="getCellStyle(ri, ci)"
              @click="activeCell = { r: ri, c: ci }"
            >
              <input
                class="cell-input"
                :value="cell"
                @input="onCellInput(ri, ci, ($event.target as HTMLInputElement).value)"
                @focus="activeCell = { r: ri, c: ci }"
              />
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import SheetToolbar from './SheetToolbar.vue'

interface CellStyle {
  bgColor?: string
  textColor?: string
  align?: 'left' | 'center' | 'right'
  format?: string
}

const props = defineProps<{
  documentId: number
  content: string
}>()

const emit = defineEmits<{
  (e: 'save', content: string): void
}>()

interface SheetData {
  cells: string[][]
  styles: Record<string, CellStyle>
  mergedCells?: { start: { r: number; c: number }; end: { r: number; c: number } }[]
  frozenRows?: number
}

const data = ref<string[][]>([])
const styles = ref<Record<string, CellStyle>>({})
const activeCell = ref<{ r: number; c: number } | null>(null)
let saveTimer: ReturnType<typeof setTimeout> | null = null

function initData() {
  if (props.content) {
    try {
      const parsed = JSON.parse(props.content)
      if (Array.isArray(parsed)) {
        // Old format: just cells array
        data.value = parsed
        styles.value = {}
      } else if (parsed.cells) {
        // New format: cells + styles
        data.value = parsed.cells
        styles.value = parsed.styles || {}
      }
      if (data.value.length > 0) return
    } catch {}
  }
  // Default 10x5 empty grid
  data.value = Array.from({ length: 10 }, () => Array(5).fill(''))
}

function getColLabel(index: number): string {
  let label = ''
  let i = index
  while (i >= 0) {
    label = String.fromCharCode(65 + (i % 26)) + label
    i = Math.floor(i / 26) - 1
  }
  return label
}

function getCellStyle(ri: number, ci: number): Record<string, string> {
  const key = `${ri}-${ci}`
  const style = styles.value[key] || {}
  const result: Record<string, string> = {}
  if (style.bgColor && style.bgColor !== '#ffffff') {
    result.backgroundColor = style.bgColor
  }
  if (style.textColor && style.textColor !== '#000000') {
    result.color = style.textColor
  }
  if (style.align) {
    result.textAlign = style.align
  }
  return result
}

function onCellInput(ri: number, ci: number, val: string) {
  data.value[ri][ci] = val
  scheduleSave()
}

function handleToolbarAction(event: { action: string; params?: any }) {
  if (!activeCell.value) {
    return
  }
  const { r, c } = activeCell.value
  const key = `${r}-${c}`

  switch (event.action) {
    case 'addRowAbove':
      insertRowAt(r)
      break
    case 'addRowBelow':
      insertRowAt(r + 1)
      break
    case 'addColLeft':
      insertColAt(c)
      break
    case 'addColRight':
      insertColAt(c + 1)
      break
    case 'deleteRow':
      deleteRowAt(r)
      break
    case 'deleteCol':
      deleteColAt(c)
      break
    case 'clearCells':
      data.value[r][c] = ''
      scheduleSave()
      break
    case 'mergeCells':
      // TODO: implement merge cells
      break
    case 'freezeRow':
      // TODO: implement freeze row
      break
  }
}

function insertRowAt(index: number) {
  const cols = data.value[0]?.length || 5
  data.value.splice(index, 0, Array(cols).fill(''))
  // Update styles keys
  const newStyles: Record<string, CellStyle> = {}
  for (const [key, style] of Object.entries(styles.value)) {
    const [ri, ci] = key.split('-').map(Number)
    if (ri >= index) {
      newStyles[`${ri + 1}-${ci}`] = style
    } else {
      newStyles[key] = style
    }
  }
  styles.value = newStyles
  scheduleSave()
}

function insertColAt(index: number) {
  for (const row of data.value) {
    row.splice(index, 0, '')
  }
  // Update styles keys
  const newStyles: Record<string, CellStyle> = {}
  for (const [key, style] of Object.entries(styles.value)) {
    const [ri, ci] = key.split('-').map(Number)
    if (ci >= index) {
      newStyles[`${ri}-${ci + 1}`] = style
    } else {
      newStyles[key] = style
    }
  }
  styles.value = newStyles
  scheduleSave()
}

function deleteRowAt(index: number) {
  if (data.value.length <= 1) return
  data.value.splice(index, 1)
  // Update styles keys
  const newStyles: Record<string, CellStyle> = {}
  for (const [key, style] of Object.entries(styles.value)) {
    const [ri, ci] = key.split('-').map(Number)
    if (ri < index) {
      newStyles[key] = style
    } else if (ri > index) {
      newStyles[`${ri - 1}-${ci}`] = style
    }
  }
  styles.value = newStyles
  if (activeCell.value && activeCell.value.r >= data.value.length) {
    activeCell.value = { r: data.value.length - 1, c: activeCell.value.c }
  }
  scheduleSave()
}

function deleteColAt(index: number) {
  if ((data.value[0]?.length || 0) <= 1) return
  for (const row of data.value) {
    row.splice(index, 1)
  }
  // Update styles keys
  const newStyles: Record<string, CellStyle> = {}
  for (const [key, style] of Object.entries(styles.value)) {
    const [ri, ci] = key.split('-').map(Number)
    if (ci < index) {
      newStyles[key] = style
    } else if (ci > index) {
      newStyles[`${ri}-${ci - 1}`] = style
    }
  }
  styles.value = newStyles
  if (activeCell.value && activeCell.value.c >= (data.value[0]?.length || 0)) {
    activeCell.value = { r: activeCell.value.r, c: (data.value[0]?.length || 1) - 1 }
  }
  scheduleSave()
}

function scheduleSave() {
  if (saveTimer) clearTimeout(saveTimer)
  saveTimer = setTimeout(() => {
    emit('save', JSON.stringify({
      cells: data.value,
      styles: styles.value,
    }))
  }, 2000)
}

onMounted(() => initData())
watch(() => props.content, () => initData(), { once: true })
</script>

<style scoped>
.sheet-editor {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}
.sheet-container {
  flex: 1;
  overflow: auto;
  padding: 0;
}
.sheet-table {
  border-collapse: collapse;
  width: max-content;
  min-width: 100%;
}
.sheet-table th, .sheet-table td {
  border: 1px solid #dee0e3;
  padding: 0;
  height: 28px;
  min-width: 80px;
}
.row-header {
  background: #f5f6f7;
  text-align: center;
  width: 40px;
  min-width: 40px;
  font-size: 12px;
  color: #646a73;
  user-select: none;
  padding: 0 4px;
}
.col-header {
  background: #f5f6f7;
  text-align: center;
  font-size: 12px;
  color: #646a73;
  font-weight: 500;
  user-select: none;
  height: 24px;
}
.sheet-cell {
  position: relative;
  padding: 0;
}
.sheet-cell.active {
  outline: 2px solid #3370ff;
  outline-offset: -1px;
  z-index: 1;
}
.cell-input {
  width: 100%;
  height: 100%;
  border: none;
  outline: none;
  padding: 2px 6px;
  font-size: 13px;
  background: transparent;
  box-sizing: border-box;
}
</style>
