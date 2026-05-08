<template>
  <div class="home-page">
<!-- Top Action Bar (fixed, 3 cards) -->
    <div class="action-bar">
      <div class="action-card" @click.stop="showNewMenu = !showNewMenu; showUploadMenu = false">
        <div class="action-icon" style="background: #e8f0fe"><el-icon :size="20" color="#3370ff"><Document /></el-icon></div>
        <div class="action-info">
          <div class="action-label">{{ $t('home.new') }}</div>
          <div class="action-desc">{{ $t('home.newDesc') }}</div>
        </div>
        <el-icon class="action-arrow"><ArrowDown /></el-icon>
        <div v-if="showNewMenu" class="action-dropdown action-dropdown-large" @click.stop>
          <div class="dropdown-item" @click="handleCreate('doc')"><el-icon color="#3370ff"><Document /></el-icon>{{ $t('home.doc') }}</div>
          <div class="dropdown-item" @click="handleCreate('sheet')"><el-icon color="#36b37e"><Grid /></el-icon>{{ $t('home.sheet') }}</div>
          <div class="dropdown-item" @click="handleCreate('slide')"><el-icon color="#ff7d00"><Monitor /></el-icon>{{ $t('home.slide') }}</div>
          <div class="dropdown-item" @click="handleCreate('bitable')"><el-icon color="#00b8d9"><Tickets /></el-icon>{{ $t('home.bitable') }}</div>
          <div class="dropdown-item" @click="handleCreate('survey')"><el-icon color="#f54a45"><Notebook /></el-icon>{{ $t('home.survey') }}</div>
          <div class="dropdown-item" @click="handleCreate('mindnote')"><el-icon color="#9254de"><Share /></el-icon>{{ $t('home.mindNote') }}</div>
          <div class="dropdown-item dropdown-item-arrow" @click.stop="showMoreTypes = !showMoreTypes"><el-icon color="#36b37e"><Grid /></el-icon>{{ $t('home.moreTypes') }}<el-icon class="arrow-right"><ArrowRight /></el-icon>
            <div v-if="showMoreTypes" class="dropdown-submenu" @click.stop>
              <div class="dropdown-item" @click="handleCreate('whiteboard')"><el-icon color="#3370ff"><Document /></el-icon>{{ $t('home.whiteboard') }}</div>
              <div class="dropdown-item" @click="handleCreate('uml')"><el-icon color="#ff7d00"><Connection /></el-icon>{{ $t('home.uml') }}</div>
              <div class="dropdown-item" @click="handleCreate('gantt')"><el-icon color="#36b37e"><TrendCharts /></el-icon>{{ $t('home.gantt') }}</div>
              <div class="dropdown-item" @click="handleCreate('flowchart')"><el-icon color="#9254de"><Share /></el-icon>{{ $t('home.orgChart') }}</div>
            </div>
          </div>
          <div class="dropdown-sep" />
          <div class="dropdown-item" @click="handleCreate('folder')"><el-icon color="#f5a623"><Folder /></el-icon>{{ $t('home.folder') }}</div>
          <div class="dropdown-group-title">{{ $t('home.docApps') }}</div>
          <div class="dropdown-item" @click="handleCreate('doc')"><el-icon color="#36b37e"><EditPen /></el-icon>{{ $t('home.canvas') }}</div>
          <div class="dropdown-item" @click="handleCreate('mindnote')"><el-icon color="#9254de"><Share /></el-icon>{{ $t('document.mindMap') }}</div>
          <div class="dropdown-item" @click="handleCreate('doc')"><el-icon color="#ff7d00"><Connection /></el-icon>{{ $t('home.flowchart') }}</div>
        </div>
      </div>
      <div class="action-card" @click.stop="showUploadMenu = !showUploadMenu; showNewMenu = false">
        <div class="action-icon" style="background: #e6f7ef"><el-icon :size="20" color="#36b37e"><Upload /></el-icon></div>
        <div class="action-info">
          <div class="action-label">{{ $t('home.upload') }}</div>
          <div class="action-desc">{{ $t('home.uploadDesc') }}</div>
        </div>
        <el-icon class="action-arrow"><ArrowDown /></el-icon>
        <div v-if="showUploadMenu" class="action-dropdown" @click.stop>
          <div class="dropdown-item" @click="triggerUpload('file')"><el-icon color="#f5a623"><Document /></el-icon>{{ $t('home.uploadFile') }}</div>
          <div class="dropdown-item" @click="triggerUpload('folder')"><el-icon color="#f5a623"><Folder /></el-icon>{{ $t('home.uploadFolder') }}</div>
          <div class="dropdown-item" @click="triggerUpload('import')"><el-icon color="#3370ff"><DocumentCopy /></el-icon>{{ $t('home.importOnline') }}</div>
        </div>
      </div>
      <div class="action-card" @click="$router.push('/templates')">
        <div class="action-icon" style="background: #fef3e0"><el-icon :size="20" color="#f5a623"><Files /></el-icon></div>
        <div class="action-info">
          <div class="action-label">{{ $t('home.templateLib') }}</div>
          <div class="action-desc">{{ $t('home.templateLibDesc') }}</div>
        </div>
      </div>
    </div>

    <!-- View Tabs + Filter Bar -->
    <div class="view-bar">
      <div class="view-tabs">
        <span class="view-tab" :class="{ active: activeTab === 'recent' }" @click="switchTab('recent')">{{ $t('home.recentVisit') }}</span>
        <span class="view-tab" :class="{ active: activeTab === 'owned' }" @click="switchTab('owned')">{{ $t('home.ownedByMe') }}</span>
        <span class="view-tab" :class="{ active: activeTab === 'shared' }" @click="switchTab('shared')">{{ $t('home.sharedWithMe') }}</span>
        <span class="view-tab" :class="{ active: activeTab === 'favorites' }" @click="switchTab('favorites')">{{ $t('home.favorites') }}</span>
        <span class="view-tab" :class="{ active: activeTab === 'recycle' }" @click="switchTab('recycle')">{{ $t('document.trash') }}</span>
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
          <span class="custom-view-text">{{ $t('home.unnamedView') }}</span>
          <el-icon :size="14"><Plus /></el-icon>
        </span>
      </div>
      <div class="view-actions">
        <span class="view-action" @click.stop="showFilterPanel = !showFilterPanel"><el-icon><Filter /></el-icon> {{ $t('home.filter') }}</span>
        <span class="view-action" @click.stop="showDisplaySettings = !showDisplaySettings"><el-icon><Setting /></el-icon> {{ $t('home.displaySettings') }}</span>
        <el-icon class="view-toggle" :class="{ active: viewMode === 'list' }" @click="viewMode = 'list'"><List /></el-icon>
        <el-icon class="view-toggle" :class="{ active: viewMode === 'grid' }" @click="viewMode = 'grid'"><Grid /></el-icon>
      </div>
    </div>

    <!-- Filter Panel -->
    <div v-if="showFilterPanel" class="filter-panel" @click.stop>
      <div class="filter-row">
        <div class="filter-group">
          <label>{{ $t('home.docType') }}</label>
          <el-select v-model="filterType" :placeholder="$t('home.allTypes')" clearable size="small" style="width:140px">
            <el-option :label="$t('home.doc')" value="doc" />
            <el-option :label="$t('home.sheet')" value="sheet" />
            <el-option :label="$t('home.slide')" value="slide" />
            <el-option :label="$t('home.bitable')" value="bitable" />
            <el-option :label="$t('home.survey')" value="survey" />
            <el-option :label="$t('home.mindNote')" value="mindnote" />
            <el-option :label="$t('home.folder')" value="folder" />
          </el-select>
        </div>
        <div class="filter-group">
          <label>{{ $t('document.owner') }}</label>
          <el-input v-model="filterOwner" :placeholder="$t('home.enterUsername')" clearable size="small" style="width:140px" />
        </div>
        <div class="filter-group">
          <label>{{ $t('home.createdTime') }}</label>
          <el-select v-model="filterCreatedRange" :placeholder="$t('common.noLimit')" clearable size="small" style="width:120px">
            <el-option :label="$t('common.today')" value="today" />
            <el-option :label="$t('common.yesterday')" value="yesterday" />
            <el-option :label="$t('common.last7days')" value="7days" />
            <el-option :label="$t('common.last30days')" value="30days" />
          </el-select>
        </div>
        <div class="filter-group">
          <label>{{ $t('home.updatedTime') }}</label>
          <el-select v-model="filterUpdatedRange" :placeholder="$t('common.noLimit')" clearable size="small" style="width:120px">
            <el-option :label="$t('common.today')" value="today" />
            <el-option :label="$t('common.yesterday')" value="yesterday" />
            <el-option :label="$t('common.last7days')" value="7days" />
            <el-option :label="$t('common.last30days')" value="30days" />
          </el-select>
        </div>
        <el-button size="small" @click="applyFilter">{{ $t('common.apply') }}</el-button>
        <el-button size="small" text @click="resetFilter">{{ $t('common.reset') }}</el-button>
      </div>
    </div>

    <!-- Display Settings Panel -->
    <div v-if="showDisplaySettings" class="display-settings-panel" @click.stop>
      <div class="display-title">{{ $t('home.displayColumns') }}</div>
      <el-checkbox v-model="colVisible.title" disabled>{{ $t('home.title') }}</el-checkbox>
      <el-checkbox v-model="colVisible.fileType">{{ $t('home.fileType') }}</el-checkbox>
      <el-checkbox v-model="colVisible.fileSize">{{ $t('home.fileSize') }}</el-checkbox>
      <el-checkbox v-model="colVisible.location">{{ $t('home.location') }}</el-checkbox>
      <el-checkbox v-model="colVisible.owner">{{ $t('document.owner') }}</el-checkbox>
      <el-checkbox v-model="colVisible.createdAt">{{ $t('home.createdTime') }}</el-checkbox>
      <el-checkbox v-model="colVisible.updatedAt">{{ $t('home.updatedTime') }}</el-checkbox>
    </div>

    <!-- Document List -->
    <div class="doc-list-area">
      <!-- Skeleton Loading -->
      <div v-if="loading" class="doc-list-skeleton">
        <div v-for="i in 4" :key="i" class="doc-skeleton-row">
          <el-skeleton animated :loading="true">
            <template #template>
              <div class="skeleton-row-inner">
                <el-skeleton-item variant="circle" style="width: 16px; height: 16px; flex-shrink: 0" />
                <el-skeleton-item variant="text" style="width: 40%; height: 16px" />
                <el-skeleton-item variant="text" style="width: 60px; height: 16px" />
                <el-skeleton-item variant="text" style="width: 80px; height: 16px" />
                <el-skeleton-item variant="text" style="width: 100px; height: 16px" />
              </div>
            </template>
          </el-skeleton>
        </div>
      </div>

      <!-- List View -->
      <el-table
        v-if="!loading && viewMode === 'list'"
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
        <el-table-column :label="$t('home.title')" prop="title" min-width="240" sortable="custom">
          <template #default="{ row }">
            <div class="doc-name-cell">
              <el-icon
                class="favorite-star"
                :class="{ 'is-favorited': row.isFavorite }"
                :size="16"
                @click.stop="toggleFavorite(row)"
              >
                <StarFilled v-if="row.isFavorite" />
                <Star v-else />
              </el-icon>
              <el-icon :color="getTypeColor(row.type, row.fileExt)" :size="16"><component :is="getTypeIcon(row.type, row.fileExt)" /></el-icon>
              <span class="doc-title-text">{{ getDisplayName(row) }}</span>
              <el-icon v-if="row.isPinned" class="pin-badge" color="#3370ff" :size="12"><Flag /></el-icon>
            </div>
          </template>
        </el-table-column>
        <el-table-column v-if="colVisible.fileType" :label="$t('home.fileType')" min-width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="getTypeTagType(row.type)" disable-transitions>{{ getTypeName(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column v-if="colVisible.fileSize" :label="$t('home.fileSize')" min-width="100">
          <template #default="{ row }">
            <span class="size-cell">{{ row.fileSize > 0 ? formatFileSize(row.fileSize) : (row.type === 'folder' ? '—' : '—') }}</span>
          </template>
        </el-table-column>
        <el-table-column v-if="colVisible.location" :label="$t('home.location')" min-width="140">
          <template #default="{ row }">
            <div class="location-cell">
              <el-icon :size="14" color="#f5a623"><FolderOpened /></el-icon>
              <span>{{ row.parentId ? $t('home.myDocLib') : $t('home.cloudDrive') }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column v-if="colVisible.owner" :label="$t('document.owner')" min-width="100">
          <template #default="{ row }">
            <div class="owner-cell">
              <el-avatar :size="20" :style="{ background: getAvatarColor(row.ownerId) }">{{ (row.ownerName || 'U')[0] }}</el-avatar>
              <span>{{ row.ownerName || $t('common.unknown') }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column v-if="colVisible.createdAt" :label="$t('home.createdTime')" prop="createdAt" min-width="130" sortable="custom">
          <template #header>
            <span>{{ $t('home.createdTime') }} {{ sortField === 'createdAt' ? (sortOrder === 'descending' ? '↓' : '↑') : '' }}</span>
          </template>
          <template #default="{ row }">{{ formatDate(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column v-if="colVisible.updatedAt" :label="$t('home.updatedTime')" prop="updatedAt" min-width="130" sortable="custom">
          <template #header>
            <span>{{ $t('home.updatedTime') }} {{ sortField === 'updatedAt' ? (sortOrder === 'descending' ? '↓' : '↑') : '' }}</span>
          </template>
          <template #default="{ row }">{{ formatDate(row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column v-if="activeTab === 'recycle'" :label="$t('home.autoCleanCountdown')" min-width="140">
          <template #default="{ row }">
            <el-tag :type="getRemainingDays(row.deletedAt) <= 7 ? 'danger' : 'warning'" size="small">
              {{ $t('home.remainingDays', { days: getRemainingDays(row.deletedAt) }) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column v-if="activeTab === 'recycle'" :label="$t('home.operations')" min-width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" link @click.stop="handleRestoreSingle(row)">{{ $t('home.restore') }}</el-button>
            <el-button size="small" type="danger" link @click.stop="handlePermanentDeleteSingle(row)">{{ $t('home.permanentDelete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- Grid View -->
      <div v-else-if="!loading && viewMode === 'grid'" class="doc-grid">
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
          <div class="doc-card-star" @click.stop="toggleFavorite(doc)">
            <el-icon :class="{ 'is-favorited': doc.isFavorite }">
              <StarFilled v-if="doc.isFavorite" />
              <Star v-else />
            </el-icon>
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
        <el-empty :description="$t('home.noDocument')" />
      </div>
    </div>

    <div class="end-marker" v-if="filteredDocuments.length">
      <span>{{ $t('home.endOfList') }}</span>
    </div>

    <!-- Batch Action Bar -->
    <div v-if="selectedDocs.length && activeTab !== 'recycle'" class="batch-bar">
      <span class="batch-info">{{ $t('common.selected', { count: selectedDocs.length }) }}</span>
      <el-button size="small" @click="batchAction('move')">{{ $t('home.moveTo') }}</el-button>
      <el-button size="small" @click="batchAction('share')">{{ $t('common.share') }}</el-button>
      <el-button size="small" type="danger" @click="batchAction('delete')">{{ $t('common.delete') }}</el-button>
      <el-button size="small" text @click="clearSelection">{{ $t('home.cancelSelect') }}</el-button>
    </div>

    <!-- Recycle Bin Batch Action Bar -->
    <div v-if="selectedDocs.length && activeTab === 'recycle'" class="batch-bar">
      <span class="batch-info">{{ $t('common.selected', { count: selectedDocs.length }) }}</span>
      <el-button size="small" type="primary" @click="handleBatchRestore">{{ $t('home.batchRestore') }}</el-button>
      <el-button size="small" type="danger" @click="handleBatchPermanentDelete">{{ $t('home.permanentDelete') }}</el-button>
      <el-button size="small" text @click="clearSelection">{{ $t('home.cancelSelect') }}</el-button>
    </div>

    <!-- Context Menu (right-click / ... button) -->
    <div v-if="contextMenu.visible" class="context-menu" :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }">
      <div class="ctx-item" @click="handleAction('share')"><el-icon><Share /></el-icon>{{ $t('common.share') }}</div>
      <div class="ctx-item" @click="handleAction('copyLink')"><el-icon><Link /></el-icon>{{ $t('home.copyLink') }}</div>
      <div v-if="contextMenu.doc?.type === 'file' || contextMenu.doc?.fileSize" class="ctx-item" @click="handleAction('download')"><el-icon><Download /></el-icon>{{ $t('home.downloadOriginal') }}</div>
      <div class="ctx-sep" />
      <div class="ctx-item" @click="handleAction('copy')"><el-icon><DocumentCopy /></el-icon>{{ $t('home.createCopy') }}</div>
      <div class="ctx-item" @click="handleAction('shortcut')"><el-icon><Position /></el-icon>{{ $t('home.addShortcut') }}</div>
      <div class="ctx-item" @click="handleAction('pin')">
        <el-icon><Flag /></el-icon>{{ contextMenu.doc?.isPinned ? $t('home.removeFromTop') : $t('home.addToTop') }}
      </div>
      <div class="ctx-item" @click="handleAction('favorite')">
        <el-icon><Star /></el-icon>{{ contextMenu.doc?.isFavorite ? $t('home.cancelFavorite') : $t('home.addFavorite') }}
      </div>
      <div class="ctx-sep" />
      <div class="ctx-item ctx-item-toggle" @click.stop="handleAction('offline')">
        <el-icon><Download /></el-icon>{{ $t('home.offlineAvailable') }}
        <el-switch v-model="offlineToggle" size="small" class="ctx-switch" @click.stop />
      </div>
      <div class="ctx-item ctx-item-toggle" @click.stop="handleAction('follow')">
        <el-icon><BellFilled /></el-icon>{{ $t('home.followUpdates') }}
        <el-switch v-model="followToggle" size="small" class="ctx-switch" @click.stop />
      </div>
      <div class="ctx-sep" />
      <div class="ctx-item" @click="handleAction('move')"><el-icon><Rank /></el-icon>{{ $t('home.moveTo') }}</div>
      <div class="ctx-item" @click="handleAction('rename')"><el-icon><EditPen /></el-icon>{{ $t('common.rename') }}</div>
      <div class="ctx-item" @click="handleAction('transfer')"><el-icon><Switch /></el-icon>{{ $t('home.transferOwnership') }}</div>
      <div class="ctx-item danger" @click="handleAction('delete')"><el-icon><Delete /></el-icon>{{ $t('common.delete') }}</div>
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
        :on-exceed="() => ElMessage.warning(t('home.fileLimitExceed'))"
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">{{ t('home.dragFolderHint') }} <em>{{ t('home.clickSelectFolder') }}</em></div>
        <template #tip><div class="el-upload__tip">{{ t('home.batchUploadHint') }}</div></template>
      </el-upload>
      <el-upload
        v-else-if="uploadType === 'file'"
        drag
        :auto-upload="false"
        :limit="10"
        multiple
        :on-change="handleFileSelect"
        :on-exceed="() => ElMessage.warning(t('home.maxUpload10'))"
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">{{ t('home.dragFileHint') }} <em>{{ t('home.clickUpload') }}</em></div>
        <template #tip><div class="el-upload__tip">{{ t('home.supportAllFormats') }}</div></template>
      </el-upload>
      <el-upload
        v-else
        drag
        :auto-upload="false"
        :limit="1"
        accept=".md,.json,.txt,.html,.docx,.doc,.xlsx,.xls,.pptx,.ppt,.pdf,.png,.jpg,.jpeg,.gif,.webp"
        :on-change="(f: any) => importFile = f.raw"
        :on-exceed="() => ElMessage.warning(t('home.uploadOnly1'))"
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">{{ t('home.dragFileHint') }} <em>{{ t('home.clickUpload') }}</em></div>
        <template #tip><div class="el-upload__tip">{{ t('home.supportFormats') }}</div></template>
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
        <el-button @click="showImportDialog = false; importFile = null; importFiles = []">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="importLoading" :disabled="!importFile && !importFiles.length" @click="handleImport">{{ $t('common.import') }}</el-button>
      </template>
    </el-dialog>

    <!-- Move Dialog -->
    <el-dialog v-model="showMoveDialog" :title="$t('home.moveTo')" width="420px" destroy-on-close>
      <div class="move-folder-list">
        <div class="move-folder-item" :class="{ active: moveTarget === null }" @click="moveTarget = null">
          <el-icon color="#f5a623"><FolderOpened /></el-icon><span>{{ $t('home.rootDir') }}</span>
        </div>
        <div v-for="f in allFolders" :key="f.id" class="move-folder-item" :class="{ active: moveTarget === f.id }" @click="moveTarget = f.id">
          <el-icon color="#f5a623"><Folder /></el-icon><span>{{ f.title }}</span>
        </div>
      </div>
      <template #footer>
        <el-button @click="showMoveDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="confirmMove">{{ $t('home.confirmMove') }}</el-button>
      </template>
    </el-dialog>

    <!-- Rename Dialog -->
    <el-dialog v-model="showRenameDialog" :title="$t('common.rename')" width="400px" destroy-on-close>
      <el-input v-model="renameValue" :placeholder="$t('home.enterNewName')" maxlength="100" show-word-limit @keyup.enter="confirmRename" />
      <template #footer>
        <el-button @click="showRenameDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="renaming" @click="confirmRename">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- Transfer Dialog -->
    <el-dialog v-model="showTransferDialog" :title="$t('home.transferOwnership')" width="420px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item :label="$t('home.selectTargetUser')">
          <el-input v-model="transferUser" :placeholder="$t('home.searchUserOrEmail')" prefix-icon="Search" />
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="transferKeepPerm">{{ $t('home.keepPermission') }}</el-checkbox>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showTransferDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="confirmTransfer">{{ $t('home.confirmTransfer') }}</el-button>
      </template>
    </el-dialog>

    <!-- Share Dialog -->
    <el-dialog v-model="showShareDialog" :title="$t('common.share')" width="520px" destroy-on-close>
      <div class="share-section">
        <div class="share-subtitle">{{ $t('home.inviteCollaborator') }}</div>
        <div class="share-invite-row">
          <el-input v-model="shareInvite" :placeholder="$t('home.enterUserOrEmail')" style="flex:1" />
          <el-select v-model="sharePermLevel" style="width:120px">
            <el-option :label="$t('home.canManage')" value="manage" />
            <el-option :label="$t('home.canEdit')" value="edit" />
            <el-option :label="$t('home.canView')" value="view" />
            <el-option :label="$t('home.readOnly')" value="readonly" />
          </el-select>
          <el-button type="primary" @click="ElMessage.info(t('home.inviteSent'))">{{ $t('common.invite') }}</el-button>
        </div>
      </div>
      <div class="share-section" style="margin-top:16px">
        <div class="share-subtitle">{{ $t('home.shareLink') }}</div>
        <div class="share-link-row">
          <el-input :model-value="shareLink" readonly style="flex:1" />
          <el-button @click="copyShareLink">{{ $t('common.copy') }}</el-button>
        </div>
        <div class="share-options">
          <el-select v-model="shareLinkScope" size="small" style="width:160px;margin-top:8px">
            <el-option :label="$t('home.collaboratorOnly')" value="collaborator" />
            <el-option :label="$t('home.orgVisible')" value="org" />
            <el-option :label="$t('home.publicVisible')" value="public" />
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
  copyDocument, importDocument, moveDocument, updateDocument, transferOwnership,
  getRecycleBin, permanentDeleteDocument, batchRestoreDocuments, restoreDocument
} from '@/api/modules/document'
import type { Document, DocumentType, KnowledgeBase } from '@/types'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const router = useRouter()
const route = useRoute()
const loading = ref(false)
const viewMode = ref<'list' | 'grid'>('list')
const documents = ref<Document[]>([])
const allFolders = ref<Document[]>([])
const activeTab = ref('recent')
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

function getTypeNameMap(): Record<string, string> {
  return {
    folder: t('home.folder'), doc: t('home.doc'), sheet: t('home.sheet'), slide: t('home.slide'),
    mindnote: t('home.mindNote'), bitable: t('home.bitable'), survey: t('home.survey'),
    file: t('home.other'), image: t('home.canvas'), code: t('document.codeEditor'),
  }
}
function getExtTypeMap(): Record<string, string> {
  return {
    // 文档类
    txt: t('fileType.textDoc'),
    doc: t('fileType.wordDoc'),
    docx: t('fileType.wordDoc'),
    pdf: t('fileType.pdfDoc'),
    rtf: t('fileType.richText'),
    epub: t('fileType.ebook'),
    mobi: t('fileType.ebook'),

    // 表格/数据类
    xls: t('fileType.excelSheet'),
    xlsx: t('fileType.excelSheet'),
    csv: t('fileType.csvSheet'),
    db: t('fileType.database'),
    sqlite: t('fileType.database'),
    sql: t('fileType.database'),

    // 演示文稿
    ppt: t('fileType.presentation'),
    pptx: t('fileType.presentation'),
    pot: t('fileType.presentationTemplate'),

    // 图片类
    png: t('fileType.pngImage'),
    jpg: t('fileType.jpegImage'),
    jpeg: t('fileType.jpegImage'),
    gif: t('fileType.gifImage'),
    bmp: t('fileType.bmpImage'),
    webp: t('fileType.webpImage'),
    svg: t('fileType.vectorImage'),
    ico: t('fileType.iconFile'),

    // 音频
    mp3: t('fileType.audioFile'),
    wav: t('fileType.losslessAudio'),
    flac: t('fileType.flacAudio'),
    aac: t('fileType.aacAudio'),
    ogg: t('fileType.oggAudio'),
    m4a: t('fileType.m4aAudio'),

    // 视频
    mp4: t('fileType.videoFile'),
    mkv: t('fileType.mkvVideo'),
    avi: t('fileType.aviVideo'),
    mov: t('fileType.movVideo'),
    wmv: t('fileType.wmvVideo'),
    flv: t('fileType.flvVideo'),
    webm: t('fileType.webmVideo'),

    // 压缩包
    zip: t('fileType.archive'),
    rar: t('fileType.archive'),
    '7z': t('fileType.7zArchive'),
    tar: t('fileType.tarArchive'),
    gz: t('fileType.gzArchive'),
    'tar.gz': t('fileType.tarGzArchive'),
    bz2: t('fileType.bz2Archive'),

    // 系统/安装包
    exe: t('fileType.winProgram'),
    msi: t('fileType.winInstaller'),
    dll: t('fileType.systemLib'),
    apk: t('fileType.androidApp'),
    aab: t('fileType.androidBundle'),
    ipa: t('fileType.iosApp'),
    app: t('fileType.macApp'),
    dmg: t('fileType.macDiskImage'),
    deb: t('fileType.debPackage'),
    rpm: t('fileType.rpmPackage'),
    AppImage: t('fileType.linuxPortable'),
    sh: t('fileType.shellScript'),
    iso: t('fileType.discImage'),

    // 网页/代码
    html: t('fileType.webPage'),
    htm: t('fileType.webPage'),
    css: t('fileType.stylesheet'),
    js: 'JavaScript',
    ts: 'TypeScript',
    vue: t('fileType.vueComponent'),
    jsx: t('fileType.reactComponent'),
    md: t('fileType.markdownDoc'),
    json: t('fileType.jsonConfig'),
    xml: t('fileType.xmlFile'),
    yaml: t('fileType.yamlFile'),
    yml: t('fileType.yamlFile'),

    // 编程源码
    c: t('fileType.cCode'),
    cpp: t('fileType.cppCode'),
    h: t('fileType.headerFile'),
    java: t('fileType.javaCode'),
    class: t('fileType.javaCompiled'),
    jar: t('fileType.javaPackage'),
    py: t('fileType.pythonCode'),
    go: t('fileType.goCode'),
    php: t('fileType.phpScript'),

    // 其他
    log: t('fileType.logFile'),
    tmp: t('fileType.tempFile'),
    vmdk: t('fileType.vmDisk'),
    vdi: t('fileType.vmDisk'),
  }
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
  const extTypeMap = getExtTypeMap()
  if (ext && extTypeMap[ext]) return extTypeMap[ext]
  const nameMap = getTypeNameMap()
  if (nameMap[doc.type]) return nameMap[doc.type]
  return t('home.other')
}

function getTypeTagType(type: string): '' | 'success' | 'warning' | 'info' | 'danger' {
  const m: Record<string, '' | 'success' | 'warning' | 'info' | 'danger'> = {
    folder: 'warning', doc: '', sheet: 'success', slide: 'warning',
    mindnote: '', bitable: 'info', survey: 'danger', file: 'info',
  }
  return m[type] || 'info'
}

function formatDate(dateStr: string) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const now = new Date()
  const isToday = d.toDateString() === now.toDateString()
  if (isToday) return `${t('common.today')} ${d.getHours().toString().padStart(2,'0')}:${d.getMinutes().toString().padStart(2,'0')}`
  return d.toLocaleDateString()
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
    } else if (activeTab.value === 'recycle') {
      res = await getRecycleBin({ page: page.value, pageSize })
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
    const res: any = await createDocument({ title: type === 'folder' ? t('document.newFolder') : `${t('document.untitled')}${typeLabel(type)}`, type, parentId: null })
    if (type !== 'folder') {
      // Open in new browser tab
      window.open(`/doc/${res.data.id}`, '_blank')
    } else {
      ElMessage.success(t('home.folderCreated'))
      fetchDocuments()
    }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || t('home.createFailed'))
  }
}

function typeLabel(type: string): string {
  const nameMap = getTypeNameMap()
  return nameMap[type] || t('home.doc')
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
    await ElMessageBox.confirm(t('home.deleteConfirmBatch', { count: selectedDocs.value.length }), t('home.deleteConfirmTitle'))
    for (const d of selectedDocs.value) { await deleteDocument(d.id) }
    ElMessage.success(t('home.movedToTrash'))
    fetchDocuments()
  } else if (action === 'move') {
    showMoveDialog.value = true
  } else if (action === 'share') {
    // Batch share - open share dialog with batch link
    const ids = selectedDocs.value.map(d => d.id).join(',')
    showShareDialog.value = true
    // shareLink is computed, so we set contextMenu.doc to null to show batch link
    ElMessage.info(t('home.batchShareGenerated', { count: selectedDocs.value.length }))
    navigator.clipboard.writeText(`${window.location.origin}/share?docs=${ids}`)
  }
  clearSelection()
}

// Custom views
function addCustomView() {
  customViews.value.push({ id: cvIdCounter++, name: t('home.unnamedView') })
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
    case 'file': return t('home.uploadDialogFile')
    case 'folder': return t('home.uploadDialogFolder')
    default: return t('home.uploadDialogImport')
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
      ElMessage.success(t('home.importSuccess'))
      showImportDialog.value = false
      importFile.value = null
      importProgress.value = 0
      fetchDocuments()
      if (res.data?.id && res.data?.type !== 'folder') {
        window.open(`/doc/${res.data.id}`, '_blank')
      }
    } catch { ElMessage.error(t('home.importFailed')) } finally { importLoading.value = false }
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
      } catch (e: any) {
        console.error('[HomePage] importDocument failed:', e)
        failCount++
      }
    }

    importLoading.value = false
    importProgress.value = 0

    if (successCount > 0) {
      ElMessage.success(failCount > 0 ? t('home.uploadSuccessWithFail', { success: successCount, fail: failCount }) : t('home.uploadSuccess', { success: successCount }))
      showImportDialog.value = false
      importFiles.value = []
      fetchDocuments()
    } else {
      ElMessage.error(t('home.uploadFailed'))
    }
  }
}

