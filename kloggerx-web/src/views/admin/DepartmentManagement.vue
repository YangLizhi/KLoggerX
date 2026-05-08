<template>
  <div class="department-management">
    <div class="page-header">
      <h2>{{ $t('admin.departmentManagement') }}</h2>
      <div class="header-actions">
        <el-button type="primary" @click="createDepartment(null)">
          <el-icon><Plus /></el-icon>{{ $t('admin.dept.addDept') }}
        </el-button>
        <el-button @click="showImportDialog = true">
          <el-icon><Upload /></el-icon>{{ $t('admin.dept.importOrg') }}
        </el-button>
      </div>
    </div>

    <div class="content-wrapper">
      <!-- Department Tree -->
      <div class="tree-panel">
        <div class="panel-header">
          <span>{{ $t('admin.dept.orgStructure') }}</span>
          <el-input v-model="searchKeyword" :placeholder="$t('admin.dept.searchDept')" prefix-icon="Search" size="small" clearable style="width: 180px" />
        </div>
        <div class="tree-content">
          <el-tree
            ref="treeRef"
            :data="departmentTree"
            :props="{ label: 'name', children: 'children' }"
            node-key="id"
            :expand-on-click-node="false"
            :default-expanded-keys="expandedKeys"
            highlight-current
            @node-click="selectDepartment"
          >
            <template #default="{ data }">
              <div class="tree-node">
                <el-icon color="#f5a623"><Folder /></el-icon>
                <span class="node-name">{{ data.name }}</span>
                <span class="node-count">({{ data.memberCount || 0 }}{{ $t('admin.dept.personUnit') }})</span>
                <div class="node-actions">
                  <el-icon v-if="getDepth(data) < 8" class="action-btn" @click.stop="createDepartment(data)" :title="$t('admin.dept.addSubDept')"><Plus /></el-icon>
                  <el-icon class="action-btn" @click.stop="editDepartment(data)" :title="$t('common.edit')"><EditPen /></el-icon>
                  <el-icon class="action-btn danger" @click.stop="deleteDepartment(data)" :title="$t('common.delete')"><Delete /></el-icon>
                </div>
              </div>
            </template>
          </el-tree>
        </div>
      </div>

      <!-- Department Detail -->
      <div class="detail-panel">
        <template v-if="selectedDept">
          <div class="detail-header">
            <h3>{{ selectedDept.name }}</h3>
            <el-button size="small" @click="editDepartment(selectedDept)">{{ $t('admin.dept.editDept') }}</el-button>
          </div>

          <!-- Members -->
          <div class="section">
            <div class="section-header">
              <span>{{ $t('admin.dept.deptMembers') }} ({{ deptMembers.length }}{{ $t('admin.dept.personUnit') }})</span>
              <el-button size="small" type="primary" @click="showAddMemberDialog = true">
                <el-icon><Plus /></el-icon>{{ $t('admin.dept.addMember') }}
              </el-button>
            </div>
            <div class="member-list">
              <div v-for="m in deptMembers" :key="m.id" class="member-item">
                <el-avatar :size="32" :src="m.avatar">{{ m.name?.[0] }}</el-avatar>
                <div class="member-info">
                  <span class="member-name">{{ m.name }}</span>
                  <span class="member-role">{{ m.role === 'manager' ? $t('admin.dept.manager') : $t('admin.dept.member') }}</span>
                </div>
                <el-button link size="small" @click="setAsManager(m)" v-if="m.role !== 'manager'">{{ $t('admin.dept.setAsManager') }}</el-button>
                <el-button link size="small" type="danger" @click="removeMember(m)">{{ $t('common.remove') }}</el-button>
              </div>
              <div v-if="!deptMembers.length" class="empty-tip">{{ $t('admin.dept.noMembers') }}</div>
            </div>
          </div>

          <!-- Permissions -->
          <div class="section">
            <div class="section-header">
              <span>{{ $t('admin.dept.deptPerms') }}</span>
              <el-button size="small" @click="showPermDialog = true">{{ $t('admin.dept.configPerms') }}</el-button>
            </div>
            <div class="perm-list">
              <div v-for="p in deptPermissions" :key="p.id" class="perm-item">
                <div class="perm-info">
                  <el-icon :color="p.type === 'doc' ? '#3370ff' : p.type === 'folder' ? '#f5a623' : '#36b37e'">
                    <component :is="p.type === 'doc' ? 'Document' : p.type === 'folder' ? 'Folder' : 'Collection'" />
                  </el-icon>
                  <span>{{ p.name }}</span>
                </div>
                <el-tag size="small">{{ permLevelMap[p.level] }}</el-tag>
              </div>
              <div v-if="!deptPermissions.length" class="empty-tip">{{ $t('admin.dept.noPerms') }}</div>
            </div>
          </div>
        </template>
        <div v-else class="empty-state">
          <el-icon :size="48" color="#c0c4cc"><Folder /></el-icon>
          <p>{{ $t('admin.dept.selectDeptHint') }}</p>
        </div>
      </div>
    </div>

    <!-- Create/Edit Department Dialog -->
    <el-dialog v-model="showDeptDialog" :title="editingDept ? $t('admin.dept.editDept') : $t('admin.dept.newDept')" width="450px" destroy-on-close>
      <el-form :model="deptForm" :rules="deptRules" ref="deptFormRef" label-width="80px">
        <el-form-item :label="$t('admin.dept.deptName')" prop="name">
          <el-input v-model="deptForm.name" :placeholder="$t('admin.dept.inputDeptName')" maxlength="50" />
        </el-form-item>
        <el-form-item :label="$t('admin.dept.parentDept')">
          <el-tree-select
            v-model="deptForm.parentId"
            :data="departmentTree"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            :placeholder="$t('admin.dept.noParent')"
            clearable
            check-strictly
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="$t('admin.dept.deptDesc')">
          <el-input v-model="deptForm.description" type="textarea" :rows="3" :placeholder="$t('admin.dept.inputDeptDesc')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDeptDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="submitDepartment" :loading="submitting">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- Add Member Dialog -->
    <el-dialog v-model="showAddMemberDialog" :title="$t('admin.dept.addMember')" width="500px" destroy-on-close>
      <el-transfer
        v-model="selectedMembers"
        :data="allUsers"
        :titles="[$t('admin.dept.availableUsers'), $t('admin.dept.selectedUsers')]"
        :props="{ key: 'id', label: 'name' }"
        filterable
        :filter-placeholder="$t('admin.dept.searchUser')"
      />
      <template #footer>
        <el-button @click="showAddMemberDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="addMembers" :loading="addingMembers">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- Import Dialog -->
    <el-dialog v-model="showImportDialog" :title="$t('admin.dept.importOrg')" width="500px" destroy-on-close>
      <el-upload
        drag
        :auto-upload="false"
        :limit="1"
        accept=".xlsx,.xls,.csv"
        :on-change="handleFileChange"
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">{{ $t('admin.dept.dragFile') }} <em>{{ $t('admin.dept.clickUpload') }}</em></div>
        <template #tip>
          <div class="el-upload__tip">{{ $t('admin.dept.uploadTip') }}</div>
        </template>
      </el-upload>
      <div v-if="importPreview.length" class="import-preview">
        <div class="preview-header">{{ $t('admin.dept.preview', { count: importPreview.length }) }}</div>
        <el-tree :data="importPreview" :props="{ label: 'name', children: 'children' }" default-expand-all />
      </div>
      <template #footer>
        <el-button @click="showImportDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="confirmImport" :loading="importing" :disabled="!importPreview.length">{{ $t('admin.dept.confirmImport') }}</el-button>
      </template>
    </el-dialog>

    <!-- Permission Config Dialog -->
    <el-dialog v-model="showPermDialog" :title="$t('admin.dept.configDeptPerms')" width="600px" destroy-on-close>
      <el-tabs v-model="permTab">
        <el-tab-pane :label="$t('admin.dept.docPerms')" name="doc">
          <div class="perm-config-list">
            <div v-for="p in docPermList" :key="p.id" class="perm-config-item">
              <span>{{ p.name }}</span>
              <el-select v-model="p.level" size="small" style="width: 120px">
                <el-option :label="$t('admin.dept.noAccess')" value="none" />
                <el-option :label="$t('admin.dept.canView')" value="view" />
                <el-option :label="$t('admin.dept.canEdit')" value="edit" />
                <el-option :label="$t('admin.dept.canManage')" value="manage" />
              </el-select>
            </div>
          </div>
        </el-tab-pane>
        <el-tab-pane :label="$t('admin.dept.folderPerms')" name="folder">
          <div class="perm-config-list">
            <div v-for="p in folderPermList" :key="p.id" class="perm-config-item">
              <span>{{ p.name }}</span>
              <el-select v-model="p.level" size="small" style="width: 120px">
                <el-option :label="$t('admin.dept.noAccess')" value="none" />
                <el-option :label="$t('admin.dept.canView')" value="view" />
                <el-option :label="$t('admin.dept.canEdit')" value="edit" />
                <el-option :label="$t('admin.dept.canManage')" value="manage" />
              </el-select>
            </div>
          </div>
        </el-tab-pane>
        <el-tab-pane :label="$t('admin.dept.kbPerms')" name="kb">
          <div class="perm-config-list">
            <div v-for="p in kbPermList" :key="p.id" class="perm-config-item">
              <span>{{ p.name }}</span>
              <el-select v-model="p.level" size="small" style="width: 120px">
                <el-option :label="$t('admin.dept.noAccess')" value="none" />
                <el-option :label="$t('admin.dept.canView')" value="view" />
                <el-option :label="$t('admin.dept.canEdit')" value="edit" />
                <el-option :label="$t('admin.dept.canManage')" value="manage" />
              </el-select>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="showPermDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="savePermissions" :loading="savingPerms">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'

