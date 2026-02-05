<script setup lang="ts">
import type { Replay } from '@/api/client'

defineProps<{
  replays: Replay[]
}>()

const emit = defineEmits<{
  click: [id: number]
  delete: [id: number]
}>()

function handleDelete(event: Event, replayId: number) {
  event.stopPropagation() // Verhindert, dass der Click-Event auf der Karte ausgelöst wird
  emit('delete', replayId)
}

function formatDuration(seconds: number): string {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleDateString('de-DE', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function getRaceClass(race: string): string {
  switch (race.toLowerCase()) {
    case 'terran':
      return 'text-blue-400'
    case 'zerg':
      return 'text-purple-400'
    case 'protoss':
      return 'text-yellow-400'
    default:
      return 'text-gray-400'
  }
}

function getResultClass(result: string): string {
  switch (result.toLowerCase()) {
    case 'win':
      return 'bg-green-500/20 text-green-400'
    case 'loss':
      return 'bg-red-500/20 text-red-400'
    default:
      return 'bg-gray-500/20 text-gray-400'
  }
}
</script>

<template>
  <div class="space-y-3">
    <div
      v-for="replay in replays"
      :key="replay.id"
      class="bg-gray-800 rounded-lg p-4 hover:bg-gray-750 cursor-pointer transition-colors border border-gray-700 hover:border-gray-600"
      @click="emit('click', replay.id)"
    >
      <div class="flex items-center justify-between">
        <div class="flex-1">
          <div class="flex items-center gap-3">
            <h3 class="font-medium text-white">{{ replay.map }}</h3>
            <span class="text-gray-500 text-sm">{{ formatDuration(replay.duration) }}</span>
          </div>

          <div class="mt-2 flex items-center gap-4">
            <template v-for="(player, index) in replay.players" :key="player.player_id">
              <span v-if="index > 0" class="text-gray-600">vs</span>
              <div class="flex items-center gap-2">
                <span :class="getRaceClass(player.race)" class="font-medium">
                  {{ player.name }}
                </span>
                <span class="text-gray-500 text-sm">({{ player.race }})</span>
                <span
                  v-if="player.is_human"
                  :class="getResultClass(player.result)"
                  class="text-xs px-2 py-0.5 rounded"
                >
                  {{ player.result }}
                </span>
              </div>
            </template>
          </div>
        </div>

        <div class="flex items-center gap-3">
          <div class="text-right text-sm text-gray-500">
            {{ formatDate(replay.played_at) }}
          </div>
          <button
            @click="handleDelete($event, replay.id)"
            class="p-2 text-gray-500 hover:text-red-400 hover:bg-red-500/10 rounded transition-colors"
            title="Replay löschen"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
