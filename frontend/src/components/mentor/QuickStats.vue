<script setup lang="ts">
import type { DailyProgress, WeekStats } from '@/api/client'

defineProps<{
  todayStats: DailyProgress | null
  weekStats: WeekStats | null
}>()

function formatPlayTime(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  if (hours > 0) {
    return `${hours}h ${minutes}m`
  }
  return `${minutes}m`
}

function formatChange(value: number): string {
  if (value > 0) return `+${value.toFixed(1)}%`
  if (value < 0) return `${value.toFixed(1)}%`
  return '0%'
}
</script>

<template>
  <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
    <!-- Heute -->
    <div class="bg-gray-800 rounded-lg p-6">
      <h3 class="text-lg font-semibold text-white mb-4">Heute</h3>
      <div v-if="todayStats" class="space-y-4">
        <div class="flex justify-between items-center">
          <span class="text-gray-400">Spiele</span>
          <span class="text-2xl font-bold text-white">{{ todayStats.games_played }}</span>
        </div>
        <div class="flex justify-between items-center">
          <span class="text-gray-400">Wins / Losses</span>
          <span class="text-lg">
            <span class="text-green-400">{{ todayStats.wins }}</span>
            <span class="text-gray-500"> / </span>
            <span class="text-red-400">{{ todayStats.losses }}</span>
          </span>
        </div>
        <div class="flex justify-between items-center">
          <span class="text-gray-400">APM</span>
          <span class="text-lg text-white">{{ todayStats.avg_apm.toFixed(0) }}</span>
        </div>
        <div class="flex justify-between items-center">
          <span class="text-gray-400">Supply Block</span>
          <span class="text-lg" :class="todayStats.avg_supply_block_pct < 10 ? 'text-green-400' : todayStats.avg_supply_block_pct < 20 ? 'text-yellow-400' : 'text-red-400'">
            {{ todayStats.avg_supply_block_pct.toFixed(1) }}%
          </span>
        </div>
        <div class="flex justify-between items-center">
          <span class="text-gray-400">Spielzeit</span>
          <span class="text-lg text-white">{{ formatPlayTime(todayStats.total_play_time) }}</span>
        </div>
      </div>
      <div v-else class="text-gray-500 text-center py-8">
        Noch keine Spiele heute
      </div>
    </div>

    <!-- Diese Woche -->
    <div class="bg-gray-800 rounded-lg p-6">
      <h3 class="text-lg font-semibold text-white mb-4">Diese Woche</h3>
      <div v-if="weekStats && weekStats.games_played > 0" class="space-y-4">
        <div class="flex justify-between items-center">
          <span class="text-gray-400">Spiele</span>
          <span class="text-2xl font-bold text-white">{{ weekStats.games_played }}</span>
        </div>
        <div class="flex justify-between items-center">
          <span class="text-gray-400">Win Rate</span>
          <div class="text-right">
            <span class="text-lg" :class="weekStats.win_rate >= 50 ? 'text-green-400' : 'text-red-400'">
              {{ weekStats.win_rate.toFixed(1) }}%
            </span>
            <span v-if="weekStats.win_rate_change !== 0" class="text-sm ml-2" :class="weekStats.win_rate_change > 0 ? 'text-green-400' : 'text-red-400'">
              {{ formatChange(weekStats.win_rate_change) }}
            </span>
          </div>
        </div>
        <div class="flex justify-between items-center">
          <span class="text-gray-400">APM</span>
          <div class="text-right">
            <span class="text-lg text-white">{{ weekStats.avg_apm.toFixed(0) }}</span>
            <span v-if="weekStats.apm_change !== 0" class="text-sm ml-2" :class="weekStats.apm_change > 0 ? 'text-green-400' : 'text-red-400'">
              {{ formatChange(weekStats.apm_change) }}
            </span>
          </div>
        </div>
        <div class="flex justify-between items-center">
          <span class="text-gray-400">SQ</span>
          <div class="text-right">
            <span class="text-lg text-white">{{ weekStats.avg_sq.toFixed(0) }}</span>
            <span v-if="weekStats.sq_change !== 0" class="text-sm ml-2" :class="weekStats.sq_change > 0 ? 'text-green-400' : 'text-red-400'">
              {{ formatChange(weekStats.sq_change) }}
            </span>
          </div>
        </div>
        <div class="flex justify-between items-center">
          <span class="text-gray-400">Spielzeit</span>
          <span class="text-lg text-white">{{ formatPlayTime(weekStats.total_play_time) }}</span>
        </div>
      </div>
      <div v-else class="text-gray-500 text-center py-8">
        Noch keine Spiele diese Woche
      </div>
    </div>
  </div>
</template>
