<script setup lang="ts">
/** Переключатель лент «Интерьеров» — tabs(active) из оригинала */
import type { InteriorsKind } from '@/types'

defineProps<{ active: InteriorsKind }>()

const tabs: { kind: InteriorsKind; title: string }[] = [
  { kind: 'rooms', title: 'Интерьеры со сметой' },
  { kind: 'apartments', title: 'Квартиры со сметой' },
  { kind: 'ideas', title: 'Идеи интерьеров' },
]
</script>

<template>
  <div class="pills">
    <RouterLink v-for="t in tabs" :key="t.kind" class="pill" :class="{ active: active === t.kind }"
                :to="`/interiors/${t.kind}`">
      {{ t.title }}
    </RouterLink>
  </div>
</template>
