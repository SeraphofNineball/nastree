<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  rootPath: string
  currentPath: string
}>()

const emit = defineEmits<{ navigate: [path: string] }>()

const segments = computed(() => {
  const root = props.rootPath.replace(/[\\/]+$/, '')
  let rest = props.currentPath
  if (rest.startsWith(root)) {
    rest = rest.slice(root.length)
  }
  const parts = rest.split(/[\\/]+/).filter(Boolean)
  const out: { name: string; path: string }[] = [{ name: root || '/', path: root }]
  let acc = root
  for (const part of parts) {
    acc = acc + (acc.endsWith('/') || acc.endsWith('\\') ? '' : '/') + part
    out.push({ name: part, path: acc })
  }
  return out
})
</script>

<template>
  <div class="crumbs">
    <template v-for="(seg, i) in segments" :key="seg.path">
      <span v-if="i > 0" class="sep">/</span>
      <button class="crumb" :disabled="i === segments.length - 1" @click="emit('navigate', seg.path)">
        {{ seg.name }}
      </button>
    </template>
  </div>
</template>

<style scoped>
.crumbs {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  font-size: 13px;
  padding: 4px 2px;
}
.sep {
  color: var(--text-muted);
  margin: 0 4px;
}
.crumb {
  background: none;
  border: none;
  color: var(--series-1);
  cursor: pointer;
  padding: 2px 4px;
  font-size: 13px;
  border-radius: 4px;
}
.crumb:hover:not(:disabled) {
  background: var(--surface-1);
}
.crumb:disabled {
  color: var(--text-primary);
  font-weight: 600;
  cursor: default;
}
</style>
