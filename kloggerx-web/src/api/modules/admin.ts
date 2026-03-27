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
