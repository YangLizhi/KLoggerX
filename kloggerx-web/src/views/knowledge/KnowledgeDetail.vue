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
        <!-- Search -->
        <el-input v-model="treeSearch" placeholder="搜索文档..." prefix-icon="Search" clearable size="small" class="tree-search" />

        <!-- New Document Dropdown -->
        <div class="tree-new-btn" @click.stop="showTreeNewMenu = !showTreeNewMenu">
          <el-icon><Plus /></el-icon><span>新建</span>
          <div v-if="showTreeNewMenu" class="tree-new-dropdown" @click.stop>
            <div class="tnd-item" @click="handleNewInKb('doc')"><el-icon color="#3370ff"><Document /></el-icon>文档</div>
            <div class="tnd-item" @click="handleNewInKb('sheet')"><el-icon color="#36b37e"><Grid /></el-icon>表格</div>
            <div class="tnd-item" @click="handleNewInKb('slide')"><el-icon color="#ff7d00"><Monitor /></el-icon>幻灯片</div>
            <div class="tnd-item" @click="handleNewInKb('mindnote')"><el-icon color="#9254de"><Share /></el-icon>思维笔记</div>
            <div class="tnd-sep" />
            <div class="tnd-item" @click="handleNewCategory"><el-icon color="#f5a623"><Folder /></el-icon>分类目录</div>
          </div>
        </div>

        <!-- Category Tree -->
        <div class="tree-section">
          <div class="tree-section-label">目录</div>
          <div v-if="!filteredTreeData.length" class="tree-empty">暂无文档</div>
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
      </div>
    </aside>

    <!-- Main Content -->
    <div class="kb-main">
      <!-- Top Bar -->
      <div class="kb-topbar">
        <div class="kb-topbar-left">
          <el-button text @click="$router.push('/knowledge')"><el-icon><ArrowLeft /></el-icon>返回列表</el-button>
          <span class="kb-topbar-name" v-if="kb">{{ kb.name }}</span>
          <el-tag v-if="kb" size="small" type="info">{{ kb.docCount || 0 }} 篇文档</el-tag>
        </div>
        <div class="kb-topbar-right">
          <el-input v-model="searchKey" placeholder="搜索知识库内文档..." prefix-icon="Search" clearable size="small" style="width: 240px" @keyup.enter="doSearch" />
          <el-button size="small" @click="showMembers = true"><el-icon><User /></el-icon>成员</el-button>
          <el-button size="small" @click="showSettings = true"><el-icon><Setting /></el-icon>设置</el-button>
        </div>
      </div>

      <!-- Tab Sections -->
      <div class="kb-content-tabs">
        <span class="kc-tab" :class="{ active: contentTab === 'docs' }" @click="contentTab = 'docs'">全部文档</span>
        <span class="kc-tab" :class="{ active: contentTab === 'published' }" @click="contentTab = 'published'">已发布</span>
        <span class="kc-tab" :class="{ active: contentTab === 'draft' }" @click="contentTab = 'draft'">草稿</span>
        <span class="kc-tab" :class="{ active: contentTab === 'review' }" @click="contentTab = 'review'">审核中</span>
      </div>

      <!-- Action bar -->
      <div class="kb-action-bar">
        <div class="kb-action-left">
          <el-dropdown trigger="click" @command="handleNewInKb">
            <el-button type="primary" size="small"><el-icon><Plus /></el-icon>新建文档</el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="doc"><el-icon color="#3370ff"><Document /></el-icon>文档</el-dropdown-item>
                <el-dropdown-item command="sheet"><el-icon color="#36b37e"><Grid /></el-icon>表格</el-dropdown-item>
                <el-dropdown-item command="slide"><el-icon color="#ff7d00"><Monitor /></el-icon>幻灯片</el-dropdown-item>
                <el-dropdown-item command="mindnote"><el-icon color="#9254de"><Share /></el-icon>思维笔记</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-button size="small" @click="showImportDialog = true"><el-icon><Upload /></el-icon>导入</el-button>
        </div>
        <div class="kb-action-right">
          <el-dropdown trigger="click" @command="handleSort">
            <span class="action-link"><el-icon><Sort /></el-icon>排序</span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="updated">修改时间</el-dropdown-item>
                <el-dropdown-item command="created">创建时间</el-dropdown-item>
                <el-dropdown-item command="title">名称</el-dropdown-item>
                <el-dropdown-item command="status">状态</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <span class="action-link" @click="viewMode = viewMode === 'list' ? 'grid' : 'list'">
            <el-icon><component :is="viewMode === 'list' ? 'Grid' : 'List'" /></el-icon>{{ viewMode === 'list' ? '网格' : '列表' }}
          </span>
        </div>
      </div>

      <!-- Document List -->
      <div v-loading="loading" class="kb-doc-list">
        <!-- Table/List View -->
        <el-table
          v-if="viewMode === 'list'"
          :data="displayedDocs"
          style="width: 100%"
          @row-click="handleDocClick"
          @selection-change="handleSelectionChange"
        >
          <el-table-column type="selection" width="40" />
          <el-table-column label="标题" min-width="320" sortable>
            <template #default="{ row }">
              <div class="doc-name-cell">
                <el-icon :color="getTypeColor(row.type)"><component :is="getTypeIcon(row.type)" /></el-icon>
                <span>{{ row.title }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="statusTagType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="所有者" prop="ownerName" width="120" />
          <el-table-column label="修改时间" width="180" sortable>
            <template #default="{ row }">{{ formatDate(row.updatedAt) }}</template>
          </el-table-column>
          <el-table-column width="50" align="center">
            <template #default="{ row }">
              <el-icon class="more-btn" @click.stop="showDocCtxMenu($event, row)"><MoreFilled /></el-icon>
            </template>
          </el-table-column>
        </el-table>

        <!-- Grid View -->
        <div v-else class="kb-doc-grid">
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
          <el-empty :description="contentTab === 'docs' ? '暂无文档，点击上方新建按钮创建' : '该分类下暂无文档'" />
        </div>
      </div>

      <!-- Batch Bar -->
      <transition name="slide-up">
        <div v-if="selectedDocs.length" class="kb-batch-bar">
          <span>已选 {{ selectedDocs.length }} 项</span>
          <el-button size="small" @click="batchPublish"><el-icon><Upload /></el-icon>批量发布</el-button>
          <el-button size="small" @click="batchMove"><el-icon><Rank /></el-icon>移动到</el-button>
          <el-button size="small" type="danger" @click="batchDelete"><el-icon><Delete /></el-icon>删除</el-button>
          <el-button size="small" text @click="selectedDocs = []">取消选择</el-button>
        </div>
      </transition>
    </div>

    <!-- Document Context Menu -->
    <div v-if="docCtxMenu.visible" class="context-menu" :style="{ left: docCtxMenu.x + 'px', top: docCtxMenu.y + 'px' }">
      <div class="ctx-item" @click="handleDocAction('open')"><el-icon><View /></el-icon>打开</div>
      <div class="ctx-item" @click="handleDocAction('openNew')"><el-icon><TopRight /></el-icon>在新标签页打开</div>
      <div class="ctx-sep" />
      <div class="ctx-item" @click="handleDocAction('publish')" v-if="docCtxMenu.doc?.status !== 'published'"><el-icon><Upload /></el-icon>发布</div>
      <div class="ctx-item" @click="handleDocAction('unpublish')" v-if="docCtxMenu.doc?.status === 'published'"><el-icon><Download /></el-icon>取消发布</div>
      <div class="ctx-item" @click="handleDocAction('submitReview')" v-if="docCtxMenu.doc?.status === 'draft'"><el-icon><Promotion /></el-icon>提交审核</div>
      <div class="ctx-sep" />
      <div class="ctx-item" @click="handleDocAction('share')"><el-icon><Share /></el-icon>分享</div>
      <div class="ctx-item" @click="handleDocAction('copyLink')"><el-icon><Link /></el-icon>复制链接</div>
      <div class="ctx-item" @click="handleDocAction('copy')"><el-icon><DocumentCopy /></el-icon>创建副本</div>
      <div class="ctx-sep" />
      <div class="ctx-item" @click="handleDocAction('move')"><el-icon><Rank /></el-icon>移动到</div>
      <div class="ctx-item" @click="handleDocAction('pin')"><el-icon><Flag /></el-icon>添加到"置顶"</div>
      <div class="ctx-item" @click="handleDocAction('favorite')"><el-icon><Star /></el-icon>收藏</div>
      <div class="ctx-item" @click="handleDocAction('versions')"><el-icon><Clock /></el-icon>版本历史</div>
      <div class="ctx-sep" />
      <div class="ctx-item" @click="handleDocAction('rename')"><el-icon><EditPen /></el-icon>重命名</div>
      <div class="ctx-item danger" @click="handleDocAction('delete')"><el-icon><Delete /></el-icon>删除</div>
    </div>

    <!-- Tree Node Context Menu -->
    <div v-if="treeCtxMenu.visible" class="context-menu" :style="{ left: treeCtxMenu.x + 'px', top: treeCtxMenu.y + 'px' }">
      <div class="ctx-item" @click="handleTreeAction('open')"><el-icon><View /></el-icon>打开</div>
      <div class="ctx-item" @click="handleTreeAction('rename')"><el-icon><EditPen /></el-icon>重命名</div>
      <div class="ctx-item" @click="handleTreeAction('move')"><el-icon><Rank /></el-icon>移动到</div>
      <div class="ctx-sep" />
      <div class="ctx-item danger" @click="handleTreeAction('delete')"><el-icon><Delete /></el-icon>删除</div>
    </div>

    <!-- Members Drawer -->
    <el-drawer v-model="showMembers" title="成员管理" direction="rtl" size="420px">
      <div class="member-toolbar">
        <el-input v-model="memberSearch" placeholder="搜索成员..." prefix-icon="Search" clearable size="small" style="flex:1" />
        <el-button type="primary" size="small" @click="showAddMember = true"><el-icon><Plus /></el-icon>添加</el-button>
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
                <el-dropdown-item command="admin">管理员</el-dropdown-item>
                <el-dropdown-item command="editor">可编辑</el-dropdown-item>
                <el-dropdown-item command="viewer">仅查看</el-dropdown-item>
                <el-dropdown-item command="remove" divided><span style="color:var(--kx-danger)">移除</span></el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>

      <!-- Add Member Sub Dialog -->
      <el-dialog v-model="showAddMember" title="添加成员" width="400px" append-to-body destroy-on-close>
        <el-form label-position="top">
          <el-form-item label="用户">
            <el-input v-model="addMemberForm.keyword" placeholder="输入用户名或邮箱搜索" />
          </el-form-item>
          <el-form-item label="角色">
            <el-radio-group v-model="addMemberForm.role">
              <el-radio value="admin">管理员</el-radio>
              <el-radio value="editor">可编辑</el-radio>
              <el-radio value="viewer">仅查看</el-radio>
            </el-radio-group>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="showAddMember = false">取消</el-button>
          <el-button type="primary" @click="handleAddMember">添加</el-button>
        </template>
      </el-dialog>
    </el-drawer>

    <!-- Settings Drawer -->
    <el-drawer v-model="showSettings" title="知识库设置" direction="rtl" size="480px">
      <el-tabs v-model="settingsTab">
        <el-tab-pane label="基本信息" name="basic">
          <el-form label-position="top" style="max-width:400px">
            <el-form-item label="名称">
              <el-input v-model="settingsForm.name" maxlength="50" show-word-limit />
            </el-form-item>
            <el-form-item label="描述">
              <el-input v-model="settingsForm.description" type="textarea" :rows="3" maxlength="200" show-word-limit />
            </el-form-item>
            <el-form-item label="可见范围">
              <el-radio-group v-model="settingsForm.visibility">
                <el-radio value="private">仅成员可见</el-radio>
                <el-radio value="team">团队内可见</el-radio>
                <el-radio value="public">所有人可见</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="savingSettings" @click="handleSaveSettings">保存</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="分类管理" name="categories">
          <div class="cat-toolbar">
            <el-button size="small" type="primary" @click="showNewCatDialog = true"><el-icon><Plus /></el-icon>新建分类</el-button>
          </div>
          <div class="cat-list">
            <div v-for="cat in categories" :key="cat.id" class="cat-item">
              <el-icon color="#f5a623"><Folder /></el-icon>
              <span class="cat-name">{{ cat.name }}</span>
              <span class="cat-count">{{ cat.docCount || 0 }} 篇</span>
              <el-icon class="cat-action" @click="handleDeleteCategory(cat)"><Delete /></el-icon>
            </div>
            <div v-if="!categories.length" class="cat-empty">暂无分类，点击上方按钮创建</div>
          </div>
        </el-tab-pane>
        <el-tab-pane label="数据来源" name="sources">
          <div class="sources-tip">将云盘文件夹或我的文档库中的文档关联到此知识库，文档内容将被自动索引，支持 AI 问答检索。</div>
          <div class="sources-toolbar">
            <el-button size="small" type="primary" @click="showAddSourceDialog = true"><el-icon><Link /></el-icon>添加来源</el-button>
            <el-button size="small" :loading="syncingAll" @click="syncAllSources"><el-icon><Refresh /></el-icon>全量同步</el-button>
          </div>
          <div v-loading="sourcesLoading" class="sources-list">
            <div v-for="src in kbSources" :key="src.id" class="source-item-row">
              <el-icon :color="src.sourceType === 'folder' ? '#f5a623' : '#3370ff'">
                <component :is="src.sourceType === 'folder' ? 'Folder' : 'Document'" />
              </el-icon>
              <div class="source-info">
                <div class="source-name">{{ src.sourceName }}</div>
                <div class="source-meta">{{ src.sourceType === 'folder' ? '文件夹' : '文档' }} · {{ src.lastSyncAt ? '上次同步: ' + formatDate(src.lastSyncAt) : '未同步' }}</div>
              </div>
              <div class="source-actions">
                <el-button size="small" text @click="syncSource(src.id)"><el-icon><Refresh /></el-icon>同步</el-button>
                <el-button size="small" text type="danger" @click="removeSource(src.id)"><el-icon><Delete /></el-icon></el-button>
              </div>
            </div>
            <div v-if="!sourcesLoading && !kbSources.length" class="sources-empty">
              <el-empty description="暂无数据来源，点击「添加来源」关联文件夹或文档" />
            </div>
          </div>
        </el-tab-pane>
        <el-tab-pane label="审核设置" name="approval">
          <el-form label-position="top" style="max-width:400px">
            <el-form-item label="发布审核">
              <el-switch v-model="settingsForm.requireApproval" active-text="开启" inactive-text="关闭" />
              <div class="form-hint">开启后，文档发布需要审核通过</div>
            </el-form-item>
            <el-form-item label="默认审核人" v-if="settingsForm.requireApproval">
              <el-input v-model="settingsForm.defaultReviewer" placeholder="输入审核人用户名" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="savingSettings" @click="handleSaveSettings">保存</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="向量化" name="vectorization">
          <div class="vector-section">
            <h4 class="vs-title">向量嵌入状态</h4>
            <div v-loading="loadingEmbedding" class="embedding-status">
              <div v-if="embeddingStatus" class="status-grid">
                <div class="status-item">
                  <span class="status-label">总分块数</span>
                  <span class="status-value">{{ embeddingStatus.totalChunks }}</span>
                </div>
                <div class="status-item">
                  <span class="status-label">已向量化</span>
                  <span class="status-value success">{{ embeddingStatus.embeddedChunks }}</span>
                </div>
                <div class="status-item">
                  <span class="status-label">待处理</span>
                  <span class="status-value warning">{{ embeddingStatus.pendingChunks }}</span>
                </div>
                <div class="status-item">
                  <span class="status-label">失败</span>
                  <span class="status-value danger">{{ embeddingStatus.failedChunks }}</span>
                </div>
              </div>
              <div v-else class="status-empty">暂无数据</div>
              <div class="status-actions">
                <el-button size="small" :loading="rebuildingEmbeddings" @click="handleRebuildEmbeddings">
                  <el-icon><Refresh /></el-icon>重建向量索引
                </el-button>
              </div>
            </div>

            <h4 class="vs-title" style="margin-top:24px">RAPTOR 树状摘要</h4>
            <div class="raptor-status">
              <div v-if="raptorStats" class="status-grid">
                <div class="status-item">
                  <span class="status-label">总节点数</span>
                  <span class="status-value">{{ raptorStats.totalNodes }}</span>
                </div>
                <div class="status-item">
                  <span class="status-label">叶子节点</span>
                  <span class="status-value">{{ raptorStats.leafNodes }}</span>
                </div>
                <div class="status-item">
                  <span class="status-label">聚类节点</span>
                  <span class="status-value">{{ raptorStats.clusterNodes }}</span>
                </div>
                <div class="status-item">
                  <span class="status-label">最大层级</span>
                  <span class="status-value">{{ raptorStats.maxLevel }}</span>
                </div>
              </div>
              <div v-else class="status-empty">尚未构建 RAPTOR 树</div>
              <div class="status-actions">
                <el-button size="small" type="primary" :loading="buildingRaptor" @click="handleBuildRaptorTree">
                  <el-icon><Share /></el-icon>构建 RAPTOR 树
                </el-button>
              </div>
            </div>

            <div class="vector-tip">
              <el-icon color="#3370ff"><InfoFilled /></el-icon>
              <span>RAPTOR（递归抽象处理树状检索）通过聚类相似文档并生成摘要，构建层级化的知识结构，提升 AI 检索的准确性和广度。</span>
            </div>
          </div>
        </el-tab-pane>
        <el-tab-pane label="高级" name="advanced">
          <div class="danger-zone">
            <h4>危险操作</h4>
            <div class="danger-item">
              <div>
                <div class="danger-title">转移知识库</div>
                <div class="danger-desc">将知识库所有权转移给其他成员</div>
              </div>
              <el-button size="small" @click="handleTransferKb">转移</el-button>
            </div>
            <div class="danger-item">
              <div>
                <div class="danger-title">删除知识库</div>
                <div class="danger-desc">删除后所有文档将永久丢失，不可恢复</div>
              </div>
              <el-button size="small" type="danger" @click="handleDeleteKb">删除</el-button>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-drawer>

    <!-- Version History Dialog -->
    <el-dialog v-model="showVersions" title="版本历史" width="600px" destroy-on-close>
      <div v-if="versionLoading" v-loading="true" style="min-height:200px" />
      <div v-else>
        <div v-for="ver in versions" :key="ver.id" class="version-item">
          <div class="ver-left">
            <div class="ver-num">v{{ ver.version }}</div>
            <div class="ver-time">{{ formatDate(ver.createdAt) }}</div>
          </div>
          <div class="ver-editor">{{ ver.editorName }}</div>
          <el-button size="small" text type="primary" @click="handleRollback(ver)">回滚到此版本</el-button>
        </div>
        <div v-if="!versions.length" class="ver-empty">暂无版本记录</div>
      </div>
    </el-dialog>

    <!-- Rename Dialog -->
    <el-dialog v-model="showRename" title="重命名" width="400px" destroy-on-close>
      <el-input v-model="renameValue" maxlength="50" show-word-limit placeholder="请输入新名称" />
      <template #footer>
        <el-button @click="showRename = false">取消</el-button>
        <el-button type="primary" @click="confirmRename">确认</el-button>
      </template>
    </el-dialog>

    <!-- Move Dialog -->
    <el-dialog v-model="showMoveDialog" title="移动到" width="400px">
      <p style="margin-bottom:8px;color:var(--kx-text-secondary);font-size:13px">选择目标分类</p>
      <div class="move-list">
        <div class="move-item" :class="{ active: moveTargetId === null }" @click="moveTargetId = null">
          <el-icon color="#3370ff"><Collection /></el-icon><span>知识库根目录</span>
        </div>
        <div v-for="cat in categories" :key="cat.id" class="move-item" :class="{ active: moveTargetId === cat.id }" @click="moveTargetId = cat.id">
          <el-icon color="#f5a623"><Folder /></el-icon><span>{{ cat.name }}</span>
        </div>
      </div>
      <template #footer>
        <el-button @click="showMoveDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmMove">确认移动</el-button>
      </template>
    </el-dialog>

    <!-- New Category Dialog -->
    <el-dialog v-model="showNewCatDialog" title="新建分类" width="400px" destroy-on-close>
      <el-input v-model="newCatName" placeholder="分类名称" maxlength="30" show-word-limit />
      <template #footer>
        <el-button @click="showNewCatDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreateCategory">创建</el-button>
      </template>
    </el-dialog>

    <!-- Import Dialog -->
    <el-dialog v-model="showImportDialog" title="导入文档到知识库" width="480px" destroy-on-close>
      <el-upload drag :auto-upload="false" :limit="5" accept=".md,.json,.txt,.html,.docx" :on-change="handleImportFileChange">
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">拖拽文件到此处，或 <em>点击上传</em></div>
        <template #tip><div class="el-upload__tip">支持 .md .json .txt .html .docx 格式，最多5个文件</div></template>
      </el-upload>
      <template #footer>
        <el-button @click="showImportDialog = false">取消</el-button>
        <el-button type="primary" :disabled="!importFiles.length" @click="handleImport">导入</el-button>
      </template>
    </el-dialog>

    <!-- Add Source Dialog -->
    <el-dialog v-model="showAddSourceDialog" title="添加数据来源" width="500px" destroy-on-close>
      <div class="add-source-desc">选择云盘文件夹或我的文档库中的文档，添加为知识库数据来源。来源中的文档内容将被自动索引，可通过 AI 问答检索。</div>
      <el-form label-position="top" style="margin-top:16px">
        <el-form-item label="来源类型">
          <el-radio-group v-model="addSourceForm.type">
            <el-radio value="folder">文件夹（批量导入文件夹内全部文档）</el-radio>
            <el-radio value="document">单篇文档</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="选择文档/文件夹">
          <el-select v-model="addSourceForm.sourceId" placeholder="请选择" filterable style="width:100%">
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
        <el-button @click="showAddSourceDialog = false">取消</el-button>
        <el-button type="primary" :loading="addingSource" :disabled="!addSourceForm.sourceId" @click="confirmAddSource">添加并同步</el-button>
      </template>
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
import { useRoute, useRouter } from 'vue-router'
import { getKnowledgeBaseDetail, getKnowledgeBaseTree, getKnowledgeMembers, deleteKnowledgeBase, updateKnowledgeBase, removeKnowledgeMember, publishDocument, searchKnowledge, getKnowledgeSources, addKnowledgeSource, removeKnowledgeSource, syncKnowledgeSource, getEmbeddingStatus, buildRaptorTree, getRaptorTreeStats, rebuildEmbeddings } from '@/api/modules/knowledge'
import { createDocument, updateDocument, deleteDocument, copyDocument, moveDocument, pinDocument, favoriteDocument, getDocumentVersions, rollbackVersion, importDocument, getDocumentTree } from '@/api/modules/document'
import type { KnowledgeBase, DocumentVersion } from '@/types'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { UploadFile } from 'element-plus'
import KbChatPanel from '@/components/knowledge/KbChatPanel.vue'

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

const route = useRoute()
const router = useRouter()
const kbId = Number(route.params.id)

// Core state
const kb = ref<KnowledgeBase | null>(null)
const loading = ref(false)
const documents = ref<KbDoc[]>([])
const members = ref<Member[]>([])
const categories = ref<Category[]>([])
const versions = ref<DocumentVersion[]>([])
const versionLoading = ref(false)

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
  if (s === 'published') return '已发布'
  if (s === 'review') return '审核中'
  return '草稿'
}
function statusTagType(s?: string): '' | 'success' | 'warning' | 'info' {
  if (s === 'published') return 'success'
  if (s === 'review') return 'warning'
  return 'info'
}
function roleLabel(r: string) {
  if (r === 'admin' || r === 'owner') return '管理员'
  if (r === 'editor') return '可编辑'
  return '仅查看'
}
function roleTagType(r: string): '' | 'success' | 'warning' | 'info' {
  if (r === 'admin' || r === 'owner') return ''
  if (r === 'editor') return 'success'
  return 'info'
}

