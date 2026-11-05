<script setup lang="ts">
/**
 * Кнопки карточки интерьера — actions(x) из оригинала.
 *
 * Вместо data-action="book" стоит ссылка на страницу записи: модальное окно
 * записи в приложении «Интерьеры» заменено отдельной страницей прототипа.
 */
import { computed } from 'vue'
import type { InteriorItem } from '@/types'
import { useFavoritesStore } from '@/stores/favorites'

const props = defineProps<{ item: InteriorItem }>()

const favorites = useFavoritesStore()
const saved = computed(() => favorites.ideas.includes(props.item.id))
</script>

<template>
  <div class="actions">
    <button class="btn secondary" type="button" :aria-pressed="saved"
            @click="favorites.toggleIdea(item.id)">
      {{ saved ? '♥ В избранном' : '♡ Сохранить в избранное' }}
    </button>
    <RouterLink class="btn" to="/appointment">Записаться в Интерьерный центр</RouterLink>
  </div>
</template>
