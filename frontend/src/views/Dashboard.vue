<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref } from 'vue'
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
import FileSearchView from '../components/FileSearchView.vue'

const status = ref<ScanStatus | null>(null)
const currentPath = ref('')
const children = ref<TreeNode[]>([])
const treeData = ref<TreeNode | null>(null)
const fileTypes = ref<FileTypeStat[]>([])
const loading = ref(false)
const error = ref('')
const vizMode = ref<'treemap' | 'pie' | 'files'>('treemap')
const focusMode = ref(false)

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
    const [kids, tree] = await Promise.all([api.children(path), api.tree(path)])
    children.value = kids
    treeData.value = tree
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

function onFileViewEnter(path: string) {
  // jump back to the treemap so the user sees where the file/folder they picked lives
  vizMode.value = 'treemap'
  navigate(path)
}

function goUp() {
  if (!status.value?.rootPath) return
  navigate(parentPath(currentPath.value, status.value.rootPath))
}

async function toggleFocusMode() {
  focusMode.value = !focusMode.value
  // the chart's container size changes synchronously with this toggle; nudge
  // ECharts to recompute immediately rather than waiting on its resize observer
  await nextTick()
  window.dispatchEvent(new Event('resize'))
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
    <SummaryBar
      :status="status"
      :scanning="status?.running ?? false"
      :focus-mode="focusMode"
      @rescan="rescan"
      @toggle-focus="toggleFocusMode"
    />

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
        <div class="spacer" />
        <div class="viz-toolbar">
          <button
            class="icon-btn"
            :class="{ active: vizMode === 'treemap' }"
            title="Treemap view"
            aria-label="Treemap view"
            @click="vizMode = 'treemap'"
          >
            <svg viewBox="0 0 20 20" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
              <rect x="1" y="1" width="10" height="10" />
              <rect x="12" y="1" width="7" height="6" />
              <rect x="12" y="8" width="7" height="3" />
              <rect x="1" y="12" width="6" height="7" />
              <rect x="8" y="12" width="11" height="7" />
            </svg>
          </button>
          <button
            class="icon-btn"
            :class="{ active: vizMode === 'pie' }"
            title="Pie chart view"
            aria-label="Pie chart view"
            @click="vizMode = 'pie'"
          >
            <svg viewBox="0 0 20 20" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
              <circle cx="10" cy="10" r="8" />
              <path d="M10 2 A8 8 0 0 1 18 10 L10 10 Z" fill="currentColor" stroke="none" />
            </svg>
          </button>
          <button
            class="icon-btn"
            :class="{ active: vizMode === 'files' }"
            title="File view (search & duplicates)"
            aria-label="File view"
            @click="vizMode = 'files'"
          >
            <svg viewBox="0 0 20 20" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
              <circle cx="8.5" cy="8.5" r="6" />
              <line x1="13" y1="13" x2="18" y2="18" stroke-linecap="round" />
            </svg>
          </button>
        </div>
      </div>

      <div v-if="!focusMode" class="panes">
        <DirectoryTree
          class="pane tree-pane"
          :root-path="status?.rootPath ?? ''"
          :current-path="currentPath"
          @navigate="navigate"
        />
        <FolderTable class="pane folder-table" :children="children" @enter="navigate" />
        <FileTypeTable class="pane filetype-table" :file-types="fileTypes" />
      </div>

      <div class="viz-pane">
        <TreemapChart v-if="vizMode === 'treemap'" :root="treeData" :file-types="fileTypes" @enter="navigate" />
        <PieChart v-else-if="vizMode === 'pie'" :children="children" :file-types="fileTypes" @enter="navigate" />
        <FileSearchView v-else :total-size="status?.totalSize ?? 0" @enter="onFileViewEnter" />
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
.spacer {
  flex: 1;
}
.panes {
  display: grid;
  grid-template-columns: 220px 3fr 2fr;
  gap: 12px;
  height: 28vh;
  min-height: 220px;
}
.pane {
  min-width: 0;
}
.viz-toolbar {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}
.icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  padding: 0;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: var(--surface-1);
  color: var(--text-secondary);
  cursor: pointer;
}
.icon-btn.active {
  background: var(--series-1);
  color: #fff;
  border-color: var(--series-1);
}
.viz-pane {
  flex: 1;
  min-height: 400px;
}
</style>
