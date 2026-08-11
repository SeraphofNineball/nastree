<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { api } from '../api'
import type { TreeNode } from '../types'
import { expandedPaths } from '../lib/treeState'

defineOptions({ name: 'TreeNodeItem' })

const props = defineProps<{
  node: TreeNode
  depth: number
  currentPath: string
}>()

const emit = defineEmits<{ navigate: [path: string] }>()

const hasChildren = computed(() => props.node.isDir && props.node.dirs > 0)
const expanded = computed(() => expandedPaths.has(props.node.path))
const isCurrent = computed(() => props.node.path === props.currentPath)

const childDirs = ref<TreeNode[] | null>(null)
const loading = ref(false)

async function loadChildren() {
  if (childDirs.value !== null || loading.value) return
  loading.value = true
  try {
    const kids = await api.children(props.node.path)
    childDirs.value = kids
      .filter((k) => k.isDir)
      .sort((a, b) => a.name.localeCompare(b.name, undefined, { sensitivity: 'base', numeric: true }))
  } finally {
    loading.value = false
  }
}

watch(
  expanded,
  (isExpanded) => {
    if (isExpanded) loadChildren()
  },
  { immediate: true },
)

function toggle() {
  if (!hasChildren.value) return
  if (expanded.value) {
    expandedPaths.delete(props.node.path)
  } else {
    expandedPaths.add(props.node.path)
  }
}

function select() {
  emit('navigate', props.node.path)
  // clicking a row both navigates and reveals its contents, mirroring the treemap's one-click drill-in
  if (hasChildren.value && !expanded.value) {
    expandedPaths.add(props.node.path)
  }
}

function onChildNavigate(path: string) {
  emit('navigate', path)
}
</script>

<template>
  <div class="node">
    <div class="row" :class="{ current: isCurrent }" :style="{ paddingLeft: `${depth * 16 + 4}px` }" @click="select">
      <span class="toggle" :class="{ visible: hasChildren }" @click.stop="toggle">
        {{ hasChildren ? (expanded ? '−' : '+') : '' }}
      </span>
      <span class="icon">📁</span>
      <span class="name">{{ node.name }}</span>
    </div>
    <div v-if="expanded" class="children">
      <div v-if="loading" class="loading" :style="{ paddingLeft: `${(depth + 1) * 16 + 4}px` }">Loading…</div>
      <TreeNodeItem
        v-for="child in childDirs"
        :key="child.path"
        :node="child"
        :depth="depth + 1"
        :current-path="currentPath"
        @navigate="onChildNavigate"
      />
    </div>
  </div>
</template>

<style scoped>
.row {
  display: flex;
  align-items: center;
  gap: 4px;
  padding-top: 3px;
  padding-bottom: 3px;
  padding-right: 8px;
  cursor: pointer;
  white-space: nowrap;
  font-size: 13px;
  color: var(--text-primary);
}
.row:hover {
  background: var(--surface-2);
}
.row.current {
  background: var(--series-1);
  color: #fff;
}
.toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  flex-shrink: 0;
  font-size: 11px;
  color: var(--text-muted);
  visibility: hidden;
}
.toggle.visible {
  visibility: visible;
  border: 1px solid var(--border);
  border-radius: 2px;
}
.row.current .toggle {
  color: #fff;
}
.icon {
  flex-shrink: 0;
}
.name {
  overflow: hidden;
  text-overflow: ellipsis;
}
.loading {
  font-size: 12px;
  color: var(--text-muted);
  padding-top: 2px;
  padding-bottom: 2px;
}
</style>
