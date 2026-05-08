<template>
  <div class="kb-detail">
    <!-- Left Sidebar: Document Tree -->
    <aside class="kb-sidebar" :class="{ collapsed: sidebarCollapsed }">
      <div class="kb-sidebar-header">
        <div class="kb-sidebar-title" v-if="!sidebarCollapsed">
          <el-icon color="#3370ff"><Collection /></el-icon>
          <span class="kb-name-text" :title="kb?.name">{{ kb?.name }}</span>
        </div>
        <el-icon class="collapse-btn" @click="sidebarCollapsed = !sidebarCollapsed">
          <component :is="sidebarCollapsed ? 'Expand' : 'Fold'" />
        </el-icon>
      </div>

      <div v-if="!sidebarCollapsed" class="kb-sidebar-body">
        <!-- Sidebar Skeleton -->
        <div v-if="loading" class="sidebar-tree-skeleton">
          <el-skeleton v-for="i in 8" :key="i" animated :loading="true" style="padding: 4px 12px">
            <template #template>
              <div style="display: flex; align-items: center; gap: 8px; padding: 6px 0">
                <el-skeleton-item variant="rect" :style="{ width: '14px', height: '14px', flexShrink: 0, marginLeft: (i % 3 === 0 ? '16px' : '0') }" />
                <el-skeleton-item variant="text" :style="{ width: (40 + (i * 7) % 40) + '%', height: '14px' }" />
              </div>
            </template>
          </el-skeleton>
        </div>

        <template v-else>
        <!-- Search -->
        <el-input v-model="treeSearch" :placeholder="$t('knowledge.detail.searchDoc')" prefix-icon="Search" clearable size="small" class="tree-search" />

        <!-- New Document Dropdown -->
        <div class="tree-new-btn" @click.stop="showTreeNewMenu = !showTreeNewMenu">
          <el-icon><Plus /></el-icon><span>{{ $t('common.create') }}</span>
          <div v-if="showTreeNewMenu" class="tree-new-dropdown" @click.stop>
            <div class="tnd-item" @click="handleNewInKb('doc')"><el-icon color="#3370ff"><Document /></el-icon>{{ $t('home.doc') }}</div>
            <div class="tnd-item" @click="handleNewInKb('sheet')"><el-icon color="#36b37e"><Grid /></el-icon>{{ $t('home.sheet') }}</div>
            <div class="tnd-item" @click="handleNewInKb('slide')"><el-icon color="#ff7d00"><Monitor /></el-icon>{{ $t('home.slide') }}</div>
            <div class="tnd-item" @click="handleNewInKb('mindnote')"><el-icon color="#9254de"><Share /></el-icon>{{ $t('home.mindNote') }}</div>
            <div class="tnd-sep" />
            <div class="tnd-item" @click="handleNewCategory"><el-icon color="#f5a623"><Folder /></el-icon>{{ $t('knowledge.detail.categoryDir') }}</div>
          </div>
        </div>

        <!-- Category Tree -->
        <div class="tree-section">
          <div class="tree-section-label">{{ $t('knowledge.detail.directory') }}</div>
          <div v-if="!filteredTreeData.length" class="tree-empty">{{ $t('home.noDocument') }}</div>
          <div v-for="node in filteredTreeData" :key="node.id" class="tree-node-wrap">
            <div
              class="tree-node"
              :class="{ active: selectedDocId === node.id, category: node.type === 'folder' }"
              @click="handleTreeNodeClick(node)"
              @contextmenu.prevent="showTreeCtxMenu($event, node)"
            >
              <el-icon v-if="node.type === 'folder'" class="tree-expand" :class="{ expanded: node.isExpanded }" @click.stop="node.isExpanded = !node.isExpanded"><ArrowRight /></el-icon>
              <el-icon v-else class="tree-expand invisible"><ArrowRight /></el-icon>
              <el-icon :color="getTypeColor(node.type)" :size="14"><component :is="getTypeIcon(node.type)" /></el-icon>
              <span class="tree-node-label">{{ node.title }}</span>
              <span v-if="node.status" class="tree-node-status" :class="node.status">{{ statusLabel(node.status) }}</span>
            </div>
            <div v-if="node.isExpanded && node.children?.length" class="tree-children">
              <div
                v-for="child in node.children"
                :key="child.id"
                class="tree-node level-2"
                :class="{ active: selectedDocId === child.id }"
                @click="handleTreeNodeClick(child)"
                @contextmenu.prevent="showTreeCtxMenu($event, child)"
              >
                <el-icon :color="getTypeColor(child.type)" :size="14"><component :is="getTypeIcon(child.type)" /></el-icon>
                <span class="tree-node-label">{{ child.title }}</span>
                <span v-if="child.status" class="tree-node-status" :class="child.status">{{ statusLabel(child.status) }}</span>
              </div>
            </div>
          </div>
        </div>
        </template>
      </div>
    </aside>

    <!-- Main Content -->
    <div class="kb-main">
      <!-- Top Bar -->
      <div class="kb-topbar">
        <div class="kb-topbar-left">
          <el-button text @click="$router.push('/knowledge')"><el-icon><ArrowLeft /></el-icon>{{ $t('knowledge.detail.backToList') }}</el-button>
          <span class="kb-topbar-name" v-if="kb">{{ kb.name }}</span>
          <el-tag v-if="kb" size="small" type="info">{{ $t('knowledge.detail.docCountTag', { count: kb.docCount || 0 }) }}</el-tag>
        </div>
        <div class="kb-topbar-right">
          <el-input v-model="searchKey" :placeholder="$t('knowledge.detail.searchInKb')" prefix-icon="Search" clearable size="small" style="width: 240px" @keyup.enter="doSearch" />
          <el-button size="small" @click="showMembers = true"><el-icon><User /></el-icon>{{ $t('knowledge.members') }}</el-button>
          <el-button size="small" @click="openGraphPanel"><el-icon><Share /></el-icon>{{ $t('knowledge.graph') }}</el-button>
          <el-button size="small" @click="showSettings = true"><el-icon><Setting /></el-icon>{{ $t('knowledge.settings') }}</el-button>
        </div>
      </div>

      <!-- Tab Sections -->
      <div class="kb-content-tabs">
        <span class="kc-tab" :class="{ active: contentTab === 'docs' }" @click="contentTab = 'docs'">{{ $t('knowledge.detail.allDocs') }}</span>
        <span class="kc-tab" :class="{ active: contentTab === 'published' }" @click="contentTab = 'published'">{{ $t('knowledge.detail.published') }}</span>
        <span class="kc-tab" :class="{ active: contentTab === 'draft' }" @click="contentTab = 'draft'">{{ $t('knowledge.detail.draft') }}</span>
        <span class="kc-tab" :class="{ active: contentTab === 'review' }" @click="contentTab = 'review'">{{ $t('knowledge.detail.inReview') }}</span>
      </div>

      <!-- Action bar -->
      <div class="kb-action-bar">
        <div class="kb-action-left">
          <el-dropdown trigger="click" @command="handleNewInKb">
            <el-button type="primary" size="small"><el-icon><Plus /></el-icon>{{ $t('document.newDocument') }}</el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="doc"><el-icon color="#3370ff"><Document /></el-icon>{{ $t('home.doc') }}</el-dropdown-item>
                <el-dropdown-item command="sheet"><el-icon color="#36b37e"><Grid /></el-icon>{{ $t('home.sheet') }}</el-dropdown-item>
                <el-dropdown-item command="slide"><el-icon color="#ff7d00"><Monitor /></el-icon>{{ $t('home.slide') }}</el-dropdown-item>
                <el-dropdown-item command="mindnote"><el-icon color="#9254de"><Share /></el-icon>{{ $t('home.mindNote') }}</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-button size="small" @click="showImportDialog = true"><el-icon><Upload /></el-icon>{{ $t('common.import') }}</el-button>
        </div>
        <div class="kb-action-right">
          <el-dropdown trigger="click" @command="handleSort">
            <span class="action-link"><el-icon><Sort /></el-icon>{{ $t('knowledge.detail.sort') }}</span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="updated">{{ $t('knowledge.detail.updatedTime') }}</el-dropdown-item>
                <el-dropdown-item command="created">{{ $t('knowledge.detail.createdTime') }}</el-dropdown-item>
                <el-dropdown-item command="title">{{ $t('common.name') }}</el-dropdown-item>
                <el-dropdown-item command="status">{{ $t('knowledge.detail.status') }}</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <span class="action-link" @click="viewMode = viewMode === 'list' ? 'grid' : 'list'">
            <el-icon><component :is="viewMode === 'list' ? 'Grid' : 'List'" /></el-icon>{{ viewMode === 'list' ? $t('knowledge.detail.grid') : $t('knowledge.detail.list') }}
          </span>
        </div>
      </div>

      <!-- Document List -->
      <div v-if="loadError" class="kb-error-state">
        <el-icon :size="48" color="#f56c6c"><WarningFilled /></el-icon>
        <p class="error-message">{{ loadError }}</p>
        <el-button type="primary" @click="retryLoad">{{ $t('knowledge.detail.reload') }}</el-button>
      </div>
      <div v-else class="kb-doc-list">
        <!-- Document List Skeleton -->
        <div v-if="loading" class="kb-doc-skeleton">
          <el-skeleton v-for="i in 5" :key="i" animated :loading="true" style="padding: 12px 16px; border-bottom: 1px solid var(--kx-border)">
            <template #template>
              <div style="display: flex; align-items: center; gap: 12px">
                <el-skeleton-item variant="rect" style="width: 20px; height: 20px; flex-shrink: 0" />
                <el-skeleton-item variant="text" style="width: 35%; height: 16px" />
                <el-skeleton-item variant="text" style="width: 60px; height: 14px; margin-left: auto" />
                <el-skeleton-item variant="text" style="width: 100px; height: 14px" />
              </div>
            </template>
          </el-skeleton>
        </div>

        <!-- Table/List View -->
        <el-table
          v-if="!loading && viewMode === 'list'"
          :data="displayedDocs"
          style="width: 100%"
          @row-click="handleDocClick"
          @selection-change="handleSelectionChange"
        >
          <el-table-column type="selection" width="40" />
          <el-table-column :label="$t('knowledge.detail.title')" min-width="320" sortable>
            <template #default="{ row }">
              <div class="doc-name-cell">
                <el-icon :color="getTypeColor(row.type)"><component :is="getTypeIcon(row.type)" /></el-icon>
                <span>{{ row.title }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column :label="$t('knowledge.detail.status')" width="100">
            <template #default="{ row }">
              <el-tag :type="statusTagType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column :label="$t('knowledge.detail.owner')" prop="ownerName" width="120" />
          <el-table-column :label="$t('knowledge.detail.updatedTime')" width="180" sortable>
            <template #default="{ row }">{{ formatDate(row.updatedAt) }}</template>
          </el-table-column>
          <el-table-column width="50" align="center">
            <template #default="{ row }">
              <el-icon class="more-btn" @click.stop="showDocCtxMenu($event, row)"><MoreFilled /></el-icon>
            </template>
          </el-table-column>
        </el-table>

        <!-- Grid View -->
        <div v-else-if="!loading" class="kb-doc-grid">
          <div
            v-for="doc in displayedDocs"
            :key="doc.id"
            class="kb-doc-card"
            @click="handleDocClick(doc)"
            @contextmenu.prevent="showDocCtxMenu($event, doc)"
          >
            <div class="kdc-icon">
              <el-icon :size="32" :color="getTypeColor(doc.type)"><component :is="getTypeIcon(doc.type)" /></el-icon>
            </div>
            <div class="kdc-title">{{ doc.title }}</div>
            <div class="kdc-meta">
              <el-tag :type="statusTagType(doc.status)" size="small">{{ statusLabel(doc.status) }}</el-tag>
              <span>{{ formatDate(doc.updatedAt) }}</span>
            </div>
          </div>
        </div>

        <div v-if="!loading && !displayedDocs.length" class="kb-empty">
          <el-empty :description="contentTab === 'docs' ? $t('knowledge.detail.noDocHint') : $t('knowledge.detail.noDocInTab')" />
        </div>
      </div>

      <!-- Batch Bar -->
      <transition name="slide-up">
        <div v-if="selectedDocs.length" class="kb-batch-bar">
          <span>{{ $t('common.selected', { count: selectedDocs.length }) }}</span>
          <el-button size="small" @click="batchPublish"><el-icon><Upload /></el-icon>{{ $t('knowledge.detail.batchPublish') }}</el-button>
          <el-button size="small" @click="batchMove"><el-icon><Rank /></el-icon>{{ $t('common.moveTo') }}</el-button>
          <el-button size="small" type="danger" @click="batchDelete"><el-icon><Delete /></el-icon>{{ $t('common.delete') }}</el-button>
          <el-button size="small" text @click="selectedDocs = []">{{ $t('home.cancelSelect') }}</el-button>
        </div>
      </transition>
    </div>

    <!-- Document Context Menu -->
    <div v-if="docCtxMenu.visible" class="context-menu" :style="{ left: docCtxMenu.x + 'px', top: docCtxMenu.y + 'px' }">
      <div class="ctx-item" @click="handleDocAction('open')"><el-icon><View /></el-icon>{{ $t('knowledge.detail.openDoc') }}</div>
      <div class="ctx-item" @click="handleDocAction('openNew')"><el-icon><TopRight /></el-icon>{{ $t('knowledge.detail.openInNewTab') }}</div>
      <div class="ctx-sep" />
      <div class="ctx-item" @click="handleDocAction('publish')" v-if="docCtxMenu.doc?.status !== 'published'"><el-icon><Upload /></el-icon>{{ $t('knowledge.detail.publish') }}</div>
      <div class="ctx-item" @click="handleDocAction('unpublish')" v-if="docCtxMenu.doc?.status === 'published'"><el-icon><Download /></el-icon>{{ $t('knowledge.detail.unpublish') }}</div>
      <div class="ctx-item" @click="handleDocAction('submitReview')" v-if="docCtxMenu.doc?.status === 'draft'"><el-icon><Promotion /></el-icon>{{ $t('knowledge.detail.submitReview') }}</div>
      <div class="ctx-sep" />
      <div class="ctx-item" @click="handleDocAction('share')"><el-icon><Share /></el-icon>{{ $t('common.share') }}</div>
      <div class="ctx-item" @click="handleDocAction('copyLink')"><el-icon><Link /></el-icon>{{ $t('knowledge.detail.copyLink') }}</div>
      <div class="ctx-item" @click="handleDocAction('copy')"><el-icon><DocumentCopy /></el-icon>{{ $t('knowledge.detail.createCopy') }}</div>
      <div class="ctx-sep" />
      <div class="ctx-item" @click="handleDocAction('move')"><el-icon><Rank /></el-icon>{{ $t('common.moveTo') }}</div>
      <div class="ctx-item" @click="handleDocAction('pin')"><el-icon><Flag /></el-icon>{{ $t('knowledge.detail.addToTop') }}</div>
      <div class="ctx-item" @click="handleDocAction('favorite')"><el-icon><Star /></el-icon>{{ $t('knowledge.detail.favorite') }}</div>
      <div class="ctx-item" @click="handleDocAction('versions')"><el-icon><Clock /></el-icon>{{ $t('knowledge.detail.versionHistory') }}</div>
      <div class="ctx-sep" />
      <div class="ctx-item" @click="handleDocAction('rename')"><el-icon><EditPen /></el-icon>{{ $t('common.rename') }}</div>
      <div class="ctx-item danger" @click="handleDocAction('delete')"><el-icon><Delete /></el-icon>{{ $t('common.delete') }}</div>
    </div>

    <!-- Tree Node Context Menu -->
    <div v-if="treeCtxMenu.visible" class="context-menu" :style="{ left: treeCtxMenu.x + 'px', top: treeCtxMenu.y + 'px' }">
      <div class="ctx-item" @click="handleTreeAction('open')"><el-icon><View /></el-icon>{{ $t('knowledge.detail.openDoc') }}</div>
      <div class="ctx-item" @click="handleTreeAction('rename')"><el-icon><EditPen /></el-icon>{{ $t('common.rename') }}</div>
      <div class="ctx-item" @click="handleTreeAction('move')"><el-icon><Rank /></el-icon>{{ $t('common.moveTo') }}</div>
      <div class="ctx-sep" />
      <div class="ctx-item danger" @click="handleTreeAction('delete')"><el-icon><Delete /></el-icon>{{ $t('common.delete') }}</div>
    </div>

    <!-- Members Drawer -->
    <el-drawer v-model="showMembers" :title="$t('knowledge.detail.memberManagement')" direction="rtl" size="420px">
      <div class="member-toolbar">
        <el-input v-model="memberSearch" :placeholder="$t('knowledge.detail.searchMember')" prefix-icon="Search" clearable size="small" style="flex:1" />
        <el-button type="primary" size="small" @click="showAddMember = true"><el-icon><Plus /></el-icon>{{ $t('knowledge.detail.add') }}</el-button>
      </div>
      <div class="member-list">
        <div v-for="m in filteredMembers" :key="m.userId" class="member-item">
          <el-avatar :size="32" :src="m.userAvatar">{{ m.userName?.[0] }}</el-avatar>
          <div class="member-info">
            <div class="member-name">{{ m.userName }}</div>
            <div class="member-role-text">{{ roleLabel(m.role) }}</div>
          </div>
          <el-dropdown trigger="click" @command="(cmd: string) => handleMemberRole(m, cmd)">
            <el-tag :type="roleTagType(m.role)" size="small" class="role-tag">{{ roleLabel(m.role) }}<el-icon class="el-icon--right"><ArrowDown /></el-icon></el-tag>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="admin">{{ $t('knowledge.detail.admin') }}</el-dropdown-item>
                <el-dropdown-item command="editor">{{ $t('knowledge.detail.canEdit') }}</el-dropdown-item>
                <el-dropdown-item command="viewer">{{ $t('knowledge.detail.viewOnly') }}</el-dropdown-item>
                <el-dropdown-item command="remove" divided><span style="color:var(--kx-danger)">{{ $t('knowledge.detail.remove') }}</span></el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>

      <!-- Add Member Sub Dialog -->
      <el-dialog v-model="showAddMember" :title="$t('knowledge.detail.addMember')" width="400px" append-to-body destroy-on-close>
        <el-form label-position="top">
          <el-form-item :label="$t('knowledge.detail.user')">
            <el-input v-model="addMemberForm.keyword" :placeholder="$t('knowledge.detail.searchUserOrEmail')" />
          </el-form-item>
          <el-form-item :label="$t('knowledge.detail.role')">
            <el-radio-group v-model="addMemberForm.role">
              <el-radio value="admin">{{ $t('knowledge.detail.admin') }}</el-radio>
              <el-radio value="editor">{{ $t('knowledge.detail.canEdit') }}</el-radio>
              <el-radio value="viewer">{{ $t('knowledge.detail.viewOnly') }}</el-radio>
            </el-radio-group>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="showAddMember = false">{{ $t('common.cancel') }}</el-button>
          <el-button type="primary" @click="handleAddMember">{{ $t('knowledge.detail.add') }}</el-button>
        </template>
      </el-dialog>
    </el-drawer>

    <!-- Settings Drawer -->
    <el-drawer v-model="showSettings" :title="$t('knowledge.detail.kbSettings')" direction="rtl" size="480px">
      <el-tabs v-model="settingsTab">
        <el-tab-pane :label="$t('knowledge.detail.basicInfo')" name="basic">
          <el-form label-position="top" style="max-width:400px">
            <el-form-item :label="$t('common.name')">
              <el-input v-model="settingsForm.name" maxlength="50" show-word-limit />
            </el-form-item>
            <el-form-item :label="$t('common.description')">
              <el-input v-model="settingsForm.description" type="textarea" :rows="3" maxlength="200" show-word-limit />
            </el-form-item>
            <el-form-item :label="$t('knowledge.detail.visibility')">
              <el-radio-group v-model="settingsForm.visibility">
                <el-radio value="private">{{ $t('knowledge.detail.memberOnly') }}</el-radio>
                <el-radio value="team">{{ $t('knowledge.detail.teamVisible') }}</el-radio>
                <el-radio value="public">{{ $t('knowledge.detail.publicVisible') }}</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="savingSettings" @click="handleSaveSettings">{{ $t('common.save') }}</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane :label="$t('knowledge.detail.categoryManagement')" name="categories">
          <div class="cat-toolbar">
            <el-button size="small" type="primary" @click="showNewCatDialog = true"><el-icon><Plus /></el-icon>{{ $t('knowledge.detail.newCategory') }}</el-button>
          </div>
          <div class="cat-list">
            <div v-for="cat in categories" :key="cat.id" class="cat-item">
              <el-icon color="#f5a623"><Folder /></el-icon>
              <span class="cat-name">{{ cat.name }}</span>
              <span class="cat-count">{{ cat.docCount || 0 }} {{ $t('knowledge.detail.docCountUnit') }}</span>
              <el-icon class="cat-action" @click="handleDeleteCategory(cat)"><Delete /></el-icon>
            </div>
            <div v-if="!categories.length" class="cat-empty">{{ $t('knowledge.detail.noCategoryHint') }}</div>
          </div>
        </el-tab-pane>
        <el-tab-pane :label="$t('knowledge.detail.dataSources')" name="sources">
          <div class="sources-tip">{{ $t('knowledge.detail.sourceTip') }}</div>
          <div class="sources-toolbar">
            <el-button size="small" type="primary" @click="showAddSourceDialog = true"><el-icon><Link /></el-icon>{{ $t('knowledge.detail.addSource') }}</el-button>
            <el-button size="small" :loading="syncingAll" @click="syncAllSources"><el-icon><Refresh /></el-icon>{{ $t('knowledge.detail.syncAll') }}</el-button>
          </div>
          <div v-loading="sourcesLoading" class="sources-list">
            <div v-for="src in kbSources" :key="src.id" class="source-item-row">
              <el-icon :color="src.sourceType === 'folder' ? '#f5a623' : '#3370ff'">
                <component :is="src.sourceType === 'folder' ? 'Folder' : 'Document'" />
              </el-icon>
              <div class="source-info">
                <div class="source-name">{{ src.sourceName }}</div>
                <div class="source-meta">{{ src.sourceType === 'folder' ? $t('knowledge.detail.folder') : $t('knowledge.detail.document') }} · {{ src.lastSyncAt ? $t('knowledge.detail.lastSync', { time: formatDate(src.lastSyncAt) }) : $t('knowledge.detail.notSynced') }}</div>
              </div>
              <div class="source-actions">
                <el-button size="small" text @click="syncSource(src.id)"><el-icon><Refresh /></el-icon>{{ $t('knowledge.detail.sync') }}</el-button>
                <el-button size="small" text type="danger" @click="removeSource(src.id)"><el-icon><Delete /></el-icon></el-button>
              </div>
            </div>
            <div v-if="!sourcesLoading && !kbSources.length" class="sources-empty">
              <el-empty :description="$t('knowledge.detail.noSourceHint')" />
            </div>
          </div>
        </el-tab-pane>
        <el-tab-pane :label="$t('knowledge.detail.approvalSettings')" name="approval">
          <el-form label-position="top" style="max-width:400px">
            <el-form-item :label="$t('knowledge.detail.publishApproval')">
              <el-switch v-model="settingsForm.requireApproval" :active-text="$t('knowledge.detail.enabled')" :inactive-text="$t('knowledge.detail.disabled')" />
              <div class="form-hint">{{ $t('knowledge.detail.approvalHint') }}</div>
            </el-form-item>
            <el-form-item :label="$t('knowledge.detail.defaultReviewer')" v-if="settingsForm.requireApproval">
              <el-input v-model="settingsForm.defaultReviewer" :placeholder="$t('knowledge.detail.reviewerPlaceholder')" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="savingSettings" @click="handleSaveSettings">{{ $t('common.save') }}</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane :label="$t('knowledge.detail.vectorization')" name="vectorization">
          <div class="vector-section">
            <h4 class="vs-title">{{ $t('knowledge.detail.embeddingStatus') }}</h4>
            <div v-loading="loadingEmbedding" class="embedding-status">
              <div v-if="embeddingStatus" class="status-grid">
                <div class="status-item">
                  <span class="status-label">{{ $t('knowledge.detail.totalChunks') }}</span>
                  <span class="status-value">{{ embeddingStatus.totalChunks }}</span>
                </div>
                <div class="status-item">
                  <span class="status-label">{{ $t('knowledge.detail.embeddedChunks') }}</span>
                  <span class="status-value success">{{ embeddingStatus.embeddedChunks }}</span>
                </div>
                <div class="status-item">
                  <span class="status-label">{{ $t('knowledge.detail.pendingChunks') }}</span>
                  <span class="status-value warning">{{ embeddingStatus.pendingChunks }}</span>
                </div>
                <div class="status-item">
                  <span class="status-label">{{ $t('knowledge.detail.failedChunks') }}</span>
                  <span class="status-value danger">{{ embeddingStatus.failedChunks }}</span>
                </div>
              </div>
              <div v-else class="status-empty">{{ $t('common.noData') }}</div>
              <div class="status-actions">
                <el-button size="small" :loading="rebuildingEmbeddings" @click="handleRebuildEmbeddings">
                  <el-icon><Refresh /></el-icon>{{ $t('knowledge.detail.rebuildIndex') }}
                </el-button>
              </div>
            </div>

            <h4 class="vs-title" style="margin-top:24px">{{ $t('knowledge.detail.raptorTitle') }}</h4>
            <div class="raptor-status">
              <div v-if="raptorStats" class="status-grid">
                <div class="status-item">
                  <span class="status-label">{{ $t('knowledge.detail.totalNodes') }}</span>
                  <span class="status-value">{{ raptorStats.totalNodes }}</span>
                </div>
                <div class="status-item">
                  <span class="status-label">{{ $t('knowledge.detail.leafNodes') }}</span>
                  <span class="status-value">{{ raptorStats.leafNodes }}</span>
                </div>
                <div class="status-item">
                  <span class="status-label">{{ $t('knowledge.detail.clusterNodes') }}</span>
                  <span class="status-value">{{ raptorStats.clusterNodes }}</span>
                </div>
                <div class="status-item">
                  <span class="status-label">{{ $t('knowledge.detail.maxLevel') }}</span>
                  <span class="status-value">{{ raptorStats.maxLevel }}</span>
                </div>
              </div>
              <div v-else class="status-empty">{{ $t('knowledge.detail.noRaptorTree') }}</div>
              <div class="status-actions">
                <el-button size="small" type="primary" :loading="buildingRaptor" @click="handleBuildRaptorTree">
                  <el-icon><Share /></el-icon>{{ $t('knowledge.detail.buildRaptorTree') }}
                </el-button>
              </div>
            </div>

            <div class="vector-tip">
              <el-icon color="#3370ff"><InfoFilled /></el-icon>
              <span>{{ $t('knowledge.detail.raptorTip') }}</span>
            </div>
          </div>
        </el-tab-pane>
        <el-tab-pane :label="$t('knowledge.detail.advanced')" name="advanced">
          <div class="danger-zone">
            <h4>{{ $t('knowledge.detail.dangerZone') }}</h4>
            <div class="danger-item">
              <div>
                <div class="danger-title">{{ $t('knowledge.detail.transferKb') }}</div>
                <div class="danger-desc">{{ $t('knowledge.detail.transferKbDesc') }}</div>
              </div>
              <el-button size="small" @click="handleTransferKb">{{ $t('knowledge.detail.transfer') }}</el-button>
            </div>
            <div class="danger-item">
              <div>
                <div class="danger-title">{{ $t('knowledge.detail.deleteKb') }}</div>
                <div class="danger-desc">{{ $t('knowledge.detail.deleteKbDesc') }}</div>
              </div>
              <el-button size="small" type="danger" @click="handleDeleteKb">{{ $t('common.delete') }}</el-button>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-drawer>

    <!-- Version History Dialog -->
    <el-dialog v-model="showVersions" :title="$t('knowledge.detail.versionHistory')" width="600px" destroy-on-close>
      <div v-if="versionLoading" v-loading="true" style="min-height:200px" />
      <div v-else>
        <div style="margin-bottom: 12px; display: flex; justify-content: space-between; align-items: center;">
          <el-button size="small" :type="diffCompareMode ? 'warning' : 'primary'" @click="toggleDiffMode">
            {{ diffCompareMode ? $t('knowledge.detail.cancelCompare') : $t('knowledge.detail.versionCompare') }}
          </el-button>
          <el-button v-if="diffCompareMode && diffSelectedVersions.length === 2" size="small" type="success" @click="handleDiffCompare">
            {{ $t('knowledge.detail.compareSelected') }}
          </el-button>
          <span v-if="diffCompareMode" style="font-size:12px;color:#909399;">{{ $t('knowledge.detail.selectTwoVersions') }}</span>
        </div>
        <div v-for="ver in versions" :key="ver.id" class="version-item">
          <div class="ver-left">
            <el-checkbox v-if="diffCompareMode" :model-value="diffSelectedVersions.includes(ver.version)" @change="(val: any) => toggleDiffVersion(ver.version, val)" :disabled="!diffSelectedVersions.includes(ver.version) && diffSelectedVersions.length >= 2" />
            <div class="ver-num">v{{ ver.version }}</div>
            <div class="ver-time">{{ formatDate(ver.createdAt) }}</div>
          </div>
          <div class="ver-editor">{{ ver.editorName }}</div>
          <el-button size="small" text type="primary" @click="handleRollback(ver)">{{ $t('knowledge.detail.rollbackTo') }}</el-button>
        </div>
        <div v-if="!versions.length" class="ver-empty">{{ $t('knowledge.detail.noVersions') }}</div>
      </div>
    </el-dialog>

    <!-- Version Diff Dialog -->
    <el-dialog v-model="showDiffDialog" :title="$t('knowledge.detail.versionDiff')" width="800px" destroy-on-close>
      <div v-if="diffLoading" v-loading="true" style="min-height:200px" />
      <VersionDiff
        v-else-if="diffResult"
        :old-version="diffResult.old_version"
        :new-version="diffResult.new_version"
        :lines="diffResult.lines"
        :stats="diffResult.stats"
      />
    </el-dialog>

    <!-- Rename Dialog -->
    <el-dialog v-model="showRename" :title="$t('common.rename')" width="400px" destroy-on-close>
      <el-input v-model="renameValue" maxlength="50" show-word-limit :placeholder="$t('knowledge.detail.inputNewName')" />
      <template #footer>
        <el-button @click="showRename = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="confirmRename">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- Move Dialog -->
    <el-dialog v-model="showMoveDialog" :title="$t('common.moveTo')" width="400px">
      <p style="margin-bottom:8px;color:var(--kx-text-secondary);font-size:13px">{{ $t('knowledge.detail.selectCategory') }}</p>
      <div class="move-list">
        <div class="move-item" :class="{ active: moveTargetId === null }" @click="moveTargetId = null">
          <el-icon color="#3370ff"><Collection /></el-icon><span>{{ $t('knowledge.detail.kbRootDir') }}</span>
        </div>
        <div v-for="cat in categories" :key="cat.id" class="move-item" :class="{ active: moveTargetId === cat.id }" @click="moveTargetId = cat.id">
          <el-icon color="#f5a623"><Folder /></el-icon><span>{{ cat.name }}</span>
        </div>
      </div>
      <template #footer>
        <el-button @click="showMoveDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="confirmMove">{{ $t('knowledge.detail.confirmMove') }}</el-button>
      </template>
    </el-dialog>

    <!-- New Category Dialog -->
    <el-dialog v-model="showNewCatDialog" :title="$t('knowledge.detail.newCategory')" width="400px" destroy-on-close>
      <el-input v-model="newCatName" :placeholder="$t('knowledge.detail.categoryName')" maxlength="30" show-word-limit />
      <template #footer>
        <el-button @click="showNewCatDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleCreateCategory">{{ $t('common.create') }}</el-button>
      </template>
    </el-dialog>

    <!-- Import Dialog -->
    <el-dialog v-model="showImportDialog" :title="$t('knowledge.detail.importToKb')" width="480px" destroy-on-close>
      <el-upload drag :auto-upload="false" :limit="5" accept=".md,.json,.txt,.html,.docx" :on-change="handleImportFileChange">
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">{{ $t('knowledge.detail.dragOrClick') }} <em>{{ $t('knowledge.detail.clickUpload') }}</em></div>
        <template #tip><div class="el-upload__tip">{{ $t('knowledge.detail.importTip') }}</div></template>
      </el-upload>
      <template #footer>
        <el-button @click="showImportDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :disabled="!importFiles.length" @click="handleImport">{{ $t('common.import') }}</el-button>
      </template>
    </el-dialog>

    <!-- Add Source Dialog -->
    <el-dialog v-model="showAddSourceDialog" :title="$t('knowledge.detail.addSourceTitle')" width="500px" destroy-on-close>
      <div class="add-source-desc">{{ $t('knowledge.detail.addSourceDesc') }}</div>
      <el-form label-position="top" style="margin-top:16px">
        <el-form-item :label="$t('knowledge.detail.sourceType')">
          <el-radio-group v-model="addSourceForm.type">
            <el-radio value="folder">{{ $t('knowledge.detail.folderBatch') }}</el-radio>
            <el-radio value="document">{{ $t('knowledge.detail.singleDoc') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('knowledge.detail.selectDocOrFolder')">
          <el-select v-model="addSourceForm.sourceId" :placeholder="$t('knowledge.detail.pleaseSelect')" filterable style="width:100%">
            <el-option
              v-for="item in addSourceForm.type === 'folder' ? allFolders : allDocs"
              :key="item.id"
              :value="item.id"
              :label="item.title || item.originalName"
            >
              <el-icon :color="item.type === 'folder' ? '#f5a623' : '#3370ff'">
                <component :is="item.type === 'folder' ? 'Folder' : 'Document'" />
              </el-icon>
              <span style="margin-left:8px">{{ item.title || item.originalName }}</span>
            </el-option>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddSourceDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="addingSource" :disabled="!addSourceForm.sourceId" @click="confirmAddSource">{{ $t('knowledge.detail.addAndSync') }}</el-button>
      </template>
    </el-dialog>

    <!-- Knowledge Graph Dialog -->
    <el-dialog v-model="showGraphPanel" :title="$t('knowledge.detail.graphTitle')" fullscreen destroy-on-close>
      <div class="graph-panel">
        <!-- 构建控制区 -->
        <div class="graph-toolbar">
          <div class="graph-status">
            <el-tag v-if="graphStatus?.status === 'ready'" type="success">{{ $t('knowledge.detail.graphReady') }}</el-tag>
            <el-tag v-else-if="graphStatus?.status === 'building'" type="warning">{{ $t('knowledge.detail.graphBuilding') }}</el-tag>
            <el-tag v-else-if="graphStatus?.status === 'failed'" type="danger">{{ $t('knowledge.detail.graphFailed') }}</el-tag>
            <el-tag v-else type="info">{{ $t('knowledge.detail.graphIdle') }}</el-tag>
            <span v-if="graphStatus?.status === 'ready'" class="graph-stats">
              {{ $t('knowledge.detail.graphStats', { nodes: graphStatus.nodeCount, edges: graphStatus.edgeCount, communities: graphStatus.communityCount }) }}
            </span>
          </div>
          <div class="graph-actions">
            <el-button type="primary" @click="handleBuildGraph"
                       :loading="graphStatus?.status === 'building'"
                       :disabled="graphStatus?.status === 'building'">
              {{ graphStatus?.status === 'ready' ? $t('knowledge.detail.rebuildGraph') : $t('knowledge.detail.buildGraph') }}
            </el-button>
          </div>
        </div>

        <!-- 构建中状态 -->
        <div v-if="graphStatus?.status === 'building'" class="graph-building">
          <el-icon class="is-loading" :size="48"><Loading /></el-icon>
          <p>{{ $t('knowledge.detail.graphBuildingMsg') }}</p>
          <p class="graph-building-tip">{{ $t('knowledge.detail.graphBuildingTip') }}</p>
        </div>

        <!-- 失败状态 -->
        <div v-else-if="graphStatus?.status === 'failed'" class="graph-error">
          <el-result icon="error" :title="$t('knowledge.detail.graphBuildFailed')" :sub-title="graphStatus?.errorMessage">
            <template #extra>
              <el-button type="primary" @click="handleBuildGraph">{{ $t('knowledge.detail.rebuildGraph') }}</el-button>
            </template>
          </el-result>
        </div>

        <!-- 未构建状态 -->
        <div v-else-if="!graphStatus || graphStatus?.status === 'idle'" class="graph-empty">
          <el-empty :description="$t('knowledge.detail.graphNotBuilt')">
            <el-button type="primary" @click="handleBuildGraph">{{ $t('knowledge.detail.buildKnowledgeGraph') }}</el-button>
          </el-empty>
          <p class="graph-empty-tip">{{ $t('knowledge.detail.graphEmptyTip') }}</p>
        </div>

        <!-- 图谱展示 -->
        <KnowledgeGraphView v-else :graph-data="graphData" :loading="graphLoading"
                            style="height: calc(100vh - 130px)" />
      </div>
    </el-dialog>

    <!-- AI Chat Button -->
    <div class="chat-fab" @click="showChatPanel = true" v-if="!showChatPanel">
      <el-icon :size="24"><ChatDotRound /></el-icon>
    </div>

    <!-- AI Chat Panel -->
    <KbChatPanel
      :is-open="showChatPanel"
      :kb-id="kbId"
      :kb-name="kb?.name || ''"
      @close="showChatPanel = false"
      @open-doc="handleOpenDocFromChat"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { getKnowledgeBaseDetail, getKnowledgeBaseTree, getKnowledgeMembers, deleteKnowledgeBase, updateKnowledgeBase, removeKnowledgeMember, publishDocument, searchKnowledge, getKnowledgeSources, addKnowledgeSource, removeKnowledgeSource, syncKnowledgeSource, getEmbeddingStatus, buildRaptorTree, getRaptorTreeStats, rebuildEmbeddings, buildKnowledgeGraph, getKnowledgeGraph, getKnowledgeGraphStatus, transferKnowledgeBase } from '@/api/modules/knowledge'
import { createDocument, updateDocument, deleteDocument, copyDocument, moveDocument, pinDocument, favoriteDocument, getDocumentVersions, rollbackVersion, importDocument, getDocumentTree, getVersionDiff } from '@/api/modules/document'
import type { KnowledgeBase, DocumentVersion } from '@/types'
import type { DiffLine, VersionDiffResult } from '@/api/modules/document'
import { ElMessage, ElMessageBox, ElLoading } from 'element-plus'
import type { UploadFile } from 'element-plus'
import { WarningFilled } from '@element-plus/icons-vue'
import KbChatPanel from '@/components/knowledge/KbChatPanel.vue'
import KnowledgeGraphView from '@/components/knowledge/KnowledgeGraphView.vue'
import VersionDiff from '@/components/document/VersionDiff.vue'

interface KbDoc {
  id: number
  title: string
  type: string
  parentId: number | null
  ownerName: string
  status: 'draft' | 'review' | 'published'
  isPinned: boolean
  isFavorite: boolean
  isExpanded: boolean
  createdAt: string
  updatedAt: string
  children?: KbDoc[]
  [key: string]: any
}

interface Category {
  id: number
  name: string
  docCount: number
}

interface Member {
  userId: number
  userName: string
  userAvatar: string
  role: string
}

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const kbId = computed(() => Number(route.params.id))

// Core state
const kb = ref<KnowledgeBase | null>(null)
const loading = ref(false)
const loadError = ref<string | null>(null)
const documents = ref<KbDoc[]>([])
const members = ref<Member[]>([])
const categories = ref<Category[]>([])
const versions = ref<DocumentVersion[]>([])
const versionLoading = ref(false)

// Version diff state
const diffCompareMode = ref(false)
const diffSelectedVersions = ref<number[]>([])
const showDiffDialog = ref(false)
const diffLoading = ref(false)
const diffResult = ref<VersionDiffResult | null>(null)
const versionDocId = ref<number | null>(null)

// UI state
const sidebarCollapsed = ref(false)
const treeSearch = ref('')
const showTreeNewMenu = ref(false)
const selectedDocId = ref<number | null>(null)
const searchKey = ref('')
const contentTab = ref('docs')
const viewMode = ref<'list' | 'grid'>('list')
const sortBy = ref('updated')
const selectedDocs = ref<KbDoc[]>([])

// Dialogs
const showMembers = ref(false)
const showSettings = ref(false)
const showVersions = ref(false)
const showRename = ref(false)
const showMoveDialog = ref(false)
const showNewCatDialog = ref(false)
const showImportDialog = ref(false)
const showAddMember = ref(false)
const showChatPanel = ref(false)

// Share dialog
const shareLink = ref('')
const shareDialogVisible = ref(false)

// Knowledge source state
const kbSources = ref<any[]>([])
const sourcesLoading = ref(false)
const syncingAll = ref(false)
const showAddSourceDialog = ref(false)
const addingSource = ref(false)
const addSourceForm = reactive({ type: 'folder', sourceId: undefined as number | undefined })
const allFolders = ref<any[]>([])
const allDocs = ref<any[]>([])

// Form state
const settingsTab = ref('basic')
const settingsForm = reactive({ name: '', description: '', visibility: 'private', requireApproval: false, defaultReviewer: '' })
const savingSettings = ref(false)
const memberSearch = ref('')
const addMemberForm = reactive({ keyword: '', role: 'editor' })
const renameValue = ref('')
const renameTarget = ref<KbDoc | null>(null)
const moveTargetId = ref<number | null>(null)
const moveTargetDoc = ref<KbDoc | null>(null)
const newCatName = ref('')
const importFiles = ref<File[]>([])

// 知识图谱
const showGraphPanel = ref(false)
const graphData = ref<any>(null)
const graphLoading = ref(false)
const graphStatus = ref<any>(null)
const graphPollingTimer = ref<any>(null)

// Vectorization state
const embeddingStatus = ref<{ totalChunks: number; embeddedChunks: number; pendingChunks: number; failedChunks: number; progress: number } | null>(null)
const raptorStats = ref<{ totalNodes: number; leafNodes: number; clusterNodes: number; rootNodes: number; maxLevel: number } | null>(null)
const loadingEmbedding = ref(false)
const buildingRaptor = ref(false)
const rebuildingEmbeddings = ref(false)

// Context menus
const docCtxMenu = reactive({ visible: false, x: 0, y: 0, doc: null as KbDoc | null })
const treeCtxMenu = reactive({ visible: false, x: 0, y: 0, node: null as KbDoc | null })

// Type helpers
const typeMap: Record<string, { icon: string; color: string }> = {
  folder: { icon: 'Folder', color: '#f5a623' },
  doc: { icon: 'Document', color: '#3370ff' },
  sheet: { icon: 'Grid', color: '#36b37e' },
  slide: { icon: 'Monitor', color: '#ff7d00' },
  mindnote: { icon: 'Share', color: '#9254de' },
  bitable: { icon: 'Tickets', color: '#00b8d9' },
  survey: { icon: 'Notebook', color: '#f54a45' },
}
function getTypeIcon(type: string) { return typeMap[type]?.icon || 'Document' }
function getTypeColor(type: string) { return typeMap[type]?.color || '#3370ff' }

function statusLabel(s?: string) {
  if (s === 'published') return t('knowledge.detail.statusPublished')
  if (s === 'review') return t('knowledge.detail.statusReview')
  return t('knowledge.detail.statusDraft')
}
function statusTagType(s?: string): '' | 'success' | 'warning' | 'info' {
  if (s === 'published') return 'success'
  if (s === 'review') return 'warning'
  return 'info'
}
function roleLabel(r: string) {
  if (r === 'admin' || r === 'owner') return t('knowledge.detail.admin')
  if (r === 'editor') return t('knowledge.detail.canEdit')
  return t('knowledge.detail.viewOnly')
}
function roleTagType(r: string): '' | 'success' | 'warning' | 'info' {
  if (r === 'admin' || r === 'owner') return ''
  if (r === 'editor') return 'success'
  return 'info'
}

function formatDate(t: string) {
  if (!t) return ''
  const d = new Date(t)
  return `${d.getFullYear()}/${d.getMonth()+1}/${d.getDate()}`
}

// Computed
const filteredTreeData = computed(() => {
  const q = treeSearch.value.toLowerCase()
  if (!q) return documents.value
  return documents.value.filter(d => d.title.toLowerCase().includes(q) || d.children?.some(c => c.title.toLowerCase().includes(q)))
})

const displayedDocs = computed(() => {
  let list = documents.value.flatMap(d => d.type === 'folder' ? (d.children || []) : [d])
  // Also include top-level non-folder docs
  list = [...documents.value.filter(d => d.type !== 'folder'), ...list]
  // Deduplicate
  const seen = new Set<number>()
  list = list.filter(d => { if (seen.has(d.id)) return false; seen.add(d.id); return true })

  // Tab filter
  if (contentTab.value === 'published') list = list.filter(d => d.status === 'published')
  else if (contentTab.value === 'draft') list = list.filter(d => d.status === 'draft')
  else if (contentTab.value === 'review') list = list.filter(d => d.status === 'review')

  // Sort
  list.sort((a, b) => {
    if (sortBy.value === 'title') return a.title.localeCompare(b.title)
    if (sortBy.value === 'created') return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
    if (sortBy.value === 'status') return (a.status || '').localeCompare(b.status || '')
    return new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
  })
  return list
})

const filteredMembers = computed(() => {
  const q = memberSearch.value.toLowerCase()
  if (!q) return members.value
  return members.value.filter(m => m.userName?.toLowerCase().includes(q))
})

// Data fetching
async function fetchDetail() {
  try {
    const res: any = await getKnowledgeBaseDetail(kbId.value)
    kb.value = res.data
    if (kb.value) {
      settingsForm.name = kb.value.name
      settingsForm.description = kb.value.description
    }
    loadError.value = null
  } catch (e: any) {
    loadError.value = e?.response?.data?.message || t('knowledge.detail.kbLoadFailed')
    console.error('[KnowledgeDetail] fetchDetail failed:', e)
  }
}

function retryLoad() {
  loadError.value = null
  initKbData()
}

async function fetchDocs() {
  loading.value = true
  try {
    const res: any = await getKnowledgeBaseTree(kbId.value)
    const rawDocs = res.data || []
    documents.value = rawDocs.map((d: any) => ({
      ...d,
      status: d.status || 'draft',
      isExpanded: false,
      children: d.children?.map((c: any) => ({ ...c, status: c.status || 'draft', isExpanded: false })) || [],
    }))
    // Extract categories (folders)
    categories.value = rawDocs
      .filter((d: any) => d.type === 'folder')
      .map((d: any) => ({ id: d.id, name: d.title, docCount: d.children?.length || 0 }))
  } finally { loading.value = false }
}

async function fetchMembers() {
  try {
    const res: any = await getKnowledgeMembers(kbId.value)
    members.value = res.data || []
  } catch (e: any) {
    console.error('[KnowledgeDetail] fetchMembers failed:', e)
  }
}

async function doSearch() {
  if (!searchKey.value.trim()) { fetchDocs(); return }
  loading.value = true
  try {
    const res: any = await searchKnowledge({ keyword: searchKey.value, knowledgeBaseId: kbId.value, page: 1, pageSize: 50 })
    const list = res.data?.list || []
    documents.value = list.map((d: any) => ({ ...d, status: d.status || 'draft', isExpanded: false, children: [] }))
  } finally { loading.value = false }
}

// Actions
function handleTreeNodeClick(node: KbDoc) {
  if (node.type === 'folder') {
    node.isExpanded = !node.isExpanded
  } else {
    selectedDocId.value = node.id
    // Open in new browser tab
    window.open(`/doc/${node.id}`, '_blank')
  }
}

function handleDocClick(doc: any) {
  // Open in new browser tab
  window.open(`/doc/${doc.id}`, '_blank')
}

async function handleNewInKb(type: string) {
  showTreeNewMenu.value = false
  try {
    const res: any = await createDocument({ title: t('document.untitled'), type: type as any, parentId: null })
    if (res.data?.id) {
      // Auto-add to knowledge base as source
      try {
        await addKnowledgeSource(kbId.value, { sourceType: 'document', sourceId: res.data.id })
      } catch (e: any) {
        console.error('[KnowledgeDetail] addKnowledgeSource after create failed:', e)
      }
      ElMessage.success(t('knowledge.detail.createdAndAdded'))
      fetchDocs()
      loadSources()
      // Open in new browser tab
      window.open(`/doc/${res.data.id}`, '_blank')
    }
  } catch { ElMessage.error(t('knowledge.detail.createFailed')) }
}

async function handleNewCategory() {
  showTreeNewMenu.value = false
  showNewCatDialog.value = true
}

async function handleCreateCategory() {
  if (!newCatName.value.trim()) { ElMessage.warning(t('knowledge.detail.categoryNameRequired')); return }
  try {
    await createDocument({ title: newCatName.value.trim(), type: 'folder', parentId: null })
    ElMessage.success(t('knowledge.detail.categoryCreated'))
    showNewCatDialog.value = false
    newCatName.value = ''
    fetchDocs()
  } catch { ElMessage.error(t('knowledge.detail.createFailed')) }
}

async function handleDeleteCategory(cat: Category) {
  await ElMessageBox.confirm(t('knowledge.detail.deleteCategoryConfirm', { name: cat.name }), t('knowledge.detail.deleteConfirmTitle'))
  try {
    await deleteDocument(cat.id)
    ElMessage.success(t('knowledge.detail.removed'))
    fetchDocs()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || t('knowledge.detail.deleteCategoryFailed'))
  }
}

function handleSort(cmd: string) {
  sortBy.value = cmd
}

function handleSelectionChange(rows: KbDoc[]) {
  selectedDocs.value = rows
}

// Document context menu
function showDocCtxMenu(e: MouseEvent, doc: KbDoc) {
  e.preventDefault()
  e.stopPropagation()
  treeCtxMenu.visible = false
  docCtxMenu.visible = true
  docCtxMenu.x = e.clientX
  docCtxMenu.y = e.clientY
  docCtxMenu.doc = doc
}

function showTreeCtxMenu(e: MouseEvent, node: KbDoc) {
  e.preventDefault()
  e.stopPropagation()
  docCtxMenu.visible = false
  treeCtxMenu.visible = true
  treeCtxMenu.x = e.clientX
  treeCtxMenu.y = e.clientY
  treeCtxMenu.node = node
}

function handleOpenDocFromChat(docId: number) {
  window.open(`/doc/${docId}`, '_blank')
}

async function handleDocAction(action: string) {
  const doc = docCtxMenu.doc
  docCtxMenu.visible = false
  if (!doc) return
  switch (action) {
    case 'open':
      // Open in new browser tab
      window.open(`/doc/${doc.id}`, '_blank')
      break
    case 'openNew':
      window.open(`${window.location.origin}/doc/${doc.id}`, '_blank')
      break
    case 'publish':
      try {
        await publishDocument(kbId.value, doc.id)
        ElMessage.success(t('knowledge.detail.publishSuccess'))
        fetchDocs()
      } catch { ElMessage.error(t('knowledge.detail.publishFailed')) }
      break
    case 'unpublish':
      doc.status = 'draft'
      ElMessage.success(t('knowledge.detail.unpublished'))
      fetchDocs()
      break
    case 'submitReview':
      doc.status = 'review'
      ElMessage.success(t('knowledge.detail.reviewSubmitted'))
      fetchDocs()
      break
    case 'share':
      shareLink.value = `${window.location.origin}/doc/${doc.id}`
      shareDialogVisible.value = true
      break
    case 'copyLink': {
      const link = `${window.location.origin}/doc/${doc.id}`
      navigator.clipboard.writeText(link)
      ElMessage.success(t('common.linkCopied'))
      break
    }
    case 'copy':
      try {
        await copyDocument(doc.id, false)
        ElMessage.success(t('knowledge.detail.copyCreated'))
        fetchDocs()
      } catch { ElMessage.error(t('knowledge.detail.copyFailed')) }
      break
    case 'move':
      moveTargetDoc.value = doc
      moveTargetId.value = null
      showMoveDialog.value = true
      break
    case 'pin':
      await pinDocument(doc.id, !doc.isPinned)
      ElMessage.success(doc.isPinned ? t('knowledge.detail.unpinned') : t('knowledge.detail.pinned'))
      fetchDocs()
      break
    case 'favorite':
      await favoriteDocument(doc.id, !doc.isFavorite)
      ElMessage.success(doc.isFavorite ? t('knowledge.detail.unfavorited') : t('knowledge.detail.favorited'))
      fetchDocs()
      break
    case 'versions':
      showVersions.value = true
      versionLoading.value = true
      versionDocId.value = doc.id
      diffCompareMode.value = false
      diffSelectedVersions.value = []
      try {
        const res: any = await getDocumentVersions(doc.id)
        versions.value = res.data || []
      } finally { versionLoading.value = false }
      break
    case 'rename':
      renameTarget.value = doc
      renameValue.value = doc.title
      showRename.value = true
      break
    case 'delete':
      await ElMessageBox.confirm(t('knowledge.detail.deleteDocConfirm', { title: doc.title }), t('knowledge.detail.deleteConfirmTitle'))
      await deleteDocument(doc.id)
      ElMessage.success(t('knowledge.detail.removed'))
      fetchDocs()
      break
  }
}

async function handleTreeAction(action: string) {
  const node = treeCtxMenu.node
  treeCtxMenu.visible = false
  if (!node) return
  switch (action) {
    case 'open':
      if (node.type === 'folder') { node.isExpanded = true }
      else window.open(`/doc/${node.id}`, '_blank')
      break
    case 'rename':
      renameTarget.value = node
      renameValue.value = node.title
      showRename.value = true
      break
    case 'move':
      moveTargetDoc.value = node
      moveTargetId.value = null
      showMoveDialog.value = true
      break
    case 'delete':
      await ElMessageBox.confirm(t('knowledge.detail.deleteDocConfirm', { title: node.title }), t('knowledge.detail.deleteConfirmTitle'))
      await deleteDocument(node.id)
      ElMessage.success(t('knowledge.detail.removed'))
      fetchDocs()
      break
  }
}

async function confirmRename() {
  if (!renameTarget.value || !renameValue.value.trim()) { ElMessage.warning(t('knowledge.detail.nameRequired')); return }
  try {
    await updateDocument(renameTarget.value.id, { title: renameValue.value.trim() })
    ElMessage.success(t('knowledge.detail.renameSuccess'))
    showRename.value = false
    fetchDocs()
  } catch { ElMessage.error(t('knowledge.detail.renameFailed')) }
}

async function confirmMove() {
  if (!moveTargetDoc.value) return
  try {
    await moveDocument(moveTargetDoc.value.id, moveTargetId.value)
    ElMessage.success(t('knowledge.detail.moved'))
    showMoveDialog.value = false
    fetchDocs()
  } catch { ElMessage.error(t('knowledge.detail.moveFailed')) }
}

async function handleRollback(ver: DocumentVersion) {
  await ElMessageBox.confirm(t('knowledge.detail.rollbackConfirm', { version: ver.version }), t('knowledge.detail.rollbackTitle'))
  try {
    await rollbackVersion(ver.documentId, ver.version)
    ElMessage.success(t('knowledge.detail.rolledBack'))
    showVersions.value = false
  } catch { ElMessage.error(t('knowledge.detail.rollbackFailed')) }
}

function toggleDiffMode() {
  diffCompareMode.value = !diffCompareMode.value
  diffSelectedVersions.value = []
}

function toggleDiffVersion(version: number, checked: any) {
  if (checked) {
    if (diffSelectedVersions.value.length < 2) {
      diffSelectedVersions.value.push(version)
    }
  } else {
    diffSelectedVersions.value = diffSelectedVersions.value.filter(v => v !== version)
  }
}

async function handleDiffCompare() {
  if (diffSelectedVersions.value.length !== 2) return
  const sorted = [...diffSelectedVersions.value].sort((a, b) => a - b)
  const docId = versionDocId.value
  if (!docId) return
  diffLoading.value = true
  showDiffDialog.value = true
  diffResult.value = null
  try {
    const res: any = await getVersionDiff(docId, String(sorted[0]), String(sorted[1]))
    diffResult.value = res.data || null
  } catch {
    ElMessage.error(t('knowledge.detail.diffFailed'))
    showDiffDialog.value = false
  } finally {
    diffLoading.value = false
  }
}

// Batch operations
async function batchPublish() {
  const total = selectedDocs.value.length
  let completed = 0
  let failed = 0

  const loadingInstance = ElLoading.service({ text: t('knowledge.detail.publishing', { current: 0, total }) })

  for (const doc of selectedDocs.value) {
    try {
      await publishDocument(kbId.value, doc.id)
      completed++
    } catch {
      failed++
    }
    loadingInstance.setText(t('knowledge.detail.publishing', { current: completed + failed, total }))
  }

  loadingInstance.close()

  if (failed > 0) {
    ElMessage.warning(t('knowledge.detail.publishDone', { success: completed, fail: failed }))
  } else {
    ElMessage.success(t('knowledge.detail.publishAllSuccess', { count: completed }))
  }

  selectedDocs.value = []
  fetchDocs()
}

async function batchMove() {
  if (selectedDocs.value.length === 1) {
    moveTargetDoc.value = selectedDocs.value[0]
  }
  moveTargetId.value = null
  showMoveDialog.value = true
}

async function batchDelete() {
  await ElMessageBox.confirm(t('knowledge.detail.batchDeleteConfirm', { count: selectedDocs.value.length }), t('knowledge.detail.batchDeleteTitle'))
  const total = selectedDocs.value.length
  let completed = 0
  let failed = 0

  const loadingInstance = ElLoading.service({ text: t('knowledge.detail.deleting', { current: 0, total }) })

  for (const doc of selectedDocs.value) {
    try {
      await deleteDocument(doc.id)
      completed++
    } catch {
      failed++
    }
    loadingInstance.setText(t('knowledge.detail.deleting', { current: completed + failed, total }))
  }

  loadingInstance.close()

  if (failed > 0) {
    ElMessage.warning(t('knowledge.detail.deleteDone', { success: completed, fail: failed }))
  } else {
    ElMessage.success(t('knowledge.detail.deleteSuccess', { count: completed }))
  }

  selectedDocs.value = []
  fetchDocs()
}

// Settings
async function handleSaveSettings() {
  savingSettings.value = true
  try {
    await updateKnowledgeBase(kbId.value, { name: settingsForm.name, description: settingsForm.description })
    ElMessage.success(t('knowledge.detail.settingsSaved'))
    fetchDetail()
  } finally { savingSettings.value = false }
}

async function handleDeleteKb() {
  await ElMessageBox.confirm(t('knowledge.detail.deleteKbConfirm'), t('knowledge.detail.deleteConfirmTitle'), { type: 'warning' })
  await deleteKnowledgeBase(kbId.value)
  ElMessage.success(t('knowledge.detail.removed'))
  router.push('/knowledge')
}

async function handleTransferKb() {
  const { value } = await ElMessageBox.prompt(t('knowledge.detail.transferPrompt'), t('knowledge.detail.transferTitle'), {
    confirmButtonText: t('knowledge.detail.transferBtn'),
    cancelButtonText: t('common.cancel'),
    inputPlaceholder: t('knowledge.detail.emailPlaceholder'),
  }).catch(() => ({ value: null }))
  if (value) {
    try {
      await transferKnowledgeBase(kbId.value, value)
      ElMessage.success(t('knowledge.detail.transferSuccess', { user: value }))
      fetchDetail()
    } catch (e: any) {
      ElMessage.error(e?.response?.data?.message || t('knowledge.detail.transferFailed'))
    }
  }
}

// Members
async function handleMemberRole(m: Member, cmd: string) {
  if (cmd === 'remove') {
    await ElMessageBox.confirm(t('knowledge.detail.removeMemberConfirm', { name: m.userName }), t('knowledge.detail.removeConfirmTitle'))
    await removeKnowledgeMember(kbId.value, m.userId)
    ElMessage.success(t('knowledge.detail.removed'))
    fetchMembers()
  } else {
    ElMessage.success(t('knowledge.detail.roleChanged', { name: m.userName, role: roleLabel(cmd) }))
  }
}

async function handleAddMember() {
  if (!addMemberForm.keyword.trim()) { ElMessage.warning(t('knowledge.detail.usernameRequired')); return }
  // Add member to knowledge base
  members.value.push({
    userId: Date.now(),
    userName: addMemberForm.keyword,
    userAvatar: '',
    role: addMemberForm.role,
  })
  ElMessage.success(t('knowledge.detail.memberAdded', { name: addMemberForm.keyword }))
  showAddMember.value = false
  addMemberForm.keyword = ''
  addMemberForm.role = 'view'
}

// Import
function handleImportFileChange(file: UploadFile) {
  if (file.raw) importFiles.value.push(file.raw)
}

async function handleImport() {
  let importedCount = 0
  for (const file of importFiles.value) {
    try {
      const res: any = await importDocument(file, null)
      if (res.data?.id) {
        // Auto-add to knowledge base as source
        try {
          await addKnowledgeSource(kbId.value, { sourceType: 'document', sourceId: res.data.id })
        } catch (e: any) {
          console.error('[KnowledgeDetail] addKnowledgeSource after import failed:', e)
        }
        importedCount++
      }
    } catch (e: any) {
      console.error('[KnowledgeDetail] importDocument failed:', e)
    }
  }
  ElMessage.success(t('knowledge.detail.importedCount', { count: importedCount }))
  showImportDialog.value = false
  importFiles.value = []
  fetchDocs()
  loadSources()
}

// Close menus
function closeMenus() {
  docCtxMenu.visible = false
  treeCtxMenu.visible = false
  showTreeNewMenu.value = false
}

// Knowledge source management
async function loadSources() {
  sourcesLoading.value = true
  try {
    const res: any = await getKnowledgeSources(kbId.value)
    kbSources.value = res.data || []
  } catch (e: any) {
    console.error('[KnowledgeDetail] loadSources failed:', e)
  } finally {
    sourcesLoading.value = false
  }
}

async function loadAllDocsFolders() {
  try {
    const res: any = await getDocumentTree(null)
    const items = res.data || []
    allFolders.value = items.filter((d: any) => d.type === 'folder')
    allDocs.value = items.filter((d: any) => d.type !== 'folder')
  } catch (e: any) {
    console.error('[KnowledgeDetail] loadAllDocsFolders failed:', e)
  }
}

async function syncSource(srcId: number) {
  try {
    await syncKnowledgeSource(kbId.value, srcId)
    ElMessage.success(t('knowledge.detail.syncSuccess'))
    loadSources()
  } catch { ElMessage.error(t('knowledge.detail.syncFailed')) }
}

async function removeSource(srcId: number) {
  await ElMessageBox.confirm(t('knowledge.detail.removeSourceConfirm'), t('knowledge.detail.removeSourceTitle'))
  try {
    await removeKnowledgeSource(kbId.value, srcId)
    ElMessage.success(t('knowledge.detail.removed'))
    loadSources()
  } catch { ElMessage.error(t('knowledge.detail.removeFailed')) }
}

async function syncAllSources() {
  if (!kbSources.value.length) { ElMessage.info(t('knowledge.detail.noSources')); return }
  syncingAll.value = true
  try {
    for (const src of kbSources.value) {
      try { await syncKnowledgeSource(kbId.value, src.id) } catch (e: any) { console.error('[KnowledgeDetail] syncSource failed for id:', src.id, e) }
    }
    ElMessage.success(t('knowledge.detail.syncAllSuccess'))
    loadSources()
  } finally { syncingAll.value = false }
}

async function confirmAddSource() {
  if (!addSourceForm.sourceId) { ElMessage.warning(t('knowledge.detail.selectDocOrFolderRequired')); return }
  addingSource.value = true
  try {
    await addKnowledgeSource(kbId.value, { sourceType: addSourceForm.type, sourceId: addSourceForm.sourceId })
    ElMessage.success(t('knowledge.detail.sourceAdded'))
    showAddSourceDialog.value = false
    addSourceForm.sourceId = undefined
    loadSources()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || t('knowledge.detail.addFailed'))
  } finally { addingSource.value = false }
}

watch(() => addSourceForm.type, () => { addSourceForm.sourceId = undefined })

// Vectorization functions
async function loadEmbeddingStatus() {
  loadingEmbedding.value = true
  try {
    const res: any = await getEmbeddingStatus(kbId.value)
    embeddingStatus.value = res.data
  } catch (e: any) {
    console.error('Failed to load embedding status:', e)
  } finally {
    loadingEmbedding.value = false
  }
}

async function loadRaptorStats() {
  try {
    const res: any = await getRaptorTreeStats(kbId.value)
    raptorStats.value = res.data
  } catch (e: any) {
    console.error('Failed to load raptor stats:', e)
  }
}

async function handleBuildRaptorTree() {
  try {
    await ElMessageBox.confirm(t('knowledge.detail.raptorBuildConfirm'), t('knowledge.detail.buildConfirmTitle'))
  } catch {
    return
  }
  buildingRaptor.value = true
  try {
    await buildRaptorTree(kbId.value, { clusterCount: 10, maxLevel: 3 })
    ElMessage.success(t('knowledge.detail.raptorStarted'))
    // 轮询机制：每2秒查询一次，最多30次（60秒）
    let retryCount = 0
    const maxRetries = 30
    const pollTimer = setInterval(async () => {
      retryCount++
      await loadRaptorStats()
      // 如果已有数据或超过最大重试次数，停止轮询
      if ((raptorStats.value && raptorStats.value.totalNodes > 0) || retryCount >= maxRetries) {
        clearInterval(pollTimer)
        if (retryCount >= maxRetries) {
          ElMessage.info(t('knowledge.detail.buildStillRunning'))
        }
      }
    }, 2000)
  } catch (e: any) {
    const msg = e?.response?.data?.message || e?.message || t('knowledge.detail.buildFailed')
    ElMessage.error(msg)
  } finally {
    buildingRaptor.value = false
  }
}

async function handleRebuildEmbeddings() {
  try {
    await ElMessageBox.confirm(t('knowledge.detail.rebuildConfirm'), t('knowledge.detail.rebuildTitle'))
  } catch {
    return
  }
  rebuildingEmbeddings.value = true
  try {
    await rebuildEmbeddings(kbId.value)
    ElMessage.success(t('knowledge.detail.rebuildStarted'))
    setTimeout(() => {
      loadEmbeddingStatus()
    }, 2000)
  } catch (e: any) {
    const msg = e?.response?.data?.message || e?.message || t('knowledge.detail.rebuildFailed')
    ElMessage.error(msg)
  } finally {
    rebuildingEmbeddings.value = false
  }
}

// 知识图谱功能
async function openGraphPanel() {
  showGraphPanel.value = true
  await loadGraphStatus()
  if (graphStatus.value?.status === 'ready') {
    await loadGraphData()
  }
}

async function loadGraphStatus() {
  try {
    const res = await getKnowledgeGraphStatus(kbId.value) as any
    graphStatus.value = res.data
  } catch (e) {
    graphStatus.value = null
  }
}

async function loadGraphData() {
  graphLoading.value = true
  try {
    const res = await getKnowledgeGraph(kbId.value) as any
    graphData.value = res.data
  } catch (e) {
    graphData.value = null
  } finally {
    graphLoading.value = false
  }
}

async function handleBuildGraph() {
  try {
    await buildKnowledgeGraph(kbId.value)
    ElMessage.success(t('knowledge.detail.graphBuildStarted'))
    graphStatus.value = { status: 'building' }
    startGraphPolling()
  } catch (e) {
    ElMessage.error(t('knowledge.detail.graphBuildStartFailed'))
  }
}

function startGraphPolling() {
  stopGraphPolling()
  graphPollingTimer.value = setInterval(async () => {
    await loadGraphStatus()
    if (graphStatus.value?.status === 'ready') {
      stopGraphPolling()
      await loadGraphData()
      ElMessage.success(t('knowledge.detail.graphBuildComplete'))
    } else if (graphStatus.value?.status === 'failed') {
      stopGraphPolling()
      ElMessage.error(t('knowledge.detail.graphBuildError', { msg: graphStatus.value?.errorMessage || t('common.unknown') }))
    }
  }, 5000)
}

function stopGraphPolling() {
  if (graphPollingTimer.value) {
    clearInterval(graphPollingTimer.value)
    graphPollingTimer.value = null
  }
}

// 初始化：加载所有知识库相关数据
function initKbData() {
  fetchDetail()
  fetchDocs()
  fetchMembers()
  loadSources()
  loadAllDocsFolders()
  loadEmbeddingStatus()
  loadRaptorStats()
}

// 重置所有 UI 状态（路由切换时调用）
function resetState() {
  kb.value = null
  documents.value = []
  members.value = []
  categories.value = []
  versions.value = []
  selectedDocId.value = null
  searchKey.value = ''
  treeSearch.value = ''
  contentTab.value = 'docs'
  selectedDocs.value = []
  kbSources.value = []
  embeddingStatus.value = null
  raptorStats.value = null
  graphData.value = null
  graphStatus.value = null
  showMembers.value = false
  showSettings.value = false
  showGraphPanel.value = false
  showChatPanel.value = false
  stopGraphPolling()
}

// 监听路由参数变化（同组件内 kbId 切换）
watch(kbId, (newId, oldId) => {
  if (newId && newId !== oldId) {
    resetState()
    initKbData()
  }
})

onMounted(() => {
  initKbData()
  document.addEventListener('click', closeMenus)
})
onBeforeUnmount(() => {
  stopGraphPolling()
  document.removeEventListener('click', closeMenus)
})
</script>

<style scoped>
.kb-detail {
  display: flex;
  height: 100%;
  gap: 0;
}

/* Sidebar */
.kb-sidebar {
  width: 260px;
  min-width: 260px;
  border-right: 1px solid var(--kx-border);
  background: var(--kx-sidebar-bg, #f7f8fa);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: width 0.2s, min-width 0.2s;
}
.kb-sidebar.collapsed {
  width: 48px;
  min-width: 48px;
}
.kb-sidebar-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 12px;
  border-bottom: 1px solid var(--kx-border);
}
.kb-sidebar-title {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1;
  min-width: 0;
}
.kb-name-text {
  font-size: 14px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.collapse-btn {
  cursor: pointer;
  font-size: 16px;
  color: var(--kx-text-placeholder);
  flex-shrink: 0;
}
.collapse-btn:hover {
  color: var(--kx-primary);
}
.kb-sidebar-body {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}
.tree-search {
  margin-bottom: 8px;
}
.tree-new-btn {
  position: relative;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 10px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  color: var(--kx-primary);
  margin-bottom: 8px;
}
.tree-new-btn:hover {
  background: rgba(51,112,255,0.06);
}
.tree-new-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  min-width: 180px;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  z-index: 100;
  padding: 4px 0;
}
.tnd-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 9px 14px;
  font-size: 13px;
  cursor: pointer;
  color: var(--kx-text-primary);
}
.tnd-item:hover {
  background: var(--kx-sidebar-bg);
}
.tnd-sep {
  height: 1px;
  background: var(--kx-border);
  margin: 4px 0;
}

