<script setup lang="ts">
/**
 * Единая оболочка прототипа А — перенос shellbar + header.shell-header
 * из оригинала (файл a_page/shell.html).
 *
 * В оригинале оболочка живёт вне `#surface52` и не подменяется при переходе
 * между приложениями: собственные шапки приложений скрыты правилом
 * `#surface52>header{display:none!important}`. Поэтому на страницах
 * «Интерьеров» показывается именно эта шапка, а не шапка главной.
 */
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { storeToRefs } from 'pinia'
import { useFavoritesStore } from '@/stores/favorites'

const emit = defineEmits<{ (e: 'appointment'): void }>()

const favorites = useFavoritesStore()
const { totalCount: savedCount } = storeToRefs(favorites)

// Раскрытие подраздела «Интерьеры» и мобильного меню — как в оригинале
const interiorsOpen = ref(false)
const menuOpen = ref(false)

const sections: { title: string; to: string }[] = [
  { title: 'Идеи интерьеров', to: '/interiors/ideas' },
  { title: 'Комнаты со сметой', to: '/interiors/rooms' },
  { title: 'Квартиры со сметой', to: '/interiors/apartments' },
  { title: 'Загрузить дизайн-проект', to: '/interiors/upload' },
]

function onDocClick(e: MouseEvent) {
  const t = e.target as HTMLElement
  if (!t.closest('.interiors-menu52')) interiorsOpen.value = false
}
onMounted(() => document.addEventListener('click', onDocClick))
onBeforeUnmount(() => document.removeEventListener('click', onDocClick))
</script>

<template>
  <header class="shell-header">
    <div class="shellnav">
      <RouterLink class="brand51" to="/" aria-label="Zetta Home, на главную">
        <i></i><strong>ZETTA</strong> <span>home</span>
      </RouterLink>

      <nav class="nav51" :class="{ open: menuOpen }" aria-label="Разделы прототипа">
        <div class="interiors-menu52">
          <RouterLink to="/interiors">Интерьеры</RouterLink>
          <button type="button" aria-label="Разделы интерьеров"
                  :aria-expanded="interiorsOpen"
                  @click.stop="interiorsOpen = !interiorsOpen">⌄</button>
          <div class="drop51" v-show="interiorsOpen">
            <RouterLink v-for="s in sections" :key="s.to" :to="s.to"
                        @click="interiorsOpen = false; menuOpen = false">{{ s.title }}</RouterLink>
          </div>
        </div>
        <RouterLink to="/catalog" @click="menuOpen = false">Каталог товаров</RouterLink>
        <RouterLink to="/flow" @click="menuOpen = false">Расчёт и ЦПР</RouterLink>
        <RouterLink to="/#center" @click="menuOpen = false">Интерьерный центр</RouterLink>
        <RouterLink to="/#repair" @click="menuOpen = false">Ремонт</RouterLink>
        <RouterLink to="/#company" @click="menuOpen = false">О ZETTA HOME</RouterLink>
      </nav>

      <div class="tools51">
        <details class="fav51">
          <summary>♡ <span>{{ savedCount }}</span></summary>
          <div class="drop51">
            <RouterLink to="/interiors/favorites">Избранные интерьеры</RouterLink>
            <RouterLink to="/cart">Корзина товаров</RouterLink>
          </div>
        </details>
        <button class="book51" type="button" @click="emit('appointment')">
          Записаться в Интерьерный центр
        </button>
        <button class="mobile51" type="button" aria-label="Открыть меню"
                :aria-expanded="menuOpen" @click="menuOpen = !menuOpen">☰</button>
      </div>
    </div>
  </header>
</template>
