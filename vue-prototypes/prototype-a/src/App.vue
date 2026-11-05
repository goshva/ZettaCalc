<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { closeOnBackdrop } from '@/composables/useDialog'
import AppHeader from '@/components/AppHeader.vue'
import AppFooter from '@/components/AppFooter.vue'
import ShellHeader from '@/components/shell/ShellHeader.vue'
import InteriorsFooter from '@/components/shell/InteriorsFooter.vue'
import CenterFooter from '@/components/shell/CenterFooter.vue'

const auth = useAuthStore()
onMounted(() => {
  if (auth.isAuthenticated) auth.fetchMe()
})

// В оригинале оболочка постоянна, а содержимое #surface52 подменяется целиком
// (mount52). Шапка главной лежит внутри #surface52, поэтому на страницах
// приложений её скрывает `#surface52>header{display:none!important}`, а видна
// общая шапка shell-header. Футер приложений, наоборот, свой.
//
// Страница записи — это маршрут приложения «Центр» (#/center/appointment),
// поэтому она тоже считается страницей приложения.
const route = useRoute()
const isInteriors = computed(() => route.path.startsWith('/interiors'))
const isApp = computed(() => isInteriors.value || route.path.startsWith('/appointment'))

// Модальное окно записи — открывается любой кнопкой «Получить консультацию» / «data-meeting».
// Логика заголовков и полей — как meeting(task) в оригинале
const meetingTopic = ref('')
const meetingEl = ref<HTMLDialogElement | null>(null)
const meetingForm = ref({ task: 'Получить консультацию', name: '', phone: '', need: '', comment: '' })
const meetingSent = ref(false)

const meetingTitle = computed(() => {
  const t = meetingTopic.value
  if (t === 'Дополнительные услуги') return 'Обсудим дополнительные услуги'
  if (own.value) return 'Продолжим с вашего дизайн-проекта'
  if (t === 'Подобрать комплектацию') return 'Подберём комплектацию под ваш интерьер'
  if (t === 'Получить консультацию') return 'Поможем разобраться, с чего начать.'
  return 'Подготовим встречу под вашу задачу.'
})
const meetingIntro = computed(() => {
  const t = meetingTopic.value
  if (t === 'Дополнительные услуги') return 'Оставьте контакты и укажите интересующую услугу. Обсудим задачу, состав и стоимость.'
  if (own.value) return 'Оставьте контакты. Мы свяжемся с вами, уточним задачу и договоримся, как передать дизайн-проект для обсуждения комплектации и ремонта.'
  if (t === 'Подобрать комплектацию') return 'Оставьте контакты. Уточним, какой интерьер вы выбрали или есть ли у вас дизайн-проект, и обсудим подбор мебели и материалов для вашей квартиры.'
  return 'Расскажите, с чего хотите начать.'
})
const own = computed(() => meetingTopic.value === 'У меня есть дизайн-проект')
const submitLabel = computed(() =>
  (own.value || meetingTopic.value === 'Подобрать комплектацию' || meetingTopic.value === 'Дополнительные услуги')
    ? 'Связаться со мной' : 'Проверить сценарий записи')

function openMeeting(topic: string, comment?: string) {
  meetingTopic.value = topic
  const opts = ['Получить консультацию', 'Дополнительные услуги', 'Подобрать комплектацию', 'Встреча в Центре',
    'У меня есть дизайн-проект', 'Ремонт и смета', 'Мебель и интерьер', 'Кухня', 'Материалы',
    'Сантехника и ванная', 'Уточнить бюджет квартиры', 'Обсудить мой проект']
  meetingForm.value = { task: opts.includes(topic) ? topic : 'Обсудить мой проект', name: '', phone: '', need: '', comment: comment ?? '' }
  meetingSent.value = false
  meetingEl.value?.showModal()
}

function closeMeeting() {
  meetingEl.value?.close()
}

function sendMeeting() {
  // TODO (итерация 4): POST на API с dry-run
  meetingSent.value = true
}
</script>

