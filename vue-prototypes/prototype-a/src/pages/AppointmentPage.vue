<script setup lang="ts">
/**
 * Запись в Интерьерный центр — итерация 4.
 *
 * Разметка и тексты перенесены из оригинального блока #appointment
 * (Прототип А, приложение «Центр»), поэтому стили приходят из
 * /api/apps/center/css и рассчитаны на селекторы под #surface52.
 *
 * Write-операция выполняется по правилам проекта: сначала dry-run,
 * затем создание записи с Idempotency-Key и подтверждением человека.
 */
import { ref, computed, onMounted, nextTick } from 'vue'
import { api, API_BASE } from '@/api/client'
import { bookingStatusLabel } from '@/utils/labels'
import { useAppointmentsStore } from '@/stores/appointments'
import { useFavoritesStore } from '@/stores/favorites'
import { useAppStyles } from '@/composables/useAppStyles'
import { closeOnBackdrop } from '@/composables/useDialog'
import type { AppointmentRequest, InteriorItem } from '@/types'

const appts = useAppointmentsStore()
const favorites = useFavoritesStore()
const { options, validation, created, replayed } = storeToRefs(appts)
const { ideas: savedIds, meeting: meetingIds } = storeToRefs(favorites)

// Страница записи — маршрут приложения «Центр» (#/center/appointment), поэтому
// оформление приходит из center.css, а CSS главной на это время выключается:
// иначе его `header`, `.btn` и `.grid` перебивают стили формы записи.
useAppStyles('center', 'inner-design65')

onBeforeMount(() => {
  // html.booking67 включает стили страницы записи в оригинальной таблице
  document.documentElement.classList.add('booking67')
})
onBeforeUnmount(() => {
  document.documentElement.classList.remove('booking67')
})

// Опции формы (задачи, время, адрес) приходят из API вместе с лимитами
onMounted(async () => {
  if (!options.value) appts.fetchOptions()
  // Выбор для встречи, сделанный в избранном, подхватывается формой — как в booking()
  if (meetingIds.value.length && !picked.value.length) {
    await ensureInteriors()
    picked.value = meetingIds.value
      .map((id) => allInteriors.value.find((i) => i.id === id))
      .filter(Boolean) as InteriorItem[]
  }
  // Файлы, выбранные на странице загрузки проекта, переносятся в заявку
  if (appts.projectFiles.length && !files.value.length) {
    files.value = appts.projectFiles.map((name) => ({ name, size: 0 }))
  }
})

const apiImg = (name: string) => `${API_BASE}/api/images/file/${name}`

// ── Состояние формы ────────────────────────────────────────────────────────
const TASK_FALLBACK = ['Подбор интерьера', 'Мебель и материалы', 'Ремонт', 'Свой дизайн-проект']
const tasks = computed(() => options.value?.tasks ?? TASK_FALLBACK)
const times = computed(() => options.value?.times ?? ['Уточнить по телефону'])
const maxIdeas = computed(() => options.value?.max_ideas ?? 5)
const maxFiles = computed(() => options.value?.max_files ?? 10)

const form = ref({
  tasks: [] as string[],
  name: '',
  phone: '',
  date: '',
  time: 'Уточнить по телефону',
  wish: '',
  consent: false,
})
const picked = ref<InteriorItem[]>([])
const files = ref<{ name: string; size: number }[]>([])
const fileError = ref('')

const today = new Date().toISOString().slice(0, 10)
const fileInput = ref<HTMLInputElement | null>(null)
const uploadOpen = ref(false)

const pickedIds = computed(() => picked.value.map((i) => i.id))
const totalFileSize = computed(() => files.value.reduce((sum, f) => sum + f.size, 0))

function toggleTask(t: string) {
  const i = form.value.tasks.indexOf(t)
  if (i >= 0) form.value.tasks.splice(i, 1)
  else form.value.tasks.push(t)
}

