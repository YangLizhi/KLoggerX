<template>
  <div class="main-layout">
    <aside class="sidebar" :class="{ collapsed: sidebarCollapsed && !isCloudDrivePage && !isAdminPage && !isKnowledgePage }">
      <!-- === Admin Sub-Navigation Mode === -->
      <template v-if="isAdminPage">
        <div class="sidebar-header drive-header">
          <el-icon class="back-btn" @click="$router.push('/home')"><ArrowLeft /></el-icon>
          <div class="drive-header-info" @click="$router.push('/home')">
            <el-icon color="#646a73"><Setting /></el-icon>
            <span class="drive-header-title">系统管理</span>
          </div>
        </div>
        <div class="sidebar-nav" style="margin-top: 8px;">
          <div class="nav-item" :class="{ active: route.path === '/admin/users' }" @click="$router.push('/admin/users')">
            <el-icon :size="18"><User /></el-icon>
            <span class="nav-label">用户和权限</span>
          </div>
          <div class="nav-item" :class="{ active: route.path === '/admin/departments' }" @click="$router.push('/admin/departments')">
            <el-icon :size="18"><OfficeBuilding /></el-icon>
            <span class="nav-label">部门管理</span>
          </div>
          <div class="nav-item" :class="{ active: route.path === '/admin/ai-models' }" @click="$router.push('/admin/ai-models')">
            <el-icon :size="18"><Cpu /></el-icon>
            <span class="nav-label">AI模型设置</span>
          </div>
          <div class="nav-item" :class="{ active: route.path === '/admin/storage' }" @click="$router.push('/admin/storage')">
            <el-icon :size="18"><FolderOpened /></el-icon>
            <span class="nav-label">云盘存储</span>
          </div>
        </div>
      </template>

      <!-- === Knowledge Base Sub-Navigation Mode === -->
      <template v-else-if="isKnowledgePage">
        <div class="sidebar-header drive-header">
          <el-icon class="back-btn" @click="$router.push('/home')"><ArrowLeft /></el-icon>
          <div class="drive-header-info" @click="$router.push('/knowledge')">
            <el-icon color="#3370ff"><Collection /></el-icon>
            <span class="drive-header-title">知识库</span>
          </div>
        </div>

        <!-- Search -->
        <div class="sidebar-search">
          <el-input v-model="kbSearch" placeholder="搜索" prefix-icon="Search" size="small" clearable />
        </div>

        <!-- Navigation -->
        <div class="sidebar-nav" style="margin-top: 0;">
          <div class="nav-item" :class="{ active: route.path === '/knowledge' && !route.params.id }" @click="$router.push('/knowledge')">
            <el-icon :size="18"><HomeFilled /></el-icon>
            <span class="nav-label">首页</span>
          </div>
        </div>

        <div class="drive-tree-sections">
          <!-- Personal Knowledge Base -->
          <div class="drive-tree-section">
            <div class="drive-section-header" @click="kbPersonalExpanded = !kbPersonalExpanded">
              <el-icon class="drive-expand-arrow" :class="{ expanded: kbPersonalExpanded }"><ArrowRight /></el-icon>
              <el-icon color="#3370ff" :size="15"><User /></el-icon>
              <span class="drive-section-title">个人知识库</span>
              <el-tag size="small" type="info" style="margin-left: auto; cursor:pointer" @click.stop="$router.push('/knowledge/list')">全部</el-tag>
            </div>
            <div v-show="kbPersonalExpanded" class="drive-section-body">
              <div v-if="!filteredKbList.length" class="drive-tree-empty">暂无知识库</div>
              <div
                v-for="kb in filteredKbList"
                :key="kb.id"
                class="kb-tree-item"
                :class="{ active: route.params.id && Number(route.params.id) === kb.id }"
                @click="$router.push(`/knowledge/${kb.id}`)"
                @contextmenu.prevent="showKbTreeCtx($event, kb)"
              >
                <el-icon class="kb-tree-icon" :color="kbCoverColors[kb.id % kbCoverColors.length]"><Collection /></el-icon>
                <span class="kb-tree-name">{{ kb.name }}</span>
                <span class="kb-tree-actions" @click.stop="showKbTreeCtx($event, kb)">
                  <el-icon :size="14"><MoreFilled /></el-icon>
                </span>
              </div>
            </div>
          </div>

          <!-- Shared Knowledge Base -->
          <div class="drive-tree-section">
            <div class="drive-section-header" @click="kbSharedExpanded = !kbSharedExpanded">
              <el-icon class="drive-expand-arrow" :class="{ expanded: kbSharedExpanded }"><ArrowRight /></el-icon>
              <el-icon color="#36b37e" :size="15"><Share /></el-icon>
              <span class="drive-section-title">共享知识库</span>
              <el-tag size="small" type="info" style="margin-left: auto; cursor:pointer" @click.stop="$router.push('/knowledge/list')">全部</el-tag>
            </div>
            <div v-show="kbSharedExpanded" class="drive-section-body">
              <div class="drive-tree-empty">暂无共享知识库</div>
            </div>
          </div>

          <!-- Created by Me -->
          <div class="drive-tree-section">
            <div class="drive-section-header" @click="kbPinnedExpanded = !kbPinnedExpanded">
              <el-icon class="drive-expand-arrow" :class="{ expanded: kbPinnedExpanded }"><ArrowRight /></el-icon>
              <el-icon color="#f5a623" :size="15"><FolderOpened /></el-icon>
              <span class="drive-section-title">我创建的</span>
              <el-icon :size="14" class="section-act-btn" style="margin-left: auto;" @click.stop="handleNewKb"><Plus /></el-icon>
            </div>
            <div v-show="kbPinnedExpanded" class="drive-section-body">
              <div v-if="!filteredKbList.length" class="drive-tree-empty">暂无创建的知识库</div>
              <div
                v-for="kb in filteredKbList"
                :key="'cr-' + kb.id"
                class="kb-tree-item"
                :class="{ active: route.params.id && Number(route.params.id) === kb.id }"
                @click="$router.push(`/knowledge/${kb.id}`)"
                @contextmenu.prevent="showKbTreeCtx($event, kb)"
              >
                <el-icon class="kb-tree-icon" :color="kbCoverColors[kb.id % kbCoverColors.length]"><Collection /></el-icon>
                <span class="kb-tree-name">{{ kb.name }}</span>
                <span class="kb-tree-actions" @click.stop="showKbTreeCtx($event, kb)">
                  <el-icon :size="14"><MoreFilled /></el-icon>
                </span>
              </div>
            </div>
          </div>

          <!-- Joined -->
          <div class="drive-tree-section">
            <div class="drive-section-header" @click="kbRecentExpanded = !kbRecentExpanded">
              <el-icon class="drive-expand-arrow" :class="{ expanded: kbRecentExpanded }"><ArrowRight /></el-icon>
              <el-icon color="#9254de" :size="15"><UserFilled /></el-icon>
              <span class="drive-section-title">我加入的</span>
            </div>
            <div v-show="kbRecentExpanded" class="drive-section-body">
              <div class="drive-tree-empty">暂无加入的知识库</div>
            </div>
          </div>
        </div>

        <!-- New Knowledge Base button -->
        <div class="drive-bottom-action" @click="handleNewKb">
          <el-icon><Plus /></el-icon>
          <span>新建知识库</span>
        </div>

        <!-- Knowledge Base Context Menu -->
        <div v-if="kbTreeCtx.visible" class="context-menu" :style="{ left: kbTreeCtx.x + 'px', top: kbTreeCtx.y + 'px' }">
          <div class="ctx-item" @click="handleKbTreeAction('open')"><el-icon :size="13"><View /></el-icon>打开</div>
          <div class="ctx-item" @click="handleKbTreeAction('openNew')"><el-icon :size="13"><TopRight /></el-icon>在新标签页打开</div>
          <div class="ctx-sep" />
          <div class="ctx-item" @click="handleKbTreeAction('share')"><el-icon :size="13"><Share /></el-icon>分享</div>
          <div class="ctx-item" @click="handleKbTreeAction('copyLink')"><el-icon :size="13"><Link /></el-icon>复制链接</div>
          <div class="ctx-sep" />
          <div class="ctx-item" @click="handleKbTreeAction('settings')"><el-icon :size="13"><Setting /></el-icon>设置</div>
          <div class="ctx-item danger" @click="handleKbTreeAction('delete')"><el-icon :size="13"><Delete /></el-icon>删除</div>
        </div>
      </template>

      <!-- === Cloud Drive Sub-Navigation Mode === -->
      <template v-else-if="isCloudDrivePage">
        <div class="sidebar-header drive-header">
          <el-icon class="back-btn" @click="$router.push('/home')"><ArrowLeft /></el-icon>
          <div class="drive-header-info" @click="$router.push('/home')">
            <el-icon color="#3370ff"><FolderOpened /></el-icon>
            <span class="drive-header-title">云盘</span>
          </div>
        </div>

        <!-- Search -->
        <div class="sidebar-search">
          <el-input v-model="driveFolderSearch" placeholder="搜索" prefix-icon="Search" size="small" clearable />
        </div>

        <!-- My Folders -->
        <div class="drive-tree-sections">
          <div class="drive-tree-section">
            <div class="drive-section-header" @click="handleMyFoldersClick">
              <el-icon class="drive-expand-arrow" :class="{ expanded: driveMyFoldersExpanded }"><ArrowRight /></el-icon>
              <el-icon color="#f5a623" :size="15"><FolderOpened /></el-icon>
              <span class="drive-section-title" :class="{ active: driveSelectedFolderId === null }">我的文件夹</span>
            </div>
            <div v-show="driveMyFoldersExpanded" class="drive-section-body">
              <div v-if="!driveFolderTree.length" class="drive-tree-empty">暂无文件夹</div>
              <FolderTreeNode
                v-for="node in filteredDriveFolders"
                :key="node.id"
                :node="node"
                :depth="0"
                :selected-id="driveSelectedFolderId"
                :search-keyword="driveFolderSearch"
                icon-color="#f5a623"
                @select="selectDriveFolder($event as any)"
                @contextmenu="(e: MouseEvent, n: any) => showDriveFolderCtx(e, n)"
                @toggle-expand="saveDriveExpandState"
                @load-children="(n: any) => loadDriveFolderChildren(n)"
              />
            </div>
          </div>

          <!-- Shared Folders -->
          <div class="drive-tree-section">
            <div class="drive-section-header" @click="driveSharedFoldersExpanded = !driveSharedFoldersExpanded">
              <el-icon class="drive-expand-arrow" :class="{ expanded: driveSharedFoldersExpanded }"><ArrowRight /></el-icon>
              <el-icon color="#3370ff" :size="15"><User /></el-icon>
              <span class="drive-section-title">共享文件夹</span>
            </div>
            <div v-show="driveSharedFoldersExpanded" class="drive-section-body">
              <div class="drive-tree-empty">暂无共享文件夹</div>
            </div>
          </div>
        </div>

        <!-- New Folder button -->
        <div class="drive-bottom-action" @click="handleNewDriveFolder">
          <el-icon><Plus /></el-icon>
          <span>新建文件夹</span>
        </div>
      </template>

      <!-- === Normal Main Navigation Mode === -->
      <template v-else>
        <!-- Header -->
        <div class="sidebar-header">
          <el-icon class="collapse-btn" @click="sidebarCollapsed = !sidebarCollapsed">
            <Fold v-if="!sidebarCollapsed" />
            <Expand v-else />
          </el-icon>
          <div class="logo" @click="$router.push('/home')">
            <span class="logo-icon">K</span>
            <span v-if="!sidebarCollapsed" class="logo-text">KLoggerX</span>
          </div>
        </div>

        <template v-if="!sidebarCollapsed">
          <!-- Search -->
          <div class="sidebar-search">
            <el-input v-model="searchKeyword" placeholder="搜索" prefix-icon="Search" size="small" clearable @keyup.enter="handleSearch" />
          </div>

          <!-- Nav Items -->
          <div class="sidebar-nav">
            <div class="nav-item" :class="{ active: activeMenu === '/home' }" @click="$router.push('/home')">
              <el-icon :size="18"><HomeFilled /></el-icon>
              <span class="nav-label">主页</span>
            </div>
            <div class="nav-item" :class="{ active: activeMenu === '/documents' }" @click="$router.push('/documents')">
              <el-icon :size="18"><FolderOpened /></el-icon>
              <span class="nav-label">云盘</span>
            </div>
            <div class="nav-item" :class="{ active: activeMenu === '/knowledge' }" @click="$router.push('/knowledge')">
              <el-icon :size="18"><Collection /></el-icon>
              <span class="nav-label">知识库</span>
            </div>
            <div v-if="isAdmin" class="nav-item" :class="{ active: activeMenu.startsWith('/admin') }" @click="$router.push('/admin/users')">
              <el-icon :size="18"><Setting /></el-icon>
              <span class="nav-label">系统管理</span>
            </div>
          </div>

          <!-- Sidebar Sections -->
          <div class="sidebar-sections">
            <!-- Pinned Documents -->
            <div class="sidebar-section">
              <div class="section-header" @click="pinnedDocsExpanded = !pinnedDocsExpanded">
                <span class="section-title">置顶文档</span>
              </div>
              <div v-show="pinnedDocsExpanded" class="section-body">
                <div v-for="doc in pinnedDocs" :key="doc.id" class="section-item" @click="openDocInNewTab(doc.id)" @contextmenu.prevent="showPinnedCtxMenu($event, doc)">
                  <el-icon :color="getTypeColor(doc.type)" :size="14"><component :is="getTypeIcon(doc.type)" /></el-icon>
                  <span class="section-item-text">{{ doc.title }}</span>
                </div>
                <div v-if="!pinnedDocs.length" class="section-empty">暂无置顶文档</div>
              </div>
            </div>

            <!-- Pinned Knowledge Base -->
            <div class="sidebar-section">
              <div class="section-header">
                <span class="section-title" @click="pinnedKbExpanded = !pinnedKbExpanded">置顶知识库</span>
                <div class="section-actions" @click.stop>
                  <el-icon :size="14" class="section-act-btn" @click.stop="showKbNewMenu = !showKbNewMenu"><Plus /></el-icon>
                  <el-icon :size="14" class="section-act-btn" @click.stop="ElMessage.info('知识库管理')"><MoreFilled /></el-icon>
                </div>
              </div>
              <!-- KB New Dropdown -->
              <div v-if="showKbNewMenu" class="sidebar-dropdown kb-dropdown" @click.stop>
                <div class="dropdown-group-title">新建</div>
                <div class="dropdown-item" @click="handleKbNew('blank')"><el-icon color="#3370ff"><Collection /></el-icon>空白知识库</div>
                <div class="dropdown-item" @click="handleKbNew('team')"><el-icon color="#36b37e"><UserFilled /></el-icon>团队项目</div>
                <div class="dropdown-item" @click="handleKbNew('product')"><el-icon color="#f5a623"><Compass /></el-icon>产品部门</div>
                <div class="dropdown-item" @click="handleKbNew('dev')"><el-icon color="#3370ff"><Monitor /></el-icon>研发部门</div>
                <div class="dropdown-item" @click="handleKbNew('design')"><el-icon color="#9254de"><Brush /></el-icon>设计部门</div>
                <div class="dropdown-item" @click="handleKbNew('marketing')"><el-icon color="#f54a45"><TrendCharts /></el-icon>市场营销</div>
                <div class="dropdown-item" @click="handleKbNew('more')"><el-icon color="#646a73"><Grid /></el-icon>更多模板</div>
                <div class="dropdown-sep" />
                <div class="dropdown-group-title">添加</div>
                <div class="dropdown-item" @click="handleKbNew('existing')"><el-icon color="#3370ff"><FolderAdd /></el-icon>已有知识库</div>
              </div>
              <div v-show="pinnedKbExpanded" class="section-body">
                <div class="section-item section-item-muted" @click="$router.push('/knowledge')">
                  <el-icon :size="14"><Plus /></el-icon>
                  <span class="section-item-text">新建或置顶知识库</span>
                </div>
              </div>
            </div>

            <!-- My Document Library -->
            <div class="sidebar-section">
              <div class="section-header">
                <span class="section-title" @click="docLibExpanded = !docLibExpanded">我的文档库</span>
                <el-icon class="section-collapse-arrow" :class="{ expanded: docLibExpanded }" @click="docLibExpanded = !docLibExpanded"><ArrowDown /></el-icon>
                <div class="section-actions" @click.stop>
                  <el-icon :size="14" class="section-act-btn" @click.stop="showDocLibMenu = !showDocLibMenu"><Plus /></el-icon>
                  <el-icon :size="14" class="section-act-btn" @click.stop="ElMessage.info('文档库设置')"><Setting /></el-icon>
                </div>
              </div>
              <!-- Doc Lib + Dropdown -->
              <div v-if="showDocLibMenu" class="sidebar-dropdown doclib-dropdown" @click.stop>
                <div class="dropdown-item" @click="handleDocLibCreate('doc')"><el-icon color="#3370ff"><Document /></el-icon>文档</div>
                <div class="dropdown-item" @click="handleDocLibCreate('sheet')"><el-icon color="#36b37e"><Grid /></el-icon>表格</div>
                <div class="dropdown-item" @click="handleDocLibCreate('slide')"><el-icon color="#ff7d00"><Monitor /></el-icon>幻灯片</div>
                <div class="dropdown-item" @click="handleDocLibCreate('bitable')"><el-icon color="#00b8d9"><Tickets /></el-icon>多维表格</div>
                <div class="dropdown-item" @click="handleDocLibCreate('survey')"><el-icon color="#f54a45"><Notebook /></el-icon>问卷</div>
                <div class="dropdown-item" @click="handleDocLibCreate('mindnote')"><el-icon color="#9254de"><Share /></el-icon>思维笔记</div>
                <div class="dropdown-group-title">文档应用</div>
                <div class="dropdown-item" @click="handleDocLibCreate('doc')"><el-icon color="#36b37e"><EditPen /></el-icon>画板</div>
                <div class="dropdown-item" @click="handleDocLibCreate('mindnote')"><el-icon color="#9254de"><Share /></el-icon>思维导图</div>
                <div class="dropdown-item" @click="handleDocLibCreate('doc')"><el-icon color="#ff7d00"><Connection /></el-icon>流程图</div>
                <div class="dropdown-sep" />
                <div class="dropdown-item"><el-icon color="#646a73"><Upload /></el-icon>上传及导入<el-icon class="arrow-right"><ArrowRight /></el-icon></div>
                <div class="dropdown-item" @click="handleImportDocs"><el-icon color="#3370ff"><FolderAdd /></el-icon>迁入已有云文档</div>
              </div>
              <div v-show="docLibExpanded" class="section-body doc-tree-body">
                <DocumentTree ref="docTreeRef" :show-header="false" @refresh="onTreeRefresh" />
              </div>
            </div>
          </div>

          <!-- Bottom Actions -->
          <div class="sidebar-bottom">
            <div class="bottom-item" @click="$router.push('/recycle-bin')">
              <el-icon :size="16"><Delete /></el-icon>
              <span>回收站</span>
            </div>
            <div class="bottom-item" @click="$router.push('/settings')">
              <el-icon :size="16"><Setting /></el-icon>
              <span>设置</span>
            </div>
          </div>
        </template>
      </template>
    </aside>

    <!-- Main Content -->
    <div class="main-content">
      <header class="topbar">
        <div class="topbar-left">
          <h2 class="topbar-title">{{ currentPageTitle }}</h2>
        </div>
        <div class="topbar-right">
          <el-icon class="topbar-icon" @click="focusSearch"><Search /></el-icon>
          <el-badge :value="unreadCount" :hidden="unreadCount === 0" :max="99" class="notification-badge">
            <el-icon class="topbar-icon" @click="showNotifications = true"><Bell /></el-icon>
          </el-badge>
          <el-icon class="topbar-icon" @click="ElMessage.info('帮助')"><QuestionFilled /></el-icon>
          <el-icon class="topbar-icon" @click="ElMessage.info('应用')"><Menu /></el-icon>
          <el-dropdown trigger="click" @command="handleUserCommand">
            <el-avatar :size="28" :src="userStore.user?.avatar" :style="{ background: '#3370ff', cursor: 'pointer' }">{{ userStore.user?.nickname?.[0] || 'U' }}</el-avatar>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="settings">个人设置</el-dropdown-item>
                <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>
      <main class="page-content" :class="{ 'no-padding': isCloudDrivePage || isAdminPage || isKnowledgePage }">
        <router-view />
      </main>
    </div>

    <!-- Pinned doc context menu -->
    <div v-if="pinnedCtxMenu.visible" class="ctx-menu-global" :style="{ left: pinnedCtxMenu.x + 'px', top: pinnedCtxMenu.y + 'px' }">
      <div class="ctx-item" @click="handlePinnedAction('open')"><el-icon><View /></el-icon>打开</div>
      <div class="ctx-item" @click="handlePinnedAction('unpin')"><el-icon><Flag /></el-icon>取消置顶</div>
      <div class="ctx-item" @click="handlePinnedAction('copyLink')"><el-icon><Link /></el-icon>复制链接</div>
    </div>

    <!-- Drive folder context menu -->
    <div v-if="driveFolderCtx.visible" class="ctx-menu-global" :style="{ left: driveFolderCtx.x + 'px', top: driveFolderCtx.y + 'px' }">
      <div class="ctx-item" @click="handleDriveFolderCtx('open')"><el-icon><View /></el-icon>打开</div>
      <div class="ctx-item" @click="handleDriveFolderCtx('rename')"><el-icon><EditPen /></el-icon>重命名</div>
      <div class="ctx-item" @click="handleDriveFolderCtx('share')"><el-icon><Share /></el-icon>分享</div>
      <div class="ctx-sep-global" />
      <div class="ctx-item danger" @click="handleDriveFolderCtx('delete')"><el-icon><Delete /></el-icon>删除</div>
    </div>

    <!-- Notification Drawer -->
    <el-drawer v-model="showNotifications" title="通知中心" direction="rtl" size="400px" class="notification-drawer">
      <template #header>
        <div class="notification-drawer-header">
          <span class="notification-title">通知中心</span>
          <el-button v-if="unreadCount > 0" link type="primary" size="small" @click="handleMarkAllRead">
            全部已读
          </el-button>
        </div>
      </template>
      <div class="notification-tabs">
        <span class="notif-tab" :class="{ active: notifTab === 'all' }" @click="notifTab = 'all'">全部</span>
        <span class="notif-tab" :class="{ active: notifTab === 'unread' }" @click="notifTab = 'unread'">未读 ({{ unreadCount }})</span>
      </div>
      <div class="notification-list" v-loading="notifLoading">
        <div v-if="!filteredNotifications.length" class="notification-empty">
          <el-icon :size="48" color="#c0c4cc"><Bell /></el-icon>
          <p>暂无通知</p>
        </div>
        <div
          v-for="n in filteredNotifications"
          :key="n.id"
          class="notification-item"
          :class="{ unread: !n.isRead }"
          @click="handleNotificationClick(n)"
        >
          <div class="notif-icon" :class="n.type">
            <el-icon>
              <component :is="getNotifIcon(n.type)" />
            </el-icon>
          </div>
          <div class="notif-content">
            <div class="notif-title">{{ n.title }}</div>
            <div class="notif-text">{{ n.content }}</div>
            <div class="notif-meta">
              <span class="notif-time">{{ formatNotifTime(n.createdAt) }}</span>
              <span v-if="n.docTitle" class="notif-doc">{{ n.docTitle }}</span>
            </div>
          </div>
          <el-icon v-if="!n.isRead" class="unread-dot"><CircleCheck /></el-icon>
        </div>
      </div>
      <div class="notification-footer" v-if="notifications.length >= 20">
        <el-button link @click="ElMessage.info('查看更多通知')">查看更多</el-button>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount, provide } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/store/modules/user'
