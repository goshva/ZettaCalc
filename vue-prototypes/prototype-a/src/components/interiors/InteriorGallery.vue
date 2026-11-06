<script setup lang="ts">
/**
 * Галерея карточки интерьера — перенос gallery(x) и обработчиков
 * prev / next / plan из оригинала + миниатюры в формате .thumbs.
 *
 * Состояние то же: galleryIndex и planMode. Планировка показывается
 * в той же рамке с классом .plan. Отдельной кнопки «Увеличить» нет:
 * lightbox открывается двойным кликом по изображению, но только когда
 * увеличение реально возможно (см. canZoom).
 */
import { computed, ref } from 'vue'
import type { InteriorItem } from '@/types'
import { imgUrl } from '@/composables/useImg'

const props = defineProps<{ item: InteriorItem }>()

const index = ref(0)
const planMode = ref(false)
const mainimage = ref<HTMLImageElement | null>(null)
const lightbox = ref<HTMLDialogElement | null>(null)

const gallery = computed(() => (props.item.gallery?.length ? props.item.gallery : [props.item.image]))
/** Карусель (‹ 1/N › и миниатюры) имеет смысл только при нескольких фото. */
const multiple = computed(() => gallery.value.length > 1)
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

/**
 * Увеличение допустимо, если есть изображение большего размера (можно
 * показать в оригинале) ИЛИ текущее сжато вёрсткой: naturalWidth больше
 * отображаемого clientWidth. В каталоге версий «большого» размера нет,
 * поэтому критерий проверяется по реально загруженному пиксельному размеру.
 */
const canZoom = computed(() => {
  const el = mainimage.value
  if (!el || !el.naturalWidth || !el.clientWidth) return false
  return el.naturalWidth > el.clientWidth + 1 || el.naturalHeight > el.clientHeight + 1
})

function onDblclick() {
  if (!canZoom.value) return
  lightbox.value?.showModal()
}
</script>

<template>
  <div class="gallery">
    <img
      id="mainimage"
      ref="mainimage"
      :class="{ plan: planMode, zoomable: canZoom }"
      :src="imgUrl(src)"
      :alt="item.title"
      @dblclick="onDblclick"
    >

    <div v-if="multiple || item.plan" class="gallerybar">
      <div v-if="multiple">
        <button type="button" aria-label="Предыдущее фото" @click="step(-1)">‹</button>
        <span id="gallerycount" class="small">{{ count }}</span>
        <button type="button" aria-label="Следующее фото" @click="step(1)">›</button>
      </div>
      <span v-else class="small muted">Визуализация интерьера</span>
      <button v-if="item.plan" type="button" @click="planMode = !planMode">Планировка</button>
    </div>

    <!-- Миниатюры как в каталоге Прототипа А: .thumbs + data-a="photo". -->
    <div v-if="multiple" class="thumbs">
      <button
        v-for="(g, i) in gallery"
        :key="g"
        type="button"
        data-a="photo"
        :data-index="i"
        :aria-label="`Фото ${i + 1}`"
        :aria-current="!planMode && i === index"
        @click="index = i"
      >
        <img :src="imgUrl(g)" :alt="`Ракурс ${i + 1}`">
      </button>
    </div>

    <dialog id="lightbox" ref="lightbox" class="lightbox" aria-label="Изображение интерьера">
      <div class="dialoghead">
        <span>{{ item.title }}</span>
        <button class="close" type="button" aria-label="Закрыть изображение" @click="lightbox?.close()">×</button>
      </div>
      <img style="width: 100%" :src="imgUrl(src)" :alt="item.title">
    </dialog>
  </div>
</template>

<style scoped>
/* Ряд галереи: счётчик/стрелки слева, «Планировка» справа. */
.gallery > .gallerybar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-top: 12px;
}
.gallery > .gallerybar > div {
  display: flex;
  align-items: center;
  gap: 10px;
}
.gallery > .gallerybar button {
  border: 1px solid var(--line, #dddfd7);
  background: transparent;
  border-radius: 50%;
  width: 34px;
  height: 34px;
  cursor: pointer;
  font-size: 16px;
  line-height: 1;
}
/* Кнопка «Планировка» текстовая — не круглая. */
.gallery > .gallerybar button:last-child {
  border-radius: 8px;
  width: auto;
  height: auto;
  padding: 8px 14px;
  font-size: 13px;
}
#mainimage.zoomable {
  cursor: zoom-in;
}

/* Миниатюры — формат .thumbs из каталога Прототипа А. */
.thumbs {
  display: flex;
  gap: 10px;
  margin: 13px 0 0;
  overflow-x: auto;
  scrollbar-width: thin;
  padding-bottom: 4px;
}
.thumbs button {
  flex: 0 0 auto;
  width: 86px;
  padding: 4px;
  border: 1px solid var(--line, #dddfd7);
  border-radius: 7px;
  background: transparent;
  cursor: pointer;
}
.thumbs button[aria-current='true'] {
  border: 2px solid var(--red, #cf3d2b);
  padding: 3px;
}
.thumbs img {
  height: 60px;
  width: 100%;
  object-fit: contain;
  display: block;
}
</style>
