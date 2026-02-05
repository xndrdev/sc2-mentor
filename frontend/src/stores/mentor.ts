import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { MentorDashboard, Goal, GoalTemplate, DailyProgress, WeeklyReport, CoachingFocus } from '@/api/client'
import {
  getMentorDashboard,
  getGoals,
  createGoal as apiCreateGoal,
  deleteGoal as apiDeleteGoal,
  getProgress,
  getWeeklyReport as apiGetWeeklyReport,
  setCoachingFocus as apiSetCoachingFocus,
  getGoalTemplates,
} from '@/api/client'

export const useMentorStore = defineStore('mentor', () => {
  const dashboard = ref<MentorDashboard | null>(null)
  const goals = ref<Goal[]>([])
  const goalTemplates = ref<GoalTemplate[]>([])
  const progressHistory = ref<DailyProgress[]>([])
  const weeklyReport = ref<WeeklyReport | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchDashboard() {
    loading.value = true
    error.value = null
    try {
      dashboard.value = await getMentorDashboard()
      // Aktualisiere auch die lokalen Ziele
      if (dashboard.value.active_goals) {
        goals.value = dashboard.value.active_goals
      }
    } catch (e: unknown) {
      const axiosError = e as { response?: { data?: { error?: string } } }
      error.value = axiosError.response?.data?.error || 'Fehler beim Laden des Dashboards'
    } finally {
      loading.value = false
    }
  }

  async function fetchGoals() {
    loading.value = true
    error.value = null
    try {
      const data = await getGoals()
      goals.value = data.goals || []
      goalTemplates.value = data.templates || []
    } catch (e: unknown) {
      const axiosError = e as { response?: { data?: { error?: string } } }
      error.value = axiosError.response?.data?.error || 'Fehler beim Laden der Ziele'
    } finally {
      loading.value = false
    }
  }

  async function createGoal(goalType: string, metricName: string, targetValue: number, comparison?: string) {
    loading.value = true
    error.value = null
    try {
      const newGoal = await apiCreateGoal(goalType, metricName, targetValue, comparison)
      goals.value.push(newGoal)
      return newGoal
    } catch (e: unknown) {
      const axiosError = e as { response?: { data?: { error?: string } } }
      error.value = axiosError.response?.data?.error || 'Fehler beim Erstellen des Ziels'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function deleteGoal(goalId: number) {
    try {
      await apiDeleteGoal(goalId)
      goals.value = goals.value.filter(g => g.id !== goalId)
    } catch (e: unknown) {
      const axiosError = e as { response?: { data?: { error?: string } } }
      error.value = axiosError.response?.data?.error || 'Fehler beim Löschen des Ziels'
      throw e
    }
  }

  async function fetchProgress(days = 14) {
    loading.value = true
    error.value = null
    try {
      const data = await getProgress(days)
      progressHistory.value = data.progress || []
    } catch (e: unknown) {
      const axiosError = e as { response?: { data?: { error?: string } } }
      error.value = axiosError.response?.data?.error || 'Fehler beim Laden des Fortschritts'
    } finally {
      loading.value = false
    }
  }

  async function fetchWeeklyReport(generate = false) {
    loading.value = true
    error.value = null
    try {
      weeklyReport.value = await apiGetWeeklyReport(generate)
    } catch (e: unknown) {
      const axiosError = e as { response?: { data?: { error?: string } } }
      error.value = axiosError.response?.data?.error || 'Fehler beim Laden des Wochenberichts'
    } finally {
      loading.value = false
    }
  }

  async function setCoachingFocus(focusArea: string, description: string): Promise<CoachingFocus | null> {
    loading.value = true
    error.value = null
    try {
      const focus = await apiSetCoachingFocus(focusArea, description)
      if (dashboard.value) {
        dashboard.value.current_focus = focus
      }
      return focus
    } catch (e: unknown) {
      const axiosError = e as { response?: { data?: { error?: string } } }
      error.value = axiosError.response?.data?.error || 'Fehler beim Setzen des Fokus'
      return null
    } finally {
      loading.value = false
    }
  }

  async function fetchGoalTemplates() {
    try {
      goalTemplates.value = await getGoalTemplates()
    } catch {
      // Ignoriere Fehler, Templates sind optional
    }
  }

  function reset() {
    dashboard.value = null
    goals.value = []
    progressHistory.value = []
    weeklyReport.value = null
    error.value = null
  }

  return {
    dashboard,
    goals,
    goalTemplates,
    progressHistory,
    weeklyReport,
    loading,
    error,
    fetchDashboard,
    fetchGoals,
    createGoal,
    deleteGoal,
    fetchProgress,
    fetchWeeklyReport,
    setCoachingFocus,
    fetchGoalTemplates,
    reset,
  }
})
