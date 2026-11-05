import { defineStore } from 'pinia'
import { api } from '@/api/client'
import type { EstimateResponse } from '@/types'

export const useEstimateStore = defineStore('est_b', {
  state: () => ({ current: null as EstimateResponse | null, loading: false, error: '' }),
  getters: { total: (s) => s.current?.total_rub ?? 0 },
  actions: {
    async calculate() {
      this.loading = true; this.error = ''
      try { const { data } = await api.post<EstimateResponse>('/api/estimate/calculate', {}); this.current = data }
      catch (e: unknown) { this.error = (e as { message?: string }).message ?? 'Ошибка расчёта' }
      finally { this.loading = false }
    },
  },
})
