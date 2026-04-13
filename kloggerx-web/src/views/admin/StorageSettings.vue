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

    <!-- System Default Storage Path -->
    <div class="settings-section">
      <div class="section-header">
        <span>系统默认存储路径</span>
      </div>
      <el-form :model="systemStoragePaths" label-width="120px">
        <el-form-item label="云盘存储路径">
          <el-input v-model="systemStoragePaths.uploadPath" style="width: 400px">
            <template #append>
              <el-button @click="openDirSelector('upload')">浏览</el-button>
            </template>
          </el-input>
          <div class="form-tip">云盘文件的默认存储位置（系统级设置）</div>
        </el-form-item>
        <el-form-item label="远程文件路径">
          <el-input v-model="systemStoragePaths.remotePath" style="width: 400px">
            <template #append>
              <el-button @click="openDirSelector('remote')">浏览</el-button>
            </template>
          </el-input>
          <div class="form-tip">远程存储接入的文件更新存放位置</div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="saveSystemStoragePaths" :loading="savingSystemPaths">保存路径设置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- Directory Selector Dialog -->
    <el-dialog v-model="showDirSelector" title="选择目录" width="500px">
      <div class="dir-selector">
        <div class="dir-path">
          <el-input v-model="currentDirPath" placeholder="输入或选择目录路径">
            <template #prepend>
              <el-button @click="goToParentDir" :disabled="!currentDirPath || currentDirPath === '/'">
                <el-icon><ArrowUp /></el-icon>
              </el-button>
            </template>
          </el-input>
        </div>
        <div class="dir-list" v-loading="loadingDirs">
          <div
            v-for="dir in availableDirs"
            :key="dir.path"
            class="dir-item"
            :class="{ selected: selectedDirPath === dir.path }"
            @click="selectDir(dir)"
            @dblclick="enterDir(dir)"
          >
            <el-icon><Folder /></el-icon>
            <span>{{ dir.name }}</span>
          </div>
          <div v-if="availableDirs.length === 0 && !loadingDirs" class="no-dirs">
            暂无子目录
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="showDirSelector = false">取消</el-button>
        <el-button type="primary" @click="confirmDirSelection">确定</el-button>
      </template>
    </el-dialog>

    <!-- Remote Storage -->
    <div class="settings-section">
      <div class="section-header">
        <span>远程存储接入</span>
        <el-button type="primary" @click="showAddStorageDialog = true">
          <el-icon><Plus /></el-icon>添加存储
        </el-button>
      </div>
      <el-table :data="remoteStorages" stripe>
        <el-table-column prop="name" label="名称" min-width="100" show-overflow-tooltip align="center"/>
        <el-table-column prop="type" label="类型" width="120" align="center">
          <template #default="{ row }">
            <el-tag size="small">{{ row.type.toUpperCase() }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="server" label="服务器地址" min-width="130" show-overflow-tooltip align="center"/>
        <el-table-column prop="status" label="状态" width="120" align="center">
          <template #default="{ row }">
            <span :class="['status-badge', row.status]">
              {{ row.status === 'connected' ? '已连接' : row.status === 'error' ? '连接失败' : '未连接' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="80" align="center">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button link size="small" type="primary" @click="testStorageConnection(row)" :loading="row.testing">测试</el-button>
              <el-button link size="small" type="primary" @click="reconnectStorage(row)" :loading="row.reconnecting">重连</el-button>
              <el-button link size="small" type="warning" @click="disconnectStorage(row)" :loading="row.disconnecting">断开</el-button>
              <el-button link size="small" type="primary" @click="editStorage(row)">编辑</el-button>
              <el-button link size="small" type="danger" @click="deleteStorage(row)">删除</el-button>
            </div>
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
          <el-select v-model="storageForm.type" style="width: 100%" @change="onStorageTypeChange">
            <el-option-group label="网络协议">
              <el-option label="FTP" value="ftp" />
              <el-option label="SFTP" value="sftp" />
              <el-option label="SMB/CIFS" value="smb" />
              <el-option label="NFS" value="nfs" />
              <el-option label="WebDAV" value="webdav" />
            </el-option-group>
            <el-option-group label="云盘服务">
              <el-option label="百度网盘" value="baidu" />
              <el-option label="阿里云盘" value="aliyun" />
              <el-option label="腾讯微盘" value="tencent" />
            </el-option-group>
          </el-select>
        </el-form-item>
        <el-form-item label="服务器地址" prop="server" v-if="!['baidu', 'aliyun', 'tencent'].includes(storageForm.type)">
          <el-input v-model="storageForm.server" placeholder="如: 192.168.1.100 或 fileserver.local" />
        </el-form-item>
        <el-form-item label="端口" v-if="!['baidu', 'aliyun', 'tencent'].includes(storageForm.type)">
          <el-input-number v-model="storageForm.port" :min="1" :max="65535" style="width: 150px" />
          <span class="port-hint">{{ getDefaultPort(storageForm.type) }}</span>
        </el-form-item>
        <el-form-item label="共享路径" v-if="['smb', 'nfs', 'ftp', 'sftp'].includes(storageForm.type)">
          <el-input v-model="storageForm.sharePath" :placeholder="storageForm.type === 'ftp' || storageForm.type === 'sftp' ? '可选：如 /home/user/documents' : '如: /shared/documents'" />
          <div class="form-tip" v-if="storageForm.type === 'ftp' || storageForm.type === 'sftp'">初始目录路径（可选），留空则进入根目录</div>
        </el-form-item>
        <el-form-item label="用户名" v-if="['ftp', 'sftp', 'smb', 'webdav'].includes(storageForm.type)">
          <el-input v-model="storageForm.username" :placeholder="['ftp', 'webdav'].includes(storageForm.type) ? '留空则匿名访问' : '请输入用户名'" />
          <div class="form-tip" v-if="storageForm.type === 'ftp'">留空或填 anonymous 进行匿名访问</div>
          <div class="form-tip" v-if="storageForm.type === 'webdav'">留空进行匿名访问</div>
          <div class="form-tip" v-if="storageForm.type === 'smb'">留空使用来宾(Guest)模式访问</div>
        </el-form-item>
        <el-form-item label="密码" v-if="['ftp', 'sftp', 'smb', 'webdav'].includes(storageForm.type) && storageForm.username !== '' && storageForm.username !== 'anonymous'">
          <el-input v-model="storageForm.password" type="password" placeholder="请输入密码" show-password />
        </el-form-item>
        <el-form-item label="域" v-if="storageForm.type === 'smb'">
          <el-input v-model="storageForm.domain" placeholder="AD域（可选）" />
        </el-form-item>
        <!-- Cloud drive OAuth tokens -->
        <el-form-item label="Access Token" v-if="['baidu', 'aliyun', 'tencent'].includes(storageForm.type)">
          <el-input v-model="storageForm.accessToken" type="textarea" :rows="3" placeholder="请输入 OAuth Access Token" />
          <div class="form-tip">从云盘开放平台获取的访问令牌</div>
        </el-form-item>
        <el-form-item label="Refresh Token" v-if="['aliyun'].includes(storageForm.type)">
          <el-input v-model="storageForm.refreshToken" type="textarea" :rows="2" placeholder="请输入 OAuth Refresh Token（可选）" />
          <div class="form-tip">用于自动刷新 Access Token</div>
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
import {
  getAdminStorageUsage,
  getStorageSettings,
  saveStorageSettings,
  listRemoteStorages,
  createRemoteStorage,
  updateRemoteStorage,
  deleteRemoteStorage,
  testRemoteStorage,
  testRemoteStorageConnection,
  connectRemoteStorage,
  disconnectRemoteStorage,
  listDirectories,
} from '@/api/modules/admin'

interface RemoteStorage {
  id: number
  name: string
  type: 'ftp' | 'sftp' | 'smb' | 'nfs' | 'webdav' | 'baidu' | 'aliyun' | 'tencent'
  server: string
  port: number
  username: string
  password: string
  sharePath: string
  mountPoint: string
  status: 'connected' | 'disconnected' | 'error'
  testing: boolean
  reconnecting: boolean
  disconnecting: boolean
}

// Storage stats - initialize with 0, load from API
const storageUsed = ref(0)
const storageTotal = ref(0)
const fileCount = ref(0)
const docSize = ref(0)
const sheetSize = ref(0)
const otherSize = ref(0)
const slideSize = ref(0)
const imageSize = ref(0)

const storagePercent = computed(() => {
  if (storageTotal.value === 0) return 0
  return Math.round(storageUsed.value / storageTotal.value * 100)
})

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

// System storage paths (admin only)
const systemStoragePaths = ref({
  uploadPath: './uploads',
  remotePath: './uploads/remote',
})
const savingSystemPaths = ref(false)

// Directory selector
const showDirSelector = ref(false)
const currentDirType = ref<'upload' | 'remote'>('upload')
const currentDirPath = ref('/')
const selectedDirPath = ref('')
const availableDirs = ref<{ name: string; path: string }[]>([])
const loadingDirs = ref(false)

function openDirSelector(type: 'upload' | 'remote') {
  currentDirType.value = type
  const currentPath = type === 'upload' ? systemStoragePaths.value.uploadPath : systemStoragePaths.value.remotePath
  currentDirPath.value = currentPath || '/'
  selectedDirPath.value = ''
  showDirSelector.value = true
  loadAvailableDirs()
}

async function loadAvailableDirs() {
  loadingDirs.value = true
  try {
    const res = await listDirectories(currentDirPath.value)
    availableDirs.value = res.data || []
  } catch (err: any) {
    ElMessage.error(err.message || '获取目录列表失败')
    availableDirs.value = []
  } finally {
    loadingDirs.value = false
  }
}

function selectDir(dir: { name: string; path: string }) {
  selectedDirPath.value = dir.path
}

function enterDir(dir: { name: string; path: string }) {
  currentDirPath.value = dir.path
  selectedDirPath.value = ''
  loadAvailableDirs()
}

function goToParentDir() {
  if (!currentDirPath.value || currentDirPath.value === '/') return
  const parts = currentDirPath.value.split('/').filter(Boolean)
  parts.pop()
  currentDirPath.value = '/' + parts.join('/')
  selectedDirPath.value = ''
  loadAvailableDirs()
}

function confirmDirSelection() {
  const finalPath = selectedDirPath.value || currentDirPath.value
  if (currentDirType.value === 'upload') {
    systemStoragePaths.value.uploadPath = finalPath
  } else {
    systemStoragePaths.value.remotePath = finalPath
  }
  showDirSelector.value = false
}

async function saveSystemStoragePaths() {
  savingSystemPaths.value = true
  try {
    const settings = await getStorageSettings()
    const data = settings.data || {}
    await saveStorageSettings({
      ...data,
      systemStoragePaths: systemStoragePaths.value,
    })
    ElMessage.success('系统存储路径已保存')
  } catch (err: any) {
    ElMessage.error(err.message || '保存失败')
  } finally {
    savingSystemPaths.value = false
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
  type: 'smb' as 'ftp' | 'sftp' | 'smb' | 'nfs' | 'webdav' | 'baidu' | 'aliyun' | 'tencent',
  server: '',
  port: 445, // Default SMB port
  username: '',
  password: '',
  sharePath: '',
  domain: '',
  mountPoint: '',
  accessToken: '',
  refreshToken: '',
})

const storageRules: FormRules = {
  name: [{ required: true, message: '请输入存储名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择存储类型', trigger: 'change' }],
  mountPoint: [{ required: true, message: '请输入挂载点', trigger: 'blur' }],
}

function getDefaultPort(type: string): string {
  const ports: Record<string, number> = { ftp: 21, sftp: 22, smb: 445, nfs: 2049, webdav: 80 }
  return `默认端口: ${ports[type] || '-'}`
}

// Update port when type changes
function onStorageTypeChange(type: string) {
  const ports: Record<string, number> = { ftp: 21, sftp: 22, smb: 445, nfs: 2049, webdav: 80 }
  storageForm.value.port = ports[type] || 0
}

function editStorage(s: RemoteStorage) {
  editingStorage.value = s
  storageForm.value = {
    name: s.name,
    type: s.type,
    server: s.server,
    port: s.port,
    username: s.username,
    password: '',
    sharePath: s.sharePath,
    domain: '',
    mountPoint: s.mountPoint,
    accessToken: '',
    refreshToken: '',
  }
  showAddStorageDialog.value = true
}

async function testConnectionFromDialog() {
  testingConnection.value = true
  try {
    await testRemoteStorage({
      type: storageForm.value.type,
      server: storageForm.value.server,
      port: storageForm.value.port,
      username: storageForm.value.username,
      password: storageForm.value.password,
    })
    ElMessage.success('连接成功')
  } catch (err: any) {
    ElMessage.error(err.message || '连接失败')
  } finally {
    testingConnection.value = false
  }
}

async function submitStorage() {
  const valid = await storageFormRef.value?.validate().catch(() => false)
  if (!valid) return
  submittingStorage.value = true
  try {
    const data: any = {
      name: storageForm.value.name,
      type: storageForm.value.type,
      server: storageForm.value.server,
      port: storageForm.value.port,
      username: storageForm.value.username,
      password: storageForm.value.password,
      sharePath: storageForm.value.sharePath,
      domain: storageForm.value.domain,
      mountPoint: storageForm.value.mountPoint,
      isEnabled: true,
    }
    
    // Add cloud drive tokens if applicable
    if (['baidu', 'aliyun', 'tencent'].includes(storageForm.value.type)) {
      data.accessToken = storageForm.value.accessToken
      data.refreshToken = storageForm.value.refreshToken
      data.server = '' // Cloud drives don't need server
      data.port = 0
    }

    if (editingStorage.value) {
      await updateRemoteStorage(editingStorage.value.id, data)
      ElMessage.success('存储已更新')
    } else {
      await createRemoteStorage(data)
      ElMessage.success('存储已添加')
    }
    showAddStorageDialog.value = false
    fetchRemoteStorages()
  } catch (err: any) {
    ElMessage.error(err.message || '操作失败')
  } finally {
    submittingStorage.value = false
  }
}

async function testStorageConnection(s: RemoteStorage) {
  s.testing = true
  try {
    await testRemoteStorageConnection(s.id)
    s.status = 'connected'
    ElMessage.success('连接成功')
  } catch (err: any) {
    s.status = 'error'
    ElMessage.error(err.message || '连接失败')
  } finally {
    s.testing = false
  }
}

async function reconnectStorage(s: RemoteStorage) {
  s.reconnecting = true
  try {
    await connectRemoteStorage(s.id)
    s.status = 'connected'
    ElMessage.success('重新连接成功')
  } catch (err: any) {
    s.status = 'error'
    ElMessage.error(err.message || '重新连接失败')
  } finally {
    s.reconnecting = false
  }
}

async function disconnectStorage(s: RemoteStorage) {
  s.disconnecting = true
  try {
    await disconnectRemoteStorage(s.id)
    s.status = 'disconnected'
    ElMessage.success('已断开连接')
  } catch (err: any) {
    ElMessage.error(err.message || '断开连接失败')
  } finally {
    s.disconnecting = false
  }
}

async function deleteStorage(s: RemoteStorage) {
  await ElMessageBox.confirm(`确定要删除存储 "${s.name}" 吗？`, '删除确认')
  try {
    await deleteRemoteStorage(s.id)
    remoteStorages.value = remoteStorages.value.filter(x => x.id !== s.id)
    ElMessage.success('已删除')
  } catch (err: any) {
    ElMessage.error(err.message || '删除失败')
  }
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
    const settings = await getStorageSettings()
    const data = settings.data || {}
    await saveStorageSettings({
      ...data,
      storagePolicy: storagePolicy.value,
    })
    ElMessage.success('存储策略已保存')
  } catch (err: any) {
    ElMessage.error(err.message || '保存失败')
  } finally {
    savingPolicy.value = false
  }
}

// Fetch storage statistics
async function fetchStorageStats() {
  try {
    const res = await getAdminStorageUsage()
    const data = res.data || {}
    const diskStats = data.diskStats || {}
    const byType = data.byType || {}

    storageUsed.value = diskStats.used || 0
    storageTotal.value = diskStats.total || 0
    fileCount.value = data.fileCount || 0

    docSize.value = byType.doc || 0
    sheetSize.value = byType.sheet || 0
    slideSize.value = byType.slide || 0
    imageSize.value = byType.image || 0
    otherSize.value = byType.other || 0
  } catch (err) {
    console.error('Failed to fetch storage stats:', err)
  }
}

// Fetch remote storages from API
async function fetchRemoteStorages() {
  try {
    const res = await listRemoteStorages()
    remoteStorages.value = (res.data || []).map((s: any) => ({
      id: s.id,
      name: s.name,
      type: s.type,
      server: s.server,
      port: s.port,
      username: s.username || '',
      password: '',
      sharePath: s.sharePath || '',
      mountPoint: s.mountPoint,
      status: s.status || 'disconnected',
      testing: false,
      reconnecting: false,
      disconnecting: false,
    }))
  } catch (err) {
    console.error('Failed to fetch remote storages:', err)
    remoteStorages.value = []
  }
}

// Fetch storage settings
async function fetchStorageSettings() {
  try {
    const res = await getStorageSettings()
    const data = res.data || {}

    // Load system storage paths
    if (data.systemStoragePaths) {
      systemStoragePaths.value = {
        uploadPath: data.systemStoragePaths.uploadPath || './uploads',
        remotePath: data.systemStoragePaths.remotePath || './uploads/remote',
      }
    }

    // Load storage policy
    if (data.storagePolicy) {
      storagePolicy.value = {
        versionRetention: data.storagePolicy.versionRetention || 20,
        recycleRetention: data.storagePolicy.recycleRetention || 30,
        largeFileThreshold: data.storagePolicy.largeFileThreshold || 100,
        autoCleanCache: data.storagePolicy.autoCleanCache !== false,
      }
    }
  } catch (err) {
    console.error('Failed to fetch storage settings:', err)
  }
}

onMounted(() => {
  fetchStorageStats()
  fetchRemoteStorages()
  fetchStorageSettings()
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

/* Directory Selector */
.dir-selector {
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  overflow: hidden;
}
.dir-path {
  padding: 12px;
  background: #f5f6f7;
  border-bottom: 1px solid var(--kx-border);
}
.dir-list {
  max-height: 300px;
  overflow-y: auto;
  min-height: 150px;
}
.dir-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  cursor: pointer;
  transition: background 0.2s;
}
.dir-item:hover {
  background: #f5f6f7;
}
.dir-item.selected {
  background: #e8f0fe;
  color: #3370ff;
}
.no-dirs {
  padding: 40px;
  text-align: center;
  color: #8f959e;
}

/* Action Buttons */
.action-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.action-buttons .el-button {
  padding: 4px 8px;
}
</style>
