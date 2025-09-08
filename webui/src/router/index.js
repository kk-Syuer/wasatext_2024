import { createRouter, createWebHashHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'

const TOKEN_KEY = 'wasa_token'
const hasToken = () => !!localStorage.getItem(TOKEN_KEY)

const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/login', name: 'login', component: LoginView, meta: { plain: true } },
    { path: '/', name: 'home', component: HomeView, meta: { requiresAuth: true } },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

router.beforeEach((to, _from, next) => {
  const authed = hasToken()
  if (to.meta.requiresAuth && !authed) return next({ name: 'login' })
  if (to.name === 'login' && authed) return next({ name: 'home' })
  next()
})

export default router
