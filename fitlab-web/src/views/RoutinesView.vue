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
            @click="openCopyModal(routine)" 
            class="btn btn-secondary"
          >
            Copiar
          </button>
          <button 
            v-if="auth.isProfessor" 
            @click="openDeleteModal(routine)" 
            class="btn btn-danger"
          >
            Eliminar
          </button>
        </div>
      </div>
    </div>

    <!-- Copy Modal -->
    <div v-if="showCopyModal" class="modal-overlay" @click.self="closeCopyModal">
      <div class="modal">
        <h3>Copiar Rutina</h3>
        <p class="modal-subtitle">Copia "{{ copyRoutine?.title }}" a otro alumno</p>
        
        <form @submit.prevent="handleCopy">
          <div class="form-group">
            <label for="target_user">Alumno destino</label>
            <select id="target_user" v-model="copyForm.target_user_id" required>
              <option value="">Seleccioná un alumno...</option>
              <option v-for="student in routines.students" :key="student.id" :value="student.id">
                {{ student.name }} ({{ student.email }})
              </option>
            </select>
          </div>
          
          <div class="form-group">
            <label for="title">Título (opcional)</label>
            <input 
              type="text" 
              id="title" 
              v-model="copyForm.title" 
              placeholder="Dejar vacío para usar título original + (copia)"
            />
          </div>
          
          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" @click="closeCopyModal">
              Cancelar
            </button>
            <button type="submit" class="btn btn-primary" :disabled="copying">
              {{ copying ? 'Copiando...' : 'Copiar Rutina' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="closeDeleteModal">
      <div class="modal">
        <h3>Eliminar Rutina</h3>
        <p class="modal-subtitle">¿Estás seguro de eliminar "{{ deleteRoutineData?.title }}"?</p>
        <p class="modal-warning">Esta acción no se puede deshacer.</p>
        
        <div class="modal-actions">
          <button type="button" class="btn btn-secondary" @click="closeDeleteModal">
            Cancelar
          </button>
          <button type="button" class="btn btn-danger" :disabled="deleting" @click="confirmDelete">
            {{ deleting ? 'Eliminando...' : 'Eliminar' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRoutineStore } from '../stores/routines'

const auth = useAuthStore()
const routines = useRoutineStore()

const showCopyModal = ref(false)
const copyRoutine = ref(null)
const copying = ref(false)

const showDeleteModal = ref(false)
const deleteRoutineData = ref(null)
const deleting = ref(false)

const copyForm = reactive({
  target_user_id: '',
  title: '',
})

onMounted(() => {
  routines.fetchRoutines()
  if (auth.isProfessor) {
    routines.fetchStudents()
  }
})

function formatDate(dateStr) {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString('es-AR')
}

function openDeleteModal(routine) {
  deleteRoutineData.value = routine
  showDeleteModal.value = true
}

function closeDeleteModal() {
  showDeleteModal.value = false
  deleteRoutineData.value = null
}

async function confirmDelete() {
  deleting.value = true
  await routines.deleteRoutine(deleteRoutineData.value.id)
  deleting.value = false
  closeDeleteModal()
}

async function toggleActive(id, isActive) {
  await routines.toggleActive(id, isActive)
}

function openCopyModal(routine) {
  copyRoutine.value = routine
  copyForm.target_user_id = ''
  copyForm.title = ''
  showCopyModal.value = true
}

function closeCopyModal() {
  showCopyModal.value = false
  copyRoutine.value = null
}

async function handleCopy() {
  copying.value = true
  const result = await routines.copyRoutine(copyRoutine.value.id, {
    target_user_id: copyForm.target_user_id,
    title: copyForm.title || undefined,
  })
  copying.value = false
  
  if (result) {
    closeCopyModal()
  }
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

/* Modal */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: #363636;
  padding: 1.5rem;
  border-radius: 8px;
  max-width: 400px;
  width: 90%;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
}

.modal h3 {
  color: var(--color-primary);
  margin-bottom: 0.25rem;
}

.modal-subtitle {
  color: #888;
  font-size: 0.875rem;
  margin-bottom: 0.5rem;
}

.modal-warning {
  color: #ff6b6b;
  font-size: 0.875rem;
  margin-bottom: 1rem;
}

.modal .form-group {
  margin-bottom: 1rem;
}

.modal .form-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: #ccc;
}

.modal .form-group select,
.modal .form-group input {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #555;
  border-radius: 4px;
  background: #2a2a2a;
  color: #fff;
}

.modal-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
  margin-top: 1.5rem;
}
</style>