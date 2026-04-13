<template>
  <div class="template-center">
    <div class="tc-header">
      <div class="tc-title">模板库</div>
      <div class="tc-subtitle">选择模板，快速开始创作</div>
      <el-input v-model="keyword" placeholder="搜索模板..." prefix-icon="Search" clearable style="width:260px;margin-top:16px" />
    </div>

    <div class="tc-body">
      <!-- Category sidebar -->
      <div class="tc-sidebar">
        <div
          class="tc-cat-item"
          :class="{ active: activeCategory === '' }"
          @click="activeCategory = ''"
        >全部模板</div>
        <div
          v-for="cat in categories"
          :key="cat"
          class="tc-cat-item"
          :class="{ active: activeCategory === cat }"
          @click="activeCategory = cat"
        >{{ cat }}</div>
      </div>

      <!-- Template grid -->
      <div class="tc-content">
        <div v-loading="loading" class="tc-grid">
          <div
            v-for="tpl in filteredTemplates"
            :key="tpl.id"
            class="tc-card"
            @click="previewTemplate(tpl)"
          >
            <div class="tc-card-header">
              <div class="tc-card-icon" :style="{ background: getTypeBg(tpl.type) }">
                <el-icon :size="28" :color="getTypeColor(tpl.type)">
                  <component :is="getTypeIcon(tpl.type)" />
                </el-icon>
              </div>
              <div class="tc-card-title">{{ tpl.name }}</div>
            </div>
            <div class="tc-card-preview">
              {{ tpl.preview || tpl.description || '暂无预览内容' }}
            </div>
            <div class="tc-card-footer">
              <el-tag size="small" type="info">{{ tpl.category || '未分类' }}</el-tag>
              <el-tag size="small" :type="getTypeTagType(tpl.type)">{{ typeNameMap[tpl.type] || tpl.type }}</el-tag>
            </div>
          </div>
          <div v-if="!loading && filteredTemplates.length === 0" class="tc-empty">
            <el-empty description="暂无匹配的模板" />
          </div>
        </div>
      </div>
    </div>

    <!-- Preview Template Dialog -->
    <el-dialog 
      v-model="showPreviewDialog" 
      :title="selectedTpl?.name || '模板预览'" 
      width="700px" 
      destroy-on-close
    >
      <div v-if="selectedTpl" class="preview-content">
        <div class="preview-info">
          <el-tag size="small" type="info">{{ selectedTpl.category || '未分类' }}</el-tag>
          <el-tag size="small" :type="getTypeTagType(selectedTpl.type)">{{ typeNameMap[selectedTpl.type] || selectedTpl.type }}</el-tag>
        </div>
        <div v-if="selectedTpl.description" class="preview-description">
          {{ selectedTpl.description }}
        </div>
        <div class="preview-body">
          <div v-if="previewContent" class="preview-text">
            {{ previewContent }}
          </div>
          <div v-else class="preview-empty">
            <el-empty description="此模板暂无内容" :image-size="80" />
            <p class="preview-hint">您可以直接使用此模板创建文档，然后在编辑器中添加内容</p>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="showPreviewDialog = false">取消</el-button>
        <el-button type="primary" @click="showUseDialog = true; showPreviewDialog = false">使用此模板</el-button>
      </template>
    </el-dialog>

    <!-- Use template dialog -->
    <el-dialog v-model="showUseDialog" :title="'使用模板：' + (selectedTpl?.name || '')" width="420px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="文档标题">
          <el-input v-model="newTitle" :placeholder="selectedTpl?.name || ''" maxlength="100" />
        </el-form-item>
        <el-form-item label="保存位置">
          <el-select v-model="targetParentId" placeholder="根目录（我的文档库）" clearable style="width:100%">
            <el-option :value="undefined" label="根目录（我的文档库）" />
            <el-option v-for="f in folders" :key="f.id" :value="f.id" :label="f.title" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showUseDialog = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="confirmUse">创建文档</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { getTemplates, useTemplate as apiUseTemplate, type Template } from '@/api/modules/template'
import { getDocumentTree } from '@/api/modules/document'
import type { Document } from '@/types'

const loading = ref(false)
const keyword = ref('')
const activeCategory = ref('')
const templates = ref<Template[]>([])
const categories = ref<string[]>([])
const folders = ref<Document[]>([])
const showPreviewDialog = ref(false)
const showUseDialog = ref(false)
const selectedTpl = ref<Template | null>(null)
const newTitle = ref('')
const targetParentId = ref<number | undefined>(undefined)
const creating = ref(false)
const previewContent = ref('')

const typeNameMap: Record<string, string> = {
  doc: '文档', sheet: '表格', slide: '幻灯片', mindnote: '思维笔记',
  bitable: '多维表格', survey: '问卷',
}
const typeIconMap: Record<string, string> = {
  doc: 'Document', sheet: 'Grid', slide: 'Monitor',
  mindnote: 'Share', bitable: 'Tickets', survey: 'Notebook',
}
const typeColorMap: Record<string, string> = {
  doc: '#3370ff', sheet: '#36b37e', slide: '#ff7d00',
  mindnote: '#9254de', bitable: '#00b8d9', survey: '#f54a45',
}
const typeBgMap: Record<string, string> = {
  doc: '#e8f0fe', sheet: '#e6f7ef', slide: '#fff3e0',
  mindnote: '#f3edfd', bitable: '#e0f7fa', survey: '#fce4e4',
}

function getTypeIcon(t: string) { return typeIconMap[t] || 'Document' }
function getTypeColor(t: string) { return typeColorMap[t] || '#3370ff' }
function getTypeBg(t: string) { return typeBgMap[t] || '#f5f5f5' }
function getTypeTagType(type: string): '' | 'success' | 'warning' | 'info' | 'danger' {
  const m: Record<string, '' | 'success' | 'warning' | 'info' | 'danger'> = {
    doc: '', sheet: 'success', slide: 'warning', mindnote: '', bitable: 'info', survey: 'danger',
  }
  return m[type] || ''
}