// ── Выбор идей из избранного ───────────────────────────────────────────────
const picker = ref<HTMLDialogElement | null>(null)
const allInteriors = ref<InteriorItem[]>([])
const savedInteriors = computed(() => {
  const byId = new Map(allInteriors.value.map((i) => [i.id, i]))
  return savedIds.value.map((id) => byId.get(id)).filter(Boolean) as InteriorItem[]
})

async function ensureInteriors() {
  if (allInteriors.value.length) return
  try {
    const { data } = await api.get<InteriorItem[]>('/api/interiors')
    allInteriors.value = data
  } catch {
    /* без данных каталога список останется пустым */
  }
}

// В диалоге выбор идёт в черновик: он применяется к записи по кнопке «Готово»
const pickerDraft = ref<string[]>([])

async function openPicker() {
  await ensureInteriors()
  pickerDraft.value = [...pickedIds.value]
  picker.value?.showModal()
}

function togglePick(item: InteriorItem) {
  const i = picked.value.findIndex((p) => p.id === item.id)
  if (i >= 0) picked.value.splice(i, 1)
  else if (picked.value.length < maxIdeas.value) picked.value.push(item)
  // Выбор сразу попадает в «Для встречи»: он сохраняется при переходе между страницами
  favorites.setMeeting(pickedIds.value)
}

function toggleDraft(item: InteriorItem) {
  const i = pickerDraft.value.indexOf(item.id)
  if (i >= 0) pickerDraft.value.splice(i, 1)
  else if (pickerDraft.value.length < maxIdeas.value) pickerDraft.value.push(item.id)
}

/** «Готово» в диалоге: черновик становится выбором для встречи */
function applyPicker() {
  const byId = new Map(allInteriors.value.map((i) => [i.id, i]))
  picked.value = pickerDraft.value
    .map((id) => byId.get(id))
    .filter(Boolean) as InteriorItem[]
  favorites.setMeeting(pickerDraft.value)
  picker.value?.close()
}

// ── Файлы дизайн-проекта ───────────────────────────────────────────────────
const MAX_FILE_MB = 20
const MAX_TOTAL_MB = 50

function onFiles(e: Event) {
  const input = e.target as HTMLInputElement
  fileError.value = ''
  const chosen = Array.from(input.files ?? [])
  const accepted: { name: string; size: number }[] = []
  for (const f of chosen) {
    if (files.value.length + accepted.length >= maxFiles.value) {
      fileError.value = `Не более ${maxFiles.value} файлов`
      break
    }
    if (f.size > MAX_FILE_MB * 1024 * 1024) {
      fileError.value = `Файл «${f.name}» больше ${MAX_FILE_MB} МБ`
      continue
    }
    accepted.push({ name: f.name, size: f.size })
  }
  const next = [...files.value, ...accepted]
  if (next.reduce((s, f) => s + f.size, 0) > MAX_TOTAL_MB * 1024 * 1024) {
    fileError.value = `Суммарный размер больше ${MAX_TOTAL_MB} МБ`
  } else {
    files.value = next
  }
  input.value = ''
}

function removeFile(index: number) {
  files.value.splice(index, 1)
  fileError.value = ''
}

const sizeLabel = (bytes: number) =>
  bytes >= 1024 * 1024
    ? (bytes / 1024 / 1024).toFixed(1).replace('.', ',') + ' МБ'
    : Math.max(1, Math.round(bytes / 1024)) + ' КБ'

// ── Проверка и запись ──────────────────────────────────────────────────────
function payload(): AppointmentRequest {
  return {
    name: form.value.name.trim(),
    phone: form.value.phone.trim(),
    tasks: form.value.tasks,
    preferred_date: form.value.date || null,
    preferred_time: form.value.time,
    ideas: pickedIds.value,
    files: files.value.map((f) => f.name),
    wish: form.value.wish.trim(),
    consent: form.value.consent,
  }
}

const localErrors = computed(() => {
  const errors: string[] = []
  if (!form.value.name.trim()) errors.push('Укажите имя')
  if (!/^[+0-9\s()\-]{10,22}$/.test(form.value.phone.trim())) {
    errors.push('Телефон: 10–22 символа, цифры, +, скобки и дефисы')
  }
  if (!form.value.consent) errors.push('Нужно согласие на обработку персональных данных')
  if (form.value.date && form.value.date < today) errors.push('Дата визита уже прошла')
  return errors
})

