<template>
  <div class="feedback-review-page">
    <h2 class="page-title">{{ $t('admin.feedbackTitle') }}</h2>

    <!-- 统计卡片区 -->
    <div class="stats-cards">
      <el-card shadow="hover" class="stat-card">
        <div class="stat-content">
          <div class="stat-label">{{ $t('admin.pending') }}</div>
          <div class="stat-value pending">{{ stats.pending }}</div>
        </div>
      </el-card>
      <el-card shadow="hover" class="stat-card">
        <div class="stat-content">
          <div class="stat-label">{{ $t('admin.approved') }}</div>
          <div class="stat-value approved">{{ stats.approved }}</div>
        </div>
      </el-card>
      <el-card shadow="hover" class="stat-card">
        <div class="stat-content">
          <div class="stat-label">{{ $t('admin.rejected') }}</div>
          <div class="stat-value rejected">{{ stats.rejected }}</div>
        </div>
      </el-card>
    </div>

    <!-- 审核列表 -->
    <el-card shadow="never" class="table-card">
      <el-table
        :data="reviewList"
        v-loading="loading"
        row-key="id"
        :expand-row-keys="expandedRows"
        @expand-change="handleExpandChange"
        style="width: 100%"
      >
        <el-table-column type="expand">
          <template #default="{ row }">
            <div class="expand-detail">
              <div class="detail-section">
                <div class="detail-label">{{ $t('admin.originalQuestion') }}</div>
                <div class="detail-content">{{ row.originalQuestion }}</div>
              </div>
              <div class="detail-section">
                <div class="detail-label">{{ $t('admin.originalAnswer') }}</div>
                <div class="detail-content ai-answer" v-html="renderMarkdown(row.originalAnswer)"></div>
              </div>
              <div class="detail-section highlight">
                <div class="detail-label">{{ $t('admin.userCorrectAnswer') }}</div>
                <div class="detail-content correct-answer">{{ row.correctAnswer }}</div>
              </div>
              <div class="detail-section" v-if="row.comment">
                <div class="detail-label">{{ $t('admin.userComment') }}</div>
                <div class="detail-content">{{ row.comment }}</div>
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="userName" :label="$t('admin.submitter')" width="100" />
        <el-table-column prop="feedbackType" :label="$t('admin.feedbackType')" width="110">
          <template #default="{ row }">
            <el-tag :type="feedbackTagType(row.feedbackType)" size="small">
              {{ feedbackTypeLabel(row.feedbackType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="originalQuestion" :label="$t('admin.originalQuestion')" show-overflow-tooltip min-width="200" />
        <el-table-column prop="correctAnswer" :label="$t('admin.correctAnswer')" show-overflow-tooltip min-width="200" />
        <el-table-column prop="createdAt" :label="$t('admin.submitTime')" width="160">
          <template #default="{ row }">
            {{ formatTime(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column :label="$t('admin.operations')" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="success" size="small" @click="handleApprove(row)">{{ $t('admin.approve') }}</el-button>
            <el-button type="danger" size="small" @click="handleReject(row)">{{ $t('admin.reject') }}</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @size-change="fetchList"
          @current-change="fetchList"
        />
      </div>
    </el-card>

    <!-- 拒绝原因弹窗 -->
    <el-dialog v-model="rejectDialogVisible" :title="$t('admin.rejectReview')" width="460px">
      <el-form>
        <el-form-item :label="$t('admin.rejectReason')" required>
          <el-input
            v-model="rejectComment"
            type="textarea"
            :rows="3"
            :placeholder="$t('admin.rejectReasonPlaceholder')"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="danger" :disabled="!rejectComment.trim()" @click="confirmReject">{{ $t('admin.confirmReject') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { getPendingReviews, approveReview, rejectReview, getReviewStats } from '@/api/modules/admin'
import type { FeedbackReview } from '@/api/modules/admin'

const { t } = useI18n()

const loading = ref(false)
const reviewList = ref<FeedbackReview[]>([])
const expandedRows = ref<number[]>([])
const stats = reactive({ pending: 0, approved: 0, rejected: 0 })
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

const rejectDialogVisible = ref(false)
const rejectComment = ref('')
const currentRejectId = ref<number | null>(null)

const feedbackTypeMap: Record<string, string> = {
  inaccurate: t('admin.inaccurate'),
  incomplete: t('admin.incomplete'),
  irrelevant: t('admin.irrelevant'),
  outdated: t('admin.outdated'),
}

function feedbackTypeLabel(type: string) {
  return feedbackTypeMap[type] || type
}

function feedbackTagType(type: string): '' | 'warning' | 'danger' | 'info' {
  const map: Record<string, '' | 'warning' | 'danger' | 'info'> = {
    inaccurate: 'danger',
    incomplete: 'warning',
    irrelevant: 'info',
    outdated: '',
  }
  return map[type] || ''
}

function formatTime(dateStr: string) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  if (diff < 60000) return t('admin.justNow')
  if (diff < 3600000) return t('admin.minutesAgo', { count: Math.floor(diff / 60000) })
  if (diff < 86400000) return t('admin.hoursAgo', { count: Math.floor(diff / 3600000) })
  return d.toLocaleDateString(undefined, { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

function renderMarkdown(text: string) {
  if (!text) return ''
  // Simple markdown rendering: code blocks, bold, line breaks
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\n/g, '<br>')
}

function handleExpandChange(row: FeedbackReview, expandedRowsList: FeedbackReview[]) {
  expandedRows.value = expandedRowsList.map(r => r.id)
}

async function fetchStats() {
  try {
    const res = await getReviewStats()
    if (res.data) {
      Object.assign(stats, res.data)
    }
  } catch { /* ignore */ }
}

async function fetchList() {
  loading.value = true
  try {
    const res = await getPendingReviews({ page: pagination.page, pageSize: pagination.pageSize })
    if (res.data) {
      reviewList.value = res.data.list || []
      pagination.total = res.data.total || 0
    }
  } catch {
    ElMessage.error(t('admin.fetchReviewFailed'))
  } finally {
    loading.value = false
  }
}

async function handleApprove(row: FeedbackReview) {
  try {
    await ElMessageBox.prompt(t('admin.approveNote'), t('admin.approveReview'), {
      confirmButtonText: t('admin.confirmApprove'),
      cancelButtonText: t('common.cancel'),
      inputPlaceholder: t('admin.notePlaceholder'),
      type: 'success',
    }).then(async ({ value }) => {
      await approveReview(row.id, value || '')
      ElMessage.success(t('admin.reviewApproved'))
      fetchList()
      fetchStats()
    })
  } catch { /* cancelled */ }
}

function handleReject(row: FeedbackReview) {
  currentRejectId.value = row.id
  rejectComment.value = ''
  rejectDialogVisible.value = true
}

async function confirmReject() {
  if (!rejectComment.value.trim() || currentRejectId.value === null) return
  try {
    await rejectReview(currentRejectId.value, rejectComment.value.trim())
    ElMessage.success(t('admin.reviewRejected'))
    rejectDialogVisible.value = false
    fetchList()
    fetchStats()
  } catch {
    ElMessage.error(t('admin.operationFailed'))
  }
}

onMounted(() => {
  fetchList()
  fetchStats()
})
</script>

<style scoped>
.feedback-review-page {
  max-width: 1200px;
}
.page-title {
  font-size: 20px;
  font-weight: 600;
  margin: 0 0 20px;
  color: var(--kx-text-primary);
}
.stats-cards {
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
}
.stat-card {
  flex: 1;
  border-radius: 10px;
}
.stat-content {
  text-align: center;
  padding: 8px 0;
}
.stat-label {
  font-size: 14px;
  color: var(--kx-text-secondary, #666);
  margin-bottom: 8px;
}
.stat-value {
  font-size: 32px;
  font-weight: 700;
}
.stat-value.pending {
  color: #e6a23c;
}
.stat-value.approved {
  color: #67c23a;
}
.stat-value.rejected {
  color: #f56c6c;
}
.table-card {
  border-radius: 10px;
}
.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
.expand-detail {
  padding: 16px 24px;
}
.detail-section {
  margin-bottom: 16px;
}
.detail-section.highlight {
  background: #f0f9eb;
  border-left: 3px solid #67c23a;
  padding: 12px 16px;
  border-radius: 4px;
}
.detail-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--kx-text-secondary, #666);
  margin-bottom: 6px;
}
.detail-content {
  font-size: 14px;
  line-height: 1.6;
  color: var(--kx-text-primary, #333);
}
.detail-content.ai-answer :deep(code) {
  background: #f5f5f5;
  padding: 2px 4px;
  border-radius: 3px;
  font-size: 13px;
}
.detail-content.correct-answer {
  font-weight: 500;
}
</style>
