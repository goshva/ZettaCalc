<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { closeOnBackdrop } from '@/composables/useDialog'

// Детерминированная модель расчёта — константы дословно из оригинального прототипа
const segments = ['Базовый', 'Оптимальный', 'Комфортный'] as const
const rates = [
  [17000, 19000, 16500],
  [22500, 26500, 24000],
  [29000, 35000, 34000],
]
const fixed = [
  { work: 80000, material: 90000, furniture: 130000 },
  { work: 110000, material: 120000, furniture: 200000 },
  { work: 150000, material: 150000, furniture: 280000 },
  { work: 210000, material: 190000, furniture: 390000 },
  { work: 270000, material: 230000, furniture: 500000 },
]
const labels = ['Работы', 'Отделочные материалы', 'Мебель']
const barColors = ['#353b38', '#e44832', '#b0b798']

const roomsOptions = [
  { value: 0, label: 'Студия' },
  { value: 1, label: '1 комната' },
  { value: 2, label: '2 комнаты' },
  { value: 3, label: '3 комнаты' },
  { value: 4, label: '4 и более' },
]
const scopeOptions = [
  { value: 'all', label: 'Работы, мебель и отделочные материалы' },
  { value: 'repair', label: 'Ремонтные работы' },
  { value: 'supply', label: 'Мебель и отделочные материалы' },
]

const form = ref({ area: 54, rooms: 2, condition: 'new', scope: 'all' })
const level = ref(0)

const rub = (n: number) => new Intl.NumberFormat('ru-RU').format(n) + ' ₽'

interface Part { label: string; value: number; color: string }
const parts = ref<Part[]>([])
const total = ref(0)
const meta = ref('')
const planNote = ref('')
const price = ref('')
const valid = ref(true)

function calc(): boolean {
  const { area, rooms, condition, scope } = form.value
  if (!area || area < 20 || area > 250) { valid.value = false; return false }
  valid.value = true
  const f = fixed[rooms]
  const mul = [1, 1.35, 1.8][level.value]
  const cond = condition === 'old' ? 1.15 : 1
  let p = [
    (area * rates[level.value][0] + f.work * mul) * cond,
    area * rates[level.value][1] + f.material * mul,
    area * rates[level.value][2] + f.furniture * mul,
  ].map((x) => Math.round(x / 1000) * 1000)
  if (scope === 'repair') p = [p[0], 0, 0]
  if (scope === 'supply') p[0] = 0
  total.value = p.reduce((a, b) => a + b, 0)
  price.value = 'от ' + rub(total.value)
  const roomsLabel = roomsOptions[rooms].label
  meta.value = `${area} м² · ${roomsLabel} · ${segments[level.value]}`
  planNote.value = 'Модель расчёта: ' +
    (rooms === 0 ? 'студия с кухонной зоной' : roomsLabel.toLowerCase() + ' и отдельная кухня') + '.'
  parts.value = p
    .map((v, i) => ({ label: labels[i], value: v, color: barColors[i] }))
    .filter((x) => x.value > 0)
  return true
}

onMounted(calc)

// Тост «Ориентир пересчитан» — как в оригинале
const toast = ref('')
let toastTimer: ReturnType<typeof setTimeout> | undefined
function submit() {
  if (calc()) {
    toast.value = 'Ориентир пересчитан'
    clearTimeout(toastTimer)
    toastTimer = setTimeout(() => (toast.value = ''), 2000)
  }
}

// ── Диалоги ──
const apartments = ref<HTMLDialogElement | null>(null)
const meeting = ref<HTMLDialogElement | null>(null)
const meetingContext = ref('')
const meetingSent = ref(false)

function openApartments() { apartments.value?.showModal() }

function openMeeting() {
  // В оригинале: сначала валидация формы, затем контекст в модалку
  if (!calc()) return
  const scopeLabel = scopeOptions.find((o) => o.value === form.value.scope)?.label ?? ''
  meetingContext.value = `${meta.value} · ${scopeLabel} · демонстрационный ориентир ${rub(total.value)}`
  meetingSent.value = false
  meeting.value?.showModal()
}
function sendMeeting() {
  meetingSent.value = true
}
</script>

