<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useMentorStore } from '@/stores/mentor'
import { useAuthStore } from '@/stores/auth'
import QuickStats from '@/components/mentor/QuickStats.vue'
import GoalCard from '@/components/mentor/GoalCard.vue'
import ProgressChart from '@/components/mentor/ProgressChart.vue'
import WeeklyReportCard from '@/components/mentor/WeeklyReportCard.vue'

const mentorStore = useMentorStore()
const authStore = useAuthStore()

const showGoalModal = ref(false)
const newGoal = ref({
  goalType: 'daily',
  metricName: 'games_played',
  targetValue: 3,
  comparison: '>=',
})

onMounted(async () => {
  await mentorStore.fetchDashboard()
  await mentorStore.fetchGoalTemplates()
})

async function handleCreateGoal() {
  try {
    await mentorStore.createGoal(
      newGoal.value.goalType,
      newGoal.value.metricName,
      newGoal.value.targetValue,
      newGoal.value.comparison
    )
    showGoalModal.value = false
    // Reset Form
    newGoal.value = {
      goalType: 'daily',
      metricName: 'games_played',
      targetValue: 3,
      comparison: '>=',
    }
  } catch {
    // Error wird im Store gehandhabt
  }
}

async function handleDeleteGoal(goalId: number) {
  if (confirm('Ziel wirklich löschen?')) {
    await mentorStore.deleteGoal(goalId)
  }
}