async function toggleFavorite(doc: Document) {
  const newState = !doc.isFavorite
  try {
    await favoriteDocument(doc.id, newState)
    doc.isFavorite = newState
    ElMessage.success(newState ? t('home.favorited') : t('home.unfavorited'))
  } catch {
    ElMessage.error(t('home.operationFailed'))
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
      ElMessage.success(t('common.linkCopied'))
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
      ElMessage.success(t('home.copyCreated'))
      fetchDocuments()
      break
    case 'shortcut':
      ElMessage.success(t('home.shortcutAdded'))
      break
    case 'pin':
      await pinDocument(doc.id, !doc.isPinned)
      ElMessage.success(doc.isPinned ? t('home.removedFromTop') : t('home.addedToTop'))
      fetchDocuments()
      break
    case 'favorite':
      await favoriteDocument(doc.id, !doc.isFavorite)
      ElMessage.success(doc.isFavorite ? t('home.unfavorited') : t('home.favorited'))
      fetchDocuments()
      break
    case 'offline':
      offlineToggle.value = !offlineToggle.value
      ElMessage.success(offlineToggle.value ? t('home.offlineEnabled') : t('home.offlineDisabled'))
      break
    case 'follow':
      followToggle.value = !followToggle.value
      ElMessage.success(followToggle.value ? t('home.followEnabled') : t('home.followDisabled'))
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
      await ElMessageBox.confirm(t('home.deleteConfirmSingle', { title: doc.title }), t('home.deleteConfirmTitle'))
      await deleteDocument(doc.id)
      ElMessage.success(t('home.movedToTrash'))
      fetchDocuments()
      break
  }
}

