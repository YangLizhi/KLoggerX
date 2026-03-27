<template>
  <div class="slide-editor">
    <SlideToolbar
      :slide-count="slides.length"
      :current-index="currentIndex"
      @action="handleToolbarAction"
    />
    <div class="slide-main-area">
      <div class="slide-sidebar">
        <div class="slide-list">
          <div
            v-for="(slide, index) in slides"
            :key="index"
            class="slide-thumb"
            :class="{ active: currentIndex === index }"
            @click="currentIndex = index"
          >
            <span class="slide-num">{{ index + 1 }}</span>
            <div class="thumb-preview" :style="{ backgroundColor: slide.bgColor || '#fff' }">
              {{ slide.title || '空白幻灯片' }}
            </div>
          </div>
        </div>
      </div>
      <div class="slide-main">
        <div class="slide-canvas" v-if="currentSlide" :style="{ backgroundColor: currentSlide.bgColor || '#fff' }">
          <input
            class="slide-title-input"
            v-model="currentSlide.title"
            placeholder="点击输入标题"
            @input="scheduleSave"
            :style="{ color: currentSlide.textColor || '#1f2329' }"
          />
          <textarea
            class="slide-body-input"
            v-model="currentSlide.body"
            placeholder="点击输入内容"
            @input="scheduleSave"
            :style="{ color: currentSlide.textColor || '#1f2329' }"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import SlideToolbar from './SlideToolbar.vue'
import { ElMessageBox, ElMessage } from 'element-plus'

interface Slide {
  title: string
  body: string
  bgColor?: string
  textColor?: string
  layout?: string
}

const props = defineProps<{
  documentId: number
  content: string
}>()

const emit = defineEmits<{
  (e: 'save', content: string): void
}>()

const slides = ref<Slide[]>([])
const currentIndex = ref(0)
const currentSlide = computed({
  get: () => slides.value[currentIndex.value],
  set: (val) => { slides.value[currentIndex.value] = val }
})
let saveTimer: ReturnType<typeof setTimeout> | null = null

function initData() {
  if (props.content) {
    try {
      const parsed = JSON.parse(props.content)
      if (Array.isArray(parsed) && parsed.length > 0) {
        slides.value = parsed.map(s => ({
          title: s.title || '',
          body: s.body || '',
          bgColor: s.bgColor || '#ffffff',
          textColor: s.textColor || '#1f2329',
          layout: s.layout || 'title-content',
        }))
        return
      }
    } catch {}
  }
  slides.value = [{ title: '', body: '', bgColor: '#ffffff', textColor: '#1f2329', layout: 'title-content' }]
}

function handleToolbarAction(event: { action: string; params?: any }) {
  switch (event.action) {
    case 'addSlide':
      addSlide()
      break
    case 'duplicateSlide':
      duplicateSlide()
      break
    case 'deleteSlide':
      removeSlide()
      break
    case 'moveUp':
      moveSlide(-1)
      break
    case 'moveDown':
      moveSlide(1)
      break
    case 'setLayout':
      if (currentSlide.value) {
        currentSlide.value.layout = event.params
        scheduleSave()
      }
      break
    case 'setBgColor':
      if (currentSlide.value) {
        currentSlide.value.bgColor = event.params
        scheduleSave()
      }
      break
    case 'setTextColor':
      if (currentSlide.value) {
        currentSlide.value.textColor = event.params
        scheduleSave()
      }
      break
    case 'insertImage':
      insertImage()
      break
    case 'insertShape':
      ElMessage.info('形状功能开发中')
      break
    case 'play':
      startPresentation()
      break
  }
}

function addSlide() {
  slides.value.push({
    title: '',
    body: '',
    bgColor: '#ffffff',
    textColor: '#1f2329',
    layout: 'title-content',
  })
  currentIndex.value = slides.value.length - 1
  scheduleSave()
}

function duplicateSlide() {
  if (currentIndex.value < 0) return
  const copy = { ...slides.value[currentIndex.value] }
  slides.value.splice(currentIndex.value + 1, 0, copy)
  currentIndex.value += 1
  scheduleSave()
}

function removeSlide() {
  if (slides.value.length <= 1) return
  slides.value.splice(currentIndex.value, 1)
  if (currentIndex.value >= slides.value.length) {
    currentIndex.value = slides.value.length - 1
  }
  scheduleSave()
}

function moveSlide(direction: number) {
  const newIndex = currentIndex.value + direction
  if (newIndex < 0 || newIndex >= slides.value.length) return
  const temp = slides.value[currentIndex.value]
  slides.value[currentIndex.value] = slides.value[newIndex]
  slides.value[newIndex] = temp
  currentIndex.value = newIndex
  scheduleSave()
}

async function insertImage() {
  try {
    const { value } = await ElMessageBox.prompt('请输入图片URL', '插入图片', {
      inputPlaceholder: 'https://example.com/image.png',
      confirmButtonText: '插入',
      cancelButtonText: '取消',
    })
    if (value && currentSlide.value) {
      // For simplicity, we'll just append the image markdown to the body
      currentSlide.value.body += `\n![图片](${value})`
      scheduleSave()
    }
  } catch {
    // cancelled
  }
}

function startPresentation() {
  ElMessage.info('幻灯片放映功能开发中')
}

function scheduleSave() {
  if (saveTimer) clearTimeout(saveTimer)
  saveTimer = setTimeout(() => {
    emit('save', JSON.stringify(slides.value))
  }, 2000)
}

onMounted(() => initData())
watch(() => props.content, () => initData(), { once: true })
</script>

<style scoped>
.slide-editor {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}
.slide-main-area {
  flex: 1;
  display: flex;
  overflow: hidden;
}
.slide-sidebar {
  width: 180px;
  background: #f5f6f7;
  border-right: 1px solid #e5e6eb;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}
.slide-list {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.slide-thumb {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
}
.slide-thumb.active {
  background: rgba(51, 112, 255, 0.08);
}
.slide-thumb:hover {
  background: rgba(0, 0, 0, 0.04);
}
.slide-num {
  font-size: 11px;
  color: #646a73;
  flex-shrink: 0;
  width: 16px;
  text-align: right;
}
.thumb-preview {
  background: #fff;
  border: 1px solid #dee0e3;
  border-radius: 4px;
  padding: 8px;
  font-size: 11px;
  color: #1f2329;
  width: 120px;
  min-height: 68px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.slide-main {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #e8e9eb;
  padding: 40px;
}
.slide-canvas {
  width: 100%;
  max-width: 720px;
  aspect-ratio: 16 / 9;
  background: #fff;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  padding: 40px;
  gap: 24px;
}
.slide-title-input {
  border: none;
  outline: none;
  font-size: 28px;
  font-weight: 700;
  color: #1f2329;
  text-align: center;
  background: transparent;
}
.slide-title-input::placeholder {
  color: #bbbfc4;
}
.slide-body-input {
  border: none;
  outline: none;
  font-size: 16px;
  color: #1f2329;
  flex: 1;
  resize: none;
  text-align: center;
  background: transparent;
  line-height: 1.8;
}
.slide-body-input::placeholder {
  color: #bbbfc4;
}
</style>
