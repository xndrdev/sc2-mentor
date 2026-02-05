<script setup lang="ts">
import { computed } from 'vue'
import type { SpendingAnalysis } from '@/api/client'

const props = defineProps<{
  data: SpendingAnalysis
}>()

function formatTime(seconds: number): string {
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

const chartOptions = computed(() => ({
  chart: {
    type: 'line',
    height: 300,
    background: 'transparent',
    toolbar: { show: false },
  },
  colors: ['#3b82f6', '#22c55e'],
  stroke: { curve: 'smooth', width: 2 },
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
}))

const series = computed(() => [
  {
    name: 'Mineralien',
    data: props.data.resource_timeline.map(p => ({ x: p.time, y: p.minerals })),
  },
  {
    name: 'Gas',
    data: props.data.resource_timeline.map(p => ({ x: p.time, y: p.gas })),
  },
])

function getRatingColor(rating: string): string {
  switch (rating) {
    case 'excellent': return 'text-green-400'
    case 'good': return 'text-blue-400'
    case 'average': return 'text-yellow-400'
    case 'below_average': return 'text-orange-400'
    default: return 'text-red-400'
  }
}

function getRatingText(rating: string): string {
  switch (rating) {
    case 'excellent': return 'Exzellent'
    case 'good': return 'Gut'
    case 'average': return 'Durchschnitt'
    case 'below_average': return 'Unterdurchschnittlich'
    default: return 'Verbesserungswürdig'
  }
}
</script>

<template>
  <div>
    <apexchart
      type="line"
      height="300"
      :options="chartOptions"
      :series="series"
    />
    <div class="mt-4 grid grid-cols-2 gap-4 text-sm">
      <div>
        <p class="text-gray-400">Spending Quotient</p>
        <p class="text-xl font-bold text-white">{{ Math.round(data.spending_quotient) }}</p>
        <p :class="getRatingColor(data.rating)" class="text-sm">
          {{ getRatingText(data.rating) }}
        </p>
      </div>
      <div>
        <p class="text-gray-400">Ø Ungenutzt</p>
        <p class="text-white">
          <span class="text-blue-400">{{ Math.round(data.average_unspent.minerals) }}</span> /
          <span class="text-green-400">{{ Math.round(data.average_unspent.gas) }}</span>
        </p>
      </div>
    </div>
  </div>
</template>
