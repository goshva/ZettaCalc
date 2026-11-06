<script setup lang="ts">
/**
 * FlowPage.vue — Сквозной путь: Бриф → Геометрия → Расчёт → ЦПР → Заказ и аукцион
 *
 * Детерминированный расчёт по тем же правилам, что и src/domain/estimate:
 *  - площади → техкарты → объёмы → цены из PIM (заглушки)
 *  - допущения → неопределённость (3% за допущение, кап 25%)
 *  - статус ЦПР: ready / pending_review
 *
 * Шаг «Заказ и аукцион» воспроизводит продуктовую границу из AGENTS.md:
 * write-операции (заказ ЦПР, передача ТЗ) выполняются только после
 * подтверждения человеком, ставки подрядчиков считаются от сметы.
 *
 * ?step=N открывает шаг напрямую — нужно для проверки вёрстки на широких экранах.
 */

import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useBriefStore, useGeometryStore, useCalcStore, useCprStore } from '@/stores'
import { formatPrice } from '@/api/client'
import { docTypeLabel, docStatusLabel } from '@/utils/labels'

/* ════════════ Типы ════════════ */
interface BriefFields {
  area: number
  rooms: string
  family: string
  timeline: string
  style: string
  floorMaterial: string
  unevenness: string
}

interface Planner {
  id: string
  name: string
  area: number
  rooms: { name: string; area: number }[]
  ceilingHeight: number
}

interface EstimateLine {
  sku: string
  name: string
  qty: number
  unit: string
  price: number
  category: 'material' | 'work'
}

interface Estimate {
  lines: EstimateLine[]
  total: number
  rangeMin: number
  rangeMax: number
  days: number
  assumptions: string[]
}

interface Variant {
  id: string
  label: string
  emoji: string
  multiplier: number
}

interface CprDocument {
  title: string
  type: string
  status: string
}

interface CprPackage {
  id: string
  label: string
  price: number
  note: string
}

interface Bid {
  id: string
  name: string
  rating: number
  price: number
  days: number
  scope: string
}

/* ════════════ Данные ════════════ */
const PLANNERS: Planner[] = [
  {
    id: 'p1', name: 'Квартира 90 м² — стандартная', area: 90,
    ceilingHeight: 2.7,
    rooms: [
      { name: 'Кухня-гостиная', area: 28 },
      { name: 'Спальня', area: 18 },
      { name: 'Детская', area: 14 },
      { name: 'Санузел', area: 6 },
      { name: 'Прихожая', area: 10 },
    ],
  },
  {
    id: 'p2', name: 'Квартира 90 м² — с перепланировкой', area: 92,
    ceilingHeight: 2.7,
    rooms: [
      { name: 'Кухня-гостиная', area: 35 },
      { name: 'Спальня', area: 18 },
      { name: 'Кабинет', area: 12 },
      { name: 'Санузел совмещённый', area: 8 },
      { name: 'Прихожая', area: 9 },
    ],
  },
  {
    id: 'p3', name: 'Квартира 60 м² — студия', area: 60,
    ceilingHeight: 3.0,
    rooms: [
      { name: 'Кухня-гостиная', area: 30 },
      { name: 'Спальня', area: 16 },
      { name: 'Санузел', area: 5 },
      { name: 'Прихожая', area: 9 },
    ],
  },
]

const VARIANTS: Variant[] = [
  { id: 'economy', label: 'Эконом', emoji: '🌿', multiplier: 0.85 },
  { id: 'balanced', label: 'Баланс', emoji: '⚖️', multiplier: 1.0 },
  { id: 'fast', label: 'Быстро', emoji: '⚡', multiplier: 1.1 },
  { id: 'premium', label: 'Премиум', emoji: '✨', multiplier: 1.45 },
]

const MATERIALS: Record<string, { name: string; price: number }> = {
  keramogranit: { name: 'Керамогранит 600×600', price: 1850 },
  laminate: { name: 'Ламинат 33 класс', price: 950 },
  parquet: { name: 'Паркетная доска', price: 2400 },
  tile: { name: 'Керамическая плитка', price: 1200 },
}

const WORKS: Record<string, { name: string; price: number }> = {
  prep: { name: 'Подготовка основания', price: 450 },
  install: { name: 'Укладка покрытия', price: 850 },
  grout: { name: 'Затирка швов', price: 180 },
  plaster: { name: 'Штукатурка стен', price: 620 },
  paint: { name: 'Покраска стен (2 слоя)', price: 380 },
  primer: { name: 'Грунтовка', price: 85 },
}

const CPR_PACKAGES: CprPackage[] = [
  { id: 'basic', label: 'Базовый ЦПР', price: 29000, note: 'Чертежи, ведомости, состав работ, календарный план' },
  { id: 'options', label: 'ЦПР с опциями', price: 49000, note: 'Базовый + электрика, сантехника, спецификация по SKU' },
]

const STEPS = ['Бриф', 'Геометрия', 'Расчёт', 'ЦПР', 'Заказ и аукцион']
const MAX_STEP = STEPS.length - 1

/* ════════════ Утилиты ════════════ */
const fmt = (n: number): string =>
  new Intl.NumberFormat('ru-RU', {
    style: 'currency',
    currency: 'RUB',
    maximumFractionDigits: 0,
  }).format(n)

/* ════════════ Детерминированный расчёт ════════════ */
/*
 * Соответствует логике src/domain/estimate/__init__.py:
 *  - build_estimate(items, catalog, assumptions)
 *  - UNCERTAINTY_PER_ASSUMPTION_PCT = 3
 *  - UNCERTAINTY_CAP_PCT = 25
 *  - total = sum(qty * price)
 *  - low = total * (1 - pct/100)
 *  - high = total * (1 + pct/100)
 */
