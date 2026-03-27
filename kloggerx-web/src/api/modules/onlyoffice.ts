import { get, post } from '../request'
import type { ApiResponse } from '@/types'

// OnlyOffice editor configuration
export interface OnlyOfficeConfig {
  document: {
    key: string
    title: string
    url: string
    fileType: string
  }
  documentType: string
  editorConfig: {
    callbackUrl: string
    mode: string
    user: {
      id: string
      name: string
    }
    customization?: {
      autosave: boolean
      chat: boolean
      comments: boolean
    }
    lang: string
  }
  token: string
}

export interface OnlyOfficeServerURL {
  serverUrl: string
}

// Get OnlyOffice editor configuration for a document
export function getOnlyOfficeConfig(documentId: number, mode: 'edit' | 'view' = 'edit') {
  return post<ApiResponse<OnlyOfficeConfig>>('/api/v1/onlyoffice/config', {
    documentId,
    mode
  })
}

// Get OnlyOffice server URL
export function getOnlyOfficeServerURL() {
  return get<ApiResponse<OnlyOfficeServerURL>>('/api/v1/onlyoffice/server-url')
}
