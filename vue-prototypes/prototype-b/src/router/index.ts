import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'home', component: () => import('@/pages/HomePage.vue') },
  { path: '/gallery', name: 'gallery', component: () => import('@/pages/GalleryPage.vue') },
  { path: '/calculator', name: 'calculator', component: () => import('@/pages/CalculatorPage.vue') },
  { path: '/cpr', name: 'cpr', component: () => import('@/pages/CprPage.vue'), meta: { requiresAuth: true } },
  { path: '/cart', name: 'cart', component: () => import('@/pages/CartPage.vue'), meta: { requiresAuth: true } },
  { path: '/login', name: 'login', component: () => import('@/pages/LoginPage.vue') },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({ history: createWebHashHistory(), routes, scrollBehavior: () => ({ top: 0 }) })

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isAuthenticated) return { name: 'login', query: { redirect: to.fullPath } }
})

export default router
