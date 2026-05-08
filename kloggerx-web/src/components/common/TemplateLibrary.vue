<template>
  <div class="tpl-lib" :class="[`tpl-lib--${mode}`]">
    <!-- Page mode header -->
    <div v-if="mode === 'page'" class="tpl-header">
      <div class="tpl-header-left">
        <h2 class="tpl-title">{{ t('template.title') }}</h2>
        <span class="tpl-subtitle">{{ t('template.subtitle') }}</span>
      </div>
      <div v-if="computedShowSearch" class="tpl-header-right">
        <el-input
          v-model="keyword"
          :placeholder="t('template.searchPlaceholder')"
          prefix-icon="Search"
          clearable
          style="width: 260px"
        />
      </div>
    </div>

    <div class="tpl-body">
      <!-- Page mode: sidebar categories -->
      <div v-if="mode === 'page'" class="tpl-sidebar">
        <div
          class="tpl-cat-item"
          :class="{ active: activeCategory === '' }"
          @click="activeCategory = ''"
        >{{ t('template.allTemplates') }}</div>
        <div
          v-for="cat in categories"
          :key="cat"
          class="tpl-cat-item"
          :class="{ active: activeCategory === cat }"
          @click="activeCategory = cat"
        >{{ cat }}</div>
      </div>

      <!-- Embed mode: top tab categories -->
      <div v-if="mode === 'embed'" class="tpl-tabs">
        <div
          class="tpl-tab"
          :class="{ active: activeCategory === '' }"
          @click="activeCategory = ''"
        >{{ t('template.all') }}</div>
        <div
          v-for="cat in categories"
          :key="cat"
          class="tpl-tab"
          :class="{ active: activeCategory === cat }"
          @click="activeCategory = cat"
        >{{ cat }}</div>
      </div>

      <!-- Template grid -->
      <div class="tpl-content">
        <div v-loading="loading" class="tpl-grid">
          <div
            v-for="tpl in filteredTemplates"
            :key="tpl.id"
            class="tpl-card"
            @click="handleCardClick(tpl)"
          >
            <div class="tpl-card-header">
              <div class="tpl-card-icon" :style="{ background: getTypeBg(tpl.type) }">
                <el-icon :size="mode === 'page' ? 28 : 32" :color="getTypeColor(tpl.type)">
                  <component :is="getTypeIcon(tpl.type)" />
                </el-icon>
              </div>
              <div class="tpl-card-title">{{ tpl.name }}</div>
              <el-icon
                class="tpl-card-star"
                :class="{ 'is-favorited': isFavorite(tpl) }"
                :size="16"
                @click="toggleFavorite($event, tpl)"
              >
                <StarFilled v-if="isFavorite(tpl)" />
                <Star v-else />
              </el-icon>
            </div>
            <div class="tpl-card-desc">
              {{ tpl.preview || tpl.description || t('template.noPreview') }}
            </div>
            <div class="tpl-card-footer">
              <el-tag size="small" type="info">{{ tpl.category || t('template.uncategorized') }}</el-tag>
              <el-tag size="small" :type="getTypeTagType(tpl.type)">{{ getTypeLabel(tpl.type) }}</el-tag>
            </div>
          </div>

          <div v-if="!loading && filteredTemplates.length === 0" class="tpl-empty">
            <el-empty :description="t('template.noMatch')" />
          </div>
        </div>
      </div>
    </div>

    <!-- Preview Dialog (page mode only) -->
    <el-dialog
      v-if="computedShowPreview"
      v-model="showPreviewDialog"
      :title="selectedTpl?.name || t('template.preview')"
      width="700px"
      destroy-on-close
    >
      <div v-if="selectedTpl" class="preview-content">
        <div class="preview-info">
          <el-tag size="small" type="info">{{ selectedTpl.category || t('template.uncategorized') }}</el-tag>
          <el-tag size="small" :type="getTypeTagType(selectedTpl.type)">{{ getTypeLabel(selectedTpl.type) }}</el-tag>
        </div>
        <div v-if="selectedTpl.description" class="preview-description">
          {{ selectedTpl.description }}
        </div>
        <div class="preview-body">
          <div v-if="previewContent" class="preview-text">
            {{ previewContent }}
          </div>
          <div v-else class="preview-empty">
            <el-empty :description="t('template.noContent')" :image-size="80" />
            <p class="preview-hint">{{ t('template.previewHint') }}</p>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="showPreviewDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleUseFromPreview">{{ t('template.useTemplate') }}</el-button>
      </template>
    </el-dialog>

    <!-- Use Template Dialog (page mode) -->
    <el-dialog
      v-if="computedShowPreview"
      v-model="showUseDialog"
      :title="t('template.useTemplateTitle', { name: selectedTpl?.name || '' })"
      width="420px"
      destroy-on-close
    >
      <el-form label-position="top">
        <el-form-item :label="t('template.docTitle')">
          <el-input v-model="newTitle" :placeholder="selectedTpl?.name || ''" maxlength="100" />
        </el-form-item>
        <el-form-item :label="t('template.saveLocation')">
          <el-select v-model="targetParentId" :placeholder="t('template.rootDir')" clearable style="width: 100%">
            <el-option :value="undefined" :label="t('template.rootDir')" />
            <el-option v-for="f in folders" :key="f.id" :value="f.id" :label="f.title" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showUseDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="creating" @click="confirmUse">{{ t('template.createDoc') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Star, StarFilled } from '@element-plus/icons-vue'
import { getTemplates, useTemplate as apiUseTemplate, favoriteTemplate, unfavoriteTemplate, type Template } from '@/api/modules/template'
import { getDocumentTree } from '@/api/modules/document'
import type { Document } from '@/types'

const { t } = useI18n()

// ---- Props & Emits ----
interface Props {
  mode?: 'page' | 'embed'
  showSearch?: boolean
  showPreview?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  mode: 'page',
  showSearch: undefined,
  showPreview: undefined,
})

