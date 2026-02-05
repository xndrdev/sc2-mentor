<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useReplayStore } from '@/stores/replays'
import UploadZone from '@/components/UploadZone.vue'
import ReplayList from '@/components/ReplayList.vue'

const router = useRouter()
const store = useReplayStore()
const uploadError = ref<string | null>(null)
const lastUploadedReplayId = ref<number | null>(null)

onMounted(() => {
  store.fetchReplays()
})

async function handleUpload(file: File) {
  uploadError.value = null
  try {
    const result = await store.upload(file)
    lastUploadedReplayId.value = result.replay_id

    // Wenn keine Spielerauswahl nötig, direkt zur Analyse
    if (!result.needs_player_selection) {
      router.push(`/replay/${result.replay_id}`)
    }
    // Sonst wird der Modal automatisch angezeigt (pendingClaimReplay ist gesetzt)
  } catch (e) {
    uploadError.value = e instanceof Error ? e.message : 'Upload fehlgeschlagen'
  }
}

async function handlePlayerSelect(playerId: number) {
  if (!store.pendingClaimReplay) return

  try {
    await store.claim(store.pendingClaimReplay.id, playerId)
    // Nach erfolgreicher Zuordnung zur Analyse
    if (lastUploadedReplayId.value) {
      router.push(`/replay/${lastUploadedReplayId.value}`)
    }
  } catch (e) {
    uploadError.value = e instanceof Error ? e.message : 'Zuordnung fehlgeschlagen'
  }
}

function handleSkipPlayerSelect() {
  store.clearPendingClaim()
  if (lastUploadedReplayId.value) {
    router.push(`/replay/${lastUploadedReplayId.value}`)
  }
}

function handleReplayClick(replayId: number) {
  router.push(`/replay/${replayId}`)
}

async function handleDeleteReplay(replayId: number) {
  if (!confirm('Replay wirklich löschen? Diese Aktion kann nicht rückgängig gemacht werden.')) {
    return
  }
  try {
    await store.remove(replayId)
  } catch {
    // Fehler wird im Store angezeigt
  }
}
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
    <!-- Upload Section -->
    <section>
      <h2 class="text-xl font-semibold text-white mb-4">Replay hochladen</h2>
      <UploadZone @upload="handleUpload" :loading="store.loading" />
      <p v-if="uploadError" class="mt-2 text-red-400 text-sm">{{ uploadError }}</p>
    </section>

    <!-- Replay List Section -->
    <section>
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-xl font-semibold text-white">Deine Replays</h2>
        <span class="text-gray-400 text-sm">{{ store.total }} Replays</span>
      </div>

      <div v-if="store.loading && !store.replays.length" class="text-center py-8">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-400 mx-auto"></div>
        <p class="mt-2 text-gray-400">Lade Replays...</p>
      </div>

      <div v-else-if="store.error" class="bg-red-900/50 border border-red-700 rounded-lg p-4">
        <p class="text-red-400">{{ store.error }}</p>
      </div>

      <div v-else-if="!store.replays.length" class="text-center py-8 bg-gray-800 rounded-lg">
        <p class="text-gray-400">Noch keine Replays hochgeladen.</p>
        <p class="text-gray-500 text-sm mt-1">Ziehe eine .SC2Replay Datei hierher.</p>
      </div>

      <ReplayList
        v-else
        :replays="store.replays"
        @click="handleReplayClick"
        @delete="handleDeleteReplay"
      />
    </section>

    <!-- Player Selection Modal -->
    <div v-if="store.pendingClaimReplay" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-gray-800 rounded-lg p-6 max-w-md w-full mx-4">
        <h3 class="text-xl font-semibold text-white mb-2">Welcher Spieler bist du?</h3>
        <p class="text-gray-400 text-sm mb-6">
          Wähle deinen Spieler aus, um das Replay deinem Fortschritt zuzuordnen.
        </p>

        <div class="space-y-3 mb-6">
          <button
            v-for="player in store.pendingClaimReplay.players"
            :key="player.player_id"
            @click="handlePlayerSelect(player.player_id)"
            class="w-full flex items-center justify-between p-4 bg-gray-700 hover:bg-gray-600 rounded-lg transition-colors"
          >
            <div class="flex items-center gap-3">
              <span class="text-2xl">
                {{ player.race === 'Terran' ? '🔧' : player.race === 'Zerg' ? '🐛' : player.race === 'Protoss' ? '⚡' : '❓' }}
              </span>
              <div class="text-left">
                <div class="text-white font-medium">{{ player.name }}</div>
                <div class="text-sm text-gray-400">{{ player.race }}</div>
              </div>
            </div>
            <span :class="player.result === 'Win' ? 'text-green-400' : 'text-red-400'">
              {{ player.result === 'Win' ? 'Sieg' : 'Niederlage' }}
            </span>
          </button>
        </div>

        <button
          @click="handleSkipPlayerSelect"
          class="w-full py-2 text-gray-400 hover:text-white text-sm transition-colors"
        >
          Überspringen (nicht zuordnen)
        </button>
      </div>
    </div>
  </div>
</template>
