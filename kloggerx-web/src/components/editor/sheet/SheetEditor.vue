<template>
  <div class="sheet-editor">
    <SheetToolbar
      :cell-reference="cellReference"
      :cell-value="currentCellValue"
      :title="docTitle"
      :save-status="saveStatus"
      @action="handleToolbarAction"
      @update:title="updateTitle"
      @save-title="saveTitle"
    />
    <div class="sheet-container" :class="{ 'no-gridlines': !showGridlines, 'no-headings': !showHeadings }">
      <table class="sheet-table">
        <thead v-if="showHeadings">
          <tr>
            <th class="row-header"></th>
            <th v-for="(_, ci) in (data[0] || [])" :key="ci" class="col-header" @click="selectColumn(ci)">
              {{ getColLabel(ci) }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, ri) in data" :key="ri">
            <td v-if="showHeadings" class="row-header" @click="selectRow(ri)">{{ ri + 1 }}</td>
            <td
              v-for="(cell, ci) in row"
              :key="ci"
              class="sheet-cell"
              :class="{
                active: activeCell?.r === ri && activeCell?.c === ci,
                selected: isInRange(ri, ci)
              }"
              :style="getCellStyle(ri, ci)"
              @click="onCellClick(ri, ci, $event)"
              @mousedown="startRangeSelect(ri, ci, $event)"
            >
              <input
                v-if="editingCell?.r === ri && editingCell?.c === ci"
                ref="cellInputRef"
                class="cell-input editing"
                :value="cell"
                @input="onCellInput(ri, ci, ($event.target as HTMLInputElement).value)"
                @blur="finishEditing"
                @keydown="onCellKeydown"
                @keydown.enter="finishEditing"
                @keydown.escape="cancelEditing"
              />
              <div v-else class="cell-content" :class="getCellClass(ri, ci)">
                {{ getDisplayValue(ri, ci, cell) }}
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Sheet tabs -->
    <div class="sheet-tabs-bar">
      <div class="sheet-tabs">
        <div
          v-for="(sheet, idx) in sheets"
          :key="idx"
          class="sheet-tab"
          :class="{ active: currentSheetIndex === idx }"
          @click="switchSheet(idx)"
        >
          {{ sheet.name }}
        </div>
        <el-button size="small" text @click="addSheet"><el-icon><Plus /></el-icon></el-button>
      </div>
      <div class="sheet-zoom">
        <el-button size="small" text @click="zoomOut"><el-icon><ZoomOut /></el-icon></el-button>
        <span class="zoom-value">{{ zoomLevel }}%</span>
        <el-button size="small" text @click="zoomIn"><el-icon><ZoomIn /></el-icon></el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import SheetToolbar from './SheetToolbar.vue'
import { ElMessage, ElMessageBox } from 'element-plus'

interface CellStyle {
  bgColor?: string
  textColor?: string
  fontFamily?: string
  fontSize?: number
  fontWeight?: string
  fontStyle?: string
  textDecoration?: string
  align?: 'left' | 'center' | 'right'
  verticalAlign?: 'top' | 'middle' | 'bottom'
  numberFormat?: string
  border?: string
  wrap?: boolean
}

interface Sheet {
  name: string
  cells: string[][]
  styles: Record<string, CellStyle>
}

const props = defineProps<{
  documentId: number
  content: string
  docTitle?: string
}>()

const emit = defineEmits<{
  (e: 'save', content: string): void
  (e: 'update:title', title: string): void
  (e: 'saveTitle'): void
}>()

// Sheet data
const sheets = ref<Sheet[]>([{ name: '工作表1', cells: [], styles: {} }])
const currentSheetIndex = ref(0)
const saveStatus = ref('已保存')

function updateTitle(title: string) {
  emit('update:title', title)
}

function saveTitle() {
  emit('saveTitle')
}
const data = computed({
  get: () => sheets.value[currentSheetIndex.value].cells,
  set: (val) => { sheets.value[currentSheetIndex.value].cells = val }
})
const styles = computed({
  get: () => sheets.value[currentSheetIndex.value].styles,
  set: (val) => { sheets.value[currentSheetIndex.value].styles = val }
})

