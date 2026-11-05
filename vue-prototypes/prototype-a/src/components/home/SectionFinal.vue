<script setup lang="ts">
const emit = defineEmits<{ (e: 'meeting', topic: string): void }>()
</script>

<template>
  <section class="section final" data-block="Следующий шаг">
    <div class="wrap">
      <div>
        <span class="eyebrow">Начните с вашей квартиры</span>
        <h2>Одна задача сегодня. Целый интерьер — дальше.</h2>
        <p>Расскажите, что у вас уже есть и что нужно решить. Начнём с подходящего вам шага.</p>
      </div>
      <div class="actions">
        <button class="btn" @click="emit('meeting', 'Обсудить мой проект')">Обсудить мой проект <span class="arrow">↗</span></button>
        <a class="btn light" href="#ideas">Сначала посмотреть идеи</a>
      </div>
    </div>
  </section>
</template>