// Шаг 1 — dry-run: проверяем форму на сервере, ничего не создавая
async function check() {
  if (localErrors.value.length) {
    appts.validation = null
    appts.error = localErrors.value.join('; ')
    return
  }
  await appts.dryRun(payload())
}

// Шаг 2 — создание записи с Idempotency-Key и подтверждением
async function confirm() {
  await appts.submit(payload(), true)
}

function again() {
  appts.reset()
  picked.value = []
  files.value = []
  fileError.value = ''
  uploadOpen.value = false
  form.value = {
    tasks: [], name: '', phone: '', date: '', time: times.value[0], wish: '', consent: false,
  }
}
</script>

<template>
  <div class="wrap">
    <div class="booking-head67">
      <div class="crumb">
        <RouterLink to="/">Интерьерный центр</RouterLink> / Запись на встречу
      </div>
      <div class="eyebrow">От понравившихся идей — к решениям для квартиры</div>
      <h1>Встреча, к которой<br>мы подготовимся</h1>
      <p>Выберите удобное время и расскажите о своей задаче. Можно приложить интерьеры,
        загрузить свой проект или записаться без предварительного выбора.</p>
    </div>

    <section class="section" id="appointment" data-block="Запись в Центр">
      <div class="visit">
      <!-- Форма -->
      <div class="formbox">
        <div class="why-book60">
          <h3>Почему по записи?</h3>
          <p>Заранее изучим ваши идеи или дизайн-проект и подготовим подборку.
            Персональный менеджер выделит время для работы именно с вашей квартирой.</p>
        </div>

        <template v-if="!created">
          <h3>Записаться в Интерьерный центр</h3>
          <p class="intro">Оставьте контакты — менеджер подтвердит время. Регистрация не нужна.</p>

          <form id="visit-form" novalidate @submit.prevent="check">
            <fieldset class="task67">
              <legend>Что хотите обсудить?</legend>
              <div class="task-options67">
                <label v-for="t in tasks" :key="t">
                  <input type="checkbox" :value="t" :checked="form.tasks.includes(t)" @change="toggleTask(t)">
                  {{ t }}
                </label>
              </div>
            </fieldset>

            <div class="fields">
              <label>Ваше имя
                <input v-model="form.name" name="name" required maxlength="80" autocomplete="given-name">
              </label>
              <label>Телефон
                <input v-model="form.phone" name="phone" required type="tel"
                       placeholder="+7 ___ ___-__-__" autocomplete="tel">
              </label>
              <label>Удобная дата <span class="muted">· необязательно</span>
                <input v-model="form.date" name="date" type="date" :min="today">
              </label>
              <label>Удобное время
                <select v-model="form.time" name="time">
                  <option v-for="t in times" :key="t">{{ t }}</option>
                </select>
              </label>
            </div>

            <div class="prepare">
              <h4>Подготовимся к нашей встрече</h4>
              <p>Выберите до пяти интерьеров из избранного или приложите свой проект.
                Готовые сметы и ваши изменения обсудим на встрече. Можно записаться без выбора.</p>
              <div class="prepare-links">
                <button class="btn light" type="button" @click="openPicker">
                  {{ picked.length ? `Изменить выбор · ${picked.length} из ${maxIdeas}` : 'Выбрать из избранного' }}
                </button>
                <button class="btn light" type="button" :aria-expanded="uploadOpen"
                        @click="uploadOpen = !uploadOpen">Добавить дизайн-проект</button>
              </div>

              <div class="selected">
                <div v-if="picked.length" class="selected-items">
                  <div v-for="idea in picked" :key="idea.id" class="selected-card67">
                    <img :src="idea.image" :alt="idea.title">
                    <span>{{ idea.title }}</span>
                    <button type="button" :aria-label="'Убрать ' + idea.title"
                            @click="togglePick(idea)">×</button>
                  </div>
                </div>
                <p v-else class="small">Интерьеры для подборки пока не выбраны.</p>
              </div>

              <div class="upload-area" v-show="uploadOpen">
                <p>PDF, JPG, PNG, ZIP, DOC/DOCX. До {{ maxFiles }} файлов, до {{ MAX_FILE_MB }} МБ каждый
                  и до {{ MAX_TOTAL_MB }} МБ суммарно. В демоверсии файлы никуда не отправляются —
                  передаются только имена.</p>
                <input ref="fileInput" type="file" multiple aria-label="Файлы дизайн-проекта"
                       accept=".pdf,.jpg,.jpeg,.png,.zip,.doc,.docx" @change="onFiles">
                <div class="file-list">
                  <div v-for="(f, i) in files" :key="f.name + i" class="file-row">
                    <span>{{ f.name }} · {{ sizeLabel(f.size) }}</span>
                    <button type="button" :aria-label="'Удалить файл ' + f.name"
                            @click="removeFile(i)">×</button>
                  </div>
                  <p v-if="files.length" class="small">Всего: {{ sizeLabel(totalFileSize) }}</p>
                </div>
                <p id="file-error67" role="alert" class="formerror">{{ fileError }}</p>
              </div>
            </div>

            <label>Пожелания <span class="muted">· необязательно</span>
              <textarea v-model="form.wish" name="wish" rows="4" maxlength="1200"
                        placeholder="Помещения, пожелания, ориентир по бюджету — если уже определились"></textarea>
            </label>

            <label class="consent">
              <input v-model="form.consent" type="checkbox" name="consent" required>
              <span>Согласен/согласна с
                <a :href="options?.consent_url ?? 'https://zettahome.ru/privacy-policy/consent/'"
                   target="_blank" rel="noopener">обработкой персональных данных</a>.</span>
            </label>

            <p class="formerror" role="alert">{{ appts.error }}</p>
            <p v-if="validation && !appts.error" class="small" role="status">
              Проверка пройдена: {{ validation.summary.join(' · ') }}
            </p>
            <p v-for="w in validation?.warnings ?? []" :key="w" class="small" role="status">{{ w }}</p>

            <button v-if="!validation?.ok" class="btn redbtn submit" type="submit" :disabled="appts.checking">
              {{ appts.checking ? 'Проверяем…' : 'Проверить сценарий записи' }} <span class="arrow">↗</span>
            </button>
            <button v-else class="btn redbtn submit" type="button" :disabled="appts.saving" @click="confirm">
              {{ appts.saving ? 'Записываем…' : 'Подтвердить запись в Центр' }} <span class="arrow">↗</span>
            </button>

            <p class="demo">Демонстрационная форма. Заявка попадает в mock API с ключом идемпотентности
              и подтверждением; файлы не загружаются. Для проверки используйте вымышленные контакты.</p>
          </form>
        </template>

        <!-- Подтверждение -->
        <div v-else class="success" role="status">
          <div class="eyebrow">Пример подтверждения</div>
          <h3>Запись создана</h3>
          <p>Заявка <b>{{ created.id }}</b> сохранена со статусом «{{ bookingStatusLabel(created.status) }}»{{ replayed ? ' (повторная отправка того же запроса)' : '' }}.
            В рабочей версии менеджер подтвердит время встречи и получит выбранные интерьеры, файлы и пожелания.</p>
          <ul>
            <li>Задачи: {{ created.tasks.join(', ') || 'не выбраны' }}</li>
            <li v-if="created.preferred_date">Дата: {{ created.preferred_date }}</li>
            <li>Время: {{ created.preferred_time }}</li>
            <li v-if="created.ideas.length">Интерьеров к подборке: {{ created.ideas.length }}</li>
            <li v-if="created.files.length">Файлов дизайн-проекта: {{ created.files.length }}</li>
            <li>Адрес: {{ created.place }}</li>
          </ul>
          <p class="demo">Это mock-запись: контакты и файлы никуда не передаются.</p>
          <button class="btn light" type="button" @click="again">Вернуться к форме</button>
        </div>
      </div>

      <!-- Адрес и помощник -->
      <aside class="booking-side67">
        <div class="helper-invite67">
          <div class="eyebrow">Помощь с выбором</div>
          <h3>С чего начать именно вам?</h3>
          <p>Помощник уточнит задачу, предложит направления и соберёт пожелания для встречи.</p>
          <RouterLink class="btn" to="/">К подборке интерьеров ↗</RouterLink>
          <p class="small">Необязательно для записи. В прототипе — готовые сценарии, без подключения AI.</p>
        </div>

        <div class="visit-address66">
          <div class="eyebrow">Запись и маршрут</div>
          <h2>Приходите посмотреть решения вживую</h2>
          <div class="address">
            <strong>{{ options?.place.address ?? 'Москва, Аминьевское шоссе, 6' }}</strong>
            <p>{{ options?.place.mall ?? 'ТЦ «Квартал», 2 этаж · бесплатная парковка' }}</p>
            <p><a :href="options?.place.phone_href ?? 'tel:+74951387338'">{{ options?.place.phone ?? '+7 (495) 138-73-38' }}</a></p>
          </div>
          <div class="directions">
            <a class="btn light" :href="options?.place.map" target="_blank" rel="noopener">Построить маршрут ↗</a>
            <a class="textlink" :href="options?.place.phone_href ?? 'tel:+74951387338'">Позвонить в Центр</a>
          </div>
          <p class="small" style="margin-top:18px">{{ options?.place.note }}</p>
          <div class="entry">
            <img :src="apiImg('a_main_img_0056.jpg')" alt="Вход в Центр интерьерных решений ZETTA HOME">
            <p class="small">Вход в Центр — ориентируйтесь на вывеску ZETTA HOME.</p>
          </div>
        </div>
      </aside>
      </div>
    </section>

    <!-- Выбор интерьеров для подборки -->
    <dialog ref="picker" id="picker" aria-labelledby="picker-title" @click="closeOnBackdrop">
      <div class="dialoghead">
        <h3 id="picker-title">Какие интерьеры обсудим?</h3>
        <button class="close" aria-label="Закрыть" @click="picker?.close()">×</button>
      </div>
      <p class="small">Избранное может содержать сколько угодно интерьеров.
        Для подготовки к визиту выберите до пяти.</p>
      <div class="picker-grid">
        <button v-for="item in savedInteriors" :key="item.id" type="button" class="picker-card"
                :aria-pressed="pickerDraft.includes(item.id)" @click="toggleDraft(item)">
          <img :src="item.image" :alt="item.title">
          <span>{{ pickerDraft.includes(item.id) ? '✓ ' : '' }}{{ item.room }} · {{ item.title }}</span>
        </button>
        <p v-if="!savedInteriors.length" class="empty">
          В избранном пока пусто. Сохраните интерьеры кнопкой ♡ в разделе «Дизайн».
        </p>
      </div>
      <div class="pickerbottom">
        <p aria-live="polite">Выбрано: {{ pickerDraft.length }} из {{ maxIdeas }}</p>
        <button class="btn" type="button" @click="applyPicker">Готово</button>
      </div>
    </dialog>
  </div>
</template>

<style scoped>
/* Остальное оформление приходит из /api/apps/center/css — стили формы записи.
   Ниже только то, чего нет в оригинальной таблице стилей. */
.selected-items { display: grid; gap: 9px; }
.file-row { display: flex; align-items: center; justify-content: space-between; gap: 10px; }
.file-row button { border: 0; background: none; font-size: 16px; cursor: pointer; }
.picker-card { border: 1px solid var(--line); border-radius: 10px; padding: 0; background: #fff; overflow: hidden; text-align: left; cursor: pointer; }
.picker-card[aria-pressed='true'] { border-color: var(--red); box-shadow: 0 0 0 1px var(--red); }
.picker-card img { width: 100%; aspect-ratio: 4/3; object-fit: cover; }
.picker-card span { display: block; padding: 10px; font-size: 13px; }
</style>