import { useDocumentStore } from '@/store/modules/document'
import { createDocument, getDocumentTree, getPinnedDocuments, pinDocument, deleteDocument, updateDocument } from '@/api/modules/document'
import { createKnowledgeBase, getKnowledgeBaseList, deleteKnowledgeBase } from '@/api/modules/knowledge'
import { getNotifications, markNotificationRead, markAllNotificationsRead } from '@/api/modules/collaborate'
import DocumentTree from '@/components/document-tree/DocumentTree.vue'
import FolderTreeNode from '@/components/FolderTreeNode.vue'
import type { Document, DocumentType, Notification } from '@/types'
import { ElMessage, ElMessageBox } from 'element-plus'

interface DriveFolderNode {
  id: number
  title: string
  type: string
  parentId: number | null
  isExpanded: boolean
  isLoading?: boolean  // 加载状态
  hasMore?: boolean    // 是否有更多子项需要懒加载
  children: DriveFolderNode[]
  [key: string]: any
}

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const documentStore = useDocumentStore()
const docTreeRef = ref()

const sidebarCollapsed = ref(false)
const searchKeyword = ref('')
const showNotifications = ref(false)
const notifications = ref<Notification[]>([])
const pinnedDocs = ref<Document[]>([])
const notifTab = ref<'all' | 'unread'>('all')
const notifLoading = ref(false)
const unreadCount = computed(() => notifications.value.filter((n) => !n.isRead).length)

