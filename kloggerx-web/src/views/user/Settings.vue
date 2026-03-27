<template>
  <div class="settings-page">
    <h2 class="settings-title">个人设置</h2>
    <el-tabs v-model="activeTab">
      <!-- Profile Tab -->
      <el-tab-pane label="基本资料" name="profile">
        <el-form ref="profileFormRef" :model="profileForm" label-width="80px" style="max-width: 480px">
          <el-form-item label="头像">
            <div class="avatar-section">
              <el-avatar :size="64" :src="profileForm.avatar">{{ profileForm.nickname?.[0] || 'U' }}</el-avatar>
              <el-upload
                :show-file-list="false"
                :before-upload="handleAvatarUpload"
                accept="image/*"
              >
                <el-button size="small">更换头像</el-button>
              </el-upload>
            </div>
          </el-form-item>
          <el-form-item label="昵称" prop="nickname">
            <el-input v-model="profileForm.nickname" maxlength="20" show-word-limit />
          </el-form-item>
          <el-form-item label="邮箱">
            <el-input :model-value="profileForm.email" disabled />
          </el-form-item>
          <el-form-item label="部门">
            <el-input :model-value="profileForm.department || '未设置'" disabled />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="saving" @click="handleSaveProfile">保存</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- Password Tab -->
      <el-tab-pane label="修改密码" name="password">
        <el-form ref="pwdFormRef" :model="pwdForm" :rules="pwdRules" label-width="100px" style="max-width: 480px">
          <el-form-item label="当前密码" prop="oldPassword">
            <el-input v-model="pwdForm.oldPassword" type="password" show-password />
          </el-form-item>
          <el-form-item label="新密码" prop="newPassword">
            <el-input v-model="pwdForm.newPassword" type="password" show-password />
          </el-form-item>
          <el-form-item label="确认新密码" prop="confirmPassword">
            <el-input v-model="pwdForm.confirmPassword" type="password" show-password />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="changingPwd" @click="handleChangePassword">修改密码</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- Notification Tab -->
      <el-tab-pane label="通知设置" name="notifications">
        <div class="settings-section">
          <h3 class="section-title">消息通知</h3>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">文档评论通知</div>
              <div class="setting-desc">当有人在您的文档中发表评论时通知您</div>
            </div>
            <el-switch v-model="notifSettings.commentNotify" />
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">@提及通知</div>
              <div class="setting-desc">当有人在文档或评论中@您时通知您</div>
            </div>
            <el-switch v-model="notifSettings.mentionNotify" />
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">权限变更通知</div>
              <div class="setting-desc">当文档权限发生变化时通知您</div>
            </div>
            <el-switch v-model="notifSettings.permissionNotify" />
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">分享通知</div>
              <div class="setting-desc">当有人与您分享文档或知识库时通知您</div>
            </div>
            <el-switch v-model="notifSettings.shareNotify" />
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">审核通知</div>
              <div class="setting-desc">知识库文档审核相关的通知</div>
            </div>
            <el-switch v-model="notifSettings.approvalNotify" />
          </div>
        </div>
        <div class="settings-section">
          <h3 class="section-title">通知方式</h3>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">站内通知</div>
              <div class="setting-desc">在平台内显示通知消息</div>
            </div>
            <el-switch v-model="notifSettings.inAppNotify" />
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">邮件通知</div>
              <div class="setting-desc">通过邮件发送重要通知</div>
            </div>
            <el-switch v-model="notifSettings.emailNotify" />
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">免打扰时段</div>
              <div class="setting-desc">在指定时间段内不发送通知</div>
            </div>
            <el-switch v-model="notifSettings.dndEnabled" />
          </div>
          <div v-if="notifSettings.dndEnabled" class="setting-sub-row">
            <el-time-select v-model="notifSettings.dndStart" :max-time="notifSettings.dndEnd" placeholder="开始时间" start="00:00" step="00:30" end="23:30" />
            <span class="dnd-separator">至</span>
            <el-time-select v-model="notifSettings.dndEnd" :min-time="notifSettings.dndStart" placeholder="结束时间" start="00:00" step="00:30" end="23:30" />
          </div>
        </div>
        <el-button type="primary" style="margin-top:16px" @click="handleSaveNotifSettings">保存通知设置</el-button>
      </el-tab-pane>

      <!-- Storage Tab -->
      <el-tab-pane label="存储空间" name="storage">
        <div class="settings-section">
          <h3 class="section-title">存储概览</h3>
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
                <span class="breakdown-label">文档</span>
                <span class="breakdown-size">128.5 MB</span>
              </div>
              <div class="breakdown-item">
                <div class="breakdown-color" style="background:#36b37e" />
                <span class="breakdown-label">表格</span>
                <span class="breakdown-size">45.2 MB</span>
              </div>
              <div class="breakdown-item">
                <div class="breakdown-color" style="background:#ff7d00" />
                <span class="breakdown-label">幻灯片</span>
                <span class="breakdown-size">52.8 MB</span>
              </div>
              <div class="breakdown-item">
                <div class="breakdown-color" style="background:#9254de" />
                <span class="breakdown-label">其他文件</span>
                <span class="breakdown-size">29.5 MB</span>
              </div>
            </div>
          </div>
        </div>
        <div class="settings-section">
          <h3 class="section-title">存储管理</h3>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">自动清理回收站</div>
              <div class="setting-desc">回收站中超过30天的文档将自动清除</div>
            </div>
            <el-switch v-model="storageSettings.autoCleanRecycle" />
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">版本保留策略</div>
              <div class="setting-desc">自动保留的历史版本数量</div>
            </div>
            <el-select v-model="storageSettings.versionRetention" style="width:120px">
              <el-option :value="10" label="最近10个" />
              <el-option :value="20" label="最近20个" />
              <el-option :value="50" label="最近50个" />
              <el-option :value="100" label="最近100个" />
            </el-select>
          </div>
          <el-button type="primary" style="margin-top:16px" @click="showUpgradeDialog = true">升级存储空间</el-button>
        </div>
      </el-tab-pane>

      <!-- Appearance Tab -->
      <el-tab-pane label="外观设置" name="appearance">
        <div class="settings-section">
          <h3 class="section-title">主题</h3>
          <div class="theme-grid">
            <div class="theme-card" :class="{ active: appearance.theme === 'light' }" @click="appearance.theme = 'light'">
              <div class="theme-preview light-preview">
                <div class="tp-sidebar" /><div class="tp-content"><div class="tp-bar" /><div class="tp-line" /><div class="tp-line short" /></div>
              </div>
              <span>浅色模式</span>
            </div>
            <div class="theme-card" :class="{ active: appearance.theme === 'dark' }" @click="appearance.theme = 'dark'">
              <div class="theme-preview dark-preview">
                <div class="tp-sidebar" /><div class="tp-content"><div class="tp-bar" /><div class="tp-line" /><div class="tp-line short" /></div>
              </div>
              <span>深色模式</span>
            </div>
            <div class="theme-card" :class="{ active: appearance.theme === 'auto' }" @click="appearance.theme = 'auto'">
              <div class="theme-preview auto-preview">
                <div class="tp-sidebar" /><div class="tp-content"><div class="tp-bar" /><div class="tp-line" /><div class="tp-line short" /></div>
              </div>
              <span>跟随系统</span>
            </div>
          </div>
        </div>
        <div class="settings-section">
          <h3 class="section-title">编辑器</h3>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">字体大小</div>
              <div class="setting-desc">调整编辑器的默认字体大小</div>
            </div>
            <el-select v-model="appearance.fontSize" style="width:100px">
              <el-option :value="13" label="小" />
              <el-option :value="15" label="中" />
              <el-option :value="17" label="大" />
              <el-option :value="19" label="特大" />
            </el-select>
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">编辑器宽度</div>
              <div class="setting-desc">设置编辑区域的最大宽度</div>
            </div>
            <el-select v-model="appearance.editorWidth" style="width:120px">
              <el-option value="narrow" label="窄 (680px)" />
              <el-option value="medium" label="中 (800px)" />
              <el-option value="wide" label="宽 (960px)" />
              <el-option value="full" label="全宽" />
            </el-select>
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">代码高亮主题</div>
              <div class="setting-desc">代码块的语法高亮主题</div>
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
          <h3 class="section-title">侧边栏</h3>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">默认折叠侧边栏</div>
              <div class="setting-desc">启动时自动折叠左侧导航栏</div>
            </div>
            <el-switch v-model="appearance.sidebarCollapsed" />
          </div>
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-name">显示文档图标</div>
              <div class="setting-desc">在文档列表中显示类型图标</div>
            </div>
            <el-switch v-model="appearance.showDocIcons" />
          </div>
        </div>
        <el-button type="primary" style="margin-top:16px" @click="handleSaveAppearance">保存外观设置</el-button>
      </el-tab-pane>

      <!-- About Tab -->
      <el-tab-pane label="关于" name="about">
        <div class="settings-section">
          <h3 class="section-title">关于 KloggerX</h3>
          <div class="about-info">
            <div class="about-row"><span class="about-label">版本</span><span>v1.0.0</span></div>
            <div class="about-row"><span class="about-label">前端框架</span><span>Vue 3 + Element Plus</span></div>
            <div class="about-row"><span class="about-label">编辑器</span><span>TipTap + ProseMirror</span></div>
            <div class="about-row"><span class="about-label">协作引擎</span><span>Yjs + WebSocket</span></div>
          </div>
        </div>
        <div class="settings-section">
          <h3 class="section-title">快捷键</h3>
          <div class="shortcut-list">
            <div class="shortcut-row"><span>新建文档</span><kbd>Ctrl</kbd>+<kbd>N</kbd></div>
            <div class="shortcut-row"><span>搜索</span><kbd>Ctrl</kbd>+<kbd>K</kbd></div>
            <div class="shortcut-row"><span>保存</span><kbd>Ctrl</kbd>+<kbd>S</kbd></div>
            <div class="shortcut-row"><span>撤销</span><kbd>Ctrl</kbd>+<kbd>Z</kbd></div>
            <div class="shortcut-row"><span>重做</span><kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>Z</kbd></div>
            <div class="shortcut-row"><span>加粗</span><kbd>Ctrl</kbd>+<kbd>B</kbd></div>
            <div class="shortcut-row"><span>斜体</span><kbd>Ctrl</kbd>+<kbd>I</kbd></div>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useUserStore } from '@/store/modules/user'
