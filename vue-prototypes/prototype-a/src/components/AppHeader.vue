<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useFavoritesStore } from '@/stores/favorites'
import { storeToRefs } from 'pinia'

const emit = defineEmits<{ (e: 'meeting', topic: string): void }>()

const favorites = useFavoritesStore()
const { totalCount: savedCount } = storeToRefs(favorites)

const menuOpen = ref(false)
const catalogOpen = ref(false)

// закрытие выпадающего каталога по клику вне
function onDocClick(e: MouseEvent) {
  const t = e.target as HTMLElement
  if (!t.closest('.catalog-menu')) catalogOpen.value = false
}
onMounted(() => document.addEventListener('click', onDocClick))
onBeforeUnmount(() => document.removeEventListener('click', onDocClick))
</script>

<template>
  <header>
    <div class="wrap nav">
      <RouterLink to="/" class="brand" aria-label="Zetta Home, начало"><i></i>ZETTA <span>HOME</span></RouterLink>

      <nav class="navlinks" :class="{ open: menuOpen }" aria-label="Главное меню">
        <RouterLink to="/#ideas" @click="menuOpen = false">Дизайн</RouterLink>
        <div class="catalog-menu">
          <button @click.stop="catalogOpen = !catalogOpen">Каталог <span aria-hidden="true">⌄</span></button>
          <div class="catalog-dropdown" v-show="catalogOpen">
            <RouterLink to="/ideas" @click="catalogOpen = false; menuOpen = false">Дизайн-концепции ↗</RouterLink>
            <RouterLink to="/catalog" @click="catalogOpen = false; menuOpen = false">Мебель и товары ↗</RouterLink>
            <RouterLink to="/catalog?group=materials" @click="catalogOpen = false; menuOpen = false">Отделочные материалы ↗</RouterLink>
            <RouterLink to="/catalog?group=bath" @click="catalogOpen = false; menuOpen = false">Сантехника ↗</RouterLink>
          </div>
        </div>
        <RouterLink to="/interiors/apartments" @click="menuOpen = false">Квартиры со сметами</RouterLink>
        <RouterLink to="/#repair" @click="menuOpen = false">Ремонт</RouterLink>
        <RouterLink to="/#supply" @click="menuOpen = false">Комплектация</RouterLink>
        <RouterLink to="/#center" @click="menuOpen = false">Центр</RouterLink>
        <RouterLink to="/#how" @click="menuOpen = false">Как это работает</RouterLink>
        <button class="btn redbtn navlinks-cta" @click="emit('meeting', 'Получить консультацию'); menuOpen = false">
          Получить консультацию ↗
        </button>
      </nav>

      <div class="navtools">
        <RouterLink to="/favorites" class="saved-btn" aria-label="Мои идеи, открыть сохранённое">
          ♡<span class="lbl">Мои идеи</span>
          <span class="count" v-if="savedCount > 0">{{ savedCount }}</span>
        </RouterLink>
        <button class="btn" @click="emit('meeting', 'Получить консультацию')">
          Получить консультацию <span class="arrow">↗</span>
        </button>
        <button class="mobile-toggle" :aria-expanded="menuOpen" aria-label="Открыть меню" @click="menuOpen = !menuOpen">☰</button>
      </div>
    </div>
  </header>
</template>

<style scoped>
/* точная 1:1 с оригиналом: header > .wrap.nav > .brand / .navlinks / .navtools
   все стили приходят из /api/css/a (#surface52 header, .navlinks, .catalog-dropdown…) */
.catalog-menu > button {
  border: 0; background: none; font: inherit; color: inherit; cursor: pointer; padding: 0;
}
.catalog-dropdown { display: flex; flex-direction: column; }
.catalog-dropdown a { display: block; padding: 12px; border-radius: 6px; font-size: 14px; }
.catalog-dropdown a:hover { background: #eeefe9; }
.saved-btn { display: inline-flex; align-items: center; gap: 8px; }
.saved-btn .count {
  background: var(--red); color: #fff; font-size: 11px; font-weight: 700;
  min-width: 18px; height: 18px; border-radius: 9px;
  display: inline-flex; align-items: center; justify-content: center; padding: 0 5px;
}
</style>
