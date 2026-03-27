<template>
  <div class="bitable-editor">
    <BitableToolbar
      :has-selection="selectedRecord !== null"
      :selected-field="selectedField"
      @action="handleToolbarAction"
    />
    <div class="bitable-main">
      <!-- View Tabs -->
      <div class="view-tabs">
        <div
          v-for="view in views"
          :key="view.id"
          class="view-tab"
          :class="{ active: currentViewId === view.id }"
          @click="currentViewId = view.id"
        >
          <el-icon><component :is="getViewIcon(view.type)" /></el-icon>
          <span>{{ view.name }}</span>
        </div>
      </div>
      <!-- Table -->
      <div class="bitable-container">
        <table class="bitable-table">
          <thead>
            <tr>
              <th class="row-header">#</th>
              <th
                v-for="field in fields"
                :key="field.id"
                class="field-header"
                :class="{ selected: selectedField === field.id }"
                @click="selectedField = field.id"
              >
                <div class="field-header-content">
                  <el-icon :size="14"><component :is="getFieldIcon(field.type)" /></el-icon>
                  <span>{{ field.name }}</span>
                </div>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(record, ri) in records"
              :key="record.id"
              :class="{ selected: selectedRecord === record.id }"
              @click="selectedRecord = record.id"
            >
              <td class="row-header">{{ ri + 1 }}</td>
              <td
                v-for="field in fields"
                :key="field.id"
                class="bitable-cell"
                @dblclick="startEdit(record.id, field.id)"
              >
                <template v-if="editingCell?.recordId === record.id && editingCell?.fieldId === field.id">
                  <input
                    class="cell-edit-input"
                    :value="record.data[field.id]"
                    @input="updateCell(record.id, field.id, ($event.target as HTMLInputElement).value)"
                    @blur="editingCell = null"
                    @keyup.enter="editingCell = null"
                    ref="editInputRef"
                    autofocus
                  />
                </template>
                <template v-else>
                  <span class="cell-value">{{ formatCellValue(record.data[field.id], field) }}</span>
                </template>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Field Edit Dialog -->
    <el-dialog v-model="showFieldDialog" title="编辑字段" width="400px">
      <el-form label-width="80px">
        <el-form-item label="字段名称">
          <el-input v-model="editingFieldName" />
        </el-form-item>
        <el-form-item label="字段类型">
          <el-select v-model="editingFieldType" style="width: 100%">
            <el-option label="文本" value="text" />
            <el-option label="数字" value="number" />
            <el-option label="单选" value="select" />
            <el-option label="多选" value="multiSelect" />
            <el-option label="日期" value="date" />
            <el-option label="人员" value="person" />
            <el-option label="附件" value="attachment" />
            <el-option label="链接" value="link" />
            <el-option label="复选框" value="checkbox" />
            <el-option label="评分" value="rating" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="editingFieldType === 'select' || editingFieldType === 'multiSelect'" label="选项">
          <div v-for="(opt, i) in editingFieldOptions" :key="i" style="display:flex;gap:8px;margin-bottom:4px">
            <el-input v-model="editingFieldOptions[i]" size="small" />
            <el-button size="small" @click="editingFieldOptions.splice(i, 1)"><el-icon><Close /></el-icon></el-button>
          </div>
          <el-button size="small" text @click="editingFieldOptions.push('新选项')"><el-icon><Plus /></el-icon>添加选项</el-button>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showFieldDialog = false">取消</el-button>
        <el-button type="primary" @click="saveField">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, nextTick } from 'vue'
import BitableToolbar from './BitableToolbar.vue'
import { ElMessage } from 'element-plus'

interface Field {
  id: string
  name: string
  type: string
  options?: string[]
}

interface Record {
  id: string
  data: Record<string, any>
}

interface View {
  id: string
  name: string
  type: 'table' | 'kanban' | 'calendar' | 'gallery'
}

