<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { api } from '../api'
import type { TreeNode, FileTypeStat, ScanStatus } from '../types'
import { parentPath } from '../lib/paths'
import { expandAncestors } from '../lib/treeState'
import SummaryBar from '../components/SummaryBar.vue'
import Breadcrumb from '../components/Breadcrumb.vue'
import DirectoryTree from '../components/DirectoryTree.vue'
import FolderTable from '../components/FolderTable.vue'
import FileTypeTable from '../components/FileTypeTable.vue'
import TreemapChart from '../components/TreemapChart.vue'
import PieChart from '../components/PieChart.vue'

const status = ref<ScanStatus | null>(null)
const currentPath = ref('')
const children = ref<TreeNode[]>([])
const fileTypes = ref<FileTypeStat[]>([])
const loading = ref(false)
const error = ref('')
const vizMode = ref<'treemap' | 'pie'>('treemap')

let pollHandle: ReturnType<typeof setInterval> | undefined

async function loadStatus() {
  const s = await api.status()
  if ('rootPath' in s) {
    const wasRunning = status.value?.running
    status.value = s
    if (wasRunning && !s.running) {
      // a scan just finished - refresh whatever the user is looking at
      await loadPath(currentPath.value)
      await loadFileTypes()
    }
  } else {
    status.value = null
  }
}

async function loadPath(path: string) {
  loading.value = true
  error.value = ''
  try {
    const kids = await api.children(path)
    children.value = kids
    currentPath.value = path || status.value?.rootPath || ''
    if (status.value?.rootPath) {
      expandAncestors(currentPath.value, status.value.rootPath)
    }
  } catch (e: any) {
    error.value = e.message ?? String(e)
  } finally {
    loading.value = false
  }
}

async function loadFileTypes() {
  try {
    fileTypes.value = await api.fileTypes()
  } catch {
    fileTypes.value = []
  }
}

async function rescan() {
  await api.triggerScan()
  await loadStatus()
}

function navigate(path: string) {
  loadPath(path)
}

function goUp() {
  if (!status.value?.rootPath) return
  navigate(parentPath(currentPath.value, status.value.rootPath))
}

onMounted(async () => {
  await loadStatus()
  if (status.value?.rootPath) {
    await loadPath(status.value.rootPath)
    await loadFileTypes()
  }
  pollHandle = setInterval(loadStatus, 5000)
})

onUnmounted(() => {
  if (pollHandle) clearInterval(pollHandle)
})
</script>

<template>
  <div class="dashboard">
    <SummaryBar :status="status" :scanning="status?.running ?? false" @rescan="rescan" />

    <div v-if="error" class="error">{{ error }}</div>

    <div v-if="!status && !loading" class="waiting">
      No scan has completed yet. The first background scan is likely still running —
      it can take a while on a large NAS.
    </div>

    <template v-else>
      <div class="nav-row">
        <button
          class="up-btn"
          :disabled="currentPath === status?.rootPath"
          title="Go to parent folder"
          @click="goUp"
        >
          ▲ Up
        </button>
        <Breadcrumb :root-path="status?.rootPath ?? ''" :current-path="currentPath" @navigate="navigate" />
      </div>

      <div class="panes">
        <DirectoryTree
          class="pane tree-pane"
          :root-path="status?.rootPath ?? ''"
          :current-path="currentPath"
          @navigate="navigate"
        />
        <FolderTable class="pane folder-table" :children="children" @enter="navigate" />
        <FileTypeTable class="pane filetype-table" :file-types="fileTypes" />
      </div>

      <div class="viz-toolbar">
        <button :class="{ active: vizMode === 'treemap' }" @click="vizMode = 'treemap'">Treemap</button>
        <button :class="{ active: vizMode === 'pie' }" @click="vizMode = 'pie'">Pie</button>
      </div>

      <div class="treemap-pane">
        <TreemapChart v-if="vizMode === 'treemap'" :children="children" :file-types="fileTypes" @enter="navigate" />
        <PieChart v-else :children="children" :file-types="fileTypes" />
      </div>
    </template>
  </div>
</template>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 12px;
  height: 100%;
  padding: 16px;
}
.error {
  color: var(--series-8);
  font-size: 13px;
}
.waiting {
  color: var(--text-secondary);
  font-size: 14px;
  padding: 32px;
  text-align: center;
}
.nav-row {
  display: flex;
  align-items: center;
  gap: 10px;
}
.up-btn {
  padding: 4px 10px;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: var(--surface-1);
  color: var(--text-primary);
  font-size: 12px;
  cursor: pointer;
  flex-shrink: 0;
}
.up-btn:hover:not(:disabled) {
  background: var(--surface-2);
}
.up-btn:disabled {
  opacity: 0.4;
  cursor: default;
}
.panes {
  display: grid;
  grid-template-columns: 220px 3fr 2fr;
  gap: 12px;
  height: 38vh;
  min-height: 260px;
}
.pane {
  min-width: 0;
}
.viz-toolbar {
  display: flex;
  gap: 6px;
}
.viz-toolbar button {
  padding: 6px 14px;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: var(--surface-1);
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
}
.viz-toolbar button.active {
  background: var(--series-1);
  color: #fff;
  border-color: var(--series-1);
}
.treemap-pane {
  flex: 1;
  min-height: 320px;
}
</style>
