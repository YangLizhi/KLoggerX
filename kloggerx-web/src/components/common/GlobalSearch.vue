<template>
  <el-dialog v-model="visible" :show-close="false" class="global-search-dialog"
             width="600px" top="15vh" @open="focusInput">
    <div class="search-container">
      <!-- 搜索输入 -->
      <div class="search-input-wrapper">
        <el-icon><Search /></el-icon>
        <input ref="inputRef" v-model="query" :placeholder="$t('search.placeholder')"
               @input="debouncedSearch" @keydown.esc="close" @keydown.enter="doSearch" />
        <span class="search-shortcut">ESC</span>
      </div>

      <!-- 搜索历史 (无查询时显示) -->
      <div v-if="!query && searchHistory.length" class="search-history">
        <div class="section-title">{{ $t('search.recentSearch') }}</div>
        <div v-for="item in searchHistory" :key="item" class="history-item"
             @click="query = item; doSearch()">
          <el-icon><Clock /></el-icon>
          <span>{{ item }}</span>
          <el-icon class="remove-btn" @click.stop="removeHistory(item)"><Close /></el-icon>
        </div>
      </div>

      <!-- 搜索结果 -->
      <div v-if="query && !loading" class="search-results">
        <div class="results-count" v-if="results.length">
          {{ $t('search.foundResults', { count: totalCount }) }}
        </div>
        <div v-for="result in results" :key="result.id" class="result-item"
             @click="navigateTo(result)">
          <div class="result-title" v-html="highlightText(result.title, query)"></div>
          <div class="result-snippet" v-html="highlightText(result.snippet, query)"></div>
          <div class="result-meta">
            <span class="result-type">{{ getTypeLabel(result.type) }}</span>
            <span class="result-date">{{ formatDate(result.updatedAt) }}</span>
          </div>
        </div>
        <div v-if="!results.length" class="no-results">
          {{ $t('search.noResults') }}
        </div>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading" class="search-loading">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>{{ $t('search.searching') }}</span>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Search, Clock, Close, Loading } from '@element-plus/icons-vue'
import { searchDocuments } from '@/api/modules/document'

interface SearchResult {
  id: number
  title: string
  snippet: string
  type: string
  updatedAt: string
}

const visible = defineModel<boolean>('visible', { default: false })
const router = useRouter()
const { t } = useI18n()

const query = ref('')
const results = ref<SearchResult[]>([])
const loading = ref(false)
const totalCount = ref(0)
const inputRef = ref<HTMLInputElement | null>(null)
const searchHistory = ref<string[]>(
  JSON.parse(localStorage.getItem('kx_search_history') || '[]')
)

// Simple debounce implementation
let debounceTimer: ReturnType<typeof setTimeout> | null = null
function debouncedSearch() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    doSearch()
  }, 300)
}

async function doSearch() {
  if (!query.value.trim()) {
    results.value = []
    totalCount.value = 0
    return
  }
  loading.value = true
  try {
    const res: any = await searchDocuments({ keyword: query.value, page: 1, pageSize: 20 })
    const items = res.data?.items || res.data?.list || []
    results.value = items.map((item: any) => ({
      id: item.id,
      title: item.title || '',
      snippet: item.preview || item.content?.substring(0, 100) || '',
      type: item.type || 'doc',
      updatedAt: item.updatedAt || '',
    }))
    totalCount.value = res.data?.total || results.value.length
    // 保存搜索历史
    saveHistory(query.value)
  } catch {
    results.value = []
    totalCount.value = 0
  }
  loading.value = false
}

