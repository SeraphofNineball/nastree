<script setup lang="ts">
import { computed } from 'vue'
import type { TreeNode } from '../types'
import { formatBytes, formatPercent, formatDateFromUnix } from '../lib/format'
import { useSort, type SortDir } from '../lib/sort'

const props = defineProps<{
  children: TreeNode[]
}>()

const emit = defineEmits<{ enter: [path: string] }>()

const parentTotal = computed(() => props.children.reduce((sum, n) => sum + n.size, 0))

const childrenRef = computed(() => props.children)
function getter(n: TreeNode, key: string): string | number {
  switch (key) {
    case 'name':
      return n.name.toLowerCase()
    case 'size':
      return n.size
    case 'items':
      return n.files + n.dirs
    case 'files':
      return n.files
    case 'dirs':
      return n.dirs
    case 'modTime':
      return n.modTime
    default:
      return 0
  }
}
const { sortKey, sortDir, toggle, sorted } = useSort(childrenRef, getter, 'size', 'desc')

const columns: { key: string; label: string; defaultDir: SortDir; align?: 'left' }[] = [
  { key: 'name', label: 'Name', defaultDir: 'asc', align: 'left' },
  { key: 'size', label: 'Size', defaultDir: 'desc' },
  { key: 'items', label: 'Items', defaultDir: 'desc' },
  { key: 'files', label: 'Files', defaultDir: 'desc' },
  { key: 'dirs', label: 'Folders', defaultDir: 'desc' },
  { key: 'modTime', label: 'Modified', defaultDir: 'desc' },
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
          <th class="pct-col">% of Parent</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="n in sorted"
          :key="n.path"
          :class="{ clickable: n.isDir }"
          @click="n.isDir && emit('enter', n.path)"
        >
          <td class="name-col">
            <span class="icon">{{ n.isDir ? '📁' : '📄' }}</span>{{ n.name }}
          </td>
          <td class="num">{{ formatBytes(n.size) }}</td>
          <td class="num">{{ n.isDir ? n.files + n.dirs : '' }}</td>
          <td class="num">{{ n.isDir ? n.files : '' }}</td>
          <td class="num">{{ n.isDir ? n.dirs : '' }}</td>
          <td>{{ formatDateFromUnix(n.modTime) }}</td>
          <td>
            <div class="pct-bar">
              <div class="pct-fill" :style="{ width: formatPercent(n.size, parentTotal) }" />
              <span class="pct-label">{{ formatPercent(n.size, parentTotal) }}</span>
            </div>
          </td>
        </tr>
        <tr v-if="!children.length">
          <td colspan="7" class="empty">This folder has no items.</td>
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
  max-width: 260px;
  overflow: hidden;
  text-overflow: ellipsis;
}
.icon {
  margin-right: 6px;
}
tr.clickable {
  cursor: pointer;
}
tr.clickable:hover td {
  background: var(--surface-2);
}
.pct-bar {
  position: relative;
  width: 100px;
  height: 14px;
  background: var(--gridline);
  border-radius: 3px;
  overflow: hidden;
  display: inline-block;
}
.pct-fill {
  position: absolute;
  inset: 0;
  right: auto;
  background: var(--series-1);
}
.pct-label {
  position: relative;
  font-size: 11px;
  line-height: 14px;
  color: var(--text-secondary);
  padding: 0 4px;
}
.empty {
  text-align: center;
  color: var(--text-muted);
  padding: 24px;
}
</style>
