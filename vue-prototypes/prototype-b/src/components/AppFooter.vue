<template>
  <footer class="ftr">
    <div class="wrap inner">
      <div>
        <div class="brand"><b>ZETTA</b><em>home</em></div>
        <p class="muted">Интерьерный центр · Москва · Аминьевское шоссе, 6</p>
      </div>
      <div>
        <h4>Навигация</h4>
        <RouterLink to="/">Главная</RouterLink>
        <RouterLink to="/gallery">Галерея</RouterLink>
        <RouterLink to="/calculator">Калькулятор</RouterLink>
      </div>
      <div>
        <h4>API</h4>
        <a href="http://127.0.0.1:8000/docs" target="_blank">Swagger UI</a>
        <a href="http://127.0.0.1:8000/api/health" target="_blank">Health Check</a>
      </div>
    </div>
    <div class="wrap bottom">
      <span>© 1998–2026 ZETTA · Vue 3 прототип B</span>
    </div>
  </footer>
</template>
<style scoped>
.ftr { background: var(--ink); color: #fff; padding: 48px 0 20px; margin-top: 60px; }
.inner { display: grid; grid-template-columns: 2fr 1fr 1fr; gap: 40px; }
.brand { font-size: 22px; font-weight: 800; letter-spacing: 0.08em; margin-bottom: 12px; }
.brand em { font-style: italic; font-weight: 400; color: var(--gold); margin-left: 4px; }
.muted { color: #aeb7a6; font-size: 14px; }
h4 { margin: 0 0 12px; font-size: 14px; }
.ftr a { display: block; font-size: 13px; color: #dddfd7; margin: 8px 0; }
.ftr a:hover { color: #fff; }
.bottom { border-top: 1px solid #3a3d38; margin-top: 32px; padding-top: 16px; font-size: 12px; color: #aeb7a6; }
</style>
