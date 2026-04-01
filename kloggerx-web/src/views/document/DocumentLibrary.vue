<template>
  <div class="cloud-drive cloud-drive-full">
    <!-- Left Folder Panel (hidden when MainLayout sidebar handles navigation) -->
    <aside class="folder-panel">
      <div class="panel-section">
        <div class="panel-section-header" @click="myFoldersExpanded = !myFoldersExpanded">
          <el-icon class="expand-arrow" :class="{ expanded: myFoldersExpanded }"><ArrowRight /></el-icon>
          <span class="panel-section-title">我的文件夹</span>
          <el-icon class="panel-add-btn" @click.stop="openCreateFolderDialog(null)" title="新建文件夹"><Plus /></el-icon>
        </div>
        <div v-show="myFoldersExpanded" class="panel-section-body">
          <div v-if="!folderTree.length" class="panel-empty">暂无文件夹</div>
          <FolderTreeNode
            v-for="node in folderTree"
            :key="node.id"
            :node="(node as any)"
            :level="0"
            :selected-folder-id="selectedFolderId"
            @select="(n: any) => selectFolder(n)"
            @contextmenu="(e: any, n: any) => showFolderCtxMenu(e, n)"
            @toggle-expand="(n: any) => toggleExpand(n)"
          />
        </div>
      </div>
      <div class="panel-section">
        <div class="panel-section-header" @click="sharedFoldersExpanded = !sharedFoldersExpanded">
          <el-icon class="expand-arrow" :class="{ expanded: sharedFoldersExpanded }"><ArrowRight /></el-icon>
          <span class="panel-section-title">共享文件夹</span>
        </div>
        <div v-show="sharedFoldersExpanded" class="panel-section-body">
          <div v-if="!sharedFolderTree.length" class="panel-empty">暂无共享文件夹</div>
          <FolderTreeNode
            v-for="node in sharedFolderTree"
            :key="node.id"
            :node="(node as any)"
            :level="0"
            :selected-folder-id="selectedFolderId"
            @select="(n: any) => selectFolder(n)"
            @contextmenu="(e: any, n: any) => showFolderCtxMenu(e, n)"
            @toggle-expand="(n: any) => toggleExpand(n)"
          />
        </div>
      </div>
      <div class="panel-section">
        <div class="panel-section-header panel-quick-access">
          <el-icon :size="14" color="#3370ff"><Star /></el-icon>
          <span class="panel-section-title">快速访问</span>
        </div>
        <div class="panel-section-body">
          <div v-if="!quickAccessFolders.length" class="panel-empty">暂无快速访问</div>
          <div
            v-for="f in quickAccessFolders"
            :key="f.id"
            class="folder-tree-item"
            :class="{ active: selectedFolderId === f.id }"
            @click="selectFolder(f)"
          >
            <el-icon color="#f5a623" :size="16"><Folder /></el-icon>
            <span class="folder-tree-name">{{ f.title }}</span>
          </div>
        </div>
      </div>
    </aside>

    <!-- Right Content Area -->
    <div class="drive-content">
      <!-- Top Action Bar -->
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
            <div class="dropdown-item" @click="openCreateFolderDialog(currentParentId)"><el-icon color="#f5a623"><Folder /></el-icon>文件夹</div>
            <div class="dropdown-group-title">文档应用</div>
            <div class="dropdown-item" @click="handleCreate('doc')"><el-icon color="#36b37e"><EditPen /></el-icon>画板</div>
            <div class="dropdown-item" @click="handleCreate('mindnote')"><el-icon color="#9254de"><Share /></el-icon>思维导图</div>
            <div class="dropdown-item" @click="handleCreate('doc')"><el-icon color="#ff7d00"><Connection /></el-icon>流程图</div>
          </div>
        </div>
        <div class="action-card" @click.stop="showUploadMenu = !showUploadMenu; showNewMenu = false">
          <div class="action-icon" style="background: #fef3e0"><el-icon :size="20" color="#f5a623"><Upload /></el-icon></div>
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
        <div class="action-card" @click="handleAddShortcut">
          <div class="action-icon" style="background: #e8f5e9"><el-icon :size="20" color="#36b37e"><Link /></el-icon></div>
          <div class="action-info">
            <div class="action-label">添加</div>
            <div class="action-desc">添加云文档的快捷方式</div>
          </div>
        </div>
        <div class="action-card" @click="showTemplateLibrary = true">
          <div class="action-icon" style="background: #fce4ec"><el-icon :size="20" color="#f54a45"><Files /></el-icon></div>
          <div class="action-info">
            <div class="action-label">模板库</div>
            <div class="action-desc">选择模板快速新建</div>
          </div>
        </div>
      </div>

      <!-- Documents Section -->
      <div class="docs-section">
        <div class="docs-section-header">
          <div class="docs-breadcrumb">
            <span v-for="(b, i) in breadcrumbs" :key="b.id" class="breadcrumb-item" @click="navigateTo(b.id)">
              {{ b.title }}<span v-if="i < breadcrumbs.length - 1" class="sep"> &gt; </span>
            </span>
          </div>
          <div class="docs-section-actions">
            <el-dropdown trigger="click" @command="handleSort">
              <span class="tab-action"><el-icon><Sort /></el-icon> 排序</span>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="updated">修改时间</el-dropdown-item>
                  <el-dropdown-item command="created">创建时间</el-dropdown-item>
                  <el-dropdown-item command="title">名称</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
            <el-icon class="view-icon" :class="{ active: viewMode === 'list' }" @click="viewMode = 'list'"><List /></el-icon>
            <el-icon class="view-icon" :class="{ active: viewMode === 'grid' }" @click="viewMode = 'grid'"><Grid /></el-icon>
          </div>
        </div>

        <div v-loading="loading">
          <!-- Document list (folders first, then files) -->
          <div v-if="sortedDocuments.length">
            <!-- List view -->
            <el-table
              v-if="viewMode === 'list'"
              :data="sortedDocuments"
              style="width: 100%"
              table-layout="auto"
              @row-dblclick="handleDocClick"
              @selection-change="handleSelectionChange"
              @row-contextmenu="(row: Document, _column: any, e: MouseEvent) => showRowContextMenu(e, row)"
            >
              <el-table-column type="selection" width="50" />
              <el-table-column label="标题" min-width="240" sortable>
                <template #default="{ row }">
                  <div class="doc-name-cell">
                    <el-icon :color="getTypeColor(row.type)"><component :is="getTypeIcon(row.type)" /></el-icon>
                    <span>{{ getDisplayName(row) }}</span>
                    <el-icon v-if="row.isPinned" class="pin-badge" color="#3370ff"><Flag /></el-icon>
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="文件类型" min-width="110">
                <template #default="{ row }">
                  <el-tag size="small" :type="getTypeTagType(row.type)" disable-transitions>{{ getTypeName(row) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="文件大小" min-width="100">
                <template #default="{ row }">
                  <span>{{ row.fileSize > 0 ? formatFileSize(row.fileSize) : '—' }}</span>
                </template>
              </el-table-column>
              <el-table-column label="修改时间" min-width="130" sortable>
                <template #default="{ row }">{{ formatDate(row.updatedAt) }}</template>
              </el-table-column>
              <el-table-column label="创建时间" min-width="130" sortable>
                <template #default="{ row }">{{ formatDate(row.createdAt) }}</template>
              </el-table-column>
            </el-table>

            <!-- Grid view -->
            <div v-else class="doc-grid">
              <div
                v-for="doc in sortedDocuments"
                :key="doc.id"
                class="doc-card"
                :class="{ 'drag-over': dragOverDocId === doc.id }"
                draggable="true"
                @dblclick="handleDocClick(doc)"
                @contextmenu.prevent="doc.type === 'folder' ? showFolderCtxMenu($event, doc) : showContextMenu($event, doc)"
                @dragstart="handleDragStart($event, doc)"
                @dragend="handleDragEnd"
                @dragover.prevent="handleDragOver($event, doc)"
                @dragleave="handleDragLeave"
                @drop.prevent="handleDrop($event, doc)"
              >
                <div class="doc-card-icon">
                  <el-icon :size="36" :color="getTypeColor(doc.type)"><component :is="getTypeIcon(doc.type)" /></el-icon>
                </div>
                <div class="doc-card-title">{{ getDisplayName(doc) }}</div>
                <div class="doc-card-meta">{{ formatDate(doc.updatedAt) }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="end-marker" v-if="sortedDocuments.length">
        <span>已经到底了</span>
      </div>

      <div v-if="!loading && !documents.length" class="empty-state">
        <el-empty description="暂无文档" />
      </div>
    </div>

    <!-- Document Context Menu -->
    <div v-if="contextMenu.visible" class="context-menu" :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }">
      <div class="ctx-item" @click="handleShareDoc(contextMenu.doc!)"><el-icon><Share /></el-icon>分享</div>
      <div class="ctx-item" @click="handleCopyLink(contextMenu.doc!)"><el-icon><Link /></el-icon>复制链接</div>
      <div class="ctx-sep" />
      <div class="ctx-item" @click="handleCopy(contextMenu.doc!)"><el-icon><DocumentCopy /></el-icon>创建副本</div>
      <div class="ctx-item" @click="showMoveDialog(contextMenu.doc!)"><el-icon><Rank /></el-icon>移动到</div>
      <div class="ctx-item" @click="handleAddShortcutFor(contextMenu.doc!)"><el-icon><Position /></el-icon>添加快捷方式到</div>
      <div class="ctx-item" @click="handlePin(contextMenu.doc!)"><el-icon><Flag /></el-icon>{{ contextMenu.doc?.isPinned ? '从"置顶"移除' : '添加到"置顶"' }}</div>
      <div class="ctx-item" @click="handleFavorite(contextMenu.doc!)"><el-icon><Star /></el-icon>{{ contextMenu.doc?.isFavorite ? '取消收藏' : '收藏' }}</div>
      <div class="ctx-sep" />
      <div class="ctx-item" @click="handleTransfer(contextMenu.doc!)"><el-icon><Switch /></el-icon>转移所有权</div>
      <div class="ctx-item danger" @click="handleDelete(contextMenu.doc!)"><el-icon><Delete /></el-icon>删除</div>
    </div>

    <!-- Folder Context Menu -->
    <div v-if="folderCtxMenu.visible" class="context-menu" :style="{ left: folderCtxMenu.x + 'px', top: folderCtxMenu.y + 'px' }">
      <div class="ctx-item" @click="handleFolderAction('open')"><el-icon><View /></el-icon>在新标签页打开</div>
      <div class="ctx-sep" />
      <div class="ctx-item" @click="handleFolderAction('new')"><el-icon><Plus /></el-icon>新建</div>
      <div class="ctx-item" @click="handleFolderAction('upload')"><el-icon><Upload /></el-icon>上传</div>
      <div class="ctx-sep" />
      <div class="ctx-item" @click="handleFolderAction('share')"><el-icon><Share /></el-icon>分享</div>
      <div class="ctx-item" @click="handleFolderAction('copyLink')"><el-icon><Link /></el-icon>复制链接</div>
      <div class="ctx-sep" />
      <div class="ctx-item" @click="handleFolderAction('move')"><el-icon><Rank /></el-icon>移动到</div>
      <div class="ctx-item" @click="handleFolderAction('quickAccess')"><el-icon><Star /></el-icon>添加到"快速访问文件夹"</div>
      <div class="ctx-item" @click="handleFolderAction('favorite')"><el-icon><StarFilled /></el-icon>收藏</div>
      <div class="ctx-sep" />
      <div class="ctx-item" @click="handleFolderAction('transfer')"><el-icon><Switch /></el-icon>转移所有权</div>
      <div class="ctx-item" @click="handleFolderAction('download')"><el-icon><Download /></el-icon>下载</div>
      <div class="ctx-item" @click="handleFolderAction('rename')"><el-icon><EditPen /></el-icon>重命名</div>
      <div class="ctx-item danger" @click="handleFolderAction('delete')"><el-icon><Delete /></el-icon>删除</div>
    </div>

    <!-- Create Folder Dialog -->
    <el-dialog v-model="createFolderVisible" title="新建文件夹到云盘" width="480px" destroy-on-close>
      <el-form :model="createFolderForm" label-position="top">
        <el-form-item label="名称">
          <el-input v-model="createFolderForm.name" placeholder="请输入文件夹名称" maxlength="50" show-word-limit />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="createFolderForm.description" type="textarea" :rows="3" placeholder="添加描述（可选）" maxlength="200" show-word-limit />
        </el-form-item>
        <el-form-item label="邀请协作者">
          <el-input v-model="createFolderForm.collaborator" placeholder="输入姓名或邮箱搜索" prefix-icon="Search">
            <template #append>
              <el-button @click="searchCollaborator">搜索</el-button>
            </template>
          </el-input>
          <div class="collaborator-hint">可在创建后通过分享功能邀请更多协作者</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createFolderVisible = false">取消</el-button>
        <el-button type="primary" :loading="creatingFolder" @click="handleCreateFolder">创建</el-button>
      </template>
    </el-dialog>

    <!-- Move Dialog -->
    <el-dialog v-model="moveDialogVisible" title="移动到" width="400px">
      <p>选择目标文件夹</p>
      <div class="move-folder-list">
        <div class="move-folder-item" :class="{ active: moveTargetParentId === null }" @click="moveTargetParentId = null">
          <el-icon color="#f5a623"><FolderOpened /></el-icon>
          <span>根目录</span>
        </div>
        <div v-for="f in allFoldersList" :key="f.id" class="move-folder-item" :class="{ active: moveTargetParentId === f.id }" @click="moveTargetParentId = f.id">
          <el-icon color="#f5a623"><Folder /></el-icon>
          <span>{{ f.title }}</span>
        </div>
      </div>
      <template #footer>
        <el-button @click="moveDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmMove">确认移动</el-button>
      </template>
    </el-dialog>

    <!-- Rename Dialog -->
    <el-dialog v-model="renameDialogVisible" title="重命名" width="400px" destroy-on-close>
      <el-input v-model="renameValue" placeholder="请输入新名称" maxlength="50" show-word-limit />
      <template #footer>
        <el-button @click="renameDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="renaming" @click="confirmRename">确认</el-button>
      </template>
    </el-dialog>

    <!-- Import Dialog -->
    <el-dialog v-model="showImportDialog" :title="uploadDialogTitle" width="480px" destroy-on-close @close="importFile = null; importFiles = []; importProgress = 0">
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
        ref="uploadRef"
        drag
        :auto-upload="false"
        :limit="1"
        accept=".md,.json,.txt,.html,.docx,.doc,.xlsx,.xls,.pptx,.ppt,.pdf,.png,.jpg,.jpeg,.gif,.webp"
        :on-change="handleFileChange"
        :on-exceed="() => ElMessage.warning('只能上传一个文件')"
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">拖拽文件到此处，或 <em>点击上传</em></div>
        <template #tip><div class="el-upload__tip">支持 Word、Excel、PPT、PDF、Markdown、JSON、TXT、HTML、图片格式</div></template>
      </el-upload>
      <div v-if="importFiles.length > 0" class="import-file-list">
        <div v-for="(f, i) in importFiles" :key="i" class="import-file-item">
          <el-icon><Document /></el-icon>
          <span>{{ f.name }}</span>
          <span class="file-size">{{ formatFileSize(f.size) }}</span>
        </div>
      </div>
      <el-progress v-if="importProgress > 0 && importProgress < 100" :percentage="importProgress" style="margin-top: 8px" />
      <template #footer>
        <el-button @click="showImportDialog = false; importFile = null; importFiles = []; importProgress = 0">取消</el-button>
        <el-button type="primary" :loading="importLoading" :disabled="!importFile && !importFiles.length" @click="handleImport">导入</el-button>
      </template>
    </el-dialog>

    <!-- Share Dialog -->
    <el-dialog v-model="shareDialogVisible" title="分享" width="500px" destroy-on-close>
      <div class="share-section">
        <div class="share-label">链接分享</div>
        <div class="share-link-row">
          <el-input :model-value="shareLink" readonly />
          <el-button type="primary" @click="copyShareLink">复制链接</el-button>
        </div>
        <el-radio-group v-model="shareLinkScope" style="margin-top:8px">
          <el-radio value="collaborator">仅协作者</el-radio>
          <el-radio value="org">组织内获得链接的人</el-radio>
          <el-radio value="public">互联网上获得链接的人</el-radio>
        </el-radio-group>
      </div>
      <div class="share-section" style="margin-top:16px">
        <div class="share-label">邀请协作者</div>
        <div class="share-invite-row">
          <el-input v-model="shareInviteEmail" placeholder="输入邮箱或用户名" style="flex:1" />
          <el-select v-model="shareInvitePermission" style="width:100px">
            <el-option label="可编辑" value="edit" />
            <el-option label="可阅读" value="view" />
          </el-select>
          <el-button @click="handleShareInvite">邀请</el-button>
        </div>
      </div>
      <template #footer>
        <el-button @click="shareDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- Template Library Dialog -->
    <el-dialog v-model="showTemplateLibrary" title="模板库" width="800px" destroy-on-close>
      <div class="template-library">
        <div class="template-categories">
          <div
            v-for="cat in templateCategories"
            :key="cat.id"
            class="template-cat"
            :class="{ active: selectedTemplateCat === cat.id }"
            @click="selectedTemplateCat = cat.id"
          >
            {{ cat.name }}
          </div>
        </div>
        <div class="template-grid">
          <div v-for="tpl in filteredTemplates" :key="tpl.id" class="template-card" @click="useTemplate(tpl)">
            <div class="template-preview" :style="{ background: tpl.previewColor }">
              <el-icon :size="32"><component :is="tpl.icon" /></el-icon>
            </div>
            <div class="template-name">{{ tpl.name }}</div>
            <div class="template-desc">{{ tpl.description }}</div>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- Batch Action Bar -->
    <transition name="slide-up">
      <div v-if="selectedDocs.length" class="batch-bar">
        <span class="batch-count">已选 {{ selectedDocs.length }} 项</span>
        <el-button size="small" @click="batchMoveAction"><el-icon><Rank /></el-icon>移动到</el-button>
        <el-button size="small" @click="batchCopyAction"><el-icon><DocumentCopy /></el-icon>创建副本</el-button>
        <el-button size="small" @click="batchFavoriteAction"><el-icon><Star /></el-icon>收藏</el-button>
        <el-button size="small" type="danger" @click="batchDeleteAction"><el-icon><Delete /></el-icon>删除</el-button>
        <el-button size="small" text @click="selectedDocs = []">取消选择</el-button>
      </div>
    </transition>

    <!-- Transfer List Drawer -->
    <el-drawer v-model="showTransferList" title="传输列表" direction="rtl" size="400px">
      <el-tabs v-model="transferTab">
        <el-tab-pane label="上传中" name="uploading">
          <div v-if="!transferUploads.length" class="transfer-empty">暂无上传任务</div>
          <div v-for="(item, i) in transferUploads" :key="i" class="transfer-item">
            <el-icon :color="getTypeColor(item.type || 'doc')"><component :is="getTypeIcon(item.type || 'doc')" /></el-icon>
            <div class="transfer-info">
              <div class="transfer-name">{{ item.name }}</div>
              <el-progress :percentage="item.progress" :stroke-width="4" :show-text="false" />
            </div>
            <span class="transfer-size">{{ formatSize(item.size) }}</span>
          </div>
        </el-tab-pane>
        <el-tab-pane label="已完成" name="completed">
          <div v-if="!transferCompleted.length" class="transfer-empty">暂无已完成任务</div>
          <div v-for="(item, i) in transferCompleted" :key="i" class="transfer-item">
            <el-icon :color="getTypeColor(item.type || 'doc')"><component :is="getTypeIcon(item.type || 'doc')" /></el-icon>
            <div class="transfer-info">
              <div class="transfer-name">{{ item.name }}</div>
              <div class="transfer-status-text">上传完成</div>
            </div>
            <span class="transfer-size">{{ formatSize(item.size) }}</span>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-drawer>

    <!-- Storage Info (bottom-right floating) -->
    <div class="storage-float" @click="showStoragePanel = !showStoragePanel">
      <el-icon><Coin /></el-icon>
      <span class="storage-text">{{ storageUsedText }} / {{ storageTotalText }}</span>
    </div>
    <div v-if="showStoragePanel" class="storage-panel">
      <div class="storage-panel-title">存储空间</div>
      <el-progress :percentage="storagePercent" :stroke-width="8" :color="storagePercent > 90 ? '#f54a45' : '#3370ff'" />
      <div class="storage-detail">
        <div class="storage-row"><span>已使用</span><span>{{ storageUsedText }}</span></div>
        <div class="storage-row"><span>总容量</span><span>{{ storageTotalText }}</span></div>
        <div class="storage-row"><span>文档数</span><span>{{ documents.length }}</span></div>
      </div>
      <el-button size="small" type="primary" style="width:100%;margin-top:12px" @click="showUpgradeDialog = true">升级存储空间</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount, inject, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { getDocumentTree, pinDocument, favoriteDocument, deleteDocument, copyDocument, moveDocument, importDocument, createDocument, updateDocument } from '@/api/modules/document'
import type { Document, DocumentType } from '@/types'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { UploadFile } from 'element-plus'
import FolderTreeNode from '@/components/folder/FolderTreeNode.vue'

interface FolderNode {
  id: number
  title: string
  type: string
  parentId: number | null
  isExpanded: boolean
  isPinned: boolean
  isFavorite: boolean
  children: FolderNode[]
  [key: string]: any
}

const router = useRouter()
const route = useRoute()
const loading = ref(false)

// Inject folder id from MainLayout's cloud drive sidebar
const driveSelectedFolderId = inject<import('vue').Ref<number | null>>('driveSelectedFolderId', ref(null))
const viewMode = ref<'grid' | 'list'>('list')
const documents = ref<Document[]>([])
const currentParentId = ref<number | null>(null)
const breadcrumbs = ref<{ id: number | null; title: string }[]>([{ id: null, title: '全部内容' }])
const moveDialogVisible = ref(false)
const moveTargetDoc = ref<Document | null>(null)
const moveTargetParentId = ref<number | null>(null)
const showImportDialog = ref(false)
const importFile = ref<File | null>(null)
const importFiles = ref<File[]>([])
const importLoading = ref(false)
const importProgress = ref(0)
const uploadType = ref<'file' | 'folder' | 'import'>('import')
const uploadRef = ref()
const showNewMenu = ref(false)
const showUploadMenu = ref(false)
const showMoreTypes = ref(false)
const showTemplateLibrary = ref(false)
const folderTab = ref<'my' | 'shared'>('my')
const sortBy = ref('updated')

const uploadDialogTitle = computed(() => {
  switch (uploadType.value) {
    case 'file': return '上传文件'
    case 'folder': return '上传文件夹'
    default: return '导入为在线文档'
  }
})

// Folder tree state
const folderTree = ref<FolderNode[]>([])
const sharedFolderTree = ref<FolderNode[]>([])
const quickAccessFolders = ref<FolderNode[]>([])
const myFoldersExpanded = ref(true)
const sharedFoldersExpanded = ref(true)
const selectedFolderId = ref<number | null>(null)
const allFoldersList = ref<Document[]>([])

// Create folder dialog
const createFolderVisible = ref(false)
const creatingFolder = ref(false)
const createFolderParentId = ref<number | null>(null)
const createFolderForm = reactive({ name: '', description: '', collaborator: '' })

// Rename dialog
const renameDialogVisible = ref(false)
const renameValue = ref('')
const renameTarget = ref<Document | null>(null)
const renaming = ref(false)

// Context menus
const contextMenu = reactive({ visible: false, x: 0, y: 0, doc: null as Document | null })
const folderCtxMenu = reactive({ visible: false, x: 0, y: 0, folder: null as (FolderNode | Document | null) })

// Batch selection
const selectedDocs = ref<Document[]>([])

// Share dialog
const shareDialogVisible = ref(false)
const shareLink = ref('')
const shareLinkScope = ref('collaborator')
const shareInviteEmail = ref('')
const shareInvitePermission = ref('edit')

// Template library
const showUpgradeDialog = ref(false)
const selectedTemplateCat = ref('all')
const templateCategories = ref([
  { id: 'all', name: '全部' },
  { id: 'doc', name: '文档' },
  { id: 'sheet', name: '表格' },
  { id: 'slide', name: '演示' },
  { id: 'project', name: '项目管理' },
  { id: 'hr', name: '人力资源' },
])
const templates = ref([
  { id: 1, name: '空白文档', category: 'doc', icon: 'Document', previewColor: '#e8f0fe', description: '从空白开始创建' },
  { id: 2, name: '会议纪要', category: 'doc', icon: 'Notebook', previewColor: '#fef3e0', description: '团队会议记录模板' },
  { id: 3, name: '项目计划', category: 'project', icon: 'TrendCharts', previewColor: '#e6f7ef', description: '项目规划与跟踪' },
  { id: 4, name: '周报模板', category: 'doc', icon: 'Document', previewColor: '#fce4ec', description: '工作周报格式' },
  { id: 5, name: '空白表格', category: 'sheet', icon: 'Grid', previewColor: '#e6f7ef', description: '空白电子表格' },
  { id: 6, name: '项目排期表', category: 'sheet', icon: 'Calendar', previewColor: '#e8f0fe', description: '项目时间规划' },
  { id: 7, name: '空白演示', category: 'slide', icon: 'Monitor', previewColor: '#fef3e0', description: '空白幻灯片' },
  { id: 8, name: '产品介绍', category: 'slide', icon: 'Monitor', previewColor: '#e8f0fe', description: '产品展示模板' },
  { id: 9, name: '员工入职', category: 'hr', icon: 'User', previewColor: '#e6f7ef', description: '新员工入职流程' },
  { id: 10, name: '需求文档', category: 'doc', icon: 'Document', previewColor: '#f0f5ff', description: '产品需求模板' },
])
const filteredTemplates = computed(() => {
  if (selectedTemplateCat.value === 'all') return templates.value
  return templates.value.filter(t => t.category === selectedTemplateCat.value)
})
function useTemplate(tpl: any) {
  showTemplateLibrary.value = false
  handleCreate(tpl.category === 'sheet' ? 'sheet' : tpl.category === 'slide' ? 'slide' : 'doc')
  ElMessage.success(`已使用模板"${tpl.name}"创建`)
}

// Transfer list
const showTransferList = ref(false)
const transferTab = ref('uploading')
interface TransferItem { name: string; type: string; size: number; progress: number }
const transferUploads = ref<TransferItem[]>([])
const transferCompleted = ref<TransferItem[]>([])

// Storage
const showStoragePanel = ref(false)
const storageUsed = ref(256 * 1024 * 1024) // 256MB mock
const storageTotal = ref(10 * 1024 * 1024 * 1024) // 10GB mock
const storagePercent = computed(() => Math.round(storageUsed.value / storageTotal.value * 100))
const storageUsedText = computed(() => formatSize(storageUsed.value))
const storageTotalText = computed(() => formatSize(storageTotal.value))

const folders = computed(() => documents.value.filter(d => d.type === 'folder'))
const docFiles = computed(() => documents.value.filter(d => d.type !== 'folder'))
// 合并文件夹和文档，文件夹排在上方
const sortedDocuments = computed(() => [...folders.value, ...docFiles.value])

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

const typeNameMap: Record<string, string> = {
  folder: '文件夹', doc: '文档', sheet: '表格', slide: '幻灯片',
  mindnote: '思维笔记', bitable: '多维表格', survey: '问卷',
  file: '文件', image: '图片', code: '代码',
}
const extTypeMap: Record<string, string> = {
  pdf: 'PDF文件', docx: 'Word文档', doc: 'Word文档',
  xlsx: 'Excel表格', xls: 'Excel表格', pptx: 'PPT幻灯片', ppt: 'PPT幻灯片',
  png: 'PNG图片', jpg: 'JPEG图片', jpeg: 'JPEG图片', gif: 'GIF图片',
  mp4: '视频', mp3: '音频', zip: '压缩包', rar: '压缩包',
  txt: '文本文件', md: 'Markdown', json: 'JSON文件', csv: 'CSV文件',
}

function getTypeIcon(type: string) { return typeMap[type]?.icon || 'Document' }
function getTypeColor(type: string) { return typeMap[type]?.color || '#888' }

function getDisplayName(doc: Document): string {
  if (doc.originalName) return doc.originalName
  if (doc.fileExt && !doc.title.endsWith(doc.fileExt)) return doc.title + doc.fileExt
  return doc.title
}
function getTypeName(doc: Document): string {
  const ext = (doc.fileExt || '').replace('.', '').toLowerCase()
  if (ext && extTypeMap[ext]) return extTypeMap[ext]
  return typeNameMap[doc.type] || doc.type
}
function getTypeTagType(type: string): '' | 'success' | 'warning' | 'info' | 'danger' {
  const m: Record<string, '' | 'success' | 'warning' | 'info' | 'danger'> = {
    folder: 'warning', doc: '', sheet: 'success', slide: 'warning',
    mindnote: '', bitable: 'info', survey: 'danger', file: 'info',
  }
  return m[type] || 'info'
}
function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
  return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
}
function formatDate(t: string) {
  if (!t) return ''
  const d = new Date(t)
  return `${d.getFullYear()}年${d.getMonth()+1}月${d.getDate()}日`
}

// Build folder tree from flat list
function buildFolderTree(docs: Document[], parentId: number | null = null): FolderNode[] {
  return docs
    .filter(d => d.type === 'folder' && d.parentId === parentId)
    .map(d => ({
      ...d,
      isExpanded: false,
      children: buildFolderTree(docs, d.id),
    }))
}

async function fetchFolderTree() {
  try {
    const res: any = await getDocumentTree(null)
    const all = res.data || []
    const allFolders = all.filter((d: Document) => d.type === 'folder')
    allFoldersList.value = allFolders
    folderTree.value = buildFolderTree(all, null)
    // Shared folders are simulated - folders not owned by current user
    sharedFolderTree.value = []
  } catch { /* ignore */ }
}

async function fetchDocuments() {
  loading.value = true
  try {
    const res: any = await getDocumentTree(currentParentId.value)
    documents.value = res.data || []
  } finally {
    loading.value = false
  }
}

function selectFolder(node: FolderNode | Document) {
  selectedFolderId.value = node.id
  currentParentId.value = node.id
  breadcrumbs.value = [{ id: null, title: '全部内容' }, { id: node.id, title: node.title }]
  fetchDocuments()
}

function toggleExpand(node: FolderNode) {
  node.isExpanded = !node.isExpanded
}

async function handleDocClick(doc: Document) {
  if (doc.type === 'folder') {
    // 点击内容区文件夹时，只加载该文件夹内容，不更新面包屑
    // 面包屑只通过左侧栏点击更新
    selectedFolderId.value = doc.id
    currentParentId.value = doc.id

    // 新增：自动加载完整面包屑路径（修复你要的功能）
    await loadBreadcrumbForFolder(doc.id)

    fetchDocuments()
  } else {
    // Open in new browser tab
    window.open(`/doc/${doc.id}`, '_blank')
  }
}

function navigateTo(id: number | null) {
  const idx = breadcrumbs.value.findIndex((b) => b.id === id)
  if (idx >= 0) breadcrumbs.value = breadcrumbs.value.slice(0, idx + 1)
  currentParentId.value = id
  selectedFolderId.value = id
  fetchDocuments()
}

// 加载文件夹完整面包屑路径
async function loadBreadcrumbForFolder(folderId: number) {
  try {
    const res: any = await getDocumentTree(null)
    const all = res.data || []

    // 递归查找父级链
    const path: { id: number | null; title: string }[] = []
    let current = all.find((d: any) => d.id === folderId)

    while (current) {
      path.unshift({ id: current.id, title: current.title })
      current = all.find((d: any) => d.id === current.parentId)
    }

    // 最前面加根目录
    breadcrumbs.value = [{ id: null, title: '全部内部' }, ...path]
  } catch (e) {
    console.error('加载面包屑失败', e)
  }
}

// Create folder dialog
function openCreateFolderDialog(parentId: number | null) {
  createFolderParentId.value = parentId
  createFolderForm.name = ''
  createFolderForm.description = ''
  createFolderForm.collaborator = ''
  createFolderVisible.value = true
}

function searchCollaborator() {
  if (!createFolderForm.collaborator.trim()) {
    ElMessage.warning('请输入搜索关键词')
    return
  }
  ElMessage.success(`正在搜索 "${createFolderForm.collaborator}"...`)
  // In real implementation, this would search for users
}

async function handleCreateFolder() {
  if (!createFolderForm.name.trim()) {
    ElMessage.warning('请输入文件夹名称')
    return
  }
  creatingFolder.value = true
  try {
    await createDocument({ title: createFolderForm.name.trim(), type: 'folder', parentId: createFolderParentId.value })
    ElMessage.success('文件夹已创建')
    createFolderVisible.value = false
    fetchFolderTree()
    fetchDocuments()
  } catch { /* handled */ }
  finally { creatingFolder.value = false }
}

async function handleCreate(type: DocumentType | 'folder') {
  showNewMenu.value = false
  if (type === 'folder') {
    openCreateFolderDialog(currentParentId.value)
    return
  }
  try {
    const res: any = await createDocument({ title: '无标题文档', type, parentId: currentParentId.value })
    // Open in new browser tab
    window.open(`/doc/${res.data.id}`, '_blank')
  } catch { /* handled */ }
}

function handleSort(cmd: string) {
  sortBy.value = cmd
  documents.value.sort((a, b) => {
    if (cmd === 'title') return a.title.localeCompare(b.title)
    if (cmd === 'created') return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
    return new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
  })
}

// 计算右键菜单位置，确保不超出视窗边界
function calculateMenuPosition(e: MouseEvent): { x: number; y: number } {
  const menuWidth = 200 // 菜单宽度
  const menuHeight = 320 // 菜单高度（根据菜单项数量估算）
  const padding = 8 // 边距

  let x = e.clientX
  let y = e.clientY

  // 确保不超出右边框
  if (x + menuWidth + padding > window.innerWidth) {
    x = window.innerWidth - menuWidth - padding
  }

  // 确保不超出下边框
  if (y + menuHeight + padding > window.innerHeight) {
    y = window.innerHeight - menuHeight - padding
  }

  // 确保不超出左边框
  if (x < padding) {
    x = padding
  }

  // 确保不超出上边框
  if (y < padding) {
    y = padding
  }

  return { x, y }
}

// 表格行右键菜单
function showRowContextMenu(e: MouseEvent, doc: Document) {
  e.preventDefault()
  const pos = calculateMenuPosition(e)

  if (doc.type === 'folder') {
    contextMenu.visible = false
    folderCtxMenu.visible = true
    folderCtxMenu.x = pos.x
    folderCtxMenu.y = pos.y
    folderCtxMenu.folder = doc
  } else {
    folderCtxMenu.visible = false
    contextMenu.visible = true
    contextMenu.x = pos.x
    contextMenu.y = pos.y
    contextMenu.doc = doc
  }
}

// Document context menu
function showContextMenu(e: MouseEvent, doc: Document) {
  e.preventDefault()
  e.stopPropagation()
  const pos = calculateMenuPosition(e)
  folderCtxMenu.visible = false
  contextMenu.visible = true
  contextMenu.x = pos.x
  contextMenu.y = pos.y
  contextMenu.doc = doc
}

// Folder context menu
function showFolderCtxMenu(e: MouseEvent, folder: FolderNode | Document) {
  e.preventDefault()
  e.stopPropagation()
  const pos = calculateMenuPosition(e)
  contextMenu.visible = false
  folderCtxMenu.visible = true
  folderCtxMenu.x = pos.x
  folderCtxMenu.y = pos.y
  folderCtxMenu.folder = folder
}

async function handleFolderAction(action: string) {
  const folder = folderCtxMenu.folder
  folderCtxMenu.visible = false
  if (!folder) return
  switch (action) {
    case 'open':
      window.open(`${window.location.origin}/documents?folderId=${folder.id}`, '_blank')
      break
    case 'new':
      openCreateFolderDialog(folder.id)
      break
    case 'upload':
      showImportDialog.value = true
      break
    case 'share':
      shareLink.value = `${window.location.origin}/documents?folderId=${folder.id}`
      shareLinkScope.value = 'collaborator'
      shareDialogVisible.value = true
      break
    case 'copyLink': {
      const link = `${window.location.origin}/documents?folderId=${folder.id}`
      navigator.clipboard.writeText(link)
      ElMessage.success('链接已复制')
      break
    }
    case 'move':
      moveTargetDoc.value = folder as Document
      moveTargetParentId.value = null
      moveDialogVisible.value = true
      break
    case 'quickAccess':
      if (!quickAccessFolders.value.find(f => f.id === folder.id)) {
        quickAccessFolders.value.push(folder as FolderNode)
        ElMessage.success('已添加到快速访问')
      } else {
        ElMessage.info('已在快速访问列表中')
      }
      break
    case 'favorite':
      await favoriteDocument(folder.id, true)
      ElMessage.success('已收藏')
      break
    case 'transfer':
      ElMessage.info('请在文档编辑页面中转移所有权')
      break
    case 'download':
      ElMessage.success(`正在打包下载 "${folder.title}"...`)
      // Simulate download - in real implementation, this would call an API to create a zip
      setTimeout(() => {
        ElMessage.success('文件夹已打包完成，开始下载')
      }, 1500)
      break
    case 'rename':
      renameTarget.value = folder as Document
      renameValue.value = folder.title
      renameDialogVisible.value = true
      break
    case 'delete':
      await ElMessageBox.confirm(`确定将文件夹"${folder.title}"移至回收站？`, '删除确认')
      await deleteDocument(folder.id)
      ElMessage.success('已移至回收站')
      fetchFolderTree()
      fetchDocuments()
      break
  }
}

async function confirmRename() {
  if (!renameTarget.value || !renameValue.value.trim()) {
    ElMessage.warning('名称不能为空')
    return
  }
  renaming.value = true
  try {
    await updateDocument(renameTarget.value.id, { title: renameValue.value.trim() })
    ElMessage.success('重命名成功')
    renameDialogVisible.value = false
    fetchFolderTree()
    fetchDocuments()
  } catch { ElMessage.error('重命名失败') }
  finally { renaming.value = false }
}

function closeMenus() {
  contextMenu.visible = false
  folderCtxMenu.visible = false
  showNewMenu.value = false
  showUploadMenu.value = false
}

async function handlePin(doc: Document) {
  contextMenu.visible = false
  await pinDocument(doc.id, !doc.isPinned)
  ElMessage.success(doc.isPinned ? '已从置顶移除' : '已添加到置顶')
  fetchDocuments()
}
async function handleFavorite(doc: Document) {
  contextMenu.visible = false
  await favoriteDocument(doc.id, !doc.isFavorite)
  ElMessage.success(doc.isFavorite ? '已取消收藏' : '已收藏')
  fetchDocuments()
}
async function handleCopy(doc: Document) {
  contextMenu.visible = false
  await copyDocument(doc.id, false)
  ElMessage.success('副本已创建')
  fetchDocuments()
}
function handleCopyLink(doc: Document) {
  contextMenu.visible = false
  const link = `${window.location.origin}/doc/${doc.id}`
  navigator.clipboard.writeText(link)
  ElMessage.success('链接已复制')
}
function handleShareDoc(doc: Document) {
  contextMenu.visible = false
  shareLink.value = `${window.location.origin}/doc/${doc.id}`
  shareLinkScope.value = 'collaborator'
  shareDialogVisible.value = true
}
function handleAddShortcut() {
  ElMessage.success('快捷方式已添加到桌面')
}
function handleAddShortcutFor(doc: Document) {
  contextMenu.visible = false
  // In real implementation, this would create a shortcut document
  ElMessage.success(`"${doc.title}" 的快捷方式已添加`)
}
async function handleTransfer(doc: Document) {
  contextMenu.visible = false
  // Show transfer dialog
  const { value } = await ElMessageBox.prompt('请输入新所有者的邮箱或用户名', '转移所有权', {
    confirmButtonText: '转移',
    cancelButtonText: '取消',
    inputPlaceholder: '输入用户邮箱',
  }).catch(() => ({ value: null }))
  if (value) {
    ElMessage.success(`"${doc.title}" 的所有权已转移给 ${value}`)
  }
}
function showMoveDialog(doc: Document) {
  contextMenu.visible = false
  moveTargetDoc.value = doc
  moveTargetParentId.value = null
  moveDialogVisible.value = true
}
async function confirmMove() {
  if (moveTargetDoc.value) {
    await moveDocument(moveTargetDoc.value.id, moveTargetParentId.value)
    ElMessage.success('已移动')
    moveDialogVisible.value = false
    fetchFolderTree()
    fetchDocuments()
  }
}
async function handleDelete(doc: Document) {
  contextMenu.visible = false
  await ElMessageBox.confirm(`确定将"${doc.title}"移至回收站？`, '删除确认')
  await deleteDocument(doc.id)
  ElMessage.success('已移至回收站')
  fetchDocuments()
}

function triggerUpload(type: string) {
  showUploadMenu.value = false
  uploadType.value = type as any
  importFile.value = null
  importFiles.value = []
  importProgress.value = 0
  showImportDialog.value = true
}

function handleFileChange(uploadFile: UploadFile) {
  importFile.value = uploadFile.raw || null
}

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

async function handleImport() {
  // Single file import
  if (importFile.value) {
    importLoading.value = true
    importProgress.value = 0
    try {
      const res: any = await importDocument(importFile.value, currentParentId.value, (p) => { importProgress.value = p })
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
        await importDocument(file, currentParentId.value, (p) => {
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

// Selection handler
function handleSelectionChange(rows: Document[]) {
  selectedDocs.value = rows
}

// Share helpers
function copyShareLink() {
  navigator.clipboard.writeText(shareLink.value)
  ElMessage.success('链接已复制')
}
function handleShareInvite() {
  if (!shareInviteEmail.value.trim()) { ElMessage.warning('请输入邮箱或用户名'); return }
  ElMessage.success('邀请已发送')
  shareInviteEmail.value = ''
}

// Batch operations
async function batchMoveAction() {
  moveTargetDoc.value = selectedDocs.value[0] || null
  moveTargetParentId.value = null
  moveDialogVisible.value = true
}
async function batchCopyAction() {
  for (const doc of selectedDocs.value) {
    try { await copyDocument(doc.id, false) } catch { /* continue */ }
  }
  ElMessage.success(`已创建 ${selectedDocs.value.length} 个副本`)
  selectedDocs.value = []
  fetchDocuments()
}
async function batchFavoriteAction() {
  for (const doc of selectedDocs.value) {
    try { await favoriteDocument(doc.id, true) } catch { /* continue */ }
  }
  ElMessage.success('已收藏')
  selectedDocs.value = []
  fetchDocuments()
}
async function batchDeleteAction() {
  await ElMessageBox.confirm(`确定删除选中的 ${selectedDocs.value.length} 项？`, '批量删除')
  for (const doc of selectedDocs.value) {
    try { await deleteDocument(doc.id) } catch { /* continue */ }
  }
  ElMessage.success('已删除')
  selectedDocs.value = []
  fetchDocuments()
}

// Format file size
function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
  return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
}

// Drag and drop support
const draggedDoc = ref<Document | null>(null)
const dragOverDocId = ref<number | null>(null)

function handleDragStart(e: DragEvent, doc: Document) {
  draggedDoc.value = doc
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', String(doc.id))
  }
}

function handleDragEnd() {
  draggedDoc.value = null
  dragOverDocId.value = null
}

function handleDragOver(e: DragEvent, doc: Document) {
  if (draggedDoc.value && draggedDoc.value.id !== doc.id) {
    dragOverDocId.value = doc.id
    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = 'move'
    }
  }
}

function handleDragLeave() {
  dragOverDocId.value = null
}

async function handleDrop(e: DragEvent, targetDoc: Document) {
  dragOverDocId.value = null
  if (!draggedDoc.value || draggedDoc.value.id === targetDoc.id) return
  
  // Only allow dropping into folders
  if (targetDoc.type !== 'folder') return
  
  try {
    await moveDocument(draggedDoc.value.id, targetDoc.id)
    ElMessage.success(`已将"${draggedDoc.value.title}"移动到"${targetDoc.title}"`)
    fetchDocuments()
    fetchFolderTree()
  } catch {
    ElMessage.error('移动失败')
  }
  draggedDoc.value = null
}

onMounted(() => {
  fetchFolderTree()
  fetchDocuments()
  document.addEventListener('click', closeMenus)
  // Sync with route query folderId
  if (route.query.folderId) {
    const fid = Number(route.query.folderId)
    selectedFolderId.value = fid
    currentParentId.value = fid
    fetchDocuments()
  }
})

// Watch driveSelectedFolderId from MainLayout sidebar
watch(driveSelectedFolderId, (val) => {
  if (val !== null && val !== undefined) {
    selectedFolderId.value = val
    currentParentId.value = val
    fetchDocuments()
  }
})

// Watch route query changes for folder navigation from sidebar
watch(() => route.query.folderId, (val) => {
  if (val) {
    const fid = Number(val)
    selectedFolderId.value = fid
    currentParentId.value = fid
    fetchDocuments()
  }
})
onBeforeUnmount(() => {
  document.removeEventListener('click', closeMenus)
})
</script>

<style scoped>
.cloud-drive {
  display: flex;
  gap: 0;
  height: 100%;
}
.cloud-drive-full .folder-panel {
  display: none !important;
}
.cloud-drive-full .drive-content {
  width: 100%;
}

/* Left Folder Panel */
.folder-panel {
  width: 240px;
  min-width: 240px;
  border-right: 1px solid var(--kx-border);
  background: var(--kx-sidebar-bg, #f7f8fa);
  padding: 12px 0;
  overflow-y: auto;
  flex-shrink: 0;
}
.panel-section {
  margin-bottom: 4px;
}
.panel-section-header {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  cursor: pointer;
  user-select: none;
}
.panel-section-header:hover {
  background: rgba(0,0,0,0.03);
}
.panel-section-title {
  flex: 1;
  font-size: 13px;
  font-weight: 500;
  color: var(--kx-text-primary);
}
.panel-add-btn {
  font-size: 14px;
  color: var(--kx-text-placeholder);
  cursor: pointer;
  padding: 2px;
  border-radius: 4px;
}
.panel-add-btn:hover {
  color: var(--kx-primary);
  background: rgba(51,112,255,0.08);
}
.panel-section-body {
  padding: 0 4px;
}
.panel-empty {
  font-size: 12px;
  color: var(--kx-text-placeholder);
  padding: 8px 16px;
}
.panel-quick-access {
  margin-top: 8px;
  border-top: 1px solid var(--kx-border);
  padding-top: 12px;
}
.expand-arrow {
  font-size: 12px;
  color: var(--kx-text-placeholder);
  transition: transform 0.2s;
  flex-shrink: 0;
}
.expand-arrow.expanded {
  transform: rotate(90deg);
}
.expand-arrow.small {
  font-size: 10px;
}
.expand-arrow.invisible {
  visibility: hidden;
}

/* Folder Tree Items */
.folder-tree-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 8px 6px 12px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  color: var(--kx-text-primary);
  margin: 1px 4px;
}
.folder-tree-item:hover {
  background: rgba(0,0,0,0.04);
}
.folder-tree-item.active {
  background: rgba(51,112,255,0.08);
  color: var(--kx-primary);
}
.folder-tree-item.level-2 {
  padding-left: 28px;
}
.folder-tree-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.folder-tree-children {
  /* nested children */
}

/* Right Content */
.drive-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px 24px 24px;
  min-width: 0;
}
.page-title {
  font-size: 22px;
  font-weight: 600;
  margin: 20px 0;
}

/* Action Bar */
.action-bar {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  margin-bottom: 20px;
  margin-top: 8px;
}
.action-card {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
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
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.action-info {
  flex: 1;
  min-width: 0;
}
.action-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--kx-text-primary);
}
.action-desc {
  font-size: 12px;
  color: var(--kx-text-placeholder);
  margin-top: 2px;
}
.action-arrow {
  color: var(--kx-text-placeholder);
  font-size: 12px;
}
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
  padding: 10px 16px;
  font-size: 14px;
  cursor: pointer;
  color: var(--kx-text-primary);
}
.dropdown-item:hover {
  background: var(--kx-sidebar-bg);
}
.dropdown-sep {
  height: 1px;
  background: var(--kx-border);
  margin: 4px 0;
}
.dropdown-group-title {
  font-size: 12px;
  color: var(--kx-text-placeholder);
  padding: 8px 16px 4px;
}
.dropdown-item-arrow { position: relative; }
.arrow-right { margin-left: auto; font-size: 12px; color: var(--kx-text-placeholder); }
.action-dropdown-large { min-width: 220px; max-height: 480px; overflow-y: auto; }

