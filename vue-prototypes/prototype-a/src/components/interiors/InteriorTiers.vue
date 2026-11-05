<script setup lang="ts">
/**
 * Переключатель ценовых вариантов — variants(x) из оригинала.
 *
 * У интерьеров из архива квартир опубликованы только «Базовый» и «Комфортный»:
 * оптимальный вариант помечен как готовящийся и недоступен.
 */
import { computed } from 'vue'
import type { InteriorItem } from '@/types'
import { useInteriorsStore } from '@/stores/interiors'

const props = defineProps<{ item: InteriorItem }>()

const interiors = useInteriorsStore()
const tier = computed(() => interiors.tier)

const buttons = computed(() => {
  const x = props.item
  if (!x.costs?.length) return []
  if (x.realArchive) {
    return [
      { label: 'Базовый', value: 0, ready: true },
      { label: 'Оптимальный', value: 1, ready: false },
      { label: 'Комфортный', value: 1, ready: x.costs.length > 1 },
    ]
  }
  if (x.costs.length > 1) {
    return ['Базовый', 'Оптимальный', 'Комфортный'].map((label, value) => ({
      label, value, ready: true,
    }))
  }
  return []
})
</script>

<template>
  <div v-if="buttons.length" class="tiers" aria-label="Вариант бюджета">
    <button v-for="b in buttons" :key="b.label" type="button"
            :disabled="!b.ready"
            :title="b.ready ? undefined : 'Расчёт ещё не предоставлен'"
            :class="{ active: tier === b.value && b.ready }"
            @click="interiors.setTier(b.value)">
      {{ b.label }}<small v-if="!b.ready">готовится</small>
    </button>
  </div>
</template>
