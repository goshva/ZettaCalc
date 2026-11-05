<script setup lang="ts">
/**
 * Панель планировки — planpanel из detail(x) и apartment(x).
 *
 * Для помещений, показанных только планом, панель заменяет галерею:
 * отдельной визуализации в исходных файлах нет.
 */
import { computed, ref } from 'vue'
import type { InteriorItem } from '@/types'
import { imgUrl } from '@/composables/useImg'

const props = defineProps<{
  item: InteriorItem
  /** Заголовок панели: планировка квартиры или помещения */
  title?: string
  /** Текст-пояснение под изображением */
  note: string
  /** Кнопка увеличения плана */
  zoomable?: boolean
}>()

const lightbox = ref<HTMLDialogElement | null>(null)
const plan = computed(() => props.item.plan ?? props.item.image)
</script>

<template>
  <div class="planpanel">
    <h3>{{ title ?? 'Планировка квартиры' }}</h3>
    <img :src="imgUrl(plan)" :alt="`Исходный план: ${item.title}`">
    <p class="scope">{{ note }}</p>
    <button v-if="zoomable" class="btn secondary" type="button" @click="lightbox?.showModal()">
      Увеличить план
    </button>

    <dialog id="lightbox" ref="lightbox" aria-label="Планировка">
      <div class="dialoghead">
        <h3>Планировка</h3>
        <button class="close" type="button" aria-label="Закрыть план" @click="lightbox?.close()">×</button>
      </div>
      <img style="width: 100%" :src="imgUrl(plan)" :alt="`План: ${item.title}`">
    </dialog>
  </div>
</template>
