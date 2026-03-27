import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Permission, ShareSetting, PermissionLevel } from '@/types'
import { getDocumentPermissions, getShareSetting } from '@/api/modules/auth'

export const useAuthStore = defineStore('auth', () => {
  const permissions = ref<Permission[]>([])
  const shareSetting = ref<ShareSetting | null>(null)
  const currentLevel = ref<PermissionLevel | null>(null)

  async function fetchPermissions(documentId: number) {
    const res: any = await getDocumentPermissions(documentId)
    permissions.value = res.data || []
  }

  async function fetchShareSetting(documentId: number) {
    const res: any = await getShareSetting(documentId)
    shareSetting.value = res.data
  }

  return { permissions, shareSetting, currentLevel, fetchPermissions, fetchShareSetting }
})
