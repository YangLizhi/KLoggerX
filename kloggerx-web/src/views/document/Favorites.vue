<template>
  <div class="favorites-page">
    <h2>收藏夹</h2>
    <el-table v-loading="loading" :data="documents" style="width: 100%; margin-top: 16px" table-layout="auto" @row-click="handleRowClick">
      <el-table-column label="名称" min-width="200">
        <template #default="{ row }">
          <div style="display:flex;align-items:center;gap:8px">
            <el-icon><Document /></el-icon>
            <span>{{ row.title }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="所有者" prop="ownerName" min-width="100" />
      <el-table-column label="更新时间" min-width="130">
        <template #default="{ row }">{{ row.updatedAt ? new Date(row.updatedAt).toLocaleString('zh-CN') : '' }}</template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-button link type="warning" @click.stop="handleUnfavorite(row)">取消收藏</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div style="margin-top:16px;text-align:center">
      <el-pagination background layout="prev, pager, next" :total="total" :page-size="20" v-model:current-page="page" @current-change="fetchData" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getFavorites, favoriteDocument } from '@/api/modules/document'
import type { Document } from '@/types'
import { ElMessage } from 'element-plus'

const loading = ref(false)
const documents = ref<Document[]>([])
const page = ref(1)
const total = ref(0)

async function fetchData() {
  loading.value = true
  try {
    const res: any = await getFavorites({ page: page.value, pageSize: 20 })
    documents.value = res.data?.list || []
    total.value = res.data?.total || 0
  } finally { loading.value = false }
}

async function handleUnfavorite(doc: Document) {
  await favoriteDocument(doc.id, false)
  ElMessage.success('已取消收藏')
  fetchData()
}

function handleRowClick(row: any) {
  window.open(`/doc/${row.id}`, '_blank')
}

onMounted(fetchData)
</script>
