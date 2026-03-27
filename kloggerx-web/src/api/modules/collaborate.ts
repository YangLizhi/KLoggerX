import { get, post } from '../request'
import type { ApiResponse, Collaborator } from '@/types'

export function getCollaborators(documentId: number) {
  return get<ApiResponse<Collaborator[]>>(`/api/v1/collaborate/${documentId}/users`)
}

export function addComment(data: { documentId: number; content: string; selection?: any; parentId?: number }) {
  return post<ApiResponse>('/api/v1/collaborate/comment/add', data)
}

export function getComments(documentId: number) {
  return get<ApiResponse>(`/api/v1/collaborate/comment/${documentId}`)
}

export function resolveComment(commentId: number) {
  return post<ApiResponse>(`/api/v1/collaborate/comment/${commentId}/resolve`)
}

export function deleteComment(commentId: number) {
  return post<ApiResponse>(`/api/v1/collaborate/comment/${commentId}/delete`)
}

export function getNotifications(params: { page: number; pageSize: number }) {
  return get<ApiResponse>('/api/v1/collaborate/notifications', params)
}

export function markNotificationRead(id: number) {
  return post<ApiResponse>(`/api/v1/collaborate/notification/${id}/read`)
}

export function markAllNotificationsRead() {
  return post<ApiResponse>('/api/v1/collaborate/notification/read-all')
}
