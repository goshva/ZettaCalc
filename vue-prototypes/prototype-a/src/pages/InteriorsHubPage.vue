<script setup lang="ts">
/**
 * Главная приложения «Интерьеры» — перенос hub() из оригинала:
 * обложка, три перехода (решения / квартиры / идеи), подборки карточек.
 */
import { computed, onMounted } from 'vue'
import { useInteriorsStore } from '@/stores/interiors'
import { useAppStyles } from '@/composables/useAppStyles'
import InteriorCard from '@/components/interiors/InteriorCard.vue'
import InteriorBand from '@/components/interiors/InteriorBand.vue'
import { imgUrl } from '@/composables/useImg'

useAppStyles('interiors', 'inner-design65')

const interiors = useInteriorsStore()

const hero = computed(() => interiors.byId('idea0'))
const rooms = computed(() => ['room0', 'room1'].map((id) => interiors.byId(id)).filter(Boolean))
const ideas = computed(() => ['idea2', 'idea1'].map((id) => interiors.byId(id)).filter(Boolean))

onMounted(() => interiors.fetch())
</script>

<template>
  <section class="section">
    <div class="wrap">
      <div class="crumb">Главная / Интерьеры</div>

      <section class="hubhero">
        <div>
          <div class="eyebrow">Интерьеры ZETTA HOME</div>
          <h1>От понравившегося<br>интерьера —<br>к понятной смете.</h1>
          <p class="lead">Выберите образ комнаты или всей квартиры. Посмотрите мебель, материалы и стоимость работ.</p>
          <RouterLink class="btn" to="/interiors/rooms">Смотреть интерьеры со сметой</RouterLink>
        </div>
        <img v-if="hero" :src="imgUrl(hero.image)" alt="Кухня Шалфей и терракота">
      </section>

      <div class="journey">
        <RouterLink to="/interiors/rooms">
          <small>01 / КОНКРЕТНЫЕ РЕШЕНИЯ</small>
          <h3>Интерьеры со сметой</h3>
          <p>Отдельные помещения, подобранные товары и состав сметы.</p>
        </RouterLink>
        <RouterLink to="/interiors/apartments">
          <small>02 / КВАРТИРА ЦЕЛИКОМ</small>
          <h3>Квартиры со сметой</h3>
          <p>Планировка, все помещения и три варианта бюджета.</p>
        </RouterLink>
        <RouterLink to="/interiors/ideas">
          <small>03 / НАЙТИ СВОЙ ОБРАЗ</small>
          <h3>Идеи интерьеров</h3>
          <p>Помещения, стили и цветовые схемы для вашего дома.</p>
        </RouterLink>
      </div>

      <InteriorBand />

      <section class="section">
        <div class="row">
          <h2>Красивый интерьер.<br>Конкретный состав.</h2>
          <RouterLink class="textlink" to="/interiors/rooms">Все интерьеры со сметой</RouterLink>
        </div>
        <div class="grid">
          <InteriorCard v-for="x in rooms" :key="x!.id" :item="x!" />
        </div>
      </section>

      <section class="section">
        <div class="row">
          <h2>Сначала — то, что нравится</h2>
          <RouterLink class="textlink" to="/interiors/ideas">Все идеи интерьеров</RouterLink>
        </div>
        <div class="grid">
          <InteriorCard v-for="x in ideas" :key="x!.id" :item="x!" />
        </div>
      </section>

      <div class="band">
        <div>
          <h3>У вас уже есть дизайн-проект?</h3>
          <p>Приходите с ним в Центр — обсудим подбор мебели, материалов и реализацию.</p>
        </div>
        <RouterLink class="btn secondary" to="/appointment">Записаться в Интерьерный центр</RouterLink>
      </div>
    </div>
  </section>
</template>