const filteredNotifications = computed(() => {
  if (notifTab.value === 'unread') {
    return notifications.value.filter(n => !n.isRead)
  }
  return notifications.value
})

const isCloudDrivePage = computed(() => route.path.startsWith('/documents'))
const isAdminPage = computed(() => route.path.startsWith('/admin'))
const isKnowledgePage = computed(() => route.path.startsWith('/knowledge'))

const isAdmin = computed(() => userStore.user?.role === 'admin')

const activeMenu = computed(() => {
  const p = route.path
  if (p.startsWith('/home')) return '/home'
  if (p.startsWith('/documents')) return '/documents'
  if (p.startsWith('/knowledge')) return '/knowledge'
  if (p.startsWith('/admin')) return '/admin'
  if (p.startsWith('/recycle-bin')) return '/recycle-bin'
  if (p.startsWith('/settings')) return '/settings'
  return p
})

const currentPageTitle = computed(() => {
  const meta = route.meta as any
  return meta?.title || ''
})

// Section expand states
const pinnedDocsExpanded = ref(true)
const pinnedKbExpanded = ref(true)
const docLibExpanded = ref(true)

// ============= Knowledge Base sidebar state =============
const kbSearch = ref('')
const kbPersonalExpanded = ref(true)
const kbSharedExpanded = ref(true)
const kbPinnedExpanded = ref(true)
const kbRecentExpanded = ref(true)
const kbList = ref<{ id: number; name: string; description: string }[]>([])
const kbCoverColors = ['#3370ff', '#36b37e', '#ff7d00', '#f54a45', '#9254de', '#00b8d9', '#f5a623', '#7b61ff']

