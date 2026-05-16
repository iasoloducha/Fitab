<template>
  <div class="progress-view">
    <h2>{{ auth.isProfessor ? 'Progreso de Alumnos' : 'Mi Progreso' }}</h2>
    
    <!-- Selector de alumno (solo para profesor) -->
    <div v-if="auth.isProfessor" class="student-selector card">
      <label>Seleccioná un alumno:</label>
      <select v-model="selectedStudent" @change="loadStudentProgress">
        <option value="">Seleccionar alumno...</option>
        <option v-for="student in students" :key="student.id" :value="student.id">
          {{ student.name }} ({{ student.email }})
        </option>
      </select>
    </div>
    
    <div v-if="loading" class="loading">Cargando...</div>
    
    <div v-else-if="!auth.isProfessor && progress.length === 0" class="empty-state">
      <p>No tenés registros todavía.</p>
      <p>Empezá a marcar ejercicios completados en tus rutinas para ver tu progreso.</p>
    </div>
    
    <div v-else-if="auth.isProfessor && !selectedStudent" class="empty-state">
      <p>Seleccioná un alumno para ver su progreso.</p>
    </div>
    
    <div v-else-if="progress.length === 0" class="empty-state">
      <p>Este alumno no tiene registros todavía.</p>
    </div>
    
    <div v-else>
      <!-- Selector de ejercicio -->
      <div class="exercise-selector card">
        <label>Seleccioná un ejercicio:</label>
        <select v-model="selectedExercise">
          <option value="">Todos los ejercicios</option>
          <option v-for="ex in uniqueExercises" :key="ex" :value="ex">
            {{ ex }}
          </option>
        </select>
      </div>
      
      <!-- Gráfico -->
      <div v-if="chartData.labels.length > 0" class="chart-container card">
        <Line :data="chartData" :options="chartOptions" />
      </div>
      
      <div v-else class="no-data card">
        <p>No hay datos de progreso para este ejercicio todavía.</p>
        <p v-if="selectedExercise">Intentá con "Todos los ejercicios" para ver todos los registros.</p>
      </div>
      
      <!-- Lista de registros -->
      <div class="progress-list">
        <h3>Registros</h3>
        <div v-for="item in filteredProgress" :key="item.id" class="progress-item">
          <div class="progress-exercise">
            <h4>{{ item.exercise_name }}</h4>
            <span class="progress-date">{{ formatDate(item.date) }}</span>
          </div>
          <div class="progress-status">
            <span v-if="item.completed" class="badge badge-done">✓</span>
            <span class="weight">
              <template v-if="item.actual_sets || item.actual_reps">
                {{ item.actual_sets || '?' }} series x {{ item.actual_reps || '?' }} reps
              </template>
              <template v-if="item.actual_weight && (item.actual_sets || item.actual_reps)">
                — {{ item.actual_weight }}kg
              </template>
              <template v-else-if="item.actual_weight && !item.actual_sets && !item.actual_reps">
                {{ item.actual_weight }}kg
              </template>
              <template v-if="!item.actual_weight && !item.actual_sets && !item.actual_reps">
                Sin registro
              </template>
            </span>
          </div>
          <div v-if="item.notes" class="progress-notes">
            📝 {{ item.notes }}
          </div>
          <button @click="deleteLog(item.id)" class="btn-delete-log" title="Eliminar registro">
            ×
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { useRoutineStore } from '../stores/routines'
import { useAuthStore } from '../stores/auth'
import { api } from '../api'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

const auth = useAuthStore()
const routines = useRoutineStore()
const progress = ref([])
const loading = ref(true)
const selectedExercise = ref('')
const selectedStudent = ref('')
const students = ref([])

onMounted(async () => {
  if (auth.isProfessor) {
    await routines.fetchStudents()
    students.value = routines.students
  } else {
    progress.value = await routines.getProgress()
    loading.value = false
  }
})

async function loadStudentProgress() {
  if (!selectedStudent.value) {
    progress.value = []
    return
  }
  
  loading.value = true
  progress.value = await routines.getProgress(selectedStudent.value)
  loading.value = false
}

async function deleteLog(logId) {
  if (!confirm('¿Eliminar este registro? Esta acción no se puede deshacer.')) {
    return
  }
  
  try {
    await api.logs.delete(logId)
    // Recargar progreso
    if (auth.isProfessor && selectedStudent.value) {
      progress.value = await routines.getProgress(selectedStudent.value)
    } else {
      progress.value = await routines.getProgress()
    }
  } catch (err) {
    alert('Error: ' + err.message)
  }
}

const uniqueExercises = computed(() => {
  const names = [...new Set(progress.value.map(p => p.exercise_name))]
  return names.sort()
})

const filteredProgress = computed(() => {
  if (!selectedExercise.value) return progress.value
  return progress.value.filter(p => p.exercise_name === selectedExercise.value)
})

