import { onUnmounted, ref, type Ref } from 'vue'
import * as Y from 'yjs'
import { WebsocketProvider } from 'y-websocket'
import type { Collaborator } from '@/types'
import type { Editor } from '@tiptap/vue-3'

export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected'
export type SyncStatus = 'synced' | 'syncing' | 'conflict' | 'offline'

// 防抖记录：记录每个用户最近一次提示的时间
const concurrentEditToastTimestamps = new Map<string, number>()
const TOAST_DEBOUNCE_MS = 10000 // 10秒防抖

export interface ConcurrentEditInfo {
  userName: string
  color: string
  nodeId: number
}

export interface YjsCollaborationOptions {
  documentId: number
  user: {
    userId: number
    userName: string
    userAvatar: string
    color: string
  }
  onCollaboratorsUpdate?: (collaborators: Collaborator[]) => void
  onConcurrentEdit?: (info: ConcurrentEditInfo) => void
  onSync?: (synced: boolean) => void  // 新增：同步完成回调
}

export interface YjsCollaborationReturn {
  ydoc: Y.Doc
  provider: WebsocketProvider
  awareness: any
  destroy: () => void
  setEditor: (editor: Editor | null) => void
  getRemoteUserCursors: () => Map<number, { userId: number; userName: string; color: string; from: number; to: number }>
  synced: Ref<boolean>
  connectionStatus: Ref<ConnectionStatus>
  syncStatus: Ref<SyncStatus>
  lastSyncTime: Ref<Date | null>
  conflictDetected: Ref<boolean>
  dismissConflict: () => void
  retryConnection: () => void
}

/**
 * Yjs 协作组合式函数
 * 用于在 Tiptap 编辑器中实现实时协作功能
 */
