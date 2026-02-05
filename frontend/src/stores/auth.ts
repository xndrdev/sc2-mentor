import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, AuthResponse } from '@/api/client'
import { login as apiLogin, register as apiRegister, getMe, logout as apiLogout, setAuthToken } from '@/api/client'

const TOKEN_KEY = 'sc2_auth_token'
const USER_KEY = 'sc2_auth_user'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))
  const user = ref<User | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Initialisiere User aus LocalStorage
  const savedUser = localStorage.getItem(USER_KEY)
  if (savedUser) {
    try {
      user.value = JSON.parse(savedUser)
    } catch {
      localStorage.removeItem(USER_KEY)
    }
  }

  // Setze Token im API Client wenn vorhanden
  if (token.value) {
    setAuthToken(token.value)
  }

  const isAuthenticated = computed(() => !!token.value && !!user.value)

  function saveAuth(authResponse: AuthResponse) {
    token.value = authResponse.token
    user.value = authResponse.user
    localStorage.setItem(TOKEN_KEY, authResponse.token)
    localStorage.setItem(USER_KEY, JSON.stringify(authResponse.user))
    setAuthToken(authResponse.token)
  }

  function clearAuth() {
    token.value = null
    user.value = null
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
    setAuthToken(null)
  }

  async function login(email: string, password: string) {
    loading.value = true
    error.value = null
    try {
      const response = await apiLogin(email, password)
      saveAuth(response)
      return true
    } catch (e: unknown) {
      const axiosError = e as { response?: { data?: { error?: string } } }
      error.value = axiosError.response?.data?.error || 'Login fehlgeschlagen'
      return false
    } finally {
      loading.value = false
    }
  }

  async function register(email: string, password: string, sc2PlayerName: string) {
    loading.value = true
    error.value = null
    try {
      const response = await apiRegister(email, password, sc2PlayerName)
      saveAuth(response)
      return true
    } catch (e: unknown) {
      const axiosError = e as { response?: { data?: { error?: string } } }
      error.value = axiosError.response?.data?.error || 'Registrierung fehlgeschlagen'
      return false
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    try {
      await apiLogout()
    } catch {
      // Ignoriere Fehler beim Logout
    }
    clearAuth()
  }

  async function checkAuth() {
    if (!token.value) {
      return false
    }

    try {
      const userData = await getMe()
      user.value = userData
      localStorage.setItem(USER_KEY, JSON.stringify(userData))
      return true
    } catch {
      clearAuth()
      return false
    }
  }

  return {
    token,
    user,
    loading,
    error,
    isAuthenticated,
    login,
    register,
    logout,
    checkAuth,
  }
})