interface BitableData {
  fields: Field[]
  records: Record[]
  views: View[]
}

const props = defineProps<{
  documentId: number
  content: string
}>()

const emit = defineEmits<{
  (e: 'save', content: string): void
}>()

const fields = ref<Field[]>([])
const records = ref<Record[]>([])
const views = ref<View[]>([])
const currentViewId = ref('')
const selectedRecord = ref<string | null>(null)
const selectedField = ref<string | null>(null)
const editingCell = ref<{ recordId: string; fieldId: string } | null>(null)
const showFieldDialog = ref(false)
const editingFieldName = ref('')
const editingFieldType = ref('text')
const editingFieldOptions = ref<string[]>([])
const editingFieldId = ref<string | null>(null)
let saveTimer: ReturnType<typeof setTimeout> | null = null
let idCounter = 0

function genId(): string {
  return `id_${Date.now()}_${idCounter++}`
}

function initData() {
  if (props.content) {
    try {
      const parsed: BitableData = JSON.parse(props.content)
      if (parsed.fields && Array.isArray(parsed.fields)) {
        fields.value = parsed.fields
        records.value = parsed.records || []
        views.value = parsed.views || [{ id: genId(), name: '表格视图', type: 'table' }]
        if (views.value.length > 0) {
          currentViewId.value = views.value[0].id
        }
        return
      }
    } catch {}
  }
  // Default data
  fields.value = [
    { id: genId(), name: '标题', type: 'text' },
    { id: genId(), name: '状态', type: 'select', options: ['未开始', '进行中', '已完成'] },
    { id: genId(), name: '负责人', type: 'person' },
    { id: genId(), name: '截止日期', type: 'date' },
  ]
  records.value = [
    { id: genId(), data: {} },
    { id: genId(), data: {} },
    { id: genId(), data: {} },
  ]
  views.value = [{ id: genId(), name: '表格视图', type: 'table' }]
  currentViewId.value = views.value[0].id
}

function getFieldIcon(type: string): string {
  const icons: Record<string, string> = {
    text: 'EditPen',
    number: 'Hashtag',
    select: 'ArrowDown',
    multiSelect: 'ArrowDown',
    date: 'Calendar',
    person: 'User',
    attachment: 'Paperclip',
    link: 'Link',
    checkbox: 'Check',
    rating: 'StarFilled',
  }
  return icons[type] || 'Document'
}

function getViewIcon(type: string): string {
  const icons: Record<string, string> = {
    table: 'Grid',
    kanban: 'Menu',
    calendar: 'Calendar',
    gallery: 'Picture',
  }
  return icons[type] || 'Grid'
}

function formatCellValue(value: any, field: Field): string {
  if (value === undefined || value === null) return ''
  if (field.type === 'date' && value) {
    try {
      return new Date(value).toLocaleDateString('zh-CN')
    } catch { return value }
  }
  if (field.type === 'multiSelect' && Array.isArray(value)) {
    return value.join(', ')
  }
  return String(value)
}

function handleToolbarAction(event: { action: string; params?: any }) {
  switch (event.action) {
    case 'addRecord':
      addRecord()
      break
    case 'deleteRecord':
      deleteRecord()
      break
    case 'addField':
      addField(event.params)
      break
    case 'editField':
      openFieldDialog()
      break
    case 'deleteField':
      deleteField()
      break
    case 'addView':
      addView()
      break
    case 'showFilter':
    case 'showSort':
    case 'showGroup':
      ElMessage.info(`${event.action}功能开发中`)
      break
    case 'export':
      ElMessage.info('导出功能开发中')
      break
    case 'import':
      ElMessage.info('导入功能开发中')
      break
  }
}

function addRecord() {
  records.value.push({ id: genId(), data: {} })
  scheduleSave()
}

function deleteRecord() {
  if (!selectedRecord.value) return
  records.value = records.value.filter(r => r.id !== selectedRecord.value)
  selectedRecord.value = null
  scheduleSave()
}

