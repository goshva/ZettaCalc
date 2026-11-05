import { defineStore } from 'pinia'
import { api } from '@/api/client'
import type { CPRResponse, ContractorBid } from '@/types'

export const useCprStore = defineStore('cpr', {
  state: () => ({
    current: null as CPRResponse | null,
    bids: [] as ContractorBid[],
    loading: false,
    error: '',
  }),
  actions: {
    async generate(projectId: string) {
      this.loading = true
      this.error = ''
      try {
        const { data } = await api.post<CPRResponse>('/api/cpr/generate', { project_id: projectId })
        this.current = data
      } catch (e: unknown) {
        this.error = (e as { message?: string }).message ?? 'Ошибка генерации ЦПР'
      } finally {
        this.loading = false
      }
    },
    async fetchBids() {
      const { data } = await api.get<ContractorBid[]>('/api/bids')
      this.bids = data
    },
  },
})