async function confirmMove() {
  const doc = contextMenu.doc
  if (doc) {
    await moveDocument(doc.id, moveTarget.value)
    ElMessage.success(t('home.moved'))
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
  if (!doc || !renameValue.value.trim()) { ElMessage.warning(t('home.nameRequired')); return }
  renaming.value = true
  try {
    await updateDocument(doc.id, { title: renameValue.value.trim() })
    ElMessage.success(t('home.renameSuccess'))
    showRenameDialog.value = false
    fetchDocuments()
  } catch { ElMessage.error(t('home.renameFailed')) }
  finally { renaming.value = false }
}

async function confirmTransfer() {
  if (!transferUser.value?.trim()) {
    ElMessage.warning(t('home.enterTargetUser'))
    return
  }
  const doc = contextMenu.doc
  if (!doc) return
  try {
    await transferOwnership(doc.id, transferUser.value.trim(), transferKeepPerm.value)
    ElMessage.success(t('home.ownershipTransferred'))
    showTransferDialog.value = false
    fetchDocuments()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || t('home.transferFailed'))
  }
}

function copyShareLink() {
  if (shareLink.value) {
    navigator.clipboard.writeText(shareLink.value)
    ElMessage.success(t('common.linkCopied'))
  }
}

// === Recycle Bin Functions ===

