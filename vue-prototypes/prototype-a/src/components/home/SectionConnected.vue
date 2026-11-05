<script setup lang="ts">
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import { api, API_BASE } from '@/api/client'
const apiImg = (name: string) => `${API_BASE}/api/images/file/${name}`

interface Hotspot {
  id: string
  name: string
  category: string
  price: number
  unit: string
  image: string
  point: [number, number]
  external?: boolean
  url?: string
}

interface Analogue {
  id: string
  name: string
  short?: string
  price: number
  group?: string
  unit?: string
  image: string
}

const items = ref<Hotspot[]>([])
const selected = ref(0)
const similarOpen = ref(false)
const similar = ref<Analogue[]>([])
const roomImg = ref<HTMLImageElement | null>(null)
const roomBox = ref<HTMLElement | null>(null)
const positions = ref<{ left: string; top: string }[]>([])

const money = (p: { price: number; unit?: string }) =>
  new Intl.NumberFormat('ru-RU').format(p.price) + ' ₽ / ' + (p.unit || 'шт.')

onMounted(async () => {
  try {
    const { data } = await api.get<Hotspot[]>('/api/hotspot/data')
    items.value = data
    await nextTick()
    place()
  } catch {
    /* без данных секция не показывает точки */
  }
})

// Позиции точек считаем от вписанного в блок изображения (как place() в оригинале)
function place() {
  const img = roomImg.value, host = roomBox.value
  if (!img || !host || !img.naturalWidth) return
  const scale = Math.min(host.clientWidth / img.naturalWidth, host.clientHeight / img.naturalHeight)
  const w = img.naturalWidth * scale, h = img.naturalHeight * scale
  const x = (host.clientWidth - w) / 2, y = (host.clientHeight - h) / 2
  positions.value = items.value.map((it) => ({
    left: x + w * it.point[0] + 'px',
    top: y + h * it.point[1] + 'px',
  }))
}
onMounted(() => {
  window.addEventListener('resize', place)
  if (roomImg.value) roomImg.value.addEventListener('load', place)
})

const current = computed(() => items.value[selected.value])

// Товары, показанные в правой колонке превью — как в оригинале: 0,1,3,7
const previewIdx = [0, 1, 3, 7]

function draw(i: number) {
  selected.value = i
  if (similarOpen.value) loadSimilar()
}

// «Похожие товары»: фильтруем общий массив аналогов по признаку выбранной позиции
const analogues = ref<Analogue[]>([])
async function loadAnalogues() {
  try {
    const { data } = await api.get<Analogue[]>('/api/hotspot/analogues')
    analogues.value = data
  } catch { /* пусто */ }
}
onMounted(loadAnalogues)

function fits(p: Analogue): boolean {
  const n = p.name.toLowerCase()
  switch (selected.value) {
    case 0: return /диван/.test(n)
    case 1: return /подвесн/.test(n)
    case 2: return /ламинат|кварц|паркет|spc|плитк|керамогранит/.test(n)
    case 3: case 4: return /столик/.test(n)
    case 5: return /кресло/.test(n)
    case 6: return /под тв/i.test((p.short || '') + ' ' + n)
    case 7: return /ков[её]р/.test(n)
    case 8: return /^(бра|настенн)/.test(n)
    case 9: return /^цвет /.test(n)
    default: return false
  }
}

async function loadSimilar() {
  if (!analogues.value.length) await loadAnalogues()
  const seen = new Set<string>()
  similar.value = analogues.value.filter((p) => {
    const name = p.name.toLowerCase().replace(/\s+/g, ' ')
    if (p.id === current.value?.id || !fits(p) || seen.has(name)) return false
    seen.add(name)
    return true
  }).slice(0, 6)
}

watch(similarOpen, (open) => { if (open) loadSimilar() })

function swapIn(p: Analogue) {
  if (!current.value) return
  items.value[selected.value] = {
    ...items.value[selected.value],
    ...p,
    name: p.name.split(' (')[0],
    category: current.value.category,
    unit: p.unit || 'шт.',
  }
  similarOpen.value = false
}

function toggleSimilar() {
  similarOpen.value = !similarOpen.value
}
</script>

