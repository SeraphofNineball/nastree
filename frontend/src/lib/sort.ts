import { computed, ref, type Ref } from 'vue'

export type SortDir = 'asc' | 'desc'

/** Shared click-to-sort behavior for table headers: tracks the active key/direction
 *  and exposes a sorted copy of the input list. */
export function useSort<T>(items: Ref<T[]>, getter: (item: T, key: string) => string | number, initialKey: string, initialDir: SortDir = 'desc') {
  const sortKey = ref(initialKey)
  const sortDir = ref<SortDir>(initialDir)

  function toggle(key: string, defaultDir: SortDir = 'desc') {
    if (sortKey.value === key) {
      sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
    } else {
      sortKey.value = key
      sortDir.value = defaultDir
    }
  }

  const sorted = computed(() => {
    const dir = sortDir.value === 'asc' ? 1 : -1
    return [...items.value].sort((a, b) => {
      const av = getter(a, sortKey.value)
      const bv = getter(b, sortKey.value)
      if (typeof av === 'string' || typeof bv === 'string') {
        return String(av).localeCompare(String(bv)) * dir
      }
      return (av - bv) * dir
    })
  })

  return { sortKey, sortDir, toggle, sorted }
}
