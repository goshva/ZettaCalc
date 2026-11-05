<script setup lang="ts">
import { ref, computed } from 'vue'
import { closeOnBackdrop } from '@/composables/useDialog'
import SectionsIdeas from '@/components/home/SectionsIdeas.vue'
import SectionConnected from '@/components/home/SectionConnected.vue'
import SectionSupply from '@/components/home/SectionSupply.vue'
import SectionCenter from '@/components/home/SectionCenter.vue'
import SectionBudget from '@/components/home/SectionBudget.vue'
import SectionRepair from '@/components/home/SectionRepair.vue'
import SectionProject from '@/components/home/SectionProject.vue'
import SectionCompany from '@/components/home/SectionCompany.vue'
import SectionFaq from '@/components/home/SectionFaq.vue'
import SectionFinal from '@/components/home/SectionFinal.vue'

const emit = defineEmits<{ (e: 'meeting', topic: string, comment?: string): void }>()

// Универсальная модалка для вклада личного кабинета (как openModal в оригинале)
const generic = ref<HTMLDialogElement | null>(null)
const genericHtml = ref('')
function openGeneric(html: string) {
  genericHtml.value = html
  generic.value?.showModal()
}

// Модалка видео «Как это работает» (section_dialog_video в оригинале)
const videoDialog = ref<HTMLDialogElement | null>(null)
const systemVideo = ref<HTMLVideoElement | null>(null)
const videoProblem = ref(false)
const VIDEO_URL = 'https://zettahome.ru/upload/medialibrary/4ad/5wkjfrtwk4j7bx1jikggt8t27313qey4.mp4'
function openVideo() {
  videoProblem.value = false
  videoDialog.value?.showModal()
  const v = systemVideo.value
  if (!v) return
  if (!v.getAttribute('src')) v.src = VIDEO_URL
  v.play().catch(() => { videoProblem.value = true })
}
function closeVideo() {
  systemVideo.value?.pause()
  videoDialog.value?.close()
}

// Пути к картинкам через функцию — иначе Vite пытается разрешить их в сборке
import { API_BASE } from '@/api/client'
const apiImg = (name: string) => `${API_BASE}/api/images/file/${name}`

// ── «С чего начнём?» — тексты сценариев из оригинала (app.js scenarios) ──
const scenarios = [
  {
    tab: '01 / В ПОИСКЕ ИДЕИ', name: 'Начинаю с нуля',
    title: 'Начните с интерьера, в котором хочется жить.',
    text: 'Сохраните понравившиеся идеи. Свяжем их с планировкой, материалами и бюджетом вашей квартиры.',
    cta: 'Выбрать идеи', href: '#ideas',
  },
  {
    tab: '02 / ПРОЕКТ УЖЕ ЕСТЬ', name: 'Хочу его реализовать',
    title: 'Продолжим с вашего дизайн-проекта.',
    text: 'Подберём мебель и отделочные материалы под ваш дизайн-проект. Подготовим исходные данные для его реализации.',
    cta: 'Подобрать комплектацию', href: 'own',
  },
  {
    tab: '03 / НУЖНЫ РАБОТЫ', name: 'Ищу ремонт и смету',
    title: 'От предварительной сметы — к ремонту под контролем.',
    text: 'Подберём технологии ремонта, рассчитаем смету и уточним её после замера. Проведём аукцион подрядчиков — вы сравниваете предложения и выбираете исполнителя.',
    cta: 'Посмотреть путь ремонта', href: '#repair',
  },
  {
    tab: '04 / ВЫБИРАЮ НАПОЛНЕНИЕ', name: 'Нужна комплектация',
    title: 'Подберём всё необходимое для вашей квартиры.',
    text: 'Кухня, мебель, отделочные материалы и сантехника — под выбранный интерьер или ваш дизайн-проект. Все позиции соберём в единую комплектацию.',
    cta: 'Подобрать под мой проект', href: 'supply',
  },
]
const scenario = ref(0)
const active = computed(() => scenarios[scenario.value])

function onScenarioCta() {
  const s = active.value
  if (s.href === 'own') emit('meeting', 'У меня есть дизайн-проект')
  else if (s.href === 'supply') emit('meeting', 'Подобрать комплектацию')
  else window.location.hash = s.href
}

