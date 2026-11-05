<script setup lang="ts">
/**
 * Полоса «Записаться в Интерьерный центр» — band() и uploadBand() из оригинала.
 *
 * Внутри приложения «Интерьеры» кнопка ведёт на страницу записи, поэтому
 * вместо data-action="book" используется RouterLink.
 */
withDefaults(defineProps<{ variant?: 'meeting' | 'upload' }>(), { variant: 'meeting' })
</script>

<template>
  <div class="band">
    <template v-if="variant === 'upload'">
      <div>
        <h3>Есть свой дизайн-проект?</h3>
        <p>Загрузите его — до встречи изучим проект и подготовим предварительный расчёт мебели, материалов и работ.</p>
      </div>
      <RouterLink class="btn secondary" to="/interiors/upload">Загрузить дизайн-проект</RouterLink>
    </template>
    <template v-else>
      <div>
        <h3>Выберите интерьеры для встречи</h3>
        <p>Сохраняйте понравившееся сердечком. При записи выберите из избранного до пяти интерьеров —
          подготовимся к вашему визиту.</p>
      </div>
      <RouterLink class="btn" to="/appointment">Записаться в Интерьерный центр</RouterLink>
    </template>
  </div>
</template>
