<script setup lang="ts">
/**
 * Подбор по палитре — перенос matchedPage(x) из оригинального приложения.
 *
 * Открывается с идеи, когда точного воплощения нет: показываем помещения
 * того же типа в той же цветовой схеме.
 */
import { computed } from 'vue'
import { useInteriorsStore } from '@/stores/interiors'
import { useInteriorItem } from '@/composables/useInteriorItem'
import { useAppStyles } from '@/composables/useAppStyles'
import InteriorCard from '@/components/interiors/InteriorCard.vue'
import InteriorBand from '@/components/interiors/InteriorBand.vue'

useAppStyles('interiors', 'inner-design65')

const interiors = useInteriorsStore()
const { item, loading, error } = useInteriorItem()

const matched = computed(() => (item.value ? interiors.matchedRooms(item.value) : []))
</script>

<template>
  <section class="section">
    <div class="wrap">
      <p v-if="loading && !item" class="muted">Загрузка подборки…</p>
      <p v-else-if="error" class="notice">{{ error }}</p>
      <p v-else-if="!item" class="notice">Идея не найдена в демонстрационной выборке.</p>

      <template v-else>
        <div class="crumb">
          <RouterLink :to="`/interiors/idea/${item.id}`">← К идее «{{ item.title }}»</RouterLink>
        </div>
        <h1>Помещения со сметой<br>в палитре «{{ item.palette }}»</h1>
        <p class="lead">{{ item.room }} · похожая цветовая схема. Эти решения не являются
          точной реализацией выбранной идеи.</p>

        <div class="grid">
          <InteriorCard v-for="y in matched" :key="y.id" :item="y" :link-from="item.id" />
        </div>

        <p v-if="!matched.length" class="scope">Подходящих расчётов пока нет в демонстрационной
          выборке. Подготовим подборку по выбранной идее к встрече.</p>

        <InteriorBand />
      </template>
    </div>
  </section>
</template>
