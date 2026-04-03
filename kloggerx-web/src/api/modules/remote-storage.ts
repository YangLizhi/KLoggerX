import { get, post, del } from '../request'
import type { ApiResponse } from '@/types'

// Remote file info interface
export interface RemoteFileInfo {
  name: string
  size: number
  isDir: boolean
  modTime: string
  path: string
}

// List files in remote storage directory
export function listRemoteFiles(storageId: number, path: string = '/') {
  return get<ApiResponse<{ path: string; files: RemoteFileInfo[] }>>(
    `/api/v1/remote-storage/${storageId}/files`,
    { path }
  )
}

// Download file from remote storage with auth token
export async function downloadRemoteFile(storageId: number, path: string): Promise<void> {
  const token = localStorage.getItem('kx_token')
  const url = `/api/v1/remote-storage/${storageId}/download?path=${encodeURIComponent(path)}`

  console.log('[RemoteStorage] Downloading file:', path, 'token exists:', !!token)

  try {
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    })

    if (!response.ok) {
      throw new Error(`下载失败: ${response.status} ${response.statusText}`)
    }

    // Get filename from path
    const filename = path.split('/').pop() || 'download'

    // Create blob from response
    const blob = await response.blob()

    // Create download link
    const blobUrl = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = blobUrl
    a.download = filename
    document.body.appendChild(a)
    a.click()

    // Cleanup
    document.body.removeChild(a)
    URL.revokeObjectURL(blobUrl)
  } catch (error) {
    console.error('Download failed:', error)
    throw error
  }
}

// Upload file to remote storage
export function uploadRemoteFile(
  storageId: number,
  path: string,
  file: File,
  onProgress?: (percent: number) => void
) {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('path', path)

  // Use fetch for progress tracking
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest()
    xhr.open('POST', `/api/v1/remote-storage/${storageId}/upload`)

    // Add auth header
    const token = localStorage.getItem('kx_token')
    if (token) {
      xhr.setRequestHeader('Authorization', `Bearer ${token}`)
    }

    xhr.upload.onprogress = (e) => {
      if (onProgress && e.lengthComputable) {
        onProgress(Math.round((e.loaded / e.total) * 100))
      }
    }

    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        try {
          resolve(JSON.parse(xhr.responseText))
        } catch {
          resolve(xhr.responseText)
        }
      } else {
        reject(new Error(`Upload failed: ${xhr.statusText}`))
      }
    }

    xhr.onerror = () => reject(new Error('Upload failed'))
    xhr.send(formData)
  })
}

// Create folder in remote storage
export function createRemoteFolder(storageId: number, path: string) {
  return post<ApiResponse>(`/api/v1/remote-storage/${storageId}/mkdir`, { path })
}

// Delete file or folder from remote storage
export function deleteRemoteFile(storageId: number, path: string) {
  return del<ApiResponse>(`/api/v1/remote-storage/${storageId}/file`, { path })
}

// Test remote storage connection
export function testRemoteStorageConnect(storageId: number) {
  return post<ApiResponse<{ connected: boolean; message: string }>>(
    `/api/v1/remote-storage/${storageId}/test`
  )
}

// Disconnect from remote storage
export function disconnectRemoteStorageConnect(storageId: number) {
  return post<ApiResponse>(`/api/v1/remote-storage/${storageId}/disconnect`)
}
