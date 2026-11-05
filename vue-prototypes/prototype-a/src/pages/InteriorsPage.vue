<script setup lang="ts">
/**
 * Лента интерьеров — перенос listing() + updateResults() из оригинала.
 *
 * Один компонент обслуживает три маршрута: /interiors/ideas, /interiors/rooms,
 * /interiors/apartments. Набор карточек, фильтры и подписи зависят от вида.
 */
import { computed, onMounted, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useRoute } from 'vue-router'
import { useInteriorsStore } from '@/stores/interiors'
import { useAppStyles } from '@/composables/useAppStyles'
import type { InteriorsKind } from '@/types'
import InteriorCard from '@/components/interiors/InteriorCard.vue'
import InteriorBand from '@/components/interiors/InteriorBand.vue'
import InteriorTabs from '@/components/interiors/InteriorTabs.vue'
import PaletteRow from '@/components/interiors/PaletteRow.vue'

useAppStyles('interiors', 'inner-design65')

const route = useRoute()
const interiors = useInteriorsStore()
const { filter, pageLimit, loading, error } = storeToRefs(interiors)

const kind = computed<InteriorsKind>(() => {
  const k = route.meta.kind as InteriorsKind | undefined
  return k ?? 'rooms'
})

const names: Record<InteriorsKind, string> = {
  ideas: 'Идеи интерьеров',
  rooms: 'Интерьеры со сметой',
  apartments: 'Квартиры со сметой',
}
const intro: Record<InteriorsKind, string> = {
  ideas: 'Более 1500 идей в каталоге ZETTA HOME. Найдите свой образ и сохраните понравившееся.',
  rooms: 'Конкретные решения для комнат: состав, товары и стоимость.',
  apartments: 'Планировки, помещения и сметы для квартиры целиком.',
}

const isApartments = computed(() => kind.value === 'apartments')
const list = computed(() => interiors.filteredList(kind.value))
const shown = computed(() => list.value.slice(0, pageLimit.value))
const roomOptions = computed(() => Array.from(new Set(interiors.listByKind(kind.value).map((x) => x.room))))
const styleOptions = computed(() => Array.from(new Set(interiors.listByKind(kind.value).map((x) => x.style))))
const paletteOptions = computed(() => Array.from(new Set(interiors.listByKind(kind.value).map((x) => x.palette))))

const activeFilters = computed(() =>
  Object.entries(filter.value)
    .filter(([, v]) => v)
    .map(([k, v]) => ({
      key: k,
      label: ({ small: 'До 60 м²', large: 'От 60 м²', base: 'Базовый', optimal: 'Оптимальный', comfort: 'Комфортный' } as Record<string, string>)[v] ?? v,
    })),
)

function paletteFor(name: string) {
  return interiors.listByKind(kind.value).find((x) => x.palette === name)
}

function apply(key: 'room' | 'style' | 'area' | 'budget', value: string) {
  interiors.setFilter({ [key]: value })
  interiors.resetPaging()
}

function pickPalette(value: string) {
  interiors.setFilter({ palette: value })
  interiors.resetPaging()
}

function reset() {
  interiors.clearFilter()
  interiors.resetPaging()
}

onMounted(() => interiors.fetch())
// при переходе между лентами фильтры сбрасываются, как при повторном listing(kind)
watch(kind, () => reset())
</script>

<template>
  <section class="section">
    <div class="wrap">
      <div class="crumb">
        <RouterLink to="/interiors">Интерьеры</RouterLink> / {{ names[kind] }}
      </div>
      <h1>{{ names[kind] }}</h1>
      <p class="catalogintro">{{ intro[kind] }}</p>

      <InteriorTabs :active="kind" />

      <div class="compactfilters">
        <label>{{ isApartments ? 'Количество комнат' : 'Тип помещения' }}
          <select :value="filter.room" :aria-label="isApartments ? 'Количество комнат' : 'Тип помещения'"
                  @change="apply('room', ($event.target as HTMLSelectElement).value)">
            <option value="">Все</option>
            <option v-for="r in roomOptions" :key="r" :value="r">{{ r }}</option>
          </select>
        </label>

        <label>Стиль
          <select :value="filter.style" aria-label="Стиль"
                  @change="apply('style', ($event.target as HTMLSelectElement).value)">
            <option value="">Все</option>
            <option v-for="s in styleOptions" :key="s" :value="s">{{ s }}</option>
          </select>
        </label>

        <details class="palettedrop">
          <summary>
            <span>Цветовая палитра</span>
            <b>{{ filter.palette || 'Все палитры' }} ▾</b>
          </summary>
          <div class="palettepopup">
            <button type="button" data-palette="" @click="pickPalette('')">Все палитры</button>
            <button v-for="p in paletteOptions" :key="p" type="button" @click="pickPalette(p)">
              <PaletteRow v-if="paletteFor(p)" :colors="paletteFor(p)!.colors" :name="p" />
              <span v-else>{{ p }}</span>
            </button>
          </div>
        </details>

        <template v-if="isApartments">
          <label>Площадь
            <select :value="filter.area" aria-label="Площадь"
                    @change="apply('area', ($event.target as HTMLSelectElement).value)">
              <option value="">Любая</option>
              <option value="small">До 60 м²</option>
              <option value="large">От 60 м²</option>
            </select>
          </label>
          <label>Вариант бюджета
            <select :value="filter.budget" aria-label="Вариант бюджета"
                    @change="apply('budget', ($event.target as HTMLSelectElement).value)">
              <option value="">Все варианты</option>
              <option value="base">Базовый</option>
              <option value="optimal">Оптимальный</option>
              <option value="comfort">Комфортный</option>
            </select>
          </label>
        </template>

        <button class="textlink resetcompact" type="button" @click="reset">Сбросить</button>
      </div>

      <div id="activefilters">
        <button v-for="f in activeFilters" :key="f.key" class="pill" type="button"
                @click="interiors.clearFilterKey(f.key as never)">
          {{ f.label }} ×
        </button>
      </div>

      <p v-if="loading && !interiors.items.length" class="muted">Загрузка интерьеров…</p>
      <p v-else-if="error" class="error">{{ error }}</p>

      <div id="results" v-else>
        <div class="results">
          Показано {{ shown.length }} из {{ list.length }}{{ kind === 'ideas' ? ' · демонстрационная выборка каталога' : '' }}
        </div>
        <div v-if="list.length" class="grid cataloggrid" :class="kind === 'ideas' ? 'ideasgrid' : 'largegrid'">
          <InteriorCard v-for="x in shown" :key="x.id" :item="x" />
        </div>
        <div v-else class="empty">
          <h3>Подходящих вариантов пока нет в прототипе</h3>
          <button class="btn secondary" type="button" @click="reset">Сбросить фильтры</button>
        </div>
        <div class="loadmore">
          <button v-if="list.length > pageLimit" class="btn secondary" type="button"
                  @click="interiors.showMore()">Показать ещё</button>
          <a v-else-if="kind === 'ideas'" class="btn secondary" href="https://zettahome.ru/design/"
             target="_blank" rel="noopener">Продолжить в полном каталоге ↗</a>
        </div>
      </div>

      <InteriorBand />
      <InteriorBand variant="upload" />
    </div>
  </section>
</template>
