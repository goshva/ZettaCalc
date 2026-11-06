<script setup lang="ts">
import { onMounted, computed, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useReferencesStore } from '@/stores/references'
import { useEstimateStore } from '@/stores/estimate'
import ReferenceCard from '@/components/ReferenceCard.vue'
import { formatPrice } from '@/api/client'
import { useRoute, useRouter } from 'vue-router'
const router = useRouter() 
const route = useRoute()
const refs = useReferencesStore()
const est = useEstimateStore()
const { filtered, loading } = storeToRefs(refs)

/** Выбранный референс — из query ?ref=; пустой → демо-смета est_001 */
const selectedRef = computed(() => String(route.query.ref ?? ''))
const selectedRefItem = computed(
  () => (selectedRef.value ? refs.items.find((r) => r.id === selectedRef.value) ?? null : null),
)
const range = computed(() => est.current ? `${formatPrice(est.current.range_min)} — ${formatPrice(est.current.range_max)}` : '')
const days = computed(() => est.current?.days ?? 0)
const lines = computed(() => est.current?.lines ?? [])
const assumptions = computed(() => est.current?.assumptions ?? [])

function onCardCalculate(id: string) {
  if (String(route.query.ref ?? '') !== id) {
    router.push({ query: { ...route.query, ref: id } })
  } else {
    est.calculate(id) // принудительный пересчёт того же референса
  }
}
onMounted(() => {
  refs.fetch()
  est.calculate(selectedRef.value)
})

// Смена референса (клик «Рассчитать» на другой карточке) пересчитывает смету.
watch(selectedRef, (id) => est.calculate(id))
</script>

<template>
  <section class="section">
    <div class="wrap">
      <h2>Предварительный расчёт</h2>
      <p class="muted" style="margin-bottom: 32px">Бесплатно — со SKU, брендами и ценами из каталога</p>

      <div class="layout">
        <!-- Референсы слева -->
        <div class="refs">
          <h3>Выберите референс</h3>
          <div class="grid cols-2">
<ReferenceCard
  v-for="r in filtered"
  :key="r.id"
  :ref-item="r"
  :selected="r.id === selectedRef"
  @calculate="onCardCalculate"
/>
          </div>
        </div>

        <!-- Расчёт справа -->
        <div class="est-card">
          <h3>Результат расчёта</h3>
          <p v-if="selectedRefItem" class="muted ref-caption">
            По референсу: {{ selectedRefItem.title }}
          </p>
          <div v-if="est.loading" class="muted">Рассчитываю…</div>
          <template v-else-if="est.error">
            <p class="muted">{{ est.error }}</p>
          </template>
          <template v-else-if="est.current">
            <div class="total">{{ formatPrice(est.current.total_rub) }}</div>
            <div class="range">Диапазон: {{ range }}</div>
            <div class="meta">Срок: ~{{ days }} дней</div>

            <div class="lines">
              <div v-for="l in lines" :key="l.id" class="line">
                <span>{{ l.title }} · {{ l.qty }} {{ l.unit }}</span>
                <b>{{ formatPrice(l.price * l.qty) }}</b>
              </div>
            </div>

            <div class="assumptions" v-if="assumptions.length">
              <strong>Требуют проверки:</strong>
              <ul>
                <li v-for="(a, i) in assumptions" :key="i">{{ a }}</li>
              </ul>
            </div>
          </template>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.layout { display: grid; grid-template-columns: 1.3fr 1fr; gap: 40px; }
.refs { display: flex; flex-direction: column; gap: 16px; }
.est-card { background: #fff; border: 1px solid var(--line); border-radius: var(--radius); padding: 28px; align-self: start; position: sticky; top: 100px; }
.total { font-size: 42px; font-weight: 700; letter-spacing: -0.03em; }
.range { font-size: 14px; color: var(--muted); margin: 8px 0; }
.meta { font-size: 14px; margin-bottom: 20px; }
.lines { border-top: 1px solid var(--line); }
.line { display: flex; justify-content: space-between; padding: 12px 0; border-bottom: 1px solid var(--line); font-size: 14px; }
.line b { font-weight: 600; }
.assumptions { margin-top: 20px; padding: 14px; background: #fef3f0; border-radius: 8px; font-size: 13px; }
.assumptions ul { margin: 8px 0 0 20px; }
.ref-caption { font-size: 13px; margin: -6px 0 14px; }
.grid.cols-2 { grid-template-columns: 1fr 1fr; }
@media (max-width: 900px) {
  .layout { grid-template-columns: 1fr; }
  .est-card { position: static; }
  .grid.cols-2 { grid-template-columns: 1fr; }
}
</style>
