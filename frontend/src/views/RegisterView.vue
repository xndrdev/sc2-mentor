<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const passwordConfirm = ref('')
const sc2PlayerName = ref('')
const localError = ref('')

async function handleSubmit() {
  localError.value = ''

  if (!email.value || !password.value || !sc2PlayerName.value) {
    localError.value = 'Bitte alle Felder ausfüllen'
    return
  }

  if (password.value !== passwordConfirm.value) {
    localError.value = 'Passwörter stimmen nicht überein'
    return
  }

  if (password.value.length < 8) {
    localError.value = 'Passwort muss mindestens 8 Zeichen haben'
    return
  }

  const success = await authStore.register(email.value, password.value, sc2PlayerName.value)
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
        <p class="mt-2 text-gray-400">Erstelle ein Konto und starte dein Training</p>
      </div>

      <form @submit.prevent="handleSubmit" class="mt-8 space-y-6 bg-gray-800 p-8 rounded-lg">
        <div v-if="localError || authStore.error" class="bg-red-500/20 border border-red-500 text-red-300 px-4 py-3 rounded">
          {{ localError || authStore.error }}
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
          <label for="sc2PlayerName" class="block text-sm font-medium text-gray-300">SC2 Spielername</label>
          <input
            id="sc2PlayerName"
            v-model="sc2PlayerName"
            type="text"
            required
            class="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            placeholder="Dein Battle.net Name"
          />
          <p class="mt-1 text-xs text-gray-400">Dieser Name wird verwendet, um deine Replays automatisch zuzuordnen</p>
        </div>

        <div>
          <label for="password" class="block text-sm font-medium text-gray-300">Passwort</label>
          <input
            id="password"
            v-model="password"
            type="password"
            required
            minlength="8"
            class="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            placeholder="Mindestens 8 Zeichen"
          />
        </div>

        <div>
          <label for="passwordConfirm" class="block text-sm font-medium text-gray-300">Passwort bestätigen</label>
          <input
            id="passwordConfirm"
            v-model="passwordConfirm"
            type="password"
            required
            class="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            placeholder="Passwort wiederholen"
          />
        </div>

        <button
          type="submit"
          :disabled="authStore.loading"
          class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <span v-if="authStore.loading">Registrieren...</span>
          <span v-else>Konto erstellen</span>
        </button>

        <p class="text-center text-sm text-gray-400">
          Bereits ein Konto?
          <router-link to="/login" class="text-blue-400 hover:text-blue-300">Anmelden</router-link>
        </p>
      </form>
    </div>
  </div>
</template>
