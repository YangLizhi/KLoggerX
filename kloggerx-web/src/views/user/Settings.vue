<template>
  <div class="settings-page">
    <h2 class="settings-title">{{ $t('settings.title') }}</h2>
    <el-tabs v-model="activeTab">
      <!-- Profile Tab -->
      <el-tab-pane :label="$t('settings.profile')" name="profile">
        <el-form ref="profileFormRef" :model="profileForm" label-width="80px" style="max-width: 480px">
          <el-form-item :label="$t('settings.avatar')">
            <div class="avatar-section">
              <el-avatar :size="64" :src="profileForm.avatar">{{ profileForm.nickname?.[0] || 'U' }}</el-avatar>
              <el-upload
                :show-file-list="false"
                :before-upload="handleAvatarUpload"
                accept="image/*"
              >
                <el-button size="small">{{ $t('settings.changeAvatar') }}</el-button>
              </el-upload>
            </div>
          </el-form-item>
          <el-form-item :label="$t('settings.nickname')" prop="nickname">
            <el-input v-model="profileForm.nickname" maxlength="20" show-word-limit />
          </el-form-item>
          <el-form-item :label="$t('auth.email')">
            <el-input :model-value="profileForm.email" disabled />
          </el-form-item>
          <el-form-item :label="$t('settings.department')">
            <el-input :model-value="profileForm.department || $t('settings.notSet')" disabled />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="saving" @click="handleSaveProfile">{{ $t('common.save') }}</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- Password Tab -->
      <el-tab-pane :label="$t('settings.changePassword')" name="password">
        <el-form ref="pwdFormRef" :model="pwdForm" :rules="pwdRules" label-width="100px" style="max-width: 480px">
          <el-form-item :label="$t('settings.currentPassword')" prop="oldPassword">
            <el-input v-model="pwdForm.oldPassword" type="password" show-password />
          </el-form-item>
          <el-form-item :label="$t('settings.newPassword')" prop="newPassword">
            <el-input v-model="pwdForm.newPassword" type="password" show-password />
          </el-form-item>
          <el-form-item :label="$t('settings.confirmNewPassword')" prop="confirmPassword">
            <el-input v-model="pwdForm.confirmPassword" type="password" show-password />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="changingPwd" @click="handleChangePassword">{{ $t('settings.changePassword') }}</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- Notification Tab -->
      <el-tab-pane :label="$t('settings.notifications')" name="notifications">
        <div class="settings-section">
          <h3 class="section-title">{{ $t('settings.messageNotification') }}</h3>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">{{ $t('settings.commentNotify') }}</div>
              <div class="setting-desc">{{ $t('settings.commentNotifyDesc') }}</div>
            </div>
            <el-switch v-model="notifSettings.commentNotify" />
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">{{ $t('settings.mentionNotify') }}</div>
              <div class="setting-desc">{{ $t('settings.mentionNotifyDesc') }}</div>
            </div>
            <el-switch v-model="notifSettings.mentionNotify" />
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">{{ $t('settings.permissionNotify') }}</div>
              <div class="setting-desc">{{ $t('settings.permissionNotifyDesc') }}</div>
            </div>
            <el-switch v-model="notifSettings.permissionNotify" />
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">{{ $t('settings.shareNotify') }}</div>
              <div class="setting-desc">{{ $t('settings.shareNotifyDesc') }}</div>
            </div>
            <el-switch v-model="notifSettings.shareNotify" />
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">{{ $t('settings.approvalNotify') }}</div>
              <div class="setting-desc">{{ $t('settings.approvalNotifyDesc') }}</div>
            </div>
            <el-switch v-model="notifSettings.approvalNotify" />
          </div>
        </div>
        <div class="settings-section">
          <h3 class="section-title">{{ $t('settings.notifyMethod') }}</h3>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">{{ $t('settings.inAppNotify') }}</div>
              <div class="setting-desc">{{ $t('settings.inAppNotifyDesc') }}</div>
            </div>
            <el-switch v-model="notifSettings.inAppNotify" />
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">{{ $t('settings.emailNotify') }}</div>
              <div class="setting-desc">{{ $t('settings.emailNotifyDesc') }}</div>
            </div>
            <el-switch v-model="notifSettings.emailNotify" />
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">{{ $t('settings.dndMode') }}</div>
              <div class="setting-desc">{{ $t('settings.dndModeDesc') }}</div>
            </div>
            <el-switch v-model="notifSettings.dndEnabled" />
          </div>
          <div v-if="notifSettings.dndEnabled" class="setting-sub-row">
            <el-time-select v-model="notifSettings.dndStart" :max-time="notifSettings.dndEnd" :placeholder="$t('settings.startTime')" start="00:00" step="00:30" end="23:30" />
            <span class="dnd-separator">{{ $t('common.to') }}</span>
            <el-time-select v-model="notifSettings.dndEnd" :min-time="notifSettings.dndStart" :placeholder="$t('settings.endTime')" start="00:00" step="00:30" end="23:30" />
          </div>
        </div>
        <el-button type="primary" style="margin-top:16px" @click="handleSaveNotifSettings">{{ $t('settings.saveNotifSettings') }}</el-button>
      </el-tab-pane>

      <!-- Storage Tab -->
      <el-tab-pane :label="$t('settings.storage')" name="storage">
        <div class="settings-section">
          <h3 class="section-title">{{ $t('settings.storageOverview') }}</h3>
          <div class="storage-overview">
            <div class="storage-chart">
              <el-progress type="dashboard" :percentage="storagePercent" :width="140" :color="storagePercent > 90 ? '#f54a45' : '#3370ff'">
                <template #default>
                  <div class="storage-chart-text">
                    <div class="storage-used-text">{{ storageUsedText }}</div>
                    <div class="storage-total-text">/ {{ storageTotalText }}</div>
                  </div>
                </template>
              </el-progress>
            </div>
            <div class="storage-breakdown">
              <div class="breakdown-item">
                <div class="breakdown-color" style="background:#3370ff" />
                <span class="breakdown-label">{{ $t('settings.docStorage') }}</span>
                <span class="breakdown-size">{{ formatSize(storageByType.doc) }}</span>
              </div>
              <div class="breakdown-item">
                <div class="breakdown-color" style="background:#36b37e" />
                <span class="breakdown-label">{{ $t('settings.sheetStorage') }}</span>
                <span class="breakdown-size">{{ formatSize(storageByType.sheet) }}</span>
              </div>
              <div class="breakdown-item">
                <div class="breakdown-color" style="background:#ff7d00" />
                <span class="breakdown-label">{{ $t('settings.slideStorage') }}</span>
                <span class="breakdown-size">{{ formatSize(storageByType.slide) }}</span>
              </div>
              <div class="breakdown-item">
                <div class="breakdown-color" style="background:#9254de" />
                <span class="breakdown-label">{{ $t('settings.otherStorage') }}</span>
                <span class="breakdown-size">{{ formatSize(storageByType.other) }}</span>
              </div>
            </div>
          </div>
        </div>
        <div class="settings-section">
          <h3 class="section-title">{{ $t('settings.storagePathSettings') }}</h3>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">{{ $t('settings.syncDir') }}</div>
              <div class="setting-desc">{{ $t('settings.syncDirDesc') }}</div>
            </div>
            <el-input v-model="userStoragePaths.syncDir" :placeholder="$t('settings.syncDirPlaceholder')" style="width:300px" />
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">{{ $t('settings.downloadDir') }}</div>
              <div class="setting-desc">{{ $t('settings.downloadDirDesc') }}</div>
            </div>
            <el-input v-model="userStoragePaths.downloadDir" :placeholder="$t('settings.downloadDirPlaceholder')" style="width:300px" />
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">{{ $t('settings.autoSync') }}</div>
              <div class="setting-desc">{{ $t('settings.autoSyncDesc') }}</div>
            </div>
            <el-switch v-model="userStoragePaths.autoSync" />
          </div>
          <el-button type="primary" style="margin-top:16px" @click="handleSaveStoragePaths" :loading="savingStoragePaths">{{ $t('settings.savePathSettings') }}</el-button>
        </div>
        <div class="settings-section">
          <h3 class="section-title">{{ $t('settings.storageManagement') }}</h3>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">{{ $t('settings.autoCleanRecycle') }}</div>
              <div class="setting-desc">{{ $t('settings.autoCleanRecycleDesc') }}</div>
            </div>
            <el-switch v-model="storageSettings.autoCleanRecycle" />
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">{{ $t('settings.versionRetention') }}</div>
              <div class="setting-desc">{{ $t('settings.versionRetentionDesc') }}</div>
            </div>
            <el-select v-model="storageSettings.versionRetention" style="width:120px">
              <el-option :value="10" :label="$t('settings.last10')" />
              <el-option :value="20" :label="$t('settings.last20')" />
              <el-option :value="50" :label="$t('settings.last50')" />
              <el-option :value="100" :label="$t('settings.last100')" />
            </el-select>
          </div>
          <el-button type="primary" style="margin-top:16px" @click="showUpgradeDialog = true">{{ $t('settings.upgradeStorage') }}</el-button>
        </div>
      </el-tab-pane>

      <!-- Appearance Tab -->
      <el-tab-pane :label="$t('settings.appearance')" name="appearance">
        <div class="settings-section">
          <h3 class="section-title">{{ $t('settings.theme') }}</h3>
          <div class="theme-grid">
            <div class="theme-card" :class="{ active: appearance.theme === 'light' }" @click="appearance.theme = 'light'">
              <div class="theme-preview light-preview">
                <div class="tp-sidebar" /><div class="tp-content"><div class="tp-bar" /><div class="tp-line" /><div class="tp-line short" /></div>
              </div>
              <span>{{ $t('settings.lightMode') }}</span>
            </div>
            <div class="theme-card" :class="{ active: appearance.theme === 'dark' }" @click="appearance.theme = 'dark'">
              <div class="theme-preview dark-preview">
                <div class="tp-sidebar" /><div class="tp-content"><div class="tp-bar" /><div class="tp-line" /><div class="tp-line short" /></div>
              </div>
              <span>{{ $t('settings.darkMode') }}</span>
            </div>
            <div class="theme-card" :class="{ active: appearance.theme === 'auto' }" @click="appearance.theme = 'auto'">
              <div class="theme-preview auto-preview">
                <div class="tp-sidebar" /><div class="tp-content"><div class="tp-bar" /><div class="tp-line" /><div class="tp-line short" /></div>
              </div>
              <span>{{ $t('settings.followSystem') }}</span>
            </div>
          </div>
        </div>
        <div class="settings-section">
          <h3 class="section-title">{{ $t('settings.editor') }}</h3>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">{{ $t('settings.fontSize') }}</div>
              <div class="setting-desc">{{ $t('settings.fontSizeDesc') }}</div>
            </div>
            <el-select v-model="appearance.fontSize" style="width:100px">
              <el-option :value="13" :label="$t('settings.fontSmall')" />
              <el-option :value="15" :label="$t('settings.fontMedium')" />
              <el-option :value="17" :label="$t('settings.fontLarge')" />
              <el-option :value="19" :label="$t('settings.fontXLarge')" />
            </el-select>
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">{{ $t('settings.editorWidth') }}</div>
              <div class="setting-desc">{{ $t('settings.editorWidthDesc') }}</div>
            </div>
            <el-select v-model="appearance.editorWidth" style="width:120px">
              <el-option value="narrow" :label="$t('settings.widthNarrow')" />
              <el-option value="medium" :label="$t('settings.widthMedium')" />
              <el-option value="wide" :label="$t('settings.widthWide')" />
              <el-option value="full" :label="$t('settings.widthFull')" />
            </el-select>
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">{{ $t('settings.codeHighlight') }}</div>
              <div class="setting-desc">{{ $t('settings.codeHighlightDesc') }}</div>
            </div>
            <el-select v-model="appearance.codeTheme" style="width:120px">
              <el-option value="github" label="GitHub" />
              <el-option value="monokai" label="Monokai" />
              <el-option value="dracula" label="Dracula" />
              <el-option value="one-dark" label="One Dark" />
            </el-select>
          </div>
        </div>
        <div class="settings-section">
          <h3 class="section-title">{{ $t('settings.sidebar') }}</h3>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">{{ $t('settings.collapseSidebar') }}</div>
              <div class="setting-desc">{{ $t('settings.collapseSidebarDesc') }}</div>
            </div>
            <el-switch v-model="appearance.sidebarCollapsed" />
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">{{ $t('settings.showDocIcons') }}</div>
              <div class="setting-desc">{{ $t('settings.showDocIconsDesc') }}</div>
            </div>
            <el-switch v-model="appearance.showDocIcons" />
          </div>
        </div>
        <el-button type="primary" style="margin-top:16px" @click="handleSaveAppearance">{{ $t('settings.saveAppearance') }}</el-button>
      </el-tab-pane>

      <!-- About Tab -->
      <el-tab-pane :label="$t('settings.about')" name="about">
        <div class="settings-section">
          <h3 class="section-title">{{ $t('settings.aboutTitle') }}</h3>
          <div class="about-info">
            <div class="about-row"><span class="about-label">{{ $t('settings.version') }}</span><span>v1.0.0</span></div>
            <div class="about-row"><span class="about-label">{{ $t('settings.frontend') }}</span><span>Vue 3 + Element Plus</span></div>
            <div class="about-row"><span class="about-label">{{ $t('settings.editorEngine') }}</span><span>TipTap + ProseMirror</span></div>
            <div class="about-row"><span class="about-label">{{ $t('settings.collabEngine') }}</span><span>Yjs + WebSocket</span></div>
          </div>
        </div>
        <div class="settings-section">
          <h3 class="section-title">{{ $t('settings.shortcuts') }}</h3>
          <div class="shortcut-list">
            <div class="shortcut-row"><span>{{ $t('settings.shortcutNewDoc') }}</span><kbd>Ctrl</kbd>+<kbd>N</kbd></div>
            <div class="shortcut-row"><span>{{ $t('settings.shortcutSearch') }}</span><kbd>Ctrl</kbd>+<kbd>K</kbd></div>
            <div class="shortcut-row"><span>{{ $t('settings.shortcutSave') }}</span><kbd>Ctrl</kbd>+<kbd>S</kbd></div>
            <div class="shortcut-row"><span>{{ $t('settings.shortcutUndo') }}</span><kbd>Ctrl</kbd>+<kbd>Z</kbd></div>
            <div class="shortcut-row"><span>{{ $t('settings.shortcutRedo') }}</span><kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>Z</kbd></div>
            <div class="shortcut-row"><span>{{ $t('settings.shortcutBold') }}</span><kbd>Ctrl</kbd>+<kbd>B</kbd></div>
            <div class="shortcut-row"><span>{{ $t('settings.shortcutItalic') }}</span><kbd>Ctrl</kbd>+<kbd>I</kbd></div>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '@/store/modules/user'