/* Tree */
.tree-section {
  margin-top: 4px;
}
.tree-section-label {
  font-size: 11px;
  color: var(--kx-text-placeholder);
  padding: 4px 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.tree-empty {
  font-size: 12px;
  color: var(--kx-text-placeholder);
  padding: 12px 8px;
}
.tree-node-wrap {
  margin-bottom: 1px;
}
.tree-node {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 6px 8px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  color: var(--kx-text-primary);
}
.tree-node:hover {
  background: rgba(0,0,0,0.04);
}
.tree-node.active {
  background: rgba(51,112,255,0.08);
  color: var(--kx-primary);
}
.tree-node.level-2 {
  padding-left: 28px;
}
.tree-expand {
  font-size: 10px;
  color: var(--kx-text-placeholder);
  transition: transform 0.2s;
  flex-shrink: 0;
}
.tree-expand.expanded {
  transform: rotate(90deg);
}
.tree-expand.invisible {
  visibility: hidden;
}
.tree-node-label {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.tree-node-status {
  font-size: 10px;
  padding: 1px 5px;
  border-radius: 3px;
  flex-shrink: 0;
}
.tree-node-status.published {
  background: #e6f7ef;
  color: #36b37e;
}
.tree-node-status.review {
  background: #fff7e0;
  color: #f5a623;
}
.tree-node-status.draft {
  background: #f0f0f0;
  color: #999;
}
.tree-children {
  margin-left: 0;
}

/* Main Content */
.kb-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.kb-topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid var(--kx-border);
  flex-shrink: 0;
}
.kb-topbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}
.kb-topbar-name {
  font-size: 16px;
  font-weight: 600;
}
.kb-topbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* Tabs */
.kb-content-tabs {
  display: flex;
  gap: 24px;
  padding: 0 20px;
  border-bottom: 1px solid var(--kx-border);
  flex-shrink: 0;
}
.kc-tab {
  font-size: 14px;
  color: var(--kx-text-secondary);
  cursor: pointer;
  padding: 12px 0;
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
}
.kc-tab.active {
  color: var(--kx-primary);
  border-color: var(--kx-primary);
  font-weight: 500;
}
.kc-tab:hover {
  color: var(--kx-primary);
}

