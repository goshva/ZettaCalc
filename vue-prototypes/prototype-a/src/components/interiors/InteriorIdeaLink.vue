<script setup lang="ts">
/**
 * Связь идеи со сметами — ideaLink(x) из оригинала.
 *
 * Показываем помещения, воплощающие идею (realizationIds), а если таких нет —
 * помещения того же типа в той же палитре. Более двух карточек листаются
 * лентой со стрелками, как relatedrail.
 */
import { computed, ref } from 'vue'
import type { InteriorItem } from '@/types'
import { useInteriorsStore } from '@/stores/interiors'
import InteriorCard from './InteriorCard.vue'

const props = defineProps<{ item: InteriorItem }>()

const interiors = useInteriorsStore()

const exact = computed(() => interiors.realizationsOf(props.item))
const similar = computed(() => interiors.matchedRooms(props.item))

/** Подпись ссылки на подбор по палитре — как в оригинальном ideaLink(x) */
const matchingLabel = computed(() => {
  const room = props.item.room.toLowerCase()
  const noun = room === 'кухня' ? 'кухни' : room === 'спальня' ? 'спальни' : 'помещения'
  return `Смотреть ${noun} со сметой в этой палитре →`
})

const rail = ref<HTMLElement | null>(null)
function slide(delta: number) {
  rail.value?.scrollBy({ left: delta * rail.value.clientWidth * 0.8, behavior: 'smooth' })
}
</script>

<template>
  <section v-if="item.type === 'idea'" class="section ideatoestimate">
    <div class="row">
      <div>
        <div class="eyebrow">От идеи — к смете</div>
        <h2>{{ exact.length ? 'Варианты воплощения этой идеи' : 'Помещения со сметой в этой палитре' }}</h2>
      </div>
      <div v-if="exact.length > 2" class="scrolltools">
        <button type="button" aria-label="Предыдущие помещения" @click="slide(-1)">←</button>
        <button type="button" aria-label="Следующие помещения" @click="slide(1)">→</button>
      </div>
    </div>
    <p>
      {{ exact.length
        ? 'Та же идея в разных ценовых вариантах. Откройте решение, чтобы посмотреть его состав.'
        : `Сравните решения для того же типа помещения в палитре «${item.palette}». Это похожие по цветовой схеме интерьеры, а не точное воплощение этой идеи.` }}
    </p>

    <div v-if="exact.length > 2" ref="rail" class="relatedrail">
      <InteriorCard v-for="y in exact" :key="y.id" :item="y" :link-from="item.id" />
    </div>
    <div v-else-if="exact.length" class="grid largegrid">
      <InteriorCard v-for="y in exact" :key="y.id" :item="y" :link-from="item.id" />
    </div>
    <RouterLink v-else-if="similar.length" class="btn secondary" :to="`/interiors/matching/${item.id}`">
      {{ matchingLabel }}
    </RouterLink>
    <template v-else>
      <p class="scope">Подходящих смет пока нет в демонстрационной выборке.
        Подготовим подборку по выбранной идее к встрече.</p>
      <RouterLink class="btn secondary" to="/appointment">Записаться в Интерьерный центр</RouterLink>
    </template>
  </section>
</template>
