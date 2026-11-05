<template>
  <footer class="ftr">
    <div class="wrap">
      <div class="footer-grid">
        <div class="f-col f-brand">
          <RouterLink to="/" class="brand"><i></i>ZETTA <span>HOME</span></RouterLink>
          <p>Дизайн, ремонт, мебель и материалы.<br />В комплекте одной квартиры.</p>
        </div>

        <div class="f-col">
          <h3>Направления</h3>
          <RouterLink to="/ideas">Дизайн и идеи</RouterLink>
          <RouterLink to="/#repair">Ремонт</RouterLink>
          <RouterLink to="/#supply">Мебель и кухни</RouterLink>
          <RouterLink to="/#supply">Материалы и сантехника</RouterLink>
        </div>

        <div class="f-col">
          <h3>Ваш проект</h3>
          <RouterLink to="/#budget">Ориентир бюджета</RouterLink>
          <RouterLink to="/#how">Как это работает</RouterLink>
          <RouterLink to="/#center">Центр решений</RouterLink>
          <RouterLink to="/#faq">Вопросы и ответы</RouterLink>
        </div>

        <div class="f-col">
          <h3>Встреча в Центре</h3>
          <p>Москва<br />Аминьевское шоссе, 6</p>
          <button class="link" @click="$emit('appointment')">Записаться в Интерьерный центр</button>
        </div>
      </div>

      <div class="footer-bottom">
        <span>© ZETTA HOME · Прототип главной</span>
        <span>Фото — интерьерные референсы, не реализованные проекты ZETTA HOME.</span>
        <RouterLink to="/about">О ZETTA HOME</RouterLink>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
defineEmits<{ (e: 'appointment'): void }>()
</script>

<style scoped>
.ftr { background: var(--ink); color: #c9c1b5; padding: 72px 0 32px; }
.ftr a, .ftr .link { color: #c9c1b5; }
.ftr a:hover, .ftr .link:hover { color: #fff; }
.footer-grid {
  display: grid;
  grid-template-columns: 1.4fr 1fr 1fr 1.2fr;
  gap: 40px 24px;
}
.f-brand p { margin: 18px 0 0; font-size: 14px; color: #a69d90; line-height: 1.6; }
.brand { font-size: 22px; font-weight: 700; letter-spacing: -0.05em; color: #fff; }
.brand i {
  display: inline-block; width: 8px; height: 8px; background: var(--red);
  border-radius: 50%; margin-right: 6px; vertical-align: middle;
}
.brand span { font-weight: 400; color: var(--red); }
.f-col h3 { color: #f2ede4; font-size: 14px; font-weight: 600; margin: 0 0 14px; }
.f-col a, .f-col p, .f-col span { display: block; font-size: 14px; margin: 10px 0; line-height: 1.5; }
.f-col p { color: #a69d90; }
.link {
  background: none; border: 0; padding: 0; font: inherit; cursor: pointer;
  color: #f2ede4; border-bottom: 1px solid #f2ede4; text-align: left;
}
.footer-bottom {
  display: flex; justify-content: space-between; gap: 16px; flex-wrap: wrap;
  margin-top: 56px; padding-top: 22px; border-top: 1px solid #3a3631;
  font-size: 12px; color: #a69d90;
}
@media (max-width: 900px) { .footer-grid { grid-template-columns: 1fr 1fr; } }
@media (max-width: 560px) { .footer-grid { grid-template-columns: 1fr; } }
</style>