// Calculate remaining days before auto-cleanup (30 days from deletion)
function getRemainingDays(deletedAt: string): number {
  if (!deletedAt) return 30
  const deleted = new Date(deletedAt)
  const expireDate = new Date(deleted.getTime() + 30 * 24 * 60 * 60 * 1000)
  const now = new Date()
  const remaining = Math.ceil((expireDate.getTime() - now.getTime()) / (24 * 60 * 60 * 1000))
  return remaining > 0 ? remaining : 0
}

async function handleRestoreSingle(doc: Document) {
  try {
    await restoreDocument(doc.id)
    ElMessage.success(t('home.restored', { title: doc.title }))
    fetchDocuments()
  } catch {
    ElMessage.error(t('home.restoreFailed'))
  }
}

async function handlePermanentDeleteSingle(doc: Document) {
  await ElMessageBox.confirm(t('home.permanentDeleteConfirm'), t('home.permanentDeleteTitle'), { type: 'warning' })
  try {
    await permanentDeleteDocument(doc.id)
    ElMessage.success(t('home.permanentDeleted'))
    fetchDocuments()
  } catch {
    ElMessage.error(t('common.failed'))
  }
}

async function handleBatchRestore() {
  const ids = selectedDocs.value.map(d => d.id)
  try {
    await batchRestoreDocuments(ids)
    ElMessage.success(t('home.batchRestored', { count: ids.length }))
    clearSelection()
    fetchDocuments()
  } catch {
    ElMessage.error(t('home.batchRestoreFailed'))
  }
}

