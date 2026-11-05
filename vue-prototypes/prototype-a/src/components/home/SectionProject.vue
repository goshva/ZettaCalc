<script setup lang="ts">
import { ref } from 'vue'

const emit = defineEmits<{ (e: 'modal', html: string): void }>()

const features = [
  { icon: '◷', title: 'Понятно, что и когда происходит', text: 'Этапы проекта, сроки закупок и ближайшие действия.' },
  { icon: '▤', title: 'Всё нужное под рукой', text: 'Документы, история заказов и фотографии каждого этапа ремонта.' },
  { icon: '✓', title: 'Согласования прямо в кабинете', text: 'Принимайте этапы ремонта и подписывайте документы онлайн.' },
]

const tabs = ['Ход проекта', 'Документы', 'Заказы']
const tab = ref(0)

// Содержимое вкладок — из views[] оригинального прототипа
function onAction(action: 'photos' | 'accept' | 'sign') {
  if (action === 'photos') {
    emit('modal', '<h3>Фотографии этапа</h3><p>Здесь клиент увидит фото, которые подрядчик загрузил после завершения работ.</p><div class="photo-empty">Место для фотофиксации ремонта</div><p class="notice">Реальные фотографии проекта в прототип не загружены.</p>')
  } else {
    const title = action === 'accept' ? 'Приёмка этапа' : 'Подписание документов'
    emit('modal', `<h3>${title}</h3><p>В личном кабинете клиент сможет изучить документы и фотографии, принять этап или отправить замечания подрядчику.</p><p class="notice">Показан сценарий будущего сервиса. Прототип ничего не принимает и не подписывает.</p>`)
  }
}
</script>

<template>
  <section id="project" data-block="Личный кабинет">
    <div class="wrap account-grid">
      <div class="account-copy">
        <span class="eyebrow">
          <a class="route-number" href="#step-04" aria-label="К шагу 04 в схеме работы">04</a> / Управление проектом</span>
        <h2>Личный кабинет. Весь проект перед вами.</h2>
        <p>Сроки, закупки, документы и ход ремонта — в одном месте.</p>
        <div class="features">
          <div class="feature" v-for="f in features" :key="f.title">
            <span class="icon" aria-hidden="true">{{ f.icon }}</span>
            <div>
              <strong>{{ f.title }}</strong>
              <p>{{ f.text }}</p>
            </div>
          </div>
        </div>
      </div>
      <div class="panel">
        <div class="panel-top">
          <div>
            <strong>Моя квартира</strong>
            <span class="demo">Пример будущего личного кабинета</span>
          </div>
          <span class="badge">Ремонт идёт</span>
        </div>
        <div class="tabs" role="tablist" aria-label="Личный кабинет">
          <button v-for="(t, i) in tabs" :key="t" role="tab" :id="'tab' + i"
                  aria-controls="tabbody" :aria-selected="tab === i"
                  :tabindex="tab === i ? 0 : -1" @click="tab = i">{{ t }}</button>
        </div>
        <div class="tabbody" id="tabbody" role="tabpanel" :aria-labelledby="'tab' + tab">
          <!-- 0 · Ход проекта -->
          <template v-if="tab === 0">
            <div class="step-line">
              <span>Подготовка ✓</span><span class="active">Ремонт</span><span>Комплектация</span>
            </div>
            <div class="task">
              <strong>Подготовка стен завершена</strong>
              <p>Подрядчик загрузил фотографии этапа. Посмотрите результат перед приёмкой.</p>
              <div class="task-bottom">
                <button class="text-link" @click="onAction('photos')">Посмотреть фото ↗</button>
                <button class="btn" @click="onAction('accept')">Принять этап</button>
              </div>
            </div>
            <div class="next">Далее · Согласовать дату доставки отделочных материалов</div>
          </template>
          <!-- 1 · Документы -->
          <template v-else-if="tab === 1">
            <div class="doc">
              <span>Цифровой паспорт ремонта<br><small>Согласован · пример</small></span><span>PDF</span>
            </div>
            <div class="doc">
              <span>Акт приёмки этапа<br><small>Ожидает вашего решения</small></span>
              <button class="text-link" @click="onAction('sign')">Открыть ↗</button>
            </div>
            <div class="doc">
              <span>Документы по поставке мебели<br><small>История документов проекта</small></span><span>PDF</span>
            </div>
          </template>
          <!-- 2 · Заказы -->
          <template v-else>
            <div class="doc">
              <span>Отделочные материалы<br><small>Следующая закупка по плану проекта</small></span>
              <span class="badge">Согласование</span>
            </div>
            <div class="doc">
              <span>Кухня<br><small>Срок поставки учтён в плане</small></span>
              <span class="badge">В производстве</span>
            </div>
            <div class="next">Здесь сохраняется история всех заказов по квартире.</div>
          </template>
        </div>
      </div>
    </div>
  </section>
</template>