export function useYjsCollaboration(options: YjsCollaborationOptions): YjsCollaborationReturn {
  const { documentId, user, onCollaboratorsUpdate, onConcurrentEdit, onSync } = options

  // 同步状态
  const synced = ref(false)
  const connectionStatus = ref<ConnectionStatus>('connecting')
  const syncStatus = ref<SyncStatus>('syncing')
  const lastSyncTime = ref<Date | null>(null)
  const conflictDetected = ref(false)

  // 创建 Yjs Doc
  const ydoc = new Y.Doc()

  // 获取 WebSocket 基础 URL
  const wsBaseUrl = import.meta.env.VITE_WS_URL || `ws://${window.location.host}/ws`
  
  // 获取 JWT token
  const token = localStorage.getItem('kx_token') || ''

  // 创建 WebSocket Provider
  // 使用 y-websocket 连接到后端的 Yjs 通道
  // 通过 Sec-WebSocket-Protocol header 传递 token
  const provider = new WebsocketProvider(
    wsBaseUrl,
    `yjs/${documentId}`,
    ydoc,
    {
      protocols: ['access_token', token],
      resyncInterval: 10000,
      maxBackoffTime: 30000,
      connect: true,
    }
  )

  // 获取 awareness 实例
  const awareness = provider.awareness

  // 当前编辑器实例
  let currentEditor: Editor | null = null

  // 设置本地用户状态
  const setLocalUserState = () => {
    awareness.setLocalStateField('user', {
      userId: user.userId,
      userName: user.userName,
      userAvatar: user.userAvatar,
      color: user.color,
    })
  }

  // 监听连接状态
  provider.on('status', (event: { status: string }) => {
    console.log('[Yjs] WebSocket status:', event.status)
    if (event.status === 'connected') {
      connectionStatus.value = 'connected'
      setLocalUserState()
    } else if (event.status === 'connecting') {
      connectionStatus.value = 'connecting'
    } else {
      connectionStatus.value = 'disconnected'
    }
  })

  // 监听同步状态
  provider.on('sync', (isSynced: boolean) => {
    console.log('[Yjs] Sync status:', isSynced)
    synced.value = isSynced
    if (isSynced) {
      syncStatus.value = 'synced'
      lastSyncTime.value = new Date()
    } else {
      syncStatus.value = 'syncing'
    }
    if (onSync) {
      onSync(isSynced)
    }
  })

  // 监听连接断开时更新 syncStatus
  provider.on('connection-close', () => {
    syncStatus.value = 'offline'
  })

  // 监听文档更新，检测潜在冲突（版本分叉）
  ydoc.on('update', (_update: Uint8Array, origin: any) => {
    // 如果更新来自远程且本地有未同步的更改，可能存在冲突
    if (origin !== null && origin !== ydoc.clientID && !synced.value) {
      // Yjs CRDT 自动合并，标记为冲突已检测并自动解决
      conflictDetected.value = true
      syncStatus.value = 'conflict'
      // 5秒后自动恢复
      setTimeout(() => {
        if (syncStatus.value === 'conflict') {
          syncStatus.value = 'synced'
        }
      }, 5000)
    }
  })

  // 获取远程用户的光标位置信息
  const getRemoteUserCursors = () => {
    const cursors = new Map<number, { userId: number; userName: string; color: string; from: number; to: number }>()
    awareness.getStates().forEach((state: any, clientId: number) => {
      if (state.user && state.user.userId !== user.userId && state.selection) {
        cursors.set(clientId, {
          userId: state.user.userId,
          userName: state.user.userName,
          color: state.user.color,
          from: state.selection.from,
          to: state.selection.to,
        })
      }
    })
    return cursors
  }

  // 检测并发编辑：检查远程用户的光标是否与本地用户在同一或相邻段落
  const checkConcurrentEdit = () => {
    if (!currentEditor || !onConcurrentEdit) return

    const localSelection = currentEditor.state.selection
    if (!localSelection || localSelection.empty === undefined) return

    // 获取本地光标所在段落的 node
    const localPos = localSelection.from
    const localNode = currentEditor.state.doc.resolve(localPos).node()
    if (!localNode) return

    // 获取远程用户的光标位置
    const remoteCursors = getRemoteUserCursors()
    
    remoteCursors.forEach((cursorInfo, _clientId) => {
      const remotePos = cursorInfo.from
      const remoteNode = currentEditor!.state.doc.resolve(remotePos).node()
      
      if (!remoteNode) return

      // 检查是否在同一段落或相邻段落
      // 通过比较两个位置的节点深度和父节点来判断
      // const localResolved = currentEditor.state.doc.resolve(localPos)
      // const remoteResolved = currentEditor.state.doc.resolve(remotePos)
      
      // 获取段落的起始位置（块级节点）
      const getParagraphStart = (pos: number) => {
        const resolved = currentEditor!.state.doc.resolve(pos)
        // 向上查找到块级节点（paragraph, heading 等）
        for (let d = resolved.depth; d > 0; d--) {
          const node = resolved.node(d)
          if (node.type.name === 'paragraph' || node.type.name === 'heading' || 
              node.type.name === 'blockquote' || node.type.name === 'codeBlock' ||
              node.type.name === 'listItem') {
            return resolved.before(d)
          }
        }
        return pos
      }

      const localParagraphStart = getParagraphStart(localPos)
      const remoteParagraphStart = getParagraphStart(remotePos)
      
      // 计算段落距离
      // const paragraphDistance = Math.abs(localParagraphStart - remoteParagraphStart)
      
      // 如果段落距离小于等于一个段落大小（约 500 字符），认为在同一或相邻段落
      // 或者直接比较段落起始位置
      const isConcurrent = Math.abs(localParagraphStart - remoteParagraphStart) <= 500 ||
                           localParagraphStart === remoteParagraphStart

      if (isConcurrent) {
        // 检查防抖
        const key = `${cursorInfo.userId}-${remoteParagraphStart}`
        const now = Date.now()
        const lastToastTime = concurrentEditToastTimestamps.get(key) || 0
        
        if (now - lastToastTime > TOAST_DEBOUNCE_MS) {
          concurrentEditToastTimestamps.set(key, now)
          onConcurrentEdit({
            userName: cursorInfo.userName,
            color: cursorInfo.color,
            nodeId: remoteParagraphStart,
          })
        }
      }
    })
  }

  // 监听 awareness 变化（远程用户状态更新）
  const handleAwarenessChange = () => {
    const states = Array.from(awareness.getStates().values())
    const remoteCollaborators: Collaborator[] = states
      .map((state: any) => {
        if (state.user && state.user.userId !== user.userId) {
          return {
            userId: state.user.userId,
            userName: state.user.userName,
            userAvatar: state.user.userAvatar,
            color: state.user.color,
            cursorPosition: state.cursor || null,
            isOnline: true,
          } as Collaborator
        }
        return null
      })
      .filter((c): c is Collaborator => c !== null)

    if (onCollaboratorsUpdate) {
      onCollaboratorsUpdate(remoteCollaborators)
    }

    // 检测并发编辑
    checkConcurrentEdit()
  }

  awareness.on('change', handleAwarenessChange)

  // 初始化时设置本地用户状态
  // 如果已经连接，立即设置
  if (provider.wsconnected) {
    setLocalUserState()
  }

  // 设置编辑器实例（用于检测并发编辑）
  const setEditor = (editor: Editor | null) => {
    currentEditor = editor
  }

  // 消除冲突提示
  const dismissConflict = () => {
    conflictDetected.value = false
    if (syncStatus.value === 'conflict') {
      syncStatus.value = 'synced'
    }
  }

  // 手动重连
  const retryConnection = () => {
    connectionStatus.value = 'connecting'
    provider.connect()
  }

  // 清理函数
  const destroy = () => {
    awareness.off('change', handleAwarenessChange)
    provider.destroy()
    ydoc.destroy()
    currentEditor = null
    // 清理防抖记录
    concurrentEditToastTimestamps.clear()
  }

  // 组件卸载时自动清理
  onUnmounted(() => {
    destroy()
  })

  return {
    ydoc,
    provider,
    awareness,
    destroy,
    setEditor,
    getRemoteUserCursors,
    synced,
    connectionStatus,
    syncStatus,
    lastSyncTime,
    conflictDetected,
    dismissConflict,
    retryConnection,
  }
}

export default useYjsCollaboration
