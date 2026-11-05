<script setup lang="ts">
/**
 * Страница помещения — маршрут /interiors/room/:id.
 *
 * Помещения квартир приходят с parent: тогда сверху появляется навигация
 * по помещениям и ссылка на общий бюджет квартиры.
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
      <p v-if="loading && !item" class="muted">Загрузка помещения…</p>
      <p v-else-if="error" class="notice">{{ error }}</p>
      <p v-else-if="!item" class="notice">Помещение не найдено в демонстрационной выборке.</p>
      <InteriorDetail v-else :item="item" />
    </div>
  </section>
</template>