// 高亮匹配文本
function highlightText(text: string, keyword: string): string {
  if (!text || !keyword) return text || ''
  const regex = new RegExp(`(${keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi')
  return text.replace(regex, '<mark>$1</mark>')
}

// 搜索历史管理
function saveHistory(keyword: string) {
  const history = searchHistory.value.filter(h => h !== keyword)
  history.unshift(keyword)
  searchHistory.value = history.slice(0, 10)
  localStorage.setItem('kx_search_history', JSON.stringify(searchHistory.value))
}

function removeHistory(item: string) {
  searchHistory.value = searchHistory.value.filter(h => h !== item)
  localStorage.setItem('kx_search_history', JSON.stringify(searchHistory.value))
}

function close() {
  visible.value = false
}

function focusInput() {
  nextTick(() => {
    inputRef.value?.focus()
  })
}

function navigateTo(result: SearchResult) {
  visible.value = false
  window.open(`/doc/${result.id}`, '_blank')
}

function getTypeLabel(type: string): string {
  const map: Record<string, string> = {
    doc: t('admin.docType'),
    sheet: t('admin.sheetType'),
    slide: t('admin.slideType'),
    mindnote: t('admin.mindnoteType'),
    bitable: t('admin.bitableType'),
    survey: t('admin.surveyType'),
    folder: t('home.folder'),
    code: t('document.codeEditor'),
    image: t('home.canvas'),
    file: t('home.other'),
  }
  return map[type] || t('admin.docType')
}

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const days = Math.floor(diff / 86400000)
  if (days < 1) return t('search.today')
  if (days < 2) return t('search.yesterday')
  if (days < 7) return t('search.daysAgo', { count: days })
  return date.toLocaleDateString('zh-CN')
}

// Reset query when dialog closes
watch(visible, (val) => {
  if (!val) {
    query.value = ''
    results.value = []
    totalCount.value = 0
  }
})
</script>

<style scoped>
:deep(.global-search-dialog .el-dialog__body) {
  padding: 0;
}
:deep(.global-search-dialog .el-dialog__header) {
  display: none;
}

.search-container {
  max-height: 500px;
  overflow-y: auto;
}

.search-input-wrapper {
  display: flex;
  align-items: center;
  padding: 14px 16px;
  border-bottom: 1px solid #eee;
  gap: 10px;
}
.search-input-wrapper .el-icon {
  font-size: 18px;
  color: #8592a6;
  flex-shrink: 0;
}
.search-input-wrapper input {
  flex: 1;
  border: none;
  outline: none;
  font-size: 16px;
  color: var(--kx-text-primary, #1f2329);
  background: transparent;
}
.search-input-wrapper input::placeholder {
  color: #bbbfc4;
}
.search-shortcut {
  font-size: 11px;
  color: #8592a6;
  background: #f5f6f7;
  padding: 2px 6px;
  border-radius: 4px;
  flex-shrink: 0;
}

/* 搜索历史 */
.search-history {
  padding: 8px 0;
}
.search-history .section-title {
  padding: 6px 16px;
  font-size: 12px;
  color: #8592a6;
  font-weight: 500;
}
.history-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 16px;
  cursor: pointer;
  font-size: 14px;
  color: var(--kx-text-primary, #1f2329);
  transition: background 0.15s;
}
.history-item:hover {
  background: #f5f7fa;
}
.history-item .el-icon {
  color: #8592a6;
  font-size: 14px;
}
.history-item .remove-btn {
  margin-left: auto;
  opacity: 0;
  transition: opacity 0.15s;
  font-size: 12px;
}
.history-item:hover .remove-btn {
  opacity: 1;
}

/* 搜索结果 */
.search-results {
  padding: 4px 0;
}
.results-count {
  padding: 6px 16px;
  font-size: 12px;
  color: #8592a6;
}
.result-item {
  padding: 10px 16px;
  cursor: pointer;
  border-bottom: 1px solid #f5f5f5;
  transition: background 0.15s;
}
.result-item:hover {
  background: #f5f7fa;
}
.result-item:last-child {
  border-bottom: none;
}
.result-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--kx-text-primary, #1f2329);
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.result-snippet {
  color: #666;
  font-size: 13px;
  margin-top: 4px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.result-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 6px;
  font-size: 12px;
  color: #8592a6;
}
.result-type {
  background: #f0f1f2;
  padding: 1px 6px;
  border-radius: 3px;
}

:deep(mark) {
  background: #fff3cd;
  padding: 0 2px;
  border-radius: 2px;
  color: inherit;
}

/* 无结果 */
.no-results {
  padding: 40px 16px;
  text-align: center;
  color: #8592a6;
  font-size: 14px;
}

/* 加载状态 */
.search-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 40px 16px;
  color: #8592a6;
  font-size: 14px;
}
</style>
