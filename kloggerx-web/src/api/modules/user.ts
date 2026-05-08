import { post, get } from '../request'
import type { ApiResponse, User, LoginForm, RegisterForm } from '@/types'

export function login(data: LoginForm) {
  return post<ApiResponse<{ token: string; user: User }>>('/api/v1/user/login', data)
}

export function register(data: RegisterForm) {
  return post<ApiResponse>('/api/v1/user/register', data)
}

export function getUserInfo() {
  return get<ApiResponse<User>>('/api/v1/user/info')
}

export function updateUserInfo(data: Partial<User>) {
  return post<ApiResponse>('/api/v1/user/update', data)
}

export function changePassword(data: { oldPassword: string; newPassword: string }) {
  return post<ApiResponse>('/api/v1/user/change-password', data)
}

export function getUserList(params: { page: number; pageSize: number; keyword?: string }) {
  return get<ApiResponse>('/api/v1/user/list', params)
}

export function getDepartmentTree() {
  return get<ApiResponse>('/api/v1/department/tree')
}

export function createDepartment(data: { name: string; parentId: number | null }) {
  return post<ApiResponse>('/api/v1/department/create', data)
}

export function updateDepartment(data: { id: number; name: string }) {
  return post<ApiResponse>('/api/v1/department/update', data)
}

export function deleteDepartment(id: number) {
  return post<ApiResponse>('/api/v1/department/delete', { id })
}

export function searchUsers(keyword: string) {
  return get<ApiResponse>('/api/v1/users/search', { q: keyword })
}
