<script setup lang="ts">
/**
 * Футер приложения «Центр» — перенос `<footer>` из apps52/center.html.
 *
 * На странице записи CSS главной выключен (конфликтует с center.css), поэтому
 * футер главной здесь не подходит: в оригинале у приложения «Центр» свой футер.
 */
</script>

<template>
  <footer>
    <div class="wrap foot">
      <div>
        <div class="brand"><i></i><b>ZETTA</b> <span>home</span></div>
        <p>От идеи до готовой квартиры<br>Система управления ремонтом</p>
      </div>
      <div>
        <p>Москва, Аминьевское шоссе, 6</p>
        <p><a href="tel:+74951387338">+7 (495) 138-73-38</a></p>
      </div>
      <div class="footlinks">
        <RouterLink to="/appointment">Записаться в Интерьерный центр</RouterLink>
        <a href="https://zettahome.ru/privacy-policy/" target="_blank" rel="noopener">
          Политика конфиденциальности
        </a>
        <RouterLink to="/#company" class="about-footer68">О ZETTA HOME</RouterLink>
      </div>
    </div>
  </footer>
</template>
