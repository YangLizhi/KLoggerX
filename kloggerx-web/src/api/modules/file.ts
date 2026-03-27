import { get, upload, post } from '../request'
import type { ApiResponse } from '@/types'

export function uploadFile(file: File, onProgress?: (p: number) => void) {
  return upload<ApiResponse<{ url: string; name: string; size: number }>>('/api/v1/file/upload', file, onProgress)
}

export function getFileInfo(fileId: string) {
  return get<ApiResponse>(`/api/v1/file/${fileId}`)
}

export function getFilePreviewUrl(fileId: string) {
  return get<ApiResponse<{ url: string }>>(`/api/v1/file/${fileId}/preview`)
}

export function deleteFile(fileId: string) {
  return post<ApiResponse>(`/api/v1/file/${fileId}/delete`)
}
