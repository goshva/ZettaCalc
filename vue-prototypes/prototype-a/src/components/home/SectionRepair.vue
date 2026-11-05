<script setup lang="ts">
import { ref } from 'vue'
import { API_BASE } from '@/api/client'
import { closeOnBackdrop } from '@/composables/useDialog'
const apiImg = (name: string) => `${API_BASE}/api/images/file/${name}`

const passport = ref<HTMLDialogElement | null>(null)

const compare = [
  ['Согласуете с каждым исполнителем состав и способы выполнения работ',
   'Подрядчики работают по выбранной стандартизированной технологии'],
  ['Сопоставляете сметы с разным наполнением',
   'Сравниваете предложения по единому цифровому паспорту ремонта'],
  ['Определяете, кому поручить весь ремонт или отдельные работы',
   'Выбираете компанию, которая отвечает за ремонт целиком'],
]

const bids = [
  { name: 'Компания А', old: '2,60 млн ₽', offer: '2,48 млн ₽', term: '5,5 месяца' },
  { name: 'Компания Б', old: '2,50 млн ₽', offer: '2,35 млн ₽', term: '5 месяцев' },
  { name: 'Компания В', old: '2,70 млн ₽', offer: '2,61 млн ₽', term: '4,5 месяца' },
]
</script>

<template>
  <section class="repair section" id="repair" data-block="Ремонт и исполнители">
    <div class="wrap">
      <div class="repair-intro63">
        <div>
          <span class="eyebrow">
            <a class="route-number" href="#step-04" aria-label="К шагу 04 в схеме работы">04</a> / Ремонт и исполнители</span>
          <h2>Одинаковая задача. Сопоставимые предложения.</h2>
          <p class="repair-intro">Подбираем технологию ремонта по вашему брифу и выбранным материалам.
            На основе планировки или замера готовим цифровой паспорт ремонта — единое задание,
            по которому подрядчики рассчитывают стоимость и выполняют работы.</p>
        </div>
        <figure class="repair-photo42">
          <img :src="apiImg('a_main_img_0034.jpg')"
               alt="Мастера проверяют стену уровнем и сверяются с планом квартиры" width="1536" height="1024">
          <figcaption>Сгенерированная иллюстрация</figcaption>
        </figure>
      </div>
      <div class="repair-system63">
        <div class="repair-rules63">
          <div class="repair-compare">
            <table>
              <thead>
                <tr>
                  <th scope="col">Самостоятельно</th>
                  <th scope="col">С ZETTA HOME</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="[a, b] in compare" :key="a">
                  <td>{{ a }}</td>
                  <td>{{ b }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div class="cpr-foot">
            <p><b>Цифровой паспорт можно заказать отдельно</b> и использовать с любым подрядчиком.
              При заказе ремонта через ZETTA HOME он входит в стоимость ремонта.</p>
            <button class="btn white" id="passportOpen" @click="passport?.showModal()">Что входит в ЦПР ↗</button>
          </div>
          <p class="service-note">Ремонт принимается по этапам. Дополнительно можно заказать индивидуальный
            дизайн-проект и «Адвоката качества». Состав сопровождения согласуем заранее.</p>
        </div>
        <div class="bids">
          <div class="bids-title">
            <h3>Аукцион подрядчиков</h3>
            <span class="tag">Пример механики</span>
          </div>
          <p class="auction-explain">Стоимость рассчитывается по единому заданию и прайсам на стандартные работы.
            На аукционе подрядчики могут предложить скидку.</p>
          <div class="doc-strip">
            <span class="doc-icon">ЦПР</span>
            <div>Цифровой паспорт ремонта<br>
              <span class="small">Одна технология и состав работ для всех участников</span>
            </div>
          </div>
          <div class="bid" v-for="b in bids" :key="b.name">
            <strong>{{ b.name }}</strong>
            <strong class="offer"><s>{{ b.old }}</s>{{ b.offer }}</strong>
            <small>{{ b.term }}</small>
            <small>Полный объём ремонта</small>
          </div>
          <p>Сравнивайте стоимость, сроки, рейтинг и отзывы. Все предложения — на полный объём ремонта.</p>
          <p>Условные компании, цены и сроки. Пример аукциона не связан с расчётом выше.</p>
        </div>
      </div>
    </div>

    <!-- ЦПР -->
    <dialog class="modal" ref="passport" id="passportDialog" aria-labelledby="passportTitle" @click="closeOnBackdrop">
      <button class="close" aria-label="Закрыть" @click="passport?.close()">×</button>
      <div class="modal-inner">
        <span class="eyebrow red">Цифровой паспорт ремонта</span>
        <h2 id="passportTitle">Что, где и как должно быть сделано.</h2>
        <ul class="cpr-list">
          <li>Данные планировки и замеров помещений.</li>
          <li>Согласованный план расстановки мебели.</li>
          <li>Выбранные материалы и технологии ремонта.</li>
          <li>Проекты электрики и сантехники.</li>
          <li>Состав работ и требования к результату.</li>
        </ul>
        <p>Паспорт можно заказать как отдельную платную услугу и передать любому подрядчику.
          Если заказываете ремонт через ZETTA HOME, паспорт входит в стоимость ремонта.</p>
      </div>
    </dialog>
  </section>
</template>