function calcEstimate(
  area: number,
  floorMaterial: string,
  unevenness: string,
  variantMultiplier: number,
): Estimate {
  const mat = MATERIALS[floorMaterial] || MATERIALS.keramogranit
  const matQty = Math.round(area * 1.05 * 10) / 10 // 5% запас
  const lines: EstimateLine[] = []

  // Материалы
  lines.push({
    sku: 'MAT-FLOOR-' + floorMaterial.toUpperCase(),
    name: mat.name,
    qty: matQty,
    unit: 'м²',
    price: Math.round(mat.price * variantMultiplier),
    category: 'material',
  })

  // Работы (соответствуют операциям техкарты FLOOR-KER-600)
  lines.push({ sku: 'WRK-PREP', name: WORKS.prep.name, qty: matQty, unit: 'м²', price: Math.round(WORKS.prep.price * variantMultiplier), category: 'work' })
  lines.push({ sku: 'WRK-INSTALL', name: WORKS.install.name, qty: matQty, unit: 'м²', price: Math.round(WORKS.install.price * variantMultiplier), category: 'work' })
  lines.push({ sku: 'WRK-GROUT', name: WORKS.grout.name, qty: matQty, unit: 'м²', price: Math.round(WORKS.grout.price * variantMultiplier), category: 'work' })

  // Стены (для всех комнат: S_floor × 2.8)
  const wallArea = Math.round(area * 2.8 * 10) / 10
  lines.push({ sku: 'WRK-PLASTER', name: WORKS.plaster.name, qty: wallArea, unit: 'м²', price: Math.round(WORKS.plaster.price * variantMultiplier), category: 'work' })
  lines.push({ sku: 'WRK-PAINT', name: WORKS.paint.name, qty: wallArea, unit: 'м²', price: Math.round(WORKS.paint.price * variantMultiplier), category: 'work' })
  lines.push({ sku: 'WRK-PRIMER', name: WORKS.primer.name, qty: wallArea, unit: 'м²', price: Math.round(WORKS.primer.price * variantMultiplier), category: 'work' })

  const total = lines.reduce((s, l) => s + l.qty * l.price, 0)

  // Допущения (как в build_cpr: floor_unevenness === unknown)
  const assumptions: string[] = []
  if (unevenness === 'unknown') {
    assumptions.push('Ровность пола требует проверки на объекте')
    assumptions.push('Состояние стен неизвестно — возможен доп. расход')
  }

  // Неопределённость: как в src/domain/estimate — 3% за допущение, кап 25%
  const uncPct = Math.min(assumptions.length * 3, 25)
  const rangeMin = Math.round(total * (1 - uncPct / 100))
  const rangeMax = Math.round(total * (1 + uncPct / 100))

  // Сроки: ~15 рабочих дней на 15 м²
  const days = Math.round((matQty * 0.8 + wallArea * 0.6) / 15) * 15

  return { lines, total: Math.round(total), rangeMin, rangeMax, days, assumptions }
}

/* ════════════ Состояние ════════════ */
const route = useRoute()

const requestedStep = Math.trunc(Number(route.query.step ?? 0))
const step = ref(
  Number.isFinite(requestedStep) ? Math.min(Math.max(requestedStep, 0), MAX_STEP) : 0,
)

const brief = ref<BriefFields>({
  area: 90,
  rooms: '3',
  family: 'family',
  timeline: '3-6',
  style: 'modern',
  floorMaterial: 'keramogranit',
  unevenness: 'unknown',
})

const selectedPlanner = ref<Planner | null>(step.value > 0 ? PLANNERS[0] : null)
const selectedVariant = ref<Variant>(VARIANTS[1]) // баланс

// Единый источник правды. Материал пола выбирается один раз: чипы на шаге
// «Геометрия» пишут в то же поле брифа, а не в отдельное состояние.
// Площадь уточняется выбранной планировкой: пока планировка не выбрана,
// считаем по площади из брифа.
const effectiveArea = computed(() =>
  selectedPlanner.value ? selectedPlanner.value.area : brief.value.area
)

const roomsSum = computed(() =>
  selectedPlanner.value
    ? Math.round(selectedPlanner.value.rooms.reduce((s, r) => s + r.area, 0) * 10) / 10
    : 0
)

// Пересчёт при любом изменении исходных данных
const estimate = computed(() =>
  calcEstimate(
    effectiveArea.value,
    brief.value.floorMaterial,
    brief.value.unevenness,
    selectedVariant.value.multiplier,
  )
)

// ЦПР: статус зависит от допущений
const cprStatus = computed(() =>
  estimate.value.assumptions.length > 0 ? 'pending_review' : 'ready'
)

const cprDocuments = computed<CprDocument[]>(() => [
  { title: 'План помещений', type: 'drawing', status: 'ready' },
  { title: 'Развёртка стен', type: 'drawing', status: 'ready' },
  { title: 'Ведомость материалов', type: 'specification', status: 'ready' },
  {
    title: 'Календарный план',
    type: 'schedule',
    status: cprStatus.value === 'ready' ? 'ready' : 'pending',
  },
])

// Комнаты для ЦПР
const cprRooms = computed(() =>
  selectedPlanner.value
    ? selectedPlanner.value.rooms.map(r => ({
        name: r.name,
        area: r.area,
        floorMaterial: MATERIALS[brief.value.floorMaterial]?.name || brief.value.floorMaterial,
        ceilingHeight: selectedPlanner.value!.ceilingHeight,
        works: ['Подготовка основания', 'Укладка покрытия', 'Затирка швов'],
      }))
    : []
)

