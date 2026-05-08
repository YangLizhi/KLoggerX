<template>
  <div class="enhanced-search-bar">
    <!-- Search Input -->
    <div class="search-input-wrapper">
      <el-icon class="search-icon"><Search /></el-icon>
      <input
        ref="inputRef"
        v-model="searchKeyword"
        type="text"
        class="search-input"
        :placeholder="placeholder || t('search.placeholder')"
        @input="handleInput"
        @keyup.enter="handleSearch"
        @focus="handleFocus"
        @blur="handleBlur"
        @keydown="handleKeyDown"
      />
      <el-icon v-if="searchKeyword" class="clear-icon" @click="handleClear"><Close /></el-icon>
      <el-icon v-if="loading" class="loading-icon"><Loading /></el-icon>
    </div>

    <!-- Search History Dropdown -->
    <div v-if="showHistory" class="search-dropdown history-dropdown">
      <div class="dropdown-header">
        <span class="dropdown-title">{{ t('search.history') }}</span>
      </div>
      <div
        v-for="(item, index) in searchHistory"
        :key="'history-' + index"
        class="dropdown-item"
        :class="{ selected: selectedSuggestionIndex === index }"
        @click="selectHistory(item)"
      >
        <el-icon class="dropdown-icon"><Clock /></el-icon>
        <span class="dropdown-text">{{ item }}</span>
        <el-icon class="dropdown-delete" @click.stop="removeSearchHistory(item)"><Close /></el-icon>
      </div>
      <div class="dropdown-footer" @click="clearSearchHistory">
        <span>{{ t('search.clearHistory') }}</span>
      </div>
    </div>

    <!-- Search Suggestions Dropdown -->
    <div v-if="showSuggestions" class="search-dropdown suggestions-dropdown">
      <div
        v-for="(item, index) in suggestions"
        :key="'suggestion-' + index"
        class="dropdown-item"
        :class="{ selected: selectedSuggestionIndex === index }"
        @click="selectSuggestion(item)"
      >
        <el-icon class="dropdown-icon"><Search /></el-icon>
        <span class="dropdown-text" v-html="highlightKeyword(item, searchKeyword)"></span>
      </div>
    </div>

    <!-- Filter Tags -->
    <div v-if="showFilterTags" class="filter-tags">
      <span
        class="filter-tag"
        :class="{ active: filterType === 'all' }"
        @click="filterType = 'all'"
      >{{ t('search.filterAll') }}</span>
      <span
        class="filter-tag"
        :class="{ active: filterType === 'document' }"
        @click="filterType = 'document'"
      >{{ t('search.filterDocument') }}</span>
      <span
        class="filter-tag"
        :class="{ active: filterType === 'knowledge' }"
        @click="filterType = 'knowledge'"
      >{{ t('search.filterKnowledge') }}</span>
    </div>

    <!-- Search Results Panel -->
    <div v-if="showResults" class="search-results-panel">
      <!-- Loading State -->
      <div v-if="loading" class="search-loading">
        <el-icon class="loading-spinner"><Loading /></el-icon>
        <span>{{ t('search.searching') }}</span>
      </div>

      <!-- Empty State -->
      <div v-else-if="!hasResults" class="search-empty">
        <el-icon><Search /></el-icon>
        <span v-if="searchKeyword">{{ t('search.noResults') }}</span>
        <span v-else>{{ t('search.inputHint') }}</span>
      </div>

      <!-- Results -->
      <template v-else>
        <!-- Document Results -->
        <div v-if="documentResults.length > 0 && (filterType === 'all' || filterType === 'document')" class="result-group">
          <div class="group-header">
            <span class="group-title">{{ t('search.resultType.document') }}</span>
            <span class="group-count">{{ t('search.resultCount', { count: documentResults.length }) }}</span>
          </div>
          <div
            v-for="doc in documentResults"
            :key="'doc-' + doc.id"
            class="result-item"
            @click="handleSelect('document', doc)"
          >
            <div class="result-icon">
              <el-icon :color="getTypeColor(doc.type, doc.fileExt)" :size="20">
                <component :is="getTypeIcon(doc.type, doc.fileExt)" />
              </el-icon>
            </div>
            <div class="result-content">
              <div class="result-title" v-html="highlightText(doc.title || doc.originalName || t('search.unnamed'), searchKeyword)"></div>
              <div class="result-summary" v-if="doc.content">
                <span v-html="highlightText(truncateText(doc.content, 200), searchKeyword)"></span>
              </div>
              <div class="result-meta">
                <el-tag size="small" :type="getTypeTagType(doc.type)" disable-transitions>
                  {{ getTypeName(doc) }}
                </el-tag>
                <span class="result-time">{{ formatDate(doc.updatedAt) }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Knowledge Results -->
        <div v-if="knowledgeResults.length > 0 && (filterType === 'all' || filterType === 'knowledge')" class="result-group">
          <div class="group-header">
            <span class="group-title">{{ t('search.resultType.knowledge') }}</span>
            <span class="group-count">{{ t('search.resultCount', { count: knowledgeResults.length }) }}</span>
          </div>
          <div
            v-for="kb in knowledgeResults"
            :key="'kb-' + kb.id"
            class="result-item"
            @click="handleSelect('knowledge', kb)"
          >
            <div class="result-icon">
              <el-icon color="#9254de" :size="20"><Collection /></el-icon>
            </div>
            <div class="result-content">
              <div class="result-title" v-html="highlightText(kb.name || t('search.unnamedKb'), searchKeyword)"></div>
              <div class="result-summary" v-if="kb.description">
                <span v-html="highlightText(truncateText(kb.description, 200), searchKeyword)"></span>
              </div>
              <div class="result-meta">
                <el-tag size="small" type="primary" disable-transitions>{{ t('search.resultType.knowledge') }}</el-tag>
                <span class="result-time">{{ formatDate(kb.updatedAt) }}</span>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { searchDocuments, getSearchSuggestions } from '@/api/modules/document'
import { searchKnowledge } from '@/api/modules/knowledge'
import type { Document, KnowledgeBase } from '@/types'

const { t } = useI18n()

interface Props {
  placeholder?: string
}

interface SearchResult {
  type: 'document' | 'knowledge'
  data: Document | KnowledgeBase
}

const props = withDefaults(defineProps<Props>(), {
  placeholder: undefined
})

const emit = defineEmits<{
  select: [result: SearchResult]
}>()

const router = useRouter()

// Search History Constants
const HISTORY_KEY = 'kloggerx_search_history'
const MAX_HISTORY = 20

// State
const inputRef = ref<HTMLInputElement>()
const searchKeyword = ref('')
const filterType = ref<'all' | 'document' | 'knowledge'>('all')
const loading = ref(false)
const isFocused = ref(false)
const documentResults = ref<Document[]>([])
const knowledgeResults = ref<KnowledgeBase[]>([])
const searchHistory = ref<string[]>([])
const suggestions = ref<string[]>([])
const selectedSuggestionIndex = ref(-1)
let debounceTimer: ReturnType<typeof setTimeout> | null = null
let suggestionDebounceTimer: ReturnType<typeof setTimeout> | null = null

// Computed
const showFilterTags = computed(() => isFocused.value || searchKeyword.value)
const showResults = computed(() => isFocused.value || searchKeyword.value)
const hasResults = computed(() => documentResults.value.length > 0 || knowledgeResults.value.length > 0)
const showHistory = computed(() => isFocused.value && !searchKeyword.value.trim() && searchHistory.value.length > 0)
const showSuggestions = computed(() => isFocused.value && searchKeyword.value.trim() && suggestions.value.length > 0)

// Type mappings (copied from HomePage.vue for consistency)
const typeMap: Record<string, { icon: string; color: string }> = {
  folder: { icon: 'Folder', color: '#f5a623' },
  doc: { icon: 'Document', color: '#3370ff' },
  sheet: { icon: 'Grid', color: '#36b37e' },
  slide: { icon: 'Monitor', color: '#ff7d00' },
  mindnote: { icon: 'Share', color: '#9254de' },
  bitable: { icon: 'Tickets', color: '#00b8d9' },
  survey: { icon: 'Notebook', color: '#f54a45' },
  file: { icon: 'Document', color: '#888' },
  image: { icon: 'Picture', color: '#36b37e' },
  code: { icon: 'Memo', color: '#3370ff' },
}

const extIconMap: Record<string, { icon: string; color: string }> = {
  pdf: { icon: 'Document', color: '#f54a45' },
  doc: { icon: 'Document', color: '#2b579a' },
  docx: { icon: 'Document', color: '#2b579a' },
  txt: { icon: 'Document', color: '#666' },
  md: { icon: 'Memo', color: '#3370ff' },
  xls: { icon: 'Grid', color: '#217346' },
  xlsx: { icon: 'Grid', color: '#217346' },
  csv: { icon: 'Grid', color: '#36b37e' },
  ppt: { icon: 'Monitor', color: '#d24726' },
  pptx: { icon: 'Monitor', color: '#d24726' },
  png: { icon: 'Picture', color: '#36b37e' },
  jpg: { icon: 'Picture', color: '#36b37e' },
  jpeg: { icon: 'Picture', color: '#36b37e' },
  gif: { icon: 'Picture', color: '#ff7d00' },
  svg: { icon: 'Picture', color: '#ff7d00' },
}

const typeNameMap = computed<Record<string, string>>(() => ({
  folder: t('search.typeName.folder'), doc: t('search.typeName.doc'), sheet: t('search.typeName.sheet'), slide: t('search.typeName.slide'),
  mindnote: t('search.typeName.mindnote'), bitable: t('search.typeName.bitable'), survey: t('search.typeName.survey'),
  file: t('search.typeName.other'), image: t('search.typeName.image'), code: t('search.typeName.code'),
}))

const extTypeMap = computed<Record<string, string>>(() => ({
  txt: t('search.extType.txt'), doc: t('search.extType.word'), docx: t('search.extType.word'), pdf: t('search.extType.pdf'),
  xls: t('search.extType.excel'), xlsx: t('search.extType.excel'), csv: t('search.extType.csv'),
  ppt: t('search.extType.ppt'), pptx: t('search.extType.ppt'),
  png: t('search.extType.png'), jpg: t('search.extType.jpeg'), jpeg: t('search.extType.jpeg'), gif: t('search.extType.gif'),
  mp3: t('search.extType.audio'), mp4: t('search.extType.video'), zip: t('search.extType.archive'), rar: t('search.extType.archive'),
  md: t('search.extType.markdown'), json: t('search.extType.json'),
}))

// Methods
function getTypeIcon(type: string, ext?: string): string {
  if (ext) {
    const e = ext.replace('.', '').toLowerCase()
    if (extIconMap[e]) return extIconMap[e].icon
  }
  return typeMap[type]?.icon || 'Document'
}

function getTypeColor(type: string, ext?: string): string {
  if (ext) {
    const e = ext.replace('.', '').toLowerCase()
    if (extIconMap[e]) return extIconMap[e].color
  }
  return typeMap[type]?.color || '#888'
}

function getTypeName(doc: Document): string {
  const ext = (doc.fileExt || '').replace('.', '').toLowerCase()
  if (ext && extTypeMap.value[ext]) return extTypeMap.value[ext]
  if (typeNameMap.value[doc.type]) return typeNameMap.value[doc.type]
  return t('search.typeName.other')
}

function getTypeTagType(type: string): '' | 'success' | 'warning' | 'info' | 'danger' {
  const m: Record<string, '' | 'success' | 'warning' | 'info' | 'danger'> = {
    folder: 'warning', doc: '', sheet: 'success', slide: 'warning',
    mindnote: '', bitable: 'info', survey: 'danger', file: 'info',
  }
  return m[type] || 'info'
}

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const now = new Date()
  const isToday = d.toDateString() === now.toDateString()
  if (isToday) return `${t('search.today')} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
  return `${d.getFullYear()}/${d.getMonth() + 1}/${d.getDate()}`
}

function highlightText(text: string, keyword: string): string {
  if (!keyword || !text) return text
  const escapedKeyword = keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const regex = new RegExp(`(${escapedKeyword})`, 'gi')
  return text.replace(regex, '<mark>$1</mark>')
}

function truncateText(text: string, maxLength: number): string {
  if (!text) return ''
  return text.length > maxLength ? text.slice(0, maxLength) + '...' : text
}

// Search History Management
function getSearchHistory(): string[] {
  try {
    const data = localStorage.getItem(HISTORY_KEY)
    return data ? JSON.parse(data) : []
  } catch {
    return []
  }
}

function addSearchHistory(query: string) {
  if (!query.trim()) return
  const history = getSearchHistory().filter(h => h !== query)
  history.unshift(query)
  localStorage.setItem(HISTORY_KEY, JSON.stringify(history.slice(0, MAX_HISTORY)))
  searchHistory.value = getSearchHistory()
}

function removeSearchHistory(query: string) {
  const history = getSearchHistory().filter(h => h !== query)
  localStorage.setItem(HISTORY_KEY, JSON.stringify(history))
  searchHistory.value = history
}

function clearSearchHistory() {
  localStorage.removeItem(HISTORY_KEY)
  searchHistory.value = []
}

// Search Suggestions
async function fetchSuggestions() {
  const q = searchKeyword.value.trim()
  if (!q) {
    suggestions.value = []
    return
  }
  try {
    const res = await getSearchSuggestions(q)
    suggestions.value = (res as any).data?.suggestions || []
    selectedSuggestionIndex.value = -1
  } catch (error) {
    console.error('Failed to fetch suggestions:', error)
    suggestions.value = []
  }
}

function handleKeyDown(e: KeyboardEvent) {
  const list = suggestions.value.length > 0 ? suggestions.value : searchHistory.value
  if (list.length === 0) return

  if (e.key === 'ArrowDown') {
    e.preventDefault()
    selectedSuggestionIndex.value = Math.min(selectedSuggestionIndex.value + 1, list.length - 1)
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    selectedSuggestionIndex.value = Math.max(selectedSuggestionIndex.value - 1, -1)
  } else if (e.key === 'Enter' && selectedSuggestionIndex.value >= 0) {
    e.preventDefault()
    const selected = list[selectedSuggestionIndex.value]
    searchKeyword.value = selected
    handleSearch()
  }
}

function selectHistory(query: string) {
  searchKeyword.value = query
  handleSearch()
}

function selectSuggestion(suggestion: string) {
  searchKeyword.value = suggestion
  handleSearch()
}

function highlightKeyword(text: string, keyword: string): string {
  if (!keyword || !text) return text
  const escapedKeyword = keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const regex = new RegExp(`(${escapedKeyword})`, 'gi')
  return text.replace(regex, '<strong>$1</strong>')
}

function handleInput() {
  if (debounceTimer) {
    clearTimeout(debounceTimer)
  }
  if (suggestionDebounceTimer) {
    clearTimeout(suggestionDebounceTimer)
  }
  debounceTimer = setTimeout(() => {
    performSearch()
  }, 300)
  suggestionDebounceTimer = setTimeout(() => {
    fetchSuggestions()
  }, 300)
}

function handleSearch() {
  if (debounceTimer) {
    clearTimeout(debounceTimer)
  }
  performSearch()
}

async function performSearch() {
  const keyword = searchKeyword.value.trim()
  if (!keyword) {
    documentResults.value = []
    knowledgeResults.value = []
    return
  }

  // Add to search history
  addSearchHistory(keyword)

  loading.value = true
  try {
    // Execute document and knowledge search in parallel
    const [docRes, kbRes] = await Promise.allSettled([
      searchDocuments({ keyword, page: 1, pageSize: 20 }),
      searchKnowledge({ keyword, page: 1, pageSize: 20 })
    ])

    // Process document results
    if (docRes.status === 'fulfilled') {
      documentResults.value = (docRes.value as any).data?.list || []
    } else {
      documentResults.value = []
    }

    // Process knowledge results
    if (kbRes.status === 'fulfilled') {
      knowledgeResults.value = (kbRes.value as any).data?.list || []
    } else {
      knowledgeResults.value = []
    }
  } catch (error) {
    console.error('Search error:', error)
    documentResults.value = []
    knowledgeResults.value = []
  } finally {
    loading.value = false
  }
}

function handleClear() {
  searchKeyword.value = ''
  documentResults.value = []
  knowledgeResults.value = []
  inputRef.value?.focus()
}

function handleFocus() {
  isFocused.value = true
}

function handleBlur() {
  // Delay blur to allow click events on results
  setTimeout(() => {
    isFocused.value = false
  }, 200)
}

function handleSelect(type: 'document' | 'knowledge', data: Document | KnowledgeBase) {
  emit('select', { type, data })
  
  // Navigate to the selected item
  if (type === 'document') {
    const doc = data as Document
    if (doc.type === 'folder') {
      router.push('/documents')
    } else {
      window.open(`/doc/${doc.id}`, '_blank')
    }
  } else {
    const kb = data as KnowledgeBase
    router.push(`/knowledge/${kb.id}`)
  }
}

// Handle click outside to close results
function handleClickOutside(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (!target.closest('.enhanced-search-bar')) {
    isFocused.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  // Initialize search history from localStorage
  searchHistory.value = getSearchHistory()
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
  if (debounceTimer) {
    clearTimeout(debounceTimer)
  }
  if (suggestionDebounceTimer) {
    clearTimeout(suggestionDebounceTimer)
  }
})

// Expose methods for parent component
defineExpose({
  focus: () => inputRef.value?.focus(),
  clear: handleClear,
  search: performSearch
})
</script>

<style scoped>
.enhanced-search-bar {
  position: relative;
  width: 100%;
  max-width: 600px;
  margin: 0 auto;
}

/* Search Input */
.search-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  background: #fff;
  border: 2px solid var(--kx-border, #e4e7ed);
  border-radius: 12px;
  padding: 0 16px;
  height: 48px;
  transition: all 0.2s ease;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.search-input-wrapper:focus-within {
  border-color: var(--kx-primary, #3370ff);
  box-shadow: 0 4px 16px rgba(51, 112, 255, 0.15);
}

.search-icon {
  font-size: 20px;
  color: var(--kx-text-placeholder, #c0c4cc);
  flex-shrink: 0;
  margin-right: 10px;
}

.search-input {
  flex: 1;
  border: none;
  outline: none;
  font-size: 16px;
  color: var(--kx-text-primary, #303133);
  background: transparent;
}

.search-input::placeholder {
  color: var(--kx-text-placeholder, #c0c4cc);
}

.clear-icon,
.loading-icon {
  font-size: 18px;
  color: var(--kx-text-placeholder, #c0c4cc);
  cursor: pointer;
  flex-shrink: 0;
  margin-left: 8px;
  transition: color 0.2s;
}

.clear-icon:hover {
  color: var(--kx-text-secondary, #909399);
}

.loading-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* Filter Tags */
.filter-tags {
  display: flex;
  justify-content: center;
  gap: 12px;
  margin-top: 12px;
}

.filter-tag {
  padding: 6px 16px;
  font-size: 13px;
  color: var(--kx-text-secondary, #909399);
  background: #f5f7fa;
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.2s ease;
  user-select: none;
}

.filter-tag:hover {
  background: #e9ecf0;
}

.filter-tag.active {
  color: #fff;
  background: var(--kx-primary, #3370ff);
  font-weight: 500;
}

/* Search Results Panel */
.search-results-panel {
  position: absolute;
  top: calc(100% + 8px);
  left: 0;
  right: 0;
  background: #fff;
  border: 1px solid var(--kx-border, #e4e7ed);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  max-height: 70vh;
  overflow-y: auto;
  z-index: 1000;
  padding: 8px 0;
}

/* Loading & Empty States */
.search-loading,
.search-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  color: var(--kx-text-placeholder, #c0c4cc);
  font-size: 14px;
  gap: 12px;
}

.loading-spinner {
  font-size: 24px;
  animation: spin 1s linear infinite;
}

.search-empty .el-icon {
  font-size: 32px;
}

/* Result Group */
.result-group {
  margin-bottom: 8px;
}

.result-group:last-child {
  margin-bottom: 0;
}

.group-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  background: #fafbfc;
  border-bottom: 1px solid var(--kx-border, #e4e7ed);
  position: sticky;
  top: 0;
  z-index: 1;
}

.group-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--kx-text-secondary, #909399);
}

.group-count {
  font-size: 12px;
  color: var(--kx-text-placeholder, #c0c4cc);
}

/* Result Item */
.result-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px 16px;
  cursor: pointer;
  transition: background-color 0.15s ease;
}

.result-item:hover {
  background: rgba(51, 112, 255, 0.06);
}

.result-item:active {
  background: rgba(51, 112, 255, 0.1);
}

.result-icon {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
  border-radius: 8px;
}

.result-content {
  flex: 1;
  min-width: 0;
}

.result-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--kx-text-primary, #303133);
  margin-bottom: 4px;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-summary {
  font-size: 13px;
  color: var(--kx-text-secondary, #909399);
  line-height: 1.5;
  margin-bottom: 6px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.result-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.result-time {
  font-size: 12px;
  color: var(--kx-text-placeholder, #c0c4cc);
}

/* Highlight Mark */
:deep(mark) {
  background: #fff3cd;
  color: inherit;
  padding: 0 2px;
  border-radius: 2px;
}

/* Search Dropdown (History & Suggestions) */
.search-dropdown {
  position: absolute;
  top: calc(100% + 8px);
  left: 0;
  right: 0;
  background: #fff;
  border: 1px solid var(--kx-border, #e4e7ed);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  z-index: 1000;
  overflow: hidden;
}

.dropdown-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background: #fafbfc;
  border-bottom: 1px solid var(--kx-border, #e4e7ed);
}

.dropdown-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--kx-text-secondary, #909399);
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  cursor: pointer;
  transition: background-color 0.15s ease;
}

.dropdown-item:hover,
.dropdown-item.selected {
  background: rgba(51, 112, 255, 0.06);
}

.dropdown-icon {
  font-size: 16px;
  color: var(--kx-text-placeholder, #c0c4cc);
  flex-shrink: 0;
}

.dropdown-text {
  flex: 1;
  font-size: 14px;
  color: var(--kx-text-primary, #303133);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dropdown-text strong {
  color: var(--kx-primary, #3370ff);
  font-weight: 600;
}

.dropdown-delete {
  font-size: 14px;
  color: var(--kx-text-placeholder, #c0c4cc);
  opacity: 0;
  transition: opacity 0.15s ease, color 0.15s ease;
  flex-shrink: 0;
}

.dropdown-item:hover .dropdown-delete {
  opacity: 1;
}

.dropdown-delete:hover {
  color: var(--kx-danger, #f56c6c);
}

.dropdown-footer {
  padding: 10px 16px;
  text-align: center;
  color: var(--kx-primary, #3370ff);
  font-size: 13px;
  cursor: pointer;
  border-top: 1px solid var(--kx-border, #e4e7ed);
  transition: background-color 0.15s ease;
}

.dropdown-footer:hover {
  background: rgba(51, 112, 255, 0.06);
}

/* Responsive */
@media (max-width: 768px) {
  .enhanced-search-bar {
    max-width: 100%;
  }
  
  .search-input-wrapper {
    height: 44px;
  }
  
  .search-input {
    font-size: 15px;
  }
  
  .filter-tags {
    flex-wrap: wrap;
  }
  
  .filter-tag {
    padding: 5px 12px;
    font-size: 12px;
  }
}
</style>
