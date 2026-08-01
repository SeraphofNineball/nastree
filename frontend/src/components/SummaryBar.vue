<script setup lang="ts">
import { computed } from 'vue'
import type { ScanStatus } from '../types'
import { formatBytes, formatPercent, formatDate, formatDuration } from '../lib/format'
import ThemeSwitcher from './ThemeSwitcher.vue'

const props = defineProps<{
  status: ScanStatus | null
  scanning: boolean
}>()

const emit = defineEmits<{ rescan: [] }>()

const used = computed(() => (props.status ? props.status.diskTotal - props.status.diskFree : 0))
</script>

<template>
  <div class="bar">
    <div class="field">
      <span class="label">Root</span>
      <span class="value mono">{{ status?.rootPath ?? '—' }}</span>
    </div>
    <div class="field" v-if="status">
      <span class="label">Scanned</span>
      <span class="value">{{ formatBytes(status.totalSize) }} &middot; {{ status.totalFiles.toLocaleString() }} files &middot; {{ status.totalDirs.toLocaleString() }} folders</span>
    </div>
    <div class="field" v-if="status?.diskTotal">
      <span class="label">Disk</span>
      <span class="value">{{ formatBytes(used) }} used / {{ formatBytes(status.diskTotal) }} ({{ formatPercent(used, status.diskTotal) }}) &middot; {{ formatBytes(status.diskFree) }} free</span>
    </div>
    <div class="field" v-if="status?.finishedAt">
      <span class="label">Last scan</span>
      <span class="value">{{ formatDate(status.finishedAt) }} ({{ formatDuration(status.durationMs) }})</span>
    </div>
    <div class="spacer" />
    <ThemeSwitcher />
    <button class="rescan" :disabled="scanning" @click="emit('rescan')">
      {{ scanning ? 'Scanning…' : 'Rescan' }}
    </button>
  </div>
</template>

<style scoped>
.bar {
  display: flex;
  align-items: center;
  gap: 24px;
  padding: 10px 16px;
  background: var(--surface-1);
  border: 1px solid var(--border);
  border-radius: 6px;
  flex-wrap: wrap;
}
.field {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.label {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-muted);
}
.value {
  font-size: 13px;
  color: var(--text-primary);
}
.mono {
  font-family: ui-monospace, Consolas, monospace;
}
.spacer {
  flex: 1;
}
.rescan {
  padding: 8px 16px;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: var(--series-1);
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}
.rescan:disabled {
  opacity: 0.6;
  cursor: default;
}
</style>