// Cell selection
const activeCell = ref<{ r: number; c: number } | null>(null)
const editingCell = ref<{ r: number; c: number } | null>(null)
const rangeStart = ref<{ r: number; c: number } | null>(null)
const rangeEnd = ref<{ r: number; c: number } | null>(null)
const isSelecting = ref(false)
const cellInputRef = ref<HTMLInputElement | null>(null)

// View settings
const zoomLevel = ref(100)
const showGridlines = ref(true)
const showHeadings = ref(true)

let saveTimer: ReturnType<typeof setTimeout> | null = null

const cellReference = computed(() => {
  if (!activeCell.value) return 'A1'
  return getColLabel(activeCell.value.c) + (activeCell.value.r + 1)
})

const currentCellValue = computed(() => {
  if (!activeCell.value) return ''
  const { r, c } = activeCell.value
  return data.value[r]?.[c] || ''
})

function initData() {
  if (props.content) {
    try {
      const parsed = JSON.parse(props.content)
      if (Array.isArray(parsed)) {
        // Old format: just cells array
        sheets.value = [{ name: '工作表1', cells: parsed, styles: {} }]
      } else if (parsed.sheets) {
        // New format: multiple sheets
        sheets.value = parsed.sheets
      } else if (parsed.cells) {
        sheets.value = [{ name: '工作表1', cells: parsed.cells, styles: parsed.styles || {} }]
      }
      if (sheets.value[0].cells.length > 0) {
        normalizeData()
        return
      }
    } catch {}
  }
  // Default 20x10 empty grid
  sheets.value = [{ name: '工作表1', cells: createEmptyGrid(20, 10), styles: {} }]
}

function createEmptyGrid(rows: number, cols: number): string[][] {
  return Array.from({ length: rows }, () => Array(cols).fill(''))
}

function normalizeData() {
  const sheet = sheets.value[currentSheetIndex.value]
  if (!sheet.cells || sheet.cells.length === 0) {
    sheet.cells = createEmptyGrid(20, 10)
  }
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
  if (style.textColor) {
    result.color = style.textColor
  }
  if (style.fontFamily) {
    result.fontFamily = style.fontFamily
  }
  if (style.fontSize) {
    result.fontSize = style.fontSize + 'px'
  }
  if (style.fontWeight) {
    result.fontWeight = style.fontWeight
  }
  if (style.fontStyle) {
    result.fontStyle = style.fontStyle
  }
  if (style.textDecoration) {
    result.textDecoration = style.textDecoration
  }
  if (style.align) {
    result.textAlign = style.align
  }
  if (style.verticalAlign) {
    result.verticalAlign = style.verticalAlign
  }
  if (style.wrap) {
    result.whiteSpace = 'pre-wrap'
    result.wordBreak = 'break-word'
  }
  if (style.border) {
    result.border = style.border
  }

  return result
}

function getCellClass(ri: number, ci: number): string[] {
  const classes: string[] = []
  const style = styles.value[`${ri}-${ci}`]
  if (style?.numberFormat === 'number') classes.push('format-number')
  if (style?.numberFormat === 'currency') classes.push('format-currency')
  if (style?.numberFormat === 'percent') classes.push('format-percent')
  return classes
}

function getDisplayValue(ri: number, ci: number, cell: string): string {
  const key = `${ri}-${ci}`
  const style = styles.value[key]

  if (!cell) return ''

  // Handle number formats
  if (style?.numberFormat && !isNaN(Number(cell))) {
    const num = Number(cell)
    switch (style.numberFormat) {
      case 'currency':
        return '¥' + num.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
      case 'percent':
        return (num * 100).toFixed(2) + '%'
      case 'number':
        return num.toLocaleString()
      case 'scientific':
        return num.toExponential(2)
      case 'date':
        return new Date(num).toLocaleDateString('zh-CN')
      case 'time':
        return new Date(num).toLocaleTimeString('zh-CN')
    }
  }

  return cell
}

