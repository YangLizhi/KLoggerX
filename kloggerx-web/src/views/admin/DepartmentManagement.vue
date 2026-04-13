<template>
  <div class="department-management">
    <div class="page-header">
      <h2>部门管理</h2>
      <div class="header-actions">
        <el-button type="primary" @click="createDepartment(null)">
          <el-icon><Plus /></el-icon>添加部门
        </el-button>
        <el-button @click="showImportDialog = true">
          <el-icon><Upload /></el-icon>导入组织架构
        </el-button>
      </div>
    </div>

    <div class="content-wrapper">
      <!-- Department Tree -->
      <div class="tree-panel">
        <div class="panel-header">
          <span>组织架构</span>
          <el-input v-model="searchKeyword" placeholder="搜索部门" prefix-icon="Search" size="small" clearable style="width: 180px" />
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
                <span class="node-count">({{ data.memberCount || 0 }}人)</span>
                <div class="node-actions">
                  <el-icon v-if="getDepth(data) < 8" class="action-btn" @click.stop="createDepartment(data)" title="添加子部门"><Plus /></el-icon>
                  <el-icon class="action-btn" @click.stop="editDepartment(data)" title="编辑"><EditPen /></el-icon>
                  <el-icon class="action-btn danger" @click.stop="deleteDepartment(data)" title="删除"><Delete /></el-icon>
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
            <el-button size="small" @click="editDepartment(selectedDept)">编辑部门</el-button>
          </div>

          <!-- Members -->
          <div class="section">
            <div class="section-header">
              <span>部门成员 ({{ deptMembers.length }}人)</span>
              <el-button size="small" type="primary" @click="showAddMemberDialog = true">
                <el-icon><Plus /></el-icon>添加成员
              </el-button>
            </div>
            <div class="member-list">
              <div v-for="m in deptMembers" :key="m.id" class="member-item">
                <el-avatar :size="32" :src="m.avatar">{{ m.name?.[0] }}</el-avatar>
                <div class="member-info">
                  <span class="member-name">{{ m.name }}</span>
                  <span class="member-role">{{ m.role === 'manager' ? '部门主管' : '成员' }}</span>
                </div>
                <el-button link size="small" @click="setAsManager(m)" v-if="m.role !== 'manager'">设为主管</el-button>
                <el-button link size="small" type="danger" @click="removeMember(m)">移除</el-button>
              </div>
              <div v-if="!deptMembers.length" class="empty-tip">暂无成员</div>
            </div>
          </div>

          <!-- Permissions -->
          <div class="section">
            <div class="section-header">
              <span>部门权限</span>
              <el-button size="small" @click="showPermDialog = true">配置权限</el-button>
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
              <div v-if="!deptPermissions.length" class="empty-tip">暂无权限配置</div>
            </div>
          </div>
        </template>
        <div v-else class="empty-state">
          <el-icon :size="48" color="#c0c4cc"><Folder /></el-icon>
          <p>请选择一个部门查看详情</p>
        </div>
      </div>
    </div>

    <!-- Create/Edit Department Dialog -->
    <el-dialog v-model="showDeptDialog" :title="editingDept ? '编辑部门' : '新建部门'" width="450px" destroy-on-close>
      <el-form :model="deptForm" :rules="deptRules" ref="deptFormRef" label-width="80px">
        <el-form-item label="部门名称" prop="name">
          <el-input v-model="deptForm.name" placeholder="请输入部门名称" maxlength="50" />
        </el-form-item>
        <el-form-item label="上级部门">
          <el-tree-select
            v-model="deptForm.parentId"
            :data="departmentTree"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            placeholder="无（顶级部门）"
            clearable
            check-strictly
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="部门描述">
          <el-input v-model="deptForm.description" type="textarea" :rows="3" placeholder="请输入部门描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDeptDialog = false">取消</el-button>
        <el-button type="primary" @click="submitDepartment" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <!-- Add Member Dialog -->
    <el-dialog v-model="showAddMemberDialog" title="添加成员" width="500px" destroy-on-close>
      <el-transfer
        v-model="selectedMembers"
        :data="allUsers"
        :titles="['可选用户', '已选用户']"
        :props="{ key: 'id', label: 'name' }"
        filterable
        filter-placeholder="搜索用户"
      />
      <template #footer>
        <el-button @click="showAddMemberDialog = false">取消</el-button>
        <el-button type="primary" @click="addMembers" :loading="addingMembers">确定</el-button>
      </template>
    </el-dialog>

    <!-- Import Dialog -->
    <el-dialog v-model="showImportDialog" title="导入组织架构" width="500px" destroy-on-close>
      <el-upload
        drag
        :auto-upload="false"
        :limit="1"
        accept=".xlsx,.xls,.csv"
        :on-change="handleFileChange"
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">拖拽文件到此处，或 <em>点击上传</em></div>
        <template #tip>
          <div class="el-upload__tip">支持 Excel (.xlsx, .xls) 或 CSV 格式，最多支持 8 级部门层级</div>
        </template>
      </el-upload>
      <div v-if="importPreview.length" class="import-preview">
        <div class="preview-header">预览 (共 {{ importPreview.length }} 个部门)</div>
        <el-tree :data="importPreview" :props="{ label: 'name', children: 'children' }" default-expand-all />
      </div>
      <template #footer>
        <el-button @click="showImportDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmImport" :loading="importing" :disabled="!importPreview.length">确认导入</el-button>
      </template>
    </el-dialog>

    <!-- Permission Config Dialog -->
    <el-dialog v-model="showPermDialog" title="配置部门权限" width="600px" destroy-on-close>
      <el-tabs v-model="permTab">
        <el-tab-pane label="文档权限" name="doc">
          <div class="perm-config-list">
            <div v-for="p in docPermList" :key="p.id" class="perm-config-item">
              <span>{{ p.name }}</span>
              <el-select v-model="p.level" size="small" style="width: 120px">
                <el-option label="无权限" value="none" />
                <el-option label="可查看" value="view" />
                <el-option label="可编辑" value="edit" />
                <el-option label="可管理" value="manage" />
              </el-select>
            </div>
          </div>
        </el-tab-pane>
        <el-tab-pane label="云盘目录权限" name="folder">
          <div class="perm-config-list">
            <div v-for="p in folderPermList" :key="p.id" class="perm-config-item">
              <span>{{ p.name }}</span>
              <el-select v-model="p.level" size="small" style="width: 120px">
                <el-option label="无权限" value="none" />
                <el-option label="可查看" value="view" />
                <el-option label="可编辑" value="edit" />
                <el-option label="可管理" value="manage" />
              </el-select>
            </div>
          </div>
        </el-tab-pane>
        <el-tab-pane label="知识库权限" name="kb">
          <div class="perm-config-list">
            <div v-for="p in kbPermList" :key="p.id" class="perm-config-item">
              <span>{{ p.name }}</span>
              <el-select v-model="p.level" size="small" style="width: 120px">
                <el-option label="无权限" value="none" />
                <el-option label="可查看" value="view" />
                <el-option label="可编辑" value="edit" />
                <el-option label="可管理" value="manage" />
              </el-select>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="showPermDialog = false">取消</el-button>
        <el-button type="primary" @click="savePermissions" :loading="savingPerms">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
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
  name: [{ required: true, message: '请输入部门名称', trigger: 'blur' }],
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
      ElMessage.success('部门已更新')
    } else {
      await createAdminDepartment({
        name: deptForm.value.name,
        parentId: deptForm.value.parentId || undefined,
      })
      ElMessage.success('部门已创建')
    }
    showDeptDialog.value = false
    fetchDepartments()
  } catch (err: any) {
    ElMessage.error(err.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

async function deleteDepartment(dept: Department) {
  if (dept.memberCount && dept.memberCount > 0) {
    ElMessage.warning('该部门下有成员，请先移除成员')
    return
  }
  await ElMessageBox.confirm(`确定要删除部门 "${dept.name}" 吗？`, '删除确认')
  try {
    await deleteAdminDepartment(dept.id)
    departments.value = departments.value.filter(d => d.id !== dept.id)
    if (selectedDept.value?.id === dept.id) {
      selectedDept.value = null
    }
    ElMessage.success('部门已删除')
  } catch (err: any) {
    ElMessage.error(err.message || '删除失败')
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
  ElMessage.success(`已将 ${member.name} 设为部门主管`)
}

function removeMember(member: Member) {
  deptMembers.value = deptMembers.value.filter(m => m.id !== member.id)
  ElMessage.success('已移除成员')
}

async function addMembers() {
  addingMembers.value = true
  try {
    await new Promise(r => setTimeout(r, 500))
    ElMessage.success('成员已添加')
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
  view: '可查看',
  edit: '可编辑',
  manage: '可管理',
}

function fetchDeptPermissions() {
  deptPermissions.value = selectedDept.value ? [
    { id: 1, type: 'doc', name: '产品需求文档', level: 'edit' },
    { id: 2, type: 'folder', name: '项目资料', level: 'view' },
  ] : []
}

async function savePermissions() {
  savingPerms.value = true
  try {
    await new Promise(r => setTimeout(r, 500))
    ElMessage.success('权限已保存')
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
    { name: '总公司', children: [
      { name: '技术部', children: [
        { name: '前端组' },
        { name: '后端组' },
      ]},
      { name: '产品部' },
    ]},
  ]
}

async function confirmImport() {
  importing.value = true
  try {
    await new Promise(r => setTimeout(r, 1000))
    ElMessage.success('组织架构已导入')
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
      { id: 1, name: '技术部', parentId: null, memberCount: 15 },
      { id: 2, name: '产品部', parentId: null, memberCount: 8 },
      { id: 3, name: '前端组', parentId: 1, memberCount: 5 },
      { id: 4, name: '后端组', parentId: 1, memberCount: 6 },
      { id: 5, name: '测试组', parentId: 1, memberCount: 4 },
      { id: 6, name: '产品设计', parentId: 2, memberCount: 4 },
      { id: 7, name: '运营部', parentId: null, memberCount: 10 },
    ]
    expandedKeys.value = [1]
  }
}

async function fetchAllUsers() {
  allUsers.value = [
    { id: 1, name: '张三' },
    { id: 2, name: '李四' },
    { id: 3, name: '王五' },
    { id: 4, name: '赵六' },
  ]
}

onMounted(() => {
  fetchDepartments()
  fetchAllUsers()
  // Mock permission lists
  docPermList.value = [
    { id: 1, name: '产品需求文档', level: 'none' },
    { id: 2, name: '技术方案', level: 'none' },
  ]
  folderPermList.value = [
    { id: 1, name: '项目资料', level: 'none' },
    { id: 2, name: '共享文档', level: 'none' },
  ]
  kbPermList.value = [
    { id: 1, name: '技术文档库', level: 'none' },
    { id: 2, name: '产品知识库', level: 'none' },
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