/* ════════════ Заказ ЦПР и аукцион ════════════ */
const selectedPackage = ref<CprPackage>(CPR_PACKAGES[0])
const orderPlaced = ref(false)
const orderNumber = ref('')
const selectedBidId = ref<string | null>(null)
const contractorConfirmed = ref(false)

// Ставки считаются от сметы — как _make_bids в mock API: цена и срок
// привязаны к смете варианта, а не заданы константами.
const bids = computed<Bid[]>(() => {
  const base = estimate.value.total
  const days = estimate.value.days
  const rooms = selectedPlanner.value?.rooms.map(r => r.name).join(', ') || 'по ЦПР'
  return [
    { id: 'b1', name: 'РемонтПро', rating: 4.8, price: Math.round(base * 0.95), days: days + 5, scope: `Полный ремонт: ${rooms}` },
    { id: 'b2', name: 'СтройМастер', rating: 4.5, price: Math.round(base * 0.88), days: days + 10, scope: `Чистовой ремонт: ${rooms}` },
    { id: 'b3', name: 'ДизайнРемонт', rating: 4.9, price: Math.round(base * 1.10), days: Math.max(days - 5, 15), scope: `Дизайн + ремонт: ${rooms}` },
  ]
})

// Лучшее предложение — минимальная цена; при равенстве берём первый по списку.
const bestBidId = computed(() => {
  const sorted = [...bids.value].sort((a, b) => a.price - b.price)
  return sorted[0]?.id ?? null
})

const selectedBid = computed(() =>
  bids.value.find(b => b.id === selectedBidId.value) ?? null
)

const orderTotal = computed(() => selectedPackage.value.price + estimate.value.total)

function placeOrder() {
  orderPlaced.value = true
  orderNumber.value = 'ORD-2026-000128'
  selectedBidId.value = null
  contractorConfirmed.value = false
}

function chooseBid(id: string) {
  selectedBidId.value = id
  contractorConfirmed.value = false
}

function confirmContractor() {
  if (selectedBidId.value) contractorConfirmed.value = true
}

/* ════════════ Навигация ════════════ */
function nextStep() {
  if (step.value < MAX_STEP) step.value++
}

function prevStep() {
  if (step.value > 0) step.value--
}

/* ════════════ Валидация шагов ════════════ */
const canGoFromBrief = computed(
  () => brief.value.area > 0 && brief.value.rooms !== '' && brief.value.floorMaterial !== ''
)

const canGoFromGeometry = computed(() => selectedPlanner.value !== null)

</script>