const chartData = computed(() => {
  const items = selectedExercise.value 
    ? progress.value.filter(p => p.exercise_name === selectedExercise.value)
    : progress.value.filter(p => p.actual_weight)
  
  // Items con ALGÚN dato (peso, series o reps)
  const withData = items.filter(p => 
    (p.actual_weight && p.actual_weight.trim() !== '') || 
    p.actual_sets || 
    p.actual_reps
  )
  
  if (withData.length === 0) {
    return { labels: [], datasets: [] }
  }
  
  // Ordenar por fecha
  withData.sort((a, b) => new Date(a.date) - new Date(b.date))
  
  // Extraer pesos numéricos (solo la primera cifra si es progresivo)
  const weights = withData.map(p => {
    if (!p.actual_weight) return null
    const first = p.actual_weight.split(/[\s,]+/)[0]
    return parseFloat(first) || 0
  })
  
  // Extraer sets
  const sets = withData.map(p => p.actual_sets || null)
  
  // Extraer reps numéricos (solo la primera cifra si es progresivo)
  const reps = withData.map(p => {
    if (!p.actual_reps) return null
    const first = p.actual_reps.split(/[\s,]+/)[0]
    return parseFloat(first) || null
  })
  
  return {
    labels: withData.map(p => formatShortDate(p.date)),
    datasets: [
      {
        label: selectedExercise.value ? 'Peso (kg)' : 'Peso (kg)',
        data: weights,
        borderColor: '#e63946',
        backgroundColor: 'rgba(230, 57, 70, 0.1)',
        tension: 0.3,
        fill: false,
        pointRadius: 6,
        pointBackgroundColor: '#e63946',
        spanGaps: true,
      },
      {
        label: 'Series',
        data: sets,
        borderColor: '#457bff',
        backgroundColor: 'rgba(69, 123, 255, 0.1)',
        tension: 0.3,
        fill: false,
        pointRadius: 6,
        pointBackgroundColor: '#457bff',
        spanGaps: true,
      },
      {
        label: 'Reps',
        data: reps,
        borderColor: '#28a745',
        backgroundColor: 'rgba(40, 167, 69, 0.1)',
        tension: 0.3,
        fill: false,
        pointRadius: 6,
        pointBackgroundColor: '#28a745',
        spanGaps: true,
      },
    ]
  }
})

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: true,
      labels: {
        color: '#ccc',
        boxWidth: 16,
        padding: 16,
      }
    },
    tooltip: {
      callbacks: {
        label: (context) => {
          const label = context.dataset.label || ''
          const val = context.parsed.y
          if (label === 'Peso (kg)') return `${val} kg`
          if (label === 'Series') return `${val} series`
          if (label === 'Reps') return `${val} reps`
          return `${label}: ${val}`
        }
      }
    }
  },
  scales: {
    y: {
      beginAtZero: true,
      title: {
        display: true,
        text: 'Valor'
      }
    },
    x: {
      title: {
        display: true,
        text: 'Fecha'
      }
    }
  }
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString('es-AR', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  })
}

function formatShortDate(dateStr) {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString('es-AR', {
    day: 'numeric',
    month: 'short'
  })
}
</script>

<style scoped>
.progress-view {
  max-width: 800px;
}

.progress-view h2 {
  color: var(--color-primary);
}

.student-selector {
  margin-bottom: 1rem;
}

.student-selector label,
.exercise-selector label {
  color: #aaa;
}

.progress-list h3 {
  color: var(--color-primary);
}

.student-selector select {
  width: 100%;
  padding: 0.75rem;
  border: 2px solid #444;
  border-radius: 4px;
  font-size: 1rem;
  margin-top: 0.5rem;
  background: #2a2a2a;
  color: white;
}

.exercise-selector {
  margin-bottom: 1rem;
}

.exercise-selector select {
  width: 100%;
  padding: 0.75rem;
  border: 2px solid #444;
  border-radius: 4px;
  font-size: 1rem;
  margin-top: 0.5rem;
  background: #2a2a2a;
  color: white;
}

.chart-container {
  height: 300px;
  margin-bottom: 1rem;
  position: relative;
}

.no-data {
  text-align: center;
  color: #888;
  padding: 2rem;
  margin-bottom: 1rem;
}

.progress-list {
  margin-top: 2rem;
}

.progress-list h3 {
  margin-bottom: 1rem;
  color: var(--color-primary);
}

.progress-list .progress-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #363636;
  padding: 1rem 1.5rem;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  margin-bottom: 0.5rem;
}

.progress-exercise h4 {
  margin: 0;
  color: white;
}

.progress-date {
  color: #888;
  font-size: 0.875rem;
}

.progress-status {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.badge-done {
  width: 2rem;
  height: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #28a745;
  color: white;
  border-radius: 50%;
  font-weight: bold;
}

.weight {
  font-weight: 600;
  color: var(--color-primary);
}

.weight-none {
  color: #666;
  font-weight: 400;
}

.progress-notes {
  background: #2a3a2a;
  padding:0.5rem 0.75rem;
  border-radius: 4px;
  font-size:0.875rem;
  color: #aaa;
  margin-top:0.5rem;
  border-left: 3px solid var(--color-primary);
}

.btn-delete-log {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  width: 1.5rem;
  height: 1.5rem;
  border: none;
  background: #dc3545;
  color: white;
  border-radius: 50%;
  font-weight: bold;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.875rem;
  opacity: 0;
  transition: opacity 0.2s;
}

.progress-item {
  position: relative;
}

.progress-item:hover .btn-delete-log {
  opacity:1;
}

.empty-state,
.loading {
  text-align: center;
  padding: 3rem;
  color: #888;
}

@media (max-width: 768px) {
  .chart-container {
    height: 250px;
  }
  
  .progress-list .progress-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
  }
}
</style>