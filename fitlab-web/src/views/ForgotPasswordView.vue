<template>
  <div class="auth-page">
    <div class="auth-card">
      <h1>FITLAB</h1>
      <p class="subtitle">Recuperá tu contraseña</p>
      
      <div v-if="error" class="error">{{ error }}</div>
      <div v-if="message" class="success">{{ message }}</div>
      
      <form v-if="!message" @submit.prevent="handleForgotPassword">
        <div class="form-group">
          <label for="email">Tu email</label>
          <input 
            type="email" 
            id="email" 
            v-model="email" 
            required 
            placeholder="tu@email.com"
            autocomplete="email"
          />
        </div>
        
        <button type="submit" class="btn btn-primary btn-full" :disabled="loading">
          {{ loading ? 'Enviando...' : 'Enviarme un email' }}
        </button>
      </form>
      
      <p v-else class="auth-link">
        <router-link to="/login">Volver a iniciar sesión</router-link>
      </p>
      
      <p class="auth-link">
        ¿Recordaste tu contraseña? 
        <router-link to="/login">Iniciá sesión</router-link>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { api } from '../api'

const email = ref('')
const loading = ref(false)
const error = ref('')
const message = ref('')

async function handleForgotPassword() {
  loading.value = true
  error.value = ''
  message.value = ''
  
  try {
    await api.auth.forgotPassword(email.value)
    // Always show same message to not reveal if email exists
    message.value = 'Si el email está registrado, te hemos enviado las instrucciones'
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
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

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  color: #aaa;
  margin-bottom: 0.5rem;
  font-size: 0.9rem;
}

.form-group input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #444;
  border-radius: 8px;
  background: #2a2a2a;
  color: white;
  font-size: 1rem;
}

.form-group input:focus {
  outline: none;
  border-color: var(--color-primary);
}

.btn {
  padding: 0.75rem 1.5rem;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
}

.btn-primary {
  background: var(--color-primary);
  color: var(--color-dark);
}

.btn-primary:hover {
  background: #ff8c00;
}

.btn-full {
  width: 100%;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error {
  background: rgba(255, 68, 68, 0.1);
  border: 1px solid rgba(255, 68, 68, 0.3);
  color: #ff4444;
  padding: 0.75rem;
  border-radius: 8px;
  margin-bottom: 1rem;
}

.success {
  background: rgba(76, 175, 80, 0.1);
  border: 1px solid rgba(76, 175, 80, 0.3);
  color: #4caf50;
  padding: 0.75rem;
  border-radius: 8px;
  margin-bottom: 1rem;
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