/* Action Bar */
.kb-action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  flex-shrink: 0;
}
.kb-action-left {
  display: flex;
  gap: 8px;
}
.kb-action-right {
  display: flex;
  gap: 16px;
  align-items: center;
}
.action-link {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: var(--kx-text-secondary);
  cursor: pointer;
}
.action-link:hover {
  color: var(--kx-primary);
}

/* Doc List */
.kb-doc-list {
  flex: 1;
  overflow-y: auto;
  padding: 0 20px 20px;
}
.doc-name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}
.more-btn {
  cursor: pointer;
  font-size: 16px;
  color: var(--kx-text-placeholder);
}
.more-btn:hover {
  color: var(--kx-primary);
}

/* Grid */
.kb-doc-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
  padding-top: 4px;
}
.kb-doc-card {
  padding: 16px;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  background: #fff;
}
.kb-doc-card:hover {
  border-color: var(--kx-primary);
  box-shadow: 0 2px 12px rgba(51,112,255,0.1);
}
.kdc-icon { margin-bottom: 10px; }
.kdc-title {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.kdc-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  color: var(--kx-text-placeholder);
}

.kb-empty { padding: 40px 0; }

/* Error State */
.kb-error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 80px 20px;
  flex: 1;
}
.kb-error-state .error-message {
  font-size: 15px;
  color: var(--kx-text-secondary);
  text-align: center;
  max-width: 400px;
}