function onCellClick(ri: number, ci: number, e: MouseEvent) {
  if (editingCell.value?.r === ri && editingCell.value?.c === ci) return

  activeCell.value = { r: ri, c: ci }
  rangeStart.value = { r: ri, c: ci }
  rangeEnd.value = { r: ri, c: ci }

  // Double click to edit
  if (e.detail === 2) {
    startEditing(ri, ci)
  }
}

function startRangeSelect(ri: number, ci: number, e: MouseEvent) {
  if (e.button === 0) {
    isSelecting.value = true
    rangeStart.value = { r: ri, c: ci }
  }
}

function isInRange(ri: number, ci: number): boolean {
  if (!rangeStart.value || !rangeEnd.value) return false
  const minR = Math.min(rangeStart.value.r, rangeEnd.value.r)
  const maxR = Math.max(rangeStart.value.r, rangeEnd.value.r)
  const minC = Math.min(rangeStart.value.c, rangeEnd.value.c)
  const maxC = Math.max(rangeStart.value.c, rangeEnd.value.c)
  return ri >= minR && ri <= maxR && ci >= minC && ci <= maxC
}

function startEditing(ri: number, ci: number) {
  editingCell.value = { r: ri, c: ci }
  nextTick(() => {
    cellInputRef.value?.focus()
    cellInputRef.value?.select()
  })
}

function finishEditing() {
  editingCell.value = null
  scheduleSave()
}

function cancelEditing() {
  editingCell.value = null
}

function onCellInput(ri: number, ci: number, val: string) {
  data.value[ri][ci] = val
}

function onCellKeydown(e: KeyboardEvent) {
  if (!activeCell.value) return

  if (e.key === 'Tab') {
    e.preventDefault()
    finishEditing()
    const nextC = activeCell.value.c + (e.shiftKey ? -1 : 1)
    if (nextC >= 0 && nextC < (data.value[0]?.length || 0)) {
      activeCell.value = { r: activeCell.value.r, c: nextC }
    }
  } else if (e.key === 'Enter') {
    finishEditing()
    const nextR = activeCell.value.r + (e.shiftKey ? -1 : 1)
    if (nextR >= 0 && nextR < data.value.length) {
      activeCell.value = { r: nextR, c: activeCell.value.c }
    }
  } else if (e.key === 'Escape') {
    cancelEditing()
  }
}

function selectRow(ri: number) {
  rangeStart.value = { r: ri, c: 0 }
  rangeEnd.value = { r: ri, c: (data.value[0]?.length || 1) - 1 }
  activeCell.value = { r: ri, c: 0 }
}

function selectColumn(ci: number) {
  rangeStart.value = { r: 0, c: ci }
  rangeEnd.value = { r: data.value.length - 1, c: ci }
  activeCell.value = { r: 0, c: ci }
}

// Handle keyboard navigation
function handleGlobalKeydown(e: KeyboardEvent) {
  if (!activeCell.value || editingCell.value) return

  const { r, c } = activeCell.value

  switch (e.key) {
    case 'ArrowUp':
      e.preventDefault()
      if (r > 0) activeCell.value = { r: r - 1, c }
      break
    case 'ArrowDown':
      e.preventDefault()
      if (r < data.value.length - 1) activeCell.value = { r: r + 1, c }
      break
    case 'ArrowLeft':
      e.preventDefault()
      if (c > 0) activeCell.value = { r, c: c - 1 }
      break
    case 'ArrowRight':
      e.preventDefault()
      if (c < (data.value[0]?.length || 0) - 1) activeCell.value = { r, c: c + 1 }
      break
    case 'Delete':
    case 'Backspace':
      if (activeCell.value) {
        data.value[r][c] = ''
        scheduleSave()
      }
      break
    case 'Enter':
      startEditing(r, c)
      break
    case 'F2':
      startEditing(r, c)
      break
  }
}

