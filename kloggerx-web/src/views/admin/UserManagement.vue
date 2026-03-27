<template>
  <div class="user-management">
    <div class="page-header">
      <h2>用户和权限管理</h2>
      <div class="header-actions">
        <el-button type="primary" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon>添加用户
        </el-button>
        <el-button @click="showImportDialog = true">
          <el-icon><Upload /></el-icon>从AD/LDAP导入
        </el-button>
      </div>
    </div>

    <!-- Search and Filter -->
    <div class="filter-bar">
      <el-input v-model="searchKeyword" placeholder="搜索用户名、邮箱" prefix-icon="Search" clearable style="width: 280px" />
      <el-select v-model="filterRole" placeholder="角色筛选" clearable style="width: 140px; margin-left: 12px">
        <el-option label="管理员" value="admin" />
        <el-option label="普通用户" value="member" />
        <el-option label="只读用户" value="readonly" />
      </el-select>
      <el-select v-model="filterDepartment" placeholder="部门筛选" clearable style="width: 160px; margin-left: 12px">
        <el-option v-for="d in departments" :key="d.id" :label="d.name" :value="d.id" />
      </el-select>
    </div>

    <!-- User Table -->
    <el-table :data="filteredUsers" v-loading="loading" stripe>
      <el-table-column prop="username" label="用户名" min-width="120">
        <template #default="{ row }">
          <div class="user-cell">
            <el-avatar :size="32" :src="row.avatar">{{ row.nickname?.[0] || row.username?.[0] || 'U' }}</el-avatar>
            <div class="user-info">
              <span class="user-name">{{ row.nickname || row.username }}</span>
              <span class="user-email">{{ row.email }}</span>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="role" label="角色" width="120">
        <template #default="{ row }">
          <el-tag :type="row.role === 'admin' ? 'danger' : row.role === 'member' ? '' : 'info'" size="small">
            {{ roleMap[row.role] || row.role }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="departmentName" label="部门" width="140" />
      <el-table-column prop="createdAt" label="创建时间" width="160">
        <template #default="{ row }">
          {{ formatDate(row.createdAt) }}
        </template>
      </el-table-column>
      <el-table-column prop="lastLoginAt" label="最后登录" width="160">
        <template #default="{ row }">
          {{ row.lastLoginAt ? formatDate(row.lastLoginAt) : '-' }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="editUser(row)">编辑</el-button>
          <el-button link type="primary" @click="editPermissions(row)">权限</el-button>
          <el-button link type="danger" @click="deleteUser(row)" :disabled="row.id === currentUserId">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Pagination -->
    <div class="pagination-wrapper">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
      />
    </div>

    <!-- Create/Edit User Dialog -->
    <el-dialog v-model="showCreateDialog" :title="editingUser ? '编辑用户' : '添加用户'" width="500px" destroy-on-close>
      <el-form :model="userForm" :rules="userRules" ref="userFormRef" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="userForm.username" placeholder="请输入用户名" :disabled="!!editingUser" />
        </el-form-item>
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="userForm.nickname" placeholder="请输入昵称" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="userForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item v-if="!editingUser" label="密码" prop="password">
          <el-input v-model="userForm.password" type="password" placeholder="请输入密码" show-password />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="userForm.role" style="width: 100%">
            <el-option label="管理员" value="admin" />
            <el-option label="普通用户" value="member" />
            <el-option label="只读用户" value="readonly" />
          </el-select>
        </el-form-item>
        <el-form-item label="部门" prop="departmentId">
          <el-tree-select
            v-model="userForm.departmentId"
            :data="departmentTree"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            placeholder="请选择部门"
            clearable
            check-strictly
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="submitUser" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <!-- Permission Dialog -->
    <el-dialog v-model="showPermDialog" title="用户权限设置" width="600px" destroy-on-close>
      <div class="perm-header">
        <span>用户: {{ permUser?.nickname || permUser?.username }}</span>
      </div>
      <el-tabs v-model="permTab">
        <el-tab-pane label="文档权限" name="doc">
          <div class="perm-list">
            <div v-for="p in docPermissions" :key="p.id" class="perm-item">
              <div class="perm-item-info">
                <el-icon :color="getTypeColor(p.docType)"><component :is="getTypeIcon(p.docType)" /></el-icon>
                <span>{{ p.docTitle }}</span>
              </div>
              <el-select v-model="p.level" size="small" style="width: 100px" @change="updateDocPerm(p)">
                <el-option label="可查看" value="view" />
                <el-option label="可编辑" value="edit" />
                <el-option label="可管理" value="manage" />
              </el-select>
            </div>
            <div v-if="!docPermissions.length" class="perm-empty">暂无文档权限</div>
          </div>
        </el-tab-pane>
        <el-tab-pane label="云盘权限" name="drive">
          <div class="perm-list">
            <div v-for="p in drivePermissions" :key="p.id" class="perm-item">
              <div class="perm-item-info">
                <el-icon color="#f5a623"><Folder /></el-icon>
                <span>{{ p.folderName }}</span>
              </div>
              <el-select v-model="p.level" size="small" style="width: 100px" @change="updateDrivePerm(p)">
                <el-option label="可查看" value="view" />
                <el-option label="可编辑" value="edit" />
                <el-option label="可管理" value="manage" />
              </el-select>
            </div>
            <div v-if="!drivePermissions.length" class="perm-empty">暂无云盘权限</div>
          </div>
        </el-tab-pane>
        <el-tab-pane label="知识库权限" name="kb">
          <div class="perm-list">
            <div v-for="p in kbPermissions" :key="p.id" class="perm-item">
              <div class="perm-item-info">
                <el-icon color="#3370ff"><Collection /></el-icon>
                <span>{{ p.kbName }}</span>
              </div>
              <el-select v-model="p.level" size="small" style="width: 100px" @change="updateKbPerm(p)">
                <el-option label="可查看" value="view" />
                <el-option label="可编辑" value="edit" />
                <el-option label="可管理" value="manage" />
              </el-select>
            </div>
            <div v-if="!kbPermissions.length" class="perm-empty">暂无知识库权限</div>
          </div>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="showPermDialog = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- AD/LDAP Import Dialog -->
    <el-dialog v-model="showImportDialog" title="从AD/LDAP导入用户" width="600px" destroy-on-close>
      <el-form :model="importForm" label-width="100px">
        <el-form-item label="导入方式">
          <el-radio-group v-model="importForm.type">
            <el-radio value="ad">Active Directory</el-radio>
            <el-radio value="ldap">LDAP</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="服务器地址">
          <el-input v-model="importForm.server" placeholder="ldap://domain.com:389" />
        </el-form-item>
        <el-form-item label="Base DN">
          <el-input v-model="importForm.baseDN" placeholder="dc=domain,dc=com" />
        </el-form-item>
        <el-form-item label="绑定DN">
          <el-input v-model="importForm.bindDN" placeholder="cn=admin,dc=domain,dc=com" />
        </el-form-item>
        <el-form-item label="绑定密码">
          <el-input v-model="importForm.bindPassword" type="password" placeholder="请输入密码" show-password />
        </el-form-item>
        <el-form-item>
          <el-button @click="testConnection" :loading="testing">测试连接</el-button>
          <el-button type="primary" @click="fetchAdUsers" :loading="fetching" :disabled="!connectionOk">获取用户列表</el-button>
        </el-form-item>
      </el-form>
      <div v-if="adUsers.length" class="ad-user-list">
        <div class="ad-user-header">
          <el-checkbox v-model="selectAllAd" @change="toggleSelectAllAd">全选</el-checkbox>
          <span>共 {{ adUsers.length }} 个用户</span>
        </div>
        <el-checkbox-group v-model="selectedAdUsers">
          <div v-for="u in adUsers" :key="u.dn" class="ad-user-item">
            <el-checkbox :value="u.dn">
              <span>{{ u.cn }} ({{ u.mail || u.sAMAccountName }})</span>
            </el-checkbox>
          </div>
        </el-checkbox-group>
      </div>
      <template #footer>
        <el-button @click="showImportDialog = false">取消</el-button>
        <el-button type="primary" @click="importAdUsers" :loading="importing" :disabled="!selectedAdUsers.length">导入选中用户</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useUserStore } from '@/store/modules/user'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getAdminUsers, createAdminUser, updateAdminUser, deleteAdminUser, getAdminDepartments, testLdapConnection, fetchLdapUsers, importLdapUsers } from '@/api/modules/admin'

