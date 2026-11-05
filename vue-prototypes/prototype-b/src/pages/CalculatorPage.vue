<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useReferencesStore } from '@/stores/references'
import { useEstimateStore } from '@/stores/estimate'
import { imgUrl } from '@/composables/useImg'

const refs = useReferencesStore()
const est = useEstimateStore()
const { items, loading } = storeToRefs(refs)
const selectedRef = ref('')

const total = computed(() => est.current?.total_rub ?? 0)
const lines = computed(() => est.current?.lines ?? [])
const assumptions = computed(() => est.current?.assumptions ?? [])

onMounted(async () => {
  await refs.fetch()
  await est.calculate()
})
</script>
<template>
  <section class="section">
    <div class="wrap">
      <h2>Калькулятор стоимости</h2>
      <p class="muted" style="margin-bottom: 36px">Бесплатно — со SKU, брендами и диапазонами</p>

      <div class="layout">
        <div class="left">
          <h3>Выберите референс</h3>
          <div class="grid-2" style="margin-bottom: 32px">
            <div v-for="r in items" :key="r.id"
                 class="ref-item" :class="{ sel: selectedRef === r.id }"
                 @click="selectedRef = r.id">
              <img :src="imgUrl(r.image)" :alt="r.title" />
              <strong>{{ r.title }}</strong>
            </div>
          </div>
        </div>
        <div class="right">
          <div class="result-card">
            <div class="eyebrow">Предварительная оценка</div>
            <div class="total" v-if="est.current">{{ formatPrice(total) }}</div>
            <div class="range" v-if="est.current">{{ formatPrice(est.current!.range_min) }} — {{ formatPrice(est.current!.range_max) }}</div>
            <div class="days" v-if="est.current">~{{ est.current.days }} рабочих дней</div>

            <div class="lines" v-if="lines.length">
              <div v-for="l in lines" :key="l.id" class="line">
                <span>{{ l.title }} · {{ l.qty }} {{ l.unit }}</span>
                <b>{{ formatPrice(l.price * l.qty) }}</b>
              </div>
            </div>

            <div class="assumptions" v-if="assumptions.length">
              <strong>⚠ Требуют проверки:</strong>
              <ul><li v-for="(a, i) in assumptions" :key="i">{{ a }}</li></ul>
            </div>

            <RouterLink to="/cpr" class="btn gold" style="width: 100%; margin-top: 20px;">
              Получить ЦПР
            </RouterLink>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
<script lang="ts">
import { formatPrice } from '@/api/client'
export default { data() { return { formatPrice } } }
</script>
<style scoped>
.layout { display: grid; grid-template-columns: 1.4fr 1fr; gap: 48px; }
.ref-item { cursor: pointer; border: 2px solid transparent; border-radius: 12px; overflow: hidden; }
.ref-item.sel { border-color: var(--gold); }
.ref-item img { width: 100%; aspect-ratio: 16/10; object-fit: cover; }
.ref-item strong { display: block; padding: 12px 16px; font-size: 15px; }
.result-card { background: #fff; border-radius: var(--radius); padding: 32px; position: sticky; top: 100px; }
.total { font-size: 48px; font-weight: 700; letter-spacing: -0.03em; margin: 12px 0 4px; }
.range { font-size: 14px; color: #5f584f; margin-bottom: 8px; }
.days { font-size: 14px; margin-bottom: 24px; }
.lines { border-top: 1px solid var(--line); }
.line { display: flex; justify-content: space-between; padding: 12px 0; border-bottom: 1px solid var(--line); font-size: 14px; }
.line b { font-weight: 600; }
.assumptions { margin-top: 20px; padding: 16px; background: #fef8f0; border-radius: 10px; font-size: 13px; }
.assumptions ul { margin: 8px 0 0 20px; }
@media (max-width: 900px) { .layout { grid-template-columns: 1fr; } .result-card { position: static; } }
</style>
