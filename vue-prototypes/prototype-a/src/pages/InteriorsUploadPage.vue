<script setup lang="ts">
/**
 * Загрузка своего дизайн-проекта — перенос uploadPage() из оригинала.
 *
 * Файлы выбираются на устройстве и никуда не отправляются: прототип
 * проверяет размер (до 50 МБ на файл) и контакты, затем показывает
 * подтверждение со ссылкой на запись в Центр.
 */
import { computed, ref } from 'vue'
import { useAppointmentsStore } from '@/stores/appointments'
import { useAppStyles } from '@/composables/useAppStyles'

useAppStyles('interiors', 'inner-design65')

const MAX_FILE_MB = 50

const appts = useAppointmentsStore()

const files = ref<File[]>([])
const name = ref('')
const phone = ref('')
const comment = ref('')
const consent = ref(false)
const sent = ref(false)
const formError = ref('')

const fileNames = computed(() => files.value.map((f) => f.name).join(', '))
const tooLarge = computed(() => files.value.filter((f) => f.size > MAX_FILE_MB * 1024 * 1024))

function onFiles(e: Event) {
  const input = e.target as HTMLInputElement
  files.value = Array.from(input.files ?? [])
  formError.value = tooLarge.value.length ? `Выберите файлы до ${MAX_FILE_MB} МБ` : ''
}

function submit() {
  if (phone.value.replace(/\D/g, '').length < 10) {
    formError.value = 'Проверьте телефон'
    return
  }
  if (!consent.value) {
    formError.value = 'Нужно согласие на обработку персональных данных'
    return
  }
  formError.value = ''
  // Имена файлов переходят в форму записи: в оригинале её открывал data-action="book"
  appts.projectFiles = files.value.map((f) => f.name)
  sent.value = true
}
</script>

<template>
  <section class="section">
    <div class="wrap">
      <div class="crumb">
        <RouterLink to="/interiors">Интерьеры</RouterLink> / Свой дизайн-проект
      </div>

      <div class="uploadlayout">
        <div>
          <div class="eyebrow">Ваш проект — наши решения</div>
          <h1>Загрузите дизайн-проект до встречи</h1>
          <p class="lead">Изучим его, подготовим подборку мебели и материалов
            и предварительный расчёт.</p>
          <p>Приложите планировку, визуализации и спецификации, которые у вас есть.</p>
        </div>

        <form v-if="!sent" id="uploadform" class="filterbox" novalidate @submit.prevent="submit">
          <label>Файлы проекта
            <input type="file" name="project" multiple
                   accept=".pdf,.zip,.jpg,.jpeg,.png,.xlsx,.docx" @change="onFiles">
          </label>
          <p class="scope">PDF, ZIP, изображения или спецификации.
            До {{ MAX_FILE_MB }} МБ на файл в этом прототипе.</p>
          <div id="filelist">{{ fileNames }}</div>

          <label>Ваше имя
            <input v-model="name" name="name" autocomplete="given-name">
          </label>
          <label>Телефон
            <input v-model="phone" name="phone" type="tel" autocomplete="tel"
                   placeholder="+7 ___ ___-__-__">
          </label>
          <label>Что важно учесть?
            <textarea v-model="comment" name="comment" rows="3"></textarea>
          </label>

          <label class="consent">
            <input v-model="consent" type="checkbox">
            <span>Согласен с обработкой персональных данных</span>
          </label>

          <p class="formerror" role="alert">{{ formError }}</p>
          <button class="btn" type="submit">Отправить проект</button>
          <p class="protohelp">Демонстрация загрузки: файлы выбираются на устройстве,
            но никуда не отправляются.</p>
        </form>

        <div v-else class="filterbox">
          <h2>Проект выбран</h2>
          <p>В рабочей версии файлы и контакты поступят менеджеру для подготовки расчёта
            до встречи.</p>
          <p class="notice">Это демонстрация. Файлы и заявка не отправлены.</p>
          <RouterLink class="btn" to="/appointment">Выбрать время визита</RouterLink>
        </div>
      </div>
    </div>
  </section>
</template>
