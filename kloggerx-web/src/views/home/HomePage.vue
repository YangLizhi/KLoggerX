<template>
  <div class="home-page">
    <!-- Enhanced Search Bar -->
    <div class="search-section">
      <EnhancedSearchBar
        ref="searchBarRef"
        placeholder="搜索文档、知识库..."
        @select="handleSearchSelect"
      />
    </div>

    <!-- Top Action Bar (fixed, 3 cards) -->
    <div class="action-bar">
      <div class="action-card" @click.stop="showNewMenu = !showNewMenu; showUploadMenu = false">
        <div class="action-icon" style="background: #e8f0fe"><el-icon :size="20" color="#3370ff"><Document /></el-icon></div>
        <div class="action-info">
          <div class="action-label">新建</div>
          <div class="action-desc">新建文档开始协作</div>
        </div>
        <el-icon class="action-arrow"><ArrowDown /></el-icon>
        <div v-if="showNewMenu" class="action-dropdown action-dropdown-large" @click.stop>
          <div class="dropdown-item" @click="handleCreate('doc')"><el-icon color="#3370ff"><Document /></el-icon>文档</div>
          <div class="dropdown-item" @click="handleCreate('sheet')"><el-icon color="#36b37e"><Grid /></el-icon>表格</div>
          <div class="dropdown-item" @click="handleCreate('slide')"><el-icon color="#ff7d00"><Monitor /></el-icon>幻灯片</div>
          <div class="dropdown-item" @click="handleCreate('bitable')"><el-icon color="#00b8d9"><Tickets /></el-icon>多维表格</div>
          <div class="dropdown-item" @click="handleCreate('survey')"><el-icon color="#f54a45"><Notebook /></el-icon>问卷</div>
          <div class="dropdown-item" @click="handleCreate('mindnote')"><el-icon color="#9254de"><Share /></el-icon>思维笔记</div>
          <div class="dropdown-item dropdown-item-arrow" @click.stop="showMoreTypes = !showMoreTypes"><el-icon color="#36b37e"><Grid /></el-icon>更多类型<el-icon class="arrow-right"><ArrowRight /></el-icon>
            <div v-if="showMoreTypes" class="dropdown-submenu" @click.stop>
              <div class="dropdown-item" @click="handleCreate('doc')"><el-icon color="#3370ff"><Document /></el-icon>白板</div>
              <div class="dropdown-item" @click="handleCreate('doc')"><el-icon color="#ff7d00"><Connection /></el-icon>UML图</div>
              <div class="dropdown-item" @click="handleCreate('doc')"><el-icon color="#36b37e"><TrendCharts /></el-icon>甘特图</div>
              <div class="dropdown-item" @click="handleCreate('doc')"><el-icon color="#9254de"><Share /></el-icon>组织架构图</div>
            </div>
          </div>
          <div class="dropdown-sep" />
          <div class="dropdown-item" @click="handleCreate('folder')"><el-icon color="#f5a623"><Folder /></el-icon>文件夹</div>
          <div class="dropdown-group-title">文档应用</div>
          <div class="dropdown-item" @click="handleCreate('doc')"><el-icon color="#36b37e"><EditPen /></el-icon>画板</div>
          <div class="dropdown-item" @click="handleCreate('mindnote')"><el-icon color="#9254de"><Share /></el-icon>思维导图</div>
          <div class="dropdown-item" @click="handleCreate('doc')"><el-icon color="#ff7d00"><Connection /></el-icon>流程图</div>
        </div>
      </div>
      <div class="action-card" @click.stop="showUploadMenu = !showUploadMenu; showNewMenu = false">
        <div class="action-icon" style="background: #e6f7ef"><el-icon :size="20" color="#36b37e"><Upload /></el-icon></div>
        <div class="action-info">
          <div class="action-label">上传</div>
          <div class="action-desc">上传本地文件</div>
        </div>
        <el-icon class="action-arrow"><ArrowDown /></el-icon>
        <div v-if="showUploadMenu" class="action-dropdown" @click.stop>
          <div class="dropdown-item" @click="triggerUpload('file')"><el-icon color="#f5a623"><Document /></el-icon>上传文件</div>
          <div class="dropdown-item" @click="triggerUpload('folder')"><el-icon color="#f5a623"><Folder /></el-icon>上传文件夹</div>
          <div class="dropdown-item" @click="triggerUpload('import')"><el-icon color="#3370ff"><DocumentCopy /></el-icon>导入为在线文档</div>
        </div>
      </div>
      <div class="action-card" @click="$router.push('/templates')">
        <div class="action-icon" style="background: #fef3e0"><el-icon :size="20" color="#f5a623"><Files /></el-icon></div>
        <div class="action-info">
          <div class="action-label">模板库</div>
          <div class="action-desc">选择模板快速新建</div>
        </div>
      </div>
    </div>

    <!-- View Tabs + Filter Bar -->
    <div class="view-bar">
      <div class="view-tabs">
        <span class="view-tab" :class="{ active: activeTab === 'recent' }" @click="switchTab('recent')">最近访问</span>
        <span class="view-tab" :class="{ active: activeTab === 'owned' }" @click="switchTab('owned')">归我所有</span>
        <span class="view-tab" :class="{ active: activeTab === 'shared' }" @click="switchTab('shared')">与我共享</span>
        <span class="view-tab" :class="{ active: activeTab === 'favorites' }" @click="switchTab('favorites')">收藏</span>
        <span
          v-for="cv in customViews"
          :key="cv.id"
          class="view-tab custom-view-tab"
          :class="{ active: activeTab === `custom_${cv.id}` }"
          @click="switchTab(`custom_${cv.id}`)"
        >
          {{ cv.name }}
          <el-icon class="cv-close" @click.stop="removeCustomView(cv.id)"><Close /></el-icon>
        </span>
        <span class="view-tab view-tab-add" @click="addCustomView">
          <span class="custom-view-text">未命名视图</span>
          <el-icon :size="14"><Plus /></el-icon>
        </span>
      </div>
      <div class="view-actions">
        <span class="view-action" @click.stop="showFilterPanel = !showFilterPanel"><el-icon><Filter /></el-icon> 筛选</span>
        <span class="view-action" @click.stop="showDisplaySettings = !showDisplaySettings"><el-icon><Setting /></el-icon> 显示设置</span>
        <el-icon class="view-toggle" :class="{ active: viewMode === 'list' }" @click="viewMode = 'list'"><List /></el-icon>
        <el-icon class="view-toggle" :class="{ active: viewMode === 'grid' }" @click="viewMode = 'grid'"><Grid /></el-icon>
      </div>
    </div>

    <!-- Filter Panel -->
    <div v-if="showFilterPanel" class="filter-panel" @click.stop>
      <div class="filter-row">
        <div class="filter-group">
          <label>文档类型</label>
          <el-select v-model="filterType" placeholder="全部类型" clearable size="small" style="width:140px">
            <el-option label="文档" value="doc" />
            <el-option label="表格" value="sheet" />
            <el-option label="幻灯片" value="slide" />
            <el-option label="多维表格" value="bitable" />
            <el-option label="问卷" value="survey" />
            <el-option label="思维笔记" value="mindnote" />
            <el-option label="文件夹" value="folder" />
          </el-select>
        </div>
        <div class="filter-group">
          <label>所有者</label>
          <el-input v-model="filterOwner" placeholder="输入用户名" clearable size="small" style="width:140px" />
        </div>
        <div class="filter-group">
          <label>创建时间</label>
          <el-select v-model="filterCreatedRange" placeholder="不限" clearable size="small" style="width:120px">
            <el-option label="今日" value="today" />
            <el-option label="昨日" value="yesterday" />
            <el-option label="近7天" value="7days" />
            <el-option label="近30天" value="30days" />
          </el-select>
        </div>
        <div class="filter-group">
          <label>修改时间</label>
          <el-select v-model="filterUpdatedRange" placeholder="不限" clearable size="small" style="width:120px">
            <el-option label="今日" value="today" />
            <el-option label="昨日" value="yesterday" />
            <el-option label="近7天" value="7days" />
            <el-option label="近30天" value="30days" />
          </el-select>
        </div>
        <el-button size="small" @click="applyFilter">应用</el-button>
        <el-button size="small" text @click="resetFilter">重置</el-button>
      </div>
    </div>

    <!-- Display Settings Panel -->
    <div v-if="showDisplaySettings" class="display-settings-panel" @click.stop>
      <div class="display-title">显示列设置</div>
      <el-checkbox v-model="colVisible.title" disabled>标题</el-checkbox>
      <el-checkbox v-model="colVisible.fileType">文件类型</el-checkbox>
      <el-checkbox v-model="colVisible.fileSize">文件大小</el-checkbox>
      <el-checkbox v-model="colVisible.location">位置</el-checkbox>
      <el-checkbox v-model="colVisible.owner">所有者</el-checkbox>
      <el-checkbox v-model="colVisible.createdAt">创建时间</el-checkbox>
      <el-checkbox v-model="colVisible.updatedAt">修改时间</el-checkbox>
    </div>

    <!-- Document List -->
    <div v-loading="loading" class="doc-list-area">
      <!-- List View -->
      <el-table
        v-if="viewMode === 'list'"
        :data="filteredDocuments"
        style="width: 100%"
        table-layout="auto"
        @row-dblclick="handleDocClick"
        @row-contextmenu="handleRowContextMenu"
        @selection-change="handleSelectionChange"
        :default-sort="{ prop: sortField, order: sortOrder }"
        @sort-change="handleSortChange"
      >
        <el-table-column type="selection" width="50" />
        <el-table-column label="标题" prop="title" min-width="240" sortable="custom">
          <template #default="{ row }">
            <div class="doc-name-cell">
              <el-icon :color="getTypeColor(row.type, row.fileExt)" :size="16"><component :is="getTypeIcon(row.type, row.fileExt)" /></el-icon>
              <span class="doc-title-text">{{ getDisplayName(row) }}</span>
              <el-icon v-if="row.isPinned" class="pin-badge" color="#3370ff" :size="12"><Flag /></el-icon>
            </div>
          </template>
        </el-table-column>
        <el-table-column v-if="colVisible.fileType" label="文件类型" min-width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="getTypeTagType(row.type)" disable-transitions>{{ getTypeName(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column v-if="colVisible.fileSize" label="文件大小" min-width="100">
          <template #default="{ row }">
            <span class="size-cell">{{ row.fileSize > 0 ? formatFileSize(row.fileSize) : (row.type === 'folder' ? '—' : '—') }}</span>
          </template>
        </el-table-column>
        <el-table-column v-if="colVisible.location" label="位置" min-width="140">
          <template #default="{ row }">
            <div class="location-cell">
              <el-icon :size="14" color="#f5a623"><FolderOpened /></el-icon>
              <span>{{ row.parentId ? '我的文档库' : '云盘' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column v-if="colVisible.owner" label="所有者" min-width="100">
          <template #default="{ row }">
            <div class="owner-cell">
              <el-avatar :size="20" :style="{ background: getAvatarColor(row.ownerId) }">{{ (row.ownerName || 'U')[0] }}</el-avatar>
              <span>{{ row.ownerName || '未知' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column v-if="colVisible.createdAt" label="创建时间" prop="createdAt" min-width="130" sortable="custom">
          <template #header>
            <span>创建时间 {{ sortField === 'createdAt' ? (sortOrder === 'descending' ? '↓' : '↑') : '' }}</span>
          </template>
          <template #default="{ row }">{{ formatDate(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column v-if="colVisible.updatedAt" label="修改时间" prop="updatedAt" min-width="130" sortable="custom">
          <template #header>
            <span>修改时间 {{ sortField === 'updatedAt' ? (sortOrder === 'descending' ? '↓' : '↑') : '' }}</span>
          </template>
          <template #default="{ row }">{{ formatDate(row.updatedAt) }}</template>
        </el-table-column>
      </el-table>

      <!-- Grid View -->
      <div v-else class="doc-grid">
        <div
          v-for="doc in filteredDocuments"
          :key="doc.id"
          class="doc-card"
          @dblclick="handleDocClick(doc)"
          @contextmenu.prevent="showContextMenu($event, doc)"
        >
          <div class="doc-card-check" @click.stop>
            <el-checkbox :model-value="(doc as any)._selected" @update:model-value="(val: any) => (doc as any)._selected = val" />
          </div>
          <div class="doc-card-icon">
            <el-icon :size="36" :color="getTypeColor(doc.type, doc.fileExt)"><component :is="getTypeIcon(doc.type, doc.fileExt)" /></el-icon>
          </div>
          <div class="doc-card-title">{{ doc.title }}</div>
          <div class="doc-card-meta">
            <el-avatar :size="16" :style="{ background: getAvatarColor(doc.ownerId) }">{{ (doc.ownerName || 'U')[0] }}</el-avatar>
            <span>{{ doc.ownerName }}</span>
            <span class="dot">·</span>
            <span>{{ formatDate(doc.updatedAt) }}</span>
          </div>
        </div>
      </div>

      <div v-if="!loading && !filteredDocuments.length" class="empty-state">
        <el-empty description="暂无文档" />
      </div>
    </div>

    <div class="end-marker" v-if="filteredDocuments.length">
      <span>已经到底了</span>
    </div>

    <!-- Batch Action Bar -->
    <div v-if="selectedDocs.length" class="batch-bar">
      <span class="batch-info">已选 {{ selectedDocs.length }} 项</span>
      <el-button size="small" @click="batchAction('move')">移动到</el-button>
      <el-button size="small" @click="batchAction('share')">分享</el-button>
      <el-button size="small" type="danger" @click="batchAction('delete')">删除</el-button>
      <el-button size="small" text @click="clearSelection">取消选择</el-button>
    </div>

    <!-- Context Menu (right-click / ... button) -->
    <div v-if="contextMenu.visible" class="context-menu" :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }">
      <div class="ctx-item" @click="handleAction('share')"><el-icon><Share /></el-icon>分享</div>
      <div class="ctx-item" @click="handleAction('copyLink')"><el-icon><Link /></el-icon>复制链接</div>
      <div v-if="contextMenu.doc?.type === 'file' || contextMenu.doc?.fileSize" class="ctx-item" @click="handleAction('download')"><el-icon><Download /></el-icon>下载原文件</div>
      <div class="ctx-sep" />
      <div class="ctx-item" @click="handleAction('copy')"><el-icon><DocumentCopy /></el-icon>创建副本</div>
      <div class="ctx-item" @click="handleAction('shortcut')"><el-icon><Position /></el-icon>添加快捷方式到</div>
      <div class="ctx-item" @click="handleAction('pin')">
        <el-icon><Flag /></el-icon>{{ contextMenu.doc?.isPinned ? '从"置顶"移除' : '添加到"置顶"' }}
      </div>
      <div class="ctx-item" @click="handleAction('favorite')">
        <el-icon><Star /></el-icon>{{ contextMenu.doc?.isFavorite ? '取消收藏' : '收藏' }}
      </div>
      <div class="ctx-sep" />
      <div class="ctx-item ctx-item-toggle" @click.stop="handleAction('offline')">
        <el-icon><Download /></el-icon>设为离线可使用
        <el-switch v-model="offlineToggle" size="small" class="ctx-switch" @click.stop />
      </div>
      <div class="ctx-item ctx-item-toggle" @click.stop="handleAction('follow')">
        <el-icon><BellFilled /></el-icon>关注文档更新
        <el-switch v-model="followToggle" size="small" class="ctx-switch" @click.stop />
      </div>
      <div class="ctx-sep" />
      <div class="ctx-item" @click="handleAction('move')"><el-icon><Rank /></el-icon>移动到</div>
      <div class="ctx-item" @click="handleAction('rename')"><el-icon><EditPen /></el-icon>重命名</div>
      <div class="ctx-item" @click="handleAction('transfer')"><el-icon><Switch /></el-icon>转移所有权</div>
      <div class="ctx-item danger" @click="handleAction('delete')"><el-icon><Delete /></el-icon>删除</div>
    </div>

    <!-- Import Dialog -->
    <el-dialog v-model="showImportDialog" :title="uploadDialogTitle" width="480px" destroy-on-close @close="importFile = null; importFiles = []">
      <el-upload
        v-if="uploadType === 'folder'"
        drag
        :auto-upload="false"
        :limit="100"
        multiple
        directory
        webkitdirectory
        :on-change="handleFolderSelect"
        :on-exceed="() => ElMessage.warning('文件数量过多')"
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">拖拽文件夹到此处，或 <em>点击选择文件夹</em></div>
        <template #tip><div class="el-upload__tip">选择文件夹批量上传文件</div></template>
      </el-upload>
      <el-upload
        v-else-if="uploadType === 'file'"
        drag
        :auto-upload="false"
        :limit="10"
        multiple
        :on-change="handleFileSelect"
        :on-exceed="() => ElMessage.warning('最多上传10个文件')"
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">拖拽文件到此处，或 <em>点击上传</em></div>
        <template #tip><div class="el-upload__tip">支持所有常见文件格式</div></template>
      </el-upload>
      <el-upload
        v-else
        drag
        :auto-upload="false"
        :limit="1"
        accept=".md,.json,.txt,.html,.docx,.doc,.xlsx,.xls,.pptx,.ppt,.pdf,.png,.jpg,.jpeg,.gif,.webp"
        :on-change="(f: any) => importFile = f.raw"
        :on-exceed="() => ElMessage.warning('只能上传一个文件')"
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">拖拽文件到此处，或 <em>点击上传</em></div>
        <template #tip><div class="el-upload__tip">支持 Word、Excel、PPT、PDF、Markdown、JSON、TXT、HTML、图片</div></template>
      </el-upload>
      <div v-if="importFiles.length > 0" class="import-file-list">
        <div v-for="(f, i) in importFiles" :key="i" class="import-file-item">
          <el-icon><Document /></el-icon>
          <span>{{ f.name }}</span>
          <span class="file-size">{{ formatFileSize(f.size) }}</span>
        </div>
      </div>
      <el-progress v-if="importLoading && importProgress > 0" :percentage="importProgress" style="margin-top: 12px" />
      <template #footer>
        <el-button @click="showImportDialog = false; importFile = null; importFiles = []">取消</el-button>
        <el-button type="primary" :loading="importLoading" :disabled="!importFile && !importFiles.length" @click="handleImport">导入</el-button>
      </template>
    </el-dialog>

    <!-- Move Dialog -->
    <el-dialog v-model="showMoveDialog" title="移动到" width="420px" destroy-on-close>
      <div class="move-folder-list">
        <div class="move-folder-item" :class="{ active: moveTarget === null }" @click="moveTarget = null">
          <el-icon color="#f5a623"><FolderOpened /></el-icon><span>根目录</span>
        </div>
        <div v-for="f in allFolders" :key="f.id" class="move-folder-item" :class="{ active: moveTarget === f.id }" @click="moveTarget = f.id">
          <el-icon color="#f5a623"><Folder /></el-icon><span>{{ f.title }}</span>
        </div>
      </div>
      <template #footer>
        <el-button @click="showMoveDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmMove">确认移动</el-button>
      </template>
    </el-dialog>

    <!-- Rename Dialog -->
    <el-dialog v-model="showRenameDialog" title="重命名" width="400px" destroy-on-close>
      <el-input v-model="renameValue" placeholder="输入新名称" maxlength="100" show-word-limit @keyup.enter="confirmRename" />
      <template #footer>
        <el-button @click="showRenameDialog = false">取消</el-button>
        <el-button type="primary" :loading="renaming" @click="confirmRename">确认</el-button>
      </template>
    </el-dialog>

    <!-- Transfer Dialog -->
    <el-dialog v-model="showTransferDialog" title="转移所有权" width="420px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="选择目标用户">
          <el-input v-model="transferUser" placeholder="输入用户名或邮箱搜索" prefix-icon="Search" />
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="transferKeepPerm">保留我的协作权限</el-checkbox>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showTransferDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmTransfer">确认转移</el-button>
      </template>
    </el-dialog>

    <!-- Share Dialog -->
    <el-dialog v-model="showShareDialog" title="分享" width="520px" destroy-on-close>
      <div class="share-section">
        <div class="share-subtitle">邀请协作者</div>
        <div class="share-invite-row">
          <el-input v-model="shareInvite" placeholder="输入用户名或邮箱" style="flex:1" />
          <el-select v-model="sharePermLevel" style="width:120px">
            <el-option label="可管理" value="manage" />
            <el-option label="可编辑" value="edit" />
            <el-option label="可查看" value="view" />
            <el-option label="只读" value="readonly" />
          </el-select>
          <el-button type="primary" @click="ElMessage.info('已发送邀请')">邀请</el-button>
        </div>
      </div>
      <div class="share-section" style="margin-top:16px">
        <div class="share-subtitle">分享链接</div>
        <div class="share-link-row">
          <el-input :model-value="shareLink" readonly style="flex:1" />
          <el-button @click="copyShareLink">复制</el-button>
        </div>
        <div class="share-options">
          <el-select v-model="shareLinkScope" size="small" style="width:160px;margin-top:8px">
            <el-option label="仅协作者可见" value="collaborator" />
            <el-option label="组织内可见" value="org" />
            <el-option label="互联网公开可见" value="public" />
          </el-select>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  getRecentDocuments, getDocumentTree, getFavorites, searchDocuments,
  createDocument, pinDocument, favoriteDocument, deleteDocument,
  copyDocument, importDocument, moveDocument, updateDocument
} from '@/api/modules/document'
import type { Document, DocumentType, KnowledgeBase } from '@/types'
import { ElMessage, ElMessageBox } from 'element-plus'
import EnhancedSearchBar from '@/components/common/EnhancedSearchBar.vue'

const router = useRouter()
const route = useRoute()
const loading = ref(false)
const viewMode = ref<'list' | 'grid'>('list')
const documents = ref<Document[]>([])
const allFolders = ref<Document[]>([])
const activeTab = ref('recent')
// const searchBarRef = ref<InstanceType<typeof EnhancedSearchBar>>()
const showNewMenu = ref(false)
const showUploadMenu = ref(false)
const showMoreTypes = ref(false)
const showImportDialog = ref(false)
const importFile = ref<File | null>(null)
const importFiles = ref<File[]>([])
const importLoading = ref(false)
const uploadType = ref<'file' | 'folder' | 'import'>('file')
const importProgress = ref(0)
const page = ref(1)
const pageSize = 50

// Context menu
const contextMenu = reactive({ visible: false, x: 0, y: 0, doc: null as Document | null })
const offlineToggle = ref(false)
const followToggle = ref(false)

// Filter
const showFilterPanel = ref(false)
const filterType = ref('')
const filterOwner = ref('')
const filterCreatedRange = ref('')
const filterUpdatedRange = ref('')

// Display settings
const showDisplaySettings = ref(false)
const colVisible = reactive({
  title: true,
  fileType: true,
  fileSize: true,
  location: true,
  owner: true,
  createdAt: true,
  updatedAt: true,
})

// Sort
const sortField = ref('updatedAt')
const sortOrder = ref<'ascending' | 'descending'>('descending')

// Custom views
const customViews = ref<{ id: number; name: string }[]>([])
let cvIdCounter = 1

// Selection
const selectedDocs = ref<Document[]>([])

// Dialogs
const showMoveDialog = ref(false)
const moveTarget = ref<number | null>(null)
const showRenameDialog = ref(false)
const renameValue = ref('')
const renaming = ref(false)
const showTransferDialog = ref(false)
const transferUser = ref('')
const transferKeepPerm = ref(true)
const showShareDialog = ref(false)
const shareInvite = ref('')
const sharePermLevel = ref('edit')
const shareLinkScope = ref('collaborator')

const shareLink = computed(() => {
  if (contextMenu.doc) return `${window.location.origin}/doc/${contextMenu.doc.id}`
  return ''
})

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

// Extension-based icon mapping
const extIconMap: Record<string, { icon: string; color: string }> = {
  // 文档类
  pdf: { icon: 'Document', color: '#f54a45' },
  doc: { icon: 'Document', color: '#2b579a' },
  docx: { icon: 'Document', color: '#2b579a' },
  txt: { icon: 'Document', color: '#666' },
  md: { icon: 'Memo', color: '#3370ff' },
  // 表格
  xls: { icon: 'Grid', color: '#217346' },
  xlsx: { icon: 'Grid', color: '#217346' },
  csv: { icon: 'Grid', color: '#36b37e' },
  // 演示文稿
  ppt: { icon: 'Monitor', color: '#d24726' },
  pptx: { icon: 'Monitor', color: '#d24726' },
  // 图片
  png: { icon: 'Picture', color: '#36b37e' },
  jpg: { icon: 'Picture', color: '#36b37e' },
  jpeg: { icon: 'Picture', color: '#36b37e' },
  gif: { icon: 'Picture', color: '#ff7d00' },
  svg: { icon: 'Picture', color: '#ff7d00' },
  // 音频
  mp3: { icon: 'Headset', color: '#9254de' },
  wav: { icon: 'Headset', color: '#9254de' },
  flac: { icon: 'Headset', color: '#9254de' },
  // 视频
  mp4: { icon: 'VideoPlay', color: '#ff7d00' },
  mkv: { icon: 'VideoPlay', color: '#ff7d00' },
  avi: { icon: 'VideoPlay', color: '#ff7d00' },
  mov: { icon: 'VideoPlay', color: '#ff7d00' },
  // 压缩包
  zip: { icon: 'Files', color: '#f5a623' },
  rar: { icon: 'Files', color: '#f5a623' },
  '7z': { icon: 'Files', color: '#f5a623' },
  // 可执行文件
  exe: { icon: 'Monitor', color: '#3370ff' },
  apk: { icon: 'Iphone', color: '#36b37e' },
  sh: { icon: 'Memo', color: '#36b37e' },
  // 代码
  html: { icon: 'Link', color: '#ff7d00' },
  css: { icon: 'Memo', color: '#264de4' },
  js: { icon: 'Memo', color: '#f7df1e' },
  ts: { icon: 'Memo', color: '#3178c6' },
  vue: { icon: 'Memo', color: '#42b883' },
  json: { icon: 'Memo', color: '#f5a623' },
  py: { icon: 'Memo', color: '#3776ab' },
  go: { icon: 'Memo', color: '#00add8' },
  java: { icon: 'Memo', color: '#f54a45' },
}

const avatarColors = ['#3370ff', '#36b37e', '#ff7d00', '#f54a45', '#9254de', '#00b8d9']

const typeNameMap: Record<string, string> = {
  folder: '文件夹', doc: '文档', sheet: '表格', slide: '幻灯片',
  mindnote: '思维笔记', bitable: '多维表格', survey: '问卷',
  file: '其他', image: '图片', code: '代码',
}
const extTypeMap: Record<string, string> = {
 // 文档类
  txt: '文本文档',
  doc: 'Word文档',
  docx: 'Word文档',
  pdf: 'PDF文档',
  rtf: '富文本',
  epub: '电子书',
  mobi: '电子书',

  // 表格/数据类
  xls: 'Excel表格',
  xlsx: 'Excel表格',
  csv: 'Csv表格',
  db: '数据库',
  sqlite: '数据库',
  sql: '数据库',

  // 演示文稿
  ppt: '幻灯片',
  pptx: '幻灯片',
  pot: '幻灯片模板',

  // 图片类
  png: 'PNG图片',
  jpg: 'JPEG图片',
  jpeg: 'JPEG图片',
  gif: 'GIF动图',
  bmp: 'BMP图片',
  webp: 'WebP图片',
  svg: '矢量图片',
  ico: '图标文件',

  // 音频
  mp3: '音频文件',
  wav: '无损音频',
  flac: 'FLAC无损音频',
  aac: 'AAC音频',
  ogg: 'OGG音频',
  m4a: 'M4A音频',

  // 视频
  mp4: '视频文件',
  mkv: 'MKV视频',
  avi: 'AVI视频',
  mov: 'MOV视频',
  wmv: 'WMV视频',
  flv: 'FLV视频',
  webm: 'WEBM视频',

  // 压缩包
  zip: '压缩包',
  rar: '压缩包',
  '7z': '7Z压缩包',
  tar: 'TAR打压缩包',
  gz: 'GZ压缩包',
  'tar.gz': 'TAR.GZ压缩包',
  bz2: 'BZ2压缩包',

  // 系统/安装包
  exe: 'Windows程序',
  msi: 'Windows安装包',
  dll: '系统库文件',
  apk: '安卓安装包',
  aab: '安卓应用捆绑包',
  ipa: 'iOS安装包',
  app: 'macOS应用',
  dmg: '苹果磁盘镜像',
  deb: 'Ubuntu安装包',
  rpm: 'RedHat安装包',
  AppImage: 'Linux便携程序',
  sh: 'Shell脚本',
  iso: '光盘镜像',

  // 网页/代码
  html: '网页文件',
  htm: '网页文件',
  css: '样式文件',
  js: 'JavaScript',
  ts: 'TypeScript',
  vue: 'Vue组件',
  jsx: 'React组件',
  md: 'Markdown文档',
  json: 'JSON配置',
  xml: 'XML文件',
  yaml: 'YAML文件',
  yml: 'YAML文件',

  // 编程源码
  c: 'C语言代码',
  cpp: 'C++代码',
  h: '头文件',
  java: 'Java代码',
  class: 'Java编译文件',
  jar: 'Java包',
  py: 'Python代码',
  go: 'Go代码',
  php: 'PHP脚本',

  // 其他
  log: '日志文件',
  tmp: '临时文件',
  vmdk: '虚拟机磁盘',
  vdi: '虚拟机磁盘',
}

function getTypeIcon(t: string, ext?: string): string {
  if (ext) {
    const e = ext.replace('.', '').toLowerCase()
    if (extIconMap[e]) return extIconMap[e].icon
  }
  return typeMap[t]?.icon || 'Document'
}
function getTypeColor(t: string, ext?: string): string {
  if (ext) {
    const e = ext.replace('.', '').toLowerCase()
    if (extIconMap[e]) return extIconMap[e].color
  }
  return typeMap[t]?.color || '#888'
}
function getAvatarColor(id: number) { return avatarColors[(id || 0) % avatarColors.length] }

// Returns the display name: for uploaded files prefer originalName, otherwise title
function getDisplayName(doc: Document): string {
  if (doc.originalName) return doc.originalName
  if (doc.fileExt && !doc.title.endsWith(doc.fileExt)) return doc.title + doc.fileExt
  return doc.title
}

// Returns human-friendly type name, using file extension if available
function getTypeName(doc: Document): string {
  const ext = (doc.fileExt || '').replace('.', '').toLowerCase()
  if (ext && extTypeMap[ext]) return extTypeMap[ext]
  if (typeNameMap[doc.type]) return typeNameMap[doc.type]
  return '其他'
}

function getTypeTagType(type: string): '' | 'success' | 'warning' | 'info' | 'danger' {
  const m: Record<string, '' | 'success' | 'warning' | 'info' | 'danger'> = {
    folder: 'warning', doc: '', sheet: 'success', slide: 'warning',
    mindnote: '', bitable: 'info', survey: 'danger', file: 'info',
  }
  return m[type] || 'info'
}

function formatDate(t: string) {
  if (!t) return ''
  const d = new Date(t)
  const now = new Date()
  const isToday = d.toDateString() === now.toDateString()
  if (isToday) return `今天 ${d.getHours().toString().padStart(2,'0')}:${d.getMinutes().toString().padStart(2,'0')}`
  return `${d.getFullYear()}年${d.getMonth()+1}月${d.getDate()}日`
}

// Filter logic
function isInRange(dateStr: string, range: string): boolean {
  if (!range) return true
  const d = new Date(dateStr)
  const now = new Date()
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  if (range === 'today') return d >= startOfToday
  if (range === 'yesterday') { const y = new Date(startOfToday); y.setDate(y.getDate() - 1); return d >= y && d < startOfToday }
  if (range === '7days') { const s = new Date(startOfToday); s.setDate(s.getDate() - 7); return d >= s }
  if (range === '30days') { const s = new Date(startOfToday); s.setDate(s.getDate() - 30); return d >= s }
  return true
}

const filteredDocuments = computed(() => {
  let docs = documents.value
  if (filterType.value) docs = docs.filter(d => d.type === filterType.value)
  if (filterOwner.value) docs = docs.filter(d => (d.ownerName || '').includes(filterOwner.value))
  if (filterCreatedRange.value) docs = docs.filter(d => isInRange(d.createdAt, filterCreatedRange.value))
  if (filterUpdatedRange.value) docs = docs.filter(d => isInRange(d.updatedAt, filterUpdatedRange.value))
  // Sort
  const field = sortField.value as keyof Document
  const dir = sortOrder.value === 'ascending' ? 1 : -1
  docs = [...docs].sort((a, b) => {
    const va = (a[field] || '') as string
    const vb = (b[field] || '') as string
    if (field === 'title') return dir * va.localeCompare(vb)
    return dir * (new Date(va).getTime() - new Date(vb).getTime())
  })
  return docs
})

function applyFilter() {
  showFilterPanel.value = false
}
function resetFilter() {
  filterType.value = ''
  filterOwner.value = ''
  filterCreatedRange.value = ''
  filterUpdatedRange.value = ''
}

function handleSortChange({ prop, order }: any) {
  if (prop) sortField.value = prop
  sortOrder.value = order || 'descending'
}

function switchTab(tab: string) {
  activeTab.value = tab
  page.value = 1
  resetFilter()
  fetchDocuments()
}

async function fetchDocuments() {
  loading.value = true
  try {
    let res: any
    if (activeTab.value === 'recent') {
      res = await getRecentDocuments({ page: page.value, pageSize })
      documents.value = res.data?.list || []
    } else if (activeTab.value === 'owned') {
      res = await getDocumentTree(null)
      documents.value = (res.data || []).filter((d: Document) => d.type !== 'folder')
    } else if (activeTab.value === 'shared') {
      res = await searchDocuments({ keyword: '', page: 1, pageSize: 50 })
      documents.value = res.data?.list || []
    } else if (activeTab.value === 'favorites') {
      res = await getFavorites({ page: page.value, pageSize })
      documents.value = res.data?.list || []
    } else {
      // custom view - use recent as default
      res = await getRecentDocuments({ page: page.value, pageSize })
      documents.value = res.data?.list || []
    }
    // Fetch all folders for move dialog
    const fRes: any = await getDocumentTree(null)
    allFolders.value = (fRes.data || []).filter((d: Document) => d.type === 'folder')
  } finally {
    loading.value = false
  }
}

function handleDocClick(doc: Document) {
  if (doc.type === 'folder') {
    router.push('/documents')
  } else {
    // Open in new browser tab
    window.open(`/doc/${doc.id}`, '_blank')
  }
}

async function handleCreate(type: DocumentType | 'folder') {
  showNewMenu.value = false
  try {
    const res: any = await createDocument({ title: type === 'folder' ? '新建文件夹' : `未命名${typeLabel(type)}`, type, parentId: null })
    if (type !== 'folder') {
      // Open in new browser tab
      window.open(`/doc/${res.data.id}`, '_blank')
    } else {
      ElMessage.success('文件夹已创建')
      fetchDocuments()
    }
  } catch { /* handled */ }
}

function typeLabel(type: string): string {
  const m: Record<string, string> = { doc: '文档', sheet: '表格', slide: '幻灯片', bitable: '多维表格', survey: '问卷', mindnote: '思维笔记' }
  return m[type] || '文档'
}

function showContextMenu(e: MouseEvent, doc: Document) {
  e.preventDefault()
  e.stopPropagation()
  
  // 计算菜单位置，适应窗口边距
  const menuWidth = 220
  const menuHeight = 400
  const padding = 8
  
  let x = e.clientX
  let y = e.clientY
  
  // 右边界检测
  if (x + menuWidth + padding > window.innerWidth) {
    x = window.innerWidth - menuWidth - padding
  }
  // 左边界检测
  if (x < padding) {
    x = padding
  }
  // 下边界检测
  if (y + menuHeight + padding > window.innerHeight) {
    y = window.innerHeight - menuHeight - padding
  }
  // 上边界检测
  if (y < padding) {
    y = padding
  }
  
  contextMenu.visible = true
  contextMenu.x = x
  contextMenu.y = y
  contextMenu.doc = doc
  offlineToggle.value = false
  followToggle.value = false
}

function handleRowContextMenu(row: Document, _col: any, e: MouseEvent) {
  showContextMenu(e, row)
}

function closeMenus() {
  contextMenu.visible = false
  showNewMenu.value = false
  showUploadMenu.value = false
  showFilterPanel.value = false
  showDisplaySettings.value = false
}

function handleSelectionChange(selection: Document[]) {
  selectedDocs.value = selection
}
function clearSelection() {
  selectedDocs.value = []
}

async function batchAction(action: string) {
  if (action === 'delete') {
    await ElMessageBox.confirm(`确定将 ${selectedDocs.value.length} 项移至回收站？删除后30天内可恢复。`, '删除确认')
    for (const d of selectedDocs.value) { await deleteDocument(d.id) }
    ElMessage.success('已移至回收站')
    fetchDocuments()
  } else if (action === 'move') {
    showMoveDialog.value = true
  } else if (action === 'share') {
    // Batch share - open share dialog
    // shareLink.value = `${window.location.origin}/documents?ids=${selectedDocs.value.map(d => d.id).join(',')}`
    // shareLinkScope.value = 'collaborator'
    // shareDialogVisible.value = true
    ElMessage.info('批量分享功能开发中')
  }
  clearSelection()
}

// Custom views
function addCustomView() {
  customViews.value.push({ id: cvIdCounter++, name: '未命名视图' })
}
function removeCustomView(id: number) {
  customViews.value = customViews.value.filter(v => v.id !== id)
  if (activeTab.value === `custom_${id}`) switchTab('recent')
}

function triggerUpload(type: string) {
  showUploadMenu.value = false
  uploadType.value = type as any
  importFile.value = null
  importFiles.value = []
  importProgress.value = 0
  showImportDialog.value = true
}

const uploadDialogTitle = computed(() => {
  switch (uploadType.value) {
    case 'file': return '上传文件'
    case 'folder': return '上传文件夹'
    default: return '导入为在线文档'
  }
})

function handleFileSelect(file: any) {
  if (file.raw) {
    importFiles.value.push(file.raw)
  }
}

function handleFolderSelect(file: any) {
  if (file.raw) {
    importFiles.value.push(file.raw)
  }
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
  return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
}

async function handleImport() {
  // Single file import
  if (importFile.value) {
    importLoading.value = true
    importProgress.value = 0
    try {
      const res: any = await importDocument(importFile.value, null, (p) => { importProgress.value = p })
      ElMessage.success('导入成功')
      showImportDialog.value = false
      importFile.value = null
      importProgress.value = 0
      fetchDocuments()
      if (res.data?.id && res.data?.type !== 'folder') {
        window.open(`/doc/${res.data.id}`, '_blank')
      }
    } catch { ElMessage.error('导入失败') } finally { importLoading.value = false }
    return
  }

  // Multiple files upload
  if (importFiles.value.length > 0) {
    importLoading.value = true
    importProgress.value = 0
    let successCount = 0
    let failCount = 0
    const total = importFiles.value.length

    for (let i = 0; i < importFiles.value.length; i++) {
      const file = importFiles.value[i]
      try {
        await importDocument(file, null, (p) => {
          importProgress.value = Math.round(((i + p / 100) / total) * 100)
        })
        successCount++
      } catch {
        failCount++
      }
    }

    importLoading.value = false
    importProgress.value = 0

    if (successCount > 0) {
      ElMessage.success(`成功上传 ${successCount} 个文件${failCount > 0 ? `，${failCount} 个失败` : ''}`)
      showImportDialog.value = false
      importFiles.value = []
      fetchDocuments()
    } else {
      ElMessage.error('上传失败')
    }
  }
}

async function handleAction(action: string) {
  const doc = contextMenu.doc
  if (!doc) return
  contextMenu.visible = false

  switch (action) {
    case 'share':
      showShareDialog.value = true
      break
    case 'copyLink': {
      const link = `${window.location.origin}/doc/${doc.id}`
      navigator.clipboard.writeText(link)
      ElMessage.success('链接已复制')
      break
    }
    case 'download': {
      const a = document.createElement('a')
      a.href = `/api/v1/document/${doc.id}/download`
      a.click()
      break
    }
    case 'copy':
      await copyDocument(doc.id, false)
      ElMessage.success('副本已创建，标题为"' + doc.title + ' 副本"')
      fetchDocuments()
      break
    case 'shortcut':
      ElMessage.success(`"${doc.title}" 的快捷方式已添加到桌面`)
      break
    case 'pin':
      await pinDocument(doc.id, !doc.isPinned)
      ElMessage.success(doc.isPinned ? '已从置顶移除' : '已添加到置顶')
      fetchDocuments()
      break
    case 'favorite':
      await favoriteDocument(doc.id, !doc.isFavorite)
      ElMessage.success(doc.isFavorite ? '已取消收藏' : '已收藏')
      fetchDocuments()
      break
    case 'offline':
      offlineToggle.value = !offlineToggle.value
      ElMessage.success(offlineToggle.value ? '已设为离线可使用' : '已取消离线使用')
      break
    case 'follow':
      followToggle.value = !followToggle.value
      ElMessage.success(followToggle.value ? '已关注文档更新' : '已取消关注')
      break
    case 'move':
      moveTarget.value = null
      showMoveDialog.value = true
      break
    case 'rename':
      renameValue.value = doc.title
      showRenameDialog.value = true
      break
    case 'transfer':
      transferUser.value = ''
      transferKeepPerm.value = true
      showTransferDialog.value = true
      break
    case 'delete':
      await ElMessageBox.confirm(`确定要删除"${doc.title}"吗？删除后将移入回收站，30天内可恢复。`, '删除确认')
      await deleteDocument(doc.id)
      ElMessage.success('已移至回收站')
      fetchDocuments()
      break
  }
}

async function confirmMove() {
  const doc = contextMenu.doc
  if (doc) {
    await moveDocument(doc.id, moveTarget.value)
    ElMessage.success('已移动')
  }
  // batch move
  for (const d of selectedDocs.value) {
    await moveDocument(d.id, moveTarget.value)
  }
  showMoveDialog.value = false
  fetchDocuments()
}

async function confirmRename() {
  const doc = contextMenu.doc
  if (!doc || !renameValue.value.trim()) { ElMessage.warning('名称不能为空'); return }
  renaming.value = true
  try {
    await updateDocument(doc.id, { title: renameValue.value.trim() })
    ElMessage.success('重命名成功')
    showRenameDialog.value = false
    fetchDocuments()
  } catch { ElMessage.error('重命名失败') }
  finally { renaming.value = false }
}

function confirmTransfer() {
  ElMessage.info('转移所有权功能需要后端支持用户搜索')
  showTransferDialog.value = false
}

function copyShareLink() {
  if (shareLink.value) {
    navigator.clipboard.writeText(shareLink.value)
    ElMessage.success('链接已复制')
  }
}

// Handle search result selection from EnhancedSearchBar
function handleSearchSelect(_result: { type: 'document' | 'knowledge'; data: Document | KnowledgeBase }) {
  // Navigation is already handled in the component
  // Just close any open menus and ensure clean state
  closeMenus()
}

watch(() => route.query.keyword, (val) => {
  if (val) {
    loading.value = true
    searchDocuments({ keyword: val as string, page: 1, pageSize: 50 }).then((res: any) => {
      documents.value = res.data?.list || []
    }).finally(() => { loading.value = false })
  }
})

onMounted(() => {
  if (route.query.keyword) {
    loading.value = true
    searchDocuments({ keyword: route.query.keyword as string, page: 1, pageSize: 50 }).then((res: any) => {
      documents.value = res.data?.list || []
    }).finally(() => { loading.value = false })
  } else {
    fetchDocuments()
  }
  document.addEventListener('click', closeMenus)
})
onBeforeUnmount(() => {
  document.removeEventListener('click', closeMenus)
})
</script>

<style scoped>
.home-page {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 20px 24px;
  box-sizing: border-box;
}

/* Action Bar */
/* Search Section */
.search-section {
  margin-bottom: 24px;
  padding: 0;
}

.action-bar {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 24px;
}
.action-card {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 18px;
  border: 1px solid var(--kx-border);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  background: #fff;
}
.action-card:hover {
  border-color: var(--kx-primary);
  box-shadow: 0 2px 8px rgba(51,112,255,0.08);
}
.action-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.action-info { flex: 1; min-width: 0; }
.action-label { font-size: 14px; font-weight: 500; color: var(--kx-text-primary); }
.action-desc { font-size: 12px; color: var(--kx-text-placeholder); margin-top: 2px; }
.action-arrow { color: var(--kx-text-placeholder); font-size: 12px; }
.action-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  min-width: 200px;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  z-index: 100;
  padding: 4px 0;
}
.dropdown-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 16px;
  font-size: 14px;
  cursor: pointer;
  color: var(--kx-text-primary);
}
.dropdown-item:hover { background: var(--kx-sidebar-bg); }
.dropdown-sep { height: 1px; background: var(--kx-border); margin: 4px 0; }
.dropdown-group-title { font-size: 12px; color: var(--kx-text-placeholder); padding: 6px 16px 2px; }
.dropdown-item-arrow { position: relative; }
.arrow-right { margin-left: auto; font-size: 12px; color: var(--kx-text-placeholder); }
.action-dropdown-large { min-width: 220px; max-height: 480px; overflow-y: auto; }

/* View Bar */
.view-bar {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  border-bottom: 1px solid var(--kx-border);
  margin-bottom: 0;
}
.view-tabs {
  display: flex;
  gap: 0;
  align-items: flex-end;
}
.view-tab {
  font-size: 14px;
  color: var(--kx-text-secondary);
  cursor: pointer;
  padding: 8px 16px 10px;
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
  white-space: nowrap;
}
.view-tab.active {
  color: var(--kx-primary);
  border-color: var(--kx-primary);
  font-weight: 500;
}
.view-tab:hover:not(.active) {
  color: var(--kx-text-primary);
}
.view-tab-add {
  color: var(--kx-primary);
  display: flex;
  align-items: center;
  gap: 4px;
}
.custom-view-tab {
  position: relative;
  padding-right: 28px;
}
.custom-view-text {
  color: var(--kx-primary);
}
.cv-close {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 12px;
  color: var(--kx-text-placeholder);
  cursor: pointer;
}
.cv-close:hover { color: var(--kx-danger); }
.view-actions {
  display: flex;
  gap: 12px;
  align-items: center;
  padding-bottom: 8px;
}
.view-action {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: var(--kx-text-secondary);
  cursor: pointer;
  white-space: nowrap;
}
.view-action:hover { color: var(--kx-primary); }
.view-toggle {
  font-size: 18px;
  cursor: pointer;
  color: var(--kx-text-placeholder);
  padding: 4px;
  border-radius: 4px;
}
.view-toggle.active {
  color: var(--kx-primary);
  background: rgba(51,112,255,0.08);
}

/* Filter Panel */
.filter-panel {
  background: #fafbfc;
  border: 1px solid var(--kx-border);
  border-top: none;
  border-radius: 0 0 8px 8px;
  padding: 12px 16px;
  margin-bottom: 4px;
}
.filter-row {
  display: flex;
  align-items: flex-end;
  gap: 12px;
  flex-wrap: wrap;
}
.filter-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.filter-group label {
  font-size: 12px;
  color: var(--kx-text-placeholder);
}

/* Display Settings */
.display-settings-panel {
  position: absolute;
  right: 24px;
  z-index: 200;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  padding: 12px 16px;
  min-width: 180px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.display-title {
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 4px;
}

/* Table */
.doc-list-area {
  flex: 1;
  margin-top: 4px;
  overflow-x: auto;
  overflow-y: auto;
}
.doc-list-area :deep(.el-table) {
  width: 100% !important;
  min-width: 600px;
}
.doc-list-area :deep(.el-table__body-wrapper) {
  overflow-x: auto;
}
.doc-list-area :deep(.el-table__header-wrapper) {
  overflow-x: auto;
}
.doc-name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}
.doc-title-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.pin-badge { flex-shrink: 0; }
.location-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--kx-text-secondary);
}
.owner-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}
.more-btn {
  cursor: pointer;
  font-size: 16px;
  color: var(--kx-text-placeholder);
}
.more-btn:hover { color: var(--kx-primary); }

