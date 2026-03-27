import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Document, DocumentTreeNode } from '@/types'
import {
  getDocumentTree,
  getDocumentDetail,
  getPinnedDocuments,
} from '@/api/modules/document'

export const useDocumentStore = defineStore('document', () => {
  const treeData = ref<DocumentTreeNode[]>([])
  const currentDocument = ref<Document | null>(null)
  const pinnedDocuments = ref<Document[]>([])
  const loading = ref(false)

  async function fetchTree(parentId?: number | null) {
    loading.value = true
    try {
      const res: any = await getDocumentTree(parentId)
      treeData.value = res.data || []
    } finally {
      loading.value = false
    }
  }

  async function fetchDocument(id: number) {
    loading.value = true
    try {
      const res: any = await getDocumentDetail(id)
      currentDocument.value = res.data
    } finally {
      loading.value = false
    }
  }

  async function fetchPinned() {
    const res: any = await getPinnedDocuments()
    pinnedDocuments.value = res.data || []
  }

  function clearCurrent() {
    currentDocument.value = null
  }

  return { treeData, currentDocument, pinnedDocuments, loading, fetchTree, fetchDocument, fetchPinned, clearCurrent }
})