const filteredKbList = computed(() => {
  const q = kbSearch.value.toLowerCase()
  if (!q) return kbList.value
  return kbList.value.filter(kb => kb.name.toLowerCase().includes(q))
})

async function fetchKbList() {
  try {
    const res: any = await getKnowledgeBaseList({ page: 1, pageSize: 100 })
    kbList.value = (res.data?.list || []).map((kb: any) => ({
      id: kb.id,
      name: kb.name,
      description: kb.description || '',
    }))
  } catch { /* ignore */ }
}

watch(isKnowledgePage, (val) => {
  if (val) fetchKbList()
}, { immediate: true })

// Knowledge Base context menu
const kbTreeCtx = ref<{ visible: boolean; x: number; y: number; kb: { id: number; name: string } | null }>({
  visible: false, x: 0, y: 0, kb: null
})

function showKbTreeCtx(e: MouseEvent, kb: { id: number; name: string }) {
  e.preventDefault()
  e.stopPropagation()
  kbTreeCtx.value = { visible: true, x: e.clientX, y: e.clientY, kb }
}

async function handleKbTreeAction(action: string) {
  const kb = kbTreeCtx.value.kb
  kbTreeCtx.value.visible = false
  if (!kb) return
  switch (action) {
    case 'open':
      router.push(`/knowledge/${kb.id}`)
      break
    case 'openNew':
      window.open(`${window.location.origin}/knowledge/${kb.id}`, '_blank')
      break
    case 'share':
      ElMessage.success('分享链接已复制')
      navigator.clipboard.writeText(`${window.location.origin}/knowledge/${kb.id}`)
      break
    case 'copyLink':
      navigator.clipboard.writeText(`${window.location.origin}/knowledge/${kb.id}`)
      ElMessage.success('链接已复制')
      break
    case 'settings':
      router.push(`/knowledge/${kb.id}`)
      break
    case 'delete':
      await ElMessageBox.confirm(`确定删除知识库"${kb.name}"？删除后不可恢复。`, '删除确认', { type: 'warning' })
      await deleteKnowledgeBase(kb.id)
      ElMessage.success('已删除')
      fetchKbList()
      break
  }
}