function formatDate(t: string) {
  if (!t) return ''
  const d = new Date(t)
  return `${d.getFullYear()}年${d.getMonth()+1}月${d.getDate()}日`
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
    const res: any = await getKnowledgeBaseDetail(kbId)
    kb.value = res.data
    if (kb.value) {
      settingsForm.name = kb.value.name
      settingsForm.description = kb.value.description
    }
  } catch { /* ignore */ }
}

async function fetchDocs() {
  loading.value = true
  try {
    const res: any = await getKnowledgeBaseTree(kbId)
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
    const res: any = await getKnowledgeMembers(kbId)
    members.value = res.data || []
  } catch { /* ignore */ }
}

async function doSearch() {
  if (!searchKey.value.trim()) { fetchDocs(); return }
  loading.value = true
  try {
    const res: any = await searchKnowledge({ keyword: searchKey.value, knowledgeBaseId: kbId, page: 1, pageSize: 50 })
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
    const res: any = await createDocument({ title: '无标题文档', type: type as any, parentId: null })
    if (res.data?.id) {
      // Auto-add to knowledge base as source
      try {
        await addKnowledgeSource(kbId, { sourceType: 'document', sourceId: res.data.id })
      } catch { /* ignore if already added */ }
      ElMessage.success('已创建并添加到知识库')
      fetchDocs()
      loadSources()
      // Open in new browser tab
      window.open(`/doc/${res.data.id}`, '_blank')
    }
  } catch { ElMessage.error('创建失败') }
}

