<template>
  <div class="template-management">
    <div class="page-header">
      <h2>模板管理</h2>
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>新建模板
      </el-button>
    </div>

    <!-- Search and Filter -->
    <div class="filter-bar">
      <el-input v-model="searchKeyword" placeholder="搜索模板名称" prefix-icon="Search" clearable style="width: 280px" />
      <el-select v-model="filterType" placeholder="类型筛选" clearable style="width: 140px; margin-left: 12px">
        <el-option label="文档" value="doc" />
        <el-option label="表格" value="sheet" />
        <el-option label="幻灯片" value="slide" />
        <el-option label="思维导图" value="mindnote" />
        <el-option label="多维表格" value="bitable" />
        <el-option label="问卷" value="survey" />
      </el-select>
    </div>

    <!-- Template Table -->
    <el-table :data="filteredTemplates" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="名称" min-width="160">
        <template #default="{ row }">
          <div class="template-name">
            <el-icon :color="getTypeColor(row.type)" style="margin-right: 6px;"><component :is="getTypeIcon(row.type)" /></el-icon>
            <span>{{ row.name }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="描述" min-width="180" show-overflow-tooltip />
      <el-table-column prop="category" label="分类" width="120">
        <template #default="{ row }">
          <el-tag size="small" type="info">{{ row.category || '未分类' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="type" label="类型" width="100">
        <template #default="{ row }">
          <span class="type-tag" :style="{ background: getTypeColor(row.type) + '20', color: getTypeColor(row.type) }">
            {{ typeMap[row.type] || row.type }}
          </span>
        </template>
      </el-table-column>
      <el-table-column prop="isBuiltin" label="内置模板" width="100">
        <template #default="{ row }">
          <el-tag :type="row.isBuiltin ? 'success' : 'info'" size="small">
            {{ row.isBuiltin ? '是' : '否' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="创建时间" width="160">
        <template #default="{ row }">
          {{ formatDate(row.createdAt) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="editTemplate(row)">编辑</el-button>
          <el-button link type="danger" @click="deleteTemplateConfirm(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Create/Edit Dialog -->
    <el-dialog v-model="dialogVisible" :title="editingTemplate ? '编辑模板' : '新建模板'" width="600px" destroy-on-close>
      <el-form :model="templateForm" :rules="formRules" ref="formRef" label-width="90px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="templateForm.name" placeholder="请输入模板名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="templateForm.description" type="textarea" :rows="2" placeholder="请输入模板描述" />
        </el-form-item>
        <el-form-item label="分类" prop="category">
          <el-autocomplete
            v-model="templateForm.category"
            :fetch-suggestions="queryCategories"
            placeholder="请输入或选择分类"
            style="width: 100%"
            clearable
          />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="templateForm.type" placeholder="请选择模板类型" style="width: 100%">
            <el-option label="文档" value="doc" />
            <el-option label="表格" value="sheet" />
            <el-option label="幻灯片" value="slide" />
            <el-option label="思维导图" value="mindnote" />
            <el-option label="多维表格" value="bitable" />
            <el-option label="问卷" value="survey" />
          </el-select>
        </el-form-item>
        <el-form-item label="内置模板">
          <el-switch v-model="templateForm.isBuiltin" />
        </el-form-item>
        <el-form-item label="内容">
          <el-input
            v-model="templateForm.content"
            type="textarea"
            :rows="6"
            placeholder="模板内容（JSON格式）"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getAdminTemplates, createTemplate, updateTemplate, deleteTemplate, type Template } from '@/api/modules/admin'

const loading = ref(false)
const templates = ref<Template[]>([])
const categories = ref<string[]>([])
const searchKeyword = ref('')
const filterType = ref('')

const typeMap: Record<string, string> = {
  doc: '文档',
  sheet: '表格',
  slide: '幻灯片',
  mindnote: '思维导图',
  bitable: '多维表格',
  survey: '问卷',
}

const filteredTemplates = computed(() => {
  let result = templates.value
  if (searchKeyword.value) {
    const kw = searchKeyword.value.toLowerCase()
    result = result.filter(t => t.name.toLowerCase().includes(kw) || (t.description && t.description.toLowerCase().includes(kw)))
  }
  if (filterType.value) {
    result = result.filter(t => t.type === filterType.value)
  }
  return result
})

// Dialog and form
const dialogVisible = ref(false)
const editingTemplate = ref<Template | null>(null)
const formRef = ref<FormInstance>()
const submitting = ref(false)
const templateForm = ref({
  name: '',
  description: '',
  category: '',
  type: 'doc',
  content: '',
  isBuiltin: false,
})

const formRules: FormRules = {
  name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择模板类型', trigger: 'change' }],
}

function showCreateDialog() {
  editingTemplate.value = null
  templateForm.value = {
    name: '',
    description: '',
    category: '',
    type: 'doc',
    content: '',
    isBuiltin: false,
  }
  dialogVisible.value = true
}

function editTemplate(template: Template) {
  editingTemplate.value = template
  templateForm.value = {
    name: template.name,
    description: template.description || '',
    category: template.category || '',
    type: template.type,
    content: template.content || '',
    isBuiltin: template.isBuiltin,
  }
  dialogVisible.value = true
}

async function submitForm() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    if (editingTemplate.value) {
      await updateTemplate(editingTemplate.value.id, templateForm.value)
      ElMessage.success('模板已更新')
    } else {
      await createTemplate(templateForm.value)
      ElMessage.success('模板已创建')
    }
    dialogVisible.value = false
    fetchTemplates()
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

async function deleteTemplateConfirm(template: Template) {
  await ElMessageBox.confirm(`确定要删除模板 "${template.name}" 吗？`, '删除确认', { type: 'warning' })
  try {
    await deleteTemplate(template.id)
    ElMessage.success('模板已删除')
    fetchTemplates()
  } catch (e: any) {
    ElMessage.error(e.message || '删除失败')
  }
}

function queryCategories(queryString: string, cb: (results: { value: string }[]) => void) {
  const results = categories.value
    .filter(c => !queryString || c.toLowerCase().includes(queryString.toLowerCase()))
    .map(c => ({ value: c }))
  cb(results)
}

// Helpers
function formatDate(date: string) {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

function getTypeColor(type: string) {
  const map: Record<string, string> = {
    doc: '#3370ff',
    sheet: '#36b37e',
    slide: '#ff7d00',
    mindnote: '#9b59b6',
    bitable: '#e91e63',
    survey: '#00bcd4',
  }
  return map[type] || '#999'
}

function getTypeIcon(type: string) {
  const map: Record<string, string> = {
    doc: 'Document',
    sheet: 'Grid',
    slide: 'Monitor',
    mindnote: 'Share',
    bitable: 'List',
    survey: 'EditPen',
  }
  return map[type] || 'Document'
}

async function fetchTemplates() {
  loading.value = true
  try {
    const res: any = await getAdminTemplates()
    templates.value = res.data?.list || []
    categories.value = res.data?.categories || []
  } catch (e) {
    console.error('Failed to fetch templates:', e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchTemplates()
})
</script>

<style scoped>
.template-management {
  padding: 24px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.page-header h2 {
  margin: 0;
  font-size: 20px;
  color: var(--kx-text-primary);
}
.filter-bar {
  display: flex;
  align-items: center;
  margin-bottom: 16px;
}
.template-name {
  display: flex;
  align-items: center;
}
.type-tag {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}
</style>
