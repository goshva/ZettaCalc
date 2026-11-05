import { defineStore } from 'pinia'

// Избранное («Мои идеи») и выбор для встречи — хранятся в localStorage,
// как в оригинале прототипа (ключи zh43-favorites / zh43-meeting заменены на fh_*)
function load(key: string): string[] {
  try {
    const raw = JSON.parse(localStorage.getItem(key) ?? '[]')
    return Array.isArray(raw) ? raw.filter((x): x is string => typeof x === 'string') : []
  } catch {
    return []
  }
}

const MAX_MEETING = 5

export const useFavoritesStore = defineStore('favorites', {
  state: () => ({
    ideas: load('fh_fav_ideas'),
    products: load('fh_fav_products'),
    meeting: load('fh_meeting'),
    choosing: false,
  }),
  getters: {
    ideasCount: (s) => s.ideas.length,
    productsCount: (s) => s.products.length,
    totalCount: (s) => s.ideas.length + s.products.length,
    meetingCount: (s) => s.meeting.length,
    maxMeeting: () => MAX_MEETING,
    isIdeaSaved: (s) => (id: string) => s.ideas.includes(id),
    isProductSaved: (s) => (id: string) => s.products.includes(id),
    isMeetingPicked: (s) => (id: string) => s.meeting.includes(id),
  },
  actions: {
    toggleIdea(id: string) {
      const i = this.ideas.indexOf(id)
      if (i >= 0) this.ideas.splice(i, 1)
      else this.ideas.push(id)
      // интерьер убрали из избранного — снимаем и выбор для встречи
      if (i >= 0) this.meeting = this.meeting.filter((x) => x !== id)
      this.persist()
    },
    toggleProduct(id: string) {
      const i = this.products.indexOf(id)
      if (i >= 0) this.products.splice(i, 1)
      else this.products.push(id)
      localStorage.setItem('fh_fav_products', JSON.stringify(this.products))
    },
    // В оригинале выбор ограничен пятью интерьерами и только из избранного
    toggleMeeting(id: string) {
      if (!this.ideas.includes(id)) return
      const i = this.meeting.indexOf(id)
      if (i >= 0) this.meeting.splice(i, 1)
      else if (this.meeting.length < MAX_MEETING) this.meeting.push(id)
      this.persist()
    },
    // Выбор из формы записи (picker в оригинале) заменяет список целиком
    setMeeting(ids: string[]) {
      this.meeting = ids.filter((id) => this.ideas.includes(id)).slice(0, MAX_MEETING)
      this.persist()
    },
    persist() {
      localStorage.setItem('fh_fav_ideas', JSON.stringify(this.ideas))
      localStorage.setItem('fh_meeting', JSON.stringify(this.meeting))
    },
  },
})
