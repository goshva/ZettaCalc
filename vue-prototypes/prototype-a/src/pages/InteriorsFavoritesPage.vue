<script setup lang="ts">
/**
 * Избранное — перенос favs() из оригинального приложения.
 *
 * Сохранённые интерьеры без ограничений, режим «Выбрать для встречи»
 * включает кнопки выбора, полоса снизу ведёт к записи с выбранными.
 */
import { computed, onMounted } from 'vue'
import { useInteriorsStore } from '@/stores/interiors'
import { useFavoritesStore } from '@/stores/favorites'
import { useAppStyles } from '@/composables/useAppStyles'
import InteriorCard from '@/components/interiors/InteriorCard.vue'

useAppStyles('interiors', 'inner-design65')

const interiors = useInteriorsStore()
const favorites = useFavoritesStore()

// Заходим сразу на избранное — список интерьеров нужен для карточек
onMounted(() => interiors.fetch())

const saved = computed(() =>
  favorites.ideas.map((id) => interiors.byId(id)).filter(Boolean) as NonNullable<ReturnType<typeof interiors.byId>>[],
)
</script>

<template>
  <section class="section">
    <div class="wrap">
      <div class="crumb">
        <RouterLink to="/interiors">Интерьеры</RouterLink> / Избранное
      </div>
      <h1>Всё, что вам нравится</h1>
      <p class="lead">Сохраняйте без ограничений. Для подготовки к встрече выберите
        от одного до пяти интерьеров.</p>

      <template v-if="saved.length">
        <div class="row" style="margin: 25px 0">
          <span class="muted">Сохранено: {{ saved.length }}</span>
          <button class="btn secondary" type="button" @click="favorites.choosing = !favorites.choosing">
            {{ favorites.choosing ? 'Завершить выбор' : 'Выбрать для встречи' }}
          </button>
        </div>

        <div class="grid">
          <InteriorCard v-for="x in saved" :key="x.id" :item="x" :pick="favorites.choosing" />
        </div>

        <div class="selected-strip">
          <span>Для встречи: {{ favorites.meetingCount }} из {{ favorites.maxMeeting }}</span>
          <RouterLink class="btn" to="/appointment">
            {{ favorites.meetingCount ? 'Записаться с выбранными' : 'Записаться в Интерьерный центр' }}
          </RouterLink>
        </div>
      </template>

      <div v-else class="empty">
        <h2>Здесь будут ваши интерьеры</h2>
        <p class="muted">Нажимайте сердечко на понравившихся карточках.<br>
          Вход в аккаунт для этого не нужен.</p>
        <RouterLink class="btn" to="/interiors/ideas">Посмотреть идеи</RouterLink>
        <p style="margin-top: 20px">
          <RouterLink class="textlink" to="/appointment">Записаться без выбора</RouterLink>
        </p>
      </div>

      <p class="scope">Избранное сохраняется в этом браузере. Выбор для встречи остаётся
        тем же при переходе между страницами.</p>
    </div>
  </section>
</template>