<template>
  <!-- Обёртка как у остальных страниц B: контейнер .wrap задаёт ширину,
       поэтому .flow-wrap здесь отвечает только за вертикальные отступы. -->
  <section class="section">
    <div class="wrap">
      <div class="flow-page">
        <div class="flow-wrap">
      <!-- Шаги -->
      <div class="flow-steps">
        <div v-for="(s, i) in STEPS" :key="i"
             :class="['flow-step', i < step ? 'done' : '', i === step ? 'active' : '']">
          <span class="num">{{ i < step ? '✓' : i + 1 }}</span>{{ s }}
        </div>
      </div>

      <!-- Шаг 0: Бриф -->
      <div v-if="step === 0" class="fade-in">
        <div class="flow-card">
          <h2 class="flow-h2">Бриф клиента</h2>
          <p class="flow-muted" style="margin-bottom: 20px">
            Большая часть полей уже предзаполнена. Ответьте на ключевые вопросы —
            остальное система поставит как разумные допущения.
          </p>

          <div class="g2">
            <div class="flow-field">
              <label>Площадь квартиры, м²</label>
              <input type="number" :value="brief.area"
                     @input="brief.area = parseFloat(($event.target as HTMLInputElement).value) || 0" />
              <div class="hint">Целевой объект: 90 м²</div>
            </div>
            <div class="flow-field">
              <label>Количество комнат</label>
              <select :value="brief.rooms" @change="brief.rooms = ($event.target as HTMLSelectElement).value">
                <option value="">Выберите…</option>
                <option value="1">Студия</option>
                <option value="2">1 комната</option>
                <option value="3">2 комнаты</option>
                <option value="4">3 комнаты</option>
                <option value="5+">4+ комнаты</option>
              </select>
            </div>
            <div class="flow-field">
              <label>Состав семьи</label>
              <select :value="brief.family" @change="brief.family = ($event.target as HTMLSelectElement).value">
                <option value="">Выберите…</option>
                <option value="single">Один / пара без детей</option>
                <option value="family">Семья с детьми</option>
                <option value="multi">Несколько поколений</option>
              </select>
            </div>
            <div class="flow-field">
              <label>Желаемые сроки</label>
              <select :value="brief.timeline" @change="brief.timeline = ($event.target as HTMLSelectElement).value">
                <option value="">Выберите…</option>
                <option value="1-2">1–2 месяца</option>
                <option value="3-6">3–6 месяцев</option>
                <option value="6-12">6–12 месяцев</option>
                <option value="no-rush">Без спешки</option>
              </select>
            </div>
          </div>

          <div class="flow-field">
            <label>Предпочтительный стиль</label>
            <select :value="brief.style" @change="brief.style = ($event.target as HTMLSelectElement).value">
              <option value="modern">Современный</option>
              <option value="scandinavian">Скандинавский</option>
              <option value="classic">Классика</option>
              <option value="minimal">Минимализм</option>
            </select>
          </div>

          <div class="flow-field">
            <label>Материал пола</label>
            <select :value="brief.floorMaterial" @change="brief.floorMaterial = ($event.target as HTMLSelectElement).value">
              <option value="">Выберите…</option>
              <option value="keramogranit">Керамогранит 600×600</option>
              <option value="laminate">Ламинат 33 класс</option>
              <option value="parquet">Паркетная доска</option>
              <option value="tile">Керамическая плитка</option>
            </select>
          </div>

          <div class="flow-field">
            <label>Ровность основания</label>
            <select :value="brief.unevenness" @change="brief.unevenness = ($event.target as HTMLSelectElement).value">
              <option value="">Выберите…</option>
              <option value="known-good">Ровное (≤ 2 мм/м)</option>
              <option value="known-bad">Неровное (&gt; 5 мм/м)</option>
              <option value="unknown">Не знаю — требует проверки</option>
            </select>
          </div>
        </div>

        <div class="flow-actions flow-actions--end">
          <button class="flow-btn flow-btn--red" :disabled="!canGoFromBrief" @click="nextStep">
            Далее: Геометрия →
          </button>
        </div>
      </div>

      <!-- Шаг 1: Геометрия -->
      <div v-if="step === 1" class="fade-in">
        <div class="flow-card">
          <h2 class="flow-h2">Выберите планировку</h2>
          <p class="flow-muted" style="margin-bottom: 20px">
            Система предлагает подходящие планировки из библиотеки. После выбора расчёт уточнится.
          </p>

          <div class="g3" style="margin-bottom: 28px">
            <button v-for="p in PLANNERS" :key="p.id" type="button"
                    :class="['planner-card', selectedPlanner?.id === p.id ? 'sel' : '']"
                    :aria-pressed="selectedPlanner?.id === p.id"
                    @click="selectedPlanner = p">
              <span class="img">📐</span>
              <span class="info">
                <span class="name">{{ p.name }}</span>
                <span class="area">{{ p.area }} м² · {{ p.rooms.length }} комн. · h={{ p.ceilingHeight }}м</span>
              </span>
            </button>
          </div>

          <h3 class="flow-h3">Материал пола</h3>
          <div class="mat-chips">
            <button v-for="(m, k) in MATERIALS" :key="k"
                    :class="['mat-chip', brief.floorMaterial === k ? 'sel' : '']"
                    @click="brief.floorMaterial = k">
              {{ m.name }} · от {{ fmt(m.price) }}/м²
            </button>
          </div>
          <p class="flow-muted" style="font-size: 12px; margin-top: 8px">
            Выбор меняет техкарту, материалы, объём работ и срок — расчёт пересчитается целиком.
          </p>
        </div>

        <!-- Комнаты: геометрия, из которой считаются объёмы -->
        <div v-if="selectedPlanner" class="flow-card flow-card-sm">
          <h3 class="flow-h3" style="font-size: 15px">Геометрия выбранной планировки</h3>
          <div style="display: flex; gap: 6px; flex-wrap: wrap; margin-top: 10px">
            <span v-for="r in selectedPlanner.rooms" :key="r.name"
                  class="badge" style="font-size: 11px">
              {{ r.name }}: {{ r.area }} м²
            </span>
          </div>
          <p class="flow-muted" style="font-size: 12px; margin-top: 12px">
            Комнаты: {{ roomsSum }} м² · всего по планировке: {{ selectedPlanner.area }} м²
            · разница {{ Math.round((selectedPlanner.area - roomsSum) * 10) / 10 }} м² —
            коридоры и общие зоны, они входят в расчёт полов.
          </p>
          <p class="flow-muted" style="font-size: 12px; margin-top: 4px">
            Высота потолка {{ selectedPlanner.ceilingHeight }} м — по ней считаются стены ({{ Math.round(selectedPlanner.area * 2.8 * 10) / 10 }} м²).
          </p>
        </div>

        <div class="flow-actions">
          <button class="flow-btn flow-btn--ghost" @click="prevStep">← Назад</button>
          <button class="flow-btn flow-btn--red" :disabled="!canGoFromGeometry" @click="nextStep">
            Далее: Расчёт →
          </button>
        </div>
      </div>

      <!-- Шаг 2: Расчёт -->
      <div v-if="step === 2" class="fade-in">
        <div class="calc-grid">
          <div>
            <div class="flow-card">
              <h2 class="flow-h2">Предварительный расчёт</h2>
              <p class="flow-muted" style="margin-bottom: 6px">
                Бесплатно · со SKU, брендами и диапазонами
              </p>
              <div class="divider">
                <span class="badge badge--red">
                  {{ selectedVariant.emoji }} {{ selectedVariant.label }} ×{{ selectedVariant.multiplier }}
                </span>
                <div class="line" />
              </div>

              <table class="est-table">
                <thead>
                  <tr>
                    <th>Позиция</th>
                    <th class="n">Кол-во</th>
                    <th class="n">Цена</th>
                    <th class="n">Сумма</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="l in estimate.lines" :key="l.sku">
                    <td>
                      <div>{{ l.name }}</div>
                      <div class="cat" :class="l.category === 'material' ? 'cat-material' : 'cat-work'">
                        {{ l.sku }}
                      </div>
                    </td>
                    <td class="n">{{ l.qty }} {{ l.unit }}</td>
                    <td class="n">{{ fmt(l.price) }}</td>
                    <td class="n">{{ fmt(l.qty * l.price) }}</td>
                  </tr>
                  <tr class="total-row">
                    <td>Итого</td>
                    <td></td>
                    <td></td>
                    <td class="n">{{ fmt(estimate.total) }}</td>
                  </tr>
                </tbody>
              </table>

              <!-- Допущения -->
              <div v-if="estimate.assumptions.length" class="assumptions">
                <div class="title">⚠ Требуют проверки:</div>
                <ul>
                  <li v-for="(a, i) in estimate.assumptions" :key="i">{{ a }}</li>
                </ul>
              </div>
            </div>

            <div class="flow-actions">
              <button class="flow-btn flow-btn--ghost" @click="prevStep">← Назад</button>
              <button class="flow-btn flow-btn--red" @click="nextStep">Получить ЦПР →</button>
            </div>
          </div>

          <!-- Правая колонка: итого + варианты -->
          <div>
            <div class="flow-card total-box">
              <div class="label">Оценка</div>
              <div class="value">{{ fmt(estimate.total) }}</div>
              <div class="range">{{ fmt(estimate.rangeMin) }} — {{ fmt(estimate.rangeMax) }}</div>
              <div class="days">~{{ estimate.days }} рабочих дней</div>
            </div>

            <div class="flow-card flow-card-sm">
              <h3 class="flow-h3" style="font-size: 15px">Варианты реализации</h3>
              <div class="g4" style="margin-top: 10px">
                <button v-for="v in VARIANTS" :key="v.id" type="button"
                        :class="['var-card', selectedVariant.id === v.id ? 'sel' : '']"
                        :aria-pressed="selectedVariant.id === v.id"
                        @click="selectedVariant = v">
                  <span class="emoji">{{ v.emoji }}</span>
                  <span class="label">{{ v.label }}</span>
                  <span class="desc">×{{ v.multiplier }}</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Шаг 3: ЦПР -->
      <div v-if="step === 3" class="fade-in">
        <div class="flow-card">
          <div class="cpr-head">
            <div>
              <h2 class="flow-h2">Цифровой паспорт ремонта</h2>
              <p class="flow-muted">
                CPR-2026-000001 · v1
                <span class="badge badge--red" style="font-size: 9px; margin-left: 8px">
                  {{ cprStatus === 'ready' ? 'Готов' : 'На проверке' }}
                </span>
              </p>
            </div>
            <div class="cpr-sum">
              <div class="cpr-sum-value">{{ fmt(estimate.total) }}</div>
              <div class="cpr-sum-range">{{ fmt(estimate.rangeMin) }} — {{ fmt(estimate.rangeMax) }}</div>
            </div>
          </div>

          <div class="g2">
            <!-- Документы -->
            <div>
              <h3 class="flow-h3" style="font-size: 17px; margin-bottom: 12px">Документы</h3>
              <div v-for="(d, i) in cprDocuments" :key="i" class="cpr-doc">
                <div>
                  <div class="title">{{ d.title }}</div>
                  <div class="type">{{ docTypeLabel(d.type) }}</div>
                </div>
                <span :class="['status', d.status === 'ready' ? 's-ready' : 's-pending']">
                  {{ docStatusLabel(d.status) }}
                </span>
              </div>
            </div>

            <!-- Комнаты -->
            <div>
              <h3 class="flow-h3" style="font-size: 17px; margin-bottom: 12px">Комнаты</h3>
              <div v-for="(r, i) in cprRooms" :key="i" class="cpr-doc">
                <div>
                  <div class="title">{{ r.name }}</div>
                  <div class="type">{{ r.area }} м² · {{ r.floorMaterial }}</div>
                </div>
                <div style="font-size: 12px">{{ r.works.length }} операций</div>
              </div>
            </div>
          </div>

          <!-- Допущения -->
          <div v-if="estimate.assumptions.length" class="assumptions">
            <div class="title">⚠ Допущения (требуют подтверждения):</div>
            <ul>
              <li v-for="(a, i) in estimate.assumptions" :key="i">{{ a }}</li>
            </ul>
          </div>
        </div>

        <div class="flow-actions">
          <button class="flow-btn flow-btn--ghost" @click="prevStep">← Назад</button>
          <button class="flow-btn flow-btn--red" @click="nextStep">Заказать ЦПР и начать аукцион →</button>
        </div>
      </div>

      <!-- Шаг 4: Заказ ЦПР и аукцион -->
      <div v-if="step === 4" class="fade-in">
        <div class="flow-card">
          <h2 class="flow-h2">Заказ ЦПР</h2>
          <p class="flow-muted" style="margin-bottom: 20px">
            ЦПР — платный выход пути: документ становится единым ТЗ для подрядчиков.
          </p>

          <div class="order-grid">
            <div>
              <h3 class="flow-h3" style="font-size: 17px; margin-bottom: 12px">Пакет</h3>
              <div class="g2">
                <button v-for="p in CPR_PACKAGES" :key="p.id" type="button"
                        :class="['pack-card', selectedPackage.id === p.id ? 'sel' : '']"
                        :aria-pressed="selectedPackage.id === p.id"
                        @click="selectedPackage = p">
                  <span class="name">{{ p.label }}</span>
                  <span class="note">{{ p.note }}</span>
                  <span class="price">{{ fmt(p.price) }}</span>
                </button>
              </div>
            </div>

            <div>
              <h3 class="flow-h3" style="font-size: 17px; margin-bottom: 12px">Заказ</h3>
              <div class="sum-row"><span>Документ</span><b>CPR-2026-000001 · v1</b></div>
              <div class="sum-row"><span>Смета по ЦПР</span><b>{{ fmt(estimate.total) }}</b></div>
              <div class="sum-row"><span>Диапазон</span><b>{{ fmt(estimate.rangeMin) }} — {{ fmt(estimate.rangeMax) }}</b></div>
              <div class="sum-row"><span>{{ selectedPackage.label }}</span><b>{{ fmt(selectedPackage.price) }}</b></div>
              <div class="sum-row sum-row--total"><span>К оплате сейчас</span><b>{{ fmt(selectedPackage.price) }}</b></div>
              <p class="flow-muted" style="font-size: 12px; margin-top: 12px">
                Заказ создаётся только после подтверждения человеком; повторная отправка
                с тем же ключом возвращает прежний результат.
              </p>
            </div>
          </div>

          <div class="flow-actions flow-actions--end">
            <button class="flow-btn flow-btn--red" :disabled="orderPlaced" @click="placeOrder">
              {{ orderPlaced ? 'ЦПР заказан' : 'Заказать ЦПР и начать аукцион →' }}
            </button>
          </div>
        </div>

        <!-- Аукцион -->
        <div v-if="orderPlaced" class="flow-card">
          <div class="cpr-head">
            <div>
              <h2 class="flow-h2">Аукцион подрядчиков</h2>
              <p class="flow-muted">
                Заказ {{ orderNumber }} · ЦПР оплачен · ставки принимаются 3 дня
              </p>
            </div>
            <span class="badge badge--red">
              {{ contractorConfirmed ? 'Подрядчик выбран' : 'Аукцион открыт' }}
            </span>
          </div>

          <div class="g3 bids-grid">
            <div v-for="b in bids" :key="b.id"
                 :class="['bid-card', selectedBidId === b.id ? 'sel' : '']">
              <div class="top">
                <span class="name">{{ b.name }}</span>
                <span class="rating">⭐ {{ b.rating }}</span>
              </div>
              <div class="price">{{ fmt(b.price) }}</div>
              <div class="meta">{{ b.days }} дней · {{ b.scope }}</div>
              <span v-if="bestBidId === b.id" class="badge" style="align-self: flex-start">Лучшая цена</span>
              <button class="flow-btn flow-btn--ghost" :aria-pressed="selectedBidId === b.id"
                      @click="chooseBid(b.id)">
                {{ selectedBidId === b.id ? 'Выбран' : 'Выбрать' }}
              </button>
            </div>
          </div>

          <div class="sum-row sum-row--total">
            <span>Смета по ЦПР + {{ selectedPackage.label }}</span>
            <b>{{ fmt(orderTotal) }}</b>
          </div>

          <div v-if="contractorConfirmed && selectedBid" class="order-status">
            Подрядчик выбран: {{ selectedBid.name }} · {{ fmt(selectedBid.price) }} · {{ selectedBid.days }} дней.
            ТЗ передано, заказ {{ orderNumber }} передан в работу.
          </div>

          <div class="flow-actions flow-actions--end">
            <button class="flow-btn flow-btn--red" :disabled="!selectedBidId || contractorConfirmed"
                    @click="confirmContractor">
              {{ contractorConfirmed ? 'Заказ передан подрядчику' : 'Подтвердить и передать ТЗ →' }}
            </button>
          </div>
        </div>

        <div class="flow-actions">
          <button class="flow-btn flow-btn--ghost" @click="prevStep">← Назад</button>
        </div>
      </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style>
