<template>
  <div class="sheet-editor-wrapper">
    <div ref="univerContainer" class="univer-container" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import type { FUniver, Univer, IWorkbookData } from '@univerjs/presets'
import { UniverSheetsCorePreset } from '@univerjs/preset-sheets-core'
import UniverPresetSheetsCoreZhCN from '@univerjs/preset-sheets-core/locales/zh-CN'
import { createUniver, LocaleType, mergeLocales } from '@univerjs/presets'
import { useI18n } from 'vue-i18n'

import '@univerjs/preset-sheets-core/lib/index.css'

const { t } = useI18n()

const props = defineProps<{
  documentId: number
  content: string
  docTitle: string
}>()

const emit = defineEmits<{
  (e: 'save', content: string): void
  (e: 'update:title', title: string): void
  (e: 'save-title'): void
}>()

const univerContainer = ref<HTMLElement | null>(null)

// IMPORTANT: 不要用 ref() 包装这些，避免 Vue 响应式代理导致问题
let univerInstance: Univer | null = null
let univerAPI: FUniver | null = null
let saveTimer: ReturnType<typeof setTimeout> | null = null

// 旧格式数据转换为 Univer 格式
function convertOldFormatToUniver(content: string): IWorkbookData | null {
  if (!content || content.trim() === '') return null

  try {
    const parsed = JSON.parse(content)

    // 如果已经是 Univer 格式（有 sheetOrder 或 id 字段），直接返回
    if (parsed.sheetOrder || parsed.id) {
      return parsed as IWorkbookData
    }

    // 旧格式转换：{ sheets: [{ name, cells: string[][], styles }] }
    if (parsed.sheets && Array.isArray(parsed.sheets)) {
      const sheets: Record<string, any> = {}
      const sheetOrder: string[] = []

      parsed.sheets.forEach((sheet: any, index: number) => {
        const sheetId = `sheet-${index}`
        sheetOrder.push(sheetId)

        // 转换 cells 二维数组为 Univer cellData 格式
        const cellData: Record<number, Record<number, any>> = {}
        if (sheet.cells && Array.isArray(sheet.cells)) {
          sheet.cells.forEach((row: string[], rowIdx: number) => {
            if (!row) return
            cellData[rowIdx] = {}
            row.forEach((cellValue: string, colIdx: number) => {
              if (cellValue !== undefined && cellValue !== null && cellValue !== '') {
                cellData[rowIdx][colIdx] = { v: cellValue }

                // 应用样式
                const styleKey = `${rowIdx}-${colIdx}`
                if (sheet.styles && sheet.styles[styleKey]) {
                  const style = sheet.styles[styleKey]
                  const s: any = {}
                  if (style.fontWeight === 'bold') s.bl = 1
                  if (style.fontStyle === 'italic') s.it = 1
                  if (style.textColor && style.textColor !== '#000000') s.cl = { rgb: style.textColor }
                  if (style.bgColor && style.bgColor !== '#ffffff') s.bg = { rgb: style.bgColor }
                  if (style.fontSize) s.fs = style.fontSize
                  if (Object.keys(s).length > 0) {
                    cellData[rowIdx][colIdx].s = s
                  }
                }
              }
            })
          })
        }

        sheets[sheetId] = {
          id: sheetId,
          name: sheet.name || `${t('editor.sheet.worksheetPrefix')}${index + 1}`,
          cellData,
          defaultColumnWidth: 100,
          defaultRowHeight: 24,
          rowCount: Math.max(50, (sheet.cells?.length || 0) + 20),
          columnCount: 26,
        }
      })

      return {
        id: `workbook-${props.documentId}`,
        name: props.docTitle || t('editor.sheet.untitledSheet'),
        sheetOrder,
        sheets,
      } as unknown as IWorkbookData
    }

    return null
  } catch (e) {
    console.error('[SheetEditor] 解析内容失败:', e)
    return null
  }
}

// 获取默认空工作簿数据
function getDefaultWorkbook(): IWorkbookData {
  return {
    id: `workbook-${props.documentId}`,
    name: props.docTitle || t('editor.sheet.untitledSheet'),
    sheetOrder: ['sheet-0'],
    sheets: {
      'sheet-0': {
        id: 'sheet-0',
        name: `${t('editor.sheet.worksheetPrefix')}1`,
        cellData: {},
        defaultColumnWidth: 100,
        defaultRowHeight: 24,
        rowCount: 100,
        columnCount: 26,
      }
    },
  } as unknown as IWorkbookData
}

// 自动保存（防抖）
function scheduleSave() {
  if (saveTimer) clearTimeout(saveTimer)
  saveTimer = setTimeout(() => {
    if (!univerAPI) return
    const workbook = univerAPI.getActiveWorkbook()
    if (!workbook) return
    const snapshot = workbook.save()
    emit('save', JSON.stringify(snapshot))
  }, 2000)
}

onMounted(() => {
  if (!univerContainer.value) return

  // 解析内容
  const workbookData = convertOldFormatToUniver(props.content) || getDefaultWorkbook()

  // 创建 Univer 实例
  const { univer, univerAPI: api } = createUniver({
    locale: LocaleType.ZH_CN,
    locales: {
      [LocaleType.ZH_CN]: mergeLocales(UniverPresetSheetsCoreZhCN),
    },
    presets: [
      UniverSheetsCorePreset({
        container: univerContainer.value,
      }),
    ],
  })

  univerInstance = univer
  univerAPI = api

  // 加载工作簿
  univerAPI.createWorkbook(workbookData)

  // 监听变更，触发自动保存
  univerAPI.addEvent(univerAPI.Event.SheetValueChanged, () => {
    scheduleSave()
  })
})

onBeforeUnmount(() => {
  // 最后保存一次
  if (univerAPI) {
    const workbook = univerAPI.getActiveWorkbook()
    if (workbook) {
      const snapshot = workbook.save()
      emit('save', JSON.stringify(snapshot))
    }
  }

  if (saveTimer) {
    clearTimeout(saveTimer)
    saveTimer = null
  }

  // 销毁 Univer（注意顺序）
  univerInstance?.dispose()
  univerInstance = null
  univerAPI = null
})

// 当 content prop 变化时（例如版本回滚），重新加载
watch(() => props.content, (newContent) => {
  if (!univerAPI) return
  // 需要先销毁当前工作簿，再创建新的
  const currentWorkbook = univerAPI.getActiveWorkbook()
  if (currentWorkbook) {
    const unitId = currentWorkbook.getId()
    univerAPI.disposeUnit(unitId)
  }
  const workbookData = convertOldFormatToUniver(newContent) || getDefaultWorkbook()
  univerAPI.createWorkbook(workbookData)
})
</script>

<style scoped>
.sheet-editor-wrapper {
  width: 100%;
  height: 100%;
  position: relative;
}

.univer-container {
  width: 100%;
  height: 100%;
}
</style>
