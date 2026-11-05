import { defineStore } from 'pinia'
import { api } from '@/api/client'
import type { InteriorItem, InteriorsFilter, InteriorsKind } from '@/types'
import { roomType } from '@/utils/interiors'

const emptyFilter = (): InteriorsFilter => ({
  room: '', palette: '', style: '', search: '', area: '', budget: '',
})

export const useInteriorsStore = defineStore('interiors', {
  state: () => ({
    items: [] as InteriorItem[],
    categories: [] as string[],
    descriptions: [] as string[],
    selected: null as InteriorItem | null,
    filter: emptyFilter(),
    /** Выбранный ценовой вариант — глобальный tier в оригинале */
    tier: 0,
    /** Сколько карточек показывать в ленте — pageLimit в оригинале */
    pageLimit: 6,
    loading: false,
    error: '',
    loaded: false,
  }),
  getters: {
    byId: (s) => (id: string) => s.items.find((i) => i.id === id) ?? null,
    ideas(s) {
      return s.items.filter((i) => i.type === 'idea')
    },
    rooms(s) {
      return s.items.filter((i) => i.type === 'room' && !i.child)
    },
    apartments(s) {
      return s.items.filter((i) => i.type === 'apt')
    },
    allRooms(s) {
      return s.items.filter((i) => i.type === 'room')
    },
    uniqueRooms(s) {
      return Array.from(new Set(s.items.map((i) => i.room)))
    },
    uniqueStyles(s) {
      return Array.from(new Set(s.items.map((i) => i.style)))
    },
    uniquePalettes(s) {
      return Array.from(new Set(s.items.map((i) => i.palette)))
    },
    /** Список ленты: в квартирах в выборку попадают кухни-гостиные как представители помещений */
    listByKind() {
      return (kind: InteriorsKind) => {
        const type = { ideas: 'idea', rooms: 'room', apartments: 'apt' }[kind]
        return this.items.filter(
          (x) => x.type === type && (!x.child || x.id === 'flat5room1' || x.id === 'flat6room1'),
        )
      }
    },
    /** Лента с учётом фильтров — порт updateResults() */
    filteredList() {
      return (kind: InteriorsKind) => {
        const f = this.filter
        return this.listByKind(kind).filter(
          (x) =>
            (!f.room || x.room === f.room) &&
            (!f.palette || x.palette === f.palette) &&
            (!f.style || x.style === f.style) &&
            (!f.search ||
              `${x.title} ${x.style} ${x.room}`.toLowerCase().includes(f.search.toLowerCase())) &&
            (!f.area || (f.area === 'small' ? (x.area ?? 0) < 60 : (x.area ?? 0) >= 60)),
        )
      }
    },
    /** Помещения квартиры в порядке из children */
    childrenOf() {
      return (x: InteriorItem) =>
        (x.children ?? []).map((id) => this.byId(id)).filter(Boolean) as InteriorItem[]
    },
    /** Похожие интерьеры для карточки: сначала та же палитра */
    relatedTo() {
      return (x: InteriorItem) => {
        const calc = x.type !== 'idea'
        return this.items
          .filter(
            (y) =>
              y.id !== x.id &&
              roomType(y.room) === roomType(x.room) &&
              (calc ? y.type === 'room' : y.type === 'idea'),
          )
          .sort((a, b) => Number(b.palette === x.palette) - Number(a.palette === x.palette))
      }
    },
    /** Идеи, реализованные в помещении (realizationIds) */
    realizationsOf() {
      return (x: InteriorItem) =>
        (x.realizationIds ?? []).map((id) => this.byId(id)).filter(Boolean) as InteriorItem[]
    },
    /** Помещения в той же палитре — для подбора по палитре */
    matchedRooms() {
      return (x: InteriorItem) =>
        this.items.filter(
          (y) => y.type === 'room' && roomType(y.room) === roomType(x.room) && y.palette === x.palette,
        )
    },
  },
  actions: {
    async fetch() {
      if (this.loaded || this.loading) return
      this.loading = true
      this.error = ''
      try {
        const [itemsRes, catsRes] = await Promise.all([
          api.get<InteriorItem[]>('/api/interiors'),
          api.get<string[]>('/api/interiors/categories'),
        ])
        this.items = itemsRes.data
        this.categories = catsRes.data
        try {
          const descsRes = await api.get<string[]>('/api/interiors/descriptions')
          this.descriptions = descsRes.data
        } catch {
          this.descriptions = []
        }
        this.loaded = true
      } catch (e: unknown) {
        this.error = (e as { message?: string }).message ?? 'Ошибка загрузки интерьеров'
      } finally {
        this.loading = false
      }
    },
    async fetchById(id: string) {
      if (this.byId(id)) return
      this.loading = true
      this.error = ''
      try {
        const { data } = await api.get<InteriorItem>(`/api/interiors/${id}`)
        this.selected = data
        if (!this.byId(id)) this.items.push(data)
      } catch (e: unknown) {
        this.error = (e as { message?: string }).message ?? 'Ошибка загрузки детали'
      } finally {
        this.loading = false
      }
    },
    setTier(value: number) {
      this.tier = value
    },
    setFilter(patch: Partial<InteriorsFilter>) {
      this.filter = { ...this.filter, ...patch }
    },
    clearFilter() {
      this.filter = emptyFilter()
    },
    clearFilterKey(key: keyof InteriorsFilter) {
      this.filter = { ...this.filter, [key]: '' }
    },
    resetPaging() {
      this.pageLimit = 6
    },
    showMore() {
      this.pageLimit += 6
    },
  },
})
