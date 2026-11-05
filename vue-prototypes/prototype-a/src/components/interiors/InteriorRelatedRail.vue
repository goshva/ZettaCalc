<script setup lang="ts">
/**
 * Лента похожих интерьеров — блок «other» из detail(x) и relatedrail.
 *
 * Кнопки ‹ › прокручивают ленту, как обработчик data-slide в оригинале.
 */
import { ref } from 'vue'
import type { InteriorItem } from '@/types'
import InteriorCard from './InteriorCard.vue'

defineProps<{ title: string; note: string; items: InteriorItem[]; from?: string }>()

const rail = ref<HTMLElement | null>(null)

function slide(delta: number) {
  rail.value?.scrollBy({ left: delta * rail.value.clientWidth * 0.8, behavior: 'smooth' })
}
</script>

<template>
  <section class="section">
    <div class="row">
      <h2>{{ title }}</h2>
      <div class="scrolltools">
        <button type="button" aria-label="Предыдущие варианты" @click="slide(-1)">←</button>
        <button type="button" aria-label="Следующие варианты" @click="slide(1)">→</button>
      </div>
    </div>
    <p class="scope">{{ note }}</p>
    <div ref="rail" class="relatedrail">
      <InteriorCard v-for="y in items" :key="y.id" :item="y" :link-from="from" />
    </div>
  </section>
</template>