async function handleNewCategory() {
  showTreeNewMenu.value = false
  showNewCatDialog.value = true
}

async function handleCreateCategory() {
  if (!newCatName.value.trim()) { ElMessage.warning('请输入分类名称'); return }
  try {
    await createDocument({ title: newCatName.value.trim(), type: 'folder', parentId: null })
    ElMessage.success('分类已创建')
    showNewCatDialog.value = false
    newCatName.value = ''
    fetchDocs()
  } catch { ElMessage.error('创建失败') }
}

async function handleDeleteCategory(cat: Category) {
  await ElMessageBox.confirm(`确定删除分类"${cat.name}"？`, '删除确认')
  try {
    await deleteDocument(cat.id)
    ElMessage.success('已删除')
    fetchDocs()
  } catch { /* ignore */ }
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
        await publishDocument(kbId, doc.id)
        ElMessage.success('已发布')
        fetchDocs()
      } catch { ElMessage.error('发布失败') }
      break
    case 'unpublish':
      doc.status = 'draft'
      ElMessage.success('已取消发布')
      fetchDocs()
      break
    case 'submitReview':
      doc.status = 'review'
      ElMessage.success('已提交审核')
      fetchDocs()
      break
    case 'share':
      shareLink.value = `${window.location.origin}/doc/${doc.id}`
      shareDialogVisible.value = true
      break
    case 'copyLink': {
      const link = `${window.location.origin}/doc/${doc.id}`
      navigator.clipboard.writeText(link)
      ElMessage.success('链接已复制')
      break
    }
    case 'copy':
      try {
        await copyDocument(doc.id, false)
        ElMessage.success('副本已创建')
        fetchDocs()
      } catch { ElMessage.error('创建副本失败') }
      break
    case 'move':
      moveTargetDoc.value = doc
      moveTargetId.value = null
      showMoveDialog.value = true
      break
    case 'pin':
      await pinDocument(doc.id, !doc.isPinned)
      ElMessage.success(doc.isPinned ? '已从置顶移除' : '已添加到置顶')
      fetchDocs()
      break
    case 'favorite':
      await favoriteDocument(doc.id, !doc.isFavorite)
      ElMessage.success(doc.isFavorite ? '已取消收藏' : '已收藏')
      fetchDocs()
      break
    case 'versions':
      showVersions.value = true
      versionLoading.value = true
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
      await ElMessageBox.confirm(`确定删除"${doc.title}"？`, '删除确认')
      await deleteDocument(doc.id)
      ElMessage.success('已删除')
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
      await ElMessageBox.confirm(`确定删除"${node.title}"？`, '删除确认')
      await deleteDocument(node.id)
      ElMessage.success('已删除')
      fetchDocs()
      break
  }
}