import { updateUserInfo, changePassword } from '@/api/modules/user'
import { getStorageUsage, getUserStorageSettings, saveUserStorageSettings } from '@/api/modules/admin'
import { ElMessage, type FormInstance } from 'element-plus'

const { t } = useI18n()
const userStore = useUserStore()
const activeTab = ref('profile')
const saving = ref(false)
const changingPwd = ref(false)
const showUpgradeDialog = ref(false)
// const profileFormRef = ref<FormInstance>()
const pwdFormRef = ref<FormInstance>()

const profileForm = reactive({ nickname: '', email: '', avatar: '', department: '' })
const pwdForm = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' })

// Avatar upload
function handleAvatarUpload(file: File) {
  const reader = new FileReader()
  reader.onload = (e) => {
    profileForm.avatar = e.target?.result as string
    ElMessage.success(t('settings.avatarUpdated'))
  }
  reader.readAsDataURL(file)
  return false // Prevent default upload
}

// Notification settings
const notifSettings = reactive({
  commentNotify: true,
  mentionNotify: true,
  permissionNotify: true,
  shareNotify: true,
  approvalNotify: true,
  inAppNotify: true,
  emailNotify: false,
  dndEnabled: false,
  dndStart: '22:00',
  dndEnd: '08:00',
})

// Storage settings
const storageUsed = ref(0)
const storageTotal = ref(0)
const storagePercent = computed(() => {
  if (storageTotal.value === 0) return 0
  return Math.round(storageUsed.value / storageTotal.value * 100)
})
const storageUsedText = computed(() => formatSize(storageUsed.value))
const storageTotalText = computed(() => formatSize(storageTotal.value))
const storageByType = ref({ doc: 0, sheet: 0, slide: 0, image: 0, other: 0 })
const storageSettings = reactive({
  autoCleanRecycle: true,
  versionRetention: 20,
})
// User storage path settings
const userStoragePaths = reactive({
  syncDir: '',
  downloadDir: '',
  autoSync: true,
})
const savingStoragePaths = ref(false)

