<script setup lang="ts">
import { computed } from 'vue'
import type { FileTypeStat } from '../types'
import { buildExtColorMap, colorForExt } from '../lib/colors'
import { formatBytes, formatPercent } from '../lib/format'
import { currentTheme } from '../lib/theme'
import { useSort, type SortDir } from '../lib/sort'

const props = defineProps<{
  fileTypes: FileTypeStat[]
}>()

const total = computed(() => props.fileTypes.reduce((sum, f) => sum + f.size, 0))
const colorMap = computed(() => {
  void currentTheme.value
  return buildExtColorMap(props.fileTypes)
})

const fileTypesRef = computed(() => props.fileTypes)
function getter(ft: FileTypeStat, key: string): string | number {
  switch (key) {
    case 'ext':
      return ft.ext.toLowerCase()
    case 'size':
      return ft.size
    case 'count':
      return ft.count
    default:
      return 0
  }
}
const { sortKey, sortDir, toggle, sorted } = useSort(fileTypesRef, getter, 'size', 'desc')

const columns: { key: string; label: string; defaultDir: SortDir; align?: 'left' }[] = [
  { key: 'ext', label: 'Extension', defaultDir: 'asc', align: 'left' },
  { key: 'size', label: 'Size', defaultDir: 'desc' },
  { key: 'count', label: 'Files', defaultDir: 'desc' },
]
</script>

<template>
  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th
            v-for="col in columns"
            :key="col.key"
            :class="{ 'name-col': col.align === 'left', sortable: true, active: sortKey === col.key }"
            @click="toggle(col.key, col.defaultDir)"
          >
            {{ col.label }}
            <span v-if="sortKey === col.key" class="arrow">{{ sortDir === 'asc' ? '▲' : '▼' }}</span>
          </th>
          <th class="num">Percent</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="ft in sorted" :key="ft.ext">
          <td class="name-col">
            <span class="swatch" :style="{ background: colorForExt(ft.ext, colorMap) }" />
            {{ ft.ext }}
          </td>
          <td class="num">{{ formatBytes(ft.size) }}</td>
          <td class="num">{{ ft.count.toLocaleString() }}</td>
          <td class="num">{{ formatPercent(ft.size, total) }}</td>
        </tr>
        <tr v-if="!fileTypes.length">
          <td colspan="4" class="empty">No data yet.</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.table-wrap {
  overflow: auto;
  height: 100%;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--surface-1);
}
table {
  font-size: 13px;
}
thead th {
  position: sticky;
  top: 0;
  background: var(--surface-1);
  text-align: right;
  padding: 8px 10px;
  font-weight: 600;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--gridline);
  white-space: nowrap;
}
thead th.name-col {
  text-align: left;
}
thead th.sortable {
  cursor: pointer;
  user-select: none;
}
thead th.sortable:hover {
  color: var(--text-primary);
}
thead th.active {
  color: var(--text-primary);
}
.arrow {
  font-size: 9px;
  margin-left: 2px;
}
td {
  padding: 6px 10px;
  text-align: right;
  border-bottom: 1px solid var(--gridline);
  white-space: nowrap;
  color: var(--text-primary);
}
td.name-col {
  text-align: left;
}
.swatch {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 2px;
  margin-right: 8px;
}
.empty {
  text-align: center;
  color: var(--text-muted);
  padding: 24px;
}
</style>
