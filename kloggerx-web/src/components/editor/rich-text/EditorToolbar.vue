<template>
  <div class="editor-toolbar" v-if="editor">
    <div class="toolbar-group">
      <button :class="{ active: editor.isActive('bold') }" @click="editor.chain().focus().toggleBold().run()" :title="$t('editor.toolbar.bold')">
        <strong>B</strong>
      </button>
      <button :class="{ active: editor.isActive('italic') }" @click="editor.chain().focus().toggleItalic().run()" :title="$t('editor.toolbar.italic')">
        <em>I</em>
      </button>
      <button :class="{ active: editor.isActive('underline') }" @click="editor.chain().focus().toggleUnderline().run()" :title="$t('editor.toolbar.underline')">
        <span style="text-decoration:underline">U</span>
      </button>
      <button :class="{ active: editor.isActive('strike') }" @click="editor.chain().focus().toggleStrike().run()" :title="$t('editor.toolbar.strikethrough')">
        <span style="text-decoration:line-through">S</span>
      </button>
    </div>

    <span class="toolbar-divider" />

    <div class="toolbar-group">
      <button :class="{ active: editor.isActive('heading', { level: 1 }) }" @click="editor.chain().focus().toggleHeading({ level: 1 }).run()" :title="$t('editor.toolbar.heading1')">
        H1
      </button>
      <button :class="{ active: editor.isActive('heading', { level: 2 }) }" @click="editor.chain().focus().toggleHeading({ level: 2 }).run()" :title="$t('editor.toolbar.heading2')">
        H2
      </button>
      <button :class="{ active: editor.isActive('heading', { level: 3 }) }" @click="editor.chain().focus().toggleHeading({ level: 3 }).run()" :title="$t('editor.toolbar.heading3')">
        H3
      </button>
    </div>

    <span class="toolbar-divider" />

    <div class="toolbar-group">
      <button :class="{ active: editor.isActive('bulletList') }" @click="editor.chain().focus().toggleBulletList().run()" :title="$t('editor.toolbar.bulletList')">
        <el-icon><List /></el-icon>
      </button>
      <button :class="{ active: editor.isActive('orderedList') }" @click="editor.chain().focus().toggleOrderedList().run()" :title="$t('editor.toolbar.orderedList')">
        <span style="font-size:12px">1.</span>
      </button>
      <button :class="{ active: editor.isActive('taskList') }" @click="editor.chain().focus().toggleTaskList().run()" :title="$t('editor.toolbar.taskList')">
        <el-icon><Finished /></el-icon>
      </button>
    </div>

    <span class="toolbar-divider" />

    <div class="toolbar-group">
      <button @click="editor.chain().focus().setHorizontalRule().run()" :title="$t('editor.toolbar.horizontalRule')">
        &mdash;
      </button>
      <button :class="{ active: editor.isActive('blockquote') }" @click="editor.chain().focus().toggleBlockquote().run()" :title="$t('editor.toolbar.blockquote')">
        <el-icon><ChatDotRound /></el-icon>
      </button>
      <button :class="{ active: editor.isActive('codeBlock') }" @click="editor.chain().focus().toggleCodeBlock().run()" :title="$t('editor.toolbar.codeBlock')">
        &lt;/&gt;
      </button>
    </div>

    <span class="toolbar-divider" />

    <div class="toolbar-group">
      <button @click="insertImage" :title="$t('editor.toolbar.image')">
        <el-icon><Picture /></el-icon>
      </button>
      <button @click="insertLink" :title="$t('editor.toolbar.link')">
        <el-icon><Link /></el-icon>
      </button>
      <button @click="insertTable" :title="$t('editor.toolbar.table')">
        <el-icon><Grid /></el-icon>
      </button>
    </div>

    <span class="toolbar-divider" />

    <div class="toolbar-group">
      <button :class="{ active: editor.isActive({ textAlign: 'left' }) }" @click="editor.chain().focus().setTextAlign('left').run()" :title="$t('editor.toolbar.alignLeft')">
        <span style="font-size:11px">&#9776;</span>
      </button>
      <button :class="{ active: editor.isActive({ textAlign: 'center' }) }" @click="editor.chain().focus().setTextAlign('center').run()" :title="$t('editor.toolbar.alignCenter')">
        <span style="font-size:11px">&#9776;</span>
      </button>
      <button :class="{ active: editor.isActive({ textAlign: 'right' }) }" @click="editor.chain().focus().setTextAlign('right').run()" :title="$t('editor.toolbar.alignRight')">
        <span style="font-size:11px">&#9776;</span>
      </button>
    </div>

    <span class="toolbar-divider" />

    <div class="toolbar-group">
      <button :class="{ active: editor.isActive('highlight') }" @click="editor.chain().focus().toggleHighlight().run()" :title="$t('editor.toolbar.highlight')">
        <span style="background:#fef08a;padding:0 2px;border-radius:2px">A</span>
      </button>
      <button @click="editor.chain().focus().unsetAllMarks().clearNodes().run()" :title="$t('editor.toolbar.clearFormat')">
        <el-icon><Delete /></el-icon>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Editor } from '@tiptap/vue-3'
import { ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps<{
  editor: Editor | undefined
}>()

async function insertImage() {
  if (!props.editor) return
  try {
    const { value } = await ElMessageBox.prompt(t('editor.toolbar.insertImagePrompt'), t('editor.toolbar.insertImageTitle'), {
      inputPlaceholder: 'https://example.com/image.png',
      confirmButtonText: t('editor.toolbar.insert'),
      cancelButtonText: t('common.cancel'),
    })
    if (value) {
      props.editor.chain().focus().setImage({ src: value }).run()
    }
  } catch {
    // cancelled
  }
}

async function insertLink() {
  if (!props.editor) return
  try {
    const { value } = await ElMessageBox.prompt(t('editor.toolbar.insertLinkPrompt'), t('editor.toolbar.insertLinkTitle'), {
      inputPlaceholder: 'https://example.com',
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
    })
    if (value) {
      props.editor.chain().focus().setLink({ href: value }).run()
    }
  } catch {
    // cancelled
  }
}

function insertTable() {
  if (!props.editor) return
  props.editor.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run()
}
</script>

<style scoped>
.editor-toolbar {
  display: flex;
  align-items: center;
  padding: 4px 16px;
  border-bottom: 1px solid var(--kx-border, #e5e6eb);
  background: #fafafa;
  flex-wrap: wrap;
  gap: 2px;
  flex-shrink: 0;
}
.toolbar-group {
  display: flex;
  align-items: center;
  gap: 2px;
}
.toolbar-divider {
  width: 1px;
  height: 20px;
  background: var(--kx-border, #e5e6eb);
  margin: 0 6px;
}
.editor-toolbar button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 28px;
  border: none;
  background: transparent;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  color: #1f2329;
  transition: all 0.15s;
}
.editor-toolbar button:hover {
  background: #e8e9eb;
}
.editor-toolbar button.active {
  background: #dee0e3;
  color: #3370ff;
}
</style>
