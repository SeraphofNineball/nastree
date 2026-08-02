<script setup lang="ts">
import { computed } from 'vue'
import type { FileTypeStat } from '../types'
import {
  buildExtColorMap,
  colorForExt,
  customExtColors,
  setExtColor,
  resetExtColor,
  resetAllExtColors,
  directoryColor,
  customFolderColor,
  setFolderColor,
  resetFolderColor,
  customGradientTop,
  setGradientTop,
  resetGradientTop,
  customGradientBottom,
  setGradientBottom,
  resetGradientBottom,
} from '../lib/colors'
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

const hasCustomColors = computed(
  () =>
    Object.keys(customExtColors).length > 0 ||
    customFolderColor.value !== null ||
    customGradientTop.value !== null ||
    customGradientBottom.value !== null,
)
function onPick(ext: string, e: Event) {
  setExtColor(ext, (e.target as HTMLInputElement).value)
}
function onPickFolder(e: Event) {
  setFolderColor((e.target as HTMLInputElement).value)
}
function onPickGradientTop(e: Event) {
  setGradientTop((e.target as HTMLInputElement).value)
}
function onPickGradientBottom(e: Event) {
  setGradientBottom((e.target as HTMLInputElement).value)
}
function resetAllColors() {
  resetAllExtColors()
  resetFolderColor()
  resetGradientTop()
  resetGradientBottom()
}

const gradientTopValue = computed(() => customGradientTop.value ?? '#ffffff')
const gradientBottomValue = computed(() => customGradientBottom.value ?? (currentTheme.value === 'light' ? '#808080' : '#000000'))

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
          <th class="reset-col">
            <button v-if="hasCustomColors" class="reset-all-btn" title="Reset all custom colors" @click="resetAllColors">
              Reset colors
            </button>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr class="folder-row">
          <td class="name-col">
            <input
              type="color"
              class="swatch-picker"
              :value="directoryColor()"
              title="Pick a color for folders"
              @input="onPickFolder"
            />
            Folders
          </td>
          <td class="num">—</td>
          <td class="num">—</td>
          <td class="num">—</td>
          <td class="reset-col">
            <button v-if="customFolderColor !== null" class="reset-one-btn" title="Reset to default color" @click="resetFolderColor">
              ↺
            </button>
          </td>
        </tr>
        <tr class="folder-row">
          <td class="name-col">
            <input
              type="color"
              class="swatch-picker"
              :value="gradientTopValue"
              title="Pick the box gradient's highlight (top) color"
              @input="onPickGradientTop"
            />
            Gradient top
          </td>
          <td class="num">—</td>
          <td class="num">—</td>
          <td class="num">—</td>
          <td class="reset-col">
            <button v-if="customGradientTop !== null" class="reset-one-btn" title="Reset to default color" @click="resetGradientTop">
              ↺
            </button>
          </td>
        </tr>
        <tr class="folder-row">
          <td class="name-col">
            <input
              type="color"
              class="swatch-picker"
              :value="gradientBottomValue"
              title="Pick the box gradient's shadow (bottom) color"
              @input="onPickGradientBottom"
            />
            Gradient bottom
          </td>
          <td class="num">—</td>
          <td class="num">—</td>
          <td class="num">—</td>
          <td class="reset-col">
            <button v-if="customGradientBottom !== null" class="reset-one-btn" title="Reset to default color" @click="resetGradientBottom">
              ↺
            </button>
          </td>
        </tr>
        <tr v-for="ft in sorted" :key="ft.ext">
          <td class="name-col">
            <input
              type="color"
              class="swatch-picker"
              :value="colorForExt(ft.ext, colorMap)"
              :title="`Pick a color for ${ft.ext}`"
              @input="onPick(ft.ext, $event)"
            />
            {{ ft.ext }}
          </td>
          <td class="num">{{ formatBytes(ft.size) }}</td>
          <td class="num">{{ ft.count.toLocaleString() }}</td>
          <td class="num">{{ formatPercent(ft.size, total) }}</td>
          <td class="reset-col">
            <button v-if="ft.ext in customExtColors" class="reset-one-btn" title="Reset to default color" @click="resetExtColor(ft.ext)">
              ↺
            </button>
          </td>
        </tr>
        <tr v-if="!fileTypes.length">
          <td colspan="5" class="empty">No data yet.</td>
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
.swatch-picker {
  width: 16px;
  height: 16px;
  margin-right: 8px;
  padding: 0;
  border: 1px solid var(--border);
  border-radius: 3px;
  background: none;
  cursor: pointer;
  vertical-align: -3px;
}
.swatch-picker::-webkit-color-swatch-wrapper {
  padding: 1px;
}
.swatch-picker::-webkit-color-swatch {
  border: none;
  border-radius: 2px;
}
.reset-col {
  width: 1%;
  padding-left: 4px;
  padding-right: 8px;
}
.reset-all-btn,
.reset-one-btn {
  background: none;
  border: 1px solid var(--border);
  border-radius: 4px;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 11px;
  padding: 2px 6px;
  white-space: nowrap;
}
.reset-all-btn:hover,
.reset-one-btn:hover {
  color: var(--text-primary);
  background: var(--surface-2);
}
.reset-one-btn {
  padding: 1px 5px;
  font-size: 12px;
}
.folder-row td {
  border-bottom: 1px solid var(--border);
  background: var(--surface-2);
}
.empty {
  text-align: center;
  color: var(--text-muted);
  padding: 24px;
}
</style>
