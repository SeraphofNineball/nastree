<script setup lang="ts">
import { computed } from 'vue'
import { use } from 'echarts/core'
import { PieChart as EchartsPie } from 'echarts/charts'
import { TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
import type { TreeNode, FileTypeStat } from '../types'
import { buildExtColorMap, colorForExt, directoryColor, otherColor, surfaceColor, textSecondaryColor } from '../lib/colors'
import { formatBytes } from '../lib/format'
import { currentTheme } from '../lib/theme'

use([EchartsPie, TooltipComponent, LegendComponent, CanvasRenderer])

const props = defineProps<{
  children: TreeNode[]
  fileTypes: FileTypeStat[]
}>()

const emit = defineEmits<{ enter: [path: string] }>()

const MAX_SLICES = 12
const MIN_SLICE_FRACTION = 0.015 // slices under 1.5% of the total fold into "Other" too

const option = computed(() => {
  void currentTheme.value
  const extColors = buildExtColorMap(props.fileTypes)
  const sorted = [...props.children].sort((a, b) => b.size - a.size)
  const grandTotal = sorted.reduce((sum, n) => sum + n.size, 0)
  const minSize = grandTotal * MIN_SLICE_FRACTION

  const top: TreeNode[] = []
  const rest: TreeNode[] = []
  for (const n of sorted) {
    // always keep at least 3 slices, even if every item is below the threshold
    const keep = top.length < MAX_SLICES && (top.length < 3 || n.size >= minSize)
    if (keep) top.push(n)
    else rest.push(n)
  }
  const restTotal = rest.reduce((sum, n) => sum + n.size, 0)

  const textColor = textSecondaryColor()
  const data = top.map((n) => ({
    name: n.name,
    value: n.size,
    path: n.path,
    isDir: n.isDir,
    itemStyle: { color: n.isDir ? directoryColor() : colorForExt(n.ext, extColors) },
  }))
  if (restTotal > 0) {
    data.push({ name: `Other (${rest.length})`, value: restTotal, path: '', isDir: false, itemStyle: { color: otherColor() } })
  }

  return {
    tooltip: {
      formatter: (info: any) => `<strong>${info.name}</strong><br/>${formatBytes(info.value)} &middot; ${info.percent}%`,
    },
    legend: {
      type: 'scroll',
      orient: 'vertical',
      right: 8,
      top: 8,
      bottom: 8,
      textStyle: { color: textColor },
      pageIconColor: textColor,
      pageTextStyle: { color: textColor },
    },
    series: [
      {
        type: 'pie',
        radius: ['35%', '70%'],
        center: ['38%', '50%'],
        avoidLabelOverlap: true,
        itemStyle: {
          borderColor: surfaceColor(),
          borderWidth: 2,
        },
        label: {
          formatter: (p: any) => `${p.name}\n${formatBytes(p.value)}`,
          color: textColor,
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
  <div class="pie-wrap">
    <VChart v-if="children.length" class="chart" :option="option" autoresize @click="onClick" />
    <div v-else class="empty">This folder has no items.</div>
  </div>
</template>

<style scoped>
.pie-wrap {
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