import { updateUserInfo, changePassword } from '@/api/modules/user'
import { ElMessage, type FormInstance } from 'element-plus'

const userStore = useUserStore()
const activeTab = ref('profile')
const saving = ref(false)
const changingPwd = ref(false)
const showUpgradeDialog = ref(false)
const profileFormRef = ref<FormInstance>()
const pwdFormRef = ref<FormInstance>()

const profileForm = reactive({ nickname: '', email: '', avatar: '', department: '' })
const pwdForm = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' })

// Avatar upload
function handleAvatarUpload(file: File) {
  const reader = new FileReader()
  reader.onload = (e) => {
    profileForm.avatar = e.target?.result as string
    ElMessage.success('头像已更新')
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
const storageUsed = ref(256 * 1024 * 1024)
const storageTotal = ref(10 * 1024 * 1024 * 1024)
const storagePercent = computed(() => Math.round(storageUsed.value / storageTotal.value * 100))
const storageUsedText = computed(() => formatSize(storageUsed.value))
const storageTotalText = computed(() => formatSize(storageTotal.value))
const storageSettings = reactive({
  autoCleanRecycle: true,
  versionRetention: 20,
})

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
  oldPassword: [{ required: true, message: '请输入当前密码', trigger: 'blur' }],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (_r: any, v: string, cb: (e?: Error) => void) => {
        if (v !== pwdForm.newPassword) cb(new Error('两次密码不一致'))
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
    ElMessage.success('保存成功')
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
    ElMessage.success('密码修改成功')
    pwdForm.oldPassword = ''
    pwdForm.newPassword = ''
    pwdForm.confirmPassword = ''
  } finally {
    changingPwd.value = false
  }
}

function handleSaveNotifSettings() {
  ElMessage.success('通知设置已保存')
}

function handleSaveAppearance() {
  ElMessage.success('外观设置已保存')
}

onMounted(() => {
  if (userStore.user) {
    profileForm.nickname = userStore.user.nickname
    profileForm.email = userStore.user.email
    profileForm.avatar = userStore.user.avatar
  }
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
