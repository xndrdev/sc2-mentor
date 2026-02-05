<script setup lang="ts">
import { computed } from 'vue'
import type { DailyProgress } from '@/api/client'

const props = defineProps<{
  progressData: DailyProgress[]
  metric: 'apm' | 'sq' | 'supply_block' | 'win_rate'
  title: string
}>()

const chartOptions = computed(() => ({
  chart: {
    type: 'area',
    height: 200,
    toolbar: { show: false },
    background: 'transparent',
    animations: { enabled: true },
  },
  colors: [getChartColor()],
  fill: {
    type: 'gradient',
    gradient: {
      shadeIntensity: 1,
      opacityFrom: 0.4,
      opacityTo: 0.1,
      stops: [0, 100],
    },
  },
  stroke: {
    curve: 'smooth',
    width: 2,
  },
  dataLabels: { enabled: false },
  xaxis: {
    type: 'datetime',
    labels: {
      style: { colors: '#9CA3AF' },
      format: 'dd.MM',
    },
    axisBorder: { show: false },
    axisTicks: { show: false },
  },
  yaxis: {
    labels: {
      style: { colors: '#9CA3AF' },
      formatter: (val: number) => formatYAxisValue(val),
    },
  },
  grid: {
    borderColor: '#374151',
    strokeDashArray: 4,
  },
  tooltip: {
    theme: 'dark',
    x: { format: 'dd.MM.yyyy' },
    y: {
      formatter: (val: number) => formatTooltipValue(val),
    },
  },
}))

const series = computed(() => [{
  name: props.title,
  data: props.progressData.map(d => ({
    x: new Date(d.date).getTime(),
    y: getMetricValue(d),
  })),
}])

function getMetricValue(progress: DailyProgress): number {
  switch (props.metric) {
    case 'apm': return progress.avg_apm
    case 'sq': return progress.avg_spending_quotient
    case 'supply_block': return progress.avg_supply_block_pct
    case 'win_rate':
      const total = progress.wins + progress.losses
      return total > 0 ? (progress.wins / total) * 100 : 0
    default: return 0
  }
}

function getChartColor(): string {
  switch (props.metric) {
    case 'apm': return '#3B82F6'
    case 'sq': return '#10B981'
    case 'supply_block': return '#EF4444'
    case 'win_rate': return '#8B5CF6'
    default: return '#3B82F6'
  }
}

function formatYAxisValue(val: number): string {
  if (props.metric === 'supply_block' || props.metric === 'win_rate') {
    return `${val.toFixed(0)}%`
  }
  return val.toFixed(0)
}

function formatTooltipValue(val: number): string {
  if (props.metric === 'supply_block' || props.metric === 'win_rate') {
    return `${val.toFixed(1)}%`
  }
  return val.toFixed(0)
}
</script>

<template>
  <div class="bg-gray-800 rounded-lg p-4">
    <h4 class="text-white font-medium mb-4">{{ title }}</h4>
    <div v-if="progressData.length > 0">
      <apexchart
        type="area"
        height="200"
        :options="chartOptions"
        :series="series"
      />
    </div>
    <div v-else class="h-48 flex items-center justify-center text-gray-500">
      Keine Daten verfügbar
    </div>
  </div>
</template>