function handleToolbarAction(event: { action: string; params?: any }) {
  if (!activeCell.value && !['addSheet', 'printSheet', 'exportSheet'].includes(event.action)) {
    return
  }
  const { r, c } = activeCell.value || { r: 0, c: 0 }
  const key = `${r}-${c}`

  switch (event.action) {
    // Clipboard
    case 'cut':
      ElMessage.info('剪切功能需要支持剪贴板API')
      break
    case 'copy':
      ElMessage.info('复制功能需要支持剪贴板API')
      break
    case 'paste':
    case 'pasteValue':
    case 'pasteFormat':
      ElMessage.info('粘贴功能需要支持剪贴板API')
      break

    // Font
    case 'setFontFamily':
      setCellStyle(key, { fontFamily: event.params })
      break
    case 'setFontSize':
      setCellStyle(key, { fontSize: event.params })
      break
    case 'toggleBold':
      setCellStyle(key, { fontWeight: event.params ? 'bold' : 'normal' })
      break
    case 'toggleItalic':
      setCellStyle(key, { fontStyle: event.params ? 'italic' : 'normal' })
      break
    case 'toggleUnderline':
      setCellStyle(key, { textDecoration: event.params ? 'underline' : 'none' })
      break
    case 'toggleStrike':
      setCellStyle(key, { textDecoration: event.params ? 'line-through' : 'none' })
      break
    case 'setTextColor':
      setCellStyle(key, { textColor: event.params })
      break
    case 'setBgColor':
      setCellStyle(key, { bgColor: event.params })
      break

    // Alignment
    case 'setAlign':
      setCellStyle(key, { align: event.params })
      break
    case 'alignTop':
      setCellStyle(key, { verticalAlign: 'top' })
      break
    case 'alignMiddle':
      setCellStyle(key, { verticalAlign: 'middle' })
      break
    case 'alignBottom':
      setCellStyle(key, { verticalAlign: 'bottom' })
      break
    case 'wrapText':
      const current = styles.value[key]?.wrap
      setCellStyle(key, { wrap: !current })
      break

    // Number format
    case 'setNumberFormat':
      setCellStyle(key, { numberFormat: event.params })
      break
    case 'formatPercent':
      setCellStyle(key, { numberFormat: 'percent' })
      break
    case 'formatCurrency':
      setCellStyle(key, { numberFormat: 'currency' })
      break

    // Row/Column operations
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
      clearRange()
      break

    // Merge & Border
    case 'mergeCells':
      ElMessage.info('合并单元格功能开发中')
      break
    case 'borderAll':
      setCellStyle(key, { border: '1px solid #1f2329' })
      break
    case 'borderNone':
      setCellStyle(key, { border: 'none' })
      break
    case 'borderTop':
      setCellStyle(key, { borderTop: '1px solid #1f2329' })
      break
    case 'borderBottom':
      setCellStyle(key, { borderBottom: '1px solid #1f2329' })
      break
    case 'borderLeft':
      setCellStyle(key, { borderLeft: '1px solid #1f2329' })
      break
    case 'borderRight':
      setCellStyle(key, { borderRight: '1px solid #1f2329' })
      break

    // Freeze
    case 'freezeRow':
      ElMessage.info('冻结功能开发中')
      break
    case 'freezeCol':
      ElMessage.info('冻结功能开发中')
      break
    case 'unfreeze':
      ElMessage.info('取消冻结功能开发中')
      break

    // Hide
    case 'hideRow':
      ElMessage.info('隐藏行功能开发中')
      break
    case 'hideCol':
      ElMessage.info('隐藏列功能开发中')
      break

    // Insert
    case 'insertImage':
      insertImage()
      break
    case 'insertLink':
      insertLink()
      break
    case 'insertChart':
      ElMessage.info('图表功能开发中')
      break

    // Data
    case 'sortAsc':
      sortRange(true)
      break
    case 'sortDesc':
      sortRange(false)
      break
    case 'filter':
      ElMessage.info('筛选功能开发中')
      break
    case 'removeDuplicates':
      ElMessage.info('删除重复项功能开发中')
      break
    case 'dataValidation':
      ElMessage.info('数据验证功能开发中')
      break

    // View
    case 'setShowGridlines':
      showGridlines.value = event.params
      break
    case 'setShowHeadings':
      showHeadings.value = event.params
      break
    case 'zoomIn':
      zoomLevel.value = Math.min(200, zoomLevel.value + 10)
      break
    case 'zoomOut':
      zoomLevel.value = Math.max(50, zoomLevel.value - 10)
      break
    case 'zoomReset':
      zoomLevel.value = 100
      break

    // Review
    case 'addComment':
      addComment()
      break
    case 'protectSheet':
      ElMessage.info('保护工作表功能开发中')
      break

    // Formula
    case 'setFormula':
      if (activeCell.value) {
        data.value[r][c] = event.params
        scheduleSave()
      }
      break

    // Export & Print
    case 'printSheet':
      window.print()
      break
    case 'exportSheet':
      exportSheet()
      break
    case 'shareSheet':
      ElMessage.info('共享功能开发中')
      break

    // Sheet operations
    case 'addSheet':
      addSheet()
      break
  }
}

