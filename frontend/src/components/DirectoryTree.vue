<script setup lang="ts">
import { ref, watch } from 'vue'
import { api } from '../api'
import type { TreeNode } from '../types'
import { expandedPaths } from '../lib/treeState'
import TreeNodeItem from './TreeNodeItem.vue'

const props = defineProps<{
  rootPath: string
  currentPath: string
}>()

const emit = defineEmits<{ navigate: [path: string] }>()

const rootNode = ref<TreeNode | null>(null)

async function loadRoot() {
  if (!props.rootPath) return
  rootNode.value = await api.node('')
  expandedPaths.add(props.rootPath.replace(/[\\/]+$/, ''))
}

watch(() => props.rootPath, loadRoot, { immediate: true })
</script>

<template>
  <div class="tree-wrap">
    <TreeNodeItem v-if="rootNode" :node="rootNode" :depth="0" :current-path="currentPath" @navigate="emit('navigate', $event)" />
  </div>
</template>

<style scoped>
.tree-wrap {
  overflow: auto;
  height: 100%;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--surface-1);
  padding: 6px 0;
}
</style>
