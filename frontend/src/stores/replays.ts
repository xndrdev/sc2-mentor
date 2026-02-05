import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Replay, ReplayAnalysis, StrategicAnalysisResponse, UploadResponse } from '@/api/client'
import { listReplays, getReplayAnalysis, uploadReplay, getStrategicAnalysis, claimReplay, deleteReplay } from '@/api/client'

export const useReplayStore = defineStore('replays', () => {
  const replays = ref<Replay[]>([])
  const currentAnalysis = ref<ReplayAnalysis | null>(null)
  const strategicAnalysis = ref<StrategicAnalysisResponse | null>(null)
  const loading = ref(false)
  const loadingStrategic = ref(false)
  const error = ref<string | null>(null)
  const total = ref(0)

  // Für Spielerauswahl nach Upload
  const pendingClaimReplay = ref<Replay | null>(null)

  async function fetchReplays(limit = 20, offset = 0) {
    loading.value = true
    error.value = null
    try {
      const data = await listReplays(limit, offset)
      replays.value = data.replays || []
      total.value = data.total || 0
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Fehler beim Laden der Replays'
      replays.value = []
    } finally {
      loading.value = false
    }
  }

  async function fetchAnalysis(replayId: number) {
    loading.value = true
    error.value = null
    try {
      currentAnalysis.value = await getReplayAnalysis(replayId)
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Fehler beim Laden der Analyse'
      currentAnalysis.value = null
    } finally {
      loading.value = false
    }
  }

  async function upload(file: File): Promise<UploadResponse> {
    loading.value = true
    error.value = null
    try {
      const result = await uploadReplay(file)
      // Aktualisiere die Liste
      await fetchReplays()

      // Wenn Spielerauswahl nötig, speichere das Replay
      if (result.needs_player_selection && result.replay) {
        pendingClaimReplay.value = result.replay
      }

      return result
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Fehler beim Upload'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function claim(replayId: number, playerId: number) {
    try {
      await claimReplay(replayId, playerId)
      pendingClaimReplay.value = null
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Fehler beim Zuordnen'
      throw e
    }
  }

  function clearPendingClaim() {
    pendingClaimReplay.value = null
  }

  async function remove(replayId: number) {
    try {
      await deleteReplay(replayId)
      // Aus lokaler Liste entfernen
      replays.value = replays.value.filter(r => r.id !== replayId)
      total.value = Math.max(0, total.value - 1)
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Fehler beim Löschen'
      throw e
    }
  }

  async function fetchStrategicAnalysis(replayId: number) {
    loadingStrategic.value = true
    try {
      strategicAnalysis.value = await getStrategicAnalysis(replayId)
    } catch (e) {
      // Strategische Analyse ist optional, also kein Fehler
      strategicAnalysis.value = null
    } finally {
      loadingStrategic.value = false
    }
  }

  return {
    replays,
    currentAnalysis,
    strategicAnalysis,
    loading,
    loadingStrategic,
    error,
    total,
    pendingClaimReplay,
    fetchReplays,
    fetchAnalysis,
    fetchStrategicAnalysis,
    upload,
    claim,
    clearPendingClaim,
    remove,
  }
})
