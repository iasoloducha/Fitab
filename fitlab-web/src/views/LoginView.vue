<template>
  <div class="auth-page">
    <div class="auth-card">
      <h1>FITLAB</h1>
      <p class="subtitle">Entrená smarter, progresá más rápido</p>
      
      <div v-if="auth.error" class="error">{{ auth.error }}</div>
      
      <form @submit.prevent="handleLogin">
        <div class="form-group">
          <label for="email">Email</label>
          <input 
            type="email" 
            id="email" 
            v-model="email" 
            required 
            placeholder="tu@email.com"
            autocomplete="email"
          />
        </div>
        
        <div class="form-group">
          <label for="password">Contraseña</label>
          <input 
            type="password" 
            id="password" 
            v-model="password" 
            required 
            placeholder="••••••"
            autocomplete="current-password"
          />
        </div>
        
        <button type="submit" class="btn btn-primary btn-full" :disabled="auth.loading">
          {{ auth.loading ? 'Ingresando...' : 'Iniciar Sesión' }}
        </button>
      </form>
      
      <p class="auth-link">
        ¿No tenés cuenta? 
        <router-link to="/register">Registrate</router-link>
      </p>
      
      <p class="auth-link">
        ¿Olvidaste tu contraseña? 
        <router-link to="/forgot-password">Recuperála</router-link>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()

const email = ref('')
const password = ref('')

async function handleLogin() {
  const success = await auth.login(email.value, password.value)
  if (success) {
    router.push('/app/routines')
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  background: linear-gradient(135deg, var(--color-dark), var(--color-secondary));
}

.auth-card {
  background: #2d2d2d;
  padding: 2rem;
  border-radius: 12px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
}

.auth-card input:-webkit-autofill,
.auth-card input:-webkit-autofill:hover,
.auth-card input:-webkit-autofill:focus,
.auth-card input:-moz-autofill,
.auth-card input:-moz-autofill:hover,
.auth-card input:-moz-autofill:focus {
  -webkit-box-shadow: 0 0 0 1000px #2a2a2a inset !important;
  -moz-box-shadow: 0 0 0 1000px #2a2a2a inset !important;
  box-shadow: 0 0 0 1000px #2a2a2a inset !important;
  -webkit-text-fill-color: white !important;
  -moz-text-fill-color: white !important;
  color: white !important;
}

.auth-card h1 {
  color: var(--color-primary);
  text-align: center;
  font-size: 2.5rem;
  margin-bottom: 0.25rem;
}

.subtitle {
  text-align: center;
  color: #aaa;
  margin-bottom: 1.5rem;
}

.auth-link {
  text-align: center;
  margin-top: 1rem;
  color: #aaa;
}

.auth-link a {
  color: var(--color-primary);
  font-weight: 600;
}
</style>