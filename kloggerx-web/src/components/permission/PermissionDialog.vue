<template>
  <el-dialog :model-value="true" title="权限管理" width="520px" @close="$emit('close')">
    <div class="perm-add">
      <el-input v-model="searchUser" placeholder="搜索用户..." size="small" style="flex:1" />
      <el-select v-model="addLevel" size="small" style="width:100px">
        <el-option label="可编辑" value="edit" />
        <el-option label="可查看" value="view" />
        <el-option label="可管理" value="manage" />
      </el-select>
      <el-button type="primary" size="small" @click="handleAdd">添加</el-button>
    </div>
    <div class="perm-list">
      <div v-for="p in permissions" :key="p.id" class="perm-item">
        <el-avatar :size="28" :src="p.userAvatar">{{ p.userName[0] }}</el-avatar>
        <span class="perm-name">{{ p.userName }}</span>
        <el-tag v-if="p.inherited" size="small" type="info">继承</el-tag>
        <el-select v-model="p.level" size="small" style="width:90px" :disabled="p.level === 'owner'" @change="handleChange(p)">
          <el-option label="所有者" value="owner" disabled />
          <el-option label="可管理" value="manage" />
          <el-option label="可编辑" value="edit" />
          <el-option label="可查看" value="view" />
        </el-select>
        <el-button v-if="p.level !== 'owner'" link type="danger" @click="handleRemove(p)"><el-icon><Delete /></el-icon></el-button>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getDocumentPermissions, setPermission, removePermission } from '@/api/modules/auth'
import type { Permission, PermissionLevel } from '@/types'
import { ElMessage } from 'element-plus'

const props = defineProps<{ documentId: number }>()
defineEmits(['close'])
const permissions = ref<Permission[]>([])
const searchUser = ref('')
const addLevel = ref<PermissionLevel>('edit')

async function fetchPerms() {
  const res: any = await getDocumentPermissions(props.documentId)
  permissions.value = res.data || []
}

async function handleAdd() {
  if (!searchUser.value.trim()) return
  ElMessage.info('请通过后端用户搜索接口选择用户后添加')
}

async function handleChange(p: Permission) {
  await setPermission({ documentId: props.documentId, userId: p.userId, level: p.level })
  ElMessage.success('权限已更新')
}

async function handleRemove(p: Permission) {
  await removePermission({ documentId: props.documentId, userId: p.userId })
  ElMessage.success('已移除')
  fetchPerms()
}

onMounted(fetchPerms)
</script>

<style scoped>
.perm-add { display: flex; gap: 8px; margin-bottom: 16px; }
.perm-list { max-height: 320px; overflow-y: auto; }
.perm-item { display: flex; align-items: center; gap: 10px; padding: 8px 0; border-bottom: 1px solid var(--kx-border); }
.perm-name { flex: 1; font-size: 13px; }
</style>
