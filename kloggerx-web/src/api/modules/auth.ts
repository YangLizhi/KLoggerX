import { get, post } from '../request'
import type { ApiResponse, Permission, PermissionLevel, ShareSetting } from '@/types'

export function getDocumentPermissions(documentId: number) {
  return get<ApiResponse<Permission[]>>(`/api/v1/auth/document/${documentId}`)
}

export function setPermission(data: { documentId: number; userId: number; level: PermissionLevel }) {
  return post<ApiResponse>('/api/v1/auth/permission/set', data)
}

export function removePermission(data: { documentId: number; userId: number }) {
  return post<ApiResponse>('/api/v1/auth/permission/remove', data)
}

export function getShareSetting(documentId: number) {
  return get<ApiResponse<ShareSetting>>(`/api/v1/auth/share/${documentId}`)
}

export function updateShareSetting(data: ShareSetting) {
  return post<ApiResponse>('/api/v1/auth/share/update', data)
}

export function checkPermission(documentId: number) {
  return get<ApiResponse<{ level: PermissionLevel }>>(`/api/v1/auth/check/${documentId}`)
}
