import { defineStore } from 'pinia'
import { api } from '@/api/client'
import type { ReferenceItem } from '@/types'

export const useReferencesStore = defineStore('references', {
  state: () => ({
    items: [] as ReferenceItem[],
    loading: false,
    filter: '' as string,
    error: '',
  }),
  getters: {
    filtered(s) {
      if (!s.filter) return s.items
      return s.items.filter((i) => i.category === s.filter || i.styles.includes(s.filter))
    },
    byId: (s) => (id: string) => s.items.find((i) => i.id === id),
  },
  actions: {
    async fetch(category?: string) {
      this.loading = true
      this.error = ''
      try {
        const { data } = await api.get<ReferenceItem[]>('/api/references', { params: category ? { category } : {} })
        this.items = data
      } catch (e: unknown) {
        this.error = (e as { message?: string }).message ?? 'Ошибка загрузки'
      } finally {
        this.loading = false
      }
    },
    setFilter(f: string) { this.filter = f },
  },
})