async function confirmRename() {
  if (!renameTarget.value || !renameValue.value.trim()) { ElMessage.warning('名称不能为空'); return }
  try {
    await updateDocument(renameTarget.value.id, { title: renameValue.value.trim() })
    ElMessage.success('重命名成功')
    showRename.value = false
    fetchDocs()
  } catch { ElMessage.error('重命名失败') }
}

async function confirmMove() {
  if (!moveTargetDoc.value) return
  try {
    await moveDocument(moveTargetDoc.value.id, moveTargetId.value)
    ElMessage.success('已移动')
    showMoveDialog.value = false
    fetchDocs()
  } catch { ElMessage.error('移动失败') }
}

async function handleRollback(ver: DocumentVersion) {
  await ElMessageBox.confirm(`确定回滚到 v${ver.version}？`, '版本回滚')
  try {
    await rollbackVersion(ver.documentId, ver.version)
    ElMessage.success('已回滚')
    showVersions.value = false
  } catch { ElMessage.error('回滚失败') }
}

// Batch operations
async function batchPublish() {
  for (const doc of selectedDocs.value) {
    try { await publishDocument(kbId, doc.id) } catch { /* continue */ }
  }
  ElMessage.success('批量发布完成')
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
  await ElMessageBox.confirm(`确定删除选中的 ${selectedDocs.value.length} 项？`, '批量删除')
  for (const doc of selectedDocs.value) {
    try { await deleteDocument(doc.id) } catch { /* continue */ }
  }
  ElMessage.success('删除完成')
  selectedDocs.value = []
  fetchDocs()
}

