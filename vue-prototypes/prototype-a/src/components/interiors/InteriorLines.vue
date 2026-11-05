<script setup lang="ts">
/**
 * Состав сметы построчно — перенос lineRows(lines) и groupedLines(lines).
 *
 * grouped=true повторяет раскрывающиеся разделы сметы: группа, её сумма
 * и первые четыре позиции, остальные — во вложенном раскрытии.
 */
import { computed } from 'vue'
import type { RoomLine } from '@/types'
import { useInteriorsStore } from '@/stores/interiors'
import { groupLines, groupTotal, lineHref, lineTotal, money } from '@/utils/interiors'

const props = withDefaults(defineProps<{ lines: RoomLine[]; grouped?: boolean }>(), { grouped: false })

const interiors = useInteriorsStore()
const tier = computed(() => interiors.tier)
const groups = computed(() => groupLines(props.lines))
</script>

<template>
  <div v-if="!grouped">
    <div v-for="(l, i) in lines" :key="l.name + i" class="lineitem">
      <div>
        <strong>{{ l.name }}</strong>
        <p>{{ l.description }}</p>
        <small>{{ l.qty }}<template v-if="l.note"> · {{ l.note }}</template></small>
        <a v-if="l.url" :href="lineHref(l, tier)" target="_blank" rel="noopener">Карточка товара ↗</a>
        <small v-if="tier && l.altUrl">Альтернатива этого варианта — по ссылке</small>
      </div>
      <b>{{ money(lineTotal(l, tier)) }}</b>
    </div>
  </div>

  <template v-else>
    <details v-for="g in groups" :key="g.name">
      <summary>
        <span>{{ g.name }}</span>
        <b>{{ money(groupTotal(g.items, tier)) }} +</b>
      </summary>
      <div v-for="(l, i) in g.items.slice(0, 4)" :key="l.name + i" class="lineitem">
        <div>
          <strong>{{ l.name }}</strong>
          <p>{{ l.description }}</p>
          <small>{{ l.qty }}<template v-if="l.note"> · {{ l.note }}</template></small>
          <a v-if="l.url" :href="lineHref(l, tier)" target="_blank" rel="noopener">Карточка товара ↗</a>
          <small v-if="tier && l.altUrl">Альтернатива этого варианта — по ссылке</small>
        </div>
        <b>{{ money(lineTotal(l, tier)) }}</b>
      </div>
      <details v-if="g.items.length > 4">
        <summary>Показать все позиции ({{ g.items.length }})</summary>
        <div v-for="(l, i) in g.items.slice(4)" :key="l.name + i" class="lineitem">
          <div>
            <strong>{{ l.name }}</strong>
            <p>{{ l.description }}</p>
            <small>{{ l.qty }}<template v-if="l.note"> · {{ l.note }}</template></small>
            <a v-if="l.url" :href="lineHref(l, tier)" target="_blank" rel="noopener">Карточка товара ↗</a>
          </div>
          <b>{{ money(lineTotal(l, tier)) }}</b>
        </div>
      </details>
    </details>
  </template>
</template>
