import { get, post, put, del } from '../request'
import type { ApiResponse, PaginatedData } from '@/types'

// User Management
export function getAdminUsers(params: { page: number; pageSize: number; keyword?: string; role?: string }) {
  return get<ApiResponse<PaginatedData<any>>>('/api/v1/admin/users', params)
}

export function createAdminUser(data: { username: string; email: string; password: string; nickname?: string; role?: string; departmentId?: number }) {
  return post<ApiResponse<any>>('/api/v1/admin/users', data)
}

export function updateAdminUser(id: number, data: { nickname?: string; email?: string; role?: string; departmentId?: number }) {
  return put<ApiResponse>(`/api/v1/admin/users/${id}`, data)
}

export function deleteAdminUser(id: number) {
  return del<ApiResponse>(`/api/v1/admin/users/${id}`)
}

export function resetUserPassword(id: number, password: string) {
  return post<ApiResponse>(`/api/v1/admin/users/${id}/reset-password`, { password })
}

// Department Management
export function getAdminDepartments() {
  return get<ApiResponse<any[]>>('/api/v1/admin/departments')
}

export function createAdminDepartment(data: { name: string; parentId?: number }) {
  return post<ApiResponse<any>>('/api/v1/admin/departments', data)
}

export function updateAdminDepartment(id: number, data: { name: string; parentId?: number }) {
  return put<ApiResponse>(`/api/v1/admin/departments/${id}`, data)
}

export function deleteAdminDepartment(id: number) {
  return del<ApiResponse>(`/api/v1/admin/departments/${id}`)
}

export function getDepartmentMembers(id: number) {
  return get<ApiResponse<any[]>>(`/api/v1/admin/departments/${id}/members`)
}

// AI Model Settings
export function getAIModelSettings() {
  return get<ApiResponse<any>>('/api/v1/admin/settings/ai')
}

export function saveAIModelSettings(data: any) {
  return post<ApiResponse>('/api/v1/admin/settings/ai', data)
}

export function detectAIModels(data: { baseUrl: string; apiKey: string }) {
  return post<ApiResponse<{ models: any[]; count: number }>>('/api/v1/admin/settings/ai/detect', data)
}

export function testAIModel(data: { baseUrl: string; apiKey: string; model: string; prompt?: string }) {
  return post<ApiResponse<{ response: string; success: boolean }>>('/api/v1/admin/settings/ai/test', data)
}

// Storage Settings
export function getStorageSettings() {
  return get<ApiResponse<any>>('/api/v1/admin/settings/storage')
}

export function saveStorageSettings(data: any) {
  return post<ApiResponse>('/api/v1/admin/settings/storage', data)
}

export function testRemoteStorage(data: { type: string; server: string; port?: number; username?: string; password?: string }) {
  return post<ApiResponse<any>>('/api/v1/admin/settings/storage/test', data)
}

// Storage Statistics
export function getStorageStats() {
  return get<ApiResponse<{ total: number; used: number; available: number; fileCount: number }>>('/api/v1/storage/stats')
}

export function getStorageUsage() {
  return get<ApiResponse<{ byType: { doc: number; sheet: number; slide: number; image: number; other: number }; totalSize: number; fileCount: number }>>('/api/v1/storage/usage')
}

export function getAdminStorageUsage() {
  return get<ApiResponse<{ diskStats: any; byType: any; totalSize: number; fileCount: number }>>('/api/v1/admin/storage/usage')
}

// Remote Storage Management
export function listRemoteStorages() {
  return get<ApiResponse<any[]>>('/api/v1/admin/remote-storages')
}

export function createRemoteStorage(data: { name: string; type: string; server: string; port?: number; username?: string; password?: string; sharePath?: string; domain?: string; mountPoint: string; isEnabled?: boolean }) {
  return post<ApiResponse<any>>('/api/v1/admin/remote-storages', data)
}

export function updateRemoteStorage(id: number, data: any) {
  return put<ApiResponse>(`/api/v1/admin/remote-storages/${id}`, data)
}

export function deleteRemoteStorage(id: number) {
  return del<ApiResponse>(`/api/v1/admin/remote-storages/${id}`)
}

export function testRemoteStorageConnection(id: number) {
  return post<ApiResponse<{ connected: boolean; message: string }>>(`/api/v1/admin/remote-storages/${id}/test`)
}

export function connectRemoteStorage(id: number) {
  return post<ApiResponse<any>>(`/api/v1/admin/remote-storages/${id}/connect`)
}

export function disconnectRemoteStorage(id: number) {
  return post<ApiResponse<any>>(`/api/v1/admin/remote-storages/${id}/disconnect`)
}

// User Storage Settings
export function getUserStorageSettings() {
  return get<ApiResponse<{ syncDir: string; downloadDir: string; autoSync: boolean }>>('/api/v1/user/storage-settings')
}

export function saveUserStorageSettings(data: { syncDir?: string; downloadDir?: string; autoSync?: boolean }) {
  return post<ApiResponse>('/api/v1/user/storage-settings', data)
}

// AD/LDAP
export function testLdapConnection(data: { type: string; server: string; baseDN: string; bindDN: string; bindPassword: string }) {
  return post<ApiResponse<{ success: boolean }>>('/api/v1/admin/ldap/test', data)
}

export function fetchLdapUsers(data: { type: string; server: string; baseDN: string; bindDN: string; bindPassword: string }) {
  return post<ApiResponse<{ users: any[] }>>('/api/v1/admin/ldap/users', data)
}

export function importLdapUsers(data: { type: string; server: string; baseDN: string; bindDN: string; bindPassword: string; users: string[] }) {
  return post<ApiResponse>('/api/v1/admin/ldap/import', data)
}