// Appearance settings
const appearance = reactive({
  theme: 'light',
  fontSize: 15,
  editorWidth: 'medium',
  codeTheme: 'github',
  sidebarCollapsed: false,
  showDocIcons: true,
})

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
  return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
}

const pwdRules = {
  oldPassword: [{ required: true, message: t('settings.currentPasswordRequired'), trigger: 'blur' }],
  newPassword: [
    { required: true, message: t('settings.newPasswordRequired'), trigger: 'blur' },
    { min: 6, message: t('auth.passwordMin'), trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: t('settings.confirmNewPasswordRequired'), trigger: 'blur' },
    {
      validator: (_r: any, v: string, cb: (e?: Error) => void) => {
        if (v !== pwdForm.newPassword) cb(new Error(t('auth.passwordMismatch')))
        else cb()
      },
      trigger: 'blur',
    },
  ],
}

async function handleSaveProfile() {
  saving.value = true
  try {
    await updateUserInfo({ nickname: profileForm.nickname })
    ElMessage.success(t('settings.saveSuccess'))
    userStore.fetchUserInfo()
  } finally {
    saving.value = false
  }
}

async function handleChangePassword() {
  const valid = await pwdFormRef.value?.validate().catch(() => false)
  if (!valid) return
  changingPwd.value = true
  try {
    await changePassword({ oldPassword: pwdForm.oldPassword, newPassword: pwdForm.newPassword })
    ElMessage.success(t('settings.passwordChanged'))
    pwdForm.oldPassword = ''
    pwdForm.newPassword = ''
    pwdForm.confirmPassword = ''
  } finally {
    changingPwd.value = false
  }
}

