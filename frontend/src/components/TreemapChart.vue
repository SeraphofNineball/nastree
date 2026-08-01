<script setup lang="ts">
import { computed } from 'vue'
import { use } from 'echarts/core'
import { TreemapChart as EchartsTreemap } from 'echarts/charts'
import { TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
import type { TreeNode, FileTypeStat } from '../types'
import { buildExtColorMap, colorForExt, directoryColor, surfaceColor } from '../lib/colors'
import { formatBytes } from '../lib/format'
import { currentTheme } from '../lib/theme'

use([EchartsTreemap, TooltipComponent, CanvasRenderer])

const props = defineProps<{
  root: TreeNode | null
  fileTypes: FileTypeStat[]
}>()

const emit = defineEmits<{ enter: [path: string] }>()

// Below this fraction of the root's total size, a region's name isn't worth labeling -
// keeps small nested boxes as pure color, like WizTree's mosaic.
const LABEL_MIN_FRACTION = 0.008

interface EchartsTreemapDatum {
  name: string
  value: number
  path: string
  isDir: boolean
  files: number
  dirs: number
  ext?: string
  itemStyle: { color: string }
  label?: { show: boolean }
  children?: EchartsTreemapDatum[]
}

function mapNode(n: TreeNode, extColors: Map<string, string>, labelMinSize: number): EchartsTreemapDatum {
  const datum: EchartsTreemapDatum = {
    name: n.name,
    value: n.size,
    path: n.path,
    isDir: n.isDir,
    files: n.files,
    dirs: n.dirs,
    ext: n.ext,
    itemStyle: { color: n.isDir ? directoryColor() : colorForExt(n.ext, extColors) },
  }
  if (!n.isDir || n.size < labelMinSize) {
    datum.label = { show: false }
  }
  if (n.children?.length) {
    datum.children = n.children.map((c) => mapNode(c, extColors, labelMinSize))
  }
  return datum
}

const option = computed(() => {
  void currentTheme.value
  const extColors = buildExtColorMap(props.fileTypes)
  const topChildren = props.root?.children ?? []
  const total = topChildren.reduce((sum, n) => sum + n.size, 0)
  const labelMinSize = total * LABEL_MIN_FRACTION
  const data = topChildren.map((n) => mapNode(n, extColors, labelMinSize))

  return {
    tooltip: {
      formatter: (info: any) => {
        const d = info.data as EchartsTreemapDatum
        const kind = d.isDir ? `${d.files} files, ${d.dirs} folders` : d.ext || 'file'
        return `<strong>${d.name}</strong><br/>${formatBytes(d.value)} &middot; ${kind}`
      },
    },
    series: [
      {
        type: 'treemap',
        roam: false,
        nodeClick: false,
        breadcrumb: { show: false },
        upperLabel: { show: false },
        label: {
          show: true,
          position: 'insideTopLeft',
          formatter: (p: any) => p.name,
          fontSize: 11,
          color: '#fff',
          backgroundColor: 'rgba(0,0,0,0.55)',
          padding: [2, 4],
          overflow: 'truncate',
        },
        itemStyle: {
          borderColor: surfaceColor(),
          borderWidth: 1,
          gapWidth: 1,
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
