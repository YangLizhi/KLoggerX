<template>
  <div class="slide-toolbar-wrapper">
    <!-- Top Menu Bar -->
    <div class="top-menu-bar">
      <div class="menu-left">
        <el-button text @click="goBack"><el-icon><ArrowLeft /></el-icon></el-button>
        <input
          :value="title"
          class="title-input"
          placeholder="无标题幻灯片"
          @input="$emit('update:title', ($event.target as HTMLInputElement).value)"
          @blur="$emit('saveTitle')"
        />
      </div>
      <div class="menu-center">
        <span
          v-for="menu in menus"
          :key="menu.key"
          class="menu-item"
          :class="{ active: activeMenu === menu.key }"
          @click="activeMenu = menu.key"
        >
          {{ menu.label }}
        </span>
      </div>
      <div class="menu-right">
        <el-button size="small" text @click="emitAction('comment')"><el-icon><ChatDotRound /></el-icon>评论</el-button>
        <el-button size="small" text @click="emitAction('permission')"><el-icon><Lock /></el-icon>权限</el-button>
        <el-button size="small" text @click="emitAction('share')"><el-icon><Share /></el-icon>分享</el-button>
        <span class="save-status">{{ saveStatus }}</span>
        <el-button size="small" text @click="exportPPT"><el-icon><Download /></el-icon>导出</el-button>
        <el-button size="small" text @click="printSlides"><el-icon><Printer /></el-icon>打印</el-button>
      </div>
    </div>

    <!-- Toolbar Content -->
    <div class="toolbar-content">
      <!-- 开始 菜单 -->
      <template v-if="activeMenu === 'home'">
        <!-- 剪贴板 -->
        <div class="toolbar-section">
          <div class="section-label">剪贴板</div>
          <div class="toolbar-group">
            <el-button size="small" @click="emitAction('cut')"><el-icon><Scissors /></el-icon></el-button>
            <el-button size="small" @click="emitAction('copy')"><el-icon><CopyDocument /></el-icon></el-button>
            <el-button size="small" @click="emitAction('paste')"><el-icon><DocumentCopy /></el-icon></el-button>
          </div>
        </div>

        <span class="toolbar-divider" />

        <!-- 幻灯片 -->
        <div class="toolbar-section">
          <div class="section-label">幻灯片</div>
          <div class="toolbar-group">
            <el-button size="small" @click="addSlide"><el-icon><Plus /></el-icon>新建</el-button>
            <el-button size="small" @click="duplicateSlide" :disabled="!hasSelection"><el-icon><CopyDocument /></el-icon>复制</el-button>
            <el-button size="small" @click="deleteSlide" :disabled="!canDelete"><el-icon><Delete /></el-icon>删除</el-button>
            <el-button size="small" @click="moveUp" :disabled="!canMoveUp"><el-icon><Top /></el-icon></el-button>
            <el-button size="small" @click="moveDown" :disabled="!canMoveDown"><el-icon><Bottom /></el-icon></el-button>
          </div>
        </div>

        <span class="toolbar-divider" />

        <!-- 字体 -->
        <div class="toolbar-section">
          <div class="section-label">字体</div>
          <div class="toolbar-group">
            <el-select v-model="fontFamily" size="small" style="width: 100px" @change="onFontFamilyChange">
              <el-option label="微软雅黑" value="Microsoft YaHei" />
              <el-option label="宋体" value="SimSun" />
              <el-option label="黑体" value="SimHei" />
              <el-option label="楷体" value="KaiTi" />
              <el-option label="Arial" value="Arial" />
              <el-option label="Times New Roman" value="Times New Roman" />
            </el-select>
            <el-select v-model="fontSize" size="small" style="width: 70px" @change="onFontSizeChange">
              <el-option v-for="size in fontSizes" :key="size" :label="size + 'px'" :value="size" />
            </el-select>
            <div class="toolbar-btn-group">
              <el-button size="small" :type="boldActive ? 'primary' : 'default'" @click="toggleBold">
                <el-icon><Bold /></el-icon>
              </el-button>
              <el-button size="small" :type="italicActive ? 'primary' : 'default'" @click="toggleItalic">
                <el-icon><Italic /></el-icon>
              </el-button>
              <el-button size="small" :type="underlineActive ? 'primary' : 'default'" @click="toggleUnderline">
                <el-icon><Underline /></el-icon>
              </el-button>
            </div>
            <el-color-picker v-model="textColor" size="small" title="文字颜色" @change="onTextColorChange" />
            <el-color-picker v-model="bgColor" size="small" title="背景颜色" @change="onBgColorChange" />
          </div>
        </div>

        <span class="toolbar-divider" />

        <!-- 段落 -->
        <div class="toolbar-section">
          <div class="section-label">段落</div>
          <div class="toolbar-group">
            <el-button size="small" @click="emitAction('alignLeft')"><el-icon><AlignLeft /></el-icon></el-button>
            <el-button size="small" @click="emitAction('alignCenter')"><el-icon><AlignCenter /></el-icon></el-button>
            <el-button size="small" @click="emitAction('alignRight')"><el-icon><AlignRight /></el-icon></el-button>
            <el-button size="small" @click="emitAction('bulletList')"><el-icon><List /></el-icon></el-button>
            <el-button size="small" @click="emitAction('numberList')"><el-icon><More /></el-icon></el-button>
          </div>
        </div>
      </template>

      <!-- 插入 菜单 -->
      <template v-else-if="activeMenu === 'insert'">
        <div class="toolbar-section">
          <div class="section-label">幻灯片</div>
          <div class="toolbar-group">
            <el-button size="small" @click="addSlide"><el-icon><Plus /></el-icon>新建幻灯片</el-button>
          </div>
        </div>

        <span class="toolbar-divider" />

        <div class="toolbar-section">
          <div class="section-label">表格</div>
          <div class="toolbar-group">
            <el-dropdown trigger="click" @command="onInsertTable">
              <el-button size="small"><el-icon><Grid /></el-icon>表格</el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item :command="3">3 x 3</el-dropdown-item>
                  <el-dropdown-item :command="4">4 x 4</el-dropdown-item>
                  <el-dropdown-item :command="5">5 x 5</el-dropdown-item>
                  <el-dropdown-item :command="6">6 x 6</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>

        <span class="toolbar-divider" />

        <div class="toolbar-section">
          <div class="section-label">图像</div>
          <div class="toolbar-group">
            <el-button size="small" @click="insertImage"><el-icon><Picture /></el-icon>图片</el-button>
            <el-button size="small" @click="insertOnlineImage"><el-icon><Link /></el-icon>在线图片</el-button>
          </div>
        </div>

        <span class="toolbar-divider" />

        <div class="toolbar-section">
          <div class="section-label">形状</div>
          <div class="toolbar-group">
            <el-dropdown trigger="click" @command="onInsertShape">
              <el-button size="small"><el-icon><Share /></el-icon>形状</el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="rect">矩形</el-dropdown-item>
                  <el-dropdown-item command="circle">圆形</el-dropdown-item>
                  <el-dropdown-item command="triangle">三角形</el-dropdown-item>
                  <el-dropdown-item command="arrow">箭头</el-dropdown-item>
                  <el-dropdown-item command="line">直线</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>

        <span class="toolbar-divider" />

        <div class="toolbar-section">
          <div class="section-label">文本</div>
          <div class="toolbar-group">
            <el-button size="small" @click="insertTextBox"><el-icon><EditPen /></el-icon>文本框</el-button>
            <el-button size="small" @click="insertLink"><el-icon><Link /></el-icon>链接</el-button>
          </div>
        </div>

        <span class="toolbar-divider" />

        <div class="toolbar-section">
          <div class="section-label">媒体</div>
          <div class="toolbar-group">
            <el-button size="small" @click="insertVideo"><el-icon><VideoPlay /></el-icon>视频</el-button>
            <el-button size="small" @click="insertAudio"><el-icon><Headset /></el-icon>音频</el-button>
          </div>
        </div>
      </template>

      <!-- 设计 菜单 -->
      <template v-else-if="activeMenu === 'design'">
        <div class="toolbar-section">
          <div class="section-label">主题</div>
          <div class="toolbar-group theme-list">
            <div
              v-for="theme in themes"
              :key="theme.name"
              class="theme-item"
              :class="{ active: currentTheme === theme.name }"
              :style="{ background: theme.bg }"
              @click="applyTheme(theme.name)"
            >
              <span :style="{ color: theme.text }">A</span>
            </div>
          </div>
        </div>

        <span class="toolbar-divider" />

        <div class="toolbar-section">
          <div class="section-label">幻灯片大小</div>
          <div class="toolbar-group">
            <el-select v-model="slideRatio" size="small" style="width: 120px" @change="onRatioChange">
              <el-option label="16:9 宽屏" value="16:9" />
              <el-option label="4:3 标准" value="4:3" />
              <el-option label="16:10" value="16:10" />
            </el-select>
          </div>
        </div>

        <span class="toolbar-divider" />

        <div class="toolbar-section">
          <div class="section-label">背景</div>
          <div class="toolbar-group">
            <el-color-picker v-model="bgColor" size="small" title="纯色背景" @change="onBgColorChange" />
            <el-button size="small" @click="applyGradientBg"><el-icon><Share /></el-icon>渐变</el-button>
            <el-button size="small" @click="applyImageBg"><el-icon><Picture /></el-icon>图片</el-button>
          </div>
        </div>
      </template>

      <!-- 动画 菜单 -->
      <template v-else-if="activeMenu === 'animation'">
        <div class="toolbar-section">
          <div class="section-label">切换效果</div>
          <div class="toolbar-group">
            <el-select v-model="transition" size="small" style="width: 120px" @change="onTransitionChange">
              <el-option label="无" value="none" />
              <el-option label="淡入淡出" value="fade" />
              <el-option label="推进" value="push" />
              <el-option label="擦除" value="wipe" />
              <el-option label="分割" value="split" />
              <el-option label="翻转" value="flip" />
              <el-option label="缩放" value="zoom" />
            </el-select>
          </div>
        </div>

        <span class="toolbar-divider" />

        <div class="toolbar-section">
          <div class="section-label">切换时间</div>
          <div class="toolbar-group">
            <el-input-number v-model="transitionDuration" size="small" :min="0.1" :max="5" :step="0.1" style="width: 100px" @change="onDurationChange" />
            <span class="unit-label">秒</span>
          </div>
        </div>

        <span class="toolbar-divider" />

        <div class="toolbar-section">
          <div class="section-label">入场动画</div>
          <div class="toolbar-group">
            <el-select v-model="entranceAnim" size="small" style="width: 120px" @change="onEntranceAnimChange">
              <el-option label="无" value="none" />
              <el-option label="飞入" value="fly-in" />
              <el-option label="淡入" value="fade-in" />
              <el-option label="缩放" value="zoom-in" />
              <el-option label="旋转" value="spin" />
              <el-option label="弹跳" value="bounce" />
            </el-select>
          </div>
        </div>
      </template>

      <!-- 放映 菜单 -->
      <template v-else-if="activeMenu === 'slideshow'">
        <div class="toolbar-section">
          <div class="section-label">开始放映</div>
          <div class="toolbar-group">
            <el-button type="primary" size="small" @click="playFromStart"><el-icon><VideoPlay /></el-icon>从头开始</el-button>
            <el-button size="small" @click="playFromCurrent" :disabled="!hasSelection"><el-icon><VideoPlay /></el-icon>从当前页</el-button>
          </div>
        </div>

        <span class="toolbar-divider" />

        <div class="toolbar-section">
          <div class="section-label">放映设置</div>
          <div class="toolbar-group">
            <el-checkbox v-model="autoPlay" @change="onAutoPlayChange">自动播放</el-checkbox>
            <el-input-number
              v-model="autoPlayInterval"
              size="small"
              :min="1"
              :max="60"
              :disabled="!autoPlay"
              style="width: 80px; margin-left: 8px"
            />
            <span class="unit-label">秒/页</span>
          </div>
        </div>

        <span class="toolbar-divider" />

        <div class="toolbar-section">
          <div class="section-label">演讲者工具</div>
          <div class="toolbar-group">
            <el-checkbox v-model="showSpeakerNotes">显示演讲者备注</el-checkbox>
            <el-checkbox v-model="showTimer">显示计时器</el-checkbox>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'

