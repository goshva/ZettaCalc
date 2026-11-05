<script setup lang="ts">
/**
 * Страница квартиры — перенос apartment(x) из оригинального приложения.
 *
 * Обложка с галереей и планом, перечень помещений с карточками и раздел
 * «Подборка для всей квартиры»: ценовые варианты и раскрывающиеся сметы
 * каждого помещения по позициям.
 */
import { computed } from 'vue'
import { useInteriorsStore } from '@/stores/interiors'
import { useInteriorItem } from '@/composables/useInteriorItem'
import { useAppStyles } from '@/composables/useAppStyles'
import { money, total } from '@/utils/interiors'
import InteriorGallery from '@/components/interiors/InteriorGallery.vue'
import InteriorPlanPanel from '@/components/interiors/InteriorPlanPanel.vue'
import InteriorActions from '@/components/interiors/InteriorActions.vue'
import InteriorTiers from '@/components/interiors/InteriorTiers.vue'
import InteriorLines from '@/components/interiors/InteriorLines.vue'
import InteriorWorks from '@/components/interiors/InteriorWorks.vue'
import InteriorBand from '@/components/interiors/InteriorBand.vue'
import InteriorCard from '@/components/interiors/InteriorCard.vue'

useAppStyles('interiors', 'inner-design65')

const interiors = useInteriorsStore()
const { item, loading, error } = useInteriorItem()

const rooms = computed(() => (item.value ? interiors.childrenOf(item.value) : []))
const tier = computed(() => interiors.tier)
</script>

<template>
  <section class="section">
    <div class="wrap">
      <p v-if="loading && !item" class="muted">Загрузка квартиры…</p>
      <p v-else-if="error" class="notice">{{ error }}</p>
      <p v-else-if="!item" class="notice">Квартира не найдена в демонстрационной выборке.</p>

      <template v-else>
        <div class="crumb">
          <RouterLink to="/interiors/apartments">Квартиры со сметой</RouterLink> / {{ item.title }}
        </div>

        <div class="apttitle">
          <div>
            <div class="eyebrow">{{ item.room }} · {{ item.area }} м²</div>
            <h1>{{ item.title }}</h1>
            <p class="complex">{{ item.complex }}</p>
            <p class="lead">{{ item.description }}</p>
          </div>
          <InteriorActions :item="item" />
        </div>

        <section class="apthero">
          <InteriorGallery :item="item" />
          <InteriorPlanPanel :item="item"
                             :note="`Планировка из проекта. Нажмите «Увеличить план», чтобы рассмотреть размеры.`"
                             zoomable />
        </section>

        <section class="section">
          <div class="row">
            <div>
              <h2>Посмотрите каждое помещение</h2>
              <p class="muted">Образ, конкретные товары и состав подборки — внутри карточки.</p>
            </div>
          </div>
          <nav class="pills">
            <RouterLink v-for="r in rooms" :key="r.id" class="pill" :to="`/interiors/room/${r.id}`">
              {{ r.room }}
            </RouterLink>
          </nav>
          <div class="grid">
            <InteriorCard v-for="r in rooms" :key="r.id" :item="r" :tier="tier" />
          </div>
        </section>

        <section class="section apartmentbudget">
          <div>
            <div class="eyebrow">Стоимость по помещениям</div>
            <h2>Подборка для всей квартиры</h2>
            <p>Выберите ценовой вариант и раскройте нужное помещение, чтобы посмотреть все позиции.</p>
            <p class="scope">В исходных файлах представлены
              {{ item.variants?.length === 1
                ? 'один расчёт (колонки «Бюджет» и «Премиум» совпадают)'
                : 'варианты «Бюджет» и «Премиум»' }}.
              В интерфейсе «Бюджет» показан как «Базовый», «Премиум» — как «Комфортный».
              Оптимальный вариант и смета работ будут добавлены после подготовки.</p>
            <p class="scope">Стиль и палитра для этих двух квартир предложены для прототипа
              и требуют утверждения.</p>
          </div>
          <div>
            <InteriorTiers :item="item" />
            <details v-for="r in rooms" :key="r.id">
              <summary>
                <span>{{ r.room }}</span>
                <strong>{{ money(total(r, tier)) }} +</strong>
              </summary>
              <InteriorLines :lines="r.lines ?? []" />
            </details>
            <div class="estimate-total">
              <span>Мебель, материалы и оснащение</span>
              <strong>{{ money(total(item, tier)) }}</strong>
            </div>
            <p class="scope">Ремонтные работы не включены. Заказная мебель и отмеченные
              в спецификации позиции требуют уточнения.</p>
          </div>
        </section>

        <InteriorWorks />
        <InteriorBand />
        <InteriorBand variant="upload" />
      </template>
    </div>
  </section>
</template>
