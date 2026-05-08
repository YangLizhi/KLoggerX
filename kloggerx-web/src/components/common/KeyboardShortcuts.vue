<template>
  <el-dialog v-model="visible" :title="$t('shortcuts.title')" width="500px" :append-to-body="true">
    <div class="shortcuts-container">
      <div v-for="group in shortcutGroups" :key="group.title" class="shortcut-group">
        <h4>{{ group.title }}</h4>
        <div v-for="item in group.items" :key="item.keys" class="shortcut-item">
          <span class="shortcut-desc">{{ item.description }}</span>
          <span class="shortcut-keys">
            <kbd v-for="key in item.keys.split('+')" :key="key">{{ key.trim() }}</kbd>
          </span>
        </div>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const visible = defineModel<boolean>('visible', { default: false })

interface ShortcutItem {
  keys: string
  description: string
}

interface ShortcutGroup {
  title: string
  items: ShortcutItem[]
}

const shortcutGroups = computed<ShortcutGroup[]>(() => [
  {
    title: t('shortcuts.general'),
    items: [
      { keys: '?', description: t('shortcuts.showHelp') },
      { keys: 'Esc', description: t('shortcuts.closeDialog') },
      { keys: 'Ctrl + K', description: t('shortcuts.globalSearch') },
    ],
  },
  {
    title: t('shortcuts.navigation'),
    items: [
      { keys: 'Ctrl + 1', description: t('shortcuts.goHome') },
      { keys: 'Ctrl + 2', description: t('shortcuts.goKnowledge') },
      { keys: 'Ctrl + N', description: t('shortcuts.newDoc') },
    ],
  },
  {
    title: t('shortcuts.editor'),
    items: [
      { keys: 'Ctrl + S', description: t('shortcuts.save') },
      { keys: 'Ctrl + B', description: t('shortcuts.bold') },
      { keys: 'Ctrl + I', description: t('shortcuts.italic') },
      { keys: 'Ctrl + U', description: t('shortcuts.underline') },
      { keys: 'Ctrl + Z', description: t('shortcuts.undo') },
      { keys: 'Ctrl + Shift + Z', description: t('shortcuts.redo') },
    ],
  },
])
</script>

<style scoped>
.shortcuts-container {
  max-height: 60vh;
  overflow-y: auto;
}

.shortcut-group {
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--kx-border, #e5e6eb);
}

.shortcut-group:last-child {
  border-bottom: none;
  margin-bottom: 0;
  padding-bottom: 0;
}

.shortcut-group h4 {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 600;
  color: var(--kx-text-primary, #1f2329);
}

.shortcut-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 0;
}

.shortcut-desc {
  font-size: 13px;
  color: var(--kx-text-secondary, #646a73);
}

.shortcut-keys {
  display: flex;
  align-items: center;
  gap: 4px;
}

kbd {
  display: inline-block;
  padding: 2px 6px;
  font-size: 12px;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  color: var(--kx-text-primary, #1f2329);
  background: #f2f3f5;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  min-width: 20px;
  text-align: center;
  line-height: 1.5;
}
</style>
