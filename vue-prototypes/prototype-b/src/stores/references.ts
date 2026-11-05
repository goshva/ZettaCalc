import { defineStore } from 'pinia'
import { api } from '@/api/client'
import type { ReferenceItem } from '@/types'

export const useReferencesStore = defineStore('refs_b', {
  state: () => ({ items: [] as ReferenceItem[], loading: false, error: '' }),
  actions: {
    async fetch(category?: string) {
      this.loading = true; this.error = ''
      try { const { data } = await api.get<ReferenceItem[]>('/api/references', { params: category ? { category } : {} }); this.items = data }
      catch (e: unknown) { this.error = (e as { message?: string }).message ?? 'Ошибка загрузки' }
      finally { this.loading = false }
    },
  },
})
