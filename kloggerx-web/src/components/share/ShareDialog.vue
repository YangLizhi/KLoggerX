<template>
  <el-dialog :model-value="true" :title="$t('share.settings')" width="480px" @close="$emit('close')">
    <el-form v-if="setting" label-position="top">
      <el-form-item :label="$t('share.scope')">
        <el-radio-group v-model="setting.scope">
          <el-radio value="collaborator">{{ $t('share.collaboratorOnly') }}</el-radio>
          <el-radio value="organization">{{ $t('share.orgVisible') }}</el-radio>
          <el-radio value="public">{{ $t('share.publicVisible') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item :label="$t('share.defaultPerm')">
        <el-select v-model="setting.defaultPermission">
          <el-option :label="$t('share.canEdit')" value="edit" />
          <el-option :label="$t('share.canView')" value="view" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('share.shareLink')">
        <el-switch v-model="setting.linkEnabled" />
        <div v-if="setting.linkEnabled && setting.shareLink" class="share-link-row">
          <el-input :model-value="setting.shareLink" readonly size="small" style="flex:1" />
          <el-button size="small" @click="copyLink">{{ $t('share.copyLink') }}</el-button>
        </div>
      </el-form-item>
      <el-form-item>
        <el-checkbox v-model="setting.includeChildren">{{ $t('share.includeChildren') }}</el-checkbox>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="$emit('close')">{{ $t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="saving" @click="handleSave">{{ $t('common.save') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { getShareSetting, updateShareSetting } from '@/api/modules/auth'
import type { ShareSetting } from '@/types'
import { ElMessage } from 'element-plus'

const props = defineProps<{ documentId: number }>()
defineEmits(['close'])
const { t } = useI18n()
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
    ElMessage.success(t('share.saved'))
  } finally { saving.value = false }
}

function copyLink() {
  if (setting.value?.shareLink) {
    navigator.clipboard.writeText(setting.value.shareLink)
    ElMessage.success(t('common.copiedToClipboard'))
  }
}

onMounted(fetchSetting)
</script>

<style scoped>
.share-link-row { display: flex; gap: 8px; margin-top: 8px; width: 100%; }
</style>
