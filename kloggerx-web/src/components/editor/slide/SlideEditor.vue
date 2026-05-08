<template>
  <div class="slide-editor">
    <SlideToolbar
      :slide-count="slides.length"
      :current-index="currentIndex"
      :title="docTitle"
      :save-status="saveStatus"
      @action="handleToolbarAction"
      @update:title="updateTitle"
      @save-title="saveTitle"
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
              {{ slide.title || $t('editor.slide.blankSlide') }}
            </div>
          </div>
        </div>
      </div>
      <div class="slide-main">
        <div class="slide-canvas-wrapper">
          <div
            class="slide-canvas"
            v-if="currentSlide"
            :style="canvasStyle"
            :class="{ 'presentation-mode': isPresenting }"
          >
            <!-- Slide content -->
            <div class="slide-content" :style="{ color: currentSlide.textColor || '#1f2329' }">
              <input
                class="slide-title-input"
                v-model="currentSlide.title"
                :placeholder="$t('editor.slide.clickInputTitle')"
                @input="scheduleSave"
                :style="titleStyle"
              />
              <textarea
                class="slide-body-input"
                v-model="currentSlide.body"
                :placeholder="$t('editor.slide.clickInputContent')"
                @input="scheduleSave"
                :style="bodyStyle"
              />
            </div>

            <!-- Shapes and elements -->
            <div
              v-for="(element, idx) in currentSlide.elements"
              :key="idx"
              class="slide-element"
              :style="element.style"
              @click="selectElement(idx)"
            >
              <img v-if="element.type === 'image'" :src="element.src" style="width: 100%; height: 100%; object-fit: contain;" />
              <div v-else-if="element.type === 'text'" v-html="element.content"></div>
              <div v-else-if="element.type === 'shape'" :style="shapeStyle(element)"></div>
            </div>
          </div>
        </div>

        <!-- Presentation controls -->
        <div v-if="isPresenting" class="presentation-controls">
          <el-button @click="prevSlide" :disabled="currentIndex === 0"><el-icon><ArrowLeft /></el-icon>{{ $t('editor.slide.prevSlide') }}</el-button>
          <span class="slide-counter">{{ currentIndex + 1 }} / {{ slides.length }}</span>
          <el-button @click="nextSlide" :disabled="currentIndex >= slides.length - 1">{{ $t('editor.slide.nextSlide') }}<el-icon><ArrowRight /></el-icon></el-button>
          <el-button type="danger" @click="exitPresentation">{{ $t('editor.slide.exitPresentation') }}</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import SlideToolbar from './SlideToolbar.vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

interface SlideElement {
  type: 'image' | 'text' | 'shape' | 'table' | 'video' | 'audio'
  style: Record<string, string>
  content?: string
  src?: string
  shape?: string
}

interface Slide {
  title: string
  body: string
  bgColor?: string
  textColor?: string
  layout?: string
  fontFamily?: string
  fontSize?: number
  fontWeight?: string
  fontStyle?: string
  textDecoration?: string
  transition?: string
  transitionDuration?: number
  entranceAnim?: string
  elements?: SlideElement[]
}

const props = defineProps<{
  documentId: number
  content: string
  docTitle?: string
}>()

const emit = defineEmits<{
  (e: 'save', content: string): void
  (e: 'update:title', title: string): void
  (e: 'saveTitle'): void
}>()

const slides = ref<Slide[]>([])
const currentIndex = ref(0)
const isPresenting = ref(false)
const selectedElement = ref(-1)
const saveStatus = ref(t('editor.slide.saved'))

let saveTimer: ReturnType<typeof setTimeout> | null = null
let presentationTimer: ReturnType<typeof setInterval> | null = null

function updateTitle(title: string) {
  emit('update:title', title)
}

function saveTitle() {
  emit('saveTitle')
}

const currentSlide = computed({
  get: () => slides.value[currentIndex.value],
  set: (val) => { slides.value[currentIndex.value] = val }
})

const canvasStyle = computed(() => ({
  backgroundColor: currentSlide.value?.bgColor || '#fff',
  fontFamily: currentSlide.value?.fontFamily || 'Microsoft YaHei',
}))

const titleStyle = computed(() => ({
  fontSize: (currentSlide.value?.fontSize || 28) + 'px',
  fontWeight: currentSlide.value?.fontWeight || '700',
  fontStyle: currentSlide.value?.fontStyle || 'normal',
  textDecoration: currentSlide.value?.textDecoration || 'none',
}))

