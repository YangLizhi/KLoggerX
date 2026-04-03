<template>
  <div class="sheet-toolbar-wrapper">
    <!-- Top Menu Bar -->
    <div class="top-menu-bar">
      <div class="menu-left">
        <el-button text @click="goBack"><el-icon><ArrowLeft /></el-icon></el-button>
        <input
          :value="title"
          class="title-input"
          placeholder="无标题表格"
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
        <el-button size="small" text @click="addComment"><el-icon><ChatDotRound /></el-icon>评论</el-button>
        <el-button size="small" text @click="shareSheet"><el-icon><Share /></el-icon>分享</el-button>
        <el-button size="small" text @click="protectSheet"><el-icon><Lock /></el-icon>权限</el-button>
        <span class="save-status">{{ saveStatus }}</span>
        <el-button size="small" text @click="printSheet"><el-icon><Printer /></el-icon>打印</el-button>
        <el-button size="small" text @click="exportSheet"><el-icon><Download /></el-icon>导出</el-button>
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
            <el-dropdown trigger="click" @command="onClipboardAction">
              <el-button size="small"><el-icon><DocumentCopy /></el-icon>粘贴</el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="paste">粘贴</el-dropdown-item>
                  <el-dropdown-item command="pasteValue">粘贴值</el-dropdown-item>
                  <el-dropdown-item command="pasteFormat">粘贴格式</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
            <el-button size="small" @click="emitAction('cut')"><el-icon><Scissors /></el-icon>剪切</el-button>
            <el-button size="small" @click="emitAction('copy')"><el-icon><CopyDocument /></el-icon>复制</el-button>
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
              <el-option label="Courier New" value="Courier New" />
            </el-select>
            <el-select v-model="fontSize" size="small" style="width: 65px" @change="onFontSizeChange">
              <el-option v-for="size in fontSizes" :key="size" :label="size" :value="size" />
            </el-select>
            <div class="toolbar-btn-group">
              <el-button size="small" :type="boldActive ? 'primary' : 'default'" @click="toggleBold">
                <strong>B</strong>
              </el-button>
              <el-button size="small" :type="italicActive ? 'primary' : 'default'" @click="toggleItalic">
                <em>I</em>
              </el-button>
              <el-button size="small" :type="underlineActive ? 'primary' : 'default'" @click="toggleUnderline">
                <u>U</u>
              </el-button>
              <el-button size="small" :type="strikeActive ? 'primary' : 'default'" @click="toggleStrike">
                <s>S</s>
              </el-button>
            </div>
            <el-color-picker v-model="textColor" size="small" title="文字颜色" @change="onTextColorChange" />
            <el-color-picker v-model="bgColor" size="small" title="填充颜色" @change="onBgColorChange" />
          </div>
        </div>

        <span class="toolbar-divider" />

        <!-- 对齐 -->
        <div class="toolbar-section">
          <div class="section-label">对齐</div>
          <div class="toolbar-group">
            <el-dropdown trigger="click" @command="onAlignChange">
              <el-button size="small">
                <el-icon><component :is="alignIcon" /></el-icon>
                <el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="left"><el-icon><AlignLeft /></el-icon> 左对齐</el-dropdown-item>
                  <el-dropdown-item command="center"><el-icon><AlignCenter /></el-icon> 居中</el-dropdown-item>
                  <el-dropdown-item command="right"><el-icon><AlignRight /></el-icon> 右对齐</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
            <div class="toolbar-btn-group">
              <el-button size="small" @click="emitAction('alignTop')" title="顶端对齐"><el-icon><Top /></el-icon></el-button>
              <el-button size="small" @click="emitAction('alignMiddle')" title="垂直居中"><el-icon><Minus /></el-icon></el-button>
              <el-button size="small" @click="emitAction('alignBottom')" title="底端对齐"><el-icon><Bottom /></el-icon></el-button>
            </div>
            <el-button size="small" @click="emitAction('wrapText')" title="自动换行"><el-icon><Document /></el-icon></el-button>
          </div>
        </div>

        <span class="toolbar-divider" />

        <!-- 数字格式 -->
        <div class="toolbar-section">
          <div class="section-label">数字</div>
          <div class="toolbar-group">
            <el-select v-model="numberFormat" size="small" style="width: 90px" @change="onNumberFormatChange">
              <el-option label="常规" value="normal" />
              <el-option label="数字" value="number" />
              <el-option label="货币" value="currency" />
              <el-option label="百分比" value="percent" />
              <el-option label="科学计数" value="scientific" />
              <el-option label="日期" value="date" />
              <el-option label="时间" value="time" />
              <el-option label="文本" value="text" />
            </el-select>
            <div class="toolbar-btn-group">
              <el-button size="small" @click="emitAction('formatPercent')" title="百分比">%</el-button>
              <el-button size="small" @click="emitAction('formatCurrency')" title="货币">¥</el-button>
              <el-button size="small" @click="emitAction('decreaseDecimal')" title="减少小数位">.0</el-button>
              <el-button size="small" @click="emitAction('increaseDecimal')" title="增加小数位">.00</el-button>
            </div>
          </div>
        </div>

        <span class="toolbar-divider" />

        <!-- 单元格 -->
        <div class="toolbar-section">
          <div class="section-label">单元格</div>
          <div class="toolbar-group">
            <el-dropdown trigger="click" @command="onInsertAction">
              <el-button size="small"><el-icon><Plus /></el-icon>插入</el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="addRowAbove">上方插入行</el-dropdown-item>
                  <el-dropdown-item command="addRowBelow">下方插入行</el-dropdown-item>
                  <el-dropdown-item command="addColLeft">左侧插入列</el-dropdown-item>
                  <el-dropdown-item command="addColRight">右侧插入列</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
            <el-dropdown trigger="click" @command="onDeleteAction">
              <el-button size="small"><el-icon><Delete /></el-icon>删除</el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="deleteRow">删除行</el-dropdown-item>
                  <el-dropdown-item command="deleteCol">删除列</el-dropdown-item>
                  <el-dropdown-item command="clearCells">清空内容</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
            <el-button size="small" @click="emitAction('mergeCells')"><el-icon><Grid /></el-icon>合并</el-button>
            <el-dropdown trigger="click" @command="onBorderAction">
              <el-button size="small"><el-icon><EditPen /></el-icon>边框</el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="borderAll">所有边框</el-dropdown-item>
                  <el-dropdown-item command="borderNone">无边框</el-dropdown-item>
                  <el-dropdown-item command="borderOuter">外边框</el-dropdown-item>
                  <el-dropdown-item command="borderInner">内边框</el-dropdown-item>
                  <el-dropdown-item command="borderTop">上边框</el-dropdown-item>
                  <el-dropdown-item command="borderBottom">下边框</el-dropdown-item>
                  <el-dropdown-item command="borderLeft">左边框</el-dropdown-item>
                  <el-dropdown-item command="borderRight">右边框</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>

        <span class="toolbar-divider" />

        <!-- 行列操作 -->
        <div class="toolbar-section">
          <div class="section-label">行列</div>
          <div class="toolbar-group">
            <el-button size="small" @click="emitAction('hideRow')"><el-icon><Hide /></el-icon>隐藏行</el-button>
            <el-button size="small" @click="emitAction('hideCol')"><el-icon><Hide /></el-icon>隐藏列</el-button>
            <el-button size="small" @click="emitAction('freezeRow')"><el-icon><Lock /></el-icon>冻结</el-button>
          </div>
        </div>
      </template>

      <!-- 插入 菜单 -->
      <template v-else-if="activeMenu === 'insert'">
        <div class="toolbar-section">
          <div class="section-label">图表</div>
          <div class="toolbar-group">
            <el-dropdown trigger="click" @command="onInsertChart">
              <el-button size="small"><el-icon><TrendCharts /></el-icon>图表</el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="bar">柱状图</el-dropdown-item>
                  <el-dropdown-item command="line">折线图</el-dropdown-item>
                  <el-dropdown-item command="pie">饼图</el-dropdown-item>
                  <el-dropdown-item command="area">面积图</el-dropdown-item>
                  <el-dropdown-item command="scatter">散点图</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>

        <span class="toolbar-divider" />

        <div class="toolbar-section">
          <div class="section-label">图片</div>
          <div class="toolbar-group">
            <el-button size="small" @click="insertImage"><el-icon><Picture /></el-icon>图片</el-button>
          </div>
        </div>

        <span class="toolbar-divider" />

        <div class="toolbar-section">
          <div class="section-label">链接</div>
          <div class="toolbar-group">
            <el-button size="small" @click="insertLink"><el-icon><Link /></el-icon>链接</el-button>
          </div>
        </div>

        <span class="toolbar-divider" />

        <div class="toolbar-section">
          <div class="section-label">批注</div>
          <div class="toolbar-group">
            <el-button size="small" @click="addComment"><el-icon><ChatDotRound /></el-icon>批注</el-button>
          </div>
        </div>
      </template>

      <!-- 数据 菜单 -->
      <template v-else-if="activeMenu === 'data'">
        <div class="toolbar-section">
          <div class="section-label">排序</div>
          <div class="toolbar-group">
            <el-button size="small" @click="emitAction('sortAsc')"><el-icon><SortUp /></el-icon>升序</el-button>
            <el-button size="small" @click="emitAction('sortDesc')"><el-icon><SortDown /></el-icon>降序</el-button>
            <el-button size="small" @click="emitAction('customSort')">自定义排序</el-button>
          </div>
        </div>

        <span class="toolbar-divider" />

        <div class="toolbar-section">
          <div class="section-label">筛选</div>
          <div class="toolbar-group">
            <el-button size="small" @click="emitAction('filter')"><el-icon><Filter /></el-icon>筛选</el-button>
            <el-button size="small" @click="emitAction('clearFilter')">清除筛选</el-button>
          </div>
        </div>

        <span class="toolbar-divider" />

        <div class="toolbar-section">
          <div class="section-label">数据工具</div>
          <div class="toolbar-group">
            <el-button size="small" @click="emitAction('removeDuplicates')">删除重复项</el-button>
            <el-button size="small" @click="emitAction('textToColumns')">分列</el-button>
            <el-button size="small" @click="emitAction('dataValidation')">数据验证</el-button>
          </div>
        </div>

        <span class="toolbar-divider" />

        <div class="toolbar-section">
          <div class="section-label">分析</div>
          <div class="toolbar-group">
            <el-button size="small" @click="emitAction('pivotTable')">数据透视表</el-button>
          </div>
        </div>
      </template>

      <!-- 视图 菜单 -->
      <template v-else-if="activeMenu === 'view'">
        <div class="toolbar-section">
          <div class="section-label">显示</div>
          <div class="toolbar-group">
            <el-checkbox v-model="showGridlines" @change="onShowGridlinesChange">显示网格线</el-checkbox>
            <el-checkbox v-model="showHeadings" @change="onShowHeadingsChange">显示行列标题</el-checkbox>
          </div>
        </div>

        <span class="toolbar-divider" />

        <div class="toolbar-section">
          <div class="section-label">缩放</div>
          <div class="toolbar-group">
            <el-button size="small" @click="emitAction('zoomIn')"><el-icon><ZoomIn /></el-icon></el-button>
            <span class="zoom-level">{{ zoomLevel }}%</span>
            <el-button size="small" @click="emitAction('zoomOut')"><el-icon><ZoomOut /></el-icon></el-button>
            <el-button size="small" @click="emitAction('zoomReset')">100%</el-button>
          </div>
        </div>

        <span class="toolbar-divider" />

        <div class="toolbar-section">
          <div class="section-label">冻结</div>
          <div class="toolbar-group">
            <el-button size="small" @click="emitAction('freezeRow')">冻结首行</el-button>
            <el-button size="small" @click="emitAction('freezeCol')">冻结首列</el-button>
            <el-button size="small" @click="emitAction('unfreeze')">取消冻结</el-button>
          </div>
        </div>
      </template>

      <!-- 审阅 菜单 -->
      <template v-else-if="activeMenu === 'review'">
        <div class="toolbar-section">
          <div class="section-label">批注</div>
          <div class="toolbar-group">
            <el-button size="small" @click="addComment"><el-icon><ChatDotRound /></el-icon>新建批注</el-button>
            <el-button size="small" @click="emitAction('deleteComment')">删除批注</el-button>
            <el-button size="small" @click="emitAction('showComments')">显示所有批注</el-button>
          </div>
        </div>

        <span class="toolbar-divider" />

        <div class="toolbar-section">
          <div class="section-label">保护</div>
          <div class="toolbar-group">
            <el-button size="small" @click="protectSheet"><el-icon><Lock /></el-icon>保护工作表</el-button>
            <el-button size="small" @click="emitAction('unprotectSheet')">撤销保护</el-button>
          </div>
        </div>
      </template>
    </div>

    <!-- Formula Bar -->
    <div class="formula-bar">
      <div class="cell-reference">{{ cellReference || 'A1' }}</div>
      <div class="formula-divider"></div>
      <div class="formula-input">
        <span class="fx-label">fx</span>
        <input
          type="text"
          v-model="formulaValue"
          placeholder="输入值或公式"
          @change="onFormulaChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'

