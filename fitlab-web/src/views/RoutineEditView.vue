<template>
  <div class="routine-edit">
    <router-link to="/app/routines" class="back-link">← Cancelar</router-link>
    
    <h2>Nueva Rutina</h2>
    
    <form @submit.prevent="handleSubmit" class="card">
      <div class="form-group">
        <label for="student">Alumno</label>
        <select id="student" v-model="form.user_id" required>
          <option value="">Seleccioná un alumno...</option>
          <option v-for="student in routines.students" :key="student.id" :value="student.id">
            {{ student.name }} ({{ student.email }})
          </option>
        </select>
      </div>
      
      <div class="form-group">
        <label for="title">Título</label>
        <input 
          type="text" 
          id="title" 
          v-model="form.title" 
          required 
          placeholder="Ej: Rutina de Fuerza - Nivel Intermedio"
        />
      </div>
      
      <div class="form-row">
        <div class="form-group">
          <label for="start_date">Fecha de inicio</label>
          <input type="date" id="start_date" v-model="form.start_date" />
        </div>
        <div class="form-group">
          <label for="end_date">Fecha de fin</label>
          <input type="date" id="end_date" v-model="form.end_date" />
        </div>
      </div>
      
      <button type="submit" class="btn btn-primary" :disabled="saving">
        {{ saving ? 'Guardando...' : 'Crear Rutina' }}
      </button>
    </form>
  </div>
</template>

<script setup>
import { reactive, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useRoutineStore } from '../stores/routines'

const router = useRouter()
const routines = useRoutineStore()
const saving = ref(false)

const form = reactive({
  user_id: '',
  title: '',
  start_date: '',
  end_date: '',
})

onMounted(() => {
  routines.fetchStudents()
})

async function handleSubmit() {
  saving.value = true
  const routineId = await routines.createRoutine({ ...form })
  saving.value = false
  
  if (routineId) {
    router.push({ name: 'routine-view', params: { id: routineId } })
  }
}
</script>

<style scoped>
.routine-edit {
  max-width: 600px;
}

.routine-edit h2 {
  color: var(--color-primary);
}

.back-link {
  color: var(--color-primary);
  margin-bottom: 1rem;
  display: inline-block;
}

.card {
  background: #363636;
  padding: 1.5rem;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}
</style>