<template>
  <div class="routine-view">
    <router-link to="/app/routines" class="back-link">← Volver a rutinas</router-link>
    
    <div v-if="routines.loading" class="loading">Cargando...</div>
    
    <div v-else-if="routines.error" class="error">{{ routines.error }}</div>
    
    <div v-else-if="routine" class="routine-detail">
      <div class="routine-header">
        <div>
          <h2>{{ routine.title }}</h2>
          <p v-if="auth.isStudent" class="routine-meta">
            Prescripta por tu profesor
          </p>
        </div>
      </div>
      
      <!-- Lista de ejercicios existentes -->
      <div v-for="day in daysWithExercises" :key="day.number" class="day-section">
        <h3>{{ day.title }}</h3>
        
        <div class="exercises-list">
          <div 
            v-for="(exercise, index) in day.exercises" 
            :key="exercise.id" 
            class="exercise-row"
          >
            <span class="exercise-number">{{ index + 1 }}</span>
            <div class="exercise-info">
              <span class="exercise-name">{{ exercise.name }}</span>
              <span class="exercise-details">
                {{ exercise.sets }} × {{ exercise.reps }}
                <span v-if="exercise.weight_kg" class="exercise-weight">
                  @ {{ exercise.weight_kg }}kg
                </span>
              </span>
              <span v-if="exercise.observations" class="exercise-obs">
                {{ exercise.observations }}
              </span>
            </div>
            
            <div class="exercise-actions">
              <button 
                v-if="auth.isStudent && !isCompletedToday(exercise.id)" 
                @click="showLogModal(exercise)" 
                class="btn-log"
              >
                Marcar
              </button>
              <span 
                v-if="auth.isStudent && isCompletedToday(exercise.id)" 
                class="completed-badge"
              >
                ✓ Listo
              </span>
              <button 
                v-if="auth.isProfessor" 
                @click="editExercise(exercise)" 
                class="btn-edit"
              >
                ✏️
              </button>
              <button 
                v-if="auth.isProfessor" 
                @click="deleteExercise(exercise.id)" 
                class="btn-delete"
              >
                ×
              </button>
            </div>
          </div>
          
          <div v-if="day.exercises.length === 0" class="no-exercises">
            Sin ejercicios en este día
          </div>
        </div>
      </div>
      
      <!-- Estado vacío -->
      <div v-if="daysWithExercises.length === 0" class="empty-state">
        <p>Esta rutina aún no tiene ejercicios.</p>
      </div>
      
      <!-- Formulario para agregar ejercicio (profesor) -->
      <div v-if="auth.isProfessor" class="add-exercise-form card">
        <h4>Agregar ejercicio</h4>
        <div class="form-row">
          <div class="form-group">
            <label>Día</label>
            <input 
              type="number" 
              v-model.number="newExercise.day_number" 
              min="1"
            />
          </div>
          <div class="form-group" style="flex: 2">
            <label>Nombre del ejercicio</label>
            <input 
              type="text" 
              v-model="newExercise.name" 
              placeholder="Ej: Press de banca"
            />
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>Series</label>
            <input 
              type="number" 
              v-model.number="newExercise.sets" 
              min="1"
            />
          </div>
          <div class="form-group">
            <label>Reps</label>
            <input 
              type="text" 
              v-model="newExercise.reps" 
              placeholder="12"
            />
          </div>
          <div class="form-group">
            <label>Peso (kg)</label>
            <input 
              type="text" 
              v-model="newExercise.weight_kg" 
              placeholder="20 22 26 o opcional"
            />
          </div>
        </div>
        <div class="form-group">
          <label>Observaciones</label>
          <input 
            type="text" 
            v-model="newExercise.observations" 
            placeholder="Opcional"
          />
        </div>
        <button @click="addExercise" class="btn btn-primary" :disabled="!canAddExercise">
          Agregar ejercicio
        </button>
      </div>
      
      <!-- Modal Log Exercise (Student) -->
      <div v-if="showLogModalVisible" class="modal-overlay" @click.self="cancelLog">
        <div class="modal">
          <h3>Registrar ejercicio: {{ loggingExercise?.name }}</h3>
          <p class="modal-subtitle">Peso sugerido: {{ loggingExercise?.weight_kg || 'no especificado' }}</p>
          <div class="form-row">
            <div class="form-group">
              <label>Series reales</label>
              <input 
                type="number" 
                v-model.number="logForm.actual_sets" 
                min="1"
                placeholder="Ej: 3"
              />
            </div>
            <div class="form-group">
              <label>Reps reales</label>
              <input 
                type="text" 
                v-model="logForm.actual_reps" 
                placeholder="Ej: 12"
              />
            </div>
            <div class="form-group">
              <label>Peso real (kg)</label>
              <input 
                type="text" 
                v-model="logForm.actual_weight" 
                placeholder="Ej: 20, 25, 30-35"
                autofocus
              />
            </div>
          </div>
          <div class="form-group">
            <label>Notas (opcional)</label>
            <input 
              type="text" 
              v-model="logForm.notes" 
              placeholder="Cómo te sentiste, RIR, etc."
            />
          </div>
          <div class="modal-actions">
            <button @click="cancelLog" class="btn btn-secondary">Cancelar</button>
            <button @click="saveLog" class="btn btn-primary">Guardar</button>
          </div>
        </div>
      </div>

      <!-- Modal Editar Ejercicio -->
      <div v-if="showEditModal" class="modal-overlay" @click.self="cancelEdit">
        <div class="modal">
          <h3>Editar ejercicio</h3>
          <div class="form-row">
            <div class="form-group">
              <label>Día</label>
              <input type="number" v-model.number="editForm.day_number" min="1" />
            </div>
            <div class="form-group" style="flex: 2">
              <label>Nombre</label>
              <input type="text" v-model="editForm.name" />
            </div>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>Series</label>
              <input type="number" v-model.number="editForm.sets" min="1" />
            </div>
            <div class="form-group">
              <label>Reps</label>
              <input type="text" v-model="editForm.reps" />
            </div>
            <div class="form-group">
              <label>Peso (kg)</label>
              <input type="text" v-model="editForm.weight_kg" placeholder="Opcional" />
            </div>
          </div>
          <div class="form-group">
            <label>Observaciones</label>
            <input type="text" v-model="editForm.observations" placeholder="Opcional" />
          </div>
          <div class="modal-actions">
            <button @click="cancelEdit" class="btn btn-secondary">Cancelar</button>
            <button @click="saveEdit" class="btn btn-primary">Guardar</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useRoutineStore } from '../stores/routines'