async function handleNewKb() {
  try {
    const { value } = await ElMessageBox.prompt('请输入知识库名称', '新建知识库', {
      confirmButtonText: '创建',
      cancelButtonText: '取消',
      inputPlaceholder: '输入知识库名称',
    })
    if (value?.trim()) {
      await createKnowledgeBase({ name: value.trim(), description: '' })
      ElMessage.success('知识库已创建')
      fetchKbList()
    }
  } catch { /* cancelled */ }
}

// Dropdowns
const showKbNewMenu = ref(false)
const showDocLibMenu = ref(false)

// Pinned context menu
const pinnedCtxMenu = ref({ visible: false, x: 0, y: 0, doc: null as Document | null })

// ============= Cloud Drive sidebar state =============
const driveFolderSearch = ref('')
const driveMyFoldersExpanded = ref(true)
const driveSharedFoldersExpanded = ref(true)
const driveSelectedFolderId = ref<number | null>(null)
const driveFolderTree = ref<DriveFolderNode[]>([])

// Provide selected folder to DocumentLibrary
provide('driveSelectedFolderId', driveSelectedFolderId)

const driveFolderCtx = ref({ visible: false, x: 0, y: 0, folder: null as DriveFolderNode | null })

const filteredDriveFolders = computed(() => {
  const q = driveFolderSearch.value.toLowerCase()
  if (!q) return driveFolderTree.value
  function filterTree(nodes: DriveFolderNode[]): DriveFolderNode[] {
    return nodes.reduce<DriveFolderNode[]>((acc, node) => {
      const nameMatch = node.title.toLowerCase().includes(q)
      const filteredChildren = filterTree(node.children || [])
      if (nameMatch || filteredChildren.length > 0) {
        acc.push({ ...node, isExpanded: true, children: nameMatch ? node.children : filteredChildren })
      }
      return acc
    }, [])
  }
  return filterTree(driveFolderTree.value)
})

function buildDriveFolderTree(docs: Document[], parentId: number | null = null): DriveFolderNode[] {
  return docs
    .filter(d => (d.parentId ?? null) === parentId)
    .sort((a, b) => {
      // Folders first, then documents
      if (a.type === 'folder' && b.type !== 'folder') return -1
      if (a.type !== 'folder' && b.type === 'folder') return 1
      return a.title.localeCompare(b.title)
    })
    .map(d => ({
      ...d,
      isExpanded: false,
      isLoading: false,
      children: [],
      hasMore: d.type === 'folder', // 文件夹可能有子项，需要懒加载
    }))
}

async function fetchDriveFolderTree() {
  try {
    const res: any = await getDocumentTree(null)
    const all: Document[] = res.data || []
    driveFolderTree.value = buildDriveFolderTree(all, null)
    // Restore expand state from localStorage
    restoreDriveExpandState()
  } catch { /* ignore */ }
}

// ============= Expand State Persistence =============
const DRIVE_EXPAND_KEY = 'kloggerx_drive_expanded'

function saveDriveExpandState() {
  const expandedIds: number[] = []
  function collectExpanded(nodes: DriveFolderNode[]) {
    for (const n of nodes) {
      if (n.isExpanded && n.children?.length) {
        expandedIds.push(n.id)
      }
      if (n.children?.length) collectExpanded(n.children)
    }
  }
  collectExpanded(driveFolderTree.value)
  localStorage.setItem(DRIVE_EXPAND_KEY, JSON.stringify(expandedIds))
}

function restoreDriveExpandState() {
  try {
    const saved = localStorage.getItem(DRIVE_EXPAND_KEY)
    if (!saved) return
    const expandedIds: number[] = JSON.parse(saved)
    function applyExpanded(nodes: DriveFolderNode[]) {
      for (const n of nodes) {
        if (expandedIds.includes(n.id)) {
          n.isExpanded = true
        }
        if (n.children?.length) applyExpanded(n.children)
      }
    }
    applyExpanded(driveFolderTree.value)
  } catch { /* ignore */ }
}

function handleMyFoldersClick() {
  driveMyFoldersExpanded.value = !driveMyFoldersExpanded.value
  // When clicking "我的文件夹", navigate to root level
  if (driveMyFoldersExpanded.value) {
    driveSelectedFolderId.value = null
    router.push({ path: '/documents' })
  }
}

function selectDriveFolder(node: DriveFolderNode) {
  driveSelectedFolderId.value = node.id
  // The FolderTreeNode now handles expand/collapse internally
  router.push({ path: '/documents', query: { folderId: String(node.id) } })
}

