<script setup lang="ts">
import { onMounted, computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useReplayStore } from '@/stores/replays'
import SupplyChart from '@/components/charts/SupplyChart.vue'
import ResourceChart from '@/components/charts/ResourceChart.vue'
import APMChart from '@/components/charts/APMChart.vue'
import ArmyChart from '@/components/charts/ArmyChart.vue'
import BuildOrder from '@/components/BuildOrder.vue'
import Suggestions from '@/components/Suggestions.vue'
import StrategicAnalysis from '@/components/StrategicAnalysis.vue'

const route = useRoute()
const router = useRouter()
const store = useReplayStore()
const selectedPlayerId = ref<number | null>(null)
const activeTab = ref<'metrics' | 'strategic'>('metrics')

const replayId = computed(() => Number(route.params.id))

onMounted(async () => {
  await store.fetchAnalysis(replayId.value)
  // Wähle automatisch den ersten menschlichen Spieler
  if (store.currentAnalysis?.replay.players) {
    const humanPlayer = store.currentAnalysis.replay.players.find(p => p.is_human)
    if (humanPlayer) {
      selectedPlayerId.value = humanPlayer.player_id
    }
  }
  // Lade auch die strategische Analyse
  store.fetchStrategicAnalysis(replayId.value)
})

const replay = computed(() => store.currentAnalysis?.replay)
const analyses = computed(() => store.currentAnalysis?.analyses || {})

const selectedAnalysis = computed(() => {
  if (!selectedPlayerId.value) return null
  return analyses.value[selectedPlayerId.value]
})

const selectedPlayer = computed(() => {
  if (!selectedPlayerId.value || !replay.value) return null
  return replay.value.players.find(p => p.player_id === selectedPlayerId.value)
})

