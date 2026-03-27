<template>
  <div class="folder-tree-node">
    <div
      class="folder-tree-item"
      :class="{ active: selectedFolderId === node.id }"
      :style="{ paddingLeft: `${level * 20 + 12}px` }"
      @click="handleClick"
      @dblclick="handleDoubleClick"
      @contextmenu.prevent="$emit('contextmenu', $event, node)"
    >
      <el-icon
        class="expand-arrow"
        :class="{ expanded: node.isExpanded, invisible: !hasChildren }"
        @click.stop="toggleExpand"
      >
        <ArrowRight />
      </el-icon>
      <el-icon color="#f5a623" :size="15"><Folder /></el-icon>
      <span class="folder-tree-name">{{ node.title }}</span>
      <span class="folder-actions" @click.stop="$emit('contextmenu', $event, node)">
        <el-icon :size="14"><MoreFilled /></el-icon>
      </span>
    </div>
    <transition name="tree-expand">
      <div v-show="node.isExpanded && hasChildren" class="folder-tree-children">
        <FolderTreeNode
          v-for="child in sortedChildren"
          :key="child.id"
          :node="child"
          :level="level + 1"
          :selected-folder-id="selectedFolderId"
          @select="$emit('select', $event)"
          @contextmenu="(event: MouseEvent, n: any) => $emit('contextmenu', event, n)"
          @toggle-expand="$emit('toggle-expand', $event)"
        />
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

defineOptions({ name: 'FolderTreeNode' })

interface FolderNode {
  id: number
  title: string
  type?: string
  parentId?: number | null
  isPinned?: boolean
  isFavorite?: boolean
  children?: FolderNode[]
  isExpanded?: boolean
  [key: string]: any
}

const props = defineProps<{
  node: FolderNode
  level: number
  selectedFolderId: number | null
}>()

const emit = defineEmits<{
  (e: 'select', node: FolderNode): void
  (e: 'contextmenu', event: MouseEvent, node: FolderNode): void
  (e: 'toggle-expand', node: FolderNode): void
}>()

const hasChildren = computed(() => props.node.children && props.node.children.length > 0)

const sortedChildren = computed(() => {
  if (!props.node.children) return []
  return [...props.node.children].sort((a, b) => a.title.localeCompare(b.title))
})

function handleClick() {
  // 点击文件夹名称：仅选中，不展开
  emit('select', props.node)
}

function handleDoubleClick() {
  // 双击文件夹：展开/折叠
  props.node.isExpanded = !props.node.isExpanded
  emit('toggle-expand', props.node)
}

function toggleExpand() {
  // 点击三角图标：展开/折叠
  props.node.isExpanded = !props.node.isExpanded
  emit('toggle-expand', props.node)
}
</script>

<style scoped>
.folder-tree-node {
  user-select: none;
}
.folder-tree-item {
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
.folder-tree-item:hover {
  background: rgba(0, 0, 0, 0.04);
}
.folder-tree-item.active {
  background: rgba(51, 112, 255, 0.1);
}
.folder-tree-item.active .folder-tree-name {
  color: var(--kx-primary, #3370ff);
  font-weight: 500;
}
.expand-arrow {
  transition: transform 0.2s;
  font-size: 12px;
  color: var(--kx-text-placeholder, #8f959e);
  flex-shrink: 0;
  width: 18px;
  height: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  cursor: pointer;
}
.expand-arrow:hover {
  color: var(--kx-text-primary, #1f2329);
  background: rgba(0, 0, 0, 0.06);
}
.expand-arrow.expanded {
  transform: rotate(90deg);
}
.expand-arrow.invisible {
  visibility: hidden;
}
.folder-tree-name {
  font-size: 13px;
  color: var(--kx-text-primary, #1f2329);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  line-height: 1.4;
}
.folder-actions {
  display: none;
  margin-left: auto;
  cursor: pointer;
  color: var(--kx-text-placeholder, #8f959e);
  flex-shrink: 0;
  padding: 2px;
  border-radius: 4px;
}
.folder-actions:hover {
  color: var(--kx-text-primary, #1f2329);
  background: rgba(0, 0, 0, 0.06);
}
.folder-tree-item:hover .folder-actions {
  display: flex;
}
.folder-tree-children {
  /* Children container */
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
</style>
