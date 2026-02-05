<script setup lang="ts">
import { ref } from 'vue'

defineProps<{
  loading?: boolean
}>()

const emit = defineEmits<{
  upload: [file: File]
}>()

const isDragging = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

function handleDrop(e: DragEvent) {
  isDragging.value = false
  const file = e.dataTransfer?.files[0]
  if (file && file.name.endsWith('.SC2Replay')) {
    emit('upload', file)
  }
}

function handleFileSelect(e: Event) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) {
    emit('upload', file)
  }
}

function openFileDialog() {
  fileInput.value?.click()
}
</script>

<template>
  <div
    class="relative border-2 border-dashed rounded-lg p-8 text-center transition-colors"
    :class="{
      'border-blue-500 bg-blue-500/10': isDragging,
      'border-gray-600 hover:border-gray-500': !isDragging,
      'opacity-50 pointer-events-none': loading,
    }"
    @dragover.prevent="isDragging = true"
    @dragleave="isDragging = false"
    @drop.prevent="handleDrop"
    @click="openFileDialog"
  >
    <input
      ref="fileInput"
      type="file"
      accept=".SC2Replay"
      class="hidden"
      @change="handleFileSelect"
    />

    <div v-if="loading" class="flex flex-col items-center">
      <div class="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-400"></div>
      <p class="mt-4 text-gray-400">Verarbeite Replay...</p>
    </div>

    <div v-else class="flex flex-col items-center">
      <svg
        class="w-12 h-12 text-gray-500"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
        />
      </svg>
      <p class="mt-4 text-gray-300">
        <span class="font-medium text-blue-400">Klicken</span> oder Datei hierher ziehen
      </p>
      <p class="mt-1 text-gray-500 text-sm">.SC2Replay Dateien</p>
    </div>
  </div>
</template>
