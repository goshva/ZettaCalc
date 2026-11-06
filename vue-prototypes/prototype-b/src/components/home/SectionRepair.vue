<script setup lang="ts">
/**
 * «Ремонт и исполнители» + аукцион подрядчиков — секция 10 главной прототипа В.
 *
 * Разметка повторяет эталон (b_page/14_section_repair.html): .sec.dark → .grid12 →
 * .cmp для сравнения и панель ставок .bid. В эталоне ставки строились из массива
 * BIDS, вторая была помечена .on сразу, а по клику выбор подтверждался toast() —
 * здесь то же поведение средствами Vue.
 *
 * Стили заданы локально: палитра прототипа В в этом приложении (--ink/--gold/
 * --paper из main.css), поэтому эталонные классы переопределены под неё, чтобы
 * секция не выпадала из остальных страниц.
 */
import { ref } from 'vue'
import { imgUrl } from '@/composables/useImg'

const compare = [
  ['Согласуете с каждым исполнителем состав и способы выполнения работ',
   'Подрядчики работают по выбранной стандартизированной технологии'],
  ['Сопоставляете сметы с разным наполнением',
   'Сравниваете предложения по единому цифровому паспорту ремонта'],
  ['Определяете, кому поручить весь ремонт или отдельные работы',
   'Выбираете компанию, которая отвечает за ремонт целиком'],
]

// Условные данные: цены и сроки не связаны с расчётом выше.
const bids = [
  { id: 'А', name: 'Компания А', old: '2,60 млн ₽', offer: '2,48 млн ₽', term: '5,5 месяца' },
  { id: 'Б', name: 'Компания Б', old: '2,50 млн ₽', offer: '2,35 млн ₽', term: '5 месяцев' },
  { id: 'В', name: 'Компания В', old: '2,70 млн ₽', offer: '2,61 млн ₽', term: '4,5 месяца' },
]

// В эталоне B вторая ставка помечена .on ещё до клика (`i===1`); текст выбора
// показывается только через toast.
const chosen = ref<string>('Б')
const toastText = ref('')
let toastTimer: ReturnType<typeof setTimeout> | undefined

function choose(bid: (typeof bids)[number]) {
  chosen.value = bid.id
  toastText.value = `Выбрано предложение: ${bid.name}`
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toastText.value = '' }, 2600)
}

const cpr = ref<HTMLDialogElement | null>(null)
</script>

<template>
  <section class="repair sec dark" id="repair">
    <div class="wrap">
      <div class="grid12">
        <div class="intro">
          <div class="eyebrow"><b>10</b><i></i>Ремонт и исполнители</div>
          <h2 class="h2">Одинаковая задача.<br>
            <span class="it">Сопоставимые предложения.</span></h2>
          <p class="lead">
            Подбираем технологию ремонта по вашему брифу и выбранным материалам. На основе планировки
            или замера готовим цифровой паспорт ремонта — единое задание, по которому подрядчики
            рассчитывают стоимость и выполняют работы.
          </p>
        </div>
        <figure class="shot">
          <div class="shot-media">
            <img class="photo" :src="imgUrl('repair')" alt="Мастера выполняют отделочные работы в квартире" />
          </div>
          <figcaption class="proto-note">Сгенерированная иллюстрация</figcaption>
        </figure>
      </div>

      <div class="grid12 body">
        <div class="rules">
          <table class="cmp">
            <thead><tr><th>Самостоятельно</th><th>С ZETTA HOME</th></tr></thead>
            <tbody>
              <tr v-for="[a, b] in compare" :key="a"><td>{{ a }}</td><td>{{ b }}</td></tr>
            </tbody>
          </table>
          <div class="cpr-box">
            <p><b>Цифровой паспорт можно заказать отдельно</b> и использовать с любым подрядчиком.
              При заказе ремонта через ZETTA HOME он входит в стоимость ремонта.</p>
            <button class="btn light sm" type="button" @click="cpr?.showModal()">Что входит в ЦПР ↗</button>
          </div>
          <p class="proto-note note">
            Ремонт принимается по этапам. Дополнительно можно заказать индивидуальный дизайн-проект
            и «Адвоката качества». Состав сопровождения согласуем заранее.
          </p>
        </div>

        <div class="bids">
          <div class="bids-head">
            <h3 class="h3">Аукцион подрядчиков</h3>
            <span class="tag">Пример механики</span>
          </div>
          <p class="small">
            Стоимость рассчитывается по единому заданию и прайсам на стандартные работы.
            На аукционе подрядчики могут предложить скидку.
          </p>
          <div class="doc-strip">
            <span class="doc-icon">ЦПР</span>
            <span><b>Цифровой паспорт ремонта</b><br>
              <span class="muted">Одна технология и состав работ для всех участников</span></span>
          </div>

          <div class="bid-list">
            <button
              v-for="b in bids" :key="b.id"
              type="button"
              :class="['bid', { on: chosen === b.id }]"
              :aria-pressed="chosen === b.id"
              @click="choose(b)"
            >
              <span class="serif name">{{ b.name }}</span>
              <span class="r"><s class="small">{{ b.old }}</s> <b>{{ b.offer }}</b></span>
              <span class="small">{{ b.term }}</span>
              <span class="small r">Полный объём ремонта</span>
            </button>
          </div>

          <p class="small">
            Сравнивайте стоимость, сроки, рейтинг и отзывы. Все предложения — на полный объём ремонта.
            Условные компании, цены и сроки; пример аукциона не связан с расчётом выше.
          </p>
        </div>
      </div>
    </div>

    <div class="toast" :class="{ show: toastText }" role="status">{{ toastText }}</div>

    <dialog class="modal" ref="cpr" aria-labelledby="cprTitle" @click.self="cpr?.close()">
      <button class="close" type="button" aria-label="Закрыть" @click="cpr?.close()">×</button>
      <span class="eyebrow red">Цифровой паспорт ремонта</span>
      <h3 id="cprTitle">Что, где и как должно быть сделано.</h3>
      <ul>
        <li>Данные планировки и замеров помещений.</li>
        <li>Согласованный план расстановки мебели.</li>
        <li>Выбранные материалы и технологии ремонта.</li>
        <li>Проекты электрики и сантехники.</li>
        <li>Состав работ и требования к результату.</li>
      </ul>
      <p class="small">
        Паспорт можно заказать как отдельную платную услугу и передать любому подрядчику.
        Если заказываете ремонт через ZETTA HOME, паспорт входит в стоимость ремонта.
      </p>
    </dialog>
  </section>
