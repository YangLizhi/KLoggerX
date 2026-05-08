<template>
  <el-dialog
    :model-value="visible"
    @update:model-value="$emit('update:visible', $event)"
    :title="$t('knowledge.shareDialog.title')"
    width="480px"
    destroy-on-close
  >
    <div class="share-dialog-body">
      <!-- Expiry selection -->
      <div class="share-option" v-if="!shareToken">
        <label>{{ $t('knowledge.shareDialog.expiryLabel') }}</label>
        <el-select v-model="expiryOption" style="width: 200px">
          <el-option :label="$t('knowledge.shareDialog.neverExpire')" :value="0" />
          <el-option :label="$t('knowledge.shareDialog.hours24')" :value="24" />
          <el-option :label="$t('knowledge.shareDialog.days7')" :value="168" />
          <el-option :label="$t('knowledge.shareDialog.days30')" :value="720" />
        </el-select>
      </div>

      <!-- Share link display -->
      <div v-if="shareToken" class="share-link-area">
        <div class="share-link-label">{{ $t('knowledge.shareDialog.shareLink') }}</div>
        <div class="share-link-input">
          <el-input :model-value="shareUrl" readonly>
            <template #append>
              <el-button @click="copyLink">{{ $t('common.copy') }}</el-button>
            </template>
          </el-input>
        </div>
        <div class="share-info" v-if="expiresAt">
          <el-icon><Clock /></el-icon>
          <span>{{ $t('knowledge.shareDialog.expireTime') }}: {{ expiresAt }}</span>
        </div>
        <div class="share-info" v-else>
          <el-icon><Clock /></el-icon>
          <span>{{ $t('knowledge.shareDialog.neverExpire') }}</span>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="share-dialog-footer">
        <el-button v-if="shareToken" type="danger" plain @click="handleRevoke" :loading="revoking">
          {{ $t('knowledge.shareDialog.revokeShare') }}
        </el-button>
        <el-button v-if="!shareToken" type="primary" @click="handleCreate" :loading="creating">
          {{ $t('knowledge.shareDialog.createShareLink') }}
        </el-button>
        <el-button @click="$emit('update:visible', false)">{{ $t('common.close') }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Clock } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createShareLink, revokeShareLink } from '@/api/modules/knowledge'

const props = defineProps<{
  visible: boolean
  conversationId: number | null
  conversationTitle?: string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
}>()

const { t } = useI18n()
const expiryOption = ref(0)
const shareToken = ref('')
const expiresAt = ref('')
const creating = ref(false)
const revoking = ref(false)

const shareUrl = computed(() => {
  return `${window.location.origin}/share/${shareToken.value}`
})

watch(() => props.visible, (val) => {
  if (!val) {
    shareToken.value = ''
    expiresAt.value = ''
  }
})

async function handleCreate() {
  if (!props.conversationId) return
  creating.value = true
  try {
    const res: any = await createShareLink(props.conversationId, expiryOption.value || undefined)
    if (res.data) {
      shareToken.value = res.data.shareToken
      expiresAt.value = res.data.expiresAt || ''
      ElMessage.success(t('knowledge.shareDialog.linkCreated'))
    }
  } catch {
    ElMessage.error(t('knowledge.shareDialog.createFailed'))
  } finally {
    creating.value = false
  }
}

async function handleRevoke() {
  if (!props.conversationId) return
  try {
    await ElMessageBox.confirm(t('knowledge.shareDialog.revokeConfirm'), t('knowledge.shareDialog.revokeTitle'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
  } catch {
    return
  }
  revoking.value = true
  try {
    await revokeShareLink(props.conversationId)
    shareToken.value = ''
    expiresAt.value = ''
    ElMessage.success(t('knowledge.shareDialog.revoked'))
  } catch {
    ElMessage.error(t('knowledge.shareDialog.revokeFailed'))
  } finally {
    revoking.value = false
  }
}

function copyLink() {
  navigator.clipboard.writeText(shareUrl.value)
  ElMessage.success(t('common.copiedToClipboard'))
}
</script>

<style scoped>
.share-dialog-body {
  padding: 8px 0;
}

.share-option {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.share-option label {
  font-size: 14px;
  color: var(--kx-text-primary);
  white-space: nowrap;
}

.share-link-area {
  margin-top: 8px;
}

.share-link-label {
  font-size: 13px;
  color: var(--kx-text-secondary);
  margin-bottom: 8px;
}

.share-link-input {
  margin-bottom: 12px;
}

.share-info {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--kx-text-secondary);
}

.share-dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
