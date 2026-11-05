<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { api } from '@/api/client'
import { useFavoritesStore } from '@/stores/favorites'
import { closeOnBackdrop } from '@/composables/useDialog'

interface Idea {
  id: string
  type?: string
  title: string
  room: string
  style: string
  palette?: string
  colors: string[]
  image: string
  source?: string
}

const emit = defineEmits<{ (e: 'meeting', topic: string, comment?: string): void }>()

const favorites = useFavoritesStore()
const { ideas: savedIds, ideasCount } = storeToRefs(favorites)

const ideas = ref<Idea[]>([])
onMounted(async () => {
  try {
    const { data } = await api.get<Idea[]>('/api/interiors')
    ideas.value = data.filter((i) => i.type === 'idea').slice(0, 6)
  } catch {
    /* API недоступен — секция остаётся пустой, как в оригинале без данных */
  }
})

// В диалоге показываем сохранённые идеи; если карточка не из этой шестёрки — только id
const savedList = computed(() =>
  savedIds.value.map((id) => ideas.value.find((i) => i.id === id)).filter(Boolean) as Idea[],
)

const savedDialog = ref<HTMLDialogElement | null>(null)
function openSaved() {
  savedDialog.value?.showModal()
}
function closeSaved() {
  savedDialog.value?.close()
}
</script>

<template>
  <section class="catalog section" id="ideas" data-block="Интерьерные идеи">
    <div class="wrap">
      <div class="head">
        <div>
          <span class="eyebrow">
            <a class="route-number" href="#step-01" aria-label="К шагу 01 в схеме работы">01</a> / Дизайн и вдохновение</span>
          <h2>Сначала — интерьер, который вам нравится.</h2>
        </div>
        <p class="intro">Сохраняйте идеи для разных комнат — из них складывается образ вашей будущей квартиры.</p>
      </div>
      <div class="ideas-tools">
        <button class="saved-open" id="collectionCta" @click="openSaved">
          ♡ Мои идеи <span class="count" id="ideasSavedCount">{{ ideasCount }}</span>
        </button>
      </div>
      <div class="grid ideas-editorial63">
        <article class="card idea" v-for="idea in ideas" :key="idea.id" :data-id="idea.id">
          <div class="visual">
            <a class="image-open" :href="idea.source" target="_blank" rel="noopener"
               :aria-label="`${idea.room} «${idea.title}» — открыть на сайте`">
              <img :src="idea.image" referrerpolicy="no-referrer" :alt="`${idea.room} «${idea.title}»`" loading="lazy">
            </a>
            <button class="heart"
                    :aria-pressed="savedIds.includes(idea.id)"
                    :aria-label="(savedIds.includes(idea.id) ? 'Удалить из «Моих идей»: ' : 'Сохранить: ') + idea.title"
                    @click="favorites.toggleIdea(idea.id)">
              {{ savedIds.includes(idea.id) ? '♥' : '♡' }}
            </button>
          </div>
          <p class="meta">
            <span>{{ idea.room }}</span>
            <span>{{ idea.style }}</span>
          </p>
          <a class="title" :href="idea.source" target="_blank" rel="noopener">
            <h3>{{ idea.title }}</h3>
          </a>
          <div class="palette">
            <span class="palette-name">{{ idea.palette }}</span>
            <div class="swatches" role="img" :aria-label="`Палитра: ${idea.colors.join(', ')}`">
              <i v-for="c in idea.colors" :key="c" :style="{ background: c }"></i>
            </div>
          </div>
        </article>
      </div>
      <div class="bottom">
        <RouterLink class="button" to="/ideas">Смотреть все идеи интерьера <span aria-hidden="true">↗</span></RouterLink>
      </div>
    </div>

    <!-- Диалог «Мои идеи» -->
    <dialog class="modal" id="savedDialog" ref="savedDialog" aria-labelledby="savedTitle" @click="closeOnBackdrop">
      <button class="close" aria-label="Закрыть" @click="closeSaved">×</button>
      <div class="modal-inner">
        <span class="eyebrow red">Начало вашего интерьера</span>
        <h2 id="savedTitle">Мои идеи</h2>
        <p>Идеи сохраняются только в этом браузере.</p>
        <div id="savedList">
          <div v-if="!savedList.length" class="empty">Здесь появятся понравившиеся интерьеры. Нажмите на сердечко в каталоге.</div>
          <div class="saved-item" v-for="idea in savedList" :key="idea.id">
            <img :src="idea.image" :alt="idea.title">
            <div>
              <strong>{{ idea.title }}</strong>
              <p>{{ idea.room }} · {{ idea.style }}</p>
            </div>
            <button :aria-label="'Удалить ' + idea.title" @click="favorites.toggleIdea(idea.id)">×</button>
          </div>
        </div>
        <div class="actions" style="margin-top:25px">
          <button class="btn redbtn" v-if="savedList.length"
                  @click="closeSaved(); emit('meeting', 'Обсудить мой проект', `Хочу обсудить сохранённые идеи — ${savedIds.length} шт.`)">Обсудить мои идеи ↗</button>
          <button class="btn light" @click="closeSaved">Продолжить выбор</button>
        </div>
      </div>
    </dialog>
  </section>
</template>
