<template>
  <div class="operation-log-page">
    <div class="page-header">
      <h2>{{ $t('admin.logTitle') }}</h2>
      <el-button :icon="Download" @click="handleExport" disabled>{{ $t('common.export') }}</el-button>
    </div>

    <!-- 过滤器 -->
    <div class="filter-bar">
      <el-select
        v-model="filter.userId"
        :placeholder="$t('admin.selectUser')"
        clearable
        filterable
        style="width: 160px"
        @change="handleSearch"
      >
        <el-option
          v-for="user in userList"
          :key="user.id"
          :label="user.nickname || user.username"
          :value="user.id"
        />
      </el-select>

      <el-select
        v-model="filter.action"
        :placeholder="$t('admin.actionType')"
        clearable
        style="width: 160px"
        @change="handleSearch"
      >
        <el-option label="POST" value="POST" />
        <el-option label="PUT" value="PUT" />
        <el-option label="DELETE" value="DELETE" />
      </el-select>

      <el-date-picker
        v-model="filter.dateRange"
        type="daterange"
        range-separator="-"
        :start-placeholder="$t('admin.startDate')"
        :end-placeholder="$t('admin.endDate')"
        value-format="YYYY-MM-DD"
        style="width: 280px"
        @change="handleSearch"
      />

      <el-button type="primary" @click="handleSearch">{{ $t('admin.query') }}</el-button>
      <el-button @click="handleReset">{{ $t('common.reset') }}</el-button>
    </div>

    <!-- 表格 -->
    <el-table
      :data="logList"
      v-loading="loading"
      stripe
      style="width: 100%"
    >
      <el-table-column :label="$t('admin.time')" prop="createdAt" width="170">
        <template #default="{ row }">
          {{ formatTime(row.createdAt) }}
        </template>
      </el-table-column>
      <el-table-column :label="$t('admin.user')" prop="userName" width="120" />
      <el-table-column :label="$t('admin.action')" prop="action" width="220" />
      <el-table-column :label="$t('admin.resourcePath')" prop="resource" min-width="200" show-overflow-tooltip />
      <el-table-column label="IP" prop="ip" width="140" />
      <el-table-column :label="$t('admin.statusCode')" prop="status" width="80" align="center">
        <template #default="{ row }">
          <el-tag :type="row.status < 400 ? 'success' : 'danger'" size="small">
            {{ row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="$t('admin.duration')" prop="duration" width="90" align="center">
        <template #default="{ row }">
          {{ row.duration }}ms
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-wrapper">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchLogs"
        @current-change="fetchLogs"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Download } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { getOperationLogs, getAdminUsers } from '@/api/modules/admin'

const { t } = useI18n()

interface UserItem {
  id: number
  username: string
  nickname: string
}

const loading = ref(false)
const logList = ref<any[]>([])
const userList = ref<UserItem[]>([])

const filter = reactive({
  userId: undefined as number | undefined,
  action: '',
  dateRange: null as [string, string] | null,
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

function formatTime(dateStr: string) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleString('zh-CN', { hour12: false })
}

async function fetchLogs() {
  loading.value = true
  try {
    const params: any = {
      page: pagination.page,
      page_size: pagination.pageSize,
    }
    if (filter.userId) params.user_id = filter.userId
    if (filter.action) params.action = filter.action
    if (filter.dateRange && filter.dateRange[0]) {
      params.start_time = filter.dateRange[0]
      params.end_time = filter.dateRange[1]
    }

    const res = await getOperationLogs(params)
    if (res.data.code === 0) {
      logList.value = res.data.data.list || []
      pagination.total = res.data.data.total || 0
    }
  } catch (e) {
    ElMessage.error(t('admin.fetchLogFailed'))
  } finally {
    loading.value = false
  }
}

async function fetchUsers() {
  try {
    const res = await getAdminUsers({ page: 1, pageSize: 500 })
    if (res.data.code === 0) {
      userList.value = res.data.data?.list || []
    }
  } catch (_) {
    // ignore
  }
}

function handleSearch() {
  pagination.page = 1
  fetchLogs()
}

function handleReset() {
  filter.userId = undefined
  filter.action = ''
  filter.dateRange = null
  pagination.page = 1
  fetchLogs()
}

function handleExport() {
  ElMessage.info(t('admin.exportNotReady'))
}

onMounted(() => {
  fetchLogs()
  fetchUsers()
})
</script>

<style scoped>
.operation-log-page {
  max-width: 1400px;
}
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}
.page-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}
.filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
