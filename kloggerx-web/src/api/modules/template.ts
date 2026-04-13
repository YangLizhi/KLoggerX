import { get, post } from '../request'

export interface Template {
  id: number
  name: string
  description: string
  type: string
  category: string
  content: string
  preview: string
  isBuiltin: boolean
  createdAt: string
}

export interface TemplateListResponse {
  list: Template[]
  categories: string[]
}

// Get template list (public)
export function getTemplates(category?: string) {
  const params = category ? { category } : {}
  return get<TemplateListResponse>('/api/v1/template/list', params)
}

// Get template detail
export function getTemplateDetail(id: number) {
  return get<Template>(`/api/v1/template/${id}`)
}

// Use template to create a new document
export function useTemplate(id: number, data: { title?: string; parentId?: number | null }) {
  return post<{ id: number; title: string; type: string }>(`/api/v1/template/${id}/use`, data)
}
