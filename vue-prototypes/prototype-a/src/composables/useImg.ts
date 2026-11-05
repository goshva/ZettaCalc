/**
 * Утилита для преобразования image-ключей API в URL.
 *
 * Ключи из API (idea-kitchen, hero, cat-furniture…) маппятся в:
 *   - Полные пути (/api/images/file/xxx.jpg) — оставляем как есть
 *   - DB-ключи → GET /api/images/{key} (через FastAPI FileResponse)
 *   - Static-ключи → /images/{key}.jpg (Vite public)
 */

const STATIC_KEYS = new Set([
  'hero', 'video', 'app', 'cpr', 'expert', 'auction', 'ceo', 'plan',
  'repair', 'center-store', 'center-pano', 'center-kitchen',
  'center-stone', 'center-wardrobe', 'ds-e06b4936da',
])

const API_BASE = import.meta.env.VITE_API_URL ?? 'http://127.0.0.1:8000'

export function imgUrl(key: string): string {
  if (!key) return ''
  // Уже полный путь из dataset.json (например /api/images/file/a_main_img_0003.jpg)
  if (key.startsWith('/api/images/') || key.startsWith('http')) {
    return key
  }
  if (STATIC_KEYS.has(key)) {
    return `/images/${key}.jpg`
  }
  return `${API_BASE}/api/images/${key}`
}