function formatDuration(seconds: number): string {
  const minutes = Math.floor(seconds / 60)
  return `${minutes}m`
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('de-DE', {
    day: '2-digit',
    month: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 py-8">
    <!-- Header -->
    <div class="mb-8">
      <h1 class="text-3xl font-bold text-white">Dein persönlicher SC2 Coach</h1>
      <p class="text-gray-400 mt-2">
        Willkommen zurück, <span class="text-blue-400">{{ authStore.user?.sc2_player_name }}</span>!
        Verfolge deinen Fortschritt und erreiche deine Ziele.
      </p>
    </div>

    <!-- Loading -->
    <div v-if="mentorStore.loading && !mentorStore.dashboard" class="flex justify-center py-20">
      <div class="animate-spin rounded-full h-12 w-12 border-4 border-blue-500 border-t-transparent"></div>
    </div>

    <!-- Error -->
    <div v-else-if="mentorStore.error" class="bg-red-500/20 border border-red-500 text-red-300 px-4 py-3 rounded mb-6">
      {{ mentorStore.error }}
    </div>

    <!-- Dashboard Content -->
    <div v-else class="space-y-8">
      <!-- Quick Stats -->
      <QuickStats
        :today-stats="mentorStore.dashboard?.today_stats || null"
        :week-stats="mentorStore.dashboard?.week_stats || null"
      />

      <!-- Goals Section -->
      <div class="bg-gray-800 rounded-lg p-6">
        <div class="flex justify-between items-center mb-6">
          <h2 class="text-xl font-semibold text-white">Deine Ziele</h2>
          <button
            @click="showGoalModal = true"
            class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors flex items-center gap-2"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            Neues Ziel
          </button>
        </div>

        <div v-if="mentorStore.goals.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <GoalCard
            v-for="goal in mentorStore.goals"
            :key="goal.id"
            :goal="goal"
            @delete="handleDeleteGoal"
          />
        </div>
        <div v-else class="text-center py-8 text-gray-500">
          <p class="mb-4">Du hast noch keine aktiven Ziele.</p>
          <button
            @click="showGoalModal = true"
            class="text-blue-400 hover:text-blue-300"
          >
            Erstelle dein erstes Ziel
          </button>
        </div>
      </div>

      <!-- Progress Charts -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <ProgressChart
          :progress-data="mentorStore.dashboard?.progress_trend || []"
          metric="apm"
          title="APM Entwicklung"
        />
        <ProgressChart
          :progress-data="mentorStore.dashboard?.progress_trend || []"
          metric="supply_block"
          title="Supply Block %"
        />
      </div>

      <!-- Weekly Report -->
      <WeeklyReportCard
        v-if="mentorStore.dashboard?.weekly_report"
        :report="mentorStore.dashboard.weekly_report"
      />

      <!-- Recent Games -->
      <div class="bg-gray-800 rounded-lg p-6">
        <h2 class="text-xl font-semibold text-white mb-4">Letzte Spiele</h2>
        <div v-if="mentorStore.dashboard?.recent_games?.length" class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="text-left text-gray-400 text-sm">
                <th class="pb-3">Map</th>
                <th class="pb-3">Matchup</th>
                <th class="pb-3">Ergebnis</th>
                <th class="pb-3">APM</th>
                <th class="pb-3">SQ</th>
                <th class="pb-3">Dauer</th>
                <th class="pb-3">Datum</th>
                <th class="pb-3"></th>
              </tr>
            </thead>
            <tbody class="text-gray-300">
              <tr v-for="game in mentorStore.dashboard.recent_games" :key="game.replay_id" class="border-t border-gray-700">
                <td class="py-3">{{ game.map }}</td>
                <td class="py-3">
                  <span class="font-medium">{{ game.race?.charAt(0) || '?' }}</span>
                  <span class="text-gray-500">v</span>
                  <span class="font-medium">{{ game.enemy_race?.charAt(0) || '?' }}</span>
                </td>
                <td class="py-3">
                  <span :class="game.result === 'Win' ? 'text-green-400' : 'text-red-400'">
                    {{ game.result === 'Win' ? 'Sieg' : 'Niederlage' }}
                  </span>
                </td>
                <td class="py-3">{{ game.apm.toFixed(0) }}</td>
                <td class="py-3">{{ game.sq.toFixed(0) }}</td>
                <td class="py-3">{{ formatDuration(game.duration) }}</td>
                <td class="py-3 text-sm text-gray-400">{{ formatDate(game.played_at) }}</td>
                <td class="py-3">
                  <router-link
                    :to="`/replay/${game.replay_id}`"
                    class="text-blue-400 hover:text-blue-300 text-sm"
                  >
                    Analyse
                  </router-link>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-else class="text-center py-8 text-gray-500">
          Noch keine Spiele hochgeladen. Lade dein erstes Replay hoch!
        </div>
      </div>

      <!-- Coaching Focus -->
      <div v-if="mentorStore.dashboard?.current_focus" class="bg-gradient-to-r from-blue-500/20 to-purple-500/20 border border-blue-500/30 rounded-lg p-6">
        <h3 class="text-lg font-semibold text-white mb-2">Aktueller Fokus: {{ mentorStore.dashboard.current_focus.focus_area }}</h3>
        <p class="text-gray-300">{{ mentorStore.dashboard.current_focus.description }}</p>
      </div>
    </div>

    <!-- Goal Modal -->
    <div v-if="showGoalModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-gray-800 rounded-lg p-6 max-w-md w-full mx-4">
        <h3 class="text-xl font-semibold text-white mb-6">Neues Ziel erstellen</h3>

        <form @submit.prevent="handleCreateGoal" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-2">Zieltyp</label>
            <select
              v-model="newGoal.goalType"
              class="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="daily">Täglich</option>
              <option value="weekly">Wöchentlich</option>
            </select>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-300 mb-2">Metrik</label>
            <select
              v-model="newGoal.metricName"
              @change="newGoal.comparison = newGoal.metricName === 'supply_block' ? '<=' : '>='"
              class="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="games_played">Spiele gespielt</option>
              <option value="apm">APM</option>
              <option value="supply_block">Supply Block %</option>
              <option value="win_rate">Win Rate %</option>
              <option value="sq">Spending Quotient</option>
            </select>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-300 mb-2">Zielwert</label>
            <div class="flex gap-2">
              <select
                v-model="newGoal.comparison"
                class="px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value=">=">>=</option>
                <option value="<="><=</option>
                <option value=">">&gt;</option>
                <option value="<">&lt;</option>
              </select>
              <input
                v-model.number="newGoal.targetValue"
                type="number"
                min="0"
                step="1"
                class="flex-1 px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>

          <!-- Templates -->
          <div v-if="mentorStore.goalTemplates.length > 0" class="pt-4 border-t border-gray-700">
            <label class="block text-sm font-medium text-gray-300 mb-2">Schnellauswahl</label>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="template in mentorStore.goalTemplates.filter(t => t.goal_type === newGoal.goalType)"
                :key="template.name"
                type="button"
                @click="
                  newGoal.metricName = template.metric_name;
                  newGoal.comparison = template.comparison;
                  newGoal.targetValue = template.beginner;
                "
                class="px-3 py-1 text-sm bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors"
              >
                {{ template.name }}
              </button>
            </div>
          </div>

          <div class="flex gap-3 pt-4">
            <button
              type="button"
              @click="showGoalModal = false"
              class="flex-1 px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-lg transition-colors"
            >
              Abbrechen
            </button>
            <button
              type="submit"
              :disabled="mentorStore.loading"
              class="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors disabled:opacity-50"
            >
              Erstellen
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