const props = defineProps<{
  slideCount: number
  currentIndex: number
  title?: string
  saveStatus?: string
}>()

const emit = defineEmits(['action', 'update:title', 'saveTitle'])

const router = useRouter()

// Menu structure
const menus = [
  { key: 'home', label: '开始' },
  { key: 'insert', label: '插入' },
  { key: 'design', label: '设计' },
  { key: 'animation', label: '动画' },
  { key: 'slideshow', label: '放映' },
]

const activeMenu = ref('home')

// Font settings
const fontFamily = ref('Microsoft YaHei')
const fontSize = ref(18)
const fontSizes = [12, 14, 16, 18, 20, 24, 28, 32, 36, 42, 48, 56, 64, 72]
const boldActive = ref(false)
const italicActive = ref(false)
const underlineActive = ref(false)
const textColor = ref('#1f2329')
const bgColor = ref('#ffffff')

// Design settings
const slideRatio = ref('16:9')
const currentTheme = ref('default')
const themes = [
  { name: 'default', bg: '#ffffff', text: '#1f2329' },
  { name: 'dark', bg: '#1f2329', text: '#ffffff' },
  { name: 'blue', bg: '#3370ff', text: '#ffffff' },
  { name: 'green', bg: '#36b37e', text: '#ffffff' },
  { name: 'orange', bg: '#ff7d00', text: '#ffffff' },
  { name: 'purple', bg: '#8b5cf6', text: '#ffffff' },
  { name: 'gradient-blue', bg: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)', text: '#ffffff' },
  { name: 'gradient-green', bg: 'linear-gradient(135deg, #11998e 0%, #38ef7d 100%)', text: '#ffffff' },
]

