<script setup lang="ts">
import { computed } from 'vue'
import type { SupplyAnalysis } from '@/api/client'

const props = defineProps<{
  data: SupplyAnalysis
}>()

function formatTime(seconds: number): string {
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

const chartOptions = computed(() => ({
  chart: {
    type: 'area',
    height: 300,
    background: 'transparent',
    toolbar: { show: false },
    animations: { enabled: true },
  },
  colors: ['#3b82f6', '#10b981'],
  stroke: { curve: 'smooth', width: 2 },
  fill: {
    type: 'gradient',
    gradient: {
      shadeIntensity: 0.8,
      opacityFrom: 0.4,
      opacityTo: 0.1,
    },
  },
  xaxis: {
    type: 'numeric',
    labels: {
      formatter: (val: number) => formatTime(val),
      style: { colors: '#9ca3af' },
    },
    axisBorder: { color: '#374151' },
    axisTicks: { color: '#374151' },
  },
  yaxis: {
    labels: { style: { colors: '#9ca3af' } },
  },
  grid: { borderColor: '#374151' },
  legend: {
    labels: { colors: '#9ca3af' },
    position: 'top',
  },
  tooltip: {
    theme: 'dark',
    x: {
      formatter: (val: number) => formatTime(val),
    },
  },
  annotations: {
    xaxis: props.data.blocks
      .filter(b => b.severity === 'high')
      .map(block => ({
        x: block.start_time,
        x2: block.end_time,
        fillColor: '#ef4444',
        opacity: 0.2,
        label: {
          text: 'Block',
          style: { color: '#fff', background: '#ef4444' },
        },
      })),
  },
}))

const series = computed(() => [
  {
    name: 'Supply Used',
    data: props.data.supply_timeline.map(p => ({ x: p.time, y: p.supply_used })),
  },
  {
    name: 'Supply Max',
    data: props.data.supply_timeline.map(p => ({ x: p.time, y: p.supply_max })),
  },
])
</script>

<template>
  <div>
    <apexchart
      type="area"
      height="300"
      :options="chartOptions"
      :series="series"
    />
    <div class="mt-4 flex items-center justify-between text-sm">
      <div class="text-gray-400">
        Blockzeit: <span class="text-white font-medium">{{ data.total_block_time.toFixed(1) }}s</span>
      </div>
      <div class="text-gray-400">
        Blocks: <span class="text-white font-medium">{{ data.blocks.length }}</span>
      </div>
    </div>
  </div>
</template>
