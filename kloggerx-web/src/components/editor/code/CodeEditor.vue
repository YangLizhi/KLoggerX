<template>
  <div class="code-editor">
    <div class="code-toolbar">
      <el-select v-model="language" size="small" style="width: 120px" @change="onLanguageChange">
        <el-option-group label="常用">
          <el-option label="JavaScript" value="javascript" />
          <el-option label="TypeScript" value="typescript" />
          <el-option label="Python" value="python" />
          <el-option label="Java" value="java" />
          <el-option label="Go" value="go" />
          <el-option label="Rust" value="rust" />
        </el-option-group>
        <el-option-group label="前端">
          <el-option label="HTML" value="html" />
          <el-option label="CSS" value="css" />
          <el-option label="Vue" value="vue" />
          <el-option label="React JSX" value="jsx" />
        </el-option-group>
        <el-option-group label="其他">
          <el-option label="C/C++" value="cpp" />
          <el-option label="C#" value="csharp" />
          <el-option label="PHP" value="php" />
          <el-option label="Ruby" value="ruby" />
          <el-option label="Swift" value="swift" />
          <el-option label="Kotlin" value="kotlin" />
          <el-option label="SQL" value="sql" />
          <el-option label="Shell" value="bash" />
          <el-option label="JSON" value="json" />
          <el-option label="YAML" value="yaml" />
          <el-option label="XML" value="xml" />
          <el-option label="Markdown" value="markdown" />
        </el-option-group>
      </el-select>
      <div class="toolbar-actions">
        <el-button size="small" text @click="formatCode" title="格式化代码">
          <el-icon><MagicStick /></el-icon>格式化
        </el-button>
        <el-button size="small" text @click="copyCode" title="复制代码">
          <el-icon><DocumentCopy /></el-icon>复制
        </el-button>
        <el-button size="small" text @click="downloadCode" title="下载代码">
          <el-icon><Download /></el-icon>下载
        </el-button>
      </div>
      <div class="toolbar-info">
        <span class="line-info">行 {{ currentLine }}, 列 {{ currentColumn }}</span>
        <span class="lang-info">{{ language }}</span>
      </div>
    </div>
    <div class="code-container">
      <div class="line-numbers" ref="lineNumbersRef">
        <div v-for="n in lineCount" :key="n" class="line-number" :class="{ active: n === currentLine }">
          {{ n }}
        </div>
      </div>
      <textarea
        ref="textareaRef"
        class="code-textarea"
        :value="code"
        @input="onCodeInput"
        @scroll="syncScroll"
        @keydown="onKeyDown"
        @click="updateCursorPosition"
        @keyup="updateCursorPosition"
        :placeholder="`在此输入 ${language} 代码...`"
        spellcheck="false"
      ></textarea>
    </div>
    <div class="code-statusbar">
      <span>UTF-8</span>
      <span>LF</span>
      <span>{{ lineCount }} 行</span>
      <span>{{ code.length }} 字符</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { ElMessage } from 'element-plus'

const props = defineProps<{
  documentId: number
  content: string
}>()

const emit = defineEmits<{
  (e: 'save', content: string): void
}>()

const code = ref('')
const language = ref('javascript')
const currentLine = ref(1)
const currentColumn = ref(1)
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const lineNumbersRef = ref<HTMLDivElement | null>(null)
let saveTimer: ReturnType<typeof setTimeout> | null = null

const lineCount = computed(() => {
  if (!code.value) return 1
  return code.value.split('\n').length
})

function detectLanguage(content: string): string {
  // Try to detect language from content patterns
  if (content.includes('<?php')) return 'php'
  if (content.includes('<template>') && content.includes('<script')) return 'vue'
  if (content.includes('import React') || content.includes('jsx')) return 'jsx'
  if (content.includes('def ') || content.includes('import ')) return 'python'
  if (content.includes('func ') || content.includes('package ')) return 'go'
  if (content.includes('fn ') && content.includes('let ')) return 'rust'
  if (content.includes('public class') || content.includes('import java')) return 'java'
  if (content.includes('#include') || content.includes('int main')) return 'cpp'
  if (content.includes('SELECT') || content.includes('CREATE TABLE')) return 'sql'
  if (content.startsWith('{') || content.startsWith('[')) return 'json'
  if (content.includes('---') || content.includes('## ')) return 'yaml'
  if (content.includes('<!DOCTYPE') || content.includes('<html')) return 'html'
  return 'javascript'
}

function initData() {
  if (props.content) {
    code.value = props.content
    language.value = detectLanguage(props.content)
  }
}

function onCodeInput(e: Event) {
  code.value = (e.target as HTMLTextAreaElement).value
  scheduleSave()
}

function onLanguageChange() {
  scheduleSave()
}

function scheduleSave() {
  if (saveTimer) clearTimeout(saveTimer)
  saveTimer = setTimeout(() => {
    emit('save', JSON.stringify({ language: language.value, code: code.value }))
  }, 2000)
}

function updateCursorPosition() {
  if (!textareaRef.value) return
  const textarea = textareaRef.value
  const text = textarea.value.substring(0, textarea.selectionStart)
  const lines = text.split('\n')
  currentLine.value = lines.length
  currentColumn.value = lines[lines.length - 1].length + 1
}

