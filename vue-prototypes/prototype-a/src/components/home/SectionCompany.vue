<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { API_BASE } from '@/api/client'

const emit = defineEmits<{ (e: 'meeting', topic: string): void }>()

const video = ref<HTMLVideoElement | null>(null)
const playBtn = ref<HTMLButtonElement | null>(null)
const videoError = ref(false)

// Логика как в initialize60: play по клику, пауза при уходе с экрана
function toggle() {
  const v = video.value
  if (!v) return
  videoError.value = false
  v.controls = true
  v.muted = false
  v.play().catch(() => { videoError.value = true })
}

// В оригинале видео ставится на паузу при уходе с экрана (IntersectionObserver)
let io: IntersectionObserver | undefined
onMounted(() => {
  if (!video.value || typeof IntersectionObserver === 'undefined') return
  io = new IntersectionObserver((entries) => {
    if (!entries[0].isIntersecting) video.value?.pause()
  }, { threshold: 0.1 })
  io.observe(video.value)
})
onBeforeUnmount(() => {
  io?.disconnect()
  video.value?.pause()
})
</script>

<template>
  <section id="companySection" class="section company41" data-block="Компания и дополнительные услуги">
    <div class="wrap">
      <div class="company-trust" id="company">
        <div class="company-copy">
          <span class="eyebrow red">Кто стоит за ZETTA HOME</span>
          <h2>За системой — опыт ZETTA.</h2>
          <div class="company-experience">
            <strong>25+</strong>
            <span>лет опыта в кухнях<br>и корпусной мебели</span>
          </div>
          <p>В основе ZETTA HOME — опыт компании ZETTA в кухнях и корпусной мебели.
            Развиваем его в систему, которая связывает идеи для интерьера, подбор мебели и материалов,
            ремонт и управление проектом квартиры.</p>
          <p>Этот опыт помогает смотреть на мебель и материалы как на части одного интерьера
            и подбирать их под вашу задачу.</p>
          <details class="company-more">
            <summary>Что это значит для вашего проекта</summary>
            <p>Начать можно с понравившегося интерьера или готового дизайн-проекта.
              Мебель и отделочные материалы подбираются под него, а состав решений, ориентир бюджета
              и дальнейшие шаги связываются в общий проект.</p>
            <p>За ремонтные работы отвечает выбранный исполнитель.
              Подбор, координация и контроль входят в проект в согласованном объёме.</p>
          </details>
          <RouterLink class="link company40-more" to="/about">О компании и системе ↗</RouterLink>
        </div>
        <div class="company-film">
          <article class="story60">
            <div class="story-media60">
              <video ref="video" playsinline preload="none" aria-label="О подходе ZETTA HOME"
                     :poster="`${API_BASE}/api/images/file/a_about_img_0002.jpg`"
                     @play="playBtn && (playBtn.hidden = true)"
                     @pause="playBtn && (playBtn.hidden = false)"
                     @ended="playBtn && (playBtn.hidden = false)"
                     @error="videoError = true">
                <source src="https://zetta.ru/upload/iblock/a23/0zhfh2bqne21x1wlepnvnyfjtbnoffq9/2026_10_07-11.26.56.mp4" type="video/mp4">
              </video>
              <button ref="playBtn" class="story-play60" aria-label="Смотреть видео: О подходе ZETTA HOME" @click="toggle">
                ▶<span>Смотреть · 10 сек</span>
              </button>
            </div>
            <div class="story-copy60">
              <p class="story-person60">Валерий Наврузбеков · руководитель ZETTA HOME</p>
              <p class="story-error60" v-show="videoError">Видео временно недоступно.
                <a href="https://zetta.ru/upload/iblock/a23/0zhfh2bqne21x1wlepnvnyfjtbnoffq9/2026_10_07-11.26.56.mp4"
                   target="_blank" rel="noopener">Открыть ролик ↗</a>
              </p>
            </div>
          </article>
        </div>
      </div>
      <aside class="extra-services" aria-labelledby="extraTitle">
        <div>
          <span class="eyebrow">Дополнительные платные услуги</span>
          <h3 id="extraTitle">Больше помощи — под вашу задачу.</h3>
          <p>Индивидуальный дизайн-проект, «Адвокат качества» и генподряд можно заказать отдельно.
            Состав, условия и стоимость согласуем до начала оказания услуг.</p>
        </div>
        <button class="btn" @click="emit('meeting', 'Дополнительные услуги')">Обсудить услуги <span class="arrow">↗</span></button>
      </aside>
    </div>
  </section>
</template>
