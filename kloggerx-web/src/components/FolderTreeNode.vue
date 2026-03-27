<template>
  <div class="sidebar-tree-node">
    <div
      class="sidebar-tree-item"
      :class="{ active: selectedId === node.id, 'is-folder': isFolder }"
      :style="{ paddingLeft: actualDepth * 20 + 12 + 'px' }"
      @click="handleClick"
      @dblclick="handleDoubleClick"
      @contextmenu.prevent="$emit('contextmenu', $event, node)"
    >
      <el-icon
        v-if="isFolder"
        class="tree-arrow"
        :class="{ expanded: node.isExpanded, invisible: !hasChildren && !isLoading && !node.hasMore, loading: isLoading }"
        @click.stop="toggleExpand"
      ><ArrowRight /></el-icon>
      <span v-else class="tree-arrow-placeholder" />
      <el-icon v-if="isLoading" class="is-loading" color="var(--kx-primary)" :size="15"><Loading /></el-icon>
      <el-icon v-else :color="nodeIconColor" :size="15" class="tree-node-icon"><component :is="nodeIcon" /></el-icon>
      <span class="tree-node-name" :class="{ 'match-highlight': isSearchMatch }">{{ node.title }}</span>
      <span v-if="isFolder && (hasChildren || node.hasMore)" class="tree-node-actions" @click.stop="$emit('contextmenu', $event, node)">
        <el-icon :size="14"><MoreFilled /></el-icon>
      </span>
    </div>
    <transition name="tree-expand">
      <div v-if="isFolder && node.isExpanded" class="sidebar-tree-children">
        <div v-if="isLoading" class="tree-loading-placeholder">
          <span>加载中...</span>
        </div>
        <template v-else>
          <FolderTreeNode
            v-for="child in sortedChildren"
            :key="child.id"
            :node="child"
            :depth="actualDepth + 1"
            :selected-id="selectedId"
            :icon-color="iconColor"
            :search-keyword="searchKeyword"
            @select="$emit('select', $event)"
            @contextmenu="(e: MouseEvent, n: any) => $emit('contextmenu', e, n)"
            @open-doc="(id: number) => $emit('open-doc', id)"
            @toggle-expand="(n: any) => $emit('toggle-expand', n)"
            @rename="(n: any) => $emit('rename', n)"
            @load-children="(n: any) => $emit('load-children', n)"
          />
        </template>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Loading } from '@element-plus/icons-vue'

defineOptions({ name: 'FolderTreeNode' })

interface FolderNode {
  id: number
  title: string
  type?: string
  children?: FolderNode[]
  isExpanded?: boolean
  isLoading?: boolean  // 加载状态 - 由父组件控制
  hasMore?: boolean    // 是否有更多子项需要加载
  [key: string]: any
}

const props = defineProps<{
  node: FolderNode
  depth?: number
  selectedId?: number | null
  iconColor?: string
  searchKeyword?: string
}>()

const emit = defineEmits<{
  select: [node: FolderNode]
  contextmenu: [e: MouseEvent, node: FolderNode]
  'open-doc': [id: number]
  'toggle-expand': [node: FolderNode]
  'rename': [node: FolderNode]
  'load-children': [node: FolderNode]
}>()

const actualDepth = computed(() => props.depth ?? 0)

// 使用 node.isLoading 作为加载状态，由父组件控制
const isLoading = computed(() => props.node.isLoading === true)

const isFolder = computed(() => !props.node.type || props.node.type === 'folder')

const hasChildren = computed(() => props.node.children && props.node.children.length > 0)

const sortedChildren = computed(() => {
  if (!props.node.children) return []
  return [...props.node.children].sort((a, b) => {
    const aIsFolder = !a.type || a.type === 'folder'
    const bIsFolder = !b.type || b.type === 'folder'
    if (aIsFolder && !bIsFolder) return -1
    if (!aIsFolder && bIsFolder) return 1
    return a.title.localeCompare(b.title)
  })
})

const nodeIcon = computed(() => {
  const t = props.node.type
  if (!t || t === 'folder') return 'Folder'
  if (t === 'doc') return 'Document'
  if (t === 'sheet') return 'Grid'
  if (t === 'slide') return 'Monitor'
  if (t === 'mindnote') return 'Connection'
  if (t === 'image') return 'Picture'
  if (t === 'file') return 'Paperclip'
  return 'Document'
})

