import { defineStore } from 'pinia'
import { api } from '@/api/client'
import type { CartResponse, CartItem } from '@/types'

export const useCartStore = defineStore('cart', {
  state: () => ({
    items: [] as CartItem[],
    total_rub: 0,
    loading: false,
  }),
  getters: {
    count: (s) => s.items.length,
  },
  actions: {
    async fetch() {
      this.loading = true
      try {
        const { data } = await api.get<CartResponse>('/api/cart')
        this.items = data.items
        this.total_rub = data.total_rub
      } finally {
        this.loading = false
      }
    },
    async add(item: CartItem) {
      const { data } = await api.post<CartResponse>('/api/cart/items', item)
      this.items = data.items
      this.total_rub = data.total_rub
    },
  },
})
