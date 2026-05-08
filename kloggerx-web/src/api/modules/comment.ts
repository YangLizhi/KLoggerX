import request from '../request'

export interface Comment {
  id: number
  document_id: number
  user_id: number
  user: { id: number; username: string; avatar: string }
  content: string
  parent_id: number | null
  quoted_text: string
  resolved: boolean
  created_at: string
  updated_at: string
  replies: Comment[]
}

export const createComment = (docId: number, data: { content: string; parent_id?: number; quoted_text?: string }) =>
  request.post(`/document/${docId}/comments`, data)

export const listComments = (docId: number) =>
  request.get(`/document/${docId}/comments`)

export const deleteComment = (commentId: number) =>
  request.delete(`/comment/${commentId}`)

export const resolveComment = (commentId: number) =>
  request.put(`/comment/${commentId}/resolve`)
