<script setup lang="ts">
/** Палитра интерьера — palette(x) из оригинала: кружки цветов и название схемы */
defineProps<{ colors: string[]; name: string }>()
</script>

<template>
  <div class="palette">
    <span class="swatches">
      <i v-for="c in colors" :key="c" :style="{ background: c }"></i>
    </span>
    <span>{{ name }}</span>
  </div>
</template>