import { api } from '../api'

const route = useRoute()
const auth = useAuthStore()
const routines = useRoutineStore()

const completedToday = ref([])
const showEditModal = ref(false)
const showLogModalVisible = ref(false)
const loggingExercise = ref(null)
const logForm = reactive({
  actual_weight: '',
  actual_sets: null,
  actual_reps: '',
  notes: '',
})
const editingExercise = ref(null)
const editForm = reactive({
  day_number: 1,
  name: '',
  sets: 3,
  reps: '',
  weight_kg: '',
  observations: '',
})

const newExercise = reactive({
  day_number: 1,
  name: '',
  sets: 3,
  reps: '',
  weight_kg: '',
  observations: '',
})

const routine = computed(() => routines.currentRoutine)

const canAddExercise = computed(() => {
  return newExercise.name && newExercise.reps && newExercise.sets > 0
})

const daysWithExercises = computed(() => {
  if (!routine.value?.exercises) return []
  
  const days = {}
  routine.value.exercises.forEach(ex => {
    if (!days[ex.day_number]) {
      days[ex.day_number] = { number: ex.day_number, title: `Día ${ex.day_number}`, exercises: [] }
    }
    days[ex.day_number].exercises.push(ex)
  })
  
  return Object.values(days).sort((a, b) => a.number - b.number)
})