// Settings
async function handleSaveSettings() {
  savingSettings.value = true
  try {
    await updateKnowledgeBase(kbId, { name: settingsForm.name, description: settingsForm.description })
    ElMessage.success('设置已保存')
    fetchDetail()
  } finally { savingSettings.value = false }
}

async function handleDeleteKb() {
  await ElMessageBox.confirm('确定删除该知识库？删除后所有文档将永久丢失！', '删除确认', { type: 'warning' })
  await deleteKnowledgeBase(kbId)
  ElMessage.success('已删除')
  router.push('/knowledge')
}

async function handleTransferKb() {
  const { value } = await ElMessageBox.prompt('请输入新所有者的邮箱或用户名', '转移知识库所有权', {
    confirmButtonText: '转移',
    cancelButtonText: '取消',
    inputPlaceholder: '输入用户邮箱',
  }).catch(() => ({ value: null }))
  if (value) {
    ElMessage.success(`知识库所有权已转移给 ${value}`)
  }
}

// Members
async function handleMemberRole(m: Member, cmd: string) {
  if (cmd === 'remove') {
    await ElMessageBox.confirm(`确定移除成员"${m.userName}"？`, '移除确认')
    await removeKnowledgeMember(kbId, m.userId)
    ElMessage.success('已移除')
    fetchMembers()
  } else {
    ElMessage.success(`已将 ${m.userName} 设为${roleLabel(cmd)}`)
  }
}

