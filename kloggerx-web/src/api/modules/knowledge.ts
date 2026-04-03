import { get, post, del } from '../request'
import type { ApiResponse, KnowledgeBase, PaginatedData } from '@/types'

export function getKnowledgeBaseList(params: { page: number; pageSize: number; keyword?: string }) {
  return get<ApiResponse<PaginatedData<KnowledgeBase>>>('/api/v1/knowledge/list', params)
}

export function getKnowledgeBaseDetail(id: number) {
  return get<ApiResponse<KnowledgeBase>>(`/api/v1/knowledge/${id}`)
}

export function createKnowledgeBase(data: { name: string; description: string }) {
  return post<ApiResponse<KnowledgeBase>>('/api/v1/knowledge/create', data)
}

export function updateKnowledgeBase(id: number, data: Partial<KnowledgeBase>) {
  return post<ApiResponse>(`/api/v1/knowledge/${id}/update`, data)
}

export function deleteKnowledgeBase(id: number) {
  return del<ApiResponse>(`/api/v1/knowledge/${id}`)
}

export function getKnowledgeBaseTree(id: number) {
  return get<ApiResponse>(`/api/v1/knowledge/${id}/tree`)
}

export function addKnowledgeMember(id: number, data: { userId: number; role: string }) {
  return post<ApiResponse>(`/api/v1/knowledge/${id}/member/add`, data)
}

export function removeKnowledgeMember(id: number, userId: number) {
  return post<ApiResponse>(`/api/v1/knowledge/${id}/member/remove`, { userId })
}

export function getKnowledgeMembers(id: number) {
  return get<ApiResponse>(`/api/v1/knowledge/${id}/members`)
}

export function searchKnowledge(params: { keyword: string; knowledgeBaseId?: number; page: number; pageSize: number }) {
  return get<ApiResponse>('/api/v1/knowledge/search', params)
}

export function publishDocument(knowledgeBaseId: number, documentId: number) {
  return post<ApiResponse>(`/api/v1/knowledge/${knowledgeBaseId}/publish`, { documentId })
}

export function getKnowledgeSources(kbId: number) {
  return get<ApiResponse>(`/api/v1/knowledge/${kbId}/sources`)
}

export function addKnowledgeSource(kbId: number, data: { sourceType: string; sourceId: number }) {
  return post<ApiResponse>(`/api/v1/knowledge/${kbId}/source/add`, data)
}

export function removeKnowledgeSource(kbId: number, srcId: number) {
  return del<ApiResponse>(`/api/v1/knowledge/${kbId}/source/${srcId}`)
}

export function syncKnowledgeSource(kbId: number, srcId: number) {
  return post<ApiResponse>(`/api/v1/knowledge/${kbId}/source/${srcId}/sync`, {})
}

export function chatWithKnowledge(kbId: number, data: { question: string; history?: { role: string; content: string }[]; model?: string }) {
  return post<ApiResponse>(`/api/v1/knowledge/${kbId}/chat`, data)
}

export function chatWithKnowledgeGlobal(data: { question: string; history?: { role: string; content: string }[]; model?: string; kbIds?: number[] }) {
  return post<ApiResponse>('/api/v1/knowledge/chat/global', data)
}

// RAPTOR API
export interface RaptorTreeStats {
  totalNodes: number
  leafNodes: number
  clusterNodes: number
  rootNodes: number
  maxLevel: number
}

export function buildRaptorTree(kbId: number, config?: { clusterCount?: number; maxLevel?: number }) {
  return post<ApiResponse>(`/api/v1/knowledge/${kbId}/raptor/build`, config || {})
}

export function getRaptorTreeStats(kbId: number) {
  return get<ApiResponse<RaptorTreeStats>>(`/api/v1/knowledge/${kbId}/raptor/stats`)
}

// Embedding API
export interface EmbeddingStatus {
  totalChunks: number
  embeddedChunks: number
  pendingChunks: number
  failedChunks: number
  progress: number
}

export function getEmbeddingStatus(kbId: number) {
  return get<ApiResponse<EmbeddingStatus>>(`/api/v1/knowledge/${kbId}/embedding/status`)
}

export function rebuildEmbeddings(kbId: number) {
  return post<ApiResponse>(`/api/v1/knowledge/${kbId}/embedding/rebuild`, {})
}
