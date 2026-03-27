export type DocumentType = 'doc' | 'sheet' | 'slide' | 'mindnote' | 'bitable' | 'survey' | 'code' | 'image' | 'file'

export interface User {
  id: number
  username: string
  email: string
  avatar: string
  nickname: string
  role: 'admin' | 'member' | 'readonly'
  departmentId: number
  createdAt: string
  updatedAt: string
}

export interface LoginForm {
  email: string
  password: string
}

export interface RegisterForm {
  username: string
  email: string
  password: string
  confirmPassword: string
}

export interface Document {
  id: number
  title: string
  type: DocumentType | 'folder'
  parentId: number | null
  ownerId: number
  ownerName: string
  content: string
  isPinned: boolean
  isFavorite: boolean
  isShortcut: boolean
  shortcutTargetId: number | null
  isDeleted: boolean
  deletedAt: string | null
  version: number
  createdAt: string
  updatedAt: string
  children?: Document[]
}

export interface DocumentTreeNode {
  id: number
  title: string
  type: DocumentType | 'folder'
  parentId: number | null
  children: DocumentTreeNode[]
  isPinned: boolean
  isFavorite: boolean
  isExpanded: boolean
}

export type PermissionLevel = 'owner' | 'manage' | 'edit' | 'view'

export interface Permission {
  id: number
  documentId: number
  userId: number
  userName: string
  userAvatar: string
  level: PermissionLevel
  inherited: boolean
  createdAt: string
}

export interface ShareSetting {
  documentId: number
  scope: 'collaborator' | 'organization' | 'public'
  defaultPermission: PermissionLevel
  linkEnabled: boolean
  shareLink: string
  includeChildren: boolean
}

export interface KnowledgeBase {
  id: number
  name: string
  description: string
  ownerId: number
  ownerName: string
  memberCount: number
  docCount: number
  createdAt: string
  updatedAt: string
}

export interface Collaborator {
  userId: number
  userName: string
  userAvatar: string
  color: string
  cursorPosition: any
  isOnline: boolean
}

export interface DocumentVersion {
  id: number
  documentId: number
  version: number
  content: string
  editorId: number
  editorName: string
  createdAt: string
}

export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

export interface PaginatedData<T = any> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

export interface Department {
  id: number
  name: string
  parentId: number | null
  children?: Department[]
  memberCount: number
}

export interface Notification {
  id: number
  type: 'mention' | 'comment' | 'permission' | 'share' | 'system' | 'approval'
  title: string
  content: string
  fromUserId: number
  fromUserName: string
  documentId: number
  docTitle?: string
  isRead: boolean
  createdAt: string
}
