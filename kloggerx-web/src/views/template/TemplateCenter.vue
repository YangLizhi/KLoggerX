<template>
  <div class="template-center">
    <div class="tc-header">
      <div class="tc-title">模板库</div>
      <div class="tc-subtitle">选择模板，快速开始创作</div>
      <el-input v-model="keyword" placeholder="搜索模板..." prefix-icon="Search" clearable style="width:260px;margin-top:16px" @input="filterTemplates" />
    </div>

    <div class="tc-body">
      <!-- Category sidebar -->
      <div class="tc-sidebar">
        <div
          class="tc-cat-item"
          :class="{ active: activeCategory === '' }"
          @click="activeCategory = ''; filterTemplates()"
        >全部模板</div>
        <div
          v-for="cat in categories"
          :key="cat"
          class="tc-cat-item"
          :class="{ active: activeCategory === cat }"
          @click="activeCategory = cat; filterTemplates()"
        >{{ cat }}</div>
      </div>

      <!-- Template grid -->
      <div class="tc-content">
        <div v-loading="loading" class="tc-grid">
          <div
            v-for="tpl in filteredTemplates"
            :key="tpl.id"
            class="tc-card"
            @click="selectTemplate(tpl)"
          >
            <div class="tc-card-icon" :style="{ background: getTypeBg(tpl.type) }">
              <el-icon :size="32" :color="getTypeColor(tpl.type)">
                <component :is="getTypeIcon(tpl.type)" />
              </el-icon>
            </div>
            <div class="tc-card-info">
              <div class="tc-card-name">{{ tpl.name }}</div>
              <div class="tc-card-desc">{{ tpl.description }}</div>
              <div class="tc-card-meta">
                <el-tag size="small" type="info">{{ tpl.category }}</el-tag>
                <el-tag size="small" :type="getTypeTagType(tpl.type)">{{ typeNameMap[tpl.type] || tpl.type }}</el-tag>
              </div>
            </div>
          </div>
          <div v-if="!loading && filteredTemplates.length === 0" class="tc-empty">
            <el-empty description="暂无匹配的模板" />
          </div>
        </div>
      </div>
    </div>

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
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { get, post } from '@/api/request'
import { getDocumentTree } from '@/api/modules/document'
import type { Document } from '@/types'

const router = useRouter()
const loading = ref(false)
const keyword = ref('')
const activeCategory = ref('')
const templates = ref<any[]>([])
const categories = ref<string[]>([])
const folders = ref<Document[]>([])
const showUseDialog = ref(false)
const selectedTpl = ref<any>(null)
const newTitle = ref('')
const targetParentId = ref<number | undefined>(undefined)
const creating = ref(false)

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

const filteredTemplates = computed(() => {
  let list = templates.value
  if (activeCategory.value) list = list.filter(t => t.category === activeCategory.value)
  if (keyword.value) list = list.filter(t =>
    t.name.includes(keyword.value) || t.description.includes(keyword.value) || t.category.includes(keyword.value)
  )
  return list
})

function filterTemplates() {
  // reactivity handles it via computed
}

async function loadTemplates() {
  loading.value = true
  try {
    const res: any = await get('/api/v1/template/list')
    templates.value = res.data?.list || []
    categories.value = res.data?.categories || []
  } finally {
    loading.value = false
  }
}

async function loadFolders() {
  const res: any = await getDocumentTree(null)
  folders.value = (res.data || []).filter((d: Document) => d.type === 'folder')
}

function selectTemplate(tpl: any) {
  selectedTpl.value = tpl
  newTitle.value = tpl.name
  targetParentId.value = undefined
  showUseDialog.value = true
}

async function confirmUse() {
  if (!selectedTpl.value) return
  creating.value = true
  try {
    const res: any = await post(`/api/v1/template/${selectedTpl.value.id}/use`, {
      title: newTitle.value || selectedTpl.value.name,
      parentId: targetParentId.value || null,
    })
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
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
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
.tc-card-icon {
  width: 64px;
  height: 64px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.tc-card-name {
  font-size: 15px;
  font-weight: 600;
  color: #1f2329;
}
.tc-card-desc {
  font-size: 13px;
  color: #8f959e;
  line-height: 1.5;
  min-height: 20px;
}
.tc-card-meta {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}
.tc-empty {
  grid-column: 1 / -1;
  display: flex;
  justify-content: center;
  padding: 40px 0;
}
</style>
