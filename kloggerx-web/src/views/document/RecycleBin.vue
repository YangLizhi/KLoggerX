<template>
  <div class="recycle-page">
    <div class="recycle-header">
      <h2 class="recycle-title">回收站</h2>
      <div class="recycle-actions">
        <el-input v-model="searchKey" placeholder="搜索已删除文档..." prefix-icon="Search" clearable class="recycle-search" @keyup.enter="fetchData" />
        <el-button size="small" type="danger" :disabled="!documents.length" @click="handleEmptyAll"><el-icon><Delete /></el-icon>清空回收站</el-button>
      </div>
    </div>

    <!-- Notice -->
    <el-alert title="回收站中的文档将在删除30天后自动清除，届时无法恢复。" type="warning" :closable="true" show-icon class="recycle-notice" />

    <!-- Filter tabs -->
    <div class="recycle-tabs">
      <span class="rc-tab" :class="{ active: filterType === 'all' }" @click="filterType = 'all'">全部</span>
      <span class="rc-tab" :class="{ active: filterType === 'doc' }" @click="filterType = 'doc'">文档</span>
      <span class="rc-tab" :class="{ active: filterType === 'sheet' }" @click="filterType = 'sheet'">表格</span>
      <span class="rc-tab" :class="{ active: filterType === 'slide' }" @click="filterType = 'slide'">幻灯片</span>
      <span class="rc-tab" :class="{ active: filterType === 'folder' }" @click="filterType = 'folder'">文件夹</span>
      <span class="rc-tab" :class="{ active: filterType === 'other' }" @click="filterType = 'other'">其他</span>
    </div>

    <!-- Table -->
    <el-table
      v-loading="loading"
      :data="filteredDocuments"
      style="width: 100%"
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="40" />
      <el-table-column label="名称" min-width="320" sortable>
        <template #default="{ row }">
          <div class="doc-name-cell">
            <el-icon :color="getTypeColor(row.type)"><component :is="getTypeIcon(row.type)" /></el-icon>
            <span>{{ row.title }}</span>
            <el-tag v-if="isExpiringSoon(row.deletedAt)" type="danger" size="small">即将清除</el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="类型" width="100">
        <template #default="{ row }">
          <span class="type-label">{{ getTypeLabel(row.type) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="原始位置" width="180">
        <template #default="{ row }">
          <span class="location-text">{{ row.parentTitle || '根目录' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="删除人" width="120">
        <template #default="{ row }">{{ row.ownerName || '我' }}</template>
      </el-table-column>
      <el-table-column label="删除时间" width="180" sortable>
        <template #default="{ row }">{{ formatDate(row.deletedAt) }}</template>
      </el-table-column>
      <el-table-column label="剩余天数" width="100" align="center">
        <template #default="{ row }">
          <span :class="{ 'days-warn': getRemainingDays(row.deletedAt) <= 7 }">{{ getRemainingDays(row.deletedAt) }} 天</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleRestore(row)"><el-icon><RefreshLeft /></el-icon>恢复</el-button>
          <el-button link type="danger" size="small" @click="handlePermanentDelete(row)"><el-icon><Delete /></el-icon>彻底删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Batch Bar -->
    <transition name="slide-up">
      <div v-if="selectedDocs.length" class="batch-bar">
        <span class="batch-count">已选 {{ selectedDocs.length }} 项</span>
        <el-button size="small" type="primary" @click="batchRestore"><el-icon><RefreshLeft /></el-icon>批量恢复</el-button>
        <el-button size="small" type="danger" @click="batchPermanentDelete"><el-icon><Delete /></el-icon>批量彻底删除</el-button>
        <el-button size="small" text @click="selectedDocs = []">取消选择</el-button>
      </div>
    </transition>

    <!-- Empty State -->
    <div v-if="!loading && !documents.length" class="recycle-empty">
      <el-empty description="回收站为空">
        <template #image>
          <el-icon :size="48" color="var(--kx-text-placeholder)"><Delete /></el-icon>
        </template>
      </el-empty>
    </div>

    <!-- Pagination -->
    <div v-if="total > 20" class="recycle-pagination">
      <el-pagination background layout="total, prev, pager, next" :total="total" :page-size="20" v-model:current-page="page" @current-change="fetchData" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getRecycleBin, restoreDocument, permanentDeleteDocument } from '@/api/modules/document'
import type { Document } from '@/types'
import { ElMessage, ElMessageBox } from 'element-plus'

const loading = ref(false)
const documents = ref<Document[]>([])
const page = ref(1)
const total = ref(0)
const searchKey = ref('')
const filterType = ref('all')
const selectedDocs = ref<Document[]>([])

const typeMap: Record<string, { icon: string; color: string; label: string }> = {
  folder: { icon: 'Folder', color: '#f5a623', label: '文件夹' },
  doc: { icon: 'Document', color: '#3370ff', label: '文档' },
  sheet: { icon: 'Grid', color: '#36b37e', label: '表格' },
  slide: { icon: 'Monitor', color: '#ff7d00', label: '幻灯片' },
  mindnote: { icon: 'Share', color: '#9254de', label: '思维笔记' },
  bitable: { icon: 'Tickets', color: '#00b8d9', label: '多维表格' },
  survey: { icon: 'Notebook', color: '#f54a45', label: '问卷' },
}

function getTypeIcon(type: string) { return typeMap[type]?.icon || 'Document' }
function getTypeColor(type: string) { return typeMap[type]?.color || '#3370ff' }
function getTypeLabel(type: string) { return typeMap[type]?.label || '文档' }

function formatDate(t: string | null) {
  if (!t) return ''
  const d = new Date(t)
  return `${d.getFullYear()}年${d.getMonth()+1}月${d.getDate()}日 ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

function getRemainingDays(deletedAt: string | null): number {
  if (!deletedAt) return 30
  const deleted = new Date(deletedAt)
  const expiry = new Date(deleted.getTime() + 30 * 24 * 60 * 60 * 1000)
  const now = new Date()
  const diff = Math.ceil((expiry.getTime() - now.getTime()) / (24 * 60 * 60 * 1000))
  return Math.max(0, diff)
}

function isExpiringSoon(deletedAt: string | null): boolean {
  return getRemainingDays(deletedAt) <= 3
}

const filteredDocuments = computed(() => {
  let list = documents.value
  // Search
  if (searchKey.value.trim()) {
    const q = searchKey.value.toLowerCase()
    list = list.filter(d => d.title.toLowerCase().includes(q))
  }
  // Type filter
  if (filterType.value !== 'all') {
    if (filterType.value === 'other') {
      list = list.filter(d => !['doc', 'sheet', 'slide', 'folder'].includes(d.type))
    } else {
      list = list.filter(d => d.type === filterType.value)
    }
  }
  return list
})

async function fetchData() {
  loading.value = true
  try {
    const res: any = await getRecycleBin({ page: page.value, pageSize: 20 })
    documents.value = res.data?.list || []
    total.value = res.data?.total || 0
  } finally { loading.value = false }
}

function handleSelectionChange(rows: Document[]) {
  selectedDocs.value = rows
}

async function handleRestore(doc: Document) {
  await restoreDocument(doc.id)
  ElMessage.success(`"${doc.title}" 已恢复`)
  fetchData()
}

async function handlePermanentDelete(doc: Document) {
  await ElMessageBox.confirm(`彻底删除"${doc.title}"？此操作不可撤销`, '警告', { type: 'warning' })
  await permanentDeleteDocument(doc.id)
  ElMessage.success('已彻底删除')
  fetchData()
}

async function batchRestore() {
  const count = selectedDocs.value.length
  for (const doc of selectedDocs.value) {
    try { await restoreDocument(doc.id) } catch { /* continue */ }
  }
  ElMessage.success(`已恢复 ${count} 个文档`)
  selectedDocs.value = []
  fetchData()
}

async function batchPermanentDelete() {
  const count = selectedDocs.value.length
  await ElMessageBox.confirm(`确定彻底删除选中的 ${count} 个文档？此操作不可撤销`, '批量删除', { type: 'warning' })
  for (const doc of selectedDocs.value) {
    try { await permanentDeleteDocument(doc.id) } catch { /* continue */ }
  }
  ElMessage.success(`已彻底删除 ${count} 个文档`)
  selectedDocs.value = []
  fetchData()
}

async function handleEmptyAll() {
  await ElMessageBox.confirm('清空回收站？所有文档将被彻底删除，不可恢复', '警告', { type: 'warning' })
  for (const doc of documents.value) {
    try { await permanentDeleteDocument(doc.id) } catch { /* continue */ }
  }
  ElMessage.success('回收站已清空')
  fetchData()
}

onMounted(fetchData)
</script>

<style scoped>
.recycle-page {
  max-width: 1200px;
}
.recycle-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.recycle-title {
  font-size: 22px;
  font-weight: 600;
}
.recycle-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}
.recycle-search {
  width: 240px;
}
.recycle-notice {
  margin-bottom: 16px;
}

/* Tabs */
.recycle-tabs {
  display: flex;
  gap: 20px;
  margin-bottom: 16px;
  border-bottom: 1px solid var(--kx-border);
}
.rc-tab {
  font-size: 14px;
  color: var(--kx-text-secondary);
  cursor: pointer;
  padding-bottom: 10px;
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
}
.rc-tab.active {
  color: var(--kx-primary);
  border-color: var(--kx-primary);
  font-weight: 500;
}
.rc-tab:hover {
  color: var(--kx-primary);
}

/* Doc name */
.doc-name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}
.type-label {
  font-size: 13px;
  color: var(--kx-text-secondary);
}
.location-text {
  font-size: 13px;
  color: var(--kx-text-secondary);
}
.days-warn {
  color: var(--kx-danger);
  font-weight: 500;
}

/* Batch Bar */
.batch-bar {
  position: sticky;
  bottom: 0;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  background: #fff;
  border-top: 1px solid var(--kx-border);
  box-shadow: 0 -2px 8px rgba(0,0,0,0.06);
  margin-top: 16px;
  border-radius: 8px;
}
.batch-count {
  font-size: 14px;
  font-weight: 500;
  color: var(--kx-primary);
}
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.2s;
}
.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(100%);
  opacity: 0;
}

.recycle-empty {
  padding: 60px 0;
}
.recycle-pagination {
  margin-top: 20px;
  text-align: center;
}
</style>