<template>
  <section class="connected section" id="connected" data-block="Связь решений">
    <div class="wrap">
      <div class="head">
        <div>
          <span class="eyebrow">
            <a class="route-number" href="#step-02">02</a> / От картинки — к конкретным решениям</span>
          <h2>Подборка точно к вашему интерьеру.</h2>
        </div>
        <p>Выбираете интерьер, который нравится. Мы собираем к нему мебель, свет и отделку из нашего каталога —
          с учётом вашей планировки и бюджета. Предметы можно заменить, сохранив общий образ.</p>
      </div>
      <p class="connected-explain61">Отдельные решения становятся одним проектом: проверяем размеры, сочетания и стоимость.
        Нажмите на точки, чтобы посмотреть примеры товаров для этой подборки.</p>
      <div class="connected-grid">
        <div class="room" ref="roomBox">
          <img ref="roomImg" :src="apiImg('a_main_img_0009.jpg')"
               alt="Гостиная «Тёплый баланс» — референс с сайта ZETTA HOME">
          <button v-for="(it, i) in items" :key="it.id" class="hotspot"
                  :class="{ active: selected === i }"
                  :style="positions[i]"
                  :aria-label="`${it.category}: посмотреть товар`"
                  :aria-pressed="selected === i"
                  @click="draw(i)" @focus="draw(i)"
                  @pointerenter="(e) => e.pointerType === 'mouse' && draw(i)">+</button>
          <span class="room-label">Наведите или нажмите на точку</span>
        </div>
        <div class="selection63">
          <div class="selection-heading63">
            <span class="eyebrow">Примеры товаров из каталога</span>
          </div>
          <div class="selection-items63">
            <button v-for="n in previewIdx" :key="n" class="selection-item63"
                    :class="{ active: selected === n }"
                    :aria-pressed="selected === n"
                    :aria-label="`${items[n]?.category}: посмотреть товар`"
                    @click="draw(n)">
              <img :src="items[n]?.image" :alt="items[n]?.name" loading="lazy">
              <span class="selection-name63">{{ items[n]?.category }}</span>
              <span class="selection-price63">{{ items[n] ? money(items[n]) : '' }}</span>
            </button>
          </div>
        </div>
      </div>
      <div class="connect-details connect-product60 selected-product63" v-if="current">
        <span class="eyebrow red">{{ current.category }}</span>
        <img :src="current.image" :alt="current.name">
        <h3>{{ current.name }}</h3>
        <p>Товар из каталога для подбора к интерьеру.</p>
        <strong>{{ money(current) }}</strong>
        <div class="selected-actions64">
          <RouterLink class="btn" :to="`/catalog/product/${current.id}`">Подробнее о товаре ↗</RouterLink>
          <button class="btn light" type="button"
                  :aria-controls="similarOpen ? 'similarPanel64' : undefined"
                  :aria-expanded="similarOpen"
                  @click="toggleSimilar">Похожие товары</button>
        </div>
        <p class="connected-note60">Пример подборки к интерьеру. Предметы могут отличаться от изображённых;
          точное исполнение и актуальную стоимость уточняем при подборе.</p>
      </div>
      <section v-show="similarOpen" id="similarPanel64" class="similar-panel64"
               aria-labelledby="similarTitle64" tabindex="-1">
        <div class="similar-head64">
          <div class="similar-title64">
            <span class="eyebrow">Варианты из каталога</span>
            <h3 id="similarTitle64">Похожие товары · {{ current?.category }}</h3>
          </div>
          <button type="button" aria-label="Закрыть похожие товары" @click="similarOpen = false">×</button>
        </div>
        <p class="similar-intro64">Выберите вариант, чтобы посмотреть его в подборке.</p>
        <div class="similar-list64">
          <p v-if="!similar.length" class="similar-empty64">Посмотрите другие модели этой категории в каталоге.</p>
          <article class="similar-card64" v-for="p in similar" :key="p.id">
            <img :src="p.image" :alt="p.name" loading="lazy">
            <h4>{{ p.name.split(' (')[0] }}</h4>
            <p class="similar-price64">{{ money(p) }}</p>
            <button type="button" class="btn light" @click="swapIn(p)">Посмотреть в подборке</button>
          </article>
        </div>
        <RouterLink class="link" to="/catalog">Все товары этой категории ↗</RouterLink>
      </section>
    </div>
  </section>
</template>
