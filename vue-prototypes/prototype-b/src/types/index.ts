// Общие типы
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
