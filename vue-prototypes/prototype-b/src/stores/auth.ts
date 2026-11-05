import { defineStore } from 'pinia'
import { api } from '@/api/client'
import type { TokenResponse, UserMe } from '@/types'

export const useAuthStore = defineStore('auth_b', {
  state: () => ({
    token: localStorage.getItem('auth_token_b') as string | null,
    user: null as UserMe | null,
    loading: false,
    error: '' as string,
  }),
  getters: { isAuthenticated: (s) => !!s.token },
  actions: {
    async login(email: string, password: string) {
      this.loading = true; this.error = ''
      try {
        const form = new URLSearchParams()
        form.append('username', email)
        form.append('password', password)
        const { data } = await api.post<TokenResponse>('/api/auth/login', form, {
          headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        })
        this.token = data.access_token
        localStorage.setItem('auth_token_b', data.access_token)
        await this.fetchMe()
      } catch (e: unknown) {
        const detail = (e as { response?: { data?: { detail?: string } } }).response?.data?.detail
        this.error = detail ?? 'Ошибка входа'
        throw e
      } finally { this.loading = false }
    },
    async fetchMe() {
      try { const { data } = await api.get<UserMe>('/api/auth/me'); this.user = data }
      catch { this.logout() }
    },
    logout() { this.token = null; this.user = null; localStorage.removeItem('auth_token_b') },
  },
})
