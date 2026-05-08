<template>
  <div class="admin-dashboard">
    <h2>{{ $t('admin.dashboard') }}</h2>
    <p>{{ $t('admin.dashboardDesc') }}</p>

    <!-- Stats Cards Skeleton -->
    <el-skeleton :loading="pageLoading" animated>
      <template #template>
        <div class="stats-skeleton">
          <div v-for="i in 4" :key="i" class="stat-skeleton-card">
            <el-skeleton-item variant="circle" style="width: 40px; height: 40px" />
            <div class="stat-skeleton-info">
              <el-skeleton-item variant="text" style="width: 60%; height: 14px" />
              <el-skeleton-item variant="text" style="width: 40%; height: 24px; margin-top: 8px" />
            </div>
          </div>
        </div>
      </template>
      <template #default>
        <div class="stats-cards">
          <div class="stat-card">
            <div class="stat-icon" style="background: #e8f0fe"><el-icon :size="24" color="#3370ff"><User /></el-icon></div>
            <div class="stat-info"><div class="stat-label">{{ $t('admin.totalUsers') }}</div><div class="stat-value">--</div></div>
          </div>
          <div class="stat-card">
            <div class="stat-icon" style="background: #e6f7ef"><el-icon :size="24" color="#36b37e"><Document /></el-icon></div>
            <div class="stat-info"><div class="stat-label">{{ $t('admin.totalDocs') }}</div><div class="stat-value">--</div></div>
          </div>
          <div class="stat-card">
            <div class="stat-icon" style="background: #fef3e0"><el-icon :size="24" color="#f5a623"><Collection /></el-icon></div>
            <div class="stat-info"><div class="stat-label">{{ $t('admin.totalKBs') }}</div><div class="stat-value">--</div></div>
          </div>
          <div class="stat-card">
            <div class="stat-icon" style="background: #fce4ec"><el-icon :size="24" color="#f54a45"><TrendCharts /></el-icon></div>
            <div class="stat-info"><div class="stat-label">{{ $t('admin.todayActive') }}</div><div class="stat-value">--</div></div>
          </div>
        </div>
      </template>
    </el-skeleton>

    <!-- Trend Chart Skeleton -->
    <el-skeleton :loading="pageLoading" animated>
      <template #template>
        <div class="chart-skeleton">
          <el-skeleton-item variant="text" style="width: 120px; height: 20px; margin-bottom: 16px" />
          <el-skeleton-item variant="rect" style="width: 100%; height: 240px; border-radius: 8px" />
        </div>
      </template>
      <template #default>
        <div class="chart-area">
          <div class="chart-title">{{ $t('admin.trendChart') }}</div>
          <div class="chart-placeholder">{{ $t('admin.chartPlaceholder') }}</div>
        </div>
      </template>
    </el-skeleton>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { User, Document, Collection, TrendCharts } from '@element-plus/icons-vue'

const { t } = useI18n()
const pageLoading = ref(true)

onMounted(() => {
  // Simulate data loading
  setTimeout(() => {
    pageLoading.value = false
  }, 800)
})
</script>

<style scoped>
.admin-dashboard {
  padding: 24px;
}

.stats-skeleton {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-top: 20px;
}

.stat-skeleton-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px;
  border: 1px solid var(--kx-border);
  border-radius: 10px;
  background: #fff;
}

.stat-skeleton-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.stats-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-top: 20px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px;
  border: 1px solid var(--kx-border);
  border-radius: 10px;
  background: #fff;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-info {
  flex: 1;
}

.stat-label {
  font-size: 13px;
  color: var(--kx-text-secondary);
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: var(--kx-text-primary);
  margin-top: 4px;
}

.chart-skeleton {
  margin-top: 24px;
  padding: 20px;
  border: 1px solid var(--kx-border);
  border-radius: 10px;
  background: #fff;
}

.chart-area {
  margin-top: 24px;
  padding: 20px;
  border: 1px solid var(--kx-border);
  border-radius: 10px;
  background: #fff;
}

.chart-title {
  font-size: 15px;
  font-weight: 500;
  color: var(--kx-text-primary);
  margin-bottom: 16px;
}

.chart-placeholder {
  height: 240px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--kx-text-placeholder);
  font-size: 14px;
  background: #fafbfc;
  border-radius: 8px;
}

@media (max-width: 768px) {
  .stats-skeleton,
  .stats-cards {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 480px) {
  .stats-skeleton,
  .stats-cards {
    grid-template-columns: 1fr;
  }
}
</style>