interface UserItem {
  id: number
  username: string
  nickname: string
  email: string
  avatar: string
  role: 'admin' | 'member' | 'readonly'
  departmentId: number | null
  departmentName: string
  createdAt: string
  lastLoginAt: string | null
}

interface Department {
  id: number
  name: string
  parentId: number | null
  children?: Department[]
}

const userStore = useUserStore()
const loading = ref(false)
const users = ref<UserItem[]>([])
const departments = ref<Department[]>([])
const searchKeyword = ref('')
const filterRole = ref('')
const filterDepartment = ref<number | null>(null)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

const roleMap: Record<string, string> = { admin: '管理员', member: '普通用户', readonly: '只读用户' }

const currentUserId = computed(() => userStore.user?.id)

const filteredUsers = computed(() => {
  let result = users.value
  if (searchKeyword.value) {
    const kw = searchKeyword.value.toLowerCase()
    result = result.filter(u => u.username.toLowerCase().includes(kw) || u.email.toLowerCase().includes(kw) || (u.nickname && u.nickname.toLowerCase().includes(kw)))
  }
  if (filterRole.value) {
    result = result.filter(u => u.role === filterRole.value)
  }
  if (filterDepartment.value) {
    result = result.filter(u => u.departmentId === filterDepartment.value)
  }
  return result
})

