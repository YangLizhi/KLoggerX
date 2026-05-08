<template>
  <div class="favorites-page">
    <h2>{{ $t('document.favorites') }}</h2>
    <el-table v-loading="loading" :data="documents" style="width: 100%; margin-top: 16px" table-layout="auto" @row-click="handleRowClick">
      <el-table-column :label="$t('common.name')" min-width="200">
        <template #default="{ row }">
          <div style="display:flex;align-items:center;gap:8px">
            <el-icon><Document /></el-icon>
            <span>{{ row.title }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column :label="$t('document.owner')" prop="ownerName" min-width="100" />
      <el-table-column :label="$t('home.updatedTime')" min-width="130">
        <template #default="{ row }">{{ row.updatedAt ? new Date(row.updatedAt).toLocaleString('zh-CN') : '' }}</template>
      </el-table-column>
      <el-table-column :label="$t('home.operations')" width="100">
        <template #default="{ row }">
          <el-button link type="warning" @click.stop="handleUnfavorite(row)">{{ $t('home.cancelFavorite') }}</el-button>
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
import { useI18n } from 'vue-i18n'
import { getFavorites, favoriteDocument } from '@/api/modules/document'
import type { Document } from '@/types'
import { ElMessage } from 'element-plus'

const { t } = useI18n()

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
  ElMessage.success(t('home.unfavorited'))
  fetchData()
}

function handleRowClick(row: any) {
  window.open(`/doc/${row.id}`, '_blank')
}

onMounted(fetchData)
</script>
