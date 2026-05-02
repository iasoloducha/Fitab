<template>
  <div class="routines-view">
    <div class="header">
      <h2>Mis Rutinas</h2>
      <router-link v-if="auth.isProfessor" to="/app/routines/new" class="btn btn-primary">
        + Nueva Rutina
      </router-link>
    </div>
    
    <div v-if="routines.loading" class="loading">Cargando...</div>
    
    <div v-else-if="routines.error" class="error">{{ routines.error }}</div>
    
    <div v-else-if="routines.routines.length === 0" class="empty-state">
      <p>No tenés rutinas asignadas todavía.</p>
      <p v-if="auth.isProfessor">¿Qué esperás? <router-link to="/app/routines/new">Creá la primera</router-link></p>
    </div>
    
    <div v-else class="routines-grid">
      <div 
        v-for="routine in routines.routines" 
        :key="routine.id" 
        class="routine-card"
      >
        <div class="routine-header">
          <h3>{{ routine.title }}</h3>
          <span 
            v-if="auth.isProfessor"
            class="badge" 
            :class="routine.is_active ? 'badge-active' : 'badge-inactive'"
            @click="toggleActive(routine.id, !routine.is_active)"
            :title="routine.is_active ? 'Click para desactivar' : 'Click para activar'"
          >
            {{ routine.is_active ? 'Activa' : 'Inactiva' }}
          </span>
          <span v-else-if="routine.is_active" class="badge badge-active">Activa</span>
        </div>
        <p v-if="routine.student_name" class="student-name">
          Alumno: {{ routine.student_name }}
        </p>
        <p v-if="routine.start_date" class="routine-date">
          {{ formatDate(routine.start_date) }}
          <span v-if="routine.end_date"> - {{ formatDate(routine.end_date) }}</span>
        </p>
        <div class="routine-actions">
          <router-link :to="`/app/routines/${routine.id}`" class="btn btn-secondary">
            Ver
          </router-link>
          <button 
            v-if="auth.isProfessor" 
            @click="deleteRoutine(routine.id)" 
            class="btn btn-danger"
          >
            Eliminar
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRoutineStore } from '../stores/routines'

const auth = useAuthStore()
const routines = useRoutineStore()

onMounted(() => {
  routines.fetchRoutines()
})

function formatDate(dateStr) {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString('es-AR')
}

async function deleteRoutine(id) {
  if (confirm('¿Seguro que querés eliminar esta rutina?')) {
    await routines.deleteRoutine(id)
  }
}

async function toggleActive(id, isActive) {
  await routines.toggleActive(id, isActive)
}
</script>

<style scoped>
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.header h2 {
  color: var(--color-primary);
}

.routines-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1rem;
}

.routine-card {
  background: #363636;
  padding: 1.5rem;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.routine-header {
  display: flex;
  justify-content: space-between;
  align-items: start;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.routine-header h3 {
  margin: 0;
  color: var(--color-primary);
}

.badge {
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.badge:hover {
  opacity: 0.8;
}

.badge-active {
  background: #2d4a2d;
  color: #8aff8a;
}

.badge-inactive {
  background: #4a2d2d;
  color: #ff8a8a;
}

.student-name {
  color: #aaa;
  margin-bottom: 0.5rem;
}

.routine-date {
  color: #888;
  font-size: 0.875rem;
  margin-bottom: 1rem;
}

.routine-actions {
  display: flex;
  gap: 0.5rem;
}

.btn-danger {
  background: #dc3545;
  color: white;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
}

.btn-danger:hover {
  background: #c82333;
}

.empty-state {
  text-align: center;
  padding: 3rem;
  color: #888;
}

.loading {
  text-align: center;
  padding: 2rem;
  color: #888;
}
</style>