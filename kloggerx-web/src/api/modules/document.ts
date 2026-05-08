import { get, post, del, upload } from '../request'
import type { ApiResponse, Document, DocumentType, PaginatedData } from '@/types'

export function getDocumentTree(parentId?: number | null) {
  return get<ApiResponse>('/api/v1/document/tree', { parentId })
}

export function getDocumentDetail(id: number) {
  return get<ApiResponse<Document>>(`/api/v1/document/${id}`)
}

export function createDocument(data: { title: string; type: DocumentType | 'folder'; parentId: number | null }) {
  return post<ApiResponse<Document>>('/api/v1/document/create', data)
}

export function updateDocument(id: number, data: Partial<Document>) {
  return post<ApiResponse>(`/api/v1/document/${id}/update`, data)
}

export function saveDocumentContent(id: number, content: string) {
  return post<ApiResponse>(`/api/v1/document/${id}/content`, { content })
}

export function deleteDocument(id: number) {
  return post<ApiResponse>(`/api/v1/document/${id}/delete`)
}

export function restoreDocument(id: number) {
  return post<ApiResponse>(`/api/v1/document/${id}/restore`)
}

export function permanentDeleteDocument(id: number) {
  return del<ApiResponse>(`/api/v1/document/${id}/permanent`)
}

export function batchRestoreDocuments(docIds: number[]) {
  return post<ApiResponse>('/api/v1/document/batch-restore', { doc_ids: docIds })
}

export function cleanupExpiredDocuments() {
  return post<ApiResponse<{ deleted_count: number }>>('/api/v1/admin/document/cleanup-expired')
}

export function moveDocument(id: number, targetParentId: number | null) {
  return post<ApiResponse>(`/api/v1/document/${id}/move`, { targetParentId })
}

export function copyDocument(id: number, includeChildren: boolean) {
  return post<ApiResponse>(`/api/v1/document/${id}/copy`, { includeChildren })
}

export function pinDocument(id: number, isPinned: boolean) {
  return post<ApiResponse>(`/api/v1/document/${id}/pin`, { isPinned })
}

export function favoriteDocument(id: number, isFavorite: boolean) {
  return post<ApiResponse>(`/api/v1/document/${id}/favorite`, { isFavorite })
}

export function transferOwnership(id: number, targetUser: string, keepPermission: boolean = true) {
  return post<ApiResponse>(`/api/v1/document/${id}/transfer`, { targetUser, keepPermission })
}

export function addShortcut(documentId: number, targetParentId: number | null) {
  return post<ApiResponse>('/api/v1/document/shortcut', { documentId, targetParentId })
}

export function migrateDocuments(documentIds: number[], targetParentId: number | null) {
  return post<ApiResponse>('/api/v1/document/migrate', { documentIds, targetParentId })
}

export function getRecycleBin(params: { page: number; pageSize: number }) {
  return get<ApiResponse<PaginatedData<Document>>>('/api/v1/document/recycle-bin', params)
}

export function getFavorites(params: { page: number; pageSize: number }) {
  return get<ApiResponse<PaginatedData<Document>>>('/api/v1/document/favorites', params)
}

export function getPinnedDocuments() {
  return get<ApiResponse<Document[]>>('/api/v1/document/pinned')
}

export function getRecentDocuments(params: { page: number; pageSize: number }) {
  return get<ApiResponse<PaginatedData<Document>>>('/api/v1/document/recent', params)
}

export function searchDocuments(params: { keyword: string; page: number; pageSize: number }) {
  return get<ApiResponse<PaginatedData<Document>>>('/api/v1/document/search', params)
}

export function getDocumentVersions(id: number) {
  return get<ApiResponse>(`/api/v1/document/${id}/versions`)
}

export function rollbackVersion(id: number, version: number) {
  return post<ApiResponse>(`/api/v1/document/${id}/rollback`, { version })
}

export function importDocument(file: File, parentId: number | null, onProgress?: (p: number) => void) {
  return upload<ApiResponse>(`/api/v1/document/import?parentId=${parentId ?? ''}`, file, onProgress)
}

export function exportDocument(id: number, format: string) {
  return get<ApiResponse>(`/api/v1/document/${id}/export`, { format })
}

export function getDocumentFilePreview(id: number) {
  return get<ApiResponse<{ url: string; fileType: string; fileName: string }>>(`/api/v1/document/${id}/file-preview`)
}

export function saveDocumentAsTemplate(docId: number, data: { name: string; description: string; category: string }) {
  return post<ApiResponse>(`/api/v1/document/${docId}/save-as-template`, data)
}

export interface EnhancedSearchResult {
  id: number
  title: string
  preview: string
  type: string
  updatedAt: string
  ownerId: number
}

export interface EnhancedSearchResponse {
  items: EnhancedSearchResult[]
  total: number
  page: number
  pageSize: number
}

export function enhancedSearchDocuments(data: {
  query: string
  type?: string
  page?: number
  pageSize?: number
}) {
  return post<ApiResponse<EnhancedSearchResponse>>('/api/v1/document/search', data)
}

export function getSearchSuggestions(q: string) {
  return get<ApiResponse<{ suggestions: string[] }>>('/api/v1/search/suggestions', { q })
}

export interface DiffLine {
  type: 'add' | 'delete' | 'equal'
  content: string
  old_line?: number
  new_line?: number
}

export interface VersionDiffResult {
  old_version: string
  new_version: string
  lines: DiffLine[]
  stats: { added: number; deleted: number; changed: number }
}

export function getVersionDiff(docId: number, v1: string, v2: string) {
  return get<ApiResponse<VersionDiffResult>>(`/api/v1/document/${docId}/versions/diff`, { v1, v2 })
}