function handleSaveNotifSettings() {
  ElMessage.success(t('settings.notifSettingsSaved'))
}

function handleSaveAppearance() {
  ElMessage.success(t('settings.appearanceSaved'))
}

// Fetch storage usage
async function fetchStorageUsage() {
  try {
    const res = await getStorageUsage()
    const data = res.data || {}
    storageUsed.value = data.totalSize || 0
    // Use a default total for now (could be from system config)
    storageTotal.value = 10 * 1024 * 1024 * 1024 // 10GB default
    if (data.byType) {
      storageByType.value = data.byType
    }
  } catch (err) {
    console.error('Failed to fetch storage usage:', err)
  }
}

// Fetch user storage paths
async function fetchUserStoragePaths() {
  try {
    const res = await getUserStorageSettings()
    const data = res.data || {}
    userStoragePaths.syncDir = data.syncDir || ''
    userStoragePaths.downloadDir = data.downloadDir || ''
    userStoragePaths.autoSync = data.autoSync !== false
  } catch (err) {
    console.error('Failed to fetch user storage paths:', err)
  }
}

// Save user storage paths
async function handleSaveStoragePaths() {
  savingStoragePaths.value = true
  try {
    await saveUserStorageSettings({
      syncDir: userStoragePaths.syncDir,
      downloadDir: userStoragePaths.downloadDir,
      autoSync: userStoragePaths.autoSync,
    })
    ElMessage.success(t('settings.pathSettingsSaved'))
  } catch (err: any) {
    ElMessage.error(err.message || t('common.saveFailed'))
  } finally {
    savingStoragePaths.value = false
  }
}