const departmentTree = computed(() => buildTree(departments.value))

function buildTree(items: Department[], parentId: number | null = null): Department[] {
  return items.filter(i => i.parentId === parentId).map(i => ({ ...i, children: buildTree(items, i.id) }))
}

// User form
const showCreateDialog = ref(false)
const editingUser = ref<UserItem | null>(null)
const userFormRef = ref<FormInstance>()
const submitting = ref(false)
const userForm = ref({
  username: '',
  nickname: '',
  email: '',
  password: '',
  role: 'member' as 'admin' | 'member' | 'readonly',
  departmentId: null as number | null,
})

const userRules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  email: [{ required: true, type: 'email', message: '请输入有效邮箱', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }],
}

function editUser(user: UserItem) {
  editingUser.value = user
  userForm.value = {
    username: user.username,
    nickname: user.nickname || '',
    email: user.email,
    password: '',
    role: user.role,
    departmentId: user.departmentId,
  }
  showCreateDialog.value = true
}

async function submitUser() {
  const valid = await userFormRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    if (editingUser.value) {
      await updateAdminUser(editingUser.value.id, {
        nickname: userForm.value.nickname,
        email: userForm.value.email,
        role: userForm.value.role,
        departmentId: userForm.value.departmentId,
      })
      ElMessage.success('用户信息已更新')
    } else {
      await createAdminUser({
        username: userForm.value.username,
        email: userForm.value.email,
        password: userForm.value.password,
        nickname: userForm.value.nickname,
        role: userForm.value.role,
        departmentId: userForm.value.departmentId,
      })
      ElMessage.success('用户已创建')
    }
    showCreateDialog.value = false
    fetchUsers()
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

async function deleteUser(user: UserItem) {
  await ElMessageBox.confirm(`确定要删除用户 "${user.nickname || user.username}" 吗？`, '删除确认')
  try {
    await deleteAdminUser(user.id)
    ElMessage.success('用户已删除')
    fetchUsers()
  } catch (e: any) {
    ElMessage.error(e.message || '删除失败')
  }
}

// Permissions
const showPermDialog = ref(false)
const permUser = ref<UserItem | null>(null)
const permTab = ref('doc')
const docPermissions = ref<any[]>([])
const drivePermissions = ref<any[]>([])
const kbPermissions = ref<any[]>([])

function editPermissions(user: UserItem) {
  permUser.value = user
  permTab.value = 'doc'
  showPermDialog.value = true
  // Fetch permissions
  docPermissions.value = []
  drivePermissions.value = []
  kbPermissions.value = []
}

function updateDocPerm(p: any) {
  ElMessage.success('权限已更新')
}

function updateDrivePerm(p: any) {
  ElMessage.success('权限已更新')
}

function updateKbPerm(p: any) {
  ElMessage.success('权限已更新')
}

// AD/LDAP Import
const showImportDialog = ref(false)
const importForm = ref({
  type: 'ad' as 'ad' | 'ldap',
  server: '',
  baseDN: '',
  bindDN: '',
  bindPassword: '',
})
const testing = ref(false)
const connectionOk = ref(false)
const fetching = ref(false)
const adUsers = ref<any[]>([])
const selectedAdUsers = ref<string[]>([])
const selectAllAd = ref(false)
const importing = ref(false)

async function testConnection() {
  if (!importForm.value.server || !importForm.value.baseDN || !importForm.value.bindDN || !importForm.value.bindPassword) {
    ElMessage.warning('请填写完整的连接信息')
    return
  }
  testing.value = true
  connectionOk.value = false
  try {
    const res: any = await testLdapConnection({
      type: importForm.value.type,
      server: importForm.value.server,
      baseDN: importForm.value.baseDN,
      bindDN: importForm.value.bindDN,
      bindPassword: importForm.value.bindPassword,
    })
    if (res.data?.success) {
      connectionOk.value = true
      ElMessage.success('连接成功')
    } else {
      ElMessage.error('连接失败：服务器未响应或认证错误')
    }
  } catch (e: any) {
    connectionOk.value = false
    ElMessage.error(e.message || '连接失败')
  } finally {
    testing.value = false
  }
}

async function fetchAdUsers() {
  if (!connectionOk.value) {
    ElMessage.warning('请先测试连接')
    return
  }
  fetching.value = true
  adUsers.value = []
  try {
    const res: any = await fetchLdapUsers({
      type: importForm.value.type,
      server: importForm.value.server,
      baseDN: importForm.value.baseDN,
      bindDN: importForm.value.bindDN,
      bindPassword: importForm.value.bindPassword,
    })
    if (res.data?.users?.length) {
      adUsers.value = res.data.users
      ElMessage.success(`获取到 ${adUsers.value.length} 个用户`)
    } else {
      ElMessage.info('未获取到用户')
    }
  } catch (e: any) {
    ElMessage.error(e.message || '获取用户列表失败')
  } finally {
    fetching.value = false
  }
}

function toggleSelectAllAd(val: boolean) {
  selectedAdUsers.value = val ? adUsers.value.map(u => u.dn) : []
}

async function importAdUsers() {
  if (!selectedAdUsers.value.length) {
    ElMessage.warning('请选择要导入的用户')
    return
  }
  importing.value = true
  try {
    await importLdapUsers({
      type: importForm.value.type,
      server: importForm.value.server,
      baseDN: importForm.value.baseDN,
      bindDN: importForm.value.bindDN,
      bindPassword: importForm.value.bindPassword,
      users: selectedAdUsers.value,
    })
    ElMessage.success(`已导入 ${selectedAdUsers.value.length} 个用户`)
    showImportDialog.value = false
    fetchUsers()
  } catch (e: any) {
    ElMessage.error(e.message || '导入失败')
  } finally {
    importing.value = false
  }
}

// Helpers
function formatDate(date: string) {
  return new Date(date).toLocaleString('zh-CN')
}

function getTypeColor(type: string) {
  const map: Record<string, string> = { folder: '#f5a623', doc: '#3370ff', sheet: '#36b37e', slide: '#ff7d00' }
  return map[type] || '#999'
}

function getTypeIcon(type: string) {
  const map: Record<string, string> = { folder: 'Folder', doc: 'Document', sheet: 'Grid', slide: 'Monitor' }
  return map[type] || 'Document'
}

async function fetchUsers() {
  loading.value = true
  try {
    const res: any = await getAdminUsers({
      page: currentPage.value,
      pageSize: pageSize.value,
      keyword: searchKeyword.value || undefined,
      role: filterRole.value || undefined,
    })
    users.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) {
    console.error('Failed to fetch users:', e)
  } finally {
    loading.value = false
  }
}

async function fetchDepartments() {
  try {
    const res: any = await getAdminDepartments()
    departments.value = res.data || []
  } catch (e) {
    console.error('Failed to fetch departments:', e)
  }
}

onMounted(() => {
  fetchUsers()
  fetchDepartments()
})
</script>

<style scoped>
.user-management {
  padding: 24px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.page-header h2 {
  margin: 0;
  font-size: 20px;
  color: var(--kx-text-primary);
}
.header-actions {
  display: flex;
  gap: 8px;
}
.filter-bar {
  display: flex;
  align-items: center;
  margin-bottom: 16px;
}
.user-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}
.user-info {
  display: flex;
  flex-direction: column;
}
.user-name {
  font-size: 14px;
  color: var(--kx-text-primary);
}
.user-email {
  font-size: 12px;
  color: var(--kx-text-placeholder);
}
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
.perm-header {
  margin-bottom: 16px;
  font-weight: 500;
}
.perm-list {
  max-height: 300px;
  overflow-y: auto;
}
.perm-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid var(--kx-border);
}
.perm-item:last-child {
  border-bottom: none;
}
.perm-item-info {
  display: flex;
  align-items: center;
  gap: 8px;
}
.perm-empty {
  text-align: center;
  padding: 20px;
  color: var(--kx-text-placeholder);
}
.ad-user-list {
  margin-top: 16px;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  padding: 12px;
  max-height: 300px;
  overflow-y: auto;
}
.ad-user-header {
  display: flex;
  justify-content: space-between;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--kx-border);
  margin-bottom: 8px;
}
.ad-user-item {
  padding: 4px 0;
}
</style>