// Animation settings
const transition = ref('fade')
const transitionDuration = ref(0.5)
const entranceAnim = ref('fade-in')

// Slideshow settings
const autoPlay = ref(false)
const autoPlayInterval = ref(5)
const showSpeakerNotes = ref(false)
const showTimer = ref(true)

// Computed
const hasSelection = computed(() => props.currentIndex >= 0)
const canDelete = computed(() => props.slideCount > 1 && props.currentIndex >= 0)
const canMoveUp = computed(() => props.currentIndex > 0)
const canMoveDown = computed(() => props.currentIndex >= 0 && props.currentIndex < props.slideCount - 1)

// Helper
function emitAction(action: string, params?: any) {
  emit('action', { action, params })
}

// Navigation
function goBack() {
  router.back()
}

// Home menu actions
function addSlide() { emitAction('addSlide') }
function duplicateSlide() { emitAction('duplicateSlide') }
function deleteSlide() { emitAction('deleteSlide') }
function moveUp() { emitAction('moveUp') }
function moveDown() { emitAction('moveDown') }

// Font actions
function onFontFamilyChange(val: string) { emitAction('setFontFamily', val) }
function onFontSizeChange(val: number) { emitAction('setFontSize', val) }
function toggleBold() {
  boldActive.value = !boldActive.value
  emitAction('toggleBold', boldActive.value)
}
function toggleItalic() {
  italicActive.value = !italicActive.value
  emitAction('toggleItalic', italicActive.value)
}
function toggleUnderline() {
  underlineActive.value = !underlineActive.value
  emitAction('toggleUnderline', underlineActive.value)
}
function onTextColorChange(val: string) { emitAction('setTextColor', val) }
function onBgColorChange(val: string) { emitAction('setBgColor', val) }

