import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as authApi from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<{ id: number; email: string; nickname: string; role: string; status: string } | null>(null)
  const accessToken = ref(localStorage.getItem('accessToken') || '')
  const refreshToken = ref(localStorage.getItem('refreshToken') || '')

  const isLoggedIn = computed(() => !!accessToken.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isPending = computed(() => user.value?.status === 'pending')

  async function login(email: string, password: string) {
    const res = await authApi.login(email, password)
    accessToken.value = res.data.access_token
    refreshToken.value = res.data.refresh_token
    localStorage.setItem('accessToken', res.data.access_token)
    localStorage.setItem('refreshToken', res.data.refresh_token)
    user.value = res.data.user || { id: 0, email, nickname: '', role: 'user', status: 'approved' }
  }

  async function register(email: string, password: string, nickname: string) {
    return await authApi.register(email, password, nickname)
  }

  function logout() {
    accessToken.value = ''
    refreshToken.value = ''
    user.value = null
    localStorage.removeItem('accessToken')
    localStorage.removeItem('refreshToken')
  }

  // Try restore session from API on app start
  async function tryRestoreSession() {
    if (!accessToken.value) return
    try {
      const res = await authApi.getMe()
      user.value = res.data
    } catch {
      logout()
    }
  }

  return { user, accessToken, refreshToken, isLoggedIn, isAdmin, isPending, login, register, logout, tryRestoreSession }
})