/* Batch Bar */
.kb-batch-bar {
  position: sticky;
  bottom: 0;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  background: #fff;
  border-top: 1px solid var(--kx-border);
  box-shadow: 0 -2px 8px rgba(0,0,0,0.06);
}
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.2s;
}
.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(100%);
  opacity: 0;
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

/* Members */
.member-toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}
.member-list {
  max-height: calc(100vh - 200px);
  overflow-y: auto;
}
.member-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 0;
  border-bottom: 1px solid var(--kx-border);
}
.member-info {
  flex: 1;
  min-width: 0;
}
.member-name {
  font-size: 14px;
  font-weight: 500;
}
.member-role-text {
  font-size: 12px;
  color: var(--kx-text-placeholder);
}
.role-tag {
  cursor: pointer;
}

/* Settings */
.form-hint {
  font-size: 12px;
  color: var(--kx-text-placeholder);
  margin-top: 4px;
}
.danger-zone {
  margin-top: 16px;
}
.danger-zone h4 {
  font-size: 15px;
  color: var(--kx-danger);
  margin-bottom: 16px;
}
.danger-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  margin-bottom: 12px;
}
.danger-title {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 4px;
}
.danger-desc {
  font-size: 12px;
  color: var(--kx-text-placeholder);
}

/* Categories */
.cat-toolbar {
  margin-bottom: 12px;
}
.cat-list {
  /* list */
}
.cat-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-bottom: 1px solid var(--kx-border);
}
.cat-name {
  flex: 1;
  font-size: 14px;
}
.cat-count {
  font-size: 12px;
  color: var(--kx-text-placeholder);
}
.cat-action {
  cursor: pointer;
  color: var(--kx-text-placeholder);
  font-size: 14px;
}
.cat-action:hover {
  color: var(--kx-danger);
}
.cat-empty {
  font-size: 13px;
  color: var(--kx-text-placeholder);
  padding: 20px;
  text-align: center;
}