const nodeIconColor = computed(() => {
  const t = props.node.type
  if (!t || t === 'folder') return props.iconColor || '#f5a623'
  if (t === 'doc') return '#3370ff'
  if (t === 'sheet') return '#36b37e'
  if (t === 'slide') return '#ff7d00'
  if (t === 'mindnote') return '#9254de'
  if (t === 'image') return '#00b8d9'
  if (t === 'file') return '#8592a6'
  return '#3370ff'
})

const isSearchMatch = computed(() => {
  if (!props.searchKeyword) return false
  return props.node.title.toLowerCase().includes(props.searchKeyword.toLowerCase())
})

async function toggleExpand() {
  const wasExpanded = props.node.isExpanded
  props.node.isExpanded = !wasExpanded

  // 如果是展开且还没有加载子文件夹，触发加载
  if (!wasExpanded && !hasChildren.value && props.node.hasMore !== false) {
    // 设置加载状态，父组件会在加载完成后重置
    props.node.isLoading = true
    emit('load-children', props.node)
  }

  emit('toggle-expand', props.node)
}

function handleClick() {
  if (isFolder.value) {
    // 点击文件夹名称：仅选中，不展开
    emit('select', props.node)
  } else {
    // 点击文件：在新标签页打开
    window.open(`/doc/${props.node.id}`, '_blank')
    emit('open-doc', props.node.id)
  }
}

function handleDoubleClick() {
  if (isFolder.value) {
    // 双击文件夹：展开/折叠
    props.node.isExpanded = !props.node.isExpanded
    emit('toggle-expand', props.node)
  } else {
    // 双击文件：触发重命名
    emit('rename', props.node)
  }
}
</script>

<style scoped>
.sidebar-tree-node {
  user-select: none;
}
.sidebar-tree-item {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 8px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.15s;
  margin: 1px 4px;
  min-height: 32px;
}
.sidebar-tree-item:hover {
  background: rgba(0, 0, 0, 0.04);
}
.sidebar-tree-item.active {
  background: rgba(51, 112, 255, 0.1);
}
.sidebar-tree-item.active .tree-node-name {
  color: var(--kx-primary, #3370ff);
  font-weight: 500;
}
.tree-arrow {
  font-size: 12px;
  color: var(--kx-text-placeholder, #8f959e);
  transition: transform 0.2s;
  flex-shrink: 0;
  width: 18px;
  height: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  cursor: pointer;
}
.tree-arrow:hover {
  color: var(--kx-text-primary, #1f2329);
  background: rgba(0, 0, 0, 0.06);
}
.tree-arrow.expanded {
  transform: rotate(90deg);
}
.tree-arrow.invisible {
  visibility: hidden;
}
.tree-arrow.loading {
  animation: spin 1s linear infinite;
}
@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
.tree-arrow-placeholder {
  width: 18px;
  flex-shrink: 0;
}
.tree-node-icon {
  flex-shrink: 0;
}
.tree-node-name {
  flex: 1;
  font-size: 13px;
  color: var(--kx-text-primary, #1f2329);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.4;
}
.tree-node-name.match-highlight {
  background: rgba(255, 214, 0, 0.3);
  border-radius: 2px;
}
.tree-node-actions {
  display: none;
  margin-left: auto;
  cursor: pointer;
  color: var(--kx-text-placeholder, #8f959e);
  flex-shrink: 0;
  padding: 2px;
  border-radius: 4px;
}
.tree-node-actions:hover {
  color: var(--kx-text-primary, #1f2329);
  background: rgba(0, 0, 0, 0.06);
}
.sidebar-tree-item:hover .tree-node-actions {
  display: flex;
}
.sidebar-tree-children {
  /* Container for children */
}
.tree-loading-placeholder {
  padding: 8px 12px 8px 50px;
  font-size: 12px;
  color: var(--kx-text-placeholder);
}
/* Expand transition */
.tree-expand-enter-active,
.tree-expand-leave-active {
  transition: all 0.2s ease;
  overflow: hidden;
}
.tree-expand-enter-from,
.tree-expand-leave-to {
  opacity: 0;
}
.is-loading {
  animation: spin 1s linear infinite;
}
</style>
