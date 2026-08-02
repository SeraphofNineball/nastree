<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { api } from '../api'
import type { TreeNode } from '../types'
import { formatBytes, formatPercent, formatDateFromUnix } from '../lib/format'
import { useSort, type SortDir } from '../lib/sort'

const props = defineProps<{
  totalSize: number
}>()

const emit = defineEmits<{ enter: [path: string] }>()

const query = ref('')
const matchPath = ref(false)
const foldersOnly = ref(false)
const duplicatesOnly = ref(false)
const dupMode = ref<'name_size' | 'name_size_date'>('name_size')
const results = ref<TreeNode[]>([])
const loading = ref(false)

async function fetchResults() {
  loading.value = true
  try {
    results.value = await api.files({
      q: query.value || undefined,
      matchPath: matchPath.value,
      foldersOnly: foldersOnly.value,
      duplicatesOnly: duplicatesOnly.value,
      dupMode: dupMode.value,
    })
  } catch {
    results.value = []
  } finally {
    loading.value = false
  }
}

let debounceHandle: ReturnType<typeof setTimeout> | undefined
function scheduleFetch(delay: number) {
  if (debounceHandle) clearTimeout(debounceHandle)
  debounceHandle = setTimeout(fetchResults, delay)
}
watch(query, () => scheduleFetch(250))
watch([matchPath, foldersOnly, duplicatesOnly, dupMode], () => scheduleFetch(0))
onMounted(fetchResults)

const resultsRef = computed(() => results.value)
function getter(n: TreeNode, key: string): string | number {
  switch (key) {
    case 'name':
      return n.name.toLowerCase()
    case 'path':
      return n.path.toLowerCase()
    case 'size':
      return n.size
    case 'modTime':
      return n.modTime
    case 'dupCount':
      return n.dupCount ?? 0
    case 'dupSize':
      return n.dupSize ?? 0
    default:
      return 0
  }
}
const { sortKey, sortDir, toggle, sorted } = useSort(resultsRef, getter, 'size', 'desc')

const columns: { key: string; label: string; defaultDir: SortDir; align?: 'left'; sortable?: boolean }[] = [
  { key: 'name', label: 'File Name', defaultDir: 'asc', align: 'left' },
  { key: 'path', label: 'Path', defaultDir: 'asc', align: 'left' },
  { key: 'pct', label: '% of Drive', defaultDir: 'desc', sortable: false },
  { key: 'size', label: 'Size', defaultDir: 'desc' },
  { key: 'modTime', label: 'Modified', defaultDir: 'desc' },
  { key: 'dupCount', label: 'Dup Count', defaultDir: 'desc' },
  { key: 'dupSize', label: 'Dup Size', defaultDir: 'desc' },
]
</script>

<template>
  <div class="file-search">
    <div class="controls">
      <label class="radio">
        <input type="radio" name="matchMode" :checked="!matchPath" @change="matchPath = false" />
        Match file name only
      </label>
      <label class="radio">
        <input type="radio" name="matchMode" :checked="matchPath" @change="matchPath = true" />
        Match entire path
      </label>

      <div class="search-box">
        <span class="search-icon">🔍</span>
        <input v-model="query" type="text" placeholder="Search files…" />
        <button v-if="query" class="clear-btn" @click="query = ''">Clear</button>
      </div>

      <label class="checkbox">
        <input v-model="foldersOnly" type="checkbox" />
        Folders only
      </label>

      <span class="dup-label">Duplicate Files:</span>
      <select v-model="dupMode">
        <option value="name_size">Name &amp; Size</option>
        <option value="name_size_date">Name, Size &amp; Date</option>
      </select>
      <label class="checkbox">
        <input v-model="duplicatesOnly" type="checkbox" />
        Duplicates only
      </label>

      <span v-if="loading" class="loading-indicator">Searching…</span>
    </div>

    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th
              v-for="col in columns"
              :key="col.key"
              :class="{ 'name-col': col.align === 'left', sortable: col.sortable !== false, active: sortKey === col.key }"
              @click="col.sortable !== false && toggle(col.key, col.defaultDir)"
            >
              {{ col.label }}
              <span v-if="col.sortable !== false && sortKey === col.key" class="arrow">{{ sortDir === 'asc' ? '▲' : '▼' }}</span>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="n in sorted" :key="n.path" :class="{ clickable: n.isDir }" @click="n.isDir && emit('enter', n.path)">
            <td class="name-col">
              <span class="icon">{{ n.isDir ? '📁' : '📄' }}</span>{{ n.name }}
            </td>
            <td class="path-col" :title="n.path">{{ n.path }}</td>
            <td class="num">{{ formatPercent(n.size, totalSize) }}</td>
            <td class="num">{{ formatBytes(n.size) }}</td>
            <td>{{ formatDateFromUnix(n.modTime) }}</td>
            <td class="num">{{ n.dupCount ? n.dupCount.toLocaleString() : '' }}</td>
            <td class="num">{{ n.dupSize ? formatBytes(n.dupSize) : '' }}</td>
          </tr>
          <tr v-if="!loading && !sorted.length">
            <td colspan="7" class="empty">No files match.</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.file-search {
  display: flex;
  flex-direction: column;
  gap: 8px;
  height: 100%;
}
.controls {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 14px;
  padding: 8px 10px;
  background: var(--surface-1);
  border: 1px solid var(--border);
  border-radius: 6px;
  font-size: 12px;
  color: var(--text-secondary);
  flex-shrink: 0;
}
.radio,
.checkbox {
  display: flex;
  align-items: center;
  gap: 5px;
  cursor: pointer;
  white-space: nowrap;
}
.search-box {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1;
  min-width: 180px;
  background: var(--surface-2);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 4px 8px;
}
.search-icon {
  font-size: 11px;
  opacity: 0.7;
}
.search-box input {
  flex: 1;
  background: none;
  border: none;
  outline: none;
  color: var(--text-primary);
  font-size: 12px;
}
.clear-btn {
  border: none;
  background: none;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 11px;
  padding: 2px 4px;
}
.clear-btn:hover {
  color: var(--text-primary);
}
.dup-label {
  white-space: nowrap;
}
select {
  background: var(--surface-2);
  border: 1px solid var(--border);
  border-radius: 6px;
  color: var(--text-primary);
  padding: 4px 6px;
  font-size: 12px;
}
.loading-indicator {
  color: var(--text-muted);
  font-style: italic;
}
.table-wrap {
  overflow: auto;
  flex: 1;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--surface-1);
}
table {
  font-size: 13px;
  width: 100%;
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
  max-width: 240px;
  overflow: hidden;
  text-overflow: ellipsis;
}
td.path-col {
  text-align: left;
  max-width: 360px;
  overflow: hidden;
  text-overflow: ellipsis;
  color: var(--text-secondary);
  font-family: ui-monospace, Consolas, monospace;
  font-size: 12px;
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
.empty {
  text-align: center;
  color: var(--text-muted);
  padding: 24px;
}
</style>