/* Versions */
.version-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 0;
  border-bottom: 1px solid var(--kx-border);
}
.ver-left {
  min-width: 120px;
}
.ver-num {
  font-size: 14px;
  font-weight: 600;
  color: var(--kx-primary);
}
.ver-time {
  font-size: 12px;
  color: var(--kx-text-placeholder);
}
.ver-editor {
  flex: 1;
  font-size: 13px;
}
.ver-empty {
  text-align: center;
  padding: 24px;
  color: var(--kx-text-placeholder);
}

/* Move List */
.move-list {
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid var(--kx-border);
  border-radius: 6px;
}
.move-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  cursor: pointer;
  font-size: 13px;
  border-bottom: 1px solid var(--kx-border);
}
.move-item:last-child {
  border-bottom: none;
}
.move-item:hover {
  background: var(--kx-sidebar-bg);
}
.move-item.active {
  background: rgba(51,112,255,0.08);
  color: var(--kx-primary);
}

/* Sources */
.sources-tip {
  font-size: 13px;
  color: var(--kx-text-secondary);
  margin-bottom: 16px;
  line-height: 1.6;
}
.sources-toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}
.sources-list {
  min-height: 80px;
}
.sources-empty {
  padding: 20px 0;
}
.source-item-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 0;
  border-bottom: 1px solid var(--kx-border);
}
.source-info {
  flex: 1;
  min-width: 0;
}
.source-name {
  font-size: 14px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.source-meta {
  font-size: 12px;
  color: var(--kx-text-placeholder);
  margin-top: 2px;
}
.source-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}
.add-source-desc {
  font-size: 13px;
  color: var(--kx-text-secondary);
  line-height: 1.6;
  background: #f7f8fa;
  padding: 10px 12px;
  border-radius: 6px;
}