async function handleBatchPermanentDelete() {
  await ElMessageBox.confirm(t('home.batchPermanentDeleteConfirm', { count: selectedDocs.value.length }), t('home.permanentDeleteTitle'), { type: 'warning' })
  try {
    for (const doc of selectedDocs.value) {
      await permanentDeleteDocument(doc.id)
    }
    ElMessage.success(t('home.permanentDeleted'))
    clearSelection()
    fetchDocuments()
  } catch {
    ElMessage.error(t('common.failed'))
  }
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
.favorite-star {
  cursor: pointer;
  color: var(--kx-text-placeholder);
  flex-shrink: 0;
  transition: color 0.2s;
}
.favorite-star:hover {
  color: #f5a623;
}
.favorite-star.is-favorited {
  color: #f5a623;
}
.doc-card-star {
  position: absolute;
  top: 8px;
  right: 8px;
  cursor: pointer;
  color: var(--kx-text-placeholder);
  transition: color 0.2s, opacity 0.15s;
  opacity: 0;
}
.doc-card:hover .doc-card-star { opacity: 1; }
.doc-card-star .is-favorited {
  color: #f5a623;
}
.doc-card-star:hover {
  color: #f5a623;
}
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

/* Document List Skeleton */
.doc-list-skeleton {
  padding: 12px 0;
}
.doc-skeleton-row {
  padding: 12px 16px;
  border-bottom: 1px solid var(--kx-border);
}
.skeleton-row-inner {
  display: flex;
  align-items: center;
  gap: 12px;
}

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

/* ===== Responsive: 768px - Tablet ===== */
@media (max-width: 768px) {
  .home-page {
    padding: 16px 16px;
  }
  .action-bar {
    grid-template-columns: repeat(2, 1fr);
    gap: 10px;
  }
  .action-card {
    padding: 12px 14px;
  }
  .action-desc {
    display: none;
  }
  .view-bar {
    flex-wrap: wrap;
    gap: 8px;
  }
  .view-tabs {
    overflow-x: auto;
    white-space: nowrap;
    flex: 1;
  }
  .view-actions {
    flex-shrink: 0;
  }
  .view-action span {
    display: none;
  }
  /* Grid view columns reduce */
  .doc-grid {
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)) !important;
  }
}

