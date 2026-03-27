<template>
  <div class="knowledge-page">
    <!-- Quick Action Cards (like reference image) -->
    <div class="kb-quick-bar">
      <div class="kb-quick-item" @click="showCreate = true; createMode = 'blank'">
        <el-icon :size="20" color="#3370ff"><EditPen /></el-icon>
        <div class="kb-quick-info">
          <span class="kb-quick-title">新建</span>
          <span class="kb-quick-desc">新建文档开始协作</span>
        </div>
        <el-icon class="kb-quick-arrow"><ArrowDown /></el-icon>
      </div>
      <div class="kb-quick-item" @click="showCreate = true; createMode = 'template'">
        <el-icon :size="20" color="#ff7d00"><Files /></el-icon>
        <div class="kb-quick-info">
          <span class="kb-quick-title">模板库</span>
          <span class="kb-quick-desc">选择模板快速新建</span>
        </div>
      </div>
      <div class="kb-quick-item" @click="showCreate = true; createMode = 'blank'">
        <el-icon :size="20" color="#3370ff"><Collection /></el-icon>
        <div class="kb-quick-info">
          <span class="kb-quick-title">新建知识库</span>
          <span class="kb-quick-desc">让知识创造价值</span>
        </div>
      </div>
    </div>

    <!-- Tabs + Controls Row -->
    <div class="kb-tabs-row">
      <div class="kb-tabs">
        <span class="kb-tab" :class="{ active: activeTab === 'all' }" @click="activeTab = 'all'; fetchData()">全部</span>
        <span class="kb-tab" :class="{ active: activeTab === 'created' }" @click="activeTab = 'created'; fetchData()">我创建的</span>
        <span class="kb-tab" :class="{ active: activeTab === 'joined' }" @click="activeTab = 'joined'; fetchData()">我加入的</span>
      </div>
      <div class="kb-controls">
        <el-input
          v-model="keyword"
          placeholder="搜索知识库"
          prefix-icon="Search"
          clearable
          class="kb-search-inline"
          @keyup.enter="fetchData"
        />
        <el-dropdown trigger="click" @command="handleSortChange">
          <span class="ctrl-btn"><el-icon><Sort /></el-icon>{{ sortLabel }}</span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="updated">最近更新</el-dropdown-item>
              <el-dropdown-item command="created">最近创建</el-dropdown-item>
              <el-dropdown-item command="name">名称</el-dropdown-item>
              <el-dropdown-item command="docCount">文档数</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <span class="ctrl-btn" @click="viewMode = viewMode === 'grid' ? 'list' : 'grid'">
          <el-icon><component :is="viewMode === 'grid' ? 'List' : 'Grid'" /></el-icon>{{ viewMode === 'grid' ? '列表' : '网格' }}
        </span>
      </div>
    </div>

    <!-- Grid View -->
    <div v-if="viewMode === 'grid'" v-loading="loading" class="kb-grid">
      <div
        v-for="kb in sortedList"
        :key="kb.id"
        class="kb-card"
        @click="$router.push(`/knowledge/${kb.id}`)"
        @contextmenu.prevent="showKbMenu($event, kb)"
      >
        <div class="kb-card-cover" :style="{ background: getCoverColor(kb.id) }">
          <el-icon :size="28" color="#fff"><Collection /></el-icon>
        </div>
        <div class="kb-card-body">
          <div class="kb-card-name">{{ kb.name }}</div>
          <div class="kb-card-desc">{{ kb.description || '暂无描述' }}</div>
          <div class="kb-card-footer">
            <div class="kb-card-stats">
              <span><el-icon :size="13"><Document /></el-icon>{{ kb.docCount }}</span>
              <span><el-icon :size="13"><User /></el-icon>{{ kb.memberCount }}</span>
            </div>
            <div class="kb-card-owner">
              <el-avatar :size="20" :style="{ background: getAvatarColor(kb.ownerId) }">{{ kb.ownerName?.[0] || '?' }}</el-avatar>
            </div>
          </div>
        </div>
      </div>

      <!-- New Knowledge Base Card -->
      <div class="kb-card kb-card-new" @click="showCreate = true">
        <div class="kb-card-new-inner">
          <el-icon :size="36" color="var(--kx-text-placeholder)"><Plus /></el-icon>
          <span>新建知识库</span>
        </div>
      </div>
    </div>

    <!-- List View -->
    <div v-if="viewMode === 'list'" v-loading="loading" class="kb-list-view">
      <el-table :data="sortedList" style="width:100%" @row-click="(row: any) => $router.push(`/knowledge/${row.id}`)" @row-contextmenu="handleRowCtxMenu">
        <el-table-column label="名称" min-width="300">
          <template #default="{ row }">
            <div class="kb-list-name">
              <div class="kb-list-icon" :style="{ background: getCoverColor(row.id) }"><el-icon color="#fff" :size="14"><Collection /></el-icon></div>
              <span>{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="描述" min-width="200">
          <template #default="{ row }">
            <span class="kb-list-desc">{{ row.description || '暂无描述' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="文档数" width="100" align="center" prop="docCount" />
        <el-table-column label="成员" width="100" align="center" prop="memberCount" />
        <el-table-column label="创建者" width="120" prop="ownerName" />
        <el-table-column label="更新时间" width="160">
          <template #default="{ row }">{{ formatDate(row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column width="50" align="center">
          <template #default="{ row }">
            <el-icon class="more-btn" @click.stop="showKbMenu($event, row)"><MoreFilled /></el-icon>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div v-if="!loading && !list.length" class="kb-empty">
      <el-empty description="暂无知识库">
        <el-button type="primary" @click="showCreate = true">创建知识库</el-button>
      </el-empty>
    </div>

    <div v-if="total > 20" style="margin-top:20px;text-align:center">
      <el-pagination background layout="prev, pager, next" :total="total" :page-size="20" v-model:current-page="page" @current-change="fetchData" />
    </div>

    <!-- Knowledge Base Context Menu -->
    <div v-if="kbMenu.visible" class="context-menu" :style="{ left: kbMenu.x + 'px', top: kbMenu.y + 'px' }">
      <div class="ctx-item" @click="handleKbAction('open')"><el-icon><View /></el-icon>打开</div>
      <div class="ctx-item" @click="handleKbAction('openNew')"><el-icon><TopRight /></el-icon>在新标签页打开</div>
      <div class="ctx-sep" />
      <div class="ctx-item" @click="handleKbAction('share')"><el-icon><Share /></el-icon>分享</div>
      <div class="ctx-item" @click="handleKbAction('copyLink')"><el-icon><Link /></el-icon>复制链接</div>
      <div class="ctx-sep" />
      <div class="ctx-item" @click="handleKbAction('pin')"><el-icon><Flag /></el-icon>添加到"置顶"</div>
      <div class="ctx-item" @click="handleKbAction('favorite')"><el-icon><Star /></el-icon>收藏</div>
      <div class="ctx-sep" />
      <div class="ctx-item" @click="handleKbAction('settings')"><el-icon><Setting /></el-icon>设置</div>
      <div class="ctx-item danger" @click="handleKbAction('delete')"><el-icon><Delete /></el-icon>删除</div>
    </div>

    <!-- Create Dialog -->
    <el-dialog v-model="showCreate" :title="createMode === 'template' ? '从模板创建知识库' : '新建知识库'" width="560px" destroy-on-close>
      <!-- Template Selection -->
      <div v-if="createMode === 'template'" class="template-grid">
        <div
          v-for="tpl in templates"
          :key="tpl.id"
          class="template-card"
          :class="{ selected: selectedTemplate === tpl.id }"
          @click="selectedTemplate = tpl.id"
        >
          <div class="tpl-icon" :style="{ background: tpl.color }"><el-icon color="#fff" :size="20"><Collection /></el-icon></div>
          <div class="tpl-name">{{ tpl.name }}</div>
          <div class="tpl-desc">{{ tpl.description }}</div>
        </div>
      </div>

      <el-form :model="createForm" label-position="top" style="margin-top:16px">
        <el-form-item label="名称">
          <el-input v-model="createForm.name" placeholder="输入知识库名称" maxlength="50" show-word-limit />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="createForm.description" type="textarea" :rows="3" placeholder="描述知识库的用途（可选）" maxlength="200" show-word-limit />
        </el-form-item>
        <el-form-item label="可见范围">
          <el-radio-group v-model="createForm.visibility">
            <el-radio value="private">
              <div>仅知识库成员可见</div>
              <div class="radio-desc">只有被邀请的成员才能查看</div>
            </el-radio>
            <el-radio value="team">
              <div>团队内所有人可见</div>
              <div class="radio-desc">团队内成员均可浏览</div>
            </el-radio>
            <el-radio value="public">
              <div>所有人可见</div>
              <div class="radio-desc">所有注册用户均可浏览</div>
            </el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- Share Dialog -->
    <el-dialog v-model="showShareDialog" title="分享知识库" width="500px" destroy-on-close>
      <div class="share-section">
        <div class="share-label">分享链接</div>
        <div class="share-link-row">
          <el-input :model-value="shareLink" readonly />
          <el-button type="primary" @click="copyShareLink">复制链接</el-button>
        </div>
      </div>
      <div class="share-section">
        <div class="share-label">邀请成员</div>
        <div class="share-invite-row">
          <el-input v-model="shareInviteEmail" placeholder="输入邮箱或用户名" style="flex:1" />
          <el-select v-model="shareInviteRole" style="width:100px">
            <el-option label="可编辑" value="editor" />
            <el-option label="仅查看" value="viewer" />
          </el-select>
          <el-button type="primary" @click="handleInvite">邀请</el-button>
        </div>
      </div>
      <template #footer>
        <el-button @click="showShareDialog = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { getKnowledgeBaseList, createKnowledgeBase, deleteKnowledgeBase } from '@/api/modules/knowledge'
import type { KnowledgeBase } from '@/types'
import { ElMessage, ElMessageBox } from 'element-plus'

const router = useRouter()
const loading = ref(false)
const list = ref<KnowledgeBase[]>([])
const page = ref(1)
const total = ref(0)
const keyword = ref('')
const activeTab = ref<'created' | 'joined' | 'all'>('all')
const showCreate = ref(false)
const creating = ref(false)
const createMode = ref<'blank' | 'template'>('blank')
const selectedTemplate = ref<string | null>(null)
const viewMode = ref<'grid' | 'list'>('grid')
const sortBy = ref('updated')

const createForm = reactive({ name: '', description: '', visibility: 'private' })

const kbMenu = reactive({ visible: false, x: 0, y: 0, kb: null as KnowledgeBase | null })

// Share dialog
const showShareDialog = ref(false)
const shareLink = ref('')
const shareInviteEmail = ref('')
const shareInviteRole = ref('editor')

// Templates
const templates = [
  { id: 'team-wiki', name: '团队 Wiki', description: '适合团队知识沉淀和分享', color: '#3370ff' },
  { id: 'product-docs', name: '产品文档', description: '产品需求、PRD、设计文档', color: '#36b37e' },
  { id: 'dev-docs', name: '技术文档', description: 'API 文档、架构设计、开发规范', color: '#ff7d00' },
  { id: 'onboarding', name: '新人手册', description: '新员工入职指南和培训材料', color: '#9254de' },
  { id: 'meeting', name: '会议纪要', description: '会议记录和决议跟踪', color: '#f54a45' },
  { id: 'project', name: '项目管理', description: '项目计划、进度和复盘', color: '#00b8d9' },
]

const coverColors = ['#3370ff', '#36b37e', '#ff7d00', '#f54a45', '#9254de', '#00b8d9', '#f5a623', '#7b61ff']
const avatarColors = ['#3370ff', '#36b37e', '#ff7d00', '#f54a45', '#9254de', '#00b8d9']

function getCoverColor(id: number) {
  return coverColors[id % coverColors.length]
}
function getAvatarColor(id: number) {
  return avatarColors[id % avatarColors.length]
}
function formatDate(t: string) {
  if (!t) return ''
  const d = new Date(t)
  return `${d.getFullYear()}年${d.getMonth()+1}月${d.getDate()}日`
}

const sortLabel = computed(() => {
  const map: Record<string, string> = { updated: '最近更新', created: '最近创建', name: '名称', docCount: '文档数' }
  return map[sortBy.value] || '排序'
})

const sortedList = computed(() => {
  const arr = [...list.value]
  arr.sort((a, b) => {
    if (sortBy.value === 'name') return a.name.localeCompare(b.name)
    if (sortBy.value === 'created') return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
    if (sortBy.value === 'docCount') return (b.docCount || 0) - (a.docCount || 0)
    return new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
  })
  return arr
})

async function fetchData() {
  loading.value = true
  try {
    const res: any = await getKnowledgeBaseList({ page: page.value, pageSize: 20, keyword: keyword.value || undefined })
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } finally { loading.value = false }
}

function handleCreateCommand(cmd: string) {
  createMode.value = cmd as 'blank' | 'template'
  selectedTemplate.value = null
  createForm.name = ''
  createForm.description = ''
  createForm.visibility = 'private'
  showCreate.value = true
}

function handleSortChange(cmd: string) {
  sortBy.value = cmd
}

async function handleCreate() {
  if (!createForm.name.trim()) { ElMessage.warning('请输入名称'); return }
  creating.value = true
  try {
    await createKnowledgeBase({ name: createForm.name, description: createForm.description })
    ElMessage.success('创建成功')
    showCreate.value = false
    createForm.name = ''
    createForm.description = ''
    createForm.visibility = 'private'
    fetchData()
  } finally { creating.value = false }
}

function showKbMenu(e: MouseEvent, kb: KnowledgeBase) {
  e.preventDefault()
  e.stopPropagation()
  kbMenu.visible = true
  kbMenu.x = e.clientX
  kbMenu.y = e.clientY
  kbMenu.kb = kb
}

function handleRowCtxMenu(row: any, _col: any, e: MouseEvent) {
  e.preventDefault()
  showKbMenu(e, row)
}

async function handleKbAction(action: string) {
  const kb = kbMenu.kb
  kbMenu.visible = false
  if (!kb) return
  switch (action) {
    case 'open':
      router.push(`/knowledge/${kb.id}`)
      break
    case 'openNew':
      window.open(`${window.location.origin}/knowledge/${kb.id}`, '_blank')
      break
    case 'share':
      shareLink.value = `${window.location.origin}/knowledge/${kb.id}`
      showShareDialog.value = true
      break
    case 'copyLink': {
      const link = `${window.location.origin}/knowledge/${kb.id}`
      navigator.clipboard.writeText(link)
      ElMessage.success('链接已复制')
      break
    }
    case 'pin':
      ElMessage.success('已添加到置顶')
      break
    case 'favorite':
      ElMessage.success('已收藏')
      break
    case 'settings':
      router.push(`/knowledge/${kb.id}`)
      break
    case 'delete':
      await ElMessageBox.confirm(`确定删除知识库"${kb.name}"？删除后不可恢复。`, '删除确认', { type: 'warning' })
      await deleteKnowledgeBase(kb.id)
      ElMessage.success('已删除')
      fetchData()
      break
  }
}

function copyShareLink() {
  navigator.clipboard.writeText(shareLink.value)
  ElMessage.success('链接已复制')
}

function handleInvite() {
  if (!shareInviteEmail.value.trim()) { ElMessage.warning('请输入邮箱或用户名'); return }
  ElMessage.success('邀请已发送')
  shareInviteEmail.value = ''
}

function closeMenus() {
  kbMenu.visible = false
}

onMounted(() => {
  fetchData()
  document.addEventListener('click', closeMenus)
})
onBeforeUnmount(() => {
  document.removeEventListener('click', closeMenus)
})
</script>

<style scoped>
.knowledge-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px 32px;
}

/* Quick Action Bar */
.kb-quick-bar {
  display: flex;
  gap: 16px;
  margin-bottom: 28px;
}
.kb-quick-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 20px;
  border: 1px solid var(--kx-border);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  background: #fff;
  min-width: 180px;
}
.kb-quick-item:hover {
  border-color: var(--kx-primary);
  box-shadow: 0 2px 12px rgba(51,112,255,0.08);
}
.kb-quick-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
}
.kb-quick-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--kx-text-primary);
}
.kb-quick-desc {
  font-size: 12px;
  color: var(--kx-text-placeholder);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.kb-quick-arrow {
  color: var(--kx-text-placeholder);
  font-size: 12px;
}

/* Search inline */
.kb-search-inline {
  width: 200px;
}

/* Tabs Row */
.kb-tabs-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  border-bottom: 1px solid var(--kx-border);
}
.kb-tabs {
  display: flex;
  gap: 24px;
}
.kb-tab {
  font-size: 14px;
  color: var(--kx-text-secondary);
  cursor: pointer;
  padding-bottom: 10px;
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
}
.kb-tab.active {
  color: var(--kx-primary);
  border-color: var(--kx-primary);
  font-weight: 500;
}
.kb-tab:hover {
  color: var(--kx-primary);
}
.kb-controls {
  display: flex;
  gap: 16px;
  align-items: center;
  padding-bottom: 10px;
}
.ctrl-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: var(--kx-text-secondary);
  cursor: pointer;
}
.ctrl-btn:hover {
  color: var(--kx-primary);
}

