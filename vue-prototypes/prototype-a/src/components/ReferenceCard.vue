<script setup lang="ts">
import type { ReferenceItem } from '@/types'
import { formatPrice } from '@/api/client'
import { imgUrl } from '@/composables/useImg'
import { refCategoryLabel, refStyleLabel } from '@/utils/labels'

defineProps<{ refItem: ReferenceItem; selected?: boolean }>()
</script>

<template>
  <div class="card ref-card" :class="{ selected }">
    <div class="ref-img">
      <img :src="imgUrl(refItem.image)" :alt="refItem.title" />
      <div class="ref-badge">{{ refCategoryLabel(refItem.category) }}</div>
    </div>
    <div class="ref-body">
      <h3>{{ refItem.title }}</h3>
      <div class="ref-tags">
        <span v-for="s in refItem.styles" :key="s" class="tag">{{ refStyleLabel(s) }}</span>
      </div>
      <div class="ref-price">
        от {{ formatPrice(refItem.price_from) }} до {{ formatPrice(refItem.price_to) }}
      </div>
      <RouterLink
        :to="`/estimate?ref=${refItem.id}`"
        class="btn red sm"
      >Рассчитать</RouterLink>
    </div>
  </div>
</template>

<style scoped>
.ref-card { transition: transform .2s, box-shadow .2s; background: #fff; border: 1px solid var(--line); border-radius: 14px; overflow: hidden; }
.ref-card:hover { transform: translateY(-4px); box-shadow: 0 12px 32px rgba(0,0,0,0.1); }
.ref-card.selected { border-color: #cf3d2b; box-shadow: 0 0 0 2px rgba(207,61,43,.35); }
.ref-img { position: relative; aspect-ratio: 4/3; overflow: hidden; }
.ref-img img { width: 100%; height: 100%; object-fit: cover; }
.ref-badge { position: absolute; top: 10px; left: 10px; background: #fff; padding: 4px 10px; border-radius: 6px; font-size: 11px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.1em; }
.ref-body { padding: 18px; display: flex; flex-direction: column; gap: 10px; }
.ref-body h3 { font-size: 20px; }
.ref-tags { display: flex; gap: 6px; flex-wrap: wrap; }
.ref-price { font-size: 14px; font-weight: 600; }
.btn.sm { min-height: 38px; padding: 8px 16px; font-size: 13px; align-self: flex-start; }

/* Широкие экраны: карточка — горизонтальная строка (фото слева, описание справа) */
@media (min-width: 900px) {
  .ref-card { display: grid; grid-template-columns: minmax(260px, 42%) 1fr; align-items: stretch; }
  .ref-img { aspect-ratio: auto; height: 100%; min-height: 220px; }
  .ref-body { padding: 22px 24px; justify-content: center; gap: 12px; }
}
</style>
