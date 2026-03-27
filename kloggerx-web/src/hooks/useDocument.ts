import { ref } from 'vue'
import {
  getDocumentDetail,
  saveDocumentContent,
  updateDocument,
  pinDocument,
  favoriteDocument,
} from '@/api/modules/document'

export function useDocument(docId: () => number) {
  const title = ref('')
  const content = ref('')
  const docType = ref('doc')
  const loading = ref(false)
  const saveStatus = ref('已保存')
  const isPinned = ref(false)
  const isFavorite = ref(false)
  let saveTimer: ReturnType<typeof setTimeout> | null = null

  async function load() {
    loading.value = true
    try {
      const res: any = await getDocumentDetail(docId())
      const doc = res.data
      title.value = doc.title
      content.value = doc.content || ''
      docType.value = doc.type || 'doc'
      isPinned.value = doc.isPinned
      isFavorite.value = doc.isFavorite
      return doc
    } finally {
      loading.value = false
    }
  }

  async function saveContent(jsonContent: string) {
    saveStatus.value = '保存中...'
    try {
      await saveDocumentContent(docId(), jsonContent)
      saveStatus.value = '已保存'
    } catch {
      saveStatus.value = '保存失败'
    }
  }

  function scheduleSave(getContent: () => string, delay = 2000) {
    saveStatus.value = '编辑中...'
    if (saveTimer) clearTimeout(saveTimer)
    saveTimer = setTimeout(() => saveContent(getContent()), delay)
  }

  async function saveTitle() {
    if (!title.value.trim()) title.value = '无标题文档'
    await updateDocument(docId(), { title: title.value })
  }

  async function togglePin() {
    isPinned.value = !isPinned.value
    await pinDocument(docId(), isPinned.value)
  }

  async function toggleFavorite() {
    isFavorite.value = !isFavorite.value
    await favoriteDocument(docId(), isFavorite.value)
  }

  function cleanup() {
    if (saveTimer) clearTimeout(saveTimer)
  }

  return {
    title,
    content,
    docType,
    loading,
    saveStatus,
    isPinned,
    isFavorite,
    load,
    saveContent,
    scheduleSave,
    saveTitle,
    togglePin,
    toggleFavorite,
    cleanup,
  }
}
