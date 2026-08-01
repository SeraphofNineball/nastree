<script setup lang="ts">
import { computed } from 'vue'
import type { FileTypeStat } from '../types'
import { buildExtColorMap, colorForExt } from '../lib/colors'
import { formatBytes, formatPercent } from '../lib/format'

const props = defineProps<{
  fileTypes: FileTypeStat[]
}>()

const total = computed(() => props.fileTypes.reduce((sum, f) => sum + f.size, 0))
const colorMap = computed(() => buildExtColorMap(props.fileTypes))
</script>

<template>
  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th class="name-col">Extension</th>
          <th class="num">Percent</th>
          <th class="num">Size</th>
          <th class="num">Files</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="ft in fileTypes" :key="ft.ext">
          <td class="name-col">
            <span class="swatch" :style="{ background: colorForExt(ft.ext, colorMap) }" />
            {{ ft.ext }}
          </td>
          <td class="num">{{ formatPercent(ft.size, total) }}</td>
          <td class="num">{{ formatBytes(ft.size) }}</td>
          <td class="num">{{ ft.count.toLocaleString() }}</td>
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
