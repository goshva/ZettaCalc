import type { InteriorItem, RoomLine } from '@/types'

/**
 * Расчёты и подписи интерьеров — дословный перенос вспомогательных функций
 * оригинального прототипа (total, isCalc, roomType, palette).
 *
 * Расчёты детерминированные: суммы берутся из данных, а не считаются LLM.
 */

export const money = (n: number) => new Intl.NumberFormat('ru-RU').format(Math.round(n)) + ' ₽'

/** Адрес карточки: в оригинале один маршрут /item/:id, здесь — по типу интерьера */
export function itemPath(x: InteriorItem): string {
  const kind = x.type === 'apt' ? 'apartment' : x.type === 'room' ? 'room' : 'idea'
  return `/interiors/${kind}/${x.id}`
}

/** Идея — образ без сметы; помещение и квартира считаются */
export const isCalc = (x: InteriorItem) => x.type !== 'idea'

/** Ценовой вариант: индекс в массиве costs */
export function total(x: InteriorItem, tier = 0): number {
  if (!x.costs?.length) return 0
  const row = x.costs[Math.min(tier, x.costs.length - 1)]
  return row.reduce((a, b) => a + b, 0)
}

export function categoryTotal(x: InteriorItem, tier: number, index: number): number {
  if (!x.costs?.length) return 0
  const row = x.costs[Math.min(tier, x.costs.length - 1)]
  return row[index] ?? 0
}

/** В оригинале «Ванная комната» и «Кухня-гостиная» сравниваются с более общими типами */
export function roomType(s: string): string {
  if (s === 'Ванная комната') return 'Ванная'
  if (s === 'Кухня-гостиная') return 'Кухня'
  return s
}

export const TIERS = ['Минимальный', 'Оптимальный', 'Комфортный']

export function tierLabels(x: InteriorItem): string[] {
  return x.variants?.length ? x.variants : TIERS.slice(0, Math.max(1, x.costs?.length ?? 1))
}

export function lineTotal(line: RoomLine, tier = 0): number {
  const values = line.values ?? []
  if (!values.length) return 0
  return values[Math.min(tier, values.length - 1)]
}

export function lineHref(line: RoomLine, tier: number): string {
  return tier && line.altUrl ? line.altUrl : line.url
}

/**
 * Группировка состава сметы — те же регулярные выражения, что в оригинале
 * (groupedLines), чтобы разделы совпадали с прототипом.
 */
const FINISH_RE = /покрытие|потолок|дверь|плинтус|краска|панель|молдинг|карниз|керамогранит|перегородка/
const EQUIP_RE = /свет|бра$|розет|выключ|рамка|ванна|смеситель|душ|полотенце|унитаз|раковин/

export const LINE_GROUPS = ['Мебель и декор', 'Отделочные материалы', 'Оснащение и сантехника'] as const
export type LineGroup = (typeof LINE_GROUPS)[number]

export function groupLines(lines: RoomLine[]): { name: LineGroup; items: RoomLine[] }[] {
  const groups: Record<LineGroup, RoomLine[]> = {
    'Мебель и декор': [],
    'Отделочные материалы': [],
    'Оснащение и сантехника': [],
  }
  for (const l of lines) {
    const n = l.name.toLowerCase()
    const key: LineGroup = FINISH_RE.test(n)
      ? 'Отделочные материалы'
      : EQUIP_RE.test(n)
        ? 'Оснащение и сантехника'
        : 'Мебель и декор'
    groups[key].push(l)
  }
  return LINE_GROUPS.filter((k) => groups[k].length).map((k) => ({ name: k, items: groups[k] }))
}

export function groupTotal(items: RoomLine[], tier: number): number {
  return items.reduce((sum, l) => sum + lineTotal(l, tier), 0)
}

/** Подписи площадки и плитки карточки — как в оригинальном card() */
export function cardCaption(x: InteriorItem): string {
  if (x.realArchive) return 'Мебель и материалы · работы отдельно'
  if (x.type === 'idea') return 'Идея интерьера'
  if (x.demo) return 'Демонстрационный расчёт'
  return 'Мебель, материалы и работы'
}

export function cardBadge(x: InteriorItem): string | null {
  if (x.type === 'idea') return null
  if (x.planOnly) return 'Планировка'
  if (x.demo) return 'Пример'
  return 'Есть смета'
}