// ── «Шесть причин» — из оригинала (секция why) ──
const reasons = [
  { n: '01', title: 'Один проект вместо десятка подрядчиков', text: 'Дизайн, мебель, материалы и работы связаны в общий проект вашей квартиры.' },
  { n: '02', title: 'Всё можно увидеть вживую', text: 'В Интерьерном центре мебель и материалы можно посмотреть, сравнить и выбрать вместе с экспертами.' },
  { n: '03', title: 'Единое задание на ремонт', text: 'Цифровой паспорт ремонта фиксирует планировку, состав работ, технологии и требования к результату.' },
  { n: '04', title: 'Сопоставимые предложения', text: 'Подрядчики считают стоимость по одному заданию — на аукционе можно сравнить цену, сроки и рейтинг.' },
  { n: '05', title: 'Проект всегда перед глазами', text: 'Сроки, закупки, документы и приёмка этапов — в личном кабинете.' },
  { n: '06', title: 'За системой — опыт ZETTA', text: 'Более 25 лет в кухнях и корпусной мебели — и собственный Интерьерный центр в Москве.' },
]
</script>

<template>
  <div>
    <!-- 01 · Первый экран -->
    <section class="hero" data-block="Первый экран">
      <div class="wrap">
        <div class="hero-stage63 hero-photo hero-person">
          <div class="hero-wash63"></div>
          <div class="hero-copy">
            <span class="eyebrow">Дизайн · ремонт · мебель · материалы</span>
            <h1>
              <span class="hero-line62">Ваша квартира.</span>
              <span class="hero-line62">Все решения</span>
              <span class="red hero-line62">вместе.</span>
            </h1>
            <p>Подберём мебель и материалы, рассчитаем ремонт и организуем реализацию вашей квартиры. Персональный менеджер свяжет решения, закупки и работы в один проект — вы выбираете, насколько участвовать сами.</p>
            <div class="actions">
              <a class="btn redbtn" href="#ideas">Найти свой интерьер <span class="arrow">↗</span></a>
              <button class="btn light" @click="emit('meeting', 'У меня есть дизайн-проект')">У меня есть дизайн-проект <span class="arrow">↗</span></button>
            </div>
            <div class="hero-note">
              <a class="hero-video-link" href="#how"><span aria-hidden="true">▷</span> Как работает система</a>
            </div>
          </div>
          <div class="hero-photozone64">
            <div class="hero-frame64">
              <img class="person-photo hero-cover63" :src="apiImg('a_main_img_0001.jpg')" alt="Женщина с планшетом в светлой гостиной — иллюстрация выбора решений для дома">
              <RouterLink class="hero-person-caption hero-cta61" to="/appointment">Хочу всё вместе <span aria-hidden="true">↗</span></RouterLink>
              <svg class="tablet-connections" aria-hidden="true"><path /><path /><path /><path /></svg>
            </div>
            <div class="hero-parts hero-tags64" aria-label="Четыре части системы ZETTA HOME">
              <a class="part-card idea-part" href="#ideas">
                <div class="part-heading"><strong>Дизайн</strong><span aria-hidden="true">↗</span></div>
                <p>Выбираете будущий интерьер или загружаете свой дизайн-проект</p>
              </a>
              <a class="part-card supply-part" href="#supply">
                <div class="part-heading"><strong>Мебель и материалы</strong><span aria-hidden="true">↗</span></div>
                <p>Мебель и материалы подбираются точно под ваш проект</p>
              </a>
              <a class="part-card repair-part" href="#repair">
                <div class="part-heading"><strong>Ремонт</strong><span aria-hidden="true">↗</span></div>
                <p>Понятный состав работ, этапы и проверенные исполнители</p>
              </a>
              <a class="part-card manage-part" href="#project">
                <div class="part-heading"><strong>Управление</strong><span aria-hidden="true">↗</span></div>
                <p>Бюджет, сроки, решения и контроль — в одном месте</p>
              </a>
            </div>
          </div>
        </div>

        <!-- Три преимущества общей полосой -->
        <div class="hero-bottom">
          <div class="hero-benefit">
            <h3>Одна ответственная сторона</h3>
            <p>Персональный менеджер ведёт проект от идеи до приёмки — вместо десятка отдельных договоров.</p>
          </div>
          <div class="hero-benefit">
            <h3>Смета до начала работ</h3>
            <p>Предварительный расчёт и ЦПР появляются до закупки материалов и выбора подрядчика.</p>
          </div>
          <div class="hero-benefit">
            <h3>Решения можно увидеть</h3>
            <p>Мебель, материалы и свет собраны в Интерьерном центре — подборку смотрите вживую.</p>
          </div>
        </div>
      </div>
    </section>

    <!-- 02 · Сценарии входа -->
    <section class="scenario section" id="start" data-block="Сценарии входа">
      <div class="wrap">
        <div class="head"><h2>С чего начнём?</h2><p>Выберите вашу ситуацию.</p></div>
        <div class="scenario-tabs" role="tablist" aria-label="Ваша ситуация">
          <button
            v-for="(s, i) in scenarios" :key="s.tab"
            class="scenario-tab" role="tab"
            :aria-selected="scenario === i"
            @click="scenario = i"
          ><span>{{ s.tab }}</span><b>{{ s.name }}</b></button>
        </div>
        <div class="scenario-result" role="tabpanel" aria-live="polite">
          <div>
            <strong>{{ active.title }}</strong>
            <p>{{ active.text }}</p>
          </div>
          <a v-if="active.href.startsWith('#')" class="btn" :href="active.href">{{ active.cta }} <span class="arrow">↗</span></a>
          <button v-else class="btn" @click="onScenarioCta">{{ active.cta }} <span class="arrow">↗</span></button>
        </div>
      </div>
    </section>

    <!-- 03 · Как это работает -->
    <section class="how section" id="how" data-block="Как это работает">
      <div class="wrap">
        <div class="how-layout63">
          <div class="how-story63">
            <div class="intro">
              <span class="eyebrow">Как это работает</span>
              <h2>Система, а не набор отдельных подрядчиков.</h2>
              <p>Короткое видео о том, как из сохранённой идеи получается проект квартиры, а из проекта — готовый результат.</p>
            </div>
            <div class="video-frame">
              <img :src="apiImg('a_main_img_0002.jpg')" alt="Интерьерная иллюстрация к видео о системе ZETTA HOME">
              <span class="video-note">Видео о системе</span>
              <button class="play" id="videoOpen" aria-label="Смотреть видео «Как это работает»" @click="openVideo()">▶</button>
              <div class="video-meta"><strong>Как это работает</strong><span>Смотреть со звуком</span></div>
            </div>
          </div>
          <ol class="how-steps">
            <li id="step-01">
              <span class="step-number">01</span>
              <div>
                <b>Собираете идеи</b>
                <p>Сохраняете понравившиеся интерьеры и записываетесь в Центр. Можно прийти со своим дизайн-проектом.</p>
                <a class="step-destination" href="#ideas">Каталог интерьеров ↗</a>
              </div>
            </li>
            <li id="step-02">
              <span class="step-number">02</span>
              <div>
                <b>Подбираем решение в Центре</b>
                <p>К вашему приходу подберём мебель и отделочные материалы. В Центре посмотрите подборку вживую и определитесь с выбором.</p>
                <a class="step-destination" href="#center">Центр интерьерных решений ↗</a>
              </div>
            </li>
            <li id="step-03">
              <span class="step-number">03</span>
              <div>
                <b>Уточняем решения и бюджет</b>
                <p><span>Уточняем планировку и пожелания.</span> <span>Подбираем технологию ремонта.</span> <span>Согласуем решения и предварительную смету.</span></p>
                <a class="step-destination" href="#budget">Ориентир бюджета ↗</a>
              </div>
            </li>
            <li id="step-04">
              <span class="step-number">04</span>
              <div>
                <b>Реализуем и контролируем</b>
                <p>Выбираете подрядчика на аукционе. Следите за работами и принимаете этапы в личном кабинете.</p>
                <a class="step-destination" href="#repair">Ремонт и исполнители ↗</a>
              </div>
            </li>
          </ol>
        </div>
      </div>
    </section>

    <!-- 04 · Шесть причин -->
    <section class="why section" id="why" data-block="Шесть причин">
      <div class="wrap">
        <div class="head">
          <div>
            <span class="eyebrow">Почему ZETTA HOME</span>
            <h2>Шесть причин <em>начать с нас.</em></h2>
          </div>
          <p>Коротко о том, чем работа с системой отличается от ремонта «по частям».</p>
        </div>
        <div class="why-grid">
          <div class="why-item" v-for="r in reasons" :key="r.n">
            <span class="why-num">{{ r.n }}</span>
            <h3>{{ r.title }}</h3>
            <p>{{ r.text }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- 05 · Интерьерные идеи -->
    <SectionsIdeas @meeting="(topic, comment) => emit('meeting', topic, comment)" />

    <!-- 06 · Связь решений (хотспоты) -->
    <SectionConnected />

    <!-- 07 · Мебель и материалы -->
    <SectionSupply />

    <!-- 08 · Центр интерьерных решений -->
    <SectionCenter />

    <!-- 09 · Бюджет -->
    <SectionBudget />

    <!-- 10 · Ремонт и исполнители -->
    <SectionRepair />

    <!-- 11 · Личный кабинет -->
    <SectionProject @modal="openGeneric" />

    <!-- 12 · Компания и дополнительные услуги -->
    <SectionCompany @meeting="emit('meeting', $event)" />

    <!-- 13 · Вопросы -->
    <SectionFaq />

    <!-- 14 · Следующий шаг -->
    <SectionFinal @meeting="emit('meeting', $event)" />

    <!-- Диалог видео «Как это работает» -->
    <dialog class="modal video-modal" ref="videoDialog" id="videoDialog" aria-labelledby="videoTitle"
            @close="systemVideo?.pause()" @click="closeOnBackdrop">
      <button class="close" aria-label="Закрыть" @click="closeVideo()">×</button>
      <div class="modal-inner">
        <span class="eyebrow red">ZETTA HOME</span>
        <h2 id="videoTitle">Как это работает</h2>
        <video class="real-video" ref="systemVideo" controls playsinline preload="none"
               aria-label="Видео о системе ZETTA HOME" @error="videoProblem = true"></video>
        <div class="video-problem" v-show="videoProblem" role="status">Видео пока не загрузилось.
          Проверьте интернет или
          <a href="https://zettahome.ru/upload/medialibrary/4ad/5wkjfrtwk4j7bx1jikggt8t27313qey4.mp4" target="_blank" rel="noopener">откройте ролик отдельно</a>.
        </div>
        <p class="video-status">Выбор интерьера, комплектация и реализация — в одной системе.</p>
        <div class="actions">
          <button class="btn redbtn" @click="closeVideo(); emit('meeting', 'Получить консультацию')">Получить консультацию <span class="arrow">↗</span></button>
          <button class="btn light" @click="closeVideo()">Закрыть</button>
        </div>
      </div>
    </dialog>

    <!-- Универсальная модалка (вклад кабинета, категории) -->
    <dialog class="modal" ref="generic" @click="closeOnBackdrop">
      <button class="close" aria-label="Закрыть" @click="generic?.close()">×</button>
      <div class="modal-inner" v-html="genericHtml"></div>
    </dialog>
  </div>
</template>

<style scoped>
/* «Шесть причин» — секции нет в разметке главной А, стилизуем в палитре оригинала */
.why .head h2 em { font-style: italic; font-weight: 400; }
.why-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 1px; background: var(--line); border: 1px solid var(--line); border-radius: 14px; overflow: hidden; }
.why-item { background: var(--paper); padding: 30px 26px 34px; min-width: 0; }
.why-num { display: block; font-size: 13px; font-weight: 700; color: var(--red); letter-spacing: .1em; margin-bottom: 18px; }
.why-item h3 { font-size: 20px; line-height: 1.25; margin: 0 0 10px; letter-spacing: -.02em; }
.why-item p { font-size: 14px; line-height: 1.6; }
@media (max-width: 900px) { .why-grid { grid-template-columns: 1fr 1fr; } }
@media (max-width: 600px) { .why-grid { grid-template-columns: 1fr; } }
</style>
