<script setup lang="ts">
import type { Goal } from '@/api/client'

const props = defineProps<{
  goal: Goal
}>()

const emit = defineEmits<{
  delete: [goalId: number]
}>()

function getProgress(): number {
  if (props.goal.target_value === 0) return 0

  // Für "kleiner als" Ziele
  if (props.goal.comparison === '<=' || props.goal.comparison === '<') {
    if (props.goal.current_value <= props.goal.target_value) return 100
    return (props.goal.target_value / props.goal.current_value) * 100
  }

  return Math.min((props.goal.current_value / props.goal.target_value) * 100, 100)
}

function isAchieved(): boolean {
  switch (props.goal.comparison) {
    case '>=': return props.goal.current_value >= props.goal.target_value
    case '<=': return props.goal.current_value <= props.goal.target_value
    case '>': return props.goal.current_value > props.goal.target_value
    case '<': return props.goal.current_value < props.goal.target_value
    case '=': return props.goal.current_value === props.goal.target_value
    default: return props.goal.current_value >= props.goal.target_value
  }
}

function getMetricLabel(metricName: string): string {
  const labels: Record<string, string> = {
    games_played: 'Spiele',
    apm: 'APM',
    supply_block: 'Supply Block',
    win_rate: 'Win Rate',
    sq: 'Spending Quotient',
  }
  return labels[metricName] || metricName
}

function formatValue(value: number, metricName: string): string {
  if (metricName === 'win_rate' || metricName === 'supply_block') {
    return `${value.toFixed(1)}%`
  }
  if (metricName === 'games_played') {
    return value.toFixed(0)
  }
  return value.toFixed(0)
}

function formatDeadline(deadline: string): string {
  const date = new Date(deadline)
  const now = new Date()
  const diff = date.getTime() - now.getTime()
  const hours = Math.floor(diff / (1000 * 60 * 60))

  if (hours < 0) return 'Abgelaufen'
  if (hours < 24) return `${hours}h übrig`
  const days = Math.floor(hours / 24)
  return `${days}d übrig`
}

const progress = getProgress()
const achieved = isAchieved()
</script>

<template>
  <div class="bg-gray-700 rounded-lg p-4">
    <div class="flex justify-between items-start mb-2">
      <div>
        <span class="text-xs font-medium px-2 py-1 rounded" :class="goal.goal_type === 'daily' ? 'bg-blue-500/20 text-blue-400' : 'bg-purple-500/20 text-purple-400'">
          {{ goal.goal_type === 'daily' ? 'Täglich' : 'Wöchentlich' }}
        </span>
        <h4 class="text-white font-medium mt-2">{{ getMetricLabel(goal.metric_name) }}</h4>
      </div>
      <button @click="emit('delete', goal.id)" class="text-gray-400 hover:text-red-400 transition-colors">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <div class="flex items-baseline gap-2 mb-3">
      <span class="text-2xl font-bold" :class="achieved ? 'text-green-400' : 'text-white'">
        {{ formatValue(goal.current_value, goal.metric_name) }}
      </span>
      <span class="text-gray-400">
        {{ goal.comparison }} {{ formatValue(goal.target_value, goal.metric_name) }}
      </span>
    </div>

    <!-- Progress Bar -->
    <div class="relative h-2 bg-gray-600 rounded-full overflow-hidden mb-2">
      <div
        class="absolute h-full rounded-full transition-all duration-300"
        :class="achieved ? 'bg-green-500' : 'bg-blue-500'"
        :style="{ width: `${progress}%` }"
      />
    </div>

    <div class="flex justify-between text-xs text-gray-400">
      <span>{{ progress.toFixed(0) }}%</span>
      <span>{{ formatDeadline(goal.deadline) }}</span>
    </div>
  </div>
</template>
