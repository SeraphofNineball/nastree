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

use([EchartsTreemap, TooltipComponent, CanvasRenderer])

const props = defineProps<{
  children: TreeNode[]
  fileTypes: FileTypeStat[]
}>()

const emit = defineEmits<{ enter: [path: string] }>()

const option = computed(() => {
  const extColors = buildExtColorMap(props.fileTypes)
  const data = props.children.map((n) => ({
    name: n.name,
    value: n.size,
    path: n.path,
    isDir: n.isDir,
    itemStyle: {
      color: n.isDir ? directoryColor() : colorForExt(n.ext, extColors),
    },
  }))

  return {
    tooltip: {
      formatter: (info: any) => {
        const n: TreeNode | undefined = props.children.find((c) => c.path === info.data.path)
        if (!n) return info.name
        const kind = n.isDir ? `${n.files} files, ${n.dirs} folders` : n.ext || 'file'
        return `<strong>${n.name}</strong><br/>${formatBytes(n.size)} &middot; ${kind}`
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
          formatter: (p: any) => `${p.name}\n${formatBytes(p.value)}`,
          color: '#fff',
          textShadowColor: 'rgba(0,0,0,0.6)',
          textShadowBlur: 2,
        },
        itemStyle: {
          borderColor: surfaceColor(),
          borderWidth: 2,
          gapWidth: 2,
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
    <VChart v-if="children.length" class="chart" :option="option" autoresize @click="onClick" />
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
