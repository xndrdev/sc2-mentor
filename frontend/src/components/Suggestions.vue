<script setup lang="ts">
import type { Suggestion } from '@/api/client'

defineProps<{
  items: Suggestion[]
}>()

function getPriorityClass(priority: string): string {
  switch (priority) {
    case 'high':
      return 'bg-red-500/20 border-red-500 text-red-400'
    case 'medium':
      return 'bg-yellow-500/20 border-yellow-500 text-yellow-400'
    default:
      return 'bg-blue-500/20 border-blue-500 text-blue-400'
  }
}

function getPriorityText(priority: string): string {
  switch (priority) {
    case 'high':
      return 'Hoch'
    case 'medium':
      return 'Mittel'
    default:
      return 'Niedrig'
  }
}

function getCategoryIcon(category: string): string {
  switch (category) {
    case 'macro':
      return '📦'
    case 'micro':
      return '⚔️'
    case 'strategy':
      return '🎯'
    default:
      return '💡'
  }
}

function formatTime(seconds: number): string {
  if (!seconds) return ''
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}
</script>

<template>
  <div class="space-y-3">
    <div
      v-for="(item, index) in items"
      :key="index"
      :class="getPriorityClass(item.priority)"
      class="border-l-4 rounded-r-lg p-4"
    >
      <div class="flex items-start gap-3">
        <span class="text-xl">{{ getCategoryIcon(item.category) }}</span>
        <div class="flex-1">
          <div class="flex items-center gap-2">
            <h4 class="font-medium text-white">{{ item.title }}</h4>
            <span class="text-xs px-2 py-0.5 rounded bg-gray-700 text-gray-300">
              {{ getPriorityText(item.priority) }}
            </span>
          </div>
          <p class="text-gray-400 text-sm mt-1">{{ item.description }}</p>
          <div class="flex items-center gap-4 mt-2 text-sm">
            <span v-if="item.target_value" class="text-gray-500">
              Ziel: <span class="text-gray-300">{{ item.target_value }}</span>
            </span>
            <span v-if="item.timestamp" class="text-gray-500">
              @ {{ formatTime(item.timestamp) }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