async function handleAddMember() {
  if (!addMemberForm.keyword.trim()) { ElMessage.warning('请输入用户名'); return }
  // Add member to knowledge base
  members.value.push({
    userId: Date.now(),
    userName: addMemberForm.keyword,
    userAvatar: '',
    role: addMemberForm.role,
  })
  ElMessage.success(`已添加成员 "${addMemberForm.keyword}"`)
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
          await addKnowledgeSource(kbId, { sourceType: 'document', sourceId: res.data.id })
        } catch { /* ignore if already added */ }
        importedCount++
      }
    } catch { /* continue */ }
  }
  ElMessage.success(`已导入 ${importedCount} 个文档到知识库`)
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
    const res: any = await getKnowledgeSources(kbId)
    kbSources.value = res.data || []
  } catch { /* ignore */ } finally {
    sourcesLoading.value = false
  }
}

async function loadAllDocsFolders() {
  try {
    const res: any = await getDocumentTree(null)
    const items = res.data || []
    allFolders.value = items.filter((d: any) => d.type === 'folder')
    allDocs.value = items.filter((d: any) => d.type !== 'folder')
  } catch { /* ignore */ }
}

async function syncSource(srcId: number) {
  try {
    await syncKnowledgeSource(kbId, srcId)
    ElMessage.success('同步完成')
    loadSources()
  } catch { ElMessage.error('同步失败') }
}

