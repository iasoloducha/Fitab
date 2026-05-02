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
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()

// Fetch user data if not already loaded (e.g., on direct navigation)
onMounted(async () => {
  if (!auth.user?.created_at) {
    await auth.fetchMe()
  }
})

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
</style>