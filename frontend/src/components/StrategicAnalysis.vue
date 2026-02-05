<script setup lang="ts">
interface MetricComparison {
  metric: string
  player_value: number
  enemy_value: number
  is_worse: boolean
}

interface SupplyBlockSummary {
  time: number
  duration: number
  severity: string
}

interface CriticalMoment {
  time: number
  player_loss: number
  enemy_loss: number
  assessment: string
  is_positive: boolean
}

interface IdentifiedProblem {
  title: string
  description: string
  priority: string
}

interface MatchupTips {
  opening: string[]
  mid_game: string[]
  timing: string[]
  late_game: string[]
}

interface ImprovementStep {
  category: string
  title: string
  description: string
}

interface StrategicAnalysisData {
  winner: string
  loser: string
  winner_race: string
  loser_race: string
  matchup: string
  metrics_comparison: MetricComparison[]
  supply_blocks: SupplyBlockSummary[]
  critical_moments: CriticalMoment[]
  problems: IdentifiedProblem[]
  matchup_tips: MatchupTips
  improvement_steps: ImprovementStep[]
  summary: string
}

defineProps<{
  data: StrategicAnalysisData
  playerName: string
}>()

function formatTime(seconds: number): string {
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

function formatMetricValue(metric: string, value: number): string {
  if (metric.includes('%') || metric.includes('Zeit')) {
    return `${value.toFixed(1)}%`
  }
  return Math.round(value).toString()
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="bg-gradient-to-r from-red-900/30 to-purple-900/30 rounded-lg p-6 border border-red-700/50">
      <h2 class="text-2xl font-bold text-white mb-2">Strategische Spielanalyse</h2>
      <p class="text-gray-300">
        {{ data.matchup }} · {{ data.loser }} ({{ data.loser_race }}) vs {{ data.winner }} ({{ data.winner_race }})
      </p>
    </div>

    <!-- Identifizierte Probleme -->
    <div v-if="data.problems?.length" class="bg-gray-800 rounded-lg p-6">
      <h3 class="text-lg font-semibold text-red-400 mb-4 flex items-center">
        <svg class="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"/>
        </svg>
        Identifizierte Probleme
      </h3>
      <div class="space-y-3">
        <div
          v-for="(problem, idx) in data.problems"
          :key="idx"
          class="flex items-start gap-3 p-3 rounded-lg"
          :class="{
            'bg-red-900/30 border border-red-700/50': problem.priority === 'high',
            'bg-yellow-900/30 border border-yellow-700/50': problem.priority === 'medium',
            'bg-gray-700/50': problem.priority === 'low'
          }"
        >
          <span
            class="text-xs font-bold px-2 py-0.5 rounded uppercase"
            :class="{
              'bg-red-500 text-white': problem.priority === 'high',
              'bg-yellow-500 text-black': problem.priority === 'medium',
              'bg-gray-500 text-white': problem.priority === 'low'
            }"
          >
            {{ problem.priority === 'high' ? 'Kritisch' : problem.priority === 'medium' ? 'Mittel' : 'Niedrig' }}
          </span>
          <div>
            <p class="font-medium text-white">{{ problem.title }}</p>
            <p class="text-sm text-gray-400">{{ problem.description }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Metriken-Vergleich -->
    <div v-if="data.metrics_comparison?.length" class="bg-gray-800 rounded-lg p-6">
      <h3 class="text-lg font-semibold text-white mb-4">Metriken-Vergleich</h3>
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead>
            <tr class="text-gray-400 text-sm border-b border-gray-700">
              <th class="text-left py-2 px-3">Metrik</th>
              <th class="text-right py-2 px-3">{{ data.loser }}</th>
              <th class="text-right py-2 px-3">{{ data.winner }}</th>
              <th class="w-8"></th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(metric, idx) in data.metrics_comparison"
              :key="idx"
              class="border-b border-gray-700/50"
            >
              <td class="py-2 px-3 text-gray-300">{{ metric.metric }}</td>
              <td class="py-2 px-3 text-right font-mono" :class="metric.is_worse ? 'text-red-400' : 'text-white'">
                {{ formatMetricValue(metric.metric, metric.player_value) }}
              </td>
              <td class="py-2 px-3 text-right font-mono text-white">
                {{ formatMetricValue(metric.metric, metric.enemy_value) }}
              </td>
              <td class="py-2 px-1">
                <span v-if="metric.is_worse" class="text-red-400" title="Schlechter als Gegner">!</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Supply Blocks -->
    <div v-if="data.supply_blocks?.length" class="bg-gray-800 rounded-lg p-6">
      <h3 class="text-lg font-semibold text-orange-400 mb-4">Deine Supply Blocks</h3>
      <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3">
        <div
          v-for="(block, idx) in data.supply_blocks"
          :key="idx"
          class="p-3 rounded-lg text-center"
          :class="{
            'bg-red-900/40 border border-red-700': block.severity === 'high',
            'bg-orange-900/40 border border-orange-700': block.severity === 'medium',
            'bg-yellow-900/40 border border-yellow-700': block.severity === 'low'
          }"
        >
          <p class="text-white font-mono text-lg">{{ formatTime(block.time) }}</p>
          <p class="text-sm" :class="{
            'text-red-400': block.severity === 'high',
            'text-orange-400': block.severity === 'medium',
            'text-yellow-400': block.severity === 'low'
          }">
            {{ Math.round(block.duration) }}s
            <span v-if="block.severity === 'high'" class="font-bold">SCHWER</span>
            <span v-else-if="block.severity === 'medium'">mittel</span>
            <span v-else>leicht</span>
          </p>
        </div>
      </div>
    </div>

    <!-- Kritische Momente -->
    <div v-if="data.critical_moments?.length" class="bg-gray-800 rounded-lg p-6">
      <h3 class="text-lg font-semibold text-white mb-4">Kritische Momente / Kämpfe</h3>
      <div class="space-y-2">
        <div
          v-for="(moment, idx) in data.critical_moments"
          :key="idx"
          class="flex items-center gap-4 p-2 rounded"
          :class="moment.is_positive ? 'bg-green-900/20' : 'bg-red-900/20'"
        >
          <span class="font-mono text-gray-400 w-12">{{ formatTime(moment.time) }}</span>
          <span class="text-gray-300">
            Du verlierst <span class="text-red-400 font-bold">{{ moment.player_loss }}</span>,
            Gegner verliert <span class="text-green-400 font-bold">{{ moment.enemy_loss }}</span>
          </span>
          <span :class="moment.is_positive ? 'text-green-400' : 'text-red-400'">
            {{ moment.is_positive ? '↑' : '↓' }} {{ moment.assessment }}
          </span>
        </div>
      </div>
    </div>

    <!-- Matchup Tipps -->
    <div v-if="data.matchup_tips" class="bg-gray-800 rounded-lg p-6">
      <h3 class="text-lg font-semibold text-blue-400 mb-4">
        {{ data.loser_race }} vs {{ data.winner_race }} Tipps
      </h3>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div v-if="data.matchup_tips.opening?.length" class="bg-gray-700/50 rounded-lg p-4">
          <h4 class="font-medium text-white mb-2">Opening</h4>
          <ul class="text-sm text-gray-300 space-y-1">
            <li v-for="(tip, idx) in data.matchup_tips.opening" :key="idx">• {{ tip }}</li>
          </ul>
        </div>
        <div v-if="data.matchup_tips.mid_game?.length" class="bg-gray-700/50 rounded-lg p-4">
          <h4 class="font-medium text-white mb-2">Mid Game</h4>
          <ul class="text-sm text-gray-300 space-y-1">
            <li v-for="(tip, idx) in data.matchup_tips.mid_game" :key="idx">• {{ tip }}</li>
          </ul>
        </div>
        <div v-if="data.matchup_tips.timing?.length" class="bg-gray-700/50 rounded-lg p-4">
          <h4 class="font-medium text-white mb-2">Timing Attacks</h4>
          <ul class="text-sm text-gray-300 space-y-1">
            <li v-for="(tip, idx) in data.matchup_tips.timing" :key="idx">• {{ tip }}</li>
          </ul>
        </div>
        <div v-if="data.matchup_tips.late_game?.length" class="bg-gray-700/50 rounded-lg p-4">
          <h4 class="font-medium text-white mb-2">Late Game</h4>
          <ul class="text-sm text-gray-300 space-y-1">
            <li v-for="(tip, idx) in data.matchup_tips.late_game" :key="idx">• {{ tip }}</li>
          </ul>
        </div>
      </div>
    </div>

    <!-- Verbesserungsschritte -->
    <div v-if="data.improvement_steps?.length" class="bg-gray-800 rounded-lg p-6">
      <h3 class="text-lg font-semibold text-green-400 mb-4">Konkrete Verbesserungsschritte</h3>
      <div class="space-y-3">
        <div
          v-for="(step, idx) in data.improvement_steps"
          :key="idx"
          class="flex items-start gap-3 p-3 bg-gray-700/50 rounded-lg"
        >
          <span class="text-xs font-bold px-2 py-0.5 rounded bg-blue-500 text-white uppercase">
            {{ step.category }}
          </span>
          <div>
            <p class="font-medium text-white">{{ step.title }}</p>
            <p class="text-sm text-gray-400">{{ step.description }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Zusammenfassung -->
    <div class="bg-gradient-to-r from-blue-900/30 to-purple-900/30 rounded-lg p-6 border border-blue-700/50">
      <h3 class="text-lg font-semibold text-white mb-3">Zusammenfassung</h3>
      <p class="text-gray-300 whitespace-pre-line">{{ data.summary }}</p>
      <div class="mt-4 p-3 bg-yellow-900/30 rounded-lg border border-yellow-700/50">
        <p class="text-yellow-300 text-sm">
          <strong>TIPP:</strong> Fokussiere dich auf EIN Problem pro Woche.
          Diese Woche: {{ data.problems?.[0]?.title?.split('(')[0] || 'Macro verbessern' }}!
        </p>
      </div>
    </div>
  </div>
</template>
