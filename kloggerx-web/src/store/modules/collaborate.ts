import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Collaborator } from '@/types'

export const useCollaborateStore = defineStore('collaborate', () => {
  const collaborators = ref<Collaborator[]>([])
  const isConnected = ref(false)

  function setCollaborators(list: Collaborator[]) {
    collaborators.value = list
  }

  function addCollaborator(c: Collaborator) {
    const idx = collaborators.value.findIndex((x) => x.userId === c.userId)
    if (idx >= 0) {
      collaborators.value[idx] = c
    } else {
      collaborators.value.push(c)
    }
  }

  function removeCollaborator(userId: number) {
    collaborators.value = collaborators.value.filter((c) => c.userId !== userId)
  }

  function setConnected(v: boolean) {
    isConnected.value = v
  }

  return { collaborators, isConnected, setCollaborators, addCollaborator, removeCollaborator, setConnected }
})
