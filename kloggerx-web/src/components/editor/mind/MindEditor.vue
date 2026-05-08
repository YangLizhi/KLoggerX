<template>
  <div class="mind-editor">
    <MindToolbar
      :has-selection="!!selectedNode"
      :is-root="selectedNode?.id === nodes[0]?.id"
      @action="handleToolbarAction"
    />
    <div class="mind-canvas" ref="canvasRef">
      <svg class="mind-lines" :width="canvasWidth" :height="canvasHeight">
        <line
          v-for="(line, i) in lines"
          :key="i"
          :x1="line.x1"
          :y1="line.y1"
          :x2="line.x2"
          :y2="line.y2"
          stroke="#c4c6cf"
          stroke-width="2"
        />
      </svg>
      <div
        v-for="node in flatNodes"
        :key="node.id"
        class="mind-node"
        :class="{ selected: selectedNode?.id === node.id, root: node.depth === 0 }"
        :style="getNodeStyle(node)"
        @click="selectedNode = node"
        @dblclick="startEdit(node)"
      >
        <input
          v-if="editingId === node.id"
          class="node-edit-input"
          :value="node.text"
          @input="onNodeInput(node, ($event.target as HTMLInputElement).value)"
          @blur="editingId = null"
          @keyup.enter="editingId = null"
          ref="editInputRef"
          autofocus
        />
        <span v-else class="node-text">{{ node.text || $t('editor.mind.newNode') }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import MindToolbar from './MindToolbar.vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

interface MindNode {
  id: string
  text: string
  color?: string
  children: MindNode[]
  // layout computed
  x?: number
  y?: number
  depth?: number
}

interface FlatNode {
  id: string
  text: string
  color?: string
  children: MindNode[]
  x: number
  y: number
  depth: number
  parentId: string | null
}

const props = defineProps<{
  documentId: number
  content: string
}>()

const emit = defineEmits<{
  (e: 'save', content: string): void
}>()

// const canvasRef = ref<HTMLElement>()
const nodes = ref<MindNode[]>([])
const selectedNode = ref<FlatNode | null>(null)
const editingId = ref<string | null>(null)
const layoutType = ref<'right' | 'left' | 'both'>('right')
let saveTimer: ReturnType<typeof setTimeout> | null = null
let idCounter = 0

function genId(): string {
  return `node_${Date.now()}_${idCounter++}`
}

function initData() {
  if (props.content) {
    try {
      const parsed = JSON.parse(props.content)
      if (Array.isArray(parsed) && parsed.length > 0) {
        nodes.value = parsed
        if (parsed[0]?.layoutType) {
          layoutType.value = parsed[0].layoutType
        }
        return
      }
    } catch {}
  }
  nodes.value = [{
    id: genId(),
    text: t('editor.mind.centralTopic'),
    color: '#3370ff',
    children: [
      { id: genId(), text: `${t('editor.mind.branch')}1`, children: [] },
      { id: genId(), text: `${t('editor.mind.branch')}2`, children: [] },
      { id: genId(), text: `${t('editor.mind.branch')}3`, children: [] },
    ],
  }]
}

const NODE_W = 120
const NODE_H = 36
const H_GAP = 60
const V_GAP = 16

function layoutTree(node: MindNode, depth: number, startY: number): { nodes: FlatNode[]; height: number } {
  const x = depth * (NODE_W + H_GAP) + 40
  const result: FlatNode[] = []

  if (!node.children || node.children.length === 0) {
    result.push({
      ...node,
      x,
      y: startY,
      depth,
      parentId: null,
    })
    return { nodes: result, height: NODE_H }
  }

  let totalChildHeight = 0
  const childLayouts: { nodes: FlatNode[]; height: number }[] = []

  for (const child of node.children) {
    const layout = layoutTree(child, depth + 1, startY + totalChildHeight)
    childLayouts.push(layout)
    totalChildHeight += layout.height + V_GAP
  }
  totalChildHeight -= V_GAP

  const nodeY = startY + totalChildHeight / 2 - NODE_H / 2

  result.push({
    ...node,
    x,
    y: nodeY,
    depth,
    parentId: null,
  })

  for (const layout of childLayouts) {
    for (const n of layout.nodes) {
      if (n.depth === depth + 1 && node.children.find((c) => c.id === n.id)) {
        n.parentId = node.id
      }
      result.push(n)
    }
  }

  return { nodes: result, height: Math.max(totalChildHeight, NODE_H) }
}

const flatNodes = computed<FlatNode[]>(() => {
  if (!nodes.value.length) return []
  const layout = layoutTree(nodes.value[0], 0, 40)
  return layout.nodes
})

const lines = computed(() => {
  const result: { x1: number; y1: number; x2: number; y2: number }[] = []
  const nodeMap = new Map<string, FlatNode>()
  for (const n of flatNodes.value) {
    nodeMap.set(n.id, n)
  }
  for (const n of flatNodes.value) {
    if (n.parentId) {
      const parent = nodeMap.get(n.parentId)
      if (parent) {
        result.push({
          x1: parent.x + NODE_W,
          y1: parent.y + NODE_H / 2,
          x2: n.x,
          y2: n.y + NODE_H / 2,
        })
      }
    }
  }
  return result
})

const canvasWidth = computed(() => {
  const maxX = Math.max(...flatNodes.value.map((n) => n.x + NODE_W), 800)
  return maxX + 100
})

const canvasHeight = computed(() => {
  const maxY = Math.max(...flatNodes.value.map((n) => n.y + NODE_H), 600)
  return maxY + 100
})

function getNodeStyle(node: FlatNode): Record<string, string> {
  const style: Record<string, string> = {
    left: `${node.x}px`,
    top: `${node.y}px`,
  }
  if (node.depth === 0) {
    style.backgroundColor = node.color || '#3370ff'
    style.borderColor = node.color || '#3370ff'
  } else if (node.color) {
    style.borderColor = node.color
  }
  return style
}

function findNodeById(root: MindNode[], id: string): MindNode | null {
  for (const n of root) {
    if (n.id === id) return n
    const found = findNodeById(n.children, id)
    if (found) return found
  }
  return null
}

function findParentById(root: MindNode[], id: string): MindNode | null {
  for (const n of root) {
    if (n.children.some((c) => c.id === id)) return n
    const found = findParentById(n.children, id)
    if (found) return found
  }
  return null
}

function handleToolbarAction(event: { action: string; params?: any }) {
  switch (event.action) {
    case 'addChild':
      addChild()
      break
    case 'addSibling':
      addSibling()
      break
    case 'removeNode':
      removeNode()
      break
    case 'editNode':
      if (selectedNode.value) startEdit(selectedNode.value)
      break
    case 'setColor':
      setNodeColor(event.params)
      break
    case 'setLayout':
      layoutType.value = event.params
      scheduleSave()
      break
    case 'expandAll':
    case 'collapseAll':
      ElMessage.info(t('editor.mind.expandCollapseDev'))
      break
    case 'exportImage':
      ElMessage.info(t('editor.mind.exportImageDev'))
      break
  }
}

function addChild() {
  if (!selectedNode.value) return
  const node = findNodeById(nodes.value, selectedNode.value.id)
  if (node) {
    node.children.push({ id: genId(), text: t('editor.mind.newNode'), children: [] })
    scheduleSave()
  }
}

function addSibling() {
  if (!selectedNode.value || selectedNode.value.id === nodes.value[0]?.id) return
  const parent = findParentById(nodes.value, selectedNode.value.id)
  if (parent) {
    const idx = parent.children.findIndex((c) => c.id === selectedNode.value!.id)
    parent.children.splice(idx + 1, 0, { id: genId(), text: t('editor.mind.newNode'), children: [] })
    scheduleSave()
  }
}

function removeNode() {
  if (!selectedNode.value || selectedNode.value.id === nodes.value[0]?.id) return
  const parent = findParentById(nodes.value, selectedNode.value.id)
  if (parent) {
    parent.children = parent.children.filter((c) => c.id !== selectedNode.value!.id)
    selectedNode.value = null
    scheduleSave()
  }
}

function setNodeColor(color: string) {
  if (!selectedNode.value) return
  const node = findNodeById(nodes.value, selectedNode.value.id)
  if (node) {
    node.color = color
    scheduleSave()
  }
}

function startEdit(node: FlatNode) {
  editingId.value = node.id
  nextTick(() => {
    const inputs = document.querySelectorAll('.node-edit-input')
    if (inputs.length) (inputs[inputs.length - 1] as HTMLInputElement).focus()
  })
}

function onNodeInput(node: FlatNode, val: string) {
  const n = findNodeById(nodes.value, node.id)
  if (n) {
    n.text = val
    scheduleSave()
  }
}

function scheduleSave() {
  if (saveTimer) clearTimeout(saveTimer)
  saveTimer = setTimeout(() => {
    const dataToSave = [...nodes.value]
    if (dataToSave[0]) {
      (dataToSave[0] as any).layoutType = layoutType.value
    }
    emit('save', JSON.stringify(dataToSave))
  }, 2000)
}

onMounted(() => initData())
watch(() => props.content, () => initData(), { once: true })
</script>

<style scoped>
.mind-editor {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}
.mind-canvas {
  flex: 1;
  overflow: auto;
  position: relative;
  background: #fafbfc;
}
.mind-lines {
  position: absolute;
  top: 0;
  left: 0;
  pointer-events: none;
}
.mind-node {
  position: absolute;
  width: 120px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  border: 1px solid #dee0e3;
  border-radius: 6px;
  cursor: pointer;
  user-select: none;
  transition: box-shadow 0.15s;
  font-size: 13px;
}
.mind-node:hover {
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.08);
}
.mind-node.selected {
  border-color: #3370ff;
  box-shadow: 0 0 0 2px rgba(51, 112, 255, 0.2);
}
.mind-node.root {
  background: #3370ff;
  color: #fff;
  border-color: #3370ff;
  font-weight: 600;
}
.node-text {
  padding: 0 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.node-edit-input {
  width: calc(100% - 8px);
  border: none;
  outline: none;
  text-align: center;
  font-size: 13px;
  background: transparent;
}
.mind-node.root .node-edit-input {
  color: #fff;
}
</style>