const emit = defineEmits<{
  (e: 'select', template: Template): void
  (e: 'use', payload: { template: Template; title: string; parentId?: number }): void
}>()

// ---- Computed defaults based on mode ----
const computedShowSearch = computed(() =>
  props.showSearch !== undefined ? props.showSearch : props.mode === 'page'
)
const computedShowPreview = computed(() =>
  props.showPreview !== undefined ? props.showPreview : props.mode === 'page'
)

// ---- State ----
const loading = ref(false)
const keyword = ref('')
const activeCategory = ref('')
const templates = ref<Template[]>([])
const categories = ref<string[]>([])
const favoriteIds = ref<number[]>([])
const folders = ref<Document[]>([])
const showPreviewDialog = ref(false)
const showUseDialog = ref(false)
const selectedTpl = ref<Template | null>(null)
const newTitle = ref('')
const targetParentId = ref<number | undefined>(undefined)
const creating = ref(false)
const previewContent = ref('')

// ---- Type maps ----
const typeNameMap = computed<Record<string, string>>(() => ({
  doc: t('template.type.doc'), sheet: t('template.type.sheet'), slide: t('template.type.slide'), mindnote: t('template.type.mindnote'),
  bitable: t('template.type.bitable'), survey: t('template.type.survey'),
}))
const typeIconMap: Record<string, string> = {
  doc: 'Document', sheet: 'Grid', slide: 'Monitor',
  mindnote: 'Share', bitable: 'Tickets', survey: 'Notebook',
}
const typeColorMap: Record<string, string> = {
  doc: '#409eff', sheet: '#67c23a', slide: '#e6a23c',
  mindnote: '#9b59b6', bitable: '#00b8d9', survey: '#f54a45',
}
const typeBgMap: Record<string, string> = {
  doc: '#e8f0fe', sheet: '#e6f7ef', slide: '#fff3e0',
  mindnote: '#f3edfd', bitable: '#e0f7fa', survey: '#fce4e4',
}

// ---- Helper functions ----
function getTypeIcon(type: string): string {
  return typeIconMap[type] || 'Document'
}

function getTypeColor(type: string): string {
  return typeColorMap[type] || '#409eff'
}

function getTypeBg(type: string): string {
  return typeBgMap[type] || '#f5f5f5'
}

function getTypeLabel(type: string): string {
  return typeNameMap.value[type] || type
}

function getTypeTagType(type: string): '' | 'success' | 'warning' | 'info' | 'danger' {
  const m: Record<string, '' | 'success' | 'warning' | 'info' | 'danger'> = {
    doc: '', sheet: 'success', slide: 'warning', mindnote: '', bitable: 'info', survey: 'danger',
  }
  return m[type] || ''
}

function extractTextFromContent(content: string): string {
  if (!content || content === '{}' || content === '""') return ''
  try {
    const doc = JSON.parse(content)
    return extractTextFromNode(doc)
  } catch {
    return content.replace(/<[^>]*>/g, '').trim()
  }
}