/* Import File List */
.import-file-list {
  max-height: 200px;
  overflow-y: auto;
  margin-top: 12px;
  border: 1px solid var(--kx-border);
  border-radius: 6px;
}
.import-file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-bottom: 1px solid var(--kx-border);
  font-size: 13px;
}
.import-file-item:last-child {
  border-bottom: none;
}
.import-file-item .file-size {
  margin-left: auto;
  color: var(--kx-text-placeholder);
  font-size: 12px;
}

/* Grid */
.doc-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
  padding-top: 12px;
}
.doc-card {
  position: relative;
  padding: 16px;
  border: 1px solid var(--kx-border);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  background: #fff;
}
.doc-card:hover {
  border-color: var(--kx-primary);
  box-shadow: 0 2px 12px rgba(51,112,255,0.1);
}
.doc-card-check {
  position: absolute;
  top: 8px;
  left: 8px;
  opacity: 0;
  transition: opacity 0.15s;
}
.doc-card:hover .doc-card-check { opacity: 1; }
.doc-card-icon { margin-bottom: 12px; }
.doc-card-title {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.doc-card-meta {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--kx-text-placeholder);
}
.dot { margin: 0 2px; }

.end-marker {
  text-align: center;
  padding: 24px;
  color: var(--kx-text-placeholder);
  font-size: 13px;
  position: relative;
}
.end-marker::before,
.end-marker::after {
  content: '';
  position: absolute;
  top: 50%;
  width: 80px;
  height: 1px;
  background: var(--kx-border);
}
.end-marker::before { right: calc(50% + 60px); }
.end-marker::after { left: calc(50% + 60px); }
.empty-state { padding: 60px 0; }