const { t } = useI18n()
import type { FormInstance, FormRules } from 'element-plus'
import {
  getAdminDepartments,
  createAdminDepartment,
  updateAdminDepartment,
  deleteAdminDepartment,
  getDepartmentMembers,
} from '@/api/modules/admin'

interface Department {
  id: number
  name: string
  parentId: number | null
  memberCount: number
  description?: string
  children?: Department[]
}

interface Member {
  id: number
  name: string
  avatar: string
  role: 'manager' | 'member'
}

const searchKeyword = ref('')
const departments = ref<Department[]>([])
const selectedDept = ref<Department | null>(null)
const expandedKeys = ref<number[]>([])
// const treeRef = ref()

const departmentTree = computed(() => {
  const buildTree = (parentId: number | null = null): Department[] => {
    return departments.value
      .filter(d => d.parentId === parentId)
      .map(d => ({ ...d, children: buildTree(d.id) }))
  }
  return buildTree(null)
})

function getDepth(dept: Department): number {
  let depth = 0
  let current = departments.value.find(d => d.id === dept.parentId)
  while (current) {
    depth++
    current = departments.value.find(d => d.id === current?.parentId)
  }
  return depth
}

function selectDepartment(data: Department) {
  selectedDept.value = data
  fetchDeptMembers()
  fetchDeptPermissions()
}

