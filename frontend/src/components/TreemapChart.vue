<script setup lang="ts">
import { computed } from 'vue'
import { use } from 'echarts/core'
import { TreemapChart as EchartsTreemap } from 'echarts/charts'
import { TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
import type { TreeNode, FileTypeStat } from '../types'
import {
  borderColor, buildExtColorMap, colorForExt, directoryColor, surfaceColor, textPrimaryColor,
  treemapGradient,
} from '../lib/colors'
import { formatBytes } from '../lib/format'
import { currentTheme } from '../lib/theme'

use([EchartsTreemap, TooltipComponent, CanvasRenderer])

const props = defineProps<{
  root: TreeNode | null
  fileTypes: FileTypeStat[]
}>()

const emit = defineEmits<{ enter: [path: string] }>()

interface EchartsTreemapDatum {
  name: string
  value: number // sqrt(size) - drives box area, so huge outliers don't crush everything else
  size: number // true byte size - for tooltip/label display
  path: string
  isDir: boolean
  files: number
  dirs: number
  ext?: string
  itemStyle: { color: ReturnType<typeof treemapGradient> }
  children?: EchartsTreemapDatum[]
}

function mapNode(n: TreeNode, extColors: Map<string, string>): EchartsTreemapDatum {
  const baseColor = n.isDir ? directoryColor() : colorForExt(n.ext, extColors)
  const datum: EchartsTreemapDatum = {
    name: n.name,
    value: Math.sqrt(n.size),
    size: n.size,
    path: n.path,
    isDir: n.isDir,
    files: n.files,
    dirs: n.dirs,
    ext: n.ext,
    itemStyle: { color: treemapGradient(baseColor) },
  }
  if (n.children?.length) {
    datum.children = n.children.map((c) => mapNode(c, extColors))
  }
  return datum
}

const option = computed(() => {
  void currentTheme.value
  const extColors = buildExtColorMap(props.fileTypes)
  const topChildren = props.root?.children ?? []
  const data = topChildren.map((n) => mapNode(n, extColors))
  // ECharts wraps top-level data in an implicit synthetic root for layout purposes;
  // that node has none of our custom fields, so its own upperLabel needs a fallback.
  const grandTotal = topChildren.reduce((sum, n) => sum + n.size, 0)
  const isLight = currentTheme.value === 'light'

  return {
    tooltip: {
      formatter: (info: any) => {
        const d = info.data as EchartsTreemapDatum
        const kind = d.isDir ? `${d.files} files, ${d.dirs} folders` : d.ext || 'file'
        return `<strong>${d.name}</strong><br/>${formatBytes(d.size)} &middot; ${kind}`
      },
      backgroundColor: surfaceColor(),
      borderColor: borderColor(),
      borderWidth: 1,
      textStyle: { color: textPrimaryColor() },
    },
    series: [
      {
        type: 'treemap',
        roam: false,
        nodeClick: false,
        breadcrumb: { show: false },
        // Files (leaf nodes) never get a label - just color, per WizTree's look.
        label: { show: false },
        // Folders get a reserved header strip instead, nested per level - this is
        // what actually produces the "ANIME\" / "DRAGON BALL Z\" stacked band look;
        // ECharts only renders it for nodes that have children, so files are unaffected.
        upperLabel: {
          show: true,
          height: 22,
          formatter: (p: any) => `${p.name}  (${formatBytes(p.data.size ?? grandTotal)})`,
          color: '#fff',
          backgroundColor: isLight ? 'rgba(128,128,128,0.9)' : 'rgba(20,20,20,0.85)',
          // a faint outline on each band so stacked levels (folder within folder)
          // read as distinct strips instead of merging into one solid block
          borderColor: 'rgba(255,255,255,0.18)',
          borderWidth: 1,
          fontSize: 12,
          padding: [4, 0, 0, 6],
        },
        // Once a folder's own box would render this small, stop subdividing it into
        // children - render it as one solid block instead of crushing its contents
        // into unreadable slivers.
        childrenVisibleMin: 6000,
        // Nodes smaller than this many px^2 aren't drawn as their own rect at all.
        visibleMin: 30,
        itemStyle: {
          borderColor: 'rgba(255,255,255,0.08)',
          borderWidth: 0.5,
          gapWidth: 0,
        },
        emphasis: {
          itemStyle: {
            borderColor: '#ffffff',
            borderWidth: 2,
          },
        },
        data,
      },
    ],
  }
})

function onClick(params: any) {
  if (params?.data?.isDir && params?.data?.path) {
    emit('enter', params.data.path)
  }
}
</script>

<template>
  <div class="treemap-wrap">
    <VChart v-if="root?.children?.length" class="chart" :option="option" autoresize @click="onClick" />
    <div v-else class="empty">This folder has no items.</div>
  </div>
</template>

<style scoped>
.treemap-wrap {
  height: 100%;
  min-height: 320px;
}
.chart {
  height: 100%;
  width: 100%;
}
.empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-muted);
  font-size: 14px;
}
</style>
