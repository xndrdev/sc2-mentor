<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')

async function handleSubmit() {
  if (!email.value || !password.value) {
    return
  }

  const success = await authStore.login(email.value, password.value)
  if (success) {
    router.push('/mentor')
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center px-4">
    <div class="max-w-md w-full space-y-8">
      <div class="text-center">
        <h1 class="text-3xl font-bold text-white">SC2 Analytics</h1>
        <p class="mt-2 text-gray-400">Melde dich an, um dein Mentor-Dashboard zu sehen</p>
      </div>

      <form @submit.prevent="handleSubmit" class="mt-8 space-y-6 bg-gray-800 p-8 rounded-lg">
        <div v-if="authStore.error" class="bg-red-500/20 border border-red-500 text-red-300 px-4 py-3 rounded">
          {{ authStore.error }}
        </div>

        <div>
          <label for="email" class="block text-sm font-medium text-gray-300">Email</label>
          <input
            id="email"
            v-model="email"
            type="email"
            required
            class="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            placeholder="deine@email.de"
          />
        </div>

        <div>
          <label for="password" class="block text-sm font-medium text-gray-300">Passwort</label>
          <input
            id="password"
            v-model="password"
            type="password"
            required
            class="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            placeholder="********"
          />
        </div>

        <button
          type="submit"
          :disabled="authStore.loading"
          class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <span v-if="authStore.loading">Anmelden...</span>
          <span v-else>Anmelden</span>
        </button>

        <p class="text-center text-sm text-gray-400">
          Noch kein Konto?
          <router-link to="/register" class="text-blue-400 hover:text-blue-300">Registrieren</router-link>
        </p>
      </form>
    </div>
  </div>
</template>