// Insert actions
function insertImage() { emitAction('insertImage') }
function insertOnlineImage() { emitAction('insertOnlineImage') }
function insertTextBox() { emitAction('insertTextBox') }
function insertLink() { emitAction('insertLink') }
function insertVideo() { emitAction('insertVideo') }
function insertAudio() { emitAction('insertAudio') }
function onInsertTable(size: number) { emitAction('insertTable', size) }
function onInsertShape(shape: string) { emitAction('insertShape', shape) }

// Design actions
function applyTheme(themeName: string) {
  currentTheme.value = themeName
  emitAction('applyTheme', themeName)
}
function onRatioChange(val: string) { emitAction('setRatio', val) }
function applyGradientBg() { emitAction('applyGradientBg') }
function applyImageBg() { emitAction('applyImageBg') }

// Animation actions
function onTransitionChange(val: string) { emitAction('setTransition', val) }
function onDurationChange(val: number) { emitAction('setTransitionDuration', val) }
function onEntranceAnimChange(val: string) { emitAction('setEntranceAnim', val) }

// Slideshow actions
function playFromStart() { emitAction('playFromStart') }
function playFromCurrent() { emitAction('playFromCurrent') }
function onAutoPlayChange(val: boolean) { emitAction('setAutoPlay', val) }