function extractTextFromNode(node: any): string {
  if (!node) return ''
  const parts: string[] = []
  if (node.text) parts.push(node.text)
  if (Array.isArray(node.content)) {
    for (const child of node.content) {
      const childText = extractTextFromNode(child)
      if (childText) parts.push(childText)
    }
  }
  return parts.join(' ')
}

// ---- Filtered list ----
const filteredTemplates = computed(() => {
  let list = templates.value
  if (activeCategory.value) list = list.filter(t => t.category === activeCategory.value)
  if (keyword.value) {
    const kw = keyword.value
    list = list.filter(t =>
      t.name.includes(kw) ||
      (t.description && t.description.includes(kw)) ||
      (t.category && t.category.includes(kw))
    )
  }
  return list
})

// ---- Data loading ----
async function loadTemplates() {
  loading.value = true
  try {
    const res = await getTemplates() as any
    templates.value = res.data?.list || []
    categories.value = res.data?.categories || []
    favoriteIds.value = res.data?.favoriteIds || []
  } finally {
    loading.value = false
  }
}

async function loadFolders() {
  const res = await getDocumentTree(null) as any
  folders.value = (res.data || []).filter((d: Document) => d.type === 'folder')
}

// ---- Interactions ----
function handleCardClick(tpl: Template) {
  if (props.mode === 'embed') {
    emit('select', tpl)
  } else {
    previewTemplate(tpl)
  }
}

function isFavorite(tpl: Template): boolean {
  return favoriteIds.value.includes(tpl.id)
}

async function toggleFavorite(event: Event, tpl: Template) {
  event.stopPropagation()
  try {
    if (isFavorite(tpl)) {
      await unfavoriteTemplate(tpl.id)
      favoriteIds.value = favoriteIds.value.filter(id => id !== tpl.id)
      ElMessage.success(t('template.unfavorited'))
    } else {
      await favoriteTemplate(tpl.id)
      favoriteIds.value.push(tpl.id)
      ElMessage.success(t('template.favorited'))
    }
  } catch {
    ElMessage.error(t('common.operationFailed'))
  }
}

function previewTemplate(tpl: Template) {
  selectedTpl.value = tpl
  newTitle.value = tpl.name
  targetParentId.value = undefined

  if (tpl.content && tpl.content !== '{}' && tpl.content !== '""') {
    previewContent.value = extractTextFromContent(tpl.content)
  } else if (tpl.preview) {
    previewContent.value = tpl.preview
  } else {
    previewContent.value = ''
  }

  showPreviewDialog.value = true
}

function handleUseFromPreview() {
  showPreviewDialog.value = false
  showUseDialog.value = true
}

async function confirmUse() {
  if (!selectedTpl.value) return
  creating.value = true
  try {
    const title = newTitle.value || selectedTpl.value.name
    const parentId = targetParentId.value
    const res = await apiUseTemplate(selectedTpl.value.id, {
      title,
      parentId: parentId || null,
    }) as any

    emit('use', { template: selectedTpl.value, title, parentId })

    ElMessage.success(t('template.docCreated'))
    showUseDialog.value = false
    if (res.data?.id) {
      window.open(`/doc/${res.data.id}`, '_blank')
    }
  } catch {
    ElMessage.error(t('template.createFailed'))
  } finally {
    creating.value = false
  }
}

// ---- Expose for parent to trigger reload ----
function reload() {
  loadTemplates()
}

defineExpose({ reload })

// ---- Lifecycle ----
onMounted(() => {
  loadTemplates()
  if (props.mode === 'page') {
    loadFolders()
  }
})
</script>

<style scoped>
/* ========== Layout ========== */
.tpl-lib {
  display: flex;
  flex-direction: column;
}
.tpl-lib--page {
  height: 100%;
  background: #f7f8fa;
}
.tpl-lib--embed {
  min-height: 400px;
}

/* ========== Page Header ========== */
.tpl-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  background: #fff;
  border-bottom: 1px solid #e8e8e8;
}
.tpl-header-left {
  display: flex;
  align-items: baseline;
  gap: 12px;
}
.tpl-title {
  font-size: 18px;
  font-weight: 600;
  margin: 0;
  color: #1d2129;
}
.tpl-subtitle {
  font-size: 13px;
  color: #86909c;
}
.tpl-header-right {
  flex-shrink: 0;
}

/* ========== Body ========== */
.tpl-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}
.tpl-lib--embed .tpl-body {
  flex-direction: column;
}

