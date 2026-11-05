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
  <section class="hero" :style="{ backgroundImage: `linear-gradient(rgba(0,0,0,.7), rgba(0,0,0,.7)), url(${imgUrl('hero')})` }">
    <div class="wrap">
      <div class="eyebrow">Дизайн · ремонт · мебель · материалы</div>
      <h1>Ваша квартира.<br />Все решения <em>вместе.</em></h1>
      <p class="lead">Подберём мебель и материалы, рассчитаем ремонт и организуем реализацию. Прототип B — альтернативный дизайн.</p>
      <div class="cta">
        <RouterLink to="/gallery" class="btn gold">Найти свой интерьер</RouterLink>
        <RouterLink to="/calculator" class="btn ghost">Рассчитать стоимость</RouterLink>
      </div>
    </div>
  </section>
  <section class="section">
    <div class="wrap">
      <h2>Галерея идей</h2>
      <p class="muted" style="margin-bottom: 32px">Сохраняйте понравившиеся интерьеры</p>
      <p v-if="loading" class="muted">Загрузка…</p>
      <div v-else class="grid-2">
        <div v-for="r in items" :key="r.id" class="card gallery-card">
          <div class="gallery-img">
            <img :src="imgUrl(r.image)" :alt="r.title" />
          </div>
          <div class="gallery-body">
            <h3>{{ r.title }}</h3>
            <div class="tags">
              <span v-for="s in r.styles" :key="s" class="tag">{{ s }}</span>
            </div>
            <div class="price">от {{ formatPrice(r.price_from) }} до {{ formatPrice(r.price_to) }}</div>
            <RouterLink :to="`/calculator?ref=${r.id}`" class="btn ghost sm">Рассчитать</RouterLink>
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
.hero { padding: 80px 0 60px; background-size: cover; background-position: center; background-repeat: no-repeat; min-height: 420px; display: flex; align-items: center; color: #fff; }
.lead { font-size: 18px; margin: 24px 0 32px; max-width: 560px; color: #5f584f; }
.cta { display: flex; gap: 12px; flex-wrap: wrap; }
.gallery-card { display: flex; flex-direction: column; }
.gallery-img { aspect-ratio: 16/10; overflow: hidden; }
.gallery-img img { width: 100%; height: 100%; object-fit: cover; transition: transform 0.4s; }
.gallery-card:hover .gallery-img img { transform: scale(1.05); }
.gallery-body { padding: 20px; display: flex; flex-direction: column; gap: 10px; }
.tags { display: flex; gap: 6px; flex-wrap: wrap; }
.price { font-size: 15px; font-weight: 600; }
</style>