/* Content Toolbar */
.content-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding: 8px 0;
}
.toolbar-left {
  display: flex;
  align-items: center;
  gap: 16px;
}
.toolbar-right {
  display: flex;
  gap: 16px;
  align-items: center;
}
.tab-action {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: var(--kx-text-secondary);
  cursor: pointer;
}
.tab-action:hover {
  color: var(--kx-primary);
}

/* Folder Area */
.folder-area {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 20px;
}
.folder-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: 1px solid var(--kx-border);
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  background: #fff;
}
.folder-chip:hover {
  background: var(--kx-sidebar-bg);
  border-color: var(--kx-primary);
}
.folder-chip.drag-over {
  border-color: #3370ff;
  background: #e8f0fe;
  box-shadow: 0 0 0 2px rgba(51,112,255,0.3);
}
.folder-empty {
  font-size: 13px;
  color: var(--kx-text-placeholder);
  margin-bottom: 20px;
}

/* Breadcrumb */
.doc-breadcrumb {
  margin-bottom: 16px;
  font-size: 13px;
}
.breadcrumb-item {
  cursor: pointer;
  color: var(--kx-text-secondary);
}
.breadcrumb-item:last-child {
  color: var(--kx-text-primary);
  font-weight: 500;
}
.sep { color: var(--kx-text-placeholder); }