/* ========== Sidebar (page mode) ========== */
.tpl-sidebar {
  width: 160px;
  flex-shrink: 0;
  background: #fff;
  border-right: 1px solid #e8e8e8;
  padding: 16px 0;
  overflow-y: auto;
}
.tpl-cat-item {
  padding: 10px 20px;
  font-size: 14px;
  color: #4e5969;
  cursor: pointer;
  border-radius: 0 20px 20px 0;
  margin-right: 8px;
  transition: background 0.15s;
}
.tpl-cat-item:hover {
  background: #f2f3f5;
}
.tpl-cat-item.active {
  background: #e8f0fe;
  color: #409eff;
  font-weight: 600;
}

/* ========== Tabs (embed mode) ========== */
.tpl-tabs {
  display: flex;
  gap: 4px;
  padding: 0 0 12px;
  flex-shrink: 0;
  flex-wrap: wrap;
}
.tpl-tab {
  padding: 6px 14px;
  cursor: pointer;
  border-radius: 6px;
  font-size: 14px;
  color: #4e5969;
  transition: all 0.15s;
}
.tpl-tab:hover {
  background: rgba(0, 0, 0, 0.04);
}
.tpl-tab.active {
  background: rgba(64, 158, 255, 0.08);
  color: #409eff;
  font-weight: 500;
}

/* ========== Content & Grid ========== */
.tpl-content {
  flex: 1;
  overflow-y: auto;
}
.tpl-lib--page .tpl-content {
  padding: 24px 32px;
}
.tpl-grid {
  display: grid;
  gap: 20px;
}
.tpl-lib--page .tpl-grid {
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
}
.tpl-lib--embed .tpl-grid {
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 16px;
}

/* ========== Card ========== */
.tpl-card {
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  cursor: pointer;
  border: 1.5px solid transparent;
  transition: all 0.18s;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
}
.tpl-lib--embed .tpl-card {
  border-radius: 8px;
  padding: 12px;
  gap: 8px;
  border-width: 1px;
  border-color: #e8e8e8;
}
.tpl-card:hover {
  border-color: #409eff;
  box-shadow: 0 4px 16px rgba(64, 158, 255, 0.12);
  transform: translateY(-2px);
}
.tpl-lib--embed .tpl-card:hover {
  transform: none;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

/* ========== Card Inner ========== */
.tpl-card-header {
  display: flex;
  align-items: center;
  gap: 12px;
}
.tpl-card-icon {
  width: 48px;
  height: 48px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.tpl-lib--embed .tpl-card-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
}
.tpl-card-title {
  font-size: 15px;
  font-weight: 600;
  color: #1f2329;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
.tpl-lib--embed .tpl-card-title {
  font-size: 13px;
  font-weight: 500;
}
.tpl-card-desc {
  font-size: 13px;
  color: #8f959e;
  line-height: 1.6;
  min-height: 42px;
  max-height: 84px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
}
.tpl-lib--embed .tpl-card-desc {
  font-size: 12px;
  min-height: 0;
  max-height: 40px;
  -webkit-line-clamp: 2;
}
.tpl-card-footer {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  margin-top: auto;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
}

/* ========== Empty ========== */
.tpl-empty {
  grid-column: 1 / -1;
  display: flex;
  justify-content: center;
  padding: 40px 0;
}

/* ========== Preview Dialog ========== */
.preview-content {
  max-height: 60vh;
  overflow-y: auto;
}
.preview-info {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}
.preview-description {
  font-size: 14px;
  color: #666;
  margin-bottom: 16px;
  padding: 12px;
  background: #f7f8fa;
  border-radius: 8px;
}
.preview-body {
  min-height: 200px;
}
.preview-text {
  font-size: 14px;
  line-height: 1.8;
  color: #1f2329;
  white-space: pre-wrap;
  word-break: break-word;
  padding: 16px;
  background: #fafbfc;
  border-radius: 8px;
  border: 1px solid #e8e8e8;
}
.preview-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px 0;
}
.preview-hint {
  font-size: 13px;
  color: #8f959e;
  margin-top: 8px;
}

/* ========== Favorite Star ========== */
.tpl-card-star {
  margin-left: auto;
  cursor: pointer;
  color: #c0c4cc;
  flex-shrink: 0;
  transition: color 0.2s;
}
.tpl-card-star:hover {
  color: #f7ba2a;
}
.tpl-card-star.is-favorited {
  color: #f7ba2a;
}
</style>