// 动态加载子文件夹
async function loadDriveFolderChildren(node: DriveFolderNode) {
  try {
    const res: any = await getDocumentTree(node.id)
    const children: Document[] = res.data || []

    // 按文件夹优先排序
    const sortedChildren: DriveFolderNode[] = children.sort((a, b) => {
      if (a.type === 'folder' && b.type !== 'folder') return -1
      if (a.type !== 'folder' && b.type === 'folder') return 1
      return a.title.localeCompare(b.title)
    }).map(d => ({
      ...d,
      isExpanded: false,
      isLoading: false,
      children: [] as DriveFolderNode[],
      hasMore: d.type === 'folder' // 文件夹可能有子项
    }))

    // 更新节点的 children
    node.children = sortedChildren
    node.hasMore = sortedChildren.some(c => c.type === 'folder' || c.type === undefined)

    // 保存展开状态
    saveDriveExpandState()
  } catch {
    // 加载失败时，标记为无更多内容
    node.hasMore = false
  } finally {
    // 无论成功还是失败，都要重置加载状态
    node.isLoading = false
  }
}

function showDriveFolderCtx(e: MouseEvent, node: DriveFolderNode) {
  e.preventDefault()
  e.stopPropagation()
  driveFolderCtx.value = { visible: true, x: e.clientX, y: e.clientY, folder: node }
}

async function handleDriveFolderCtx(action: string) {
  const folder = driveFolderCtx.value.folder
  driveFolderCtx.value.visible = false
  if (!folder) return
  switch (action) {
    case 'open':
      selectDriveFolder(folder)
      break
    case 'rename': {
      const { value } = await ElMessageBox.prompt('请输入新名称', '重命名', {
        inputValue: folder.title,
        confirmButtonText: '确定',
        cancelButtonText: '取消',
      }).catch(() => ({ value: null }))
      if (value) {
        await updateDocument(folder.id, { title: value })
        ElMessage.success('重命名成功')
        fetchDriveFolderTree()
      }
      break
    }
    case 'share':
      ElMessage.success(`分享链接已复制: ${window.location.origin}/documents?folderId=${folder.id}`)
      navigator.clipboard.writeText(`${window.location.origin}/documents?folderId=${folder.id}`)
      break
    case 'delete':
      await ElMessageBox.confirm(`确定将文件夹"${folder.title}"移至回收站？`, '删除确认')
      await deleteDocument(folder.id)
      ElMessage.success('已移至回收站')
      fetchDriveFolderTree()
      break
  }
}

async function handleNewDriveFolder() {
  try {
    await createDocument({ title: '新建文件夹', type: 'folder', parentId: driveSelectedFolderId.value })
    ElMessage.success('文件夹已创建')
    fetchDriveFolderTree()
  } catch { /* handled */ }
}

// Watch route to sync drive folder state
watch(() => route.query.folderId, (val) => {
  driveSelectedFolderId.value = val ? Number(val) : null
}, { immediate: true })

// Fetch drive tree when entering cloud drive page
watch(isCloudDrivePage, (val) => {
  if (val) fetchDriveFolderTree()
}, { immediate: true })

// ============= End Cloud Drive sidebar =============

const typeMap: Record<string, { icon: string; color: string }> = {
  folder: { icon: 'Folder', color: '#f5a623' },
  doc: { icon: 'Document', color: '#3370ff' },
  sheet: { icon: 'Grid', color: '#36b37e' },
  slide: { icon: 'Monitor', color: '#ff7d00' },
  mindnote: { icon: 'Share', color: '#9254de' },
  bitable: { icon: 'Tickets', color: '#00b8d9' },
  survey: { icon: 'Notebook', color: '#f54a45' },
}
function getTypeIcon(t: string) { return typeMap[t]?.icon || 'Document' }
function getTypeColor(t: string) { return typeMap[t]?.color || '#3370ff' }

function focusSearch() {
  sidebarCollapsed.value = false
  setTimeout(() => {
    const el = document.querySelector('.sidebar-search input') as HTMLInputElement
    el?.focus()
  }, 200)
}

function handleSearch() {
  if (searchKeyword.value.trim()) {
    router.push({ path: '/home', query: { keyword: searchKeyword.value.trim() } })
  }
}

function handleUserCommand(cmd: string) {
  if (cmd === 'settings') router.push('/settings')
  else if (cmd === 'logout') userStore.logout()
}

// Knowledge Base new menu
async function handleKbNew(type: string) {
  showKbNewMenu.value = false
  if (type === 'blank') {
    try {
      await createKnowledgeBase({ name: '未命名知识库', description: '' })
      ElMessage.success('知识库已创建')
      router.push('/knowledge')
    } catch { /* handled */ }
  } else if (type === 'existing') {
    router.push('/knowledge')
  } else if (type === 'more') {
    router.push('/knowledge')
  } else {
    const nameMap: Record<string, string> = { team: '团队项目', product: '产品部门', dev: '研发部门', design: '设计部门', marketing: '市场营销' }
    try {
      await createKnowledgeBase({ name: nameMap[type] || '未命名知识库', description: `基于${nameMap[type] || '模板'}创建` })
      ElMessage.success('知识库已创建')
      router.push('/knowledge')
    } catch { /* handled */ }
  }
}

// Doc Lib create
async function handleDocLibCreate(type: DocumentType | 'folder') {
  showDocLibMenu.value = false
  try {
    const res: any = await createDocument({ title: type === 'folder' ? '新建文件夹' : '无标题文档', type, parentId: null })
    docTreeRef.value?.refreshTree()
    if (type !== 'folder') {
      // Open in new browser tab
      window.open(`/doc/${res.data.id}`, '_blank')
    } else {
      ElMessage.success('文件夹已创建')
    }
  } catch { /* handled */ }
}

function openDocInNewTab(docId: number) {
  window.open(`/doc/${docId}`, '_blank')
}

function handleImportDocs() {
  ElMessage.success('请使用上传功能导入文档')
}

function onTreeRefresh() {
  documentStore.fetchPinned()
  fetchPinnedDocs()
}

// Pinned doc context menu
function showPinnedCtxMenu(e: MouseEvent, doc: Document) {
  e.preventDefault()
  pinnedCtxMenu.value = { visible: true, x: e.clientX, y: e.clientY, doc }
}

async function handlePinnedAction(action: string) {
  const doc = pinnedCtxMenu.value.doc
  pinnedCtxMenu.value.visible = false
  if (!doc) return
  if (action === 'open') {
    // Open in new browser tab
    window.open(`/doc/${doc.id}`, '_blank')
  } else if (action === 'unpin') {
    await pinDocument(doc.id, false)
    ElMessage.success('已取消置顶')
    fetchPinnedDocs()
  } else if (action === 'copyLink') {
    navigator.clipboard.writeText(`${window.location.origin}/doc/${doc.id}`)
    ElMessage.success('链接已复制')
  }
}

