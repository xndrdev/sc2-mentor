<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useMentorStore } from '@/stores/mentor'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const mentorStore = useMentorStore()
const router = useRouter()

async function handleLogout() {
  await authStore.logout()
  mentorStore.reset()
  router.push('/')
}
</script>

<template>
  <div class="min-h-screen bg-gray-900">
    <!-- Navigation -->
    <nav class="bg-gray-800 border-b border-gray-700">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between h-16">
          <div class="flex items-center">
            <RouterLink to="/" class="flex items-center">
              <span class="text-2xl font-bold text-blue-400">SC2</span>
              <span class="text-2xl font-bold text-white ml-1">Analytics</span>
            </RouterLink>
          </div>
          <div class="flex items-center space-x-4">
            <RouterLink
              to="/"
              class="text-gray-300 hover:text-white px-3 py-2 rounded-md text-sm font-medium"
              active-class="text-white bg-gray-700"
            >
              Replays
            </RouterLink>
            <RouterLink
              to="/trends"
              class="text-gray-300 hover:text-white px-3 py-2 rounded-md text-sm font-medium"
              active-class="text-white bg-gray-700"
            >
              Trends
            </RouterLink>

            <!-- Auth Links -->
            <template v-if="authStore.isAuthenticated">
              <RouterLink
                to="/mentor"
                class="text-gray-300 hover:text-white px-3 py-2 rounded-md text-sm font-medium"
                active-class="text-white bg-gray-700"
              >
                Mentor
              </RouterLink>
              <div class="flex items-center gap-3 ml-4 pl-4 border-l border-gray-700">
                <span class="text-sm text-gray-400">{{ authStore.user?.sc2_player_name }}</span>
                <button
                  @click="handleLogout"
                  class="text-gray-400 hover:text-white text-sm"
                >
                  Abmelden
                </button>
              </div>
            </template>
            <template v-else>
              <RouterLink
                to="/login"
                class="text-gray-300 hover:text-white px-3 py-2 rounded-md text-sm font-medium ml-4"
              >
                Anmelden
              </RouterLink>
              <RouterLink
                to="/register"
                class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md text-sm font-medium"
              >
                Registrieren
              </RouterLink>
            </template>
          </div>
        </div>
      </div>
    </nav>

    <!-- Main Content -->
    <main>
      <RouterView />
    </main>
  </div>
</template>
