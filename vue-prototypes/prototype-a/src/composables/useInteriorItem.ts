/**
 * Загрузка интерьера по идентификатору маршрута.
 *
 * Список интерьеров приходит одним запросом, поэтому для карточек-страниц
 * достаточно найти элемент в сторе; если его там нет, догружаем по id.
 */
import { computed, onMounted, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useRoute } from 'vue-router'
import { useInteriorsStore } from '@/stores/interiors'
import type { InteriorItem } from '@/types'

export function useInteriorItem() {
  const route = useRoute()
  const interiors = useInteriorsStore()
  const { loading, error } = storeToRefs(interiors)

  const id = computed(() => String(route.params.id ?? ''))
  const item = computed<InteriorItem | null>(() => interiors.byId(id.value) ?? interiors.selected)

  function load() {
    if (!id.value) return
    if (interiors.byId(id.value)) return
    interiors.fetchById(id.value)
  }

  onMounted(() => {
    interiors.fetch()
    load()
  })
  watch(id, () => load())

  return { item, loading, error }
}