const bodyStyle = computed(() => ({
  fontSize: ((currentSlide.value?.fontSize || 28) * 0.6) + 'px',
  fontWeight: currentSlide.value?.fontWeight || '400',
  fontStyle: currentSlide.value?.fontStyle || 'normal',
  textDecoration: currentSlide.value?.textDecoration || 'none',
}))

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
          fontFamily: s.fontFamily || 'Microsoft YaHei',
          fontSize: s.fontSize || 28,
          fontWeight: s.fontWeight,
          fontStyle: s.fontStyle,
          textDecoration: s.textDecoration,
          transition: s.transition || 'fade',
          transitionDuration: s.transitionDuration || 0.5,
          entranceAnim: s.entranceAnim || 'fade-in',
          elements: s.elements || [],
        }))
        return
      }
    } catch {}
  }
  slides.value = [{
    title: '', body: '', bgColor: '#ffffff', textColor: '#1f2329',
    layout: 'title-content', fontFamily: 'Microsoft YaHei', fontSize: 28,
    transition: 'fade', transitionDuration: 0.5, entranceAnim: 'fade-in', elements: []
  }]
}

function handleToolbarAction(event: { action: string; params?: any }) {
  const slide = currentSlide.value
  if (!slide) return

  switch (event.action) {
    // Slide operations
    case 'addSlide': addSlide(); break
    case 'duplicateSlide': duplicateSlide(); break
    case 'deleteSlide': removeSlide(); break
    case 'moveUp': moveSlide(-1); break
    case 'moveDown': moveSlide(1); break

    // Clipboard
    case 'cut': ElMessage.info(t('editor.slide.cutDev')); break
    case 'copy': ElMessage.info(t('editor.slide.copyDev')); break
    case 'paste': ElMessage.info(t('editor.slide.pasteDev')); break

    // Font
    case 'setFontFamily':
      slide.fontFamily = event.params
      scheduleSave()
      break
    case 'setFontSize':
      slide.fontSize = event.params
      scheduleSave()
      break
    case 'toggleBold':
      slide.fontWeight = event.params ? 'bold' : 'normal'
      scheduleSave()
      break
    case 'toggleItalic':
      slide.fontStyle = event.params ? 'italic' : 'normal'
      scheduleSave()
      break
    case 'toggleUnderline':
      slide.textDecoration = event.params ? 'underline' : 'none'
      scheduleSave()
      break
    case 'setTextColor':
      slide.textColor = event.params
      scheduleSave()
      break
    case 'setBgColor':
      slide.bgColor = event.params
      scheduleSave()
      break

    // Alignment
    case 'alignLeft':
    case 'alignCenter':
    case 'alignRight':
      ElMessage.info(t('editor.slide.alignDev'))
      break
    case 'bulletList':
    case 'numberList':
      ElMessage.info(t('editor.slide.listDev'))
      break

    // Insert
    case 'insertImage': insertImage(); break
    case 'insertOnlineImage': insertOnlineImage(); break
    case 'insertTextBox': insertTextBox(); break
    case 'insertLink': insertLink(); break
    case 'insertVideo': insertMedia('video'); break
    case 'insertAudio': insertMedia('audio'); break
    case 'insertTable': insertTable(event.params); break
    case 'insertShape': insertShape(event.params); break

    // Design
    case 'applyTheme': applyTheme(event.params); break
    case 'setRatio': ElMessage.info(t('editor.slide.ratioDev')); break
    case 'applyGradientBg': applyGradientBg(); break
    case 'applyImageBg': applyImageBg(); break

    // Animation
    case 'setTransition':
      slide.transition = event.params
      scheduleSave()
      break
    case 'setTransitionDuration':
      slide.transitionDuration = event.params
      scheduleSave()
      break
    case 'setEntranceAnim':
      slide.entranceAnim = event.params
      scheduleSave()
      break

    // Slideshow
    case 'playFromStart':
      currentIndex.value = 0
      startPresentation()
      break
    case 'playFromCurrent':
      startPresentation()
      break
    case 'setAutoPlay':
      if (event.params) {
        startAutoPlay(5)
      } else {
        stopAutoPlay()
      }
      break

    // UI actions (handled by parent)
    case 'comment': ElMessage.info(t('editor.slide.commentHint')); break
    case 'permission': ElMessage.info(t('editor.slide.permissionHint')); break
    case 'share': ElMessage.info(t('editor.slide.commentHint')); break

    // Export
    case 'exportPPT': exportPPT(); break
    case 'printSlides': printSlides(); break
  }
}

// Slide operations
function addSlide() {
  slides.value.push({
    title: '', body: '', bgColor: '#ffffff', textColor: '#1f2329',
    layout: 'title-content', fontFamily: 'Microsoft YaHei', fontSize: 28,
    transition: 'fade', transitionDuration: 0.5, entranceAnim: 'fade-in', elements: []
  })
  currentIndex.value = slides.value.length - 1
  scheduleSave()
}

