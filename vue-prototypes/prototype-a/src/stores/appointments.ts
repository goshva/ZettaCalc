import { defineStore } from 'pinia'
import { api } from '@/api/client'
import type {
  Appointment,
  AppointmentOptions,
  AppointmentRequest,
  AppointmentResponse,
  AppointmentValidation,
} from '@/types'

// Ключ идемпотентности живёт до первой успешной записи: повторная отправка
// формы с тем же ключом не создаёт вторую запись, а возвращает уже созданную.
function newKey(): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) return crypto.randomUUID()
  return 'appt-' + Date.now().toString(36) + '-' + Math.random().toString(36).slice(2, 10)
}

export const useAppointmentsStore = defineStore('appointments', {
  state: () => ({
    options: null as AppointmentOptions | null,
    validation: null as AppointmentValidation | null,
    created: null as Appointment | null,
    /** Имена файлов со страницы «Загрузите дизайн-проект до встречи» */
    projectFiles: [] as string[],
    idempotencyKey: newKey(),
    replayed: false,
    loading: false,
    checking: false,
    saving: false,
    error: '',
  }),
  getters: {
    isDone: (s) => s.created !== null,
  },
  actions: {
    async fetchOptions() {
      this.loading = true
      this.error = ''
      try {
        const { data } = await api.get<AppointmentOptions>('/api/center/appointment/options')
        this.options = data
      } catch (e: unknown) {
        this.error = (e as { message?: string }).message ?? 'Не удалось загрузить форму записи'
      } finally {
        this.loading = false
      }
    },
    async dryRun(payload: AppointmentRequest): Promise<boolean> {
      this.checking = true
      this.error = ''
      try {
        const { data } = await api.post<AppointmentValidation>(
          '/api/center/appointment/dry-run', payload,
        )
        this.validation = data
        return data.ok
      } catch (e: unknown) {
        this.validation = null
        this.error = (e as { message?: string }).message ?? 'Не удалось проверить форму'
        return false
      } finally {
        this.checking = false
      }
    },
    async submit(payload: AppointmentRequest, approved = true) {
      this.saving = true
      this.error = ''
      try {
        const { data } = await api.post<AppointmentResponse>(
          '/api/center/appointment',
          payload,
          {
            params: { approved },
            headers: { 'Idempotency-Key': this.idempotencyKey },
          },
        )
        this.created = data.appointment
        this.replayed = data.idempotent_replay
      } catch (e: unknown) {
        const err = e as { response?: { data?: { detail?: unknown } }; message?: string }
        const detail = err.response?.data?.detail
        this.error = typeof detail === 'string'
          ? detail
          : err.message ?? 'Не удалось создать запись'
      } finally {
        this.saving = false
      }
    },
    reset() {
      this.created = null
      this.replayed = false
      this.validation = null
      this.error = ''
      this.projectFiles = []
      this.idempotencyKey = newKey()
    },
  },
})