// Department CRUD
const showDeptDialog = ref(false)
const editingDept = ref<Department | null>(null)
const deptFormRef = ref<FormInstance>()
const submitting = ref(false)
const deptForm = ref({
  name: '',
  parentId: null as number | null,
  description: '',
})

const deptRules: FormRules = {
  name: [{ required: true, message: () => t('admin.dept.inputDeptName'), trigger: 'blur' }],
}

function createDepartment(parent: Department | null) {
  editingDept.value = null
  deptForm.value = { name: '', parentId: parent?.id || null, description: '' }
  showDeptDialog.value = true
}

function editDepartment(dept: Department) {
  editingDept.value = dept
  deptForm.value = { name: dept.name, parentId: dept.parentId, description: dept.description || '' }
  showDeptDialog.value = true
}

async function submitDepartment() {
  const valid = await deptFormRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    if (editingDept.value) {
      await updateAdminDepartment(editingDept.value.id, {
        name: deptForm.value.name,
        parentId: deptForm.value.parentId || undefined,
      })
      ElMessage.success(t('admin.dept.deptUpdated'))
    } else {
      await createAdminDepartment({
        name: deptForm.value.name,
        parentId: deptForm.value.parentId || undefined,
      })
      ElMessage.success(t('admin.dept.deptCreated'))
    }
    showDeptDialog.value = false
    fetchDepartments()
  } catch (err: any) {
    ElMessage.error(err.message || t('common.operationFailed'))
  } finally {
    submitting.value = false
  }
}

async function deleteDepartment(dept: Department) {
  if (dept.memberCount && dept.memberCount > 0) {
    ElMessage.warning(t('admin.dept.hasMembersWarning'))
    return
  }
  await ElMessageBox.confirm(t('admin.dept.deleteConfirm', { name: dept.name }), t('admin.dept.deleteTitle'))
  try {
    await deleteAdminDepartment(dept.id)
    departments.value = departments.value.filter(d => d.id !== dept.id)
    if (selectedDept.value?.id === dept.id) {
      selectedDept.value = null
    }
    ElMessage.success(t('admin.dept.deptDeleted'))
  } catch (err: any) {
    ElMessage.error(err.message || t('admin.dept.deleteFailed'))
  }
}