// Export and print
function exportPPT() { emitAction('exportPPT') }
function printSlides() { emitAction('printSlides') }
</script>

<style scoped>
.slide-toolbar-wrapper {
  display: flex;
  flex-direction: column;
  background: #fafafa;
  border-bottom: 1px solid var(--kx-border, #e5e6eb);
}

.top-menu-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 16px;
  height: 48px;
  background: #fff;
  border-bottom: 1px solid var(--kx-border, #e5e6eb);
}

.menu-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.title-input {
  border: none;
  outline: none;
  font-size: 15px;
  font-weight: 500;
  color: #1f2329;
  background: transparent;
  width: 200px;
}

.title-input::placeholder {
  color: #bbbfc4;
}

.menu-center {
  display: flex;
  gap: 4px;
  flex: 1;
  justify-content: center;
}

.menu-item {
  padding: 6px 16px;
  font-size: 13px;
  color: #646a73;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.2s;
}

.menu-item:hover {
  background: rgba(0, 0, 0, 0.04);
}

.menu-item.active {
  color: #3370ff;
  font-weight: 500;
}

.menu-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  justify-content: flex-end;
}

.save-status {
  font-size: 12px;
  color: #8f959e;
}

.toolbar-content {
  display: flex;
  align-items: flex-start;
  padding: 8px 16px;
  gap: 4px;
  flex-wrap: wrap;
  min-height: 44px;
}

.toolbar-section {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.section-label {
  font-size: 11px;
  color: #8f959e;
}

.toolbar-group {
  display: flex;
  align-items: center;
  gap: 4px;
}

.toolbar-btn-group {
  display: flex;
  gap: 0;
}

.toolbar-btn-group .el-button {
  border-radius: 0;
}

.toolbar-btn-group .el-button:first-child {
  border-radius: 4px 0 0 4px;
}

.toolbar-btn-group .el-button:last-child {
  border-radius: 0 4px 4px 0;
}

.toolbar-divider {
  width: 1px;
  height: 32px;
  background: var(--kx-border, #e5e6eb);
  margin: 0 8px;
  flex-shrink: 0;
}

.unit-label {
  font-size: 12px;
  color: #646a73;
  margin-left: 4px;
}

.theme-list {
  display: flex;
  gap: 8px;
}

.theme-item {
  width: 32px;
  height: 32px;
  border-radius: 4px;
  border: 2px solid #e5e6eb;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 14px;
  transition: all 0.2s;
}

.theme-item:hover {
  transform: scale(1.1);
}

.theme-item.active {
  border-color: #3370ff;
  box-shadow: 0 0 0 2px rgba(51, 112, 255, 0.2);
}
</style>