/* Batch Bar */
.batch-bar {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 10px;
  box-shadow: 0 4px 20px rgba(0,0,0,0.15);
  padding: 10px 20px;
  display: flex;
  align-items: center;
  gap: 12px;
  z-index: 100;
}
.batch-info {
  font-size: 13px;
  color: var(--kx-primary);
  font-weight: 500;
}

/* Context Menu */
.context-menu {
  position: fixed;
  z-index: 999;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 10px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  padding: 4px 0;
  min-width: 220px;
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
.ctx-item:hover { background: var(--kx-sidebar-bg); }
.ctx-item.danger { color: var(--kx-danger); }
.ctx-item-toggle {
  cursor: default;
}
.ctx-switch {
  margin-left: auto;
}
.ctx-sep { height: 1px; background: var(--kx-border); margin: 4px 0; }

/* Move Dialog */
.move-folder-list {
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
}
.move-folder-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  cursor: pointer;
  font-size: 13px;
  border-bottom: 1px solid var(--kx-border);
}
.move-folder-item:last-child { border-bottom: none; }
.move-folder-item:hover { background: var(--kx-sidebar-bg); }
.move-folder-item.active { background: rgba(51,112,255,0.08); color: var(--kx-primary); }

/* Share Dialog */
.share-section { }
.share-subtitle {
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 8px;
}
.share-invite-row {
  display: flex;
  gap: 8px;
}
.share-link-row {
  display: flex;
  gap: 8px;
}
.share-options {
  margin-top: 4px;
}
.dropdown-submenu {
  position: absolute;
  left: 100%;
  top: 0;
  min-width: 160px;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  padding: 4px 0;
  z-index: 101;
}
</style>