function syncScroll() {
  if (!lineNumbersRef.value || !textareaRef.value) return
  lineNumbersRef.value.scrollTop = textareaRef.value.scrollTop
}

function onKeyDown(e: KeyboardEvent) {
  const textarea = textareaRef.value
  if (!textarea) return

  // Handle Tab key
  if (e.key === 'Tab') {
    e.preventDefault()
    const start = textarea.selectionStart
    const end = textarea.selectionEnd
    const spaces = '  ' // 2 spaces
    code.value = code.value.substring(0, start) + spaces + code.value.substring(end)
    nextTick(() => {
      textarea.selectionStart = textarea.selectionEnd = start + spaces.length
    })
    scheduleSave()
  }

  // Handle Ctrl+S
  if (e.key === 's' && (e.ctrlKey || e.metaKey)) {
    e.preventDefault()
    emit('save', JSON.stringify({ language: language.value, code: code.value }))
    ElMessage.success('已保存')
  }

  // Handle auto-close brackets
  const pairs: Record<string, string> = { '(': ')', '[': ']', '{': '}', '"': '"', "'": "'", '`': '`' }
  if (pairs[e.key]) {
    const start = textarea.selectionStart
    const end = textarea.selectionEnd
    if (start === end) {
      e.preventDefault()
      code.value = code.value.substring(0, start) + e.key + pairs[e.key] + code.value.substring(end)
      nextTick(() => {
        textarea.selectionStart = textarea.selectionEnd = start + 1
      })
      scheduleSave()
    }
  }

  // Handle Enter with auto-indent
  if (e.key === 'Enter') {
    const start = textarea.selectionStart
    const lineStart = code.value.lastIndexOf('\n', start - 1) + 1
    const currentLineText = code.value.substring(lineStart, start)
    const indent = currentLineText.match(/^\s*/)?.[0] || ''
    const openBrackets = (currentLineText.match(/[\{\[\(]/g) || []).length
    const closeBrackets = (currentLineText.match(/[\}\]\)]/g) || []).length
    const extraIndent = openBrackets > closeBrackets ? '  ' : ''
    e.preventDefault()
    code.value = code.value.substring(0, start) + '\n' + indent + extraIndent + code.value.substring(start)
    nextTick(() => {
      textarea.selectionStart = textarea.selectionEnd = start + 1 + indent.length + extraIndent.length
    })
    scheduleSave()
  }
}

function formatCode() {
  // Basic formatting - in production would use prettier or similar
  try {
    if (language.value === 'json') {
      code.value = JSON.stringify(JSON.parse(code.value), null, 2)
      ElMessage.success('格式化成功')
      scheduleSave()
    } else {
      ElMessage.info('当前语言暂不支持自动格式化')
    }
  } catch {
    ElMessage.error('格式化失败，请检查代码语法')
  }
}

function copyCode() {
  navigator.clipboard.writeText(code.value)
  ElMessage.success('代码已复制到剪贴板')
}

function downloadCode() {
  const extMap: Record<string, string> = {
    javascript: 'js', typescript: 'ts', python: 'py', java: 'java', go: 'go', rust: 'rs',
    html: 'html', css: 'css', vue: 'vue', jsx: 'jsx', cpp: 'cpp', csharp: 'cs',
    php: 'php', ruby: 'rb', swift: 'swift', kotlin: 'kt', sql: 'sql', bash: 'sh',
    json: 'json', yaml: 'yaml', xml: 'xml', markdown: 'md'
  }
  const ext = extMap[language.value] || 'txt'
  const blob = new Blob([code.value], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `code.${ext}`
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success('下载成功')
}

onMounted(() => initData())
watch(() => props.content, () => initData(), { once: true })
</script>

<style scoped>
.code-editor {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #1e1e1e;
}
.code-toolbar {
  display: flex;
  align-items: center;
  padding: 6px 12px;
  background: #252526;
  border-bottom: 1px solid #3c3c3c;
  gap: 12px;
  flex-shrink: 0;
}
.toolbar-actions {
  display: flex;
  gap: 4px;
}
.toolbar-actions .el-button {
  color: #cccccc;
}
.toolbar-actions .el-button:hover {
  color: #ffffff;
}
.toolbar-info {
  margin-left: auto;
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: #858585;
}
.code-container {
  flex: 1;
  display: flex;
  overflow: hidden;
  position: relative;
}
.line-numbers {
  width: 50px;
  background: #1e1e1e;
  color: #858585;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  line-height: 1.5;
  padding: 8px 0;
  text-align: right;
  user-select: none;
  overflow-y: hidden;
  flex-shrink: 0;
}
.line-number {
  padding-right: 12px;
  height: 19.5px;
}
.line-number.active {
  color: #c6c6c6;
  background: #2a2a2a;
}
.code-textarea {
  flex: 1;
  background: #1e1e1e;
  color: #d4d4d4;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  line-height: 1.5;
  padding: 8px 12px;
  border: none;
  outline: none;
  resize: none;
  white-space: pre;
  overflow-wrap: normal;
  overflow-x: auto;
}
.code-textarea::placeholder {
  color: #6a6a6a;
}
.code-statusbar {
  display: flex;
  gap: 16px;
  padding: 4px 12px;
  background: #007acc;
  color: #fff;
  font-size: 12px;
  flex-shrink: 0;
}
</style>
