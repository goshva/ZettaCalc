<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useCprStore } from '@/stores/cpr'
import { formatPrice } from '@/api/client'
import { imgUrl } from '@/composables/useImg'
import { cprStatusLabel, docStatusLabel, docTypeLabel } from '@/utils/labels'

const cpr = useCprStore()
const { loading, error } = storeToRefs(cpr)
const projectId = ref('proj_001')

function generate() {
  cpr.generate(projectId.value)
}

onMounted(() => cpr.fetchBids())
</script>

<template>
  <section class="section">
    <div class="wrap">
      <h2>Цифровой паспорт ремонта</h2>
      <p class="muted" style="margin-bottom: 32px">Платный пакет документов: чертежи, ведомости, состав работ</p>

      <div class="action">
        <input v-model="projectId" class="input" placeholder="ID проекта" style="max-width: 240px" />
        <button class="btn red" :disabled="loading" @click="generate">
          {{ loading ? 'Генерация…' : 'Сгенерировать ЦПР' }}
        </button>
      </div>

      <div v-if="error" class="error">{{ error }}</div>

      <div v-if="cpr.current" class="cpr-card">
        <div class="cpr-header">
          <h3>{{ cpr.current.id }} · v{{ cpr.current.version }}</h3>
          <span class="tag">{{ cprStatusLabel(cpr.current.status) }}</span>
        </div>

        <div class="cpr-estimate">
          <div class="total">{{ formatPrice(cpr.current.estimate.total_rub) }}</div>
          <div class="range">Диапазон: {{ formatPrice(cpr.current.estimate.range_min) }} — {{ formatPrice(cpr.current.estimate.range_max) }}</div>
        </div>

        <div class="docs">
          <h4>Документы</h4>
          <div v-for="d in cpr.current.documents" :key="d.id" class="doc-item">
            <span>{{ d.title }}</span>
            <span class="tag">{{ docTypeLabel(d.type) }} · {{ docStatusLabel(d.status) }}</span>
          </div>
        </div>

        <div class="bids">
          <h4>Предложения подрядчиков</h4>
          <div v-for="b in cpr.bids" :key="b.id" class="bid">
            <div class="bid-img">
              <img v-if="b.image" :src="imgUrl(b.image)" :alt="b.name" />
            </div>
            <div>
              <strong>{{ b.name }}</strong> · ⭐ {{ b.rating }}
              <p>{{ b.scope }}</p>
            </div>
            <div class="bid-price">
              <b>{{ formatPrice(b.price) }}</b>
              <small>{{ b.days }} дней</small>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.action { display: flex; gap: 12px; margin-bottom: 32px; align-items: center; }
.error { color: var(--red); background: #fef3f0; padding: 14px; border-radius: 8px; margin-bottom: 24px; }
.cpr-card { background: #fff; border: 1px solid var(--line); border-radius: var(--radius); padding: 28px; }
.cpr-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.total { font-size: 42px; font-weight: 700; }
.range { font-size: 14px; color: var(--muted); }
.docs, .bids { margin-top: 24px; }
.doc-item { display: flex; justify-content: space-between; padding: 12px 0; border-bottom: 1px solid var(--line); font-size: 14px; }
.bid { display: flex; align-items: center; gap: 16px; justify-content: space-between; padding: 16px 0; border-bottom: 1px solid var(--line); }
.bid-img { width: 64px; height: 48px; border-radius: 8px; overflow: hidden; flex-shrink: 0; background: #f2f0ec; }
.bid-img img { width: 100%; height: 100%; object-fit: cover; }
.bid-price { text-align: right; }
.bid-price b { font-size: 20px; }
.bid-price small { color: var(--muted); font-size: 13px; }
</style>
