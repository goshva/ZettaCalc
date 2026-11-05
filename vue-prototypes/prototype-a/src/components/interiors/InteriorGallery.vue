<script setup lang="ts">
/**
 * Галерея карточки интерьера — перенос gallery(x) и обработчиков
 * prev / next / plan / zoom из оригинала.
 *
 * Состояние то же: galleryIndex и planMode. Планировка показывается
 * в той же рамке с классом .plan, увеличение — в диалоге #lightbox.
 */
import { computed, ref } from 'vue'
import type { InteriorItem } from '@/types'
import { imgUrl } from '@/composables/useImg'

const props = defineProps<{ item: InteriorItem }>()

const index = ref(0)
const planMode = ref(false)
const lightbox = ref<HTMLDialogElement | null>(null)

const gallery = computed(() => (props.item.gallery?.length ? props.item.gallery : [props.item.image]))
const src = computed(() =>
  planMode.value && props.item.plan ? props.item.plan : gallery.value[index.value],
)
const count = computed(() =>
  planMode.value ? 'Планировка' : `${index.value + 1} / ${gallery.value.length}`,
)

function step(delta: number) {
  planMode.value = false
  index.value = (index.value + delta + gallery.value.length) % gallery.value.length
}
</script>

<template>
  <div class="gallery">
    <img id="mainimage" :class="{ plan: planMode }" :src="imgUrl(src)" :alt="item.title">

    <div class="gallerybar">
      <div>
        <button type="button" aria-label="Предыдущее фото" @click="step(-1)">‹</button>
        <span id="gallerycount" class="small">{{ count }}</span>
        <button type="button" aria-label="Следующее фото" @click="step(1)">›</button>
      </div>
      <button v-if="item.plan" type="button" @click="planMode = !planMode">Планировка</button>
      <span v-else class="small muted">Визуализация интерьера</span>
      <button type="button" @click="lightbox?.showModal()">Увеличить</button>
    </div>

    <dialog id="lightbox" ref="lightbox" aria-label="Изображение интерьера">
      <div class="dialoghead">
        <span>{{ item.title }}</span>
        <button class="close" type="button" aria-label="Закрыть изображение" @click="lightbox?.close()">×</button>
      </div>
      <img style="width: 100%" :src="imgUrl(src)" :alt="item.title">
    </dialog>
  </div>
</template>
