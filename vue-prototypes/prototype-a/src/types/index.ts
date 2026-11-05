// Общие TypeScript-типы для Fresh House API
export interface LoginRequest { email: string; password: string }
export interface TokenResponse { access_token: string; token_type: string; user_id: string; email: string }
export interface UserMe { user_id: string; email: string; name: string }

export interface ReferenceItem {
  id: string; title: string; category: 'kitchen' | 'living' | 'bathroom' | 'bedroom'
  image: string; styles: string[]; price_from: number; price_to: number
}

export interface TechCard {
  id: string; code: string; title: string; surface: string; material: string
  operations: string[]; consumption: Record<string, number>
  min_area_m2: number; max_unevenness: number | null; version: number
}

export interface EstimateLine {
  id: string; title: string; qty: number; unit: string; price: number
  sku: string | null; brand: string | null
}

export interface EstimateResponse {
  id: string; status: string; total_rub: number; range_min: number; range_max: number
  days: number; lines: EstimateLine[]; assumptions: string[]
}

export interface CPRDocument {
  id: string; title: string; type: 'drawing' | 'specification' | 'schedule'
  status: 'ready' | 'pending_review' | 'draft'; url: string | null
}

export interface CPRResponse {
  id: string; version: number; project_id: string
  estimate: EstimateResponse; documents: CPRDocument[]; status: string
}

export interface CartItem {
  sku: string; brand: string; title: string; price: number; qty: number; image: string
}

export interface CartResponse { items: CartItem[]; total_rub: number }

export interface ContractorBid {
  id: string; name: string; rating: number; price: number; days: number; scope: string
  image?: string
}

export interface PlannerItem {
  id: string; name: string; area: number; rooms: { name: string; area: number }[]
}

export interface BriefField {
  id: string; label: string; value: string | null; type: 'select' | 'text' | 'number' | 'boolean'
  options?: string[]
}

export interface BriefResponse {
  id: string; status: 'draft' | 'complete'
  fields: BriefField[]; created_at: string; updated_at: string
}

// ── Интерьеры (итерация 5) ──────────────────────────────────────────────────

export interface RoomLine {
  name: string
  description: string
  qty: string
  values: number[]
  url: string
  altUrl?: string
  image?: string
  note?: string
  catalogId?: string
}

export interface InteriorItem {
  id: string
  type: 'idea' | 'room' | 'apt'
  title: string
  room: string
  style: string
  palette: string
  colors: string[]
  image: string
  source?: string
  gallery?: string[]
  plan?: string
  area?: number
  costs?: number[][]
  rooms?: number | string[]
  children?: string[]
  description?: string
  variants?: string[]
  realArchive?: boolean
  complex?: string
  demo?: boolean
  child?: boolean
  parent?: string
  lines?: RoomLine[]
  /** Помещение показано только планом: визуализации нет */
  planOnly?: boolean
  /** Идеи, реализованные в этом помещении (связь idea → room) */
  realizationIds?: string[]
}

/** Фильтры списка интерьеров — как filter в оригинальном прототипе */
export interface InteriorsFilter {
  room: string
  palette: string
  style: string
  search: string
  area: string
  budget: string
}

export type InteriorsKind = 'ideas' | 'rooms' | 'apartments'

// Категории и описания в dataset.json — плоские массивы строк
export type InteriorCategory = string
export type InteriorDescription = string

// ── Запись в Интерьерный центр (итерация 4) ─────────────────────────────────

export interface AppointmentRequest {
  name: string
  phone: string
  tasks: string[]
  preferred_date: string | null
  preferred_time: string
  ideas: string[]
  files: string[]
  wish: string
  consent: boolean
}

export interface AppointmentValidation {
  ok: boolean
  errors: string[]
  warnings: string[]
  summary: string[]
}

export interface Appointment {
  id: string
  status: string
  created_at: string
  name: string
  phone: string
  tasks: string[]
  preferred_date: string | null
  preferred_time: string
  ideas: string[]
  files: string[]
  wish: string
  consent: boolean
  place: string
  validation: AppointmentValidation
}

export interface AppointmentResponse {
  id: string
  status: string
  created_at: string
  idempotent_replay: boolean
  appointment: Appointment
}

export interface AppointmentPlace {
  address: string
  mall: string
  phone: string
  phone_href: string
  map: string
  note: string
}

export interface AppointmentOptions {
  tasks: string[]
  times: string[]
  max_ideas: number
  max_files: number
  place: AppointmentPlace
  consent_url: string
  policy_url: string
  dry_run: string
  create: string
}
