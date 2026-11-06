/**
 * Русские подписи машинных значений API.
 *
 * Значения в ответах (и в контракте `schemas/json/cpr.schema.json`) остаются
 * английскими слагами — интерфейс не должен показывать их пользователю.
 * Здесь только слой отображения: неизвестное значение отдаётся как есть,
 * чтобы не терять данные при расхождении mock/Go и фронтенда.
 */

const DOC_TYPE: Record<string, string> = {
  drawing: 'Чертёж',
  specification: 'Ведомость',
  schedule: 'График работ',
  estimate: 'Смета',
  attachment: 'Приложение',
}

const DOC_STATUS: Record<string, string> = {
  ready: 'Готов',
  pending_review: 'На проверке',
  pending: 'Ожидает',
  draft: 'Черновик',
}

const CPR_STATUS: Record<string, string> = {
  draft: 'Черновик',
  pending_review: 'На проверке',
  approved_for_tender: 'Утверждён для тендера',
  in_tender: 'В тендере',
  awarded: 'Подрядчик выбран',
  rejected: 'Отклонён',
  ready: 'Готов',
}

const BOOKING_STATUS: Record<string, string> = {
  pending_confirmation: 'Ожидает подтверждения',
  pending: 'Ожидает',
  confirmed: 'Подтверждена',
  cancelled: 'Отменена',
  closed: 'Закрыта',
}

const REF_CATEGORY: Record<string, string> = {
  kitchen: 'Кухня',
  living: 'Гостиная',
  bathroom: 'Санузел',
  bedroom: 'Спальня',
}

const REF_STYLE: Record<string, string> = {
  modern: 'Современный',
  contemporari: 'Контемпорари',
  warm: 'Тёплый',
  cool: 'Холодный',
  cozy: 'Уютный',
  minimalist: 'Минимализм',
  minimal: 'Минимализм',
  soft: 'Мягкий',
  bright: 'Светлый',
  classical: 'Классика',
  neoclassica: 'Неоклассика',
  calm: 'Спокойный',
  loft: 'Лофт',
}

export const docTypeLabel = (v: string): string => DOC_TYPE[v] ?? v
export const docStatusLabel = (v: string): string => DOC_STATUS[v] ?? v
export const cprStatusLabel = (v: string): string => CPR_STATUS[v] ?? v
export const bookingStatusLabel = (v: string): string => BOOKING_STATUS[v] ?? v
export const refCategoryLabel = (v: string): string => REF_CATEGORY[v] ?? v
export const refStyleLabel = (v: string): string => REF_STYLE[v] ?? v