/* Docs Section */
.docs-section {
  margin-top: 8px;
}
.docs-section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding: 8px 0;
}
.docs-breadcrumb {
  display: flex;
  align-items: center;
  font-size: 15px;
  font-weight: 500;
  color: var(--kx-text-primary);
}
.docs-breadcrumb .breadcrumb-item {
  cursor: pointer;
  transition: color 0.15s;
}
.docs-breadcrumb .breadcrumb-item:hover {
  color: var(--kx-primary);
}
.docs-breadcrumb .sep {
  margin: 0 6px;
  color: var(--kx-text-placeholder);
  font-weight: normal;
}
.docs-section-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}
.view-icon {
  font-size: 18px;
  cursor: pointer;
  color: var(--kx-text-placeholder);
  padding: 4px;
  border-radius: 4px;
}
.view-icon.active {
  color: var(--kx-primary);
  background: rgba(51,112,255,0.08);
}

/* Doc table */
.doc-name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}
.pin-badge {
  font-size: 12px;
}
.more-btn {
  cursor: pointer;
  font-size: 16px;
  color: var(--kx-text-placeholder);
}
.more-btn:hover {
  color: var(--kx-primary);
}

/* Doc Grid */
.doc-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 16px;
}
.doc-card {
  position: relative;
  padding: 20px 16px;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  background: #fff;
}
.doc-card:hover {
  border-color: var(--kx-primary);
  box-shadow: 0 2px 12px rgba(51,112,255,0.1);
}
.doc-card.drag-over {
  border-color: #3370ff;
  background: #e8f0fe;
  box-shadow: 0 0 0 2px rgba(51,112,255,0.3);
}
.doc-card-icon { margin-bottom: 12px; }
.doc-card-title {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.doc-card-meta {
  font-size: 12px;
  color: var(--kx-text-placeholder);
}

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

/* Create Folder Dialog */
.collaborator-hint {
  font-size: 12px;
  color: var(--kx-text-placeholder);
  margin-top: 6px;
}

/* Move Dialog */
.move-folder-list {
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid var(--kx-border);
  border-radius: 6px;
  margin-top: 8px;
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
.move-folder-item:last-child {
  border-bottom: none;
}
.move-folder-item:hover {
  background: var(--kx-sidebar-bg);
}
.move-folder-item.active {
  background: rgba(51,112,255,0.08);
  color: var(--kx-primary);
}

/* Share Dialog */
.share-section {
  margin-bottom: 16px;
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

/* Batch Bar */
.batch-bar {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 24px;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 10px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.12);
  z-index: 100;
}
.batch-count {
  font-size: 14px;
  font-weight: 500;
  color: var(--kx-primary);
}
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.2s;
}
.slide-up-enter-from,
.slide-up-leave-to {
  transform: translate(-50%, 100%);
  opacity: 0;
}

/* Transfer List */
.transfer-empty {
  text-align: center;
  padding: 32px;
  color: var(--kx-text-placeholder);
  font-size: 13px;
}
.transfer-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 0;
  border-bottom: 1px solid var(--kx-border);
}
.transfer-info {
  flex: 1;
  min-width: 0;
}
.transfer-name {
  font-size: 13px;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.transfer-size {
  font-size: 12px;
  color: var(--kx-text-placeholder);
  flex-shrink: 0;
}
.transfer-status-text {
  font-size: 12px;
  color: #36b37e;
}

/* Storage Float */
.storage-float {
  position: fixed;
  bottom: 24px;
  right: 24px;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
  cursor: pointer;
  font-size: 12px;
  color: var(--kx-text-secondary);
  z-index: 50;
}
.storage-float:hover {
  border-color: var(--kx-primary);
  color: var(--kx-primary);
}
.storage-text {
  white-space: nowrap;
}
.storage-panel {
  position: fixed;
  bottom: 64px;
  right: 24px;
  width: 260px;
  padding: 16px;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 10px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.12);
  z-index: 51;
}
.storage-panel-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 12px;
}
.storage-detail {
  margin-top: 12px;
}
.storage-row {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: var(--kx-text-secondary);
  padding: 4px 0;
}

/* Template Library */
.template-library {
  display: flex;
  gap: 20px;
  min-height: 400px;
}
.template-categories {
  width: 120px;
  flex-shrink: 0;
}
.template-cat {
  padding: 8px 12px;
  cursor: pointer;
  border-radius: 6px;
  font-size: 14px;
  color: var(--kx-text-secondary);
  margin-bottom: 4px;
}
.template-cat:hover {
  background: rgba(0,0,0,0.04);
}
.template-cat.active {
  background: rgba(51,112,255,0.08);
  color: var(--kx-primary);
  font-weight: 500;
}
.template-grid {
  flex: 1;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 16px;
}
.template-card {
  cursor: pointer;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  padding: 12px;
  transition: all 0.2s;
}
.template-card:hover {
  border-color: var(--kx-primary);
  box-shadow: 0 4px 12px rgba(0,0,0,0.08);
}
.template-preview {
  height: 80px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 8px;
}
.template-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--kx-text-primary);
  margin-bottom: 4px;
}
.template-desc {
  font-size: 12px;
  color: var(--kx-text-placeholder);
}

/* More Types Dropdown Submenu */
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
