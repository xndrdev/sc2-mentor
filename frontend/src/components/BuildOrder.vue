<script setup lang="ts">
import type { BuildOrderItem } from '@/api/client'

defineProps<{
  items: BuildOrderItem[]
}>()

function formatTime(seconds: number): string {
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

function getActionClass(action: string): string {
  switch (action) {
    case 'Build':
      return 'bg-blue-500/20 text-blue-400'
    case 'Train':
    case 'Train Worker':
      return 'bg-green-500/20 text-green-400'
    case 'Upgrade':
      return 'bg-purple-500/20 text-purple-400'
    default:
      return 'bg-gray-500/20 text-gray-400'
  }
}
</script>

<template>
  <div class="max-h-96 overflow-y-auto">
    <table class="w-full text-sm">
      <thead class="sticky top-0 bg-gray-800">
        <tr class="text-gray-400 text-left">
          <th class="pb-2 pr-4">Zeit</th>
          <th class="pb-2 pr-4">Supply</th>
          <th class="pb-2 pr-4">Aktion</th>
          <th class="pb-2">Einheit/Gebäude</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(item, index) in items"
          :key="index"
          class="border-t border-gray-700"
        >
          <td class="py-2 pr-4 text-gray-400">{{ formatTime(item.time) }}</td>
          <td class="py-2 pr-4 text-white">{{ item.supply }}</td>
          <td class="py-2 pr-4">
            <span :class="getActionClass(item.action)" class="px-2 py-0.5 rounded text-xs">
              {{ item.action }}
            </span>
          </td>
          <td class="py-2 text-white">{{ item.unit_or_building }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