function closeAllMenus() {
  showKbNewMenu.value = false
  showDocLibMenu.value = false
  pinnedCtxMenu.value.visible = false
  driveFolderCtx.value.visible = false
}

async function fetchNotifications() {
  notifLoading.value = true
  try {
    const res: any = await getNotifications({ page: 1, pageSize: 50 })
    notifications.value = res.data?.list || []
  } catch { /* ignore */ }
  finally { notifLoading.value = false }
}

async function handleNotificationClick(n: Notification) {
  if (!n.isRead) {
    try {
      await markNotificationRead(n.id)
      n.isRead = true
    } catch { /* ignore */ }
  }
  showNotifications.value = false
  if (n.documentId) {
    window.open(`/doc/${n.documentId}`, '_blank')
  }
}

async function handleMarkAllRead() {
  try {
    await markAllNotificationsRead()
    notifications.value.forEach(n => n.isRead = true)
    ElMessage.success('已全部标记为已读')
  } catch {
    ElMessage.error('操作失败')
  }
}

function getNotifIcon(type: string): string {
  const icons: Record<string, string> = {
    comment: 'ChatDotSquare',
    mention: 'User',
    share: 'Share',
    system: 'Bell',
    approval: 'DocumentChecked',
  }
  return icons[type] || 'Bell'
}

function formatNotifTime(time: string): string {
  if (!time) return ''
  const date = new Date(time)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)
  
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 7) return `${days}天前`
  return date.toLocaleDateString('zh-CN')
}

async function fetchPinnedDocs() {
  try {
    const res: any = await getPinnedDocuments()
    pinnedDocs.value = res.data || []
  } catch { /* ignore */ }
}

onMounted(() => {
  userStore.fetchUserInfo()
  fetchNotifications()
  fetchPinnedDocs()
  document.addEventListener('click', closeAllMenus)
})
onBeforeUnmount(() => {
  document.removeEventListener('click', closeAllMenus)
})
</script>

<style scoped>
.main-layout {
  display: flex;
  width: 100%;
  height: 100vh;
}

/* Sidebar */
.sidebar {
  width: var(--kx-sidebar-width);
  background: var(--kx-sidebar-bg);
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--kx-border);
  transition: width 0.2s;
  flex-shrink: 0;
  overflow: hidden;
}
.sidebar.collapsed {
  width: 64px;
}
.sidebar-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 16px 8px;
  flex-shrink: 0;
}
.collapse-btn {
  cursor: pointer;
  font-size: 18px;
  color: var(--kx-text-secondary);
  flex-shrink: 0;
}
.logo {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}
.logo-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  background: var(--kx-primary);
  color: #fff;
  border-radius: 6px;
  font-weight: 700;
  font-size: 14px;
}
.logo-text {
  font-size: 15px;
  font-weight: 600;
  color: var(--kx-text-primary);
}

/* Search */
.sidebar-search {
  padding: 4px 12px 8px;
  flex-shrink: 0;
}

/* Nav Items */
.sidebar-nav {
  padding: 0 8px;
  flex-shrink: 0;
}
.nav-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 8px 10px;
  border-radius: 8px;
  cursor: pointer;
  margin-bottom: 2px;
  color: var(--kx-text-primary);
  transition: background 0.15s;
}
.nav-item:hover {
  background: rgba(0,0,0,0.04);
}
.nav-item.active {
  background: rgba(51,112,255,0.08);
  color: var(--kx-primary);
}
.nav-item .el-icon {
  margin-top: 2px;
  flex-shrink: 0;
}
.nav-label {
  font-size: 14px;
  font-weight: 500;
}

/* Sidebar Sections */
.sidebar-sections {
  flex: 1;
  overflow-y: auto;
  padding: 0 8px;
  margin-top: 4px;
}
.sidebar-section {
  margin-top: 4px;
  position: relative;
}
.section-header {
  display: flex;
  align-items: center;
  padding: 6px 8px 4px;
  gap: 4px;
}
.section-title {
  font-size: 12px;
  color: var(--kx-text-placeholder);
  font-weight: 500;
  cursor: pointer;
  flex: 1;
}
.section-collapse-arrow {
  font-size: 10px;
  color: var(--kx-text-placeholder);
  cursor: pointer;
  transition: transform 0.2s;
  margin-right: 2px;
}
.section-collapse-arrow.expanded {
  transform: rotate(0deg);
}
.section-collapse-arrow:not(.expanded) {
  transform: rotate(-90deg);
}
.section-actions {
  display: flex;
  gap: 2px;
  opacity: 0;
  transition: opacity 0.15s;
}
.section-header:hover .section-actions {
  opacity: 1;
}
.section-act-btn {
  cursor: pointer;
  color: var(--kx-text-placeholder);
  padding: 2px;
  border-radius: 4px;
}
.section-act-btn:hover {
  color: var(--kx-primary);
  background: rgba(51,112,255,0.08);
}
.section-body {
  padding: 0 2px;
}
.section-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 5px 8px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  color: var(--kx-text-primary);
}
.section-item:hover {
  background: rgba(0,0,0,0.04);
}
.section-item-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}
.section-item-muted {
  color: var(--kx-text-placeholder);
}
.section-empty {
  padding: 4px 8px;
  font-size: 12px;
  color: var(--kx-text-placeholder);
}
.doc-tree-body {
  padding: 0;
}

/* Sidebar Dropdown */
.sidebar-dropdown {
  position: absolute;
  left: 8px;
  top: 28px;
  min-width: 200px;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  z-index: 200;
  padding: 4px 0;
  max-height: 400px;
  overflow-y: auto;
}
.doclib-dropdown {
  left: 30px;
  top: 28px;
}
.kb-dropdown {
  left: 30px;
  top: 28px;
}
.dropdown-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  font-size: 13px;
  cursor: pointer;
  color: var(--kx-text-primary);
}
.dropdown-item:hover {
  background: var(--kx-sidebar-bg);
}
.dropdown-group-title {
  font-size: 11px;
  color: var(--kx-text-placeholder);
  padding: 6px 14px 2px;
}
.dropdown-sep {
  height: 1px;
  background: var(--kx-border);
  margin: 4px 0;
}
.arrow-right {
  margin-left: auto;
  font-size: 11px;
  color: var(--kx-text-placeholder);
}

/* Bottom */
.sidebar-bottom {
  flex-shrink: 0;
  border-top: 1px solid var(--kx-border);
  padding: 6px 8px;
  display: flex;
  flex-direction: column;
  gap: 0;
}
.bottom-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  color: var(--kx-text-secondary);
}
.bottom-item:hover {
  background: rgba(0,0,0,0.04);
  color: var(--kx-text-primary);
}

