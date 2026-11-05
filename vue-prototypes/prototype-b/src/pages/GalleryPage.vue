<script setup lang="ts">
import { onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useReferencesStore } from '@/stores/references'
import { imgUrl } from '@/composables/useImg'

const refs = useReferencesStore()
const { items, loading } = storeToRefs(refs)
onMounted(() => refs.fetch())
</script>
<template>
  <section class="section">
    <div class="wrap">
      <h2>Галерея интерьеров</h2>
      <p class="muted" style="margin-bottom: 32px">Все референсы в одном месте</p>
      <p v-if="loading" class="muted">Загрузка…</p>
      <div v-else class="grid-3">
        <div v-for="r in items" :key="r.id" class="card">
          <div class="card-img"><img :src="imgUrl(r.image)" :alt="r.title" /></div>
          <div class="card-body">
            <h3>{{ r.title }}</h3>
            <div class="tags"><span v-for="s in r.styles" :key="s" class="tag">{{ s }}</span></div>
            <div class="price">от {{ formatPrice(r.price_from) }}</div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
<script lang="ts">
import { formatPrice } from '@/api/client'
export default { data() { return { formatPrice } } }
</script>
<style scoped>
.card-img { aspect-ratio: 4/3; overflow: hidden; }
.card-img img { width: 100%; height: 100%; object-fit: cover; }
.card-body { padding: 16px; display: flex; flex-direction: column; gap: 8px; }
.tags { display: flex; gap: 6px; flex-wrap: wrap; }
.price { font-size: 14px; font-weight: 600; }
</style>
