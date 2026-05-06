<template>
  <div class="profile-view">
    <h2>Mi Perfil</h2>
    
    <div class="profile-card">
      <div class="profile-field">
        <label>Nombre</label>
        <p>{{ auth.user?.name }}</p>
      </div>
      
      <div class="profile-field">
        <label>Email</label>
        <p>{{ auth.user?.email }}</p>
      </div>
      
      <div class="profile-field">
        <label>Rol</label>
        <p>{{ auth.user?.role === 'professor' ? 'Profesor' : 'Alumno' }}</p>
      </div>
      
      <div class="profile-field">
        <label>Miembro desde</label>
        <p>{{ formatDate(auth.user?.created_at) }}</p>
      </div>
    </div>

    <h2 class="section-title">Cambiar Contraseña</h2>
    
    <div class="profile-card">
      <div v-if="passwordMessage" class="success">{{ passwordMessage }}</div>
      <div v-if="passwordError" class="error">{{ passwordError }}</div>
      
      <form @submit.prevent="changePassword">
        <div class="form-group">
          <label for="current-password">Contraseña Actual</label>
          <input 
            type="password" 
            id="current-password" 
            v-model="currentPassword" 
            required 
            placeholder="Tu contraseña actual"
          />
        </div>
        
        <div class="form-group">
          <label for="new-password">Nueva Contraseña</label>
          <input 
            type="password" 
            id="new-password" 
            v-model="newPassword" 
            required 
            minlength="6"
            placeholder="Mínimo 6 caracteres"
          />
        </div>
        
        <button type="submit" class="btn btn-primary" :disabled="loadingPassword">
          {{ loadingPassword ? 'Cambiando...' : 'Cambiar Contraseña' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { api } from '../api'

const auth = useAuthStore()

const currentPassword = ref('')
const newPassword = ref('')
const loadingPassword = ref(false)
const passwordMessage = ref('')
const passwordError = ref('')

// Fetch user data if not already loaded (e.g., on direct navigation)
onMounted(async () => {
  if (!auth.user?.created_at) {
    await auth.fetchMe()
  }
})

async function changePassword() {
  loadingPassword.value = true
  passwordMessage.value = ''
  passwordError.value = ''
  
  try {
    await api.auth.changePassword(currentPassword.value, newPassword.value)
    passwordMessage.value = 'Contraseña cambiada correctamente'
    currentPassword.value = ''
    newPassword.value = ''
  } catch (err) {
    passwordError.value = err.message
  } finally {
    loadingPassword.value = false
  }
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString('es-AR', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}
</script>

<style scoped>
.profile-view {
  max-width: 500px;
}

.profile-view h2 {
  color: var(--color-primary);
}

.section-title {
  color: var(--color-primary);
  margin-top: 2rem;
}

.profile-card {
  background: #363636;
  padding: 1.5rem;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.profile-field {
  padding: 1rem 0;
  border-bottom: 1px solid #444;
}

.profile-field:last-child {
  border-bottom: none;
}

.profile-field label {
  display: block;
  font-size: 0.75rem;
  text-transform: uppercase;
  color: #888;
  margin-bottom: 0.25rem;
}

.profile-field p {
  margin: 0;
  font-weight: 600;
  color: white;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  font-size: 0.75rem;
  text-transform: uppercase;
  color: #888;
  margin-bottom: 0.5rem;
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

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.success {
  background: rgba(76, 175, 80, 0.1);
  border: 1px solid rgba(76, 175, 80, 0.3);
  color: #4caf50;
  padding: 0.75rem;
  border-radius: 8px;
  margin-bottom: 1rem;
}

.error {
  background: rgba(255, 68, 68, 0.1);
  border: 1px solid rgba(255, 68, 68, 0.3);
  color: #ff4444;
  padding: 0.75rem;
  border-radius: 8px;
  margin-bottom: 1rem;
}
</style>