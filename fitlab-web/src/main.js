import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'
import './style.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

// Auth guard - after pinia is ready
router.beforeEach(async (to, from, next) => {
  const { useAuthStore } = await import('./stores/auth')
  const auth = useAuthStore()
  
  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    try {
      await auth.fetchMe()
    } catch {
      // No tenemos sesión
    }
  }
  
  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    next({ name: 'login' })
  } else if (to.meta.requiresProfessor && auth.user?.role !== 'professor') {
    next({ name: 'routines' })
  } else if ((to.name === 'login' || to.name === 'register') && auth.isLoggedIn) {
    next('/app/routines')
  } else {
    next()
  }
})

app.mount('#app')