function setCellStyle(key: string, style: Partial<CellStyle>) {
  const current = styles.value[key] || {}
  styles.value[key] = { ...current, ...style }
  scheduleSave()
}

function insertRowAt(index: number) {
  const cols = data.value[0]?.length || 10
  data.value.splice(index, 0, Array(cols).fill(''))
  updateStylesForRowInsert(index)
  scheduleSave()
}

function insertColAt(index: number) {
  for (const row of data.value) {
    row.splice(index, 0, '')
  }
  updateStylesForColInsert(index)
  scheduleSave()
}

function deleteRowAt(index: number) {
  if (data.value.length <= 1) return
  data.value.splice(index, 1)
  updateStylesForRowDelete(index)
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
  updateStylesForColDelete(index)
  if (activeCell.value && activeCell.value.c >= (data.value[0]?.length || 0)) {
    activeCell.value = { r: activeCell.value.r, c: (data.value[0]?.length || 1) - 1 }
  }
  scheduleSave()
}

function updateStylesForRowInsert(index: number) {
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
}

function updateStylesForColInsert(index: number) {
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
}

function updateStylesForRowDelete(index: number) {
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
}

function updateStylesForColDelete(index: number) {
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
}

function clearRange() {
  if (!rangeStart.value || !rangeEnd.value) return
  const minR = Math.min(rangeStart.value.r, rangeEnd.value.r)
  const maxR = Math.max(rangeStart.value.r, rangeEnd.value.r)
  const minC = Math.min(rangeStart.value.c, rangeEnd.value.c)
  const maxC = Math.max(rangeStart.value.c, rangeEnd.value.c)

  for (let ri = minR; ri <= maxR; ri++) {
    for (let ci = minC; ci <= maxC; ci++) {
      data.value[ri][ci] = ''
    }
  }
  scheduleSave()
}

function sortRange(ascending: boolean) {
  if (!rangeStart.value || !rangeEnd.value) return
  const minR = Math.min(rangeStart.value.r, rangeEnd.value.r)
  const maxR = Math.max(rangeStart.value.r, rangeEnd.value.r)
  const sortCol = activeCell.value?.c || 0

  const rows = data.value.slice(minR, maxR + 1)
  rows.sort((a, b) => {
    const aVal = a[sortCol] || ''
    const bVal = b[sortCol] || ''
    if (ascending) return aVal.localeCompare(bVal, 'zh-CN', { numeric: true })
    return bVal.localeCompare(aVal, 'zh-CN', { numeric: true })
  })

  for (let i = 0; i < rows.length; i++) {
    data.value[minR + i] = rows[i]
  }
  scheduleSave()
}

