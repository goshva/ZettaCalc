import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'home', component: () => import('@/pages/HomePage.vue') },

  // Приложение «Интерьеры»: hub, три ленты, карточки, избранное, загрузка, подбор
  { path: '/interiors', name: 'interiors', component: () => import('@/pages/InteriorsHubPage.vue') },
  { path: '/interiors/ideas', name: 'interiors-ideas', component: () => import('@/pages/InteriorsPage.vue'), meta: { kind: 'ideas' } },
  { path: '/interiors/rooms', name: 'interiors-rooms', component: () => import('@/pages/InteriorsPage.vue'), meta: { kind: 'rooms' } },
  { path: '/interiors/apartments', name: 'interiors-apartments', component: () => import('@/pages/InteriorsPage.vue'), meta: { kind: 'apartments' } },
  { path: '/interiors/idea/:id', name: 'interiors-idea', component: () => import('@/pages/InteriorsIdeaPage.vue') },
  { path: '/interiors/room/:id', name: 'interiors-room', component: () => import('@/pages/InteriorsRoomPage.vue') },
  { path: '/interiors/apartment/:id', name: 'interiors-apartment', component: () => import('@/pages/InteriorsApartmentPage.vue') },
  { path: '/interiors/matching/:id', name: 'interiors-matching', component: () => import('@/pages/InteriorsMatchingPage.vue') },
  { path: '/interiors/favorites', name: 'interiors-favorites', component: () => import('@/pages/InteriorsFavoritesPage.vue') },
  { path: '/interiors/upload', name: 'interiors-upload', component: () => import('@/pages/InteriorsUploadPage.vue') },

  // Совместимость с прежними адресами главной прототипа
  { path: '/ideas', redirect: '/interiors/ideas' },
  { path: '/favorites', redirect: '/interiors/favorites' },

  { path: '/appointment', name: 'appointment', component: () => import('@/pages/AppointmentPage.vue') },
  { path: '/catalog', name: 'catalog', component: () => import('@/pages/CatalogPage.vue') },
  { path: '/estimate', name: 'estimate', component: () => import('@/pages/EstimatePage.vue') },
  { path: '/flow', name: 'flow', component: () => import('@/pages/FlowPage.vue') },
  { path: '/cpr', name: 'cpr', component: () => import('@/pages/CprPage.vue'), meta: { requiresAuth: true } },
  { path: '/cart', name: 'cart', component: () => import('@/pages/CartPage.vue'), meta: { requiresAuth: true } },
  { path: '/login', name: 'login', component: () => import('@/pages/LoginPage.vue') },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({
  // webHistory, а не hash: в оригинале хэш занят якорями секций (#ideas, #how, #center)
  history: createWebHistory(),
  routes,
  scrollBehavior(to, from, saved) {
    if (saved) return saved
    // якорь секции главной — скроллим плавно, как в оригинале
    if (to.hash) return { el: to.hash, top: 96, behavior: 'smooth' }
    return { top: 0 }
  },
})

// Guard: пускать на защищённые страницы только с токеном
router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
})

export default router
