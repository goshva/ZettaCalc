<script setup lang="ts">
/**
 * Страница интерьера — перенос detail(x) из оригинального приложения.
 *
 * Один компонент обслуживает идеи и помещения: различие только в блоке
 * со сметой (compactEstimate), разделе товаров и связи идеи со сметами.
 * Разметка, классы и тексты совпадают с прототипом.
 */
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import type { InteriorItem } from '@/types'
import { useInteriorsStore } from '@/stores/interiors'
import { isCalc } from '@/utils/interiors'
import InteriorGallery from './InteriorGallery.vue'
import InteriorPlanPanel from './InteriorPlanPanel.vue'
import InteriorCompactEstimate from './InteriorCompactEstimate.vue'
import InteriorActions from './InteriorActions.vue'
import InteriorProducts from './InteriorProducts.vue'
import InteriorIdeaLink from './InteriorIdeaLink.vue'
import InteriorWorks from './InteriorWorks.vue'
import InteriorBand from './InteriorBand.vue'
import InteriorRelatedRail from './InteriorRelatedRail.vue'
import ApartmentNav from './ApartmentNav.vue'
import PaletteRow from './PaletteRow.vue'

const props = defineProps<{ item: InteriorItem }>()

const route = useRoute()
const interiors = useInteriorsStore()

const calc = computed(() => isCalc(props.item))
const other = computed(() => interiors.relatedTo(props.item))
/** Идея-источник: ссылка «Вернуться к идее» приходит из query ?from= */
const from = computed(() => (route.query.from as string | undefined) ?? '')
const fromItem = computed(() => (from.value ? interiors.byId(from.value) : null))

const LEADS: Record<string, string> = {
  idea0: 'Шалфейные фасады и терракотовые акценты сочетаются с тёплыми древесными оттенками. '
    + 'Кухня и зона отдыха образуют единое пространство.',
  room1: 'Шалфейные фасады и тёплая терракота создают спокойную кухню-гостиную '
    + 'с выразительными цветовыми акцентами.',
}
const lead = computed(
  () => props.item.description
    || LEADS[props.item.id]
    || 'Цветовая палитра, фактуры и мебель создают цельный образ комнаты. '
      + 'Сохраните интерьер, чтобы обсудить его для вашего дома.',
)
</script>

<template>
  <div class="crumb">
    <RouterLink to="/interiors">Интерьеры</RouterLink> /
    <RouterLink :to="calc ? '/interiors/rooms' : '/interiors/ideas'">
      {{ calc ? 'Интерьеры со сметой' : 'Идеи интерьеров' }}
    </RouterLink> / {{ item.title }}
  </div>

  <p v-if="item.demo" class="notice">Демонстрационный расчёт: цены и состав показывают
    работу прототипа.</p>

  <RouterLink v-if="fromItem" class="backidea" :to="`/interiors/idea/${fromItem.id}`">
    ← Вернуться к идее «{{ fromItem.title }}»
  </RouterLink>

  <ApartmentNav :item="item" />

  <section class="detailtop split">
    <InteriorPlanPanel v-if="item.planOnly" :item="item"
                       title="Планировка квартиры"
                       note="Отдельная визуализация этого помещения пока не добавлена.
                             Его состав и стоимость доступны в смете справа." />
    <InteriorGallery v-else :item="item" />

    <div>
      <div class="eyebrow">{{ item.room }} · {{ item.style }}</div>
      <h1>{{ item.title }}</h1>
      <PaletteRow :colors="item.colors" :name="item.palette" />
      <p class="lead">{{ lead }}</p>

      <InteriorCompactEstimate v-if="calc" :item="item" />
      <p v-else class="muted">Подберём мебель и отделочные материалы к встрече.
        Стоимость определим по составу подборки.</p>

      <InteriorActions :item="item" />

      <p v-if="item.parent">
        <RouterLink class="textlink" :to="`/interiors/apartment/${item.parent}`">
          Вся квартира и планировка →
        </RouterLink>
      </p>
    </div>
  </section>

  <section v-if="calc" class="section">
    <h2>Мебель и материалы этого интерьера</h2>
    <InteriorProducts :item="item" />
  </section>

  <InteriorIdeaLink v-if="item.type === 'idea'" :item="item" />

  <InteriorWorks v-if="item.realArchive" />

  <InteriorBand />

  <InteriorRelatedRail v-if="other.length"
                        :title="calc ? 'Другие помещения этого типа со сметой' : `Другие идеи: ${item.room.toLowerCase()}`"
                        :note="(calc ? 'Только этот тип помещения.' : 'Сначала — идеи в той же палитре.')
                          + ' Листайте, чтобы сравнить варианты.'"
                        :items="other" :from="item.id" />

  <InteriorBand variant="upload" />
</template>
