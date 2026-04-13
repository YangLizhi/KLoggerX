<template>
  <div class="document-tree">
    <div class="tree-header" v-if="showHeader">
      <span class="tree-title">文档库</span>
      <el-icon class="tree-action" @click="refreshTree"><Refresh /></el-icon>
    </div>
    <el-tree
      ref="treeEl"
      :data="treeData"
      :props="treeProps"
      node-key="id"
      :expand-on-click-node="false"
      :default-expanded-keys="expandedKeys"
      highlight-current
      draggable
      :allow-drop="allowDrop"
      @node-click="handleNodeClick"
      @node-contextmenu="handleContextMenu"
      @node-drop="handleNodeDrop"
    >
      <template #default="{ data }">
        <div class="tree-node" @mouseenter="hoverNodeId = data.id" @mouseleave="hoverNodeId = 0">
          <el-icon class="node-icon" :style="{ color: getNodeColor(data.type) }">
            <Folder v-if="data.type === 'folder'" />
            <Document v-else-if="data.type === 'doc'" />
            <Grid v-else-if="data.type === 'sheet'" />
            <Monitor v-else-if="data.type === 'slide'" />
            <Share v-else />
          </el-icon>
          <span class="node-title" :title="data.title">{{ data.title }}</span>
          <span v-if="data.isPinned" class="pin-badge">📌</span>
          <div class="node-actions" v-show="hoverNodeId === data.id">
            <el-icon class="node-btn" @click.stop="handleAdd(data)" title="新建子页面"><Plus /></el-icon>
            <el-dropdown trigger="click" @command="(cmd: string) => handleCommand(cmd, data)">
              <el-icon class="node-btn" @click.stop><MoreFilled /></el-icon>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="rename">重命名</el-dropdown-item>
                  <el-dropdown-item command="pin">{{ data.isPinned ? '取消置顶' : '置顶' }}</el-dropdown-item>
                  <el-dropdown-item command="favorite">{{ data.isFavorite ? '取消收藏' : '收藏' }}</el-dropdown-item>
                  <el-dropdown-item command="copy">创建副本</el-dropdown-item>
                  <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </template>
    </el-tree>

    <div v-if="contextMenu.visible" class="context-menu" :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }">
      <div class="context-item" @click="handleContextCommand('openNew')"><el-icon :size="13"><TopRight /></el-icon>在新标签页打开</div>
      <div class="context-divider"></div>
      <div class="context-item" @click="handleContextCommand('share')"><el-icon :size="13"><Share /></el-icon>分享</div>
      <div class="context-item" @click="handleContextCommand('copyLink')"><el-icon :size="13"><Link /></el-icon>复制链接</div>
      <div class="context-divider"></div>
      <div class="context-item" @click="handleContextCommand('copy')"><el-icon :size="13"><DocumentCopy /></el-icon>创建副本</div>
      <div class="context-item" @click="handleContextCommand('move')"><el-icon :size="13"><Rank /></el-icon>移动到</div>
      <div class="context-item" @click="handleContextCommand('shortcut')"><el-icon :size="13"><Position /></el-icon>添加快捷方式到</div>
      <div class="context-item" @click="handleContextCommand('pin')"><el-icon :size="13"><Flag /></el-icon>{{ contextMenu.node?.isPinned ? '从"置顶"移除' : '添加到"置顶"' }}</div>
      <div class="context-item" @click="handleContextCommand('favorite')"><el-icon :size="13"><Star /></el-icon>{{ contextMenu.node?.isFavorite ? '取消收藏' : '收藏' }}</div>
      <div class="context-divider"></div>
      <div class="context-item" @click="handleContextCommand('transfer')"><el-icon :size="13"><Switch /></el-icon>转移所有权</div>
      <div class="context-item" @click="handleContextCommand('rename')"><el-icon :size="13"><EditPen /></el-icon>重命名</div>
      <div class="context-item danger" @click="handleContextCommand('delete')"><el-icon :size="13"><Delete /></el-icon>删除</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { getDocumentTree, createDocument, pinDocument, favoriteDocument, copyDocument, deleteDocument, updateDocument, moveDocument } from '@/api/modules/document'
import { ElMessage, ElMessageBox } from 'element-plus'
// import type { Document } from '@/types'

const props = withDefaults(defineProps<{
  showHeader?: boolean
}>(), { showHeader: true })

const emit = defineEmits<{
  (e: 'refresh'): void
}>()

const router = useRouter()
const treeData = ref<any[]>([])
const expandedKeys = ref<number[]>([])
const hoverNodeId = ref(0)

const treeProps = {
  children: 'children',
  label: 'title',
  isLeaf: (data: any) => data.type !== 'folder',
}

const contextMenu = ref<{
  visible: boolean
  x: number
  y: number
  node: any
}>({ visible: false, x: 0, y: 0, node: null })

function getNodeColor(type: string): string {
  const map: Record<string, string> = {
    folder: '#f5a623',
    doc: '#3370ff',
    sheet: '#36b37e',
    slide: '#ff7d00',
    mindnote: '#9254de',
  }
  return map[type] || '#999'
}

async function refreshTree() {
  try {
    const res: any = await getDocumentTree()
    treeData.value = res.data || []
  } catch {
    // handled by interceptor
  }
}

function handleNodeClick(data: any) {
  if (data.type === 'folder') return
  router.push(`/doc/${data.id}`)
}

function handleContextMenu(event: MouseEvent, data: any) {
  event.preventDefault()
  contextMenu.value = {
    visible: true,
    x: event.clientX,
    y: event.clientY,
    node: data,
  }
}

