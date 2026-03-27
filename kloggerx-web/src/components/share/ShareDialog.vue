<template>
  <el-dialog :model-value="true" title="分享设置" width="480px" @close="$emit('close')">
    <el-form v-if="setting" label-position="top">
      <el-form-item label="分享范围">
        <el-radio-group v-model="setting.scope">
          <el-radio value="collaborator">仅协作者</el-radio>
          <el-radio value="organization">组织内</el-radio>
          <el-radio value="public">公开</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="默认权限">
        <el-select v-model="setting.defaultPermission">
          <el-option label="可编辑" value="edit" />
          <el-option label="可查看" value="view" />
        </el-select>
      </el-form-item>
      <el-form-item label="分享链接">
        <el-switch v-model="setting.linkEnabled" />
        <div v-if="setting.linkEnabled && setting.shareLink" class="share-link-row">
          <el-input :model-value="setting.shareLink" readonly size="small" style="flex:1" />
          <el-button size="small" @click="copyLink">复制链接</el-button>
        </div>
      </el-form-item>
      <el-form-item>
        <el-checkbox v-model="setting.includeChildren">包含子文档</el-checkbox>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="$emit('close')">取消</el-button>
      <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getShareSetting, updateShareSetting } from '@/api/modules/auth'
import type { ShareSetting } from '@/types'
import { ElMessage } from 'element-plus'

const props = defineProps<{ documentId: number }>()
defineEmits(['close'])
const setting = ref<ShareSetting | null>(null)
const saving = ref(false)

async function fetchSetting() {
  const res: any = await getShareSetting(props.documentId)
  setting.value = res.data || {
    documentId: props.documentId,
    scope: 'collaborator',
    defaultPermission: 'view',
    linkEnabled: false,
    shareLink: '',
    includeChildren: false,
  }
}

async function handleSave() {
  if (!setting.value) return
  saving.value = true
  try {
    await updateShareSetting(setting.value)
    ElMessage.success('分享设置已保存')
  } finally { saving.value = false }
}

function copyLink() {
  if (setting.value?.shareLink) {
    navigator.clipboard.writeText(setting.value.shareLink)
    ElMessage.success('链接已复制')
  }
}

onMounted(fetchSetting)
</script>

<style scoped>
.share-link-row { display: flex; gap: 8px; margin-top: 8px; width: 100%; }
</style>