// Members
const deptMembers = ref<Member[]>([])
const showAddMemberDialog = ref(false)
const selectedMembers = ref<number[]>([])
const allUsers = ref<{ id: number; name: string }[]>([])
const addingMembers = ref(false)

async function fetchDeptMembers() {
  if (!selectedDept.value) {
    deptMembers.value = []
    return
  }
  try {
    const res: any = await getDepartmentMembers(selectedDept.value.id)
    deptMembers.value = (res.data || []).map((m: any) => ({
      id: m.id,
      name: m.nickname || m.username || m.name,
      avatar: m.avatar || '',
      role: m.is_manager ? 'manager' : 'member',
    }))
  } catch {
    deptMembers.value = []
  }
}

function setAsManager(member: Member) {
  deptMembers.value.forEach(m => m.role = 'member')
  member.role = 'manager'
  ElMessage.success(t('admin.dept.setManagerSuccess', { name: member.name }))
}

function removeMember(member: Member) {
  deptMembers.value = deptMembers.value.filter(m => m.id !== member.id)
  ElMessage.success(t('admin.dept.memberRemoved'))
}

async function addMembers() {
  addingMembers.value = true
  try {
    await new Promise(r => setTimeout(r, 500))
    ElMessage.success(t('admin.dept.memberAdded'))
    showAddMemberDialog.value = false
    fetchDeptMembers()
  } finally {
    addingMembers.value = false
  }
}

// Permissions
const showPermDialog = ref(false)
const permTab = ref('doc')
const deptPermissions = ref<any[]>([])
const docPermList = ref<any[]>([])
const folderPermList = ref<any[]>([])
const kbPermList = ref<any[]>([])
const savingPerms = ref(false)

const permLevelMap: Record<string, string> = {
  view: t('admin.dept.canView'),
  edit: t('admin.dept.canEdit'),
  manage: t('admin.dept.canManage'),
}

function fetchDeptPermissions() {
  deptPermissions.value = selectedDept.value ? [
    { id: 1, type: 'doc', name: t('admin.dept.mockProductReqDoc'), level: 'edit' },
    { id: 2, type: 'folder', name: t('admin.dept.mockProjectFiles'), level: 'view' },
  ] : []
}

async function savePermissions() {
  savingPerms.value = true
  try {
    await new Promise(r => setTimeout(r, 500))
    ElMessage.success(t('admin.dept.permsSaved'))
    showPermDialog.value = false
    fetchDeptPermissions()
  } finally {
    savingPerms.value = false
  }
}

// Import
const showImportDialog = ref(false)
const importFile = ref<File | null>(null)
const importPreview = ref<any[]>([])
const importing = ref(false)

function handleFileChange(file: any) {
  importFile.value = file.raw
  // Mock preview
  importPreview.value = [
    { name: t('admin.dept.mockHeadquarters'), children: [
      { name: t('admin.dept.mockTechDept'), children: [
        { name: t('admin.dept.mockFrontendTeam') },
        { name: t('admin.dept.mockBackendTeam') },
      ]},
      { name: t('admin.dept.mockProductDept') },
    ]},
  ]
}

async function confirmImport() {
  importing.value = true
  try {
    await new Promise(r => setTimeout(r, 1000))
    ElMessage.success(t('admin.dept.orgImported'))
    showImportDialog.value = false
    fetchDepartments()
  } finally {
    importing.value = false
  }
}

async function fetchDepartments() {
  try {
    const res: any = await getAdminDepartments()
    console.log('[DEBUG] API response:', JSON.stringify(res.data, null, 2))
    // Backend already returns tree structure, flatten it for internal use
    const flattenDepartments = (items: any[]): any[] => {
      const result: any[] = []
      for (const item of items) {
        result.push({
          id: item.id,
          name: item.name,
          parentId: item.parentId || item.parent_id || null,
          memberCount: item.memberCount || item.member_count || 0,
          description: item.description || '',
        })
        if (item.children && item.children.length > 0) {
          result.push(...flattenDepartments(item.children))
        }
      }
      return result
    }
    departments.value = flattenDepartments(res.data || [])
    console.log('[DEBUG] Flattened departments:', departments.value)
    // Auto expand first root
    if (departments.value.length > 0) {
      expandedKeys.value = [departments.value[0].id]
    }
  } catch (error) {
    console.error('Failed to fetch departments:', error)
    // Fallback to mock data if API fails
    departments.value = [
      { id: 1, name: t('admin.dept.mockTechDept'), parentId: null, memberCount: 15 },
      { id: 2, name: t('admin.dept.mockProductDept'), parentId: null, memberCount: 8 },
      { id: 3, name: t('admin.dept.mockFrontendTeam'), parentId: 1, memberCount: 5 },
      { id: 4, name: t('admin.dept.mockBackendTeam'), parentId: 1, memberCount: 6 },
      { id: 5, name: t('admin.dept.mockTestTeam'), parentId: 1, memberCount: 4 },
      { id: 6, name: t('admin.dept.mockProductDesign'), parentId: 2, memberCount: 4 },
      { id: 7, name: t('admin.dept.mockOperationsDept'), parentId: null, memberCount: 10 },
    ]
    expandedKeys.value = [1]
  }
}