// Extract plain text from Tiptap JSON content
function extractTextFromContent(content: string): string {
  if (!content || content === '{}' || content === '""') return ''
  
  try {
    const doc = JSON.parse(content)
    return extractTextFromNode(doc)
  } catch {
    // If not valid JSON, return as-is (might be plain text or HTML)
    return content.replace(/<[^>]*>/g, '').trim()
  }
}

function extractTextFromNode(node: any): string {
  if (!node) return ''
  
  const parts: string[] = []
  
  // If node has text property
  if (node.text) {
    parts.push(node.text)
  }
  
  // If node has content array
  if (Array.isArray(node.content)) {
    for (const child of node.content) {
      const childText = extractTextFromNode(child)
      if (childText) parts.push(childText)
    }
  }
  
  return parts.join(' ')
}

const filteredTemplates = computed(() => {
  let list = templates.value
  if (activeCategory.value) list = list.filter(t => t.category === activeCategory.value)
  if (keyword.value) list = list.filter(t =>
    t.name.includes(keyword.value) || 
    (t.description && t.description.includes(keyword.value)) || 
    (t.category && t.category.includes(keyword.value))
  )
  return list
})

async function loadTemplates() {
  loading.value = true
  try {
    const res = await getTemplates() as any
    templates.value = res.data?.list || []
    categories.value = res.data?.categories || []
  } finally {
    loading.value = false
  }
}

async function loadFolders() {
  const res = await getDocumentTree(null) as any
  folders.value = (res.data || []).filter((d: Document) => d.type === 'folder')
}

function previewTemplate(tpl: Template) {
  selectedTpl.value = tpl
  newTitle.value = tpl.name
  targetParentId.value = undefined
  
  // Extract text content for preview
  if (tpl.content && tpl.content !== '{}' && tpl.content !== '""') {
    previewContent.value = extractTextFromContent(tpl.content)
  } else if (tpl.preview) {
    previewContent.value = tpl.preview
  } else {
    previewContent.value = ''
  }
  
  showPreviewDialog.value = true
}

async function confirmUse() {
  if (!selectedTpl.value) return
  creating.value = true
  try {
    const res = await apiUseTemplate(selectedTpl.value.id, {
      title: newTitle.value || selectedTpl.value.name,
      parentId: targetParentId.value || null,
    }) as any
    ElMessage.success('文档已创建')
    showUseDialog.value = false
    if (res.data?.id) {
      window.open(`/doc/${res.data.id}`, '_blank')
    }
  } catch {
    ElMessage.error('创建失败')
  } finally {
    creating.value = false
  }
}

onMounted(() => {
  loadTemplates()
  loadFolders()
})
</script>

<style scoped>
.template-center {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #f7f8fa;
}
.tc-header {
  padding: 32px 40px 24px;
  background: #fff;
  border-bottom: 1px solid #e8e8e8;
}
.tc-title {
  font-size: 24px;
  font-weight: 700;
  color: #1f2329;
}
.tc-subtitle {
  margin-top: 4px;
  font-size: 14px;
  color: #8f959e;
}
.tc-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}
.tc-sidebar {
  width: 160px;
  flex-shrink: 0;
  background: #fff;
  border-right: 1px solid #e8e8e8;
  padding: 16px 0;
  overflow-y: auto;
}
.tc-cat-item {
  padding: 10px 20px;
  font-size: 14px;
  color: #4e5969;
  cursor: pointer;
  border-radius: 0 20px 20px 0;
  margin-right: 8px;
  transition: background 0.15s;
}
.tc-cat-item:hover { background: #f2f3f5; }
.tc-cat-item.active { background: #e8f0fe; color: #3370ff; font-weight: 600; }
.tc-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px 32px;
}
.tc-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 20px;
}
.tc-card {
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  cursor: pointer;
  border: 1.5px solid transparent;
  transition: all 0.18s;
  box-shadow: 0 1px 4px rgba(0,0,0,0.06);
}
.tc-card:hover {
  border-color: #3370ff;
  box-shadow: 0 4px 16px rgba(51,112,255,0.12);
  transform: translateY(-2px);
}
.tc-card-header {
  display: flex;
  align-items: center;
  gap: 12px;
}
.tc-card-icon {
  width: 48px;
  height: 48px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.tc-card-title {
  font-size: 15px;
  font-weight: 600;
  color: #1f2329;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
.tc-card-preview {
  font-size: 13px;
  color: #8f959e;
  line-height: 1.6;
  min-height: 62px;
  max-height: 84px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 4;
  -webkit-box-orient: vertical;
}
.tc-card-footer {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  margin-top: auto;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
}
.tc-empty {
  grid-column: 1 / -1;
  display: flex;
  justify-content: center;
  padding: 40px 0;
}

/* Preview Dialog Styles */
.preview-content {
  max-height: 60vh;
  overflow-y: auto;
}
.preview-info {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}
.preview-description {
  font-size: 14px;
  color: #666;
  margin-bottom: 16px;
  padding: 12px;
  background: #f7f8fa;
  border-radius: 8px;
}
.preview-body {
  min-height: 200px;
}
.preview-text {
  font-size: 14px;
  line-height: 1.8;
  color: #1f2329;
  white-space: pre-wrap;
  word-break: break-word;
  padding: 16px;
  background: #fafbfc;
  border-radius: 8px;
  border: 1px solid #e8e8e8;
}
.preview-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px 0;
}
.preview-hint {
  font-size: 13px;
  color: #8f959e;
  margin-top: 8px;
}
</style>
