<script setup lang="ts">
import { THEMES, currentTheme, setTheme } from '../lib/theme'
</script>

<template>
  <select class="theme-select" :value="currentTheme" @change="setTheme(($event.target as HTMLSelectElement).value)">
    <option v-for="t in THEMES" :key="t.value" :value="t.value">{{ t.label }}</option>
  </select>
</template>

<style scoped>
.theme-select {
  padding: 8px 10px;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: var(--surface-2);
  color: var(--text-primary);
  font-size: 13px;
  cursor: pointer;
}
</style>