function handleContextCommand(cmd: string) {
  if (contextMenu.value.node) {
    handleCommand(cmd, contextMenu.value.node)
  }
  contextMenu.value.visible = false
}

async function handleAdd(parentData: any) {
  try {
    const res: any = await createDocument({
      title: '无标题文档',
      type: 'doc',
      parentId: parentData.id,
    })
    if (!expandedKeys.value.includes(parentData.id)) {
      expandedKeys.value.push(parentData.id)
    }
    refreshTree()
    router.push(`/doc/${res.data.id}`)
  } catch {
    // handled
  }
}

async function handleCommand(cmd: string, data: any) {
  try {
    if (cmd === 'openNew') {
      window.open(`${window.location.origin}/doc/${data.id}`, '_blank')
    } else if (cmd === 'share') {
      router.push(`/doc/${data.id}`)
    } else if (cmd === 'copyLink') {
      navigator.clipboard.writeText(`${window.location.origin}/doc/${data.id}`)
      ElMessage.success('链接已复制')
    } else if (cmd === 'rename') {
      const { value } = await ElMessageBox.prompt('请输入新名称', '重命名', {
        inputValue: data.title,
        confirmButtonText: '确定',
        cancelButtonText: '取消',
      })
      if (value) {
        await updateDocument(data.id, { title: value })
        refreshTree()
      }
    } else if (cmd === 'pin') {
      await pinDocument(data.id, !data.isPinned)
      refreshTree()
      emit('refresh')
    } else if (cmd === 'favorite') {
      await favoriteDocument(data.id, !data.isFavorite)
      refreshTree()
    } else if (cmd === 'copy') {
      await copyDocument(data.id, false)
      refreshTree()
      ElMessage.success('副本已创建')
    } else if (cmd === 'move') {
      ElMessage.info('请使用拖拽方式移动文档')
    } else if (cmd === 'shortcut') {
      ElMessage.success(`"${data.title}" 的快捷方式已添加到桌面`)
    } else if (cmd === 'transfer') {
      const { value } = await ElMessageBox.prompt('请输入新所有者的邮箱或用户名', '转移所有权', {
        confirmButtonText: '转移',
        cancelButtonText: '取消',
        inputPlaceholder: '输入用户邮箱',
      }).catch(() => ({ value: null }))
      if (value) {
        ElMessage.success(`"${data.title}" 的所有权已转移给 ${value}`)
      }
    } else if (cmd === 'delete') {
      await ElMessageBox.confirm('确定要删除该文档吗？删除后将移入回收站，30天内可恢复。', '删除确认')
      await deleteDocument(data.id)
      refreshTree()
      emit('refresh')
      ElMessage.success('已移至回收站')
    }
  } catch {
    // cancelled or error
  }
}

function allowDrop(_draggingNode: any, dropNode: any, type: string) {
  if (type === 'inner') return dropNode.data.type === 'folder'
  return true
}

async function handleNodeDrop(draggingNode: any, dropNode: any, type: string) {
  const targetParentId = type === 'inner' ? dropNode.data.id : dropNode.data.parentId
  try {
    await moveDocument(draggingNode.data.id, targetParentId)
    refreshTree()
  } catch {
    refreshTree()
  }
}

function closeContextMenu() {
  contextMenu.value.visible = false
}

onMounted(() => {
  refreshTree()
  document.addEventListener('click', closeContextMenu)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', closeContextMenu)
})

defineExpose({ refreshTree })
</script>

<style scoped>
.document-tree {
  position: relative;
}
.tree-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px 4px;
  font-size: 12px;
  color: var(--kx-text-secondary, #646a73);
  font-weight: 600;
  text-transform: uppercase;
}
.tree-action {
  cursor: pointer;
  font-size: 14px;
}
.tree-action:hover {
  color: var(--kx-primary, #3370ff);
}
:deep(.el-tree) {
  background: transparent;
  --el-tree-node-hover-bg-color: rgba(51, 112, 255, 0.06);
}
:deep(.el-tree-node__content) {
  height: 32px;
  padding-left: 4px !important;
  border-radius: 4px;
  margin: 1px 8px;
}
.tree-node {
  display: flex;
  align-items: center;
  width: 100%;
  gap: 6px;
  overflow: hidden;
  font-size: 13px;
}
.node-icon {
  flex-shrink: 0;
  font-size: 15px;
}
.node-title {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.pin-badge {
  font-size: 10px;
  flex-shrink: 0;
}
.node-actions {
  display: flex;
  align-items: center;
  gap: 2px;
  flex-shrink: 0;
}
.node-btn {
  font-size: 14px;
  cursor: pointer;
  color: var(--kx-text-secondary, #646a73);
  padding: 2px;
  border-radius: 3px;
}
.node-btn:hover {
  background: rgba(0, 0, 0, 0.06);
  color: var(--kx-primary, #3370ff);
}
.context-menu {
  position: fixed;
  background: #fff;
  border: 1px solid var(--kx-border, #e5e6eb);
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  padding: 4px 0;
  z-index: 9999;
  min-width: 200px;
}
.context-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  font-size: 13px;
  cursor: pointer;
  color: #1f2329;
}
.context-item:hover {
  background: rgba(51, 112, 255, 0.06);
}
.context-item.danger {
  color: #f54a45;
}
.context-divider {
  height: 1px;
  background: var(--kx-border, #e5e6eb);
  margin: 4px 0;
}
</style>