/* Knowledge Base Grid */
.kb-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 16px;
}
.kb-card {
  border: 1px solid var(--kx-border);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  overflow: hidden;
  background: #fff;
}
.kb-card:hover {
  border-color: var(--kx-primary);
  box-shadow: 0 4px 16px rgba(51,112,255,0.1);
  transform: translateY(-2px);
}
.kb-card-cover {
  height: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.kb-card-body {
  padding: 14px 16px 16px;
}
.kb-card-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--kx-text-primary);
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.kb-card-desc {
  font-size: 13px;
  color: var(--kx-text-secondary);
  margin-bottom: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.kb-card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.kb-card-stats {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: var(--kx-text-placeholder);
}
.kb-card-stats span {
  display: flex;
  align-items: center;
  gap: 3px;
}
.kb-card-owner {
  display: flex;
  align-items: center;
}

/* New Card */
.kb-card-new {
  border-style: dashed;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 160px;
}
.kb-card-new:hover {
  border-color: var(--kx-primary);
  background: rgba(51,112,255,0.02);
}
.kb-card-new-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: var(--kx-text-placeholder);
  font-size: 14px;
}

/* List View */
.kb-list-view {
  margin-top: 4px;
}
.kb-list-name {
  display: flex;
  align-items: center;
  gap: 10px;
}
.kb-list-icon {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.kb-list-desc {
  color: var(--kx-text-secondary);
  font-size: 13px;
}
.more-btn {
  cursor: pointer;
  font-size: 16px;
  color: var(--kx-text-placeholder);
}
.more-btn:hover {
  color: var(--kx-primary);
}

.kb-empty {
  padding: 60px 0;
}

/* Context Menu */
.context-menu {
  position: fixed;
  z-index: 999;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  padding: 4px 0;
  min-width: 200px;
}
.ctx-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 16px;
  font-size: 13px;
  cursor: pointer;
  color: var(--kx-text-primary);
}
.ctx-item:hover {
  background: var(--kx-sidebar-bg);
}
.ctx-item.danger {
  color: var(--kx-danger);
}
.ctx-sep {
  height: 1px;
  background: var(--kx-border);
  margin: 4px 0;
}

/* Template Grid */
.template-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 8px;
}
.template-card {
  padding: 14px;
  border: 2px solid var(--kx-border);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}
.template-card:hover {
  border-color: var(--kx-primary);
}
.template-card.selected {
  border-color: var(--kx-primary);
  background: rgba(51,112,255,0.04);
}
.tpl-icon {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 8px;
}
.tpl-name {
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 4px;
}
.tpl-desc {
  font-size: 12px;
  color: var(--kx-text-placeholder);
}

/* Radio desc */
.radio-desc {
  font-size: 12px;
  color: var(--kx-text-placeholder);
  margin-top: 2px;
}

/* Share Dialog */
.share-section {
  margin-bottom: 20px;
}
.share-label {
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 8px;
  color: var(--kx-text-primary);
}
.share-link-row {
  display: flex;
  gap: 8px;
}
.share-invite-row {
  display: flex;
  gap: 8px;
}
</style>
