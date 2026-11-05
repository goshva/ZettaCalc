<script setup lang="ts">
import { computed } from 'vue'
import { useFavoritesStore } from '@/stores/favorites'
import { storeToRefs } from 'pinia'

// Полноценная лента идей — итерация 3/5; здесь избранное уже работает
const favorites = useFavoritesStore()
const { ideas } = storeToRefs(favorites)
const count = computed(() => ideas.value.length)
</script>

<template>
  <section class="section">
    <div class="wrap">
      <span class="eyebrow">Мои идеи</span>
      <h2>Сохранённые интерьеры</h2>
      <p v-if="count === 0">Пока ничего не сохранено. Нажимайте ♡ на карточках интерьеров.</p>
      <p v-else>Сохранено идей: {{ count }} — лента появится в итерации 3.</p>
      <RouterLink class="btn" to="/#ideas">← К идеям</RouterLink>
    </div>
  </section>
</template>

<style scoped>
.section { padding: 90px 0; }
.wrap h2 { font-size: 40px; margin: 12px 0 14px; }
.wrap p { margin-bottom: 26px; }
</style>