.flow-page{--f-ink:var(--ink,#1E1C19);--f-accent:var(--red,var(--gold,#B53A27));--f-paper:var(--paper,#FBF9F5);--f-line:var(--line,#D9D1C5);--f-muted:var(--muted,#5F584F);--f-on-ink:var(--on-ink,#F4F0EA);--f-soft:var(--soft,#EAE4DA);--f-radius:var(--radius,12px);--f-max:1320px;font-family:Manrope,system-ui,sans-serif;color:var(--f-ink)}

.flow-page { font-family: Manrope, system-ui, sans-serif; color: var(--f-ink); }
.flow-page .serif { font-family: 'Cormorant Garamond', Georgia, serif; font-weight: 500 }

/* Контейнер: ширину задаёт .wrap снаружи, здесь только вертикальные отступы */
.flow-page .flow-wrap { width: 100%; padding: 0 0 20px }

/* Шаги */
.flow-page .flow-steps { display: flex; gap: 0; margin-bottom: 40px; border-bottom: 1px solid var(--f-line) }
.flow-page .flow-step { flex: 1; text-align: center; padding: 14px 8px; font-size: 13px; font-weight: 600; cursor: default; min-width: 0 }
.flow-page .flow-step.done { color: var(--f-accent) }
.flow-page .flow-step.active { color: var(--f-ink); border-bottom: 2px solid var(--f-accent) }
.flow-page .flow-step .num { display: inline-block; width: 26px; height: 26px; line-height: 26px; border-radius: 50%; background: var(--f-soft); margin-right: 6px; font-size: 12px; vertical-align: middle }
.flow-page .flow-step.done .num { background: var(--f-accent); color: #fff }
.flow-page .flow-step.active .num { background: var(--f-ink); color: #fff }

/* Карточки */
.flow-page .flow-card { background: var(--f-paper); border: 1px solid var(--f-line); border-radius: 12px; padding: 28px; margin-bottom: 20px }
.flow-page .flow-card-sm { padding: 16px }

/* Кнопки */
.flow-page .flow-btn { display: inline-flex; align-items: center; justify-content: center; gap: 8px; min-height: 42px; padding: 0 22px; background: var(--f-ink); color: var(--f-on-ink); font-size: 14px; font-weight: 600; border: 1px solid var(--f-ink); border-radius: 999px; cursor: pointer; transition: all 0.2s; font-family: inherit }
.flow-page .flow-btn:hover { background: var(--f-accent); border-color: var(--f-accent) }
.flow-page .flow-btn:disabled { opacity: 0.4; cursor: not-allowed }
.flow-page .flow-btn--red { background: var(--f-accent); border-color: var(--f-accent) }
.flow-page .flow-btn--red:hover { background: var(--f-ink); border-color: var(--f-ink) }
.flow-page .flow-btn--ghost { background: transparent; color: var(--f-ink) }
.flow-page .flow-btn--ghost:hover { background: var(--f-ink); color: var(--f-on-ink) }

/* Ряд кнопок навигации */
.flow-page .flow-actions { display: flex; justify-content: space-between; gap: 12px; margin-top: 20px; flex-wrap: wrap }
.flow-page .flow-actions--end { justify-content: flex-end }

/* Поля */
.flow-page .flow-field { margin-bottom: 18px }
.flow-page .flow-field label { display: block; font-size: 11px; font-weight: 600; letter-spacing: 0.08em; text-transform: uppercase; color: var(--f-muted); margin-bottom: 6px }
.flow-page .flow-field input, .flow-page .flow-field select { width: 100%; height: 46px; border: 0; border-bottom: 1px solid var(--f-ink); background: transparent; font-family: 'Cormorant Garamond', serif; font-size: 22px; color: var(--f-ink); padding: 4px 0; outline: none; transition: border-color 0.2s }
.flow-page .flow-field input:focus, .flow-page .flow-field select:focus { border-bottom-color: var(--f-accent); border-bottom-width: 2px }
.flow-page .flow-field .hint { font-size: 11px; color: var(--f-muted); margin-top: 3px }

/* Сетка */
.flow-page .g2 { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 20px }
.flow-page .g3 { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 14px }
.flow-page .g4 { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 10px }

/* Расчёт: смета и итог */
.flow-page .calc-grid { display: grid; grid-template-columns: 1.3fr 1fr; gap: 28px; align-items: start }

/* Планировки */
.flow-page .planner-card { cursor: pointer; border: 2px solid transparent; border-radius: 10px; overflow: hidden; transition: all 0.2s; background: var(--f-paper); display: block; text-align: left; font: inherit; color: inherit; padding: 0; width: 100% }
.flow-page .planner-card:hover { border-color: var(--f-ink) }
.flow-page .planner-card.sel { border-color: var(--f-accent); box-shadow: 0 0 0 1px var(--f-accent) }
.flow-page .planner-card .img { aspect-ratio: 16/10; background: var(--f-soft); display: flex; align-items: center; justify-content: center; font-size: 40px }
.flow-page .planner-card .info { display: block; padding: 14px }
.flow-page .planner-card .name { display: block; font-family: 'Cormorant Garamond', serif; font-size: 18px; font-weight: 500 }
.flow-page .planner-card .area { display: block; font-size: 12px; color: var(--f-muted) }

/* Варианты */
.flow-page .var-card { cursor: pointer; border: 2px solid transparent; border-radius: 10px; padding: 16px; text-align: center; transition: all 0.2s; background: var(--f-paper); display: block; font: inherit; color: inherit; width: 100% }
.flow-page .var-card:hover { border-color: var(--f-ink) }
.flow-page .var-card.sel { border-color: var(--f-accent); box-shadow: 0 0 0 1px var(--f-accent) }
.flow-page .var-card .emoji { display: block; font-size: 28px; margin-bottom: 6px }
.flow-page .var-card .label { display: block; font-weight: 600; font-size: 14px }
.flow-page .var-card .desc { display: block; font-size: 11px; color: var(--f-muted); margin-top: 3px }

/* Пакеты ЦПР */
.flow-page .pack-card { cursor: pointer; border: 2px solid transparent; border-radius: 10px; padding: 18px; text-align: left; transition: all 0.2s; background: var(--f-paper); display: block; font: inherit; color: inherit; width: 100% }
.flow-page .pack-card:hover { border-color: var(--f-ink) }
.flow-page .pack-card.sel { border-color: var(--f-accent); box-shadow: 0 0 0 1px var(--f-accent) }
.flow-page .pack-card .name { display: block; font-family: 'Cormorant Garamond', serif; font-size: 19px; font-weight: 500 }
.flow-page .pack-card .note { display: block; font-size: 12px; color: var(--f-muted); margin-top: 6px; line-height: 1.5 }
.flow-page .pack-card .price { display: block; font-size: 15px; font-weight: 600; margin-top: 12px }

/* Заказ: сводка */
.flow-page .order-grid { display: grid; grid-template-columns: 1.15fr 1fr; gap: 28px; align-items: start }
.flow-page .sum-row { display: flex; justify-content: space-between; align-items: baseline; gap: 16px; padding: 10px 0; border-bottom: 1px solid var(--f-line); font-size: 14px }
.flow-page .sum-row:last-of-type { border-bottom: 0 }
.flow-page .sum-row b { font-weight: 600; white-space: nowrap }
.flow-page .sum-row--total { border-top: 2px solid var(--f-ink); border-bottom: 0; font-size: 16px; margin-top: 6px; padding-top: 14px }
.flow-page .order-status { margin-top: 18px; padding: 14px 16px; border-radius: 10px; background: #E8F5E9; color: #2E7D32; font-size: 13px; font-weight: 600; line-height: 1.5 }

/* Ставки подрядчиков */
.flow-page .bids-grid { margin-top: 18px; align-items: stretch }
.flow-page .bid-card { display: flex; flex-direction: column; gap: 8px; border: 1px solid var(--f-line); border-radius: 10px; padding: 16px; background: var(--f-paper); transition: all 0.2s }
.flow-page .bid-card.sel { border-color: var(--f-accent); box-shadow: 0 0 0 1px var(--f-accent) }
.flow-page .bid-card .top { display: flex; justify-content: space-between; align-items: baseline; gap: 10px }
.flow-page .bid-card .name { font-weight: 600; font-size: 15px }
.flow-page .bid-card .rating { font-size: 12px; color: var(--f-muted); white-space: nowrap }
.flow-page .bid-card .price { font-family: 'Cormorant Garamond', serif; font-size: 26px; font-weight: 500; line-height: 1.1 }
.flow-page .bid-card .meta { font-size: 12px; color: var(--f-muted); line-height: 1.5; flex: 1 }

/* Материал-чипы */
.flow-page .mat-chips { display: flex; gap: 6px; flex-wrap: wrap; margin: 10px 0 }
.flow-page .mat-chip { padding: 8px 14px; border: 1px solid var(--f-line); border-radius: 999px; font-size: 12px; cursor: pointer; transition: all 0.2s; background: transparent; font-family: inherit; color: inherit }
.flow-page .mat-chip:hover { border-color: var(--f-ink) }
.flow-page .mat-chip.sel { background: var(--f-ink); color: var(--f-on-ink); border-color: var(--f-ink) }

/* Таблица расчёта */
.flow-page .est-table { width: 100%; border-collapse: collapse; font-size: 13px }
.flow-page .est-table th { text-align: left; font-size: 10px; font-weight: 600; letter-spacing: 0.14em; text-transform: uppercase; color: var(--f-muted); padding: 0 10px 10px; border-bottom: 1px solid var(--f-ink) }
.flow-page .est-table td { padding: 10px; border-bottom: 1px solid var(--f-line); vertical-align: middle }
.flow-page .est-table td.n { text-align: right; font-weight: 600; white-space: nowrap }
.flow-page .est-table .total-row td { font-weight: 600; font-size: 16px; padding-top: 14px; border-top: 2px solid var(--f-ink) }
.flow-page .est-table .cat { font-size: 10px; color: var(--f-muted); text-transform: uppercase; letter-spacing: 0.1em }

/* Блок итого */
.flow-page .total-box { text-align: center; padding: 28px }
.flow-page .total-box .label { font-size: 11px; font-weight: 600; letter-spacing: 0.14em; text-transform: uppercase; color: var(--f-muted) }
.flow-page .total-box .value { font-family: 'Cormorant Garamond', serif; font-size: 56px; font-weight: 500; letter-spacing: -0.02em; line-height: 1.1 }
.flow-page .total-box .range { font-size: 13px; color: var(--f-muted); margin-top: 3px }
.flow-page .total-box .days { font-size: 13px; margin-top: 6px }

/* Шапка ЦПР и аукциона */
.flow-page .cpr-head { display: flex; justify-content: space-between; align-items: flex-start; gap: 24px; margin-bottom: 20px }
.flow-page .cpr-sum { text-align: right; flex-shrink: 0 }
.flow-page .cpr-sum-value { font-family: 'Cormorant Garamond', serif; font-size: 32px; font-weight: 500; line-height: 1.1 }
.flow-page .cpr-sum-range { font-size: 11px; color: var(--f-muted) }

/* Допущения */
.flow-page .assumptions { background: #FFF8F0; border-left: 3px solid var(--f-accent); padding: 14px; margin: 14px 0; border-radius: 0 6px 6px 0 }
.flow-page .assumptions .title { font-size: 12px; font-weight: 600; color: var(--f-accent); margin-bottom: 6px }
.flow-page .assumptions ul { margin: 0; padding-left: 18px }
.flow-page .assumptions li { font-size: 12px; color: var(--f-muted); margin-bottom: 3px }

/* Документы ЦПР */
.flow-page .cpr-doc { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 14px; border: 1px solid var(--f-line); border-radius: 8px; margin-bottom: 6px }
.flow-page .cpr-doc .title { font-weight: 600; font-size: 13px }
.flow-page .cpr-doc .type { font-size: 10px; color: var(--f-muted); text-transform: uppercase; letter-spacing: 0.1em }
.flow-page .cpr-doc .status { padding: 3px 10px; border-radius: 999px; font-size: 10px; font-weight: 600; white-space: nowrap }
.flow-page .s-ready { background: #E8F5E9; color: #2E7D32 }
.flow-page .s-pending { background: #FCE4EC; color: #C62828 }

/* Бейдж */
.flow-page .badge { display: inline-block; padding: 3px 10px; border-radius: 999px; font-size: 10px; font-weight: 600; letter-spacing: 0.1em; text-transform: uppercase; background: var(--f-ink); color: var(--f-on-ink) }
.flow-page .badge--red { background: var(--f-accent) }

/* Заголовки */
.flow-page .flow-h2 { font-family: 'Cormorant Garamond', serif; font-size: 28px; font-weight: 500; margin-bottom: 6px }
.flow-page .flow-h3 { font-family: 'Cormorant Garamond', serif; font-size: 20px; font-weight: 500; margin-bottom: 6px }
.flow-page .flow-muted { color: var(--f-muted); font-size: 14px }

.flow-page .divider { display: flex; align-items: center; gap: 10px; margin-bottom: 20px }
.flow-page .divider .line { flex: 1; height: 1px; background: var(--f-line) }

@keyframes flowFadeIn { from { opacity: 0; transform: translateY(6px) } to { opacity: 1; transform: translateY(0) } }
.flow-page .fade-in { animation: flowFadeIn 0.25s ease }

/* Узкие экраны */
@media (max-width: 1100px) {
  .flow-page .calc-grid, .flow-page .order-grid { grid-template-columns: 1fr; gap: 20px }
}
@media (max-width: 900px) {
  .flow-page .g2, .flow-page .g3, .flow-page .g4 { grid-template-columns: 1fr }
  .flow-page .total-box .value { font-size: 40px }
  .flow-page .flow-steps { flex-wrap: wrap }
  .flow-page .flow-step { flex: 1 1 50%; border-bottom: 1px solid var(--f-line) }
  .flow-page .cpr-head { flex-direction: column }
  .flow-page .cpr-sum { text-align: left }
  .flow-page .flow-actions { flex-direction: column-reverse }
  .flow-page .flow-actions .flow-btn { width: 100% }
  .flow-page .flow-actions--end { flex-direction: column }
}
@media (max-width: 600px) {
  .flow-page .flow-wrap { width: 100%; padding: 0 0 30px }
  .flow-page .flow-card { padding: 18px }
}
</style>
