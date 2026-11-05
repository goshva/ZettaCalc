<script setup lang="ts">
/**
 * Страница идеи интерьера — маршрут /interiors/idea/:id.
 *
 * Оригинал рисует идеи и помещения одной функцией detail(x), поэтому
 * страница сводится к общей карточке интерьера.
 */
import { useInteriorItem } from '@/composables/useInteriorItem'
import { useAppStyles } from '@/composables/useAppStyles'
import InteriorDetail from '@/components/interiors/InteriorDetail.vue'

useAppStyles('interiors', 'inner-design65')

const { item, loading, error } = useInteriorItem()
</script>

<template>
  <section class="section">
    <div class="wrap">
      <p v-if="loading && !item" class="muted">Загрузка идеи…</p>
      <p v-else-if="error" class="notice">{{ error }}</p>
      <p v-else-if="!item" class="notice">Интерьер не найден в демонстрационной выборке.</p>
      <InteriorDetail v-else :item="item" />
    </div>
  </section>
</template>