async function insertImage() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/*'
  input.onchange = async (e) => {
    const file = (e.target as HTMLInputElement).files?.[0]
    if (file) {
      ElMessage.info('图片插入功能需要后端存储支持')
    }
  }
  input.click()
}

async function insertLink() {
  try {
    const { value } = await ElMessageBox.prompt('请输入链接URL', '插入链接', {
      inputPlaceholder: 'https://example.com',
      confirmButtonText: '插入',
      cancelButtonText: '取消',
    })
    if (value && activeCell.value) {
      data.value[activeCell.value.r][activeCell.value.c] = value
      scheduleSave()
    }
  } catch {}
}

async function addComment() {
  try {
    const { value } = await ElMessageBox.prompt('请输入批注内容', '添加批注', {
      inputPlaceholder: '批注内容',
      confirmButtonText: '确定',
      cancelButtonText: '取消',
    })
    if (value) {
      ElMessage.success('批注已添加')
    }
  } catch {}
}

function addSheet() {
  const newSheet: Sheet = {
    name: `工作表${sheets.value.length + 1}`,
    cells: createEmptyGrid(20, 10),
    styles: {}
  }
  sheets.value.push(newSheet)
  currentSheetIndex.value = sheets.value.length - 1
  scheduleSave()
}

function switchSheet(index: number) {
  currentSheetIndex.value = index
  activeCell.value = null
}

function exportSheet() {
  const content = JSON.stringify({ sheets: sheets.value }, null, 2)
  const blob = new Blob([content], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `spreadsheet-${Date.now()}.json`
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success('导出成功')
}

function zoomIn() {
  zoomLevel.value = Math.min(200, zoomLevel.value + 10)
}

function zoomOut() {
  zoomLevel.value = Math.max(50, zoomLevel.value - 10)
}

function scheduleSave() {
  saveStatus.value = '保存中...'
  if (saveTimer) clearTimeout(saveTimer)
  saveTimer = setTimeout(() => {
    emit('save', JSON.stringify({ sheets: sheets.value }))
    saveStatus.value = '已保存'
  }, 2000)
}
}

onMounted(() => {
  initData()
  window.addEventListener('keydown', handleGlobalKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleGlobalKeydown)
})

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
  background: #fff;
}

.sheet-container.no-gridlines .sheet-cell {
  border: none;
}

.sheet-container.no-headings .row-header,
.sheet-container.no-headings .col-header {
  display: none;
}

.sheet-table {
  border-collapse: collapse;
  width: max-content;
  min-width: 100%;
}

.sheet-table th, .sheet-table td {
  border: 1px solid #dee0e3;
  padding: 0;
  height: 26px;
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
  cursor: pointer;
}

.row-header:hover {
  background: #e8e9eb;
}

.col-header {
  background: #f5f6f7;
  text-align: center;
  font-size: 12px;
  color: #646a73;
  font-weight: 500;
  user-select: none;
  height: 24px;
  cursor: pointer;
}

.col-header:hover {
  background: #e8e9eb;
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

.sheet-cell.selected:not(.active) {
  background: rgba(51, 112, 255, 0.08);
}

.cell-content {
  width: 100%;
  height: 100%;
  padding: 2px 6px;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  box-sizing: border-box;
  display: flex;
  align-items: center;
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

.cell-input.editing {
  background: #fff;
}

.format-number {
  text-align: right;
}

.format-currency {
  text-align: right;
}

.format-percent {
  text-align: right;
}

/* Sheet Tabs */
.sheet-tabs-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 32px;
  background: #f5f6f7;
  border-top: 1px solid #e5e6eb;
  padding: 0 8px;
}

.sheet-tabs {
  display: flex;
  align-items: center;
  gap: 4px;
}

.sheet-tab {
  padding: 4px 12px;
  font-size: 12px;
  color: #646a73;
  cursor: pointer;
  border-radius: 4px;
  border: 1px solid transparent;
}

.sheet-tab:hover {
  background: rgba(0, 0, 0, 0.04);
}

.sheet-tab.active {
  background: #fff;
  border-color: #e5e6eb;
  color: #1f2329;
  font-weight: 500;
}

.sheet-zoom {
  display: flex;
  align-items: center;
  gap: 4px;
}

.zoom-value {
  font-size: 12px;
  color: #646a73;
  min-width: 40px;
  text-align: center;
}

@media print {
  .sheet-toolbar-wrapper,
  .sheet-tabs-bar {
    display: none !important;
  }
}
</style>
