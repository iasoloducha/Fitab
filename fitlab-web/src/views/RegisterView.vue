<template>
  <div class="auth-page">
    <div class="auth-card">
      <h1>FITLAB</h1>
      <p class="subtitle">Creá tu cuenta gratis</p>
      
      <div v-if="auth.error" class="error">{{ auth.error }}</div>
      
      <form @submit.prevent="handleRegister">
        <div class="form-group">
          <label for="name">Nombre completo</label>
          <input 
            type="text" 
            id="name" 
            v-model="form.name" 
            required 
            placeholder="Tu nombre"
            autocomplete="name"
          />
        </div>
        
        <div class="form-group">
          <label for="email">Email</label>
          <input 
            type="email" 
            id="email" 
            v-model="form.email" 
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
            v-model="form.password" 
            required 
            minlength="6"
            placeholder="Mínimo 6 caracteres"
            autocomplete="new-password"
          />
        </div>
        
        <div class="form-group">
          <label for="role">Soy</label>
          <select id="role" v-model="form.role" required autocomplete="off">
            <option value="student">Alumno</option>
            <option value="professor">Profesor</option>
          </select>
        </div>
        
        <div v-if="form.role === 'professor'" class="form-group">
          <label for="code">Código de profesor</label>
          <input 
            type="text" 
            id="code" 
            v-model="form.professor_code" 
            placeholder="Código que te dieron"
          />
        </div>
        
        <button type="submit" class="btn btn-primary btn-full" :disabled="auth.loading">
          {{ auth.loading ? 'Creando cuenta...' : 'Crear Cuenta' }}
        </button>
      </form>
      
      <p class="auth-link">
        ¿Ya tenés cuenta? 
        <router-link to="/login">Ingresá</router-link>
      </p>
    </div>
  </div>
</template>

<script setup>
import { reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()

const form = reactive({
  name: '',
  email: '',
  password: '',
  role: 'student',
  professor_code: '',
})

async function handleRegister() {
  const success = await auth.register({ ...form })
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