<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getTrends, listReplays } from '@/api/client'
import type { TrendData } from '@/api/client'

interface PlayerOption {
  id: number
  name: string
}

const players = ref<PlayerOption[]>([])
const selectedPlayerId = ref<number | null>(null)
const trends = ref<Record<string, TrendData>>({})
const loading = ref(false)
const error = ref<string | null>(null)

onMounted(async () => {
  // Lade Spieler aus den Replays
  try {
    const data = await listReplays(100, 0)
    const playerMap = new Map<number, string>()
    for (const replay of data.replays || []) {
      for (const player of replay.players || []) {
        if (player.is_human && !playerMap.has(player.player_id)) {
          playerMap.set(player.player_id, player.name)
        }
      }
    }
    players.value = Array.from(playerMap.entries()).map(([id, name]) => ({ id, name }))
    if (players.value.length > 0) {
      selectedPlayerId.value = players.value[0].id
      await loadTrends()
    }
  } catch (e) {
    error.value = 'Fehler beim Laden der Spieler'
  }
})

async function loadTrends() {
  if (!selectedPlayerId.value) return

  loading.value = true
  error.value = null
  try {
    const data = await getTrends(selectedPlayerId.value, 20)
    trends.value = data.trends || {}
  } catch (e) {
    error.value = 'Fehler beim Laden der Trends'
    trends.value = {}
  } finally {
    loading.value = false
  }
}

function getTrendIcon(trend: string): string {
  switch (trend) {
    case 'improving':
      return '📈'
    case 'declining':
      return '📉'
    default:
      return '➡️'
  }
}

function getTrendClass(trend: string): string {
  switch (trend) {
    case 'improving':
      return 'text-green-400'
    case 'declining':
      return 'text-red-400'
    default:
      return 'text-gray-400'
  }
}

function formatChange(change: number): string {
  const sign = change >= 0 ? '+' : ''
  return `${sign}${change.toFixed(1)}%`
}

const trendItems = computed(() => {
  const items = []
  if (trends.value.apm) {
    items.push({
      key: 'apm',
      label: 'APM',
      ...trends.value.apm,
    })
  }
  if (trends.value.spending_quotient) {
    items.push({
      key: 'sq',
      label: 'Spending Quotient',
      ...trends.value.spending_quotient,
    })
  }
  return items
})
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
    <h1 class="text-2xl font-bold text-white">Verbesserungstrends</h1>

    <!-- Player Selection -->
    <div class="bg-gray-800 rounded-lg p-6">
      <label class="block text-gray-400 text-sm mb-2">Spieler auswählen</label>
      <div class="flex items-center gap-4">
        <select
          v-model="selectedPlayerId"
          @change="loadTrends"
          class="bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500"
        >
          <option v-for="player in players" :key="player.id" :value="player.id">
            {{ player.name }}
          </option>
        </select>
        <button
          @click="loadTrends"
          :disabled="loading"
          class="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 rounded-lg text-white transition-colors"
        >
          Aktualisieren
        </button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="text-center py-8">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-400 mx-auto"></div>
      <p class="mt-2 text-gray-400">Lade Trends...</p>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="bg-red-900/50 border border-red-700 rounded-lg p-4">
      <p class="text-red-400">{{ error }}</p>
    </div>

    <!-- No Data -->
    <div v-else-if="!players.length" class="bg-gray-800 rounded-lg p-8 text-center">
      <p class="text-gray-400">Keine Spielerdaten vorhanden.</p>
      <p class="text-gray-500 text-sm mt-1">Lade erst einige Replays hoch.</p>
    </div>

    <!-- Trends -->
    <div v-else-if="trendItems.length" class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div
        v-for="item in trendItems"
        :key="item.key"
        class="bg-gray-800 rounded-lg p-6"
      >
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold text-white">{{ item.label }}</h3>
          <span class="text-2xl">{{ getTrendIcon(item.trend) }}</span>
        </div>

        <div class="flex items-end gap-4">
          <div>
            <p class="text-gray-400 text-sm">Trend</p>
            <p :class="getTrendClass(item.trend)" class="text-xl font-bold capitalize">
              {{ item.trend === 'improving' ? 'Verbesserung' : item.trend === 'declining' ? 'Rückgang' : 'Stabil' }}
            </p>
          </div>
          <div>
            <p class="text-gray-400 text-sm">Änderung</p>
            <p :class="getTrendClass(item.trend)" class="text-xl font-bold">
              {{ formatChange(item.change) }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="bg-gray-800 rounded-lg p-8 text-center">
      <p class="text-gray-400">Keine Trend-Daten verfügbar.</p>
      <p class="text-gray-500 text-sm mt-1">Du benötigst mindestens 2 Replays für Trends.</p>
    </div>
  </div>
</template>