<template>
  <section class="budget section" id="budget" data-block="Бюджет">
    <div class="wrap">
      <div class="head">
        <div>
          <span class="eyebrow red">
            <a class="route-number" href="#step-03" aria-label="К шагу 03 в схеме работы">03</a> / Бюджет до первого разговора</span>
          <h2>Поймите порядок цифр. И что за ними стоит.</h2>
        </div>
        <p>Укажите площадь и количество комнат. Сравните три ценовых сегмента и состав расходов.</p>
      </div>
      <div class="budget-grid">
        <form class="budget-form" id="budgetForm" @submit.prevent="submit">
          <div class="form-pair">
            <div class="field">
              <label for="area">Площадь квартиры, м²</label>
              <input class="input" id="area" type="number" min="20" max="250" step="1" v-model.number="form.area" @change="calc" required>
            </div>
            <div class="field">
              <label for="rooms">Количество комнат</label>
              <select class="input" id="rooms" v-model.number="form.rooms" @change="calc">
                <option v-for="r in roomsOptions" :key="r.value" :value="r.value">{{ r.label }}</option>
              </select>
            </div>
          </div>
          <div class="field">
            <label for="condition">Состояние квартиры</label>
            <select class="input" id="condition" v-model="form.condition" @change="calc">
              <option value="new">Новостройка без отделки</option>
              <option value="old">Вторичное жильё</option>
            </select>
          </div>
          <div class="field">
            <label for="scope">Что нужно рассчитать</label>
            <select class="input" id="scope" v-model="form.scope" @change="calc">
              <option value="all">Работы, мебель и отделочные материалы</option>
              <option value="repair">Ремонтные работы</option>
              <option value="supply">Мебель и отделочные материалы</option>
            </select>
          </div>
          <button class="btn" type="submit">Пересчитать ориентир <span class="arrow">↗</span></button>
          <p class="small" style="margin-top:16px">Без регистрации и номера телефона.</p>
        </form>
        <div class="budget-result">
          <div class="levels" role="group" aria-label="Ценовой сегмент">
            <button v-for="(s, i) in segments" :key="s" class="level" type="button"
                    :aria-pressed="level === i" @click="level = i; calc()">{{ s }}</button>
          </div>
          <span class="tag">Демонстрационный расчёт</span>
          <div id="budgetLive" aria-live="polite">
            <div class="price" id="price">{{ price }}</div>
            <p class="budget-meta" id="budgetMeta">{{ meta }}</p>
            <p class="plan-note" id="planNote">{{ planNote }}</p>
            <div class="bars" id="bars" aria-hidden="true">
              <i v-for="p in parts" :key="p.label"
                 :style="{ width: (p.value / total * 100) + '%', background: p.color }"></i>
            </div>
            <div class="breakdown" id="breakdown">
              <div v-for="p in parts" :key="p.label">
                <span>{{ p.label }}</span><b>{{ rub(p.value) }}</b>
              </div>
            </div>
          </div>
          <p class="budget-assumption">За основу взята типовая планировка для выбранной площади и количества комнат.
            Расчёт включает выбранные категории по минимальным ценам в выбранном сегменте.
            Состав и стоимость уточняются под вашу квартиру.</p>
          <p class="demo-caption">В прототипе показаны условные суммы. Реальные цены на работы и позиции каталога
            ещё не подключены; это не смета и не предложение цены.</p>
          <div class="result-actions">
            <button class="btn redbtn" id="apartmentsOpen" @click="openApartments">Смотреть квартиры со сметами <span class="arrow">↗</span></button>
            <button class="link" id="budgetMeetingOpen" @click="openMeeting">Уточнить бюджет в Центре ↗</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Квартиры со сметами -->
    <dialog class="modal" ref="apartments" id="apartmentsDialog" aria-labelledby="apartmentsTitle" @click="closeOnBackdrop">
      <button class="close" aria-label="Закрыть" @click="apartments?.close()">×</button>
      <div class="modal-inner">
        <span class="eyebrow red">Переход из блока бюджета</span>
        <h2 id="apartmentsTitle">Квартиры со сметами</h2>
        <p>Здесь будет переход на отдельную страницу с конкретными планировками, интерьерными решениями и комплектацией.</p>
        <div class="demo-tiles"><span>Базовый</span><span>Оптимальный</span><span>Комфортный</span></div>
        <p>Для каждой квартиры — три ценовых сегмента и разбивка на работы, мебель и отделочные материалы.</p>
        <div class="notice">Страница ещё не подключена к этому прототипу.
          Здесь показано назначение кнопки, без вымышленных квартир и смет.</div>
      </div>
    </dialog>

    <!-- Уточнить бюджет в Центре -->
    <dialog class="modal" ref="meeting" id="budgetMeetingDialog" aria-labelledby="budgetMeetingTitle" @click="closeOnBackdrop">
      <button class="close" aria-label="Закрыть" @click="meeting?.close()">×</button>
      <div class="modal-inner">
        <span class="eyebrow red">Центр интерьерных решений</span>
        <h2 id="budgetMeetingTitle">Уточним бюджет вашей квартиры.</h2>
        <template v-if="!meetingSent">
          <p>Оставьте контакты, чтобы договориться о встрече в Центре и обсудить состав решения.</p>
          <form @submit.prevent="sendMeeting">
            <p class="notice" id="budgetMeetingContext">{{ meetingContext }}</p>
            <div class="field">
              <label for="budgetName">Как к вам обращаться</label>
              <input class="input" id="budgetName" autocomplete="given-name" required maxlength="80">
            </div>
            <div class="field">
              <label for="budgetPhone">Телефон</label>
              <input class="input" id="budgetPhone" type="tel" autocomplete="tel" required>
            </div>
            <p class="demo-caption">Демонстрационная форма: данные не отправляются и не сохраняются.
              Можно использовать вымышленные контакты.</p>
            <button class="btn redbtn" type="submit">Записаться в Интерьерный центр ↗</button>
          </form>
        </template>
        <div v-else>
          <p>Сценарий записи завершён. Заявка не отправлена.</p>
          <button class="btn" @click="meeting?.close()">Готово</button>
        </div>
      </div>
    </dialog>

    <div class="toast" role="status" aria-live="polite" :class="{ show: !!toast }">{{ toast }}</div>
  </section>
</template>

<style scoped>
/* Тост позиционируется как #toast в оригинале */
.toast { position: fixed; left: 50%; bottom: 28px; transform: translate(-50%, 20px); background: var(--ink); color: #fff; border-radius: 10px; padding: 12px 18px; font-size: 14px; opacity: 0; pointer-events: none; transition: .25s; z-index: 60; }
.toast.show { opacity: 1; transform: translate(-50%, 0); }
</style>
