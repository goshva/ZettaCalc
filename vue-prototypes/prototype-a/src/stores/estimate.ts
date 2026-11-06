import { defineStore } from 'pinia'
import { api } from '@/api/client'
import type { EstimateResponse } from '@/types'

export const useEstimateStore = defineStore('estimate', {
  state: () => ({
    current: null as EstimateResponse | null,
    loading: false,
    error: '',
  }),
  getters: {
    total: (s) => s.current?.total_rub ?? 0,
    range: (s) => s.current ? `${s.current.range_min} — ${s.current.range_max}` : '',
  },
  actions: {
    async calculate(referenceId = '') {
      this.loading = true
      this.error = ''
      try {
        const { data } = await api.post<EstimateResponse>('/api/estimate/calculate', {
          reference_id: referenceId,
        })
        this.current = data
      } catch (e: unknown) {
        this.error = (e as { message?: string }).message ?? 'Ошибка расчёта'
      } finally {
        this.loading = false
      }
    },
    async fetch(id: string) {
      const { data } = await api.get<EstimateResponse>(`/api/estimate/${id}`)
      this.current = data
    },
  },
})
