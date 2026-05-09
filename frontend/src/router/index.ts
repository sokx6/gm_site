import { createRouter, createWebHashHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  { path: '/', name: 'home', component: () => import('@/views/HomeView.vue') },
  { path: '/login', name: 'login', component: () => import('@/views/LoginView.vue') },
  { path: '/register', name: 'register', component: () => import('@/views/RegisterView.vue') },
  { path: '/pending', name: 'pending', component: () => import('@/views/PendingView.vue') },
  { path: '/admin', name: 'admin', component: () => import('@/views/AdminView.vue'), meta: { requiresAuth: true, requiresAdmin: true } },
  { path: '/admin/images', name: 'admin-images', component: () => import('@/views/AdminImagesView.vue'), meta: { requiresAuth: true, requiresAdmin: true } },
  { path: '/admin/users', name: 'admin-users', component: () => import('@/views/AdminUsersView.vue'), meta: { requiresAuth: true, requiresAdmin: true } },
  { path: '/admin/albums', name: 'admin-albums', component: () => import('@/views/AdminAlbumsView.vue'), meta: { requiresAuth: true, requiresAdmin: true } },
]

const router = createRouter({ history: createWebHashHistory(), routes })

router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isLoggedIn) return next('/login')
  if (to.meta.requiresAdmin && !auth.isAdmin) return next('/')
  if ((to.path === '/login' || to.path === '/register') && auth.isLoggedIn) return next('/')
  next()
})

export default router