</template>

<style scoped>
/* Палитра берётся из main.css прототипа В; тёмная полоса — как .sec.dark эталона */
.repair {
  --d-ink: var(--ink, #1a1a2e);
  --d-on: var(--paper, #f8f6f1);
  --d-muted: #C9C1B5;
  --d-dim: #A69D90;
  --d-line: rgba(237, 232, 224, .16);
  --d-accent: var(--red-l, #E9826F);
  padding: 100px 0;
  background: var(--d-ink);
  color: var(--d-on);
}
.repair .wrap { width: min(calc(100% - 40px), var(--max, 1200px)); margin: 0 auto; }
.repair .grid12 {
  display: grid; grid-template-columns: repeat(12, minmax(0, 1fr));
  gap: 24px; align-items: end; margin-bottom: 72px;
}
.repair .body { align-items: start; margin-bottom: 0; }
.repair .intro { grid-column: 1 / span 6; display: flex; flex-direction: column; gap: 28px; }
.repair .shot { grid-column: 8 / span 5; margin: 0; display: flex; flex-direction: column; gap: 10px; }
.repair .shot-media { aspect-ratio: 4 / 3; overflow: hidden; border-radius: 14px; background: #000; }
.repair .photo { width: 100%; height: 100%; object-fit: cover; }
.repair .rules { grid-column: 1 / span 6; }

.repair .eyebrow {
  display: flex; align-items: center; gap: 14px;
  font-size: 12px; font-weight: 600; letter-spacing: .16em; text-transform: uppercase;
}
.repair .eyebrow b { color: var(--d-accent); font-weight: 600; }
.repair .eyebrow i { display: block; width: 40px; height: 1px; background: currentColor; }
.repair .h2 {
  font-family: 'Cormorant Garamond', Georgia, serif; font-weight: 500;
  font-size: 72px; line-height: .98; letter-spacing: -.02em; margin: 0;
  color: var(--d-on);
}
.repair .h3 { font-family: 'Cormorant Garamond', Georgia, serif; font-weight: 500; font-size: 30px; line-height: 1.1; margin: 0; }
.repair .it { font-style: italic; color: var(--d-accent); }
.repair .lead { margin: 0; font-size: 17px; line-height: 1.6; color: var(--d-muted); max-width: 520px; }
.repair .proto-note { font-size: 12px; color: var(--d-dim); }
.repair .note { margin: 24px 0 0; font-size: 13px; line-height: 1.6; }
.repair .small { font-size: 13px; line-height: 1.55; color: var(--d-muted); margin: 0; }
.repair .muted { color: var(--d-muted); }
.repair .serif { font-family: 'Cormorant Garamond', Georgia, serif; font-weight: 500; }

/* Сравнение — .cmp эталона: две колонки, первая приглушена */
.repair .cmp { width: 100%; border-collapse: collapse; }
.repair .cmp th {
  text-align: left; padding: 0 24px 18px 0; font-size: 11px; font-weight: 600;
  letter-spacing: .16em; text-transform: uppercase; color: var(--d-dim); width: 50%;
}
.repair .cmp th + th { color: var(--d-accent); }
.repair .cmp td {
  padding: 18px 24px 18px 0; border-top: 1px solid var(--d-line);
  vertical-align: top; font-size: 15px; line-height: 1.5;
}
.repair .cmp td:first-child { color: var(--d-dim); }

.repair .cpr-box {
  margin-top: 36px; padding: 28px; border: 1px solid var(--d-line);
  display: flex; flex-direction: column; gap: 18px; align-items: flex-start; border-radius: 12px;
}
.repair .cpr-box p { margin: 0; font-size: 15px; line-height: 1.6; }
.repair .btn {
  display: inline-flex; align-items: center; gap: 12px; min-height: 44px; padding: 0 20px;
  border: 1px solid transparent; border-radius: 999px; font-size: 13px; font-weight: 600;
  cursor: pointer; transition: background .2s, color .2s, border-color .2s;
}
.repair .btn.light { background: var(--d-on); color: var(--d-ink); border-color: var(--d-on); }
.repair .btn.light:hover { background: var(--d-accent); border-color: var(--d-accent); color: var(--d-on); }

/* Панель ставок — светлая карточка на тёмной полосе, как в эталоне */
.repair .bids {
  grid-column: 8 / span 5; display: flex; flex-direction: column; gap: 18px;
  padding: 32px; border-radius: 14px; background: var(--d-on); color: var(--d-ink);
}
.repair .bids .small { color: #5f584f; }
.repair .bids .muted { color: #5f584f; }
.repair .bids-head { display: flex; justify-content: space-between; align-items: center; gap: 12px; flex-wrap: wrap; }
.repair .tag {
  display: inline-block; font-size: 11px; font-weight: 600; letter-spacing: .14em;
  text-transform: uppercase; padding: 6px 10px; border: 1px solid var(--line2, #CFC6B8);
  border-radius: 999px; color: #5f584f;
}
.repair .doc-strip {
  display: flex; gap: 14px; align-items: center; padding: 14px 16px;
  border: 1px solid var(--line, #e0dbd2); border-radius: 10px; font-size: 14px;
}
.repair .doc-icon {
  font-size: 11px; font-weight: 700; letter-spacing: .1em;
  border: 1px solid var(--d-ink); padding: 6px 8px; border-radius: 4px; flex-shrink: 0;
}

/* Ставка: сетка «имя — цена», вторая строка — срок и объём */
.repair .bid-list { display: flex; flex-direction: column; }
.repair .bid {
  display: grid; grid-template-columns: 1fr auto; gap: 4px 12px;
  padding: 16px 14px; margin: 0 -14px; width: calc(100% + 28px);
  border: 0; border-bottom: 1px solid var(--line, #e0dbd2);
  background: transparent; color: var(--d-ink);
  font: inherit; text-align: left; cursor: pointer; transition: background .2s, color .2s;
}
.repair .bid:hover { background: var(--stone, #EAE4DA); }
.repair .bid.on { background: var(--d-ink); color: var(--d-on); }
.repair .bid.on .small, .repair .bid.on s { color: #C9C1B5; }
.repair .bid .name { font-size: 24px; }
.repair .bid .r { text-align: right; }
.repair .bid s { font-size: 12px; font-weight: 400; color: #5f584f; }
.repair .bid b { font-size: 17px; font-weight: 600; }
.repair .bid.on b { color: var(--d-accent); }

/* Toast — как .toast эталона */
.repair .toast {
  position: fixed; left: 50%; bottom: 28px; transform: translate(-50%, 20px);
  background: var(--d-ink); color: var(--d-on); padding: 14px 22px; border-radius: 999px;
  font-size: 14px; font-weight: 600; opacity: 0; pointer-events: none;
  transition: opacity .3s, transform .3s; z-index: 120;
}
.repair .toast.show { opacity: 1; transform: translate(-50%, 0); }

/* Модалка ЦПР */
.repair .modal {
  width: min(560px, calc(100% - 32px)); padding: 32px; border: 0; border-radius: 16px;
  background: var(--paper, #f8f6f1); color: var(--ink, #1a1a2e);
}
.repair .modal::backdrop { background: #1E1C1999; }
.repair .modal .close {
  position: absolute; top: 14px; right: 14px; border: 0; background: none;
  font-size: 22px; line-height: 1; cursor: pointer; color: inherit;
}
.repair .modal h3 { font-family: 'Cormorant Garamond', Georgia, serif; font-weight: 500; font-size: 28px; margin: 12px 0 16px; }
.repair .modal ul { margin: 0 0 16px; padding-left: 20px; font-size: 14px; line-height: 1.7; color: #5f584f; }
.repair .modal .small { color: #5f584f; }
.repair .modal .eyebrow { color: var(--red, #B53A27); margin-bottom: 6px; }

@media (max-width: 1100px) {
  .repair .h2 { font-size: 56px; }
}
@media (max-width: 1000px) {
  .repair .grid12 { grid-template-columns: minmax(0, 1fr); gap: 28px; }
  .repair .intro, .repair .shot, .repair .rules, .repair .bids { grid-column: auto; }
  .repair .shot { max-width: 560px; }
}
@media (max-width: 700px) {
  .repair { padding: 64px 0; }
  .repair .h2 { font-size: 42px; }
  .repair .grid12 { margin-bottom: 48px; }
  .repair .bids { padding: 22px; }
  .repair .cmp td { padding-right: 10px; }
  .repair .bid { margin: 0; width: 100%; padding: 14px 0; }
  .repair .bid .name { font-size: 20px; }
}
</style>