function formatDuration(seconds: number): string {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

function getRaceClass(race: string): string {
  switch (race.toLowerCase()) {
    case 'terran': return 'text-blue-400'
    case 'zerg': return 'text-purple-400'
    case 'protoss': return 'text-yellow-400'
    default: return 'text-gray-400'
  }
}
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <!-- Back Button -->
    <button
      @click="router.push('/')"
      class="mb-6 flex items-center text-gray-400 hover:text-white transition-colors"
    >
      <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
      </svg>
      Zurück zur Übersicht
    </button>

    <!-- Loading State -->
    <div v-if="store.loading" class="text-center py-16">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-400 mx-auto"></div>
      <p class="mt-4 text-gray-400">Lade Analyse...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="store.error" class="bg-red-900/50 border border-red-700 rounded-lg p-6">
      <p class="text-red-400 text-center">{{ store.error }}</p>
    </div>

    <!-- Content -->
    <div v-else-if="replay" class="space-y-6">
      <!-- Header -->
      <div class="bg-gray-800 rounded-lg p-6">
        <div class="flex items-start justify-between">
          <div>
            <h1 class="text-2xl font-bold text-white">{{ replay.map }}</h1>
            <p class="text-gray-400 mt-1">
              {{ formatDuration(replay.duration) }} · Version {{ replay.game_version }}
            </p>
          </div>
          <div class="text-right text-gray-500 text-sm">
            {{ new Date(replay.played_at).toLocaleDateString('de-DE') }}
          </div>
        </div>

        <!-- Players -->
        <div class="mt-6 flex flex-wrap gap-4">
          <button
            v-for="player in replay.players"
            :key="player.player_id"
            @click="selectedPlayerId = player.player_id"
            class="flex items-center gap-3 px-4 py-2 rounded-lg border transition-colors"
            :class="{
              'border-blue-500 bg-blue-500/20': selectedPlayerId === player.player_id,
              'border-gray-600 hover:border-gray-500': selectedPlayerId !== player.player_id,
            }"
          >
            <span :class="getRaceClass(player.race)" class="font-medium">
              {{ player.name }}
            </span>
            <span class="text-gray-500 text-sm">({{ player.race }})</span>
            <span
              v-if="player.result === 'Win'"
              class="text-xs px-2 py-0.5 rounded bg-green-500/20 text-green-400"
            >
              Win
            </span>
            <span
              v-else-if="player.result === 'Loss'"
              class="text-xs px-2 py-0.5 rounded bg-red-500/20 text-red-400"
            >
              Loss
            </span>
          </button>
        </div>

        <!-- Quick Stats -->
        <div v-if="selectedPlayer" class="mt-6 grid grid-cols-2 sm:grid-cols-4 gap-4">
          <div class="bg-gray-700/50 rounded-lg p-3 text-center">
            <p class="text-gray-400 text-sm">APM</p>
            <p class="text-xl font-bold text-white">{{ Math.round(selectedPlayer.apm) }}</p>
          </div>
          <div class="bg-gray-700/50 rounded-lg p-3 text-center">
            <p class="text-gray-400 text-sm">Spending Quotient</p>
            <p class="text-xl font-bold text-white">{{ Math.round(selectedPlayer.spending_quotient) }}</p>
          </div>
          <div v-if="selectedAnalysis?.supply_analysis" class="bg-gray-700/50 rounded-lg p-3 text-center">
            <p class="text-gray-400 text-sm">Supply Block</p>
            <p class="text-xl font-bold text-white">
              {{ selectedAnalysis.supply_analysis.block_percentage.toFixed(1) }}%
            </p>
          </div>
          <div v-if="selectedAnalysis?.inject_analysis" class="bg-gray-700/50 rounded-lg p-3 text-center">
            <p class="text-gray-400 text-sm">Inject Effizienz</p>
            <p class="text-xl font-bold text-white">
              {{ selectedAnalysis.inject_analysis.efficiency.toFixed(0) }}%
            </p>
          </div>
        </div>

        <!-- Tab Buttons -->
        <div class="mt-6 flex gap-2 border-b border-gray-700 pb-2">
          <button
            @click="activeTab = 'metrics'"
            class="px-4 py-2 rounded-t-lg font-medium transition-colors"
            :class="{
              'bg-blue-600 text-white': activeTab === 'metrics',
              'text-gray-400 hover:text-white': activeTab !== 'metrics'
            }"
          >
            Metriken & Charts
          </button>
          <button
            @click="activeTab = 'strategic'"
            class="px-4 py-2 rounded-t-lg font-medium transition-colors flex items-center gap-2"
            :class="{
              'bg-purple-600 text-white': activeTab === 'strategic',
              'text-gray-400 hover:text-white': activeTab !== 'strategic'
            }"
          >
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
              <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"/>
            </svg>
            Strategische Analyse
          </button>
        </div>
      </div>

      <!-- Tab: Metrics & Charts -->
      <div v-if="activeTab === 'metrics' && selectedAnalysis" class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Supply Chart -->
        <div v-if="selectedAnalysis.supply_analysis" class="bg-gray-800 rounded-lg p-6">
          <h2 class="text-lg font-semibold text-white mb-4">Supply Timeline</h2>
          <SupplyChart :data="selectedAnalysis.supply_analysis" />
        </div>

        <!-- Resource Chart -->
        <div v-if="selectedAnalysis.spending_analysis" class="bg-gray-800 rounded-lg p-6">
          <h2 class="text-lg font-semibold text-white mb-4">Ressourcen</h2>
          <ResourceChart :data="selectedAnalysis.spending_analysis" />
        </div>

        <!-- APM Chart -->
        <div v-if="selectedAnalysis.apm_analysis" class="bg-gray-800 rounded-lg p-6">
          <h2 class="text-lg font-semibold text-white mb-4">APM Verlauf</h2>
          <APMChart :data="selectedAnalysis.apm_analysis" />
        </div>

        <!-- Army Chart -->
        <div v-if="selectedAnalysis.army_analysis" class="bg-gray-800 rounded-lg p-6">
          <h2 class="text-lg font-semibold text-white mb-4">Armeewert</h2>
          <ArmyChart :data="selectedAnalysis.army_analysis" />
        </div>

        <!-- Build Order -->
        <div v-if="selectedAnalysis.build_order?.length" class="bg-gray-800 rounded-lg p-6">
          <h2 class="text-lg font-semibold text-white mb-4">Build Order</h2>
          <BuildOrder :items="selectedAnalysis.build_order" />
        </div>

        <!-- Suggestions -->
        <div v-if="selectedAnalysis.suggestions?.length" class="bg-gray-800 rounded-lg p-6">
          <h2 class="text-lg font-semibold text-white mb-4">Verbesserungsvorschläge</h2>
          <Suggestions :items="selectedAnalysis.suggestions" />
        </div>
      </div>

      <!-- Tab: Strategic Analysis -->
      <div v-if="activeTab === 'strategic'">
        <!-- Loading State -->
        <div v-if="store.loadingStrategic" class="text-center py-16">
          <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-purple-400 mx-auto"></div>
          <p class="mt-4 text-gray-400">Lade strategische Analyse...</p>
        </div>

        <!-- Strategic Analysis Content -->
        <StrategicAnalysis
          v-else-if="store.strategicAnalysis?.analysis"
          :data="store.strategicAnalysis.analysis"
          :player-name="selectedPlayer?.name || ''"
        />

        <!-- No Strategic Analysis -->
        <div v-else class="bg-gray-800 rounded-lg p-8 text-center">
          <p class="text-gray-400">
            Strategische Analyse ist nur für Spiele mit eindeutigem Gewinner/Verlierer verfügbar.
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