function duplicateSlide() {
  if (currentIndex.value < 0) return
  const copy = JSON.parse(JSON.stringify(slides.value[currentIndex.value]))
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

// Insert operations
async function insertImage() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/*'
  input.onchange = async (e) => {
    const file = (e.target as HTMLInputElement).files?.[0]
    if (file) {
      const reader = new FileReader()
      reader.onload = (e) => {
        addElement('image', {
          src: e.target?.result as string,
          style: { top: '50%', left: '50%', width: '200px', height: '150px', transform: 'translate(-50%, -50%)' }
        })
      }
      reader.readAsDataURL(file)
    }
  }
  input.click()
}

async function insertOnlineImage() {
  try {
    const { value } = await ElMessageBox.prompt(t('editor.slide.enterImageUrl'), t('editor.slide.insertOnlineImageTitle'), {
      inputPlaceholder: 'https://example.com/image.png',
      confirmButtonText: t('editor.slide.insert'),
      cancelButtonText: t('editor.slide.cancel'),
    })
    if (value) {
      addElement('image', {
        src: value,
        style: { top: '50%', left: '50%', width: '200px', height: '150px', transform: 'translate(-50%, -50%)' }
      })
    }
  } catch {}
}

function insertTextBox() {
  addElement('text', {
    content: t('editor.slide.doubleClickEdit'),
    style: { top: '50%', left: '50%', width: '200px', padding: '8px', transform: 'translate(-50%, -50%)', border: '1px dashed #ccc' }
  })
}

async function insertLink() {
  try {
    const { value } = await ElMessageBox.prompt(t('editor.slide.enterLinkUrl'), t('editor.slide.insertLinkTitle'), {
      inputPlaceholder: 'https://example.com',
      confirmButtonText: t('editor.slide.insert'),
      cancelButtonText: t('editor.slide.cancel'),
    })
    if (value) {
      if (currentSlide.value) {
        currentSlide.value.body += `\n[${t('editor.slide.link')}](${value})`
        scheduleSave()
      }
    }
  } catch {}
}

async function insertMedia(type: 'video' | 'audio') {
  try {
    const title = type === 'video' ? t('editor.slide.insertVideoTitle') : t('editor.slide.insertAudioTitle')
    const { value } = await ElMessageBox.prompt(t('editor.slide.enterMediaUrl'), title, {
      inputPlaceholder: type === 'video' ? 'https://example.com/video.mp4' : 'https://example.com/audio.mp3',
      confirmButtonText: t('editor.slide.insert'),
      cancelButtonText: t('editor.slide.cancel'),
    })
    if (value) {
      addElement(type, {
        src: value,
        style: { top: '50%', left: '50%', width: type === 'video' ? '300px' : '200px', transform: 'translate(-50%, -50%)' }
      })
    }
  } catch {}
}

function insertTable(size: number) {
  let html = '<table style="width: 100%; border-collapse: collapse;">'
  for (let i = 0; i < size; i++) {
    html += '<tr>'
    for (let j = 0; j < size; j++) {
      html += '<td style="border: 1px solid #ccc; padding: 8px; text-align: center;">-</td>'
    }
    html += '</tr>'
  }
  html += '</table>'
  addElement('text', {
    content: html,
    style: { top: '50%', left: '50%', width: '300px', transform: 'translate(-50%, -50%)' }
  })
}

function insertShape(shape: string) {
  const shapeStyles: Record<string, Record<string, string>> = {
    rect: { width: '100px', height: '60px', background: '#3370ff', borderRadius: '4px' },
    circle: { width: '80px', height: '80px', background: '#36b37e', borderRadius: '50%' },
    triangle: { width: '0', height: '0', borderLeft: '40px solid transparent', borderRight: '40px solid transparent', borderBottom: '70px solid #ff7d00', background: 'transparent' },
    arrow: { width: '0', height: '0', borderTop: '20px solid transparent', borderBottom: '20px solid transparent', borderLeft: '40px solid #8b5cf6', background: 'transparent' },
    line: { width: '100px', height: '2px', background: '#1f2329' },
  }
  addElement('shape', {
    shape,
    style: { top: '50%', left: '50%', transform: 'translate(-50%, -50%)', ...shapeStyles[shape] }
  })
}

function addElement(type: SlideElement['type'], data: Partial<SlideElement>) {
  if (!currentSlide.value.elements) {
    currentSlide.value.elements = []
  }
  currentSlide.value.elements.push({ type, ...data } as SlideElement)
  scheduleSave()
}

function selectElement(idx: number) {
  selectedElement.value = idx
}

function shapeStyle(element: SlideElement): Record<string, string> {
  return element.style || {}
}

// Design operations
function applyTheme(themeName: string) {
  const themeMap: Record<string, { bg: string; text: string }> = {
    'default': { bg: '#ffffff', text: '#1f2329' },
    'dark': { bg: '#1f2329', text: '#ffffff' },
    'blue': { bg: '#3370ff', text: '#ffffff' },
    'green': { bg: '#36b37e', text: '#ffffff' },
    'orange': { bg: '#ff7d00', text: '#ffffff' },
    'purple': { bg: '#8b5cf6', text: '#ffffff' },
    'gradient-blue': { bg: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)', text: '#ffffff' },
    'gradient-green': { bg: 'linear-gradient(135deg, #11998e 0%, #38ef7d 100%)', text: '#ffffff' },
  }
  const theme = themeMap[themeName]
  if (theme && currentSlide.value) {
    currentSlide.value.bgColor = theme.bg
    currentSlide.value.textColor = theme.text
    scheduleSave()
  }
}

function applyGradientBg() {
  if (currentSlide.value) {
    currentSlide.value.bgColor = 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)'
    scheduleSave()
  }
}

async function applyImageBg() {
  try {
    const { value } = await ElMessageBox.prompt(t('editor.slide.enterBgImageUrl'), t('editor.slide.setBgImageTitle'), {
      inputPlaceholder: 'https://example.com/bg.jpg',
      confirmButtonText: t('editor.slide.set'),
      cancelButtonText: t('editor.slide.cancel'),
    })
    if (value && currentSlide.value) {
      currentSlide.value.bgColor = `url(${value}) center/cover`
      scheduleSave()
    }
  } catch {}
}

// Presentation
function startPresentation() {
  isPresenting.value = true
  document.body.style.overflow = 'hidden'
}

function exitPresentation() {
  isPresenting.value = false
  document.body.style.overflow = ''
  stopAutoPlay()
}

function prevSlide() {
  if (currentIndex.value > 0) {
    currentIndex.value--
  }
}

function nextSlide() {
  if (currentIndex.value < slides.value.length - 1) {
    currentIndex.value++
  }
}

function startAutoPlay(interval: number) {
  stopAutoPlay()
  presentationTimer = setInterval(() => {
    if (currentIndex.value < slides.value.length - 1) {
      currentIndex.value++
    } else {
      stopAutoPlay()
    }
  }, interval * 1000)
}

function stopAutoPlay() {
  if (presentationTimer) {
    clearInterval(presentationTimer)
    presentationTimer = null
  }
}

// Export
function exportPPT() {
  const content = JSON.stringify(slides.value, null, 2)
  const blob = new Blob([content], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `presentation-${Date.now()}.json`
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success(t('editor.slide.exportSuccess'))
}

function printSlides() {
  window.print()
}

// Keyboard shortcuts
function handleKeydown(e: KeyboardEvent) {
  if (isPresenting.value) {
    if (e.key === 'Escape') {
      exitPresentation()
    } else if (e.key === 'ArrowRight' || e.key === ' ' || e.key === 'Enter') {
      nextSlide()
    } else if (e.key === 'ArrowLeft') {
      prevSlide()
    }
  }
}

function scheduleSave() {
  saveStatus.value = t('editor.slide.saving')
  if (saveTimer) clearTimeout(saveTimer)
  saveTimer = setTimeout(() => {
    emit('save', JSON.stringify(slides.value))
    saveStatus.value = t('editor.slide.saved')
  }, 2000)
}

onMounted(() => {
  initData()
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
  stopAutoPlay()
})

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
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #e8e9eb;
  padding: 20px;
}

.slide-canvas-wrapper {
  width: 100%;
  max-width: 800px;
  display: flex;
  justify-content: center;
}

.slide-canvas {
  width: 100%;
  aspect-ratio: 16 / 9;
  background: #fff;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  position: relative;
  overflow: hidden;
}

.slide-canvas.presentation-mode {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  aspect-ratio: auto;
  z-index: 2000;
  border-radius: 0;
}

.slide-content {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
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
  text-align: center;
  background: transparent;
  width: 100%;
}

.slide-title-input::placeholder {
  color: #bbbfc4;
}

.slide-body-input {
  border: none;
  outline: none;
  font-size: 16px;
  flex: 1;
  resize: none;
  text-align: center;
  background: transparent;
  line-height: 1.8;
}

.slide-body-input::placeholder {
  color: #bbbfc4;
}

.slide-element {
  position: absolute;
  cursor: move;
}

.presentation-controls {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 16px;
  background: rgba(0, 0, 0, 0.7);
  padding: 12px 24px;
  border-radius: 8px;
  z-index: 2001;
}

.slide-counter {
  color: #fff;
  font-size: 14px;
}

@media print {
  .slide-toolbar-wrapper,
  .slide-sidebar,
  .presentation-controls {
    display: none !important;
  }
  .slide-main {
    background: #fff;
    padding: 0;
  }
  .slide-canvas {
    box-shadow: none;
    page-break-after: always;
  }
}
</style>