<template>
  <!-- #surface52 обязателен: весь оригинальный CSS привязан к этому id.
       Оболочка лежит вне него — как в оригинале, где shell-header не входит
       в подменяемый блок и переживает переход между приложениями. -->
  <div class="app">
    <ShellHeader v-if="isApp" @appointment="openMeeting('Встреча в Центре')" />

    <div id="surface52">
      <!-- обёртка нужна: оригинал скрывает #surface52>header как прямой ребёнок -->
      <div class="site-head" v-if="!isApp">
        <AppHeader @meeting="openMeeting" />
      </div>
      <main>
        <RouterView @meeting="openMeeting" />
      </main>
      <!-- Футер приложения свой и лежит внутри #surface52 -->
      <InteriorsFooter v-if="isInteriors" @appointment="openMeeting('Встреча в Центре')" />
      <CenterFooter v-else-if="isApp" />
      <AppFooter v-else @appointment="openMeeting('Встреча в Центре')" />

      <dialog class="modal" id="meetingDialog" ref="meetingEl" aria-labelledby="meetingTitle"
              @click="closeOnBackdrop">
        <button class="close" aria-label="Закрыть" @click="closeMeeting">✕</button>
        <div class="modal-inner">
          <template v-if="!meetingSent">
          <span class="eyebrow red">Следующий шаг</span>
          <h2 id="meetingTitle">{{ meetingTitle }}</h2>
          <p>{{ meetingIntro }}</p>
          <form @submit.prevent="sendMeeting">
            <div class="field" v-if="!own">
              <label for="meetingTask">Ваша задача</label>
              <select class="input" id="meetingTask" v-model="meetingForm.task">
                <option>Получить консультацию</option>
                <option>Дополнительные услуги</option>
                <option>Подобрать комплектацию</option>
                <option>Встреча в Центре</option>
                <option>У меня есть дизайн-проект</option>
                <option>Ремонт и смета</option>
                <option>Мебель и интерьер</option>
                <option>Кухня</option>
                <option>Материалы</option>
                <option>Сантехника и ванная</option>
                <option>Уточнить бюджет квартиры</option>
                <option>Обсудить мой проект</option>
              </select>
            </div>
            <div class="field" v-if="own">
              <label for="meetingNeed">Что вам нужно <span class="muted">· необязательно</span></label>
              <select class="input" id="meetingNeed" v-model="meetingForm.need">
                <option value="">Выберите задачу</option>
                <option>Комплектация</option>
                <option>Ремонт</option>
                <option>Всё вместе</option>
              </select>
            </div>
            <div class="field">
              <label for="meetingName">Как к вам обращаться</label>
              <input class="input" id="meetingName" v-model="meetingForm.name" placeholder="Имя" required maxlength="80" autocomplete="given-name" />
            </div>
            <div class="field">
              <label for="meetingPhone">Телефон</label>
              <input class="input" id="meetingPhone" v-model="meetingForm.phone" type="tel" placeholder="+7 ___ ___-__-__" required autocomplete="tel" />
            </div>
            <div class="field" v-if="!own">
              <label for="meetingComment">Что важно учесть <span class="muted">· необязательно</span></label>
              <textarea class="input" id="meetingComment" v-model="meetingForm.comment" placeholder="Площадь, сроки, что уже готово…" maxlength="1200"></textarea>
            </div>
            <p class="notice">Демонстрационная форма. Данные никуда не отправляются и не сохраняются. Для проверки можно использовать вымышленные имя и телефон.</p>
            <button class="btn redbtn" type="submit">{{ submitLabel }} <span class="arrow">↗</span></button>
          </form>
        </template>
        <template v-else>
          <div class="notice">Сценарий заполнения завершён. Заявка не отправлена.</div>
          <p>В рабочем сервисе здесь появятся подтверждение запроса и информация о следующем контакте.</p>
          <button class="btn" @click="closeMeeting">Готово</button>
        </template>
      </div>
    </dialog>
    </div>
  </div>
</template>

<style scoped>
.app { display: flex; flex-direction: column; min-height: 100vh; }
#surface52 { display: flex; flex-direction: column; flex: 1; }
main { flex: 1; }
.modal { border: 0; border-radius: 16px; padding: 0; max-width: 560px; width: calc(100% - 32px); background: var(--paper); color: var(--ink); }
.modal::backdrop { background: #25282499; }
.modal .close { position: absolute; top: 14px; right: 14px; border: 0; background: none; font-size: 20px; cursor: pointer; }
.modal-inner { padding: 34px 32px 30px; }
.modal-inner h2 { font-size: 30px; margin: 10px 0 12px; }
.modal-inner p { margin-bottom: 22px; }
.modal-inner form { display: grid; gap: 14px; }
.field label { display: block; font-size: 12px; font-weight: 700; letter-spacing: .08em; text-transform: uppercase; margin-bottom: 7px; }
.input { width: 100%; min-height: 46px; border: 1px solid var(--line); background: #fff; border-radius: 8px; padding: 11px 13px; font: inherit; }
</style>
