<template>
  <div class="storage-settings">
    <div class="page-header">
      <h2>云盘存储设置</h2>
    </div>

    <!-- Storage Overview -->
    <div class="storage-overview">
      <div class="overview-card">
        <div class="overview-icon" style="background: #e8f0fe">
          <el-icon :size="24" color="#3370ff"><Coin /></el-icon>
        </div>
        <div class="overview-info">
          <span class="overview-label">已用空间</span>
          <span class="overview-value">{{ formatSize(storageUsed) }}</span>
        </div>
      </div>
      <div class="overview-card">
        <div class="overview-icon" style="background: #e6f7ef">
          <el-icon :size="24" color="#36b37e"><Database /></el-icon>
        </div>
        <div class="overview-info">
          <span class="overview-label">总容量</span>
          <span class="overview-value">{{ formatSize(storageTotal) }}</span>
        </div>
      </div>
      <div class="overview-card">
        <div class="overview-icon" style="background: #fef3e0">
          <el-icon :size="24" color="#f5a623"><Document /></el-icon>
        </div>
        <div class="overview-info">
          <span class="overview-label">文件数量</span>
          <span class="overview-value">{{ fileCount }}</span>
        </div>
      </div>
    </div>

    <!-- Storage Progress -->
    <div class="storage-progress-section">
      <div class="progress-header">
        <span>存储使用情况</span>
        <span class="progress-text">{{ storagePercent }}%</span>
      </div>
      <el-progress :percentage="storagePercent" :stroke-width="12" :color="storagePercent > 90 ? '#f54a45' : '#3370ff'" />
      <div class="storage-breakdown">
        <div class="breakdown-item">
          <span class="breakdown-color" style="background: #3370ff"></span>
          <span>文档</span>
          <span class="breakdown-size">{{ formatSize(docSize) }}</span>
        </div>
        <div class="breakdown-item">
          <span class="breakdown-color" style="background: #36b37e"></span>
          <span>表格</span>
          <span class="breakdown-size">{{ formatSize(sheetSize) }}</span>
        </div>
        <div class="breakdown-item">
          <span class="breakdown-color" style="background: #ff7d00"></span>
          <span>其他</span>
          <span class="breakdown-size">{{ formatSize(otherSize) }}</span>
        </div>
      </div>
    </div>

    <!-- Default Local Directory -->
    <div class="settings-section">
      <div class="section-header">
        <span>默认本地目录</span>
      </div>
      <el-form :model="localDirSettings" label-width="120px">
        <el-form-item label="同步目录">
          <el-input v-model="localDirSettings.syncDir" style="width: 400px">
            <template #append>
              <el-button @click="selectDirectory('sync')">浏览</el-button>
            </template>
          </el-input>
          <div class="form-tip">本地同步目录，用于离线访问和备份</div>
        </el-form-item>
        <el-form-item label="下载目录">
          <el-input v-model="localDirSettings.downloadDir" style="width: 400px">
            <template #append>
              <el-button @click="selectDirectory('download')">浏览</el-button>
            </template>
          </el-input>
          <div class="form-tip">文件下载的默认保存位置</div>
        </el-form-item>
        <el-form-item label="自动同步">
          <el-switch v-model="localDirSettings.autoSync" />
          <div class="form-tip">开启后，文件变更将自动同步到本地</div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="saveLocalDirSettings" :loading="savingLocalDir">保存设置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- Remote Storage -->
    <div class="settings-section">
      <div class="section-header">
        <span>远程存储接入</span>
        <el-button type="primary" @click="showAddStorageDialog = true">
          <el-icon><Plus /></el-icon>添加存储
        </el-button>
      </div>
      <el-table :data="remoteStorages" stripe>
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ row.type.toUpperCase() }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="server" label="服务器地址" min-width="200" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <span :class="['status-badge', row.status]">
              {{ row.status === 'connected' ? '已连接' : row.status === 'error' ? '连接失败' : '未连接' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="testStorageConnection(row)" :loading="row.testing">测试连接</el-button>
            <el-button link type="primary" @click="editStorage(row)">编辑</el-button>
            <el-button link type="danger" @click="deleteStorage(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- Storage Policy -->
    <div class="settings-section">
      <div class="section-header">
        <span>存储策略</span>
      </div>
      <el-form :model="storagePolicy" label-width="140px">
        <el-form-item label="文件版本保留">
          <el-select v-model="storagePolicy.versionRetention" style="width: 200px">
            <el-option label="保留最近10个版本" :value="10" />
            <el-option label="保留最近20个版本" :value="20" />
            <el-option label="保留最近50个版本" :value="50" />
            <el-option label="保留全部版本" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="回收站保留时间">
          <el-select v-model="storagePolicy.recycleRetention" style="width: 200px">
            <el-option label="7天" :value="7" />
            <el-option label="30天" :value="30" />
            <el-option label="90天" :value="90" />
            <el-option label="永久保留" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="大文件阈值">
          <el-input-number v-model="storagePolicy.largeFileThreshold" :min="10" :max="1000" />
          <span style="margin-left: 8px">MB</span>
          <div class="form-tip">超过此大小的文件将采用分块上传</div>
        </el-form-item>
        <el-form-item label="自动清理缓存">
          <el-switch v-model="storagePolicy.autoCleanCache" />
          <div class="form-tip">定期清理临时文件和缓存</div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="saveStoragePolicy" :loading="savingPolicy">保存策略</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- Add/Edit Storage Dialog -->
    <el-dialog v-model="showAddStorageDialog" :title="editingStorage ? '编辑存储' : '添加远程存储'" width="550px" destroy-on-close>
      <el-form :model="storageForm" :rules="storageRules" ref="storageFormRef" label-width="100px">
        <el-form-item label="存储名称" prop="name">
          <el-input v-model="storageForm.name" placeholder="请输入存储名称" />
        </el-form-item>
        <el-form-item label="存储类型" prop="type">
          <el-select v-model="storageForm.type" style="width: 100%">
            <el-option label="FTP" value="ftp" />
            <el-option label="SFTP" value="sftp" />
            <el-option label="SMB/CIFS" value="smb" />
            <el-option label="NFS" value="nfs" />
            <el-option label="WebDAV" value="webdav" />
          </el-select>
        </el-form-item>
        <el-form-item label="服务器地址" prop="server">
          <el-input v-model="storageForm.server" placeholder="如: 192.168.1.100 或 fileserver.local" />
        </el-form-item>
        <el-form-item label="端口">
          <el-input-number v-model="storageForm.port" :min="1" :max="65535" style="width: 150px" />
          <span class="port-hint">{{ getDefaultPort(storageForm.type) }}</span>
        </el-form-item>
        <el-form-item label="共享路径" v-if="['smb', 'nfs'].includes(storageForm.type)">
          <el-input v-model="storageForm.sharePath" placeholder="如: /shared/documents" />
        </el-form-item>
        <el-form-item label="用户名" v-if="['ftp', 'sftp', 'smb', 'webdav'].includes(storageForm.type)">
          <el-input v-model="storageForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码" v-if="['ftp', 'sftp', 'smb', 'webdav'].includes(storageForm.type)">
          <el-input v-model="storageForm.password" type="password" placeholder="请输入密码" show-password />
        </el-form-item>
        <el-form-item label="域" v-if="storageForm.type === 'smb'">
          <el-input v-model="storageForm.domain" placeholder="AD域（可选）" />
        </el-form-item>
        <el-form-item label="挂载点">
          <el-input v-model="storageForm.mountPoint" placeholder="如: /mnt/remote-storage">
            <template #prepend>/mnt/</template>
          </el-input>
          <div class="form-tip">远程存储在系统中的挂载路径</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddStorageDialog = false">取消</el-button>
        <el-button @click="testConnectionFromDialog" :loading="testingConnection">测试连接</el-button>
        <el-button type="primary" @click="submitStorage" :loading="submittingStorage">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

interface RemoteStorage {
  id: number
  name: string
  type: 'ftp' | 'sftp' | 'smb' | 'nfs' | 'webdav'
  server: string
  port: number
  username: string
  password: string
  sharePath: string
  mountPoint: string
  status: 'connected' | 'disconnected' | 'error'
  testing: boolean
}

// Storage stats
const storageUsed = ref(256 * 1024 * 1024 * 1024) // 256GB
const storageTotal = ref(1024 * 1024 * 1024 * 1024) // 1TB
const fileCount = ref(1234)
const docSize = ref(100 * 1024 * 1024 * 1024) // 100GB
const sheetSize = ref(80 * 1024 * 1024 * 1024) // 80GB
const otherSize = ref(76 * 1024 * 1024 * 1024) // 76GB

const storagePercent = computed(() => Math.round(storageUsed.value / storageTotal.value * 100))

function formatSize(bytes: number): string {
  if (bytes >= 1024 * 1024 * 1024 * 1024) {
    return (bytes / (1024 * 1024 * 1024 * 1024)).toFixed(1) + ' TB'
  }
  if (bytes >= 1024 * 1024 * 1024) {
    return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
  }
  if (bytes >= 1024 * 1024) {
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
  }
  return (bytes / 1024).toFixed(1) + ' KB'
}

// Local directory settings
const localDirSettings = ref({
  syncDir: '/Users/username/KloggerX',
  downloadDir: '/Users/username/Downloads/KloggerX',
  autoSync: true,
})
const savingLocalDir = ref(false)

function selectDirectory(type: string) {
  ElMessage.info('请在实际环境中选择目录')
}

async function saveLocalDirSettings() {
  savingLocalDir.value = true
  try {
    await new Promise(r => setTimeout(r, 500))
    ElMessage.success('本地目录设置已保存')
  } finally {
    savingLocalDir.value = false
  }
}

// Remote storages
const remoteStorages = ref<RemoteStorage[]>([])
const showAddStorageDialog = ref(false)
const editingStorage = ref<RemoteStorage | null>(null)
const storageFormRef = ref<FormInstance>()
const submittingStorage = ref(false)
const testingConnection = ref(false)

const storageForm = ref({
  name: '',
  type: 'smb' as 'ftp' | 'sftp' | 'smb' | 'nfs' | 'webdav',
  server: '',
  port: 0,
  username: '',
  password: '',
  sharePath: '',
  domain: '',
  mountPoint: '',
})

const storageRules: FormRules = {
  name: [{ required: true, message: '请输入存储名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择存储类型', trigger: 'change' }],
  server: [{ required: true, message: '请输入服务器地址', trigger: 'blur' }],
}

function getDefaultPort(type: string): string {
  const ports: Record<string, number> = { ftp: 21, sftp: 22, smb: 445, nfs: 2049, webdav: 80 }
  return `默认端口: ${ports[type] || '-'}`
}

function editStorage(s: RemoteStorage) {
  editingStorage.value = s
  storageForm.value = {
    name: s.name,
    type: s.type,
    server: s.server,
    port: s.port,
    username: s.username,
    password: s.password,
    sharePath: s.sharePath,
    domain: '',
    mountPoint: s.mountPoint,
  }
  showAddStorageDialog.value = true
}

async function testConnectionFromDialog() {
  testingConnection.value = true
  try {
    await new Promise(r => setTimeout(r, 1500))
    ElMessage.success('连接成功')
  } catch {
    ElMessage.error('连接失败')
  } finally {
    testingConnection.value = false
  }
}

async function submitStorage() {
  const valid = await storageFormRef.value?.validate().catch(() => false)
  if (!valid) return
  submittingStorage.value = true
  try {
    await new Promise(r => setTimeout(r, 500))
    if (editingStorage.value) {
      Object.assign(editingStorage.value, storageForm.value)
      ElMessage.success('存储已更新')
    } else {
      const newStorage: RemoteStorage = {
        id: Date.now(),
        ...storageForm.value,
        status: 'disconnected',
        testing: false,
      }
      remoteStorages.value.push(newStorage)
      ElMessage.success('存储已添加')
    }
    showAddStorageDialog.value = false
  } finally {
    submittingStorage.value = false
  }
}

async function testStorageConnection(s: RemoteStorage) {
  s.testing = true
  try {
    await new Promise(r => setTimeout(r, 1500))
    s.status = 'connected'
    ElMessage.success('连接成功')
  } catch {
    s.status = 'error'
    ElMessage.error('连接失败')
  } finally {
    s.testing = false
  }
}

async function deleteStorage(s: RemoteStorage) {
  await ElMessageBox.confirm(`确定要删除存储 "${s.name}" 吗？`, '删除确认')
  remoteStorages.value = remoteStorages.value.filter(x => x.id !== s.id)
  ElMessage.success('已删除')
}

// Storage policy
const storagePolicy = ref({
  versionRetention: 20,
  recycleRetention: 30,
  largeFileThreshold: 100,
  autoCleanCache: true,
})
const savingPolicy = ref(false)

async function saveStoragePolicy() {
  savingPolicy.value = true
  try {
    await new Promise(r => setTimeout(r, 500))
    ElMessage.success('存储策略已保存')
  } finally {
    savingPolicy.value = false
  }
}

async function fetchRemoteStorages() {
  remoteStorages.value = [
    {
      id: 1,
      name: '公司文件服务器',
      type: 'smb',
      server: '192.168.1.100',
      port: 445,
      username: 'admin',
      password: '***',
      sharePath: '/shared/docs',
      mountPoint: '/mnt/company-files',
      status: 'connected',
      testing: false,
    },
  ]
}

onMounted(() => {
  fetchRemoteStorages()
})
</script>

<style scoped>
.storage-settings {
  padding: 24px;
}
.page-header {
  margin-bottom: 24px;
}
.page-header h2 {
  margin: 0;
  font-size: 20px;
  color: var(--kx-text-primary);
}
.storage-overview {
  display: flex;
  gap: 20px;
  margin-bottom: 24px;
}
.overview-card {
  display: flex;
  align-items: center;
  gap: 16px;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  padding: 20px;
  min-width: 200px;
}
.overview-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.overview-info {
  display: flex;
  flex-direction: column;
}
.overview-label {
  font-size: 13px;
  color: var(--kx-text-secondary);
}
.overview-value {
  font-size: 20px;
  font-weight: 600;
  color: var(--kx-text-primary);
}
.storage-progress-section {
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 24px;
}
.progress-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 12px;
  font-weight: 500;
}
.progress-text {
  color: var(--kx-primary);
}
.storage-breakdown {
  display: flex;
  gap: 24px;
  margin-top: 16px;
}
.breakdown-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}
.breakdown-color {
  width: 12px;
  height: 12px;
  border-radius: 3px;
}
.breakdown-size {
  color: var(--kx-text-secondary);
}
.settings-section {
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 24px;
}
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  font-weight: 500;
  font-size: 16px;
}
.form-tip {
  font-size: 12px;
  color: var(--kx-text-placeholder);
  margin-top: 4px;
}
.port-hint {
  margin-left: 12px;
  font-size: 12px;
  color: var(--kx-text-placeholder);
}
.status-badge {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
}
.status-badge.connected {
  background: #e6f7ef;
  color: #36b37e;
}
.status-badge.disconnected {
  background: #f0f0f0;
  color: #646a73;
}
.status-badge.error {
  background: #fef0f0;
  color: #f54a45;
}
</style>
