<script setup lang="ts">
import { ref } from 'vue'
import { API_BASE } from '@/api/client'
import { closeOnBackdrop } from '@/composables/useDialog'
const apiImg = (name: string) => `${API_BASE}/api/images/file/${name}`

const emit = defineEmits<{ (e: 'meeting', topic: string): void }>()

const categories = [
  {
    img: 'a_main_img_0025.jpg',
    alt: 'Прямой диван Creatica Djun — изображение из каталога ZETTA HOME',
    imgStyle: 'object-fit: contain; background: rgb(244, 244, 244);',
    title: 'Мебель',
    text: 'Диваны, кровати, кухни, системы хранения, столы и стулья.',
    cta: 'Смотреть мебель',
  },
  {
    img: 'a_catalog_img_0087.jpg',
    alt: 'Мебель и декор в экспозиции Центра ZETTA HOME',
    title: 'Аксессуары и декор',
    text: 'Зеркала, текстиль и детали, которые завершают интерьер.',
    cta: 'Смотреть аксессуары',
  },
  {
    img: 'a_main_img_0027.jpg',
    alt: 'Керамогранит TAU Ceramica DORSET в интерьере — изображение из каталога ZETTA HOME',
    title: 'Отделочные материалы',
    text: 'Пол, стены, плитка и покрытия — по стилю и техническим требованиям.',
    cta: 'Смотреть материалы',
  },
  {
    img: 'a_main_img_0028.jpg',
    alt: 'Ванна Ravak YPSILON WALL — изображение из каталога ZETTA HOME',
    title: 'Сантехника и ванная',
    text: 'Сантехника, мебель, зеркала и комплектующие в едином решении.',
    cta: 'Смотреть сантехнику',
  },
]

// Модальное окно категории / формы подбора (как openModal в оригинале)
type ModalState = null | { kind: 'category'; title: string } | { kind: 'help' } | { kind: 'sent' }
const modal = ref<ModalState>(null)
const modalEl = ref<HTMLDialogElement | null>(null)
const form = ref({ name: '', phone: '', comment: '' })

function open(m: ModalState) {
  modal.value = m
  modalEl.value?.showModal()
}
function close() {
  modal.value = null
  modalEl.value?.close()
}
function send() {
  modal.value = { kind: 'sent' }
  form.value = { name: '', phone: '', comment: '' }
}
</script>

<template>
  <section class="section" id="supply" data-block="Мебель и материалы">
    <div class="wrap">
      <div class="head">
        <div>
          <span class="eyebrow">
            <a class="route-number" href="#step-02" aria-label="К шагу 02 в схеме работы">02</a> / Мебель и материалы</span>
          <h2>Мебель, материалы и аксессуары.</h2>
        </div>
        <p>Выбирайте отдельную категорию или собирайте весь интерьер в одной логике.</p>
      </div>
      <div class="grid catalog-cards64">
        <article class="category" v-for="c in categories" :key="c.title">
          <button class="image-link" :aria-label="c.cta" @click="open({ kind: 'category', title: c.title })">
            <img :src="`/api/images/file/${c.img}`" :alt="c.alt" loading="lazy" :style="c.imgStyle">
          </button>
          <div class="card-copy">
            <h3>{{ c.title }}</h3>
            <p>{{ c.text }}</p>
            <button class="text-link" @click="open({ kind: 'category', title: c.title })">{{ c.cta }} <span>↗</span></button>
          </div>
        </article>
      </div>
      <div class="help">
        <div>
          <h3>Хотите собрать всё под ваш интерьер?</h3>
          <p>Подберём мебель, кухню, отделочные материалы и сантехнику под выбранную концепцию или ваш дизайн-проект.</p>
        </div>
        <button class="btn" id="help" @click="open({ kind: 'help' })">Помочь с подбором ↗</button>
      </div>
    </div>

    <dialog ref="modalEl" id="modal" class="modal" @click="closeOnBackdrop">
      <button class="close" aria-label="Закрыть" @click="close">×</button>
      <div id="modalBody" class="modal-inner">
        <template v-if="modal?.kind === 'category'">
          <h3>{{ modal.title }}</h3>
          <p>На сайте здесь откроется соответствующий раздел каталога: товары, фотографии и характеристики.</p>
          <p class="notice">В этом прототипе показана точка перехода. Страницы каталога будем проектировать отдельно.</p>
        </template>
        <template v-else-if="modal?.kind === 'help'">
          <h3>Подбор комплектации</h3>
          <p>Оставьте контакты — мы свяжемся с вами, уточним задачу и подскажем, с чего начать.</p>
          <form id="request" class="field-grid" @submit.prevent="send">
            <label>Ваше имя<input v-model="form.name" name="name" required autocomplete="given-name"></label>
            <label>Телефон<input v-model="form.phone" name="phone" type="tel" required autocomplete="tel"></label>
            <label>Что хотите подобрать? <span class="demo">Необязательно</span>
              <textarea v-model="form.comment" name="comment" rows="2"></textarea>
            </label>
            <button class="btn" type="submit">Посмотреть подтверждение</button>
            <p class="notice">Демонстрация формы. Контакты не сохраняются и не отправляются.</p>
          </form>
        </template>
        <template v-else-if="modal?.kind === 'sent'">
          <h3>Так будет выглядеть подтверждение</h3>
          <p>Спасибо! Мы свяжемся с вами, чтобы уточнить задачу по комплектации.</p>
          <p class="notice">Это пример. Заявка не отправлена.</p>
        </template>
      </div>
    </dialog>
  </section>
</template>

<style scoped>
/* Форма модалки — как стили #request в оригинале */
#request { display: grid; gap: 12px; }
#request label { display: grid; gap: 6px; font-size: 13px; font-weight: 600; }
#request input, #request textarea { border: 1px solid var(--line); border-radius: 8px; padding: 10px 12px; font: inherit; background: #fff; }
#request .demo { color: var(--muted); font-weight: 400; }
</style>