onMounted(async () => {
  await routines.fetchRoutine(route.params.id)
  
  if (auth.isStudent) {
    const today = new Date().toISOString().split('T')[0]
    // Parallelize all log fetch requests to avoid N+1 sequential calls
    await Promise.all(
      (routine.value?.exercises || []).map(async (ex) => {
        try {
          const response = await api.logs.list(ex.id)
          const logs = response.data || []
          if (logs.some(log => log.date === today && log.completed)) {
            completedToday.value.push(ex.id)
          }
        } catch {}
      })
    )
  }
})

function isCompletedToday(exerciseId) {
  return completedToday.value.includes(exerciseId)
}

function showLogModal(exercise) {
  loggingExercise.value = exercise
  logForm.actual_weight = exercise.weight_kg || ''
  logForm.actual_sets = exercise.sets || null
  logForm.actual_reps = exercise.reps || ''
  logForm.notes = ''
  showLogModalVisible.value = true
}

function cancelLog() {
  showLogModalVisible.value = false
  loggingExercise.value = null
}

async function saveLog() {
  if (!loggingExercise.value) return
  
  const today = new Date().toISOString().split('T')[0]
  await api.logs.create(loggingExercise.value.id, {
    date: today,
    completed: true,
    actual_weight: logForm.actual_weight,
    actual_sets: logForm.actual_sets || null,
    actual_reps: logForm.actual_reps || null,
    notes: logForm.notes || null,
  })
  
  completedToday.value.push(loggingExercise.value.id)
  cancelLog()
}

async function addExercise() {
  if (!canAddExercise.value) return
  
  console.log('Adding exercise:', route.params.id, newExercise)
  
  try {
    await api.routines.addExercise(route.params.id, { ...newExercise })
    console.log('Exercise added successfully')
  } catch (err) {
    console.error('Error adding exercise:', err)
    alert('Error: ' + err.message)
    return
  }
  
  // Reset form
  newExercise.name = ''
  newExercise.reps = ''
  newExercise.weight_kg = ''
  newExercise.observations = ''
  
  // Refresh routine
  await routines.fetchRoutine(route.params.id)
}

async function deleteExercise(exerciseId) {
  if (!confirm('¿Eliminar este ejercicio?')) return
  
  await api.routines.deleteExercise(exerciseId)
  await routines.fetchRoutine(route.params.id)
}

function editExercise(exercise) {
  editingExercise.value = exercise
  editForm.day_number = exercise.day_number
  editForm.name = exercise.name
  editForm.sets = exercise.sets
  editForm.reps = exercise.reps
  editForm.weight_kg = exercise.weight_kg || ''
  editForm.observations = exercise.observations || ''
  showEditModal.value = true
}

function cancelEdit() {
  showEditModal.value = false
  editingExercise.value = null
}

async function saveEdit() {
  if (!editingExercise.value) return
  
  try {
    await api.routines.updateExercise(editingExercise.value.id, {
      day_number: editForm.day_number,
      name: editForm.name,
      sets: editForm.sets,
      reps: editForm.reps,
      weight_kg: editForm.weight_kg || null,
      observations: editForm.observations || null,
    })
    await routines.fetchRoutine(route.params.id)
    cancelEdit()
  } catch (err) {
    alert('Error: ' + err.message)
  }
}
</script>

<style scoped>
.routine-view {
  max-width: 800px;
}

.routine-view h2 {
  color: var(--color-primary);
}

.back-link {
  color: var(--color-primary);
  margin-bottom: 1rem;
  display: inline-block;
}

.routine-header {
  display: flex;
  justify-content: space-between;
  align-items: start;
  margin-bottom: 2rem;
}

.routine-header h2 {
  color: var(--color-primary);
}