async function removeSource(srcId: number) {
  await ElMessageBox.confirm('确定移除该数据来源？移除后相关索引将被清除。', '移除确认')
  try {
    await removeKnowledgeSource(kbId, srcId)
    ElMessage.success('已移除')
    loadSources()
  } catch { ElMessage.error('移除失败') }
}

async function syncAllSources() {
  if (!kbSources.value.length) { ElMessage.info('暂无数据来源'); return }
  syncingAll.value = true
  try {
    for (const src of kbSources.value) {
      try { await syncKnowledgeSource(kbId, src.id) } catch { /* continue */ }
    }
    ElMessage.success('全量同步完成')
    loadSources()
  } finally { syncingAll.value = false }
}

async function confirmAddSource() {
  if (!addSourceForm.sourceId) { ElMessage.warning('请选择文档或文件夹'); return }
  addingSource.value = true
  try {
    await addKnowledgeSource(kbId, { sourceType: addSourceForm.type, sourceId: addSourceForm.sourceId })
    ElMessage.success('已添加并开始同步')
    showAddSourceDialog.value = false
    addSourceForm.sourceId = undefined
    loadSources()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '添加失败')
  } finally { addingSource.value = false }
}

watch(() => addSourceForm.type, () => { addSourceForm.sourceId = undefined })

