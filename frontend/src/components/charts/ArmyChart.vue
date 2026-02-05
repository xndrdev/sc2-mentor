<script setup lang="ts">
import { computed } from 'vue'
import type { ArmyAnalysis } from '@/api/client'

const props = defineProps<{
  data: ArmyAnalysis
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
  colors: ['#ef4444'],
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
}))

const series = computed(() => [
  {
    name: 'Armeewert',
    data: props.data.army_timeline.map(p => ({ x: p.time, y: p.value })),
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
    <div class="mt-4">
      <div class="flex items-center justify-between text-sm">
        <span class="text-gray-400">Peak Armeewert</span>
        <span class="text-white font-bold">{{ data.peak_army_value.toLocaleString() }}</span>
      </div>

      <!-- Unit Composition -->
      <div v-if="data.unit_composition?.length" class="mt-4">
        <p class="text-gray-400 text-sm mb-2">Einheiten-Komposition (Ende)</p>
        <div class="flex flex-wrap gap-2">
          <span
            v-for="unit in data.unit_composition.slice(0, 8)"
            :key="unit.unit_type"
            class="px-2 py-1 bg-gray-700 rounded text-xs text-gray-300"
          >
            {{ unit.unit_type }}: {{ unit.count }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>
