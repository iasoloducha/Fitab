<template>
  <div class="app-layout">
    <nav class="app-nav">
      <div class="nav-brand">
        <router-link to="/app">
          <img src="/images/logo.jpg" alt="Fitlab" class="logo-img" />
        </router-link>
      </div>
      <button class="nav-toggle" @click="menuOpen = !menuOpen">
        {{ menuOpen ? '✕' : '☰' }}
      </button>
      <div class="nav-links" :class="{ open: menuOpen }">
        <router-link to="/app/routines" @click="menuOpen = false">Rutinas</router-link>
        <router-link v-if="auth.isProfessor" to="/app/catalog" @click="menuOpen = false">Ejercicios</router-link>
        <router-link to="/app/progress" @click="menuOpen = false">
          {{ auth.isProfessor ? 'Progreso de Alumnos' : 'Mi Progreso' }}
        </router-link>
        <router-link to="/app/profile" @click="menuOpen = false">Perfil</router-link>
        <button @click="handleLogout" class="btn-logout">Salir</button>
      </div>
    </nav>
    
    <main class="app-main">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const menuOpen = ref(false)

onMounted(async () => {
  if (!auth.isLoggedIn) {
    await auth.fetchMe()
  }
})

async function handleLogout() {
  menuOpen.value = false
  await auth.logout()
  router.push({ name: 'login' })
}
</script>

<style scoped>
.app-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.app-nav {
  position: relative;
  background: var(--color-dark);
  color: white;
  padding: 1rem;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.nav-brand a {
  display: flex;
  align-items: center;
}

.logo-img {
  height: 40px;
  width: auto;
  object-fit: contain;
}

.nav-toggle {
  display: none;
  margin-left: auto;
  background: none;
  border: none;
  color: white;
  font-size: 1.5rem;
  cursor: pointer;
}

.nav-links {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.nav-links a {
  padding: 0.5rem 1rem;
  border-radius: 4px;
  transition: background 0.2s;
}

.nav-links a:hover,
.nav-links a.router-link-active {
  background: rgba(255, 255, 255, 0.1);
}

.btn-logout {
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  margin-left: 0.5rem;
}

.btn-logout:hover {
  background: rgba(255, 255, 255, 0.1);
}

.app-main {
  flex: 1;
  padding: 1rem;
}

/* Mobile responsive */
@media (max-width: 768px) {
  .nav-toggle {
    display: block;
  }
  
  .nav-links {
    display: none;
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    flex-direction: column;
    background: var(--color-dark);
    padding: 1rem;
    gap: 0.5rem;
  }
  
  .nav-links.open {
    display: flex;
  }
  
  .nav-links a,
  .btn-logout {
    width: 100%;
    text-align: center;
  }
  
  .app-main {
    padding: 0.75rem;
  }
}
</style>