<script setup lang="ts">
/**
 * Футер приложения «Интерьеры» — перенос `<footer>` из interiors.html.
 *
 * В оригинале он лежит внутри `#surface52` и не скрывается вместе с шапкой
 * приложения, поэтому оформление приходит из /api/apps/interiors/css.
 */
defineEmits<{ (e: 'appointment'): void }>()
</script>

<template>
  <footer>
    <div class="wrap">
      <div>
        <RouterLink class="logo" to="/">ZETTA <em>home</em></RouterLink>
        <p>От идеи до готовой квартиры</p>
        <p>Система управления ремонтом</p>
      </div>
      <div>
        <RouterLink to="/#center">Центр интерьерных решений</RouterLink>
        <p>Москва, Аминьевское шоссе, 6</p>
        <a href="tel:+74951387338">+7 (495) 138-73-38</a>
      </div>
      <div>
        <button class="btn" type="button" @click="$emit('appointment')">
          Записаться в Интерьерный центр
        </button>
        <p>Подбор мебели и материалов — бесплатно</p>
      </div>
      <RouterLink to="/#company" class="about-footer68">О ZETTA HOME</RouterLink>
    </div>
  </footer>
</template>
