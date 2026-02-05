<script setup lang="ts">
import { computed } from 'vue'
import type { APMAnalysis } from '@/api/client'

const props = defineProps<{
  data: APMAnalysis
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
  },
  colors: ['#f59e0b'],
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
    labels: {
      formatter: (val: number) => Math.round(val).toString(),
      style: { colors: '#9ca3af' },
    },
    min: 0,
  },
  grid: { borderColor: '#374151' },
  tooltip: {
    theme: 'dark',
    x: {
      formatter: (val: number) => formatTime(val),
    },
  },
  annotations: {
    yaxis: [
      {
        y: props.data.average_apm,
        borderColor: '#f59e0b',
        borderWidth: 2,
        strokeDashArray: 5,
        label: {
          text: `Ø ${Math.round(props.data.average_apm)} APM`,
          style: {
            color: '#fff',
            background: '#f59e0b',
          },
        },
      },
    ],
  },
}))

const series = computed(() => [
  {
    name: 'APM',
    data: props.data.apm_timeline.map(p => ({ x: p.time, y: Math.round(p.apm) })),
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
    <div class="mt-4 grid grid-cols-3 gap-4 text-sm text-center">
      <div>
        <p class="text-gray-400">Durchschnitt</p>
        <p class="text-xl font-bold text-white">{{ Math.round(data.average_apm) }}</p>
      </div>
      <div>
        <p class="text-gray-400">Peak</p>
        <p class="text-xl font-bold text-white">{{ Math.round(data.peak_apm) }}</p>
      </div>
      <div>
        <p class="text-gray-400">EAPM</p>
        <p class="text-xl font-bold text-white">{{ Math.round(data.eapm) }}</p>
      </div>
    </div>
  </div>
</template>