.routine-meta {
  color: #aaa;
  margin-top: 0.25rem;
}

.day-section {
  background: #363636;
  border-radius: 8px;
  padding: 1.5rem;
  margin-bottom: 1rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.day-section h3 {
  color: var(--color-primary);
  margin-bottom: 1rem;
  padding-bottom: 0.5rem;
  border-bottom: 2px solid #444;
}

.exercise-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 0;
  border-bottom: 1px solid #444;
}

.exercise-row:last-child {
  border-bottom: none;
}

.exercise-number {
  width: 2rem;
  height: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-primary);
  border-radius: 50%;
  font-weight: 600;
  color: white;
}

.exercise-info {
  flex: 1;
}

.exercise-name {
  display: block;
  font-weight: 600;
  color: white;
}

.exercise-details {
  color: #888;
  font-size: 0.875rem;
}

.exercise-weight {
  color: var(--color-primary);
  font-weight: 600;
}

.exercise-obs {
  display: block;
  color: #666;
  font-size: 0.75rem;
  font-style: italic;
}

.exercise-actions {
  display: flex;
  gap: 0.5rem;
}

.btn-log {
  padding: 0.5rem 1rem;
  background: #444;
  border: none;
  border-radius: 4px;
  transition: all 0.2s;
  color: var(--color-light);
}

.btn-log:hover {
  background: var(--color-primary);
  color: white;
}

.btn-log.completed {
  background: #28a745;
  color: white;
}

.completed-badge {
  display: inline-block;
  padding: 0.5rem 1rem;
  background: #2d5a2d;
  color: #8aff8a;
  border-radius: 4px;
  font-weight: 500;
}

.btn-delete {
  width: 2rem;
  height: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #dc3545;
  color: white;
  border: none;
  border-radius: 4px;
  font-weight: bold;
}

.btn-delete:hover {
  background: #c82333;
}

.no-exercises {
  color: #666;
  font-style: italic;
  padding: 1rem 0;
}

.add-exercise-form {
  margin-top: 2rem;
  padding-top: 1rem;
  border-top: 2px solid #444;
}

.add-exercise-form h4 {
  margin-bottom: 1rem;
  color: var(--color-primary);
}

.empty-state {
  text-align: center;
  padding: 2rem;
  color: #888;
  background: #363636;
  border-radius: 8px;
}

.form-row {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.form-group {
  flex: 1;
}

.form-group label {
  display: block;
  font-size: 0.75rem;
  font-weight: 600;
  color: #888;
  margin-bottom: 0.25rem;
}

.form-group input {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #444;
  border-radius: 4px;
  background: #2a2a2a;
  color: white;
}

.form-group input:focus {
  outline: none;
  border-color: var(--color-primary);
}

.btn-edit {
  width: 2rem;
  height: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-secondary);
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 0.875rem;
}

.btn-edit:hover {
  background: var(--color-dark);
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal {
  background: #363636;
  border-radius: 8px;
  padding: 1.5rem;
  width: 100%;
  max-width: 500px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
}

.modal h3 {
  margin-bottom: 1rem;
  color: var(--color-primary);
}

.modal-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
  margin-top: 1rem;
}

/* Responsive */
@media (max-width: 768px) {
  .routine-view {
    max-width: 100%;
  }
  
  .exercise-row {
    flex-wrap: wrap;
  }
  
  .exercise-info {
    flex: 1 1 calc(100% - 3rem);
  }
  
  .exercise-actions {
    flex: 1 1 100%;
    justify-content: flex-end;
    margin-top: 0.5rem;
  }
  
  .form-row {
    flex-direction: column;
  }
  
  .form-row .form-group {
    width: 100%;
  }
  
  .modal {
    margin: 0.5rem;
    padding: 1rem;
  }
  
  .modal-actions {
    flex-direction: column;
  }
  
  .modal-actions .btn {
    width: 100%;
  }
}
</style>