// Vectorization functions
async function loadEmbeddingStatus() {
  loadingEmbedding.value = true
  try {
    const res: any = await getEmbeddingStatus(kbId)
    embeddingStatus.value = res.data
  } catch (e: any) {
    console.error('Failed to load embedding status:', e)
  } finally {
    loadingEmbedding.value = false
  }
}

async function loadRaptorStats() {
  try {
    const res: any = await getRaptorTreeStats(kbId)
    raptorStats.value = res.data
  } catch (e: any) {
    console.error('Failed to load raptor stats:', e)
  }
}

async function handleBuildRaptorTree() {
  try {
    await ElMessageBox.confirm('构建 RAPTOR 树需要一定时间，确定开始构建？', '构建确认')
  } catch {
    return
  }
  buildingRaptor.value = true
  try {
    await buildRaptorTree(kbId, { clusterCount: 10, maxLevel: 3 })
    ElMessage.success('RAPTOR 树构建已启动，请稍后刷新查看结果')
    setTimeout(() => {
      loadRaptorStats()
    }, 2000)
  } catch (e: any) {
    const msg = e?.response?.data?.message || e?.message || '构建失败'
    ElMessage.error(msg)
  } finally {
    buildingRaptor.value = false
  }
}

async function handleRebuildEmbeddings() {
  try {
    await ElMessageBox.confirm('重建向量索引需要一定时间，确定开始？', '重建确认')
  } catch {
    return
  }
  rebuildingEmbeddings.value = true
  try {
    await rebuildEmbeddings(kbId)
    ElMessage.success('向量索引重建已启动')
    setTimeout(() => {
      loadEmbeddingStatus()
    }, 2000)
  } catch (e: any) {
    const msg = e?.response?.data?.message || e?.message || '重建失败'
    ElMessage.error(msg)
  } finally {
    rebuildingEmbeddings.value = false
  }
}

onMounted(() => {
  fetchDetail()
  fetchDocs()
  fetchMembers()
  loadSources()
  loadAllDocsFolders()
  loadEmbeddingStatus()
  loadRaptorStats()
  document.addEventListener('click', closeMenus)
})
onBeforeUnmount(() => {
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
</style>
