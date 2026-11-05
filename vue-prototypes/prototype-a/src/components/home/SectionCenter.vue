<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, API_BASE } from '@/api/client'
import { closeOnBackdrop } from '@/composables/useDialog'
const apiImg = (name: string) => `${API_BASE}/api/images/file/${name}`

interface CenterPhoto { name: string; src: string }

const photos = ref<CenterPhoto[]>([])
onMounted(async () => {
  try {
    const { data } = await api.get<CenterPhoto[]>('/api/center/photos')
    photos.value = data
  } catch { /* без фото галерея не откроется */ }
})

// В оригинале коллаж ссылался на галерею из 29 кадров; реальных фото в датасете 7,
// поэтому точки коллажа ведут на ближайшие по смыслу кадры
const shots = [
  { img: 'a_main_img_0029.jpg', label: 'Мебель и готовые сочетания', photo: 1, alt: 'Мебельная экспозиция Центра — реальное фото Центра' },
  { img: 'a_about_img_0003.jpg', label: 'Кухни', photo: 3, alt: 'Кухня и обеденная зона — реальное фото Центра' },
  { img: 'a_main_img_0031.jpg', label: 'Камень и фактуры', photo: 5, alt: 'Образцы камня для столешниц — реальное фото Центра' },
  { img: 'a_main_img_0032.jpg', label: 'Цвета', photo: 2, alt: 'Образцы цветов и покрытий — реальное фото Центра' },
  { img: 'a_catalog_img_0084.jpg', label: 'Отделочные материалы', photo: 6, alt: 'Образцы отделочных материалов — реальное фото Центра' },
]

const gallery = ref<HTMLDialogElement | null>(null)
const pos = ref(0)
function show(i: number) {
  if (!photos.value.length) return
  pos.value = (i + photos.value.length) % photos.value.length
  gallery.value?.showModal()
}
function close() { gallery.value?.close() }
</script>

<template>
  <section class="center section" id="center" data-block="Центр и компания">
    <div class="wrap">
      <div class="center29-head">
        <div>
          <span class="eyebrow red">
            <a class="route-number" href="#step-02" aria-label="К шагу 02 в схеме работы">02</a> / Центр интерьерных решений</span>
          <h2>За один визит — решения для вашей квартиры.</h2>
        </div>
        <p>К встрече подготовим подборку под понравившиеся интерьеры или ваш дизайн-проект.
          В Центре посмотрите мебель и материалы вживую, сравните варианты с экспертами и определитесь с выбором.</p>
      </div>
      <div class="center-stage63">
        <button class="center-shot shot-0 center-panorama63"
                :aria-label="'Открыть фото: ' + shots[0].alt" @click="show(shots[0].photo)">
          <img :src="apiImg(shots[0].img)" :alt="shots[0].alt" loading="lazy">
          <span>{{ shots[0].label }}</span>
        </button>
        <div class="center29-result">
          <b>После визита у вас</b>
          <ul>
            <li>Подборка решений под вашу квартиру</li>
            <li>Предварительная смета</li>
            <li>Согласованный следующий шаг</li>
          </ul>
          <p>Понимаете, что выбрано, сколько это стоит и как перейти к реализации.</p>
        </div>
      </div>
      <div class="center-after63">
        <div class="center29-copy">
          <span class="eyebrow red">Во время визита</span>
          <h3>Посмотрите, сравните и выберите вживую.</h3>
          <p class="center-visit61">Пройдём по подготовленной подборке, проверим сочетания мебели и материалов,
            обсудим технологию ремонта и бюджет вашей квартиры.</p>
          <div class="center-team61">
            <b>Поможем определиться</b>
            <p>Персональный менеджер подготовит подборку и будет сопровождать проект.
              Эксперты по мебели, материалам и технологии ремонта помогут с выбором.</p>
          </div>
          <RouterLink class="link center30-more" to="/center">Подробнее о Центре ↗</RouterLink>
        </div>
        <div class="center-views63">
          <div class="center-filmstrip63">
            <button v-for="(s, i) in shots.slice(1)" :key="s.label" :class="`center-shot shot-${i + 1}`"
                    :aria-label="'Открыть фото: ' + s.alt" @click="show(s.photo)">
              <img :src="apiImg(s.img)" :alt="s.alt" loading="lazy">
              <span>{{ s.label }}</span>
            </button>
          </div>
          <div class="center29-caption">
            <span>Реальные фотографии Центра ZETTA HOME</span>
            <button class="link" @click="show(0)">Все фото · {{ photos.length || '…' }} ↗</button>
          </div>
        </div>
      </div>
      <div class="center29-visit">
        <div>
          <strong>Почему по записи?</strong>
          <p>Заранее изучим вашу задачу и подготовим подборку.
            Персональный менеджер выделит время для работы именно с вашей квартирой.</p>
          <span>Москва · Аминьевское шоссе, 6 · ТЦ «Квартал», 2 этаж</span>
        </div>
        <RouterLink class="btn redbtn" to="/appointment">Записаться в Интерьерный центр ↗</RouterLink>
      </div>
    </div>

    <dialog ref="gallery" id="centerGallery" aria-labelledby="centerGalleryTitle" @click="closeOnBackdrop">
      <div class="center-gallery-head">
        <div>
          <strong id="centerGalleryTitle">Центр интерьерных решений</strong>
          <span id="centerPhotoCount" aria-live="polite">{{ photos.length ? (pos + 1) + ' / ' + photos.length : '' }}</span>
        </div>
        <button aria-label="Закрыть фотографии" @click="close">×</button>
      </div>
      <div class="center-gallery-stage">
        <img v-if="photos[pos]" :src="photos[pos].src" :alt="photos[pos].name">
      </div>
      <div class="center-gallery-foot">
        <button aria-label="Предыдущее фото" @click="show(pos - 1)">←</button>
        <p id="centerGalleryCaption" aria-live="polite">{{ photos[pos]?.name }}</p>
        <button aria-label="Следующее фото" @click="show(pos + 1)">→</button>
      </div>
    </dialog>
  </section>
</template>