/* Chat FAB Button */
.chat-fab {
  position: fixed;
  right: 24px;
  bottom: 24px;
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: var(--kx-primary);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(51, 112, 255, 0.4);
  transition: all 0.2s;
  z-index: 100;
}
.chat-fab:hover {
  transform: scale(1.1);
  box-shadow: 0 6px 16px rgba(51, 112, 255, 0.5);
}

/* Vectorization */
.vector-section {
  padding: 0 4px;
}
.vs-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 12px;
  color: var(--kx-text-primary);
}
.embedding-status, .raptor-status {
  background: #f7f8fa;
  border-radius: 8px;
  padding: 16px;
}
.status-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}
.status-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.status-label {
  font-size: 12px;
  color: var(--kx-text-placeholder);
}
.status-value {
  font-size: 20px;
  font-weight: 600;
  color: var(--kx-text-primary);
}
.status-value.success { color: #36b37e; }
.status-value.warning { color: #f5a623; }
.status-value.danger { color: #f54a45; }
.status-empty {
  font-size: 13px;
  color: var(--kx-text-placeholder);
  text-align: center;
  padding: 16px;
}
.status-actions {
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px solid var(--kx-border);
}
.vector-tip {
  display: flex;
  gap: 8px;
  margin-top: 16px;
  padding: 12px;
  background: #e8f3ff;
  border-radius: 6px;
  font-size: 12px;
  color: var(--kx-text-secondary);
  line-height: 1.6;
}

/* Knowledge Graph Panel */
.graph-panel { height: 100%; display: flex; flex-direction: column; }
.graph-toolbar { display: flex; justify-content: space-between; align-items: center; padding: 0 0 12px; }
.graph-status { display: flex; align-items: center; gap: 12px; }
.graph-stats { color: #909399; font-size: 13px; }
.graph-building, .graph-empty, .graph-error {
  flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center;
}
.graph-building p { margin-top: 16px; color: #606266; }
.graph-building-tip, .graph-empty-tip { color: #909399; font-size: 13px; margin-top: 8px; }

/* ===== Responsive: 768px - Tablet ===== */
@media (max-width: 768px) {
  .kb-sidebar {
    width: 200px;
    min-width: 200px;
  }
  .kb-topbar {
    flex-wrap: wrap;
    gap: 8px;
    padding: 8px 12px;
  }
  .kb-topbar-right {
    flex-wrap: wrap;
    gap: 6px;
  }
  .kb-topbar-right .el-input {
    width: 180px !important;
  }
  .kb-action-bar {
    flex-wrap: wrap;
    gap: 8px;
  }
  .kb-doc-grid {
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  }
}

/* ===== Responsive: 640px - Large Phone ===== */
@media (max-width: 640px) {
  .kb-detail {
    flex-direction: column;
  }
  .kb-sidebar {
    width: 100%;
    min-width: 100%;
    height: auto;
    max-height: 50vh;
    border-right: none;
    border-bottom: 1px solid var(--kx-border);
  }
  .kb-sidebar.collapsed {
    width: 100%;
    min-width: 100%;
    height: 48px;
  }
  .kb-topbar-right {
    width: 100%;
  }
  .kb-topbar-right .el-input {
    width: 100% !important;
  }
  .kb-doc-grid {
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 10px;
  }
  .kb-batch-bar {
    flex-wrap: wrap;
    gap: 6px;
  }
}

/* ===== Responsive: 480px - Small Phone ===== */
@media (max-width: 480px) {
  .kb-topbar {
    padding: 6px 8px;
  }
  .kb-topbar-left .el-tag {
    display: none;
  }
  .kb-content-tabs {
    overflow-x: auto;
  }
  .kb-doc-grid {
    grid-template-columns: 1fr 1fr;
    gap: 8px;
  }
  .kdc-meta span {
    display: none;
  }
}
</style>