async function fetchAllUsers() {
  allUsers.value = [
    { id: 1, name: t('admin.dept.mockUser1') },
    { id: 2, name: t('admin.dept.mockUser2') },
    { id: 3, name: t('admin.dept.mockUser3') },
    { id: 4, name: t('admin.dept.mockUser4') },
  ]
}

onMounted(() => {
  fetchDepartments()
  fetchAllUsers()
  // Mock permission lists
  docPermList.value = [
    { id: 1, name: t('admin.dept.mockProductReqDoc'), level: 'none' },
    { id: 2, name: t('admin.dept.mockTechPlan'), level: 'none' },
  ]
  folderPermList.value = [
    { id: 1, name: t('admin.dept.mockProjectFiles'), level: 'none' },
    { id: 2, name: t('admin.dept.mockSharedDocs'), level: 'none' },
  ]
  kbPermList.value = [
    { id: 1, name: t('admin.dept.mockTechDocLib'), level: 'none' },
    { id: 2, name: t('admin.dept.mockProductKb'), level: 'none' },
  ]
})
</script>

<style scoped>
.department-management {
  padding: 24px;
  height: 100%;
  display: flex;
  flex-direction: column;
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
.content-wrapper {
  display: flex;
  gap: 20px;
  flex: 1;
  min-height: 0;
}
.tree-panel {
  width: 320px;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  display: flex;
  flex-direction: column;
}
.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--kx-border);
}
.tree-content {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}
.tree-node {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
  padding: 4px 0;
}
.node-name {
  flex: 1;
  font-size: 13px;
}
.node-count {
  font-size: 12px;
  color: var(--kx-text-placeholder);
}
.node-actions {
  display: none;
  gap: 4px;
}
.tree-node:hover .node-actions {
  display: flex;
}
.action-btn {
  font-size: 14px;
  cursor: pointer;
  color: var(--kx-text-secondary);
  padding: 2px;
  border-radius: 4px;
}
.action-btn:hover {
  background: rgba(0,0,0,0.06);
  color: var(--kx-primary);
}
.action-btn.danger:hover {
  color: #f54a45;
}
.detail-panel {
  flex: 1;
  background: #fff;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  padding: 20px;
  overflow-y: auto;
}
.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--kx-border);
}
.detail-header h3 {
  margin: 0;
  font-size: 18px;
}
.section {
  margin-bottom: 24px;
}
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-weight: 500;
}
.member-list {
  border: 1px solid var(--kx-border);
  border-radius: 8px;
}
.member-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-bottom: 1px solid var(--kx-border);
}
.member-item:last-child {
  border-bottom: none;
}
.member-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.member-name {
  font-size: 14px;
}
.member-role {
  font-size: 12px;
  color: var(--kx-text-placeholder);
}
.perm-list {
  border: 1px solid var(--kx-border);
  border-radius: 8px;
}
.perm-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  border-bottom: 1px solid var(--kx-border);
}
.perm-item:last-child {
  border-bottom: none;
}
.perm-info {
  display: flex;
  align-items: center;
  gap: 8px;
}
.empty-tip {
  text-align: center;
  padding: 20px;
  color: var(--kx-text-placeholder);
}
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--kx-text-placeholder);
}
.empty-state p {
  margin-top: 12px;
}
.import-preview {
  margin-top: 16px;
  border: 1px solid var(--kx-border);
  border-radius: 8px;
  padding: 12px;
  max-height: 200px;
  overflow-y: auto;
}
.preview-header {
  font-weight: 500;
  margin-bottom: 8px;
}
.perm-config-list {
  max-height: 300px;
  overflow-y: auto;
}
.perm-config-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid var(--kx-border);
}
.perm-config-item:last-child {
  border-bottom: none;
}
</style>