const props = defineProps<{
  cellReference?: string
  cellValue?: string
  title?: string
  saveStatus?: string
}>()

const emit = defineEmits(['action', 'update:title', 'saveTitle'])

const router = useRouter()

// Menu structure
const menus = [
  { key: 'home', label: '开始' },
  { key: 'insert', label: '插入' },
  { key: 'data', label: '数据' },
  { key: 'view', label: '视图' },
  { key: 'review', label: '审阅' },
]

const activeMenu = ref('home')

// Font settings
const fontFamily = ref('Microsoft YaHei')
const fontSize = ref(11)
const fontSizes = [8, 9, 10, 11, 12, 14, 16, 18, 20, 22, 24, 26, 28, 36, 48, 72]
const boldActive = ref(false)
const italicActive = ref(false)
const underlineActive = ref(false)
const strikeActive = ref(false)
const textColor = ref('#000000')
const bgColor = ref('#ffffff')

// Alignment
const align = ref('left')
const alignIcon = computed(() => {
  const icons: Record<string, string> = { left: 'AlignLeft', center: 'AlignCenter', right: 'AlignRight' }
  return icons[align.value] || 'AlignLeft'
})

// Number format
const numberFormat = ref('normal')

// View settings
const zoomLevel = ref(100)
const showGridlines = ref(true)
const showHeadings = ref(true)