onMounted(() => {
  if (userStore.user) {
    profileForm.nickname = userStore.user.nickname
    profileForm.email = userStore.user.email
    profileForm.avatar = userStore.user.avatar
  }
  fetchStorageUsage()
  fetchUserStoragePaths()
})
</script>

<style scoped>
.settings-page {
  max-width: 760px;
}
.settings-title {
  font-size: 22px;
  font-weight: 600;
  margin-bottom: 24px;
}

/* Avatar */
.avatar-section {
  display: flex;
  align-items: center;
  gap: 16px;
}

/* Settings Sections */
.settings-section {
  margin-bottom: 28px;
}
.section-title {
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 16px;
  color: var(--kx-text-primary);
}
.setting-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 0;
  border-bottom: 1px solid var(--kx-border);
}
.setting-info {
  flex: 1;
  min-width: 0;
}
.setting-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--kx-text-primary);
  margin-bottom: 2px;
}
.setting-desc {
  font-size: 12px;
  color: var(--kx-text-placeholder);
}
.setting-sub-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 0 14px;
  border-bottom: 1px solid var(--kx-border);
}
.dnd-separator {
  font-size: 13px;
  color: var(--kx-text-secondary);
}

/* Storage */
.storage-overview {
  display: flex;
  gap: 40px;
  align-items: center;
}
.storage-chart {
  flex-shrink: 0;
}
.storage-chart-text {
  text-align: center;
}
.storage-used-text {
  font-size: 18px;
  font-weight: 600;
  color: var(--kx-text-primary);
}
.storage-total-text {
  font-size: 12px;
  color: var(--kx-text-placeholder);
}
.storage-breakdown {
  flex: 1;
}
.breakdown-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 0;
}
.breakdown-color {
  width: 12px;
  height: 12px;
  border-radius: 3px;
  flex-shrink: 0;
}
.breakdown-label {
  flex: 1;
  font-size: 13px;
}
.breakdown-size {
  font-size: 13px;
  color: var(--kx-text-secondary);
}

