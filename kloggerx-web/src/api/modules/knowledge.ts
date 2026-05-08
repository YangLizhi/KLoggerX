import { get, post, put, del } from '../request'
import service from '../request'
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

export function addKnowledgeMember(id: number, data: { userId?: number; keyword?: string; role: string }) {
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

// 知识图谱
export function buildKnowledgeGraph(kbId: number) {
  return post<ApiResponse>(`/api/v1/knowledge/${kbId}/graph/build`)
}

export function getKnowledgeGraph(kbId: number) {
  return get<ApiResponse>(`/api/v1/knowledge/${kbId}/graph`)
}

export function getKnowledgeGraphStatus(kbId: number) {
  return get<ApiResponse>(`/api/v1/knowledge/${kbId}/graph/status`)
}

// 对话历史相关 API
export interface Conversation {
  id: number
  title: string
  knowledgeBaseId: number | null
  model: string
  isPinned: boolean
  messageCount: number
  createdAt: string
  updatedAt: string
}

export interface ConversationMessage {
  id: number
  conversationId: number
  role: 'user' | 'assistant'
  content: string
  model?: string
  sources?: any[]
  createdAt: string
}

export interface ConversationDetail extends Conversation {
  messages: ConversationMessage[]
  msgTotal: number
}

export function getConversations(params: { page?: number; pageSize?: number; knowledgeBaseId?: number }) {
  return get<ApiResponse<PaginatedData<Conversation>>>('/api/v1/knowledge/conversations', params)
}

export function createConversation(data: { knowledgeBaseId?: number; model?: string }) {
  return post<ApiResponse<Conversation>>('/api/v1/knowledge/conversations', data)
}

export function getConversationDetail(id: number, params?: { msgPage?: number; msgPageSize?: number }) {
  return get<ApiResponse<ConversationDetail>>(`/api/v1/knowledge/conversations/${id}`, params)
}

export function updateConversation(id: number, data: { title?: string; isPinned?: boolean }) {
  return put<ApiResponse>(`/api/v1/knowledge/conversations/${id}`, data)
}

export function deleteConversation(id: number) {
  return del<ApiResponse>(`/api/v1/knowledge/conversations/${id}`)
}

export function batchDeleteConversations(ids: number[]) {
  return service.delete<any, ApiResponse>('/api/v1/knowledge/conversations', { data: { ids } })
}

// 流式聊天（返回 fetch Response 用于 ReadableStream）
export function streamChat(data: {
  conversationId?: number
  knowledgeBaseId?: number
  question: string
  history?: Array<{ role: string; content: string }>
  model?: string
}, signal?: AbortSignal): Promise<Response> {
  const token = localStorage.getItem('kx_token')
  const baseURL = import.meta.env.VITE_API_BASE_URL || ''
  return fetch(`${baseURL}/api/v1/knowledge/chat/stream`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify(data),
    signal
  })
}

// 提交反馈
export function submitFeedback(messageId: number, data: {
  rating: number
  feedbackType?: string
  comment?: string
  correctAnswer?: string
}) {
  return post<ApiResponse>(`/api/v1/knowledge/messages/${messageId}/feedback`, data)
}

// 获取反馈
export function getFeedback(messageId: number) {
  return get<ApiResponse>(`/api/v1/knowledge/messages/${messageId}/feedback`)
}

// 导出对话
export function exportConversation(id: number, format: 'markdown' | 'pdf' = 'markdown') {
  return service.get('/api/v1/knowledge/conversations/' + id + '/export', {
    params: { format },
    responseType: 'blob'
  })
}

// 创建分享链接
export function createShareLink(id: number, expiresInHours?: number) {
  return post<ApiResponse<{ shareToken: string; expiresAt?: string }>>(`/api/v1/knowledge/conversations/${id}/share`, { expiresInHours })
}

// 取消分享
export function revokeShareLink(id: number) {
  return del<ApiResponse>(`/api/v1/knowledge/conversations/${id}/share`)
}

// 公开访问分享对话（无需token认证）
export function getSharedConversation(token: string) {
  return get<ApiResponse<{ conversation: Conversation; messages: ConversationMessage[] }>>(`/api/v1/knowledge/share/${token}`)
}

// 转移知识库所有权
export function transferKnowledgeBase(kbId: number, targetUser: string) {
  return post<ApiResponse>(`/api/v1/knowledge/${kbId}/transfer`, { targetUser })
}
