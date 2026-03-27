import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { User } from '@/types'
import { getUserInfo, login as loginApi } from '@/api/modules/user'
import type { LoginForm } from '@/types'
import router from '@/router'

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null)
  const token = ref(localStorage.getItem('kx_token') || '')

  async function login(form: LoginForm) {
    const res: any = await loginApi(form)
    token.value = res.data.token
    user.value = res.data.user
    localStorage.setItem('kx_token', res.data.token)
  }

  async function fetchUserInfo() {
    if (!token.value) return
    const res: any = await getUserInfo()
    user.value = res.data
  }

  function logout() {
    user.value = null
    token.value = ''
    localStorage.removeItem('kx_token')
    router.push('/login')
  }

  return { user, token, login, fetchUserInfo, logout }
})