/* Theme */
.theme-grid {
  display: flex;
  gap: 16px;
}
.theme-card {
  width: 160px;
  padding: 12px;
  border: 2px solid var(--kx-border);
  border-radius: 10px;
  cursor: pointer;
  text-align: center;
  transition: all 0.2s;
}
.theme-card:hover {
  border-color: var(--kx-primary);
}
.theme-card.active {
  border-color: var(--kx-primary);
  background: rgba(51,112,255,0.04);
}
.theme-card span {
  font-size: 13px;
  margin-top: 8px;
  display: block;
}
.theme-preview {
  height: 80px;
  border-radius: 6px;
  display: flex;
  overflow: hidden;
  border: 1px solid var(--kx-border);
}
.light-preview {
  background: #fff;
}
.light-preview .tp-sidebar {
  width: 30%;
  background: #f7f8fa;
}
.light-preview .tp-content {
  flex: 1;
  padding: 8px;
}
.light-preview .tp-bar {
  height: 8px;
  background: #e8e8e8;
  border-radius: 2px;
  margin-bottom: 6px;
}
.light-preview .tp-line {
  height: 4px;
  background: #eee;
  border-radius: 2px;
  margin-bottom: 4px;
}
.light-preview .tp-line.short {
  width: 60%;
}
.dark-preview {
  background: #1e1e1e;
}
.dark-preview .tp-sidebar {
  width: 30%;
  background: #252525;
}
.dark-preview .tp-content {
  flex: 1;
  padding: 8px;
}
.dark-preview .tp-bar {
  height: 8px;
  background: #333;
  border-radius: 2px;
  margin-bottom: 6px;
}
.dark-preview .tp-line {
  height: 4px;
  background: #2a2a2a;
  border-radius: 2px;
  margin-bottom: 4px;
}
.dark-preview .tp-line.short {
  width: 60%;
}
.auto-preview {
  background: linear-gradient(135deg, #fff 50%, #1e1e1e 50%);
}
.auto-preview .tp-sidebar {
  width: 30%;
  background: linear-gradient(135deg, #f7f8fa 50%, #252525 50%);
}
.auto-preview .tp-content {
  flex: 1;
  padding: 8px;
}
.auto-preview .tp-bar {
  height: 8px;
  background: linear-gradient(135deg, #e8e8e8 50%, #333 50%);
  border-radius: 2px;
  margin-bottom: 6px;
}
.auto-preview .tp-line {
  height: 4px;
  background: linear-gradient(135deg, #eee 50%, #2a2a2a 50%);
  border-radius: 2px;
  margin-bottom: 4px;
}
.auto-preview .tp-line.short {
  width: 60%;
}

/* About */
.about-info {
  background: var(--kx-sidebar-bg, #f7f8fa);
  border-radius: 8px;
  padding: 16px;
}
.about-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  font-size: 14px;
  border-bottom: 1px solid var(--kx-border);
}
.about-row:last-child {
  border-bottom: none;
}
.about-label {
  color: var(--kx-text-secondary);
}

/* Shortcuts */
.shortcut-list {
  background: var(--kx-sidebar-bg, #f7f8fa);
  border-radius: 8px;
  padding: 12px 16px;
}
.shortcut-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  font-size: 13px;
  border-bottom: 1px solid var(--kx-border);
}
.shortcut-row:last-child {
  border-bottom: none;
}
kbd {
  display: inline-block;
  padding: 2px 6px;
  font-size: 12px;
  font-family: monospace;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 4px;
  box-shadow: 0 1px 1px rgba(0,0,0,0.06);
  margin: 0 2px;
}
</style>
