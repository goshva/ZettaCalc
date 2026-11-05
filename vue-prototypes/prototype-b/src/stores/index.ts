import { defineStore } from 'pinia'
import { api } from '@/api/client'
import type { CartItem, CartResponse, ContractorBid } from '@/types'

export const useCartStore = defineStore('cart_b', {
  state: () => ({ items: [] as CartItem[], total_rub: 0, loading: false }),
  getters: { count: (s) => s.items.length },
  actions: {
    async fetch() { this.loading = true; try { const { data } = await api.get<CartResponse>('/api/cart'); this.items = data.items; this.total_rub = data.total_rub } finally { this.loading = false } },
    async add(item: CartItem) { const { data } = await api.post<CartResponse>('/api/cart/items', item); this.items = data.items; this.total_rub = data.total_rub ?? 0 },
  },
})

export const useCprStore = defineStore('cpr_b', {
  state: () => ({ current: null as any, bids: [] as ContractorBid[], loading: false, error: '' }),
  actions: {
    async generate(projectId: string) {
      this.loading = true; this.error = ''
      try { const { data } = await api.post('/api/cpr/generate', { project_id: projectId }); this.current = data }
      catch (e: unknown) { this.error = (e as { message?: string }).message ?? 'Ошибка генерации' }
      finally { this.loading = false }
    },
    async fetchBids() { const { data } = await api.get<ContractorBid[]>('/api/bids'); this.bids = data },
  },
})