/* ===== Responsive: 640px - Large Phone ===== */
@media (max-width: 640px) {
  .home-page {
    padding: 12px 12px;
  }
  .action-bar {
    grid-template-columns: 1fr;
    gap: 8px;
    margin-bottom: 16px;
  }
  .action-card {
    padding: 12px 14px;
  }
    .view-tab {
    font-size: 13px;
    padding: 6px 10px;
  }
  .custom-view-text {
    display: none;
  }
  .filter-panel {
    padding: 10px 12px;
  }
  .filter-row {
    flex-direction: column;
    gap: 8px;
  }
  .doc-grid {
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr)) !important;
  }
}

/* ===== Responsive: 480px - Small Phone ===== */
@media (max-width: 480px) {
  .home-page {
    padding: 8px 8px;
  }
  .action-bar {
    margin-bottom: 12px;
  }
  .action-icon {
    width: 30px;
    height: 30px;
  }
  .action-label {
    font-size: 13px;
  }
  .view-bar {
    margin-bottom: 8px;
  }
  .view-tab {
    font-size: 12px;
    padding: 5px 8px;
  }
  .view-actions {
    gap: 4px;
  }
  .doc-grid {
    grid-template-columns: 1fr 1fr !important;
    gap: 8px !important;
  }
}
</style>
