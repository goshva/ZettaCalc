<script setup lang="ts">
/**
 * Смета в карточке интерьера — compactEstimate(x) из оригинала.
 *
 * Если у интерьера есть построчный состав (архив квартир) — показываем
 * сгруппированные разделы. Иначе суммы по категориям из датасета.
 */
import { computed } from 'vue'
import type { InteriorItem } from '@/types'
import { useInteriorsStore } from '@/stores/interiors'
import { categoryTotal, money, total } from '@/utils/interiors'
import InteriorTiers from './InteriorTiers.vue'
import InteriorLines from './InteriorLines.vue'

const props = defineProps<{ item: InteriorItem }>()

const interiors = useInteriorsStore()
const tier = computed(() => interiors.tier)
const categories = computed(() => interiors.categories)
const descriptions = computed(() => interiors.descriptions)
const sum = computed(() => total(props.item, tier.value))
</script>

<template>
  <div class="compactestimate">
    <InteriorTiers :item="item" />

    <div class="estimate-total">
      <span>{{ item.realArchive ? 'Подборка' : 'Итого' }}</span>
      <strong>{{ money(sum) }}</strong>
    </div>

    <template v-if="item.lines?.length">
      <InteriorLines :lines="item.lines" grouped />
      <p class="scope">Расчёт из предоставленной таблицы. Заказные позиции уточняются.
        Смета ремонтных работ в этот итог не включена.</p>
    </template>

    <template v-else>
      <details v-for="(c, i) in categories" :key="c">
        <summary>
          <span>{{ c }}</span>
          <b>{{ money(categoryTotal(item, tier, i)) }} +</b>
        </summary>
        <p class="small">{{ descriptions[i] }}</p>
        <p class="scope">
          {{ item.demo
            ? 'Демонстрационная сумма. Товарная спецификация для этого примера ещё не добавлена.'
            : 'В источнике опубликован итог категории. Построчная спецификация пока не предоставлена.' }}
        </p>
      </details>
    </template>
  </div>
</template>
