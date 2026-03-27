<template>
  <div class="admin-layout">
    <div class="admin-sidebar" :class="{ collapsed: sidebarCollapsed }">
      <div class="sidebar-header">
        <el-icon :size="20"><Setting /></el-icon>
        <span v-show="!sidebarCollapsed">系统管理</span>
        <el-icon
          class="collapse-btn"
          @click="sidebarCollapsed = !sidebarCollapsed"
        >
          <component :is="sidebarCollapsed ? 'ArrowRight' : 'ArrowLeft'" />
        </el-icon>
      </div>
      <div class="sidebar-menu">
        <div
          v-for="item in menuItems"
          :key="item.path"
          class="menu-item"
          :class="{ active: isActive(item.path) }"
          @click="$router.push(item.path)"
        >
          <el-tooltip :content="sidebarCollapsed ? item.label : ''" placement="right">
            <div class="menu-item-content">
              <el-icon :size="18"><component :is="item.icon" /></el-icon>
              <span v-show="!sidebarCollapsed">{{ item.label }}</span>
            </div>
          </el-tooltip>
        </div>
      </div>
      <div class="sidebar-footer">
        <el-button
          text
          @click="goBack"
          class="back-btn"
        >
          <el-icon><ArrowLeft /></el-icon>
          <span v-show="!sidebarCollapsed">返回</span>
        </el-button>
      </div>
    </div>
    <div class="admin-content">
      <router-view />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const sidebarCollapsed = ref(false)

const menuItems = [
  { path: '/admin/users', label: '用户和权限', icon: 'User' },
  { path: '/admin/departments', label: '部门管理', icon: 'OfficeBuilding' },
  { path: '/admin/ai-models', label: 'AI模型设置', icon: 'Cpu' },
  { path: '/admin/storage', label: '云盘存储', icon: 'FolderOpened' },
]

function isActive(path: string) {
  return route.path === path
}

function goBack() {
  router.push('/')
}
</script>

<style scoped>
.admin-layout {
  display: flex;
  height: 100%;
  background: #f7f8fa;
}
.admin-sidebar {
  width: 220px;
  background: #fff;
  border-right: 1px solid var(--kx-border);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  transition: width 0.2s ease;
}
.admin-sidebar.collapsed {
  width: 60px;
}
.sidebar-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 16px 12px;
  font-size: 16px;
  font-weight: 600;
  color: var(--kx-text-primary);
  border-bottom: 1px solid var(--kx-border);
  position: relative;
}
.sidebar-header span {
  white-space: nowrap;
  overflow: hidden;
}
.collapse-btn {
  margin-left: auto;
  cursor: pointer;
  color: var(--kx-text-secondary);
  transition: color 0.15s;
}
.collapse-btn:hover {
  color: var(--kx-primary);
}
.collapsed .collapse-btn {
  margin-left: 0;
}
.sidebar-menu {
  flex: 1;
  padding: 8px;
  overflow-y: auto;
}
.menu-item {
  display: flex;
  align-items: center;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  color: var(--kx-text-primary);
  transition: background 0.15s;
  margin-bottom: 4px;
}
.menu-item:hover {
  background: rgba(0,0,0,0.04);
}
.menu-item.active {
  background: rgba(51,112,255,0.08);
  color: var(--kx-primary);
}
.menu-item-content {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  width: 100%;
}
.menu-item-content span {
  white-space: nowrap;
  overflow: hidden;
}
.collapsed .menu-item-content {
  justify-content: center;
  padding: 10px;
}
.sidebar-footer {
  padding: 12px;
  border-top: 1px solid var(--kx-border);
}
.back-btn {
  width: 100%;
  justify-content: flex-start;
  color: var(--kx-text-secondary);
}
.collapsed .back-btn {
  justify-content: center;
  padding: 8px;
}
.back-btn:hover {
  color: var(--kx-primary);
}
.admin-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}
</style>
