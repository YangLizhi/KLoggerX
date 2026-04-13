<template>
  <div class="recent-docs-page">
    <h2>最近访问</h2>
    <el-table v-loading="loading" :data="documents" style="width: 100%; margin-top: 16px" table-layout="auto" @row-click="handleRowClick">
      <el-table-column label="名称" min-width="200">
        <template #default="{ row }">
          <div style="display:flex;align-items:center;gap:8px">
            <el-icon :color="typeColor(row.type)"><component :is="typeIcon(row.type)" /></el-icon>
            <span>{{ row.title }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="所有者" prop="ownerName" min-width="100" />
      <el-table-column label="访问时间" min-width="130">
        <template #default="{ row }">{{ fmtTime(row.updatedAt) }}</template>
      </el-table-column>
    </el-table>
    <div style="margin-top:16px;text-align:center">
      <el-pagination background layout="prev, pager, next" :total="total" :page-size="pageSize" v-model:current-page="page" @current-change="fetchData" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getRecentDocuments } from '@/api/modules/document'
import type { Document } from '@/types'

const loading = ref(false)
const documents = ref<Document[]>([])
const page = ref(1)
const pageSize = 20
const total = ref(0)

const iconMap: Record<string, string> = { folder:'Folder', doc:'Document', sheet:'Grid', slide:'Monitor', mindnote:'Share', bitable:'Tickets', survey:'Notebook' }
const colorMap: Record<string, string> = { folder:'#f5a623', doc:'#3370ff', sheet:'#36b37e', slide:'#ff7d00', mindnote:'#9254de' }
function typeIcon(t: string) { return iconMap[t] || 'Document' }
function typeColor(t: string) { return colorMap[t] || '#3370ff' }
function fmtTime(t: string) { return t ? new Date(t).toLocaleString('zh-CN', { month:'2-digit', day:'2-digit', hour:'2-digit', minute:'2-digit' }) : '' }

async function fetchData() {
  loading.value = true
  try {
    const res: any = await getRecentDocuments({ page: page.value, pageSize })
    documents.value = res.data?.list || []
    total.value = res.data?.total || 0
  } finally { loading.value = false }
}
function handleRowClick(row: any) {
  window.open(`/doc/${row.id}`, '_blank')
}

onMounted(fetchData)
</script>
