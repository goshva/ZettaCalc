<script setup lang="ts">
/**
 * Навигация по помещениям квартиры — apartmentNavigation(x) и apNav51.
 *
 * Показывается в карточке помещения, которое входит в квартиру.
 */
import { computed } from 'vue'
import type { InteriorItem } from '@/types'
import { useInteriorsStore } from '@/stores/interiors'

const props = defineProps<{ item: InteriorItem }>()

const interiors = useInteriorsStore()
const parent = computed(() => (props.item.parent ? interiors.byId(props.item.parent) : null))
const rooms = computed(() => (parent.value ? interiors.childrenOf(parent.value) : []))
</script>

<template>
  <nav v-if="parent" class="apartmentnav" aria-label="Помещения квартиры">
    <RouterLink class="backapt" :to="`/interiors/apartment/${parent.id}`">
      ← Вернуться к квартире {{ parent.area }} м²
    </RouterLink>
    <div>
      <RouterLink v-for="r in rooms" :key="r.id" class="pill" :class="{ active: r.id === item.id }"
                  :to="`/interiors/room/${r.id}`">
        {{ r.room }}
      </RouterLink>
    </div>
  </nav>
</template>