// Formula bar
const formulaValue = ref(props.cellValue || '')

// Helper
function emitAction(action: string, params?: any) {
  emit('action', { action, params })
}

// Navigation
function goBack() {
  router.back()
}

// Clipboard
function onClipboardAction(cmd: string) {
  emitAction(cmd)
}

// Font
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
function toggleStrike() {
  strikeActive.value = !strikeActive.value
  emitAction('toggleStrike', strikeActive.value)
}
function onTextColorChange(val: string) { emitAction('setTextColor', val) }
function onBgColorChange(val: string) { emitAction('setBgColor', val) }

// Alignment
function onAlignChange(val: string) {
  align.value = val
  emitAction('setAlign', val)
}

// Number format
function onNumberFormatChange(val: string) { emitAction('setNumberFormat', val) }

// Insert
function onInsertAction(cmd: string) { emitAction(cmd) }
function onDeleteAction(cmd: string) { emitAction(cmd) }
function onBorderAction(cmd: string) { emitAction(cmd) }

// Insert specific
function insertImage() { emitAction('insertImage') }
function insertLink() { emitAction('insertLink') }
function onInsertChart(type: string) { emitAction('insertChart', type) }

// View
function onShowGridlinesChange(val: boolean) { emitAction('setShowGridlines', val) }
function onShowHeadingsChange(val: boolean) { emitAction('setShowHeadings', val) }

