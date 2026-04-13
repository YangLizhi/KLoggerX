<template>
  <div class="image-editor">
    <div class="image-toolbar">
      <div class="toolbar-group">
        <el-button size="small" @click="zoomIn"><el-icon><ZoomIn /></el-icon>放大</el-button>
        <el-button size="small" @click="zoomOut"><el-icon><ZoomOut /></el-icon>缩小</el-button>
        <el-button size="small" @click="resetZoom"><el-icon><RefreshRight /></el-icon>重置</el-button>
      </div>
      <span class="toolbar-divider" />
      <div class="toolbar-group">
        <el-button size="small" @click="rotateLeft"><el-icon><RefreshLeft /></el-icon>左旋转</el-button>
        <el-button size="small" @click="rotateRight"><el-icon><RefreshRight /></el-icon>右旋转</el-button>
      </div>
      <span class="toolbar-divider" />
      <div class="toolbar-group">
        <el-button size="small" @click="downloadImage"><el-icon><Download /></el-icon>下载</el-button>
        <el-button size="small" @click="copyImage"><el-icon><CopyDocument /></el-icon>复制</el-button>
      </div>
      <div class="toolbar-info">
        <span v-if="imageInfo">{{ imageInfo.width }} x {{ imageInfo.height }}</span>
        <span v-if="scale !== 100" style="margin-left: 8px">{{ scale }}%</span>
      </div>
    </div>
    <div class="image-container" ref="containerRef" @wheel="handleWheel">
      <div class="image-wrapper" :style="imageWrapperStyle">
        <img
          v-if="imageUrl"
          :src="imageUrl"
          :style="imageStyle"
          @load="onImageLoad"
          @error="onImageError"
          draggable="false"
        />
        <div v-else class="image-placeholder">
          <el-icon :size="48"><Picture /></el-icon>
          <p>无法加载图片</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'

interface ImageInfo {
  width: number
  height: number
  name?: string
}

const props = defineProps<{
  documentId: number
  content: string
}>()

const emit = defineEmits<{
  (e: 'save', content: string): void
}>()

// const containerRef = ref<HTMLElement>()
const imageUrl = ref('')
const imageInfo = ref<ImageInfo | null>(null)
const scale = ref(100)
const rotation = ref(0)
const loading = ref(true)

const imageWrapperStyle = computed(() => ({
  transform: `scale(${scale.value / 100}) rotate(${rotation.value}deg)`,
  transition: 'transform 0.2s ease'
}))

const imageStyle = computed(() => ({
  maxWidth: '100%',
  maxHeight: '100%',
  objectFit: 'contain' as const
}))

function initData() {
  if (props.content) {
    try {
      const parsed = JSON.parse(props.content)
      if (parsed.url) {
        imageUrl.value = parsed.url
        imageInfo.value = {
          width: parsed.width || 0,
          height: parsed.height || 0,
          name: parsed.name
        }
      }
    } catch {}
  }
  loading.value = false
}

function onImageLoad(e: Event) {
  const img = e.target as HTMLImageElement
  if (!imageInfo.value?.width) {
    imageInfo.value = {
      width: img.naturalWidth,
      height: img.naturalHeight
    }
  }
  loading.value = false
}

function onImageError() {
  loading.value = false
  ElMessage.error('图片加载失败')
}

function zoomIn() {
  if (scale.value < 500) {
    scale.value = Math.min(500, scale.value + 25)
  }
}

function zoomOut() {
  if (scale.value > 25) {
    scale.value = Math.max(25, scale.value - 25)
  }
}

function resetZoom() {
  scale.value = 100
  rotation.value = 0
}

function rotateLeft() {
  rotation.value -= 90
}

function rotateRight() {
  rotation.value += 90
}

function handleWheel(e: WheelEvent) {
  if (e.ctrlKey) {
    e.preventDefault()
    if (e.deltaY < 0) {
      zoomIn()
    } else {
      zoomOut()
    }
  }
}

async function downloadImage() {
  if (!imageUrl.value) return
  try {
    const response = await fetch(imageUrl.value)
    const blob = await response.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = imageInfo.value?.name || 'image.png'
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success('下载成功')
  } catch {
    ElMessage.error('下载失败')
  }
}

async function copyImage() {
  if (!imageUrl.value) return
  try {
    const response = await fetch(imageUrl.value)
    const blob = await response.blob()
    await navigator.clipboard.write([
      new ClipboardItem({ [blob.type]: blob })
    ])
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

onMounted(() => initData())
watch(() => props.content, () => initData(), { once: true })
</script>

<style scoped>
.image-editor {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #1a1a1a;
}
.image-toolbar {
  display: flex;
  align-items: center;
  padding: 8px 16px;
  background: #2d2d2d;
  border-bottom: 1px solid #404040;
  gap: 4px;
}
.toolbar-group {
  display: flex;
  align-items: center;
  gap: 4px;
}
.toolbar-divider {
  width: 1px;
  height: 20px;
  background: #404040;
  margin: 0 8px;
}
.toolbar-info {
  margin-left: auto;
  color: #aaa;
  font-size: 12px;
}
.image-container {
  flex: 1;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}
.image-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
}
.image-wrapper img {
  max-width: 100%;
  max-height: 100%;
}
.image-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #666;
}
.image-placeholder p {
  margin-top: 12px;
}
</style>
