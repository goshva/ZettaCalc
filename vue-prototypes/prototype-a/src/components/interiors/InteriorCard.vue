<script setup lang="ts">
/**
 * Карточка интерьера — перенос card(x, pick, priceTier) из оригинального
 * приложения «Интерьеры».
 *
 * Используется в лентах, блоке похожих, избранном и в составе квартиры.
 * Разметка и классы совпадают с прототипом, поэтому оформление приходит
 * из /api/apps/interiors/css.
 */
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useFavoritesStore } from '@/stores/favorites'
import { useInteriorsStore } from '@/stores/interiors'
import type { InteriorItem } from '@/types'
import { cardCaption, isCalc, money, total } from '@/utils/interiors'
import { imgUrl } from '@/composables/useImg'
import PaletteRow from './PaletteRow.vue'

const props = withDefaults(defineProps<{
  item: InteriorItem
  /** Показать кнопку «Выбрать для встречи» */
  pick?: boolean
  /** Ценовой вариант для суммы; иначе берётся из стора */
  tier?: number
  /** Идея-источник: ссылки ведут с параметром ?from= */
  linkFrom?: string
}>(), { pick: false, tier: undefined, linkFrom: '' })

const favorites = useFavoritesStore()
const interiors = useInteriorsStore()
const { ideas: savedIds, meeting: meetingIds } = storeToRefs(favorites)

const saved = computed(() => savedIds.value.includes(props.item.id))
const picked = computed(() => meetingIds.value.includes(props.item.id))
const caption = computed(() => cardCaption(props.item))

/** В ленте квартир сумма зависит от фильтра «Вариант бюджета» — как в оригинале */
const budgetTier = computed(() => {
  if (props.item.type !== 'apt') return 0
  if (interiors.filter.budget === 'comfort') return Math.min(1, (props.item.costs?.length ?? 1) - 1)
  return 0
})
const tier = computed(() => props.tier ?? budgetTier.value)

/** «Расчёт готовится» — когда выбран вариант, которого в исходных файлах нет */
const pending = computed(
  () =>
    props.item.type === 'apt' &&
    (interiors.filter.budget === 'optimal' ||
      (interiors.filter.budget === 'comfort' && (props.item.costs?.length ?? 1) < 2)),
)

const badge = computed(() => {
  const x = props.item
  if (x.type === 'idea') return null
  if (x.planOnly) return 'Планировка'
  if (x.demo) return 'Пример'
  return 'Есть смета'
})

const prefix = computed(
  () => (props.item.costs?.length ?? 0) > 1 && !(props.item.type === 'apt' && interiors.filter.budget),
)

const path = computed(() => {
  const kind = props.item.type === 'apt' ? 'apartment' : props.item.type === 'room' ? 'room' : 'idea'
  return `/interiors/${kind}/${props.item.id}`
})
const to = computed(() => (props.linkFrom ? { path: path.value, query: { from: props.linkFrom } } : path.value))

const altText = computed(() => `${props.item.room} — ${props.item.title}`)
</script>

<template>
  <article class="card">
    <div class="photo">
      <RouterLink :to="to">
        <img :src="imgUrl(item.image)" :alt="altText" loading="lazy"
             :style="item.planOnly ? 'object-fit:contain;background:#fff' : undefined">
      </RouterLink>
      <span v-if="badge" class="badge">{{ badge }}</span>
      <button class="heart" :class="{ saved }" type="button"
              :title="saved ? 'Убрать из избранного' : 'Сохранить в избранное'"
              :aria-label="(saved ? 'Убрать из избранного: ' : 'Сохранить в избранное: ') + item.title"
              :aria-pressed="saved"
              @click="favorites.toggleIdea(item.id)">
        {{ saved ? '♥' : '♡' }}
      </button>
      <button v-if="pick" class="pick" :class="{ on: picked }" type="button"
              :aria-pressed="picked" @click="favorites.toggleMeeting(item.id)">
        {{ picked ? '✓ Для встречи' : 'Выбрать для встречи' }}
      </button>
    </div>

    <p class="meta">{{ item.room }} · {{ item.style }}</p>
    <p v-if="item.complex" class="complex">{{ item.complex }}</p>
    <h3><RouterLink :to="to">{{ item.title }}</RouterLink></h3>

    <PaletteRow :colors="item.colors" :name="item.palette" />

    <div class="cardbottom">
      <div>
        <strong v-if="isCalc(item) && !pending" class="price">
          {{ prefix ? 'от ' : '' }}{{ money(total(item, tier)) }}
        </strong>
        <strong v-else-if="isCalc(item)" class="price">Расчёт готовится</strong>
        <small>{{ caption }}</small>
      </div>
      <RouterLink class="textlink" :to="to">Подробнее →</RouterLink>
    </div>
  </article>
</template>