/* Main Content */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}
.topbar {
  height: var(--kx-header-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  background: #fff;
  border-bottom: 1px solid var(--kx-border);
  flex-shrink: 0;
}
.topbar-left {
  display: flex;
  align-items: center;
}
.topbar-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--kx-text-primary);
  margin: 0;
}
.topbar-right {
  display: flex;
  align-items: center;
  gap: 14px;
}
.topbar-icon {
  font-size: 20px;
  cursor: pointer;
  color: var(--kx-text-secondary);
}
.topbar-icon:hover {
  color: var(--kx-primary);
}
.notification-badge {
  line-height: 1;
}
.page-content {
  flex: 1;
  overflow-y: auto;
  padding: 0 24px 24px;
  background: #fff;
}
.page-content.no-padding {
  padding: 0;
}

/* ============= Cloud Drive Sidebar ============= */
.drive-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 12px 8px;
  flex-shrink: 0;
}
.back-btn {
  font-size: 18px;
  color: var(--kx-text-secondary);
  cursor: pointer;
  padding: 4px;
  border-radius: 6px;
  flex-shrink: 0;
}
.back-btn:hover {
  background: rgba(0,0,0,0.06);
  color: var(--kx-text-primary);
}
.drive-header-info {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}
.drive-header-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--kx-text-primary);
}
.drive-tree-sections {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
}
.drive-tree-section {
  margin-bottom: 4px;
}
.drive-section-header {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 7px 12px;
  cursor: pointer;
  user-select: none;
  border-radius: 6px;
  margin: 0 6px;
  transition: background 0.15s;
}
.drive-section-header:hover {
  background: rgba(0,0,0,0.04);
}
.drive-section-title {
  flex: 1;
  font-size: 13px;
  font-weight: 500;
  color: var(--kx-text-primary);
}
.drive-section-title.active {
  color: var(--kx-primary);
  font-weight: 600;
}
.drive-section-body {
  padding: 0;
}
.drive-expand-arrow {
  font-size: 12px;
  color: var(--kx-text-placeholder);
  transition: transform 0.2s;
  flex-shrink: 0;
}
.drive-expand-arrow.expanded {
  transform: rotate(90deg);
}
.drive-expand-arrow.small {
  font-size: 10px;
}
.drive-expand-arrow.invisible {
  visibility: hidden;
}
.drive-tree-empty {
  padding: 6px 18px;
  font-size: 12px;
  color: var(--kx-text-placeholder);
}
.drive-bottom-action {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 12px;
  margin: 4px 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  color: var(--kx-primary);
  background: rgba(51,112,255,0.06);
  border: 1px dashed rgba(51,112,255,0.3);
  transition: background 0.15s;
}
.drive-bottom-action:hover {
  background: rgba(51,112,255,0.12);
}

/* Knowledge Base Tree Item */
.kb-tree-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px 6px 28px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  color: var(--kx-text-primary);
  margin: 1px 6px;
  transition: background 0.15s;
}
.kb-tree-item:hover {
  background: rgba(0, 0, 0, 0.04);
}
.kb-tree-item.active {
  background: rgba(51, 112, 255, 0.1);
}
.kb-tree-item.active .kb-tree-name {
  color: var(--kx-primary);
  font-weight: 500;
}
.kb-tree-icon {
  flex-shrink: 0;
  font-size: 15px;
}
.kb-tree-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.kb-tree-actions {
  display: none;
  margin-left: auto;
  cursor: pointer;
  color: var(--kx-text-placeholder);
  flex-shrink: 0;
  padding: 2px;
  border-radius: 4px;
}
.kb-tree-actions:hover {
  color: var(--kx-text-primary);
  background: rgba(0, 0, 0, 0.06);
}
.kb-tree-item:hover .kb-tree-actions {
  display: flex;
}

/* Context menu separator */
.ctx-sep-global {
  height: 1px;
  background: var(--kx-border);
  margin: 4px 0;
}
.ctx-item.danger {
  color: #f54a45;
}
.ctx-item.danger:hover {
  background: rgba(245,74,69,0.06);
}

/* Global context menu */
.ctx-menu-global {
  position: fixed;
  z-index: 999;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  padding: 4px 0;
  min-width: 160px;
}
.ctx-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 14px;
  font-size: 13px;
  cursor: pointer;
  color: var(--kx-text-primary);
}
.ctx-item:hover {
  background: var(--kx-sidebar-bg);
}

/* Notification Drawer */
.notification-drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}
.notification-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--kx-text-primary);
}
.notification-tabs {
  display: flex;
  gap: 4px;
  padding: 0 0 12px;
  border-bottom: 1px solid var(--kx-border);
  margin-bottom: 8px;
}
.notif-tab {
  padding: 6px 16px;
  border-radius: 16px;
  font-size: 13px;
  cursor: pointer;
  color: var(--kx-text-secondary);
  transition: all 0.2s;
  user-select: none;
}
.notif-tab:hover {
  background: rgba(0,0,0,0.04);
}
.notif-tab.active {
  background: rgba(51,112,255,0.1);
  color: var(--kx-primary);
  font-weight: 500;
}
.notification-list {
  flex: 1;
  overflow-y: auto;
  min-height: 200px;
}
.notification-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: var(--kx-text-placeholder);
}
.notification-empty p {
  margin-top: 12px;
  font-size: 14px;
}
.notification-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.15s;
  position: relative;
}
.notification-item:hover {
  background: rgba(0,0,0,0.03);
}
.notification-item.unread {
  background: rgba(51,112,255,0.04);
}
.notification-item.unread:hover {
  background: rgba(51,112,255,0.08);
}
.notif-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 10px;
  flex-shrink: 0;
  font-size: 18px;
  color: #fff;
  background: #8592a6;
}
.notif-icon.comment {
  background: #3370ff;
}
.notif-icon.mention {
  background: #f5a623;
}
.notif-icon.share {
  background: #36b37e;
}
.notif-icon.system {
  background: #8592a6;
}
.notif-icon.approval {
  background: #9254de;
}
.notif-content {
  flex: 1;
  min-width: 0;
}
.notif-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--kx-text-primary);
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.notif-text {
  font-size: 13px;
  color: var(--kx-text-secondary);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-bottom: 6px;
}
.notif-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--kx-text-placeholder);
}
.notif-time {
  white-space: nowrap;
}
.notif-doc {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--kx-primary);
  max-width: 180px;
}
.notif-doc::before {
  content: '·';
  margin-right: 8px;
  color: var(--kx-text-placeholder);
}
.unread-dot {
  color: var(--kx-primary);
  font-size: 14px;
  flex-shrink: 0;
  margin-top: 2px;
}
.notification-footer {
  text-align: center;
  padding: 12px 0;
  border-top: 1px solid var(--kx-border);
  margin-top: 8px;
}
</style>
