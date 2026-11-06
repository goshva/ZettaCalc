<script setup lang="ts">
import { onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useReferencesStore } from '@/stores/references'
import ReferenceCard from '@/components/ReferenceCard.vue'

const refs = useReferencesStore()
const { items, loading } = storeToRefs(refs)

onMounted(() => refs.fetch())
</script>

<template>
  <section class="section">
    <div class="wrap">
      <h2>Каталог интерьеров</h2>
      <p class="muted" style="margin-bottom: 32px">Все идеи и референсы</p>

      <p v-if="loading" class="muted">Загрузка…</p>
      <div v-else class="grid cols-3">
        <ReferenceCard v-for="r in items" :key="r.id" :ref-item="r" />
      </div>
    </div>
  </section>
</template>

<style scoped>
.grid.cols-3 {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
}
@media (min-width: 1100px) {
  .grid.cols-3 { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 24px 26px; }
}
</style>
