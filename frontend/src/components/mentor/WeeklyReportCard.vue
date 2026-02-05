<script setup lang="ts">
import type { WeeklyReport } from '@/api/client'

defineProps<{
  report: WeeklyReport
}>()

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('de-DE', {
    day: '2-digit',
    month: '2-digit',
  })
}

function formatPlayTime(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  if (hours > 0) {
    return `${hours}h ${minutes}m`
  }
  return `${minutes}m`
}

function parseJson<T>(data: T | string | undefined): T | null {
  if (!data) return null
  if (typeof data === 'string') {
    try {
      return JSON.parse(data)
    } catch {
      return null
    }
  }
  return data
}
</script>

<template>
  <div class="bg-gray-800 rounded-lg p-6">
    <div class="flex justify-between items-start mb-6">
      <div>
        <h3 class="text-lg font-semibold text-white">Wochenbericht</h3>
        <p class="text-sm text-gray-400">{{ formatDate(report.week_start) }} - {{ formatDate(report.week_end) }}</p>
      </div>
      <span v-if="report.main_race" class="px-3 py-1 bg-gray-700 rounded text-sm text-gray-300">
        {{ report.main_race }}
      </span>
    </div>

    <!-- Statistiken -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
      <div class="text-center">
        <div class="text-2xl font-bold text-white">{{ report.total_games }}</div>
        <div class="text-xs text-gray-400">Spiele</div>
      </div>
      <div class="text-center">
        <div class="text-2xl font-bold" :class="report.win_rate >= 50 ? 'text-green-400' : 'text-red-400'">
          {{ report.win_rate.toFixed(0) }}%
        </div>
        <div class="text-xs text-gray-400">Win Rate</div>
      </div>
      <div class="text-center">
        <div class="text-2xl font-bold text-white">{{ report.avg_apm.toFixed(0) }}</div>
        <div class="text-xs text-gray-400">APM</div>
      </div>
      <div class="text-center">
        <div class="text-2xl font-bold text-white">{{ formatPlayTime(report.total_play_time) }}</div>
        <div class="text-xs text-gray-400">Spielzeit</div>
      </div>
    </div>

    <!-- Verbesserungen & Verschlechterungen -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
      <div v-if="parseJson(report.improvements)" class="bg-green-500/10 border border-green-500/30 rounded-lg p-4">
        <h4 class="text-green-400 font-medium mb-2">Verbesserungen</h4>
        <ul class="space-y-1">
          <li v-for="(value, key) in parseJson(report.improvements)" :key="key" class="text-sm text-green-300">
            {{ key === 'apm' ? 'APM' : key === 'sq' ? 'SQ' : key === 'supply_block' ? 'Supply Block' : key === 'win_rate' ? 'Win Rate' : key }}: {{ value }}
          </li>
        </ul>
      </div>
      <div v-if="parseJson(report.regressions)" class="bg-red-500/10 border border-red-500/30 rounded-lg p-4">
        <h4 class="text-red-400 font-medium mb-2">Verschlechterungen</h4>
        <ul class="space-y-1">
          <li v-for="(value, key) in parseJson(report.regressions)" :key="key" class="text-sm text-red-300">
            {{ key === 'apm' ? 'APM' : key === 'sq' ? 'SQ' : key === 'supply_block' ? 'Supply Block' : key === 'win_rate' ? 'Win Rate' : key }}: {{ value }}
          </li>
        </ul>
      </div>
    </div>

    <!-- Stärken & Schwächen -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
      <div v-if="parseJson(report.strengths)?.length" class="bg-gray-700 rounded-lg p-4">
        <h4 class="text-white font-medium mb-2">Stärken</h4>
        <ul class="space-y-1">
          <li v-for="strength in parseJson(report.strengths)" :key="strength" class="text-sm text-gray-300 flex items-center gap-2">
            <svg class="w-4 h-4 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
            {{ strength }}
          </li>
        </ul>
      </div>
      <div v-if="parseJson(report.weaknesses)?.length" class="bg-gray-700 rounded-lg p-4">
        <h4 class="text-white font-medium mb-2">Verbesserungsbereiche</h4>
        <ul class="space-y-1">
          <li v-for="weakness in parseJson(report.weaknesses)" :key="weakness" class="text-sm text-gray-300 flex items-center gap-2">
            <svg class="w-4 h-4 text-yellow-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            {{ weakness }}
          </li>
        </ul>
      </div>
    </div>

    <!-- Fokus-Empfehlung -->
    <div v-if="report.focus_suggestion" class="bg-blue-500/10 border border-blue-500/30 rounded-lg p-4">
      <h4 class="text-blue-400 font-medium mb-2">Fokus für nächste Woche</h4>
      <p class="text-sm text-gray-300">{{ report.focus_suggestion }}</p>
    </div>
  </div>
</template>