// Review
function addComment() { emitAction('addComment') }
function protectSheet() { emitAction('protectSheet') }

// Export & Print
function printSheet() { emitAction('printSheet') }
function exportSheet() { emitAction('exportSheet') }
function shareSheet() { emitAction('shareSheet') }

// Formula
function onFormulaChange() {
  emitAction('setFormula', formulaValue.value)
}
</script>

<style scoped>
.sheet-toolbar-wrapper {
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
  padding: 5px 8px;
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

.zoom-level {
  font-size: 12px;
  color: #646a73;
  min-width: 40px;
  text-align: center;
}

/* Formula Bar */
.formula-bar {
  display: flex;
  align-items: center;
  height: 28px;
  background: #fff;
  border-bottom: 1px solid var(--kx-border, #e5e6eb);
  padding: 0 8px;
}

.cell-reference {
  width: 60px;
  padding: 0 8px;
  font-size: 12px;
  color: #1f2329;
  font-weight: 500;
}

.formula-divider {
  width: 1px;
  height: 16px;
  background: #e5e6eb;
}

.formula-input {
  flex: 1;
  display: flex;
  align-items: center;
  padding: 0 8px;
}

.fx-label {
  font-size: 12px;
  color: #8f959e;
  font-style: italic;
  margin-right: 8px;
}

.formula-input input {
  flex: 1;
  border: none;
  outline: none;
  font-size: 13px;
  background: transparent;
}

.formula-input input::placeholder {
  color: #bbbfc4;
}
</style>
