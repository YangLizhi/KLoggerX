<template>
  <div class="slide-toolbar">
    <div class="toolbar-group">
      <el-button size="small" @click="addSlide"><el-icon><Plus /></el-icon>添加幻灯片</el-button>
      <el-button size="small" @click="duplicateSlide" :disabled="!hasSelection"><el-icon><CopyDocument /></el-icon>复制</el-button>
      <el-button size="small" @click="deleteSlide" :disabled="!canDelete"><el-icon><Delete /></el-icon>删除</el-button>
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-button size="small" @click="moveUp" :disabled="!canMoveUp"><el-icon><Top /></el-icon>上移</el-button>
      <el-button size="small" @click="moveDown" :disabled="!canMoveDown"><el-icon><Bottom /></el-icon>下移</el-button>
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-select v-model="slideLayout" size="small" placeholder="布局" style="width: 120px" @change="onLayoutChange">
        <el-option label="标题+内容" value="title-content" />
        <el-option label="仅标题" value="title-only" />
        <el-option label="两栏" value="two-column" />
        <el-option label="空白" value="blank" />
      </el-select>
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-color-picker v-model="bgColor" size="small" title="背景颜色" @change="onBgColorChange" />
      <el-color-picker v-model="textColor" size="small" title="文字颜色" @change="onTextColorChange" />
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-button size="small" @click="insertImage"><el-icon><Picture /></el-icon>插入图片</el-button>
      <el-button size="small" @click="insertShape"><el-icon><Grid /></el-icon>插入形状</el-button>
    </div>
    <span class="toolbar-divider" />
    <div class="toolbar-group">
      <el-button size="small" @click="playFromStart"><el-icon><VideoPlay /></el-icon>放映</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const props = defineProps<{
  slideCount: number
  currentIndex: number
}>()

const emit = defineEmits(['action'])

const slideLayout = ref('title-content')
const bgColor = ref('#ffffff')
const textColor = ref('#1f2329')

const hasSelection = computed(() => props.currentIndex >= 0)
const canDelete = computed(() => props.slideCount > 1 && props.currentIndex >= 0)
const canMoveUp = computed(() => props.currentIndex > 0)
const canMoveDown = computed(() => props.currentIndex >= 0 && props.currentIndex < props.slideCount - 1)

function emitAction(action: string, params?: any) {
  emit('action', { action, params })
}

function addSlide() { emitAction('addSlide') }
function duplicateSlide() { emitAction('duplicateSlide') }
function deleteSlide() { emitAction('deleteSlide') }
function moveUp() { emitAction('moveUp') }
function moveDown() { emitAction('moveDown') }
function onLayoutChange(val: string) { emitAction('setLayout', val) }
function onBgColorChange(val: string) { emitAction('setBgColor', val) }
function onTextColorChange(val: string) { emitAction('setTextColor', val) }
function insertImage() { emitAction('insertImage') }
function insertShape() { emitAction('insertShape') }
function playFromStart() { emitAction('play') }
</script>

<style scoped>
.slide-toolbar {
  display: flex;
  align-items: center;
  padding: 6px 16px;
  border-bottom: 1px solid var(--kx-border, #e5e6eb);
  background: #fafafa;
  gap: 4px;
  flex-wrap: wrap;
}
.toolbar-group {
  display: flex;
  align-items: center;
  gap: 4px;
}
.toolbar-divider {
  width: 1px;
  height: 20px;
  background: var(--kx-border, #e5e6eb);
  margin: 0 8px;
}
</style>