function addField(type: string) {
  const newField: Field = {
    id: genId(),
    name: '新字段',
    type,
    options: (type === 'select' || type === 'multiSelect') ? ['选项1', '选项2'] : undefined,
  }
  fields.value.push(newField)
  scheduleSave()
}

function openFieldDialog() {
  if (!selectedField.value) return
  const field = fields.value.find(f => f.id === selectedField.value)
  if (!field) return
  editingFieldId.value = field.id
  editingFieldName.value = field.name
  editingFieldType.value = field.type
  editingFieldOptions.value = field.options ? [...field.options] : []
  showFieldDialog.value = true
}

function saveField() {
  if (!editingFieldId.value) return
  const field = fields.value.find(f => f.id === editingFieldId.value)
  if (field) {
    field.name = editingFieldName.value
    field.type = editingFieldType.value
    if (editingFieldType.value === 'select' || editingFieldType.value === 'multiSelect') {
      field.options = editingFieldOptions.value
    }
  }
  showFieldDialog.value = false
  scheduleSave()
}

function deleteField() {
  if (!selectedField.value) return
  fields.value = fields.value.filter(f => f.id !== selectedField.value)
  selectedField.value = null
  scheduleSave()
}

function addView() {
  const newView: View = {
    id: genId(),
    name: '新视图',
    type: 'table',
  }
  views.value.push(newView)
  currentViewId.value = newView.id
  scheduleSave()
}

function startEdit(recordId: string, fieldId: string) {
  editingCell.value = { recordId, fieldId }
  nextTick(() => {
    const input = document.querySelector('.cell-edit-input') as HTMLInputElement
    if (input) input.focus()
  })
}

function updateCell(recordId: string, fieldId: string, value: string) {
  const record = records.value.find(r => r.id === recordId)
  if (record) {
    record.data[fieldId] = value
    scheduleSave()
  }
}

function scheduleSave() {
  if (saveTimer) clearTimeout(saveTimer)
  saveTimer = setTimeout(() => {
    emit('save', JSON.stringify({
      fields: fields.value,
      records: records.value,
      views: views.value,
    }))
  }, 2000)
}

onMounted(() => initData())
watch(() => props.content, () => initData(), { once: true })
</script>

<style scoped>
.bitable-editor {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}
.bitable-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.view-tabs {
  display: flex;
  padding: 8px 16px;
  gap: 8px;
  border-bottom: 1px solid #e5e6eb;
  background: #fff;
}
.view-tab {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  color: #646a73;
}
.view-tab:hover {
  background: #f5f6f7;
}
.view-tab.active {
  background: #e8f0fe;
  color: #3370ff;
}
.bitable-container {
  flex: 1;
  overflow: auto;
}
.bitable-table {
  border-collapse: collapse;
  width: max-content;
  min-width: 100%;
}
.bitable-table th, .bitable-table td {
  border: 1px solid #dee0e3;
  padding: 0;
  height: 32px;
  min-width: 120px;
}
.row-header {
  background: #f5f6f7;
  text-align: center;
  width: 50px;
  min-width: 50px;
  font-size: 12px;
  color: #646a73;
  user-select: none;
}
.field-header {
  background: #f5f6f7;
  font-weight: 500;
  font-size: 13px;
  cursor: pointer;
}
.field-header.selected {
  background: #e8f0fe;
}
.field-header-content {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 12px;
}
.bitable-cell {
  padding: 0 8px;
  cursor: pointer;
}
.bitable-cell:hover {
  background: #f9fafb;
}
tr.selected .bitable-cell {
  background: #e8f0fe;
}
.cell-value {
  font-size: 13px;
}
.cell-edit-input {
  width: 100%;
  height: 100%;
  border: none;
  outline: none;
  padding: 0 8px;
  font-size: 13px;
  background: #fff;
}
</style>
