<script setup lang="ts">
/**
 * Товары интерьера — products(x) из оригинала.
 *
 * Показываем до восьми позиций с фотографией и ссылкой; позиции с
 * альтернативой выбранного варианта остаются в раскрывающейся смете.
 */
import { computed } from 'vue'
import type { InteriorItem } from '@/types'
import { useInteriorsStore } from '@/stores/interiors'
import { lineTotal, money } from '@/utils/interiors'
import { imgUrl } from '@/composables/useImg'

const props = defineProps<{ item: InteriorItem }>()

const interiors = useInteriorsStore()
const tier = computed(() => interiors.tier)

const products = computed(() =>
  (props.item.lines ?? [])
    .filter((l) => l.image && l.url && !(tier.value && l.altUrl))
    .slice(0, 8),
)
</script>

<template>
  <p v-if="!item.lines" class="muted">Товарная спецификация этого интерьера пока не добавлена
    в прототип. Здесь будут конкретные товары выбранного ценового варианта.</p>

  <template v-else>
    <div class="products realproducts">
      <article v-for="l in products" :key="l.name" class="product">
        <img :src="imgUrl(l.image!)" :alt="l.name" loading="lazy">
        <small>{{ l.name }}</small>
        <h3>{{ l.description || l.name }}</h3>
        <p>{{ l.qty }} · {{ money(lineTotal(l, tier)) }}</p>
        <a class="textlink" :href="l.url" target="_blank" rel="noopener">Подробнее о товаре ↗</a>
      </article>
    </div>
    <p class="scope">Полный перечень — в раскрывающейся смете. Изображения и ссылки взяты
      из исходной подборки. Позиции с альтернативой без фотографии показаны в смете.</p>
  </template>
</template>
