import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '@/views/HomeView.vue'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
      meta: { requiresAuth: true },
    },
    {
      path: '/replay/:id',
      name: 'replay',
      component: () => import('@/views/ReplayView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/trends',
      name: 'trends',
      component: () => import('@/views/TrendsView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { guest: true },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/RegisterView.vue'),
      meta: { guest: true },
    },
    {
      path: '/mentor',
      name: 'mentor',
      component: () => import('@/views/MentorView.vue'),
      meta: { requiresAuth: true },
    },
  ],
})

// Navigation Guards
router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore()

  // Prüfe Auth-Status wenn Token vorhanden
  if (authStore.token && !authStore.user) {
    await authStore.checkAuth()
  }

  // Route erfordert Authentifizierung
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({ name: 'login', query: { redirect: to.fullPath } })
    return
  }

  // Route ist nur für Gäste (Login/Register)
  if (to.meta.guest && authStore.isAuthenticated) {
    next({ name: 'mentor' })
    return
  }

  next()
})

export default router
