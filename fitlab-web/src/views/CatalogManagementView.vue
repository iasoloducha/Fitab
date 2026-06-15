<template>
  <div class="catalog-view">
    <div class="catalog-header">
      <h2>Catálogo de Ejercicios</h2>
      <button @click="showCreateModal = true" class="btn btn-primary">+ Agregar ejercicio</button>
    </div>

    <div class="search-bar">
      <input
        type="text"
        v-model="searchQuery"
        @input="onSearch"
        placeholder="Buscar ejercicios..."
        class="search-input"
      />
    </div>

    <div v-if="catalog.loading" class="loading">Cargando...</div>
    <div v-else-if="catalog.error" class="error">{{ catalog.error }}</div>

    <div v-else class="catalog-table card">
      <table>
        <thead>
          <tr>
            <th>Nombre</th>
            <th>Imágenes (URLs)</th>
            <th>Acciones</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="ex in catalog.exercises" :key="ex.id">
            <td class="exercise-name">{{ ex.name }}</td>
            <td class="exercise-urls">
              <span v-if="ex.image_urls" class="urls-text">{{ ex.image_urls }}</span>
              <span v-else class="no-urls">—</span>
            </td>
            <td class="actions">
              <button @click="editExercise(ex)" class="btn-edit">✏️</button>
              <button @click="confirmDelete(ex)" class="btn-delete">×</button>
            </td>
          </tr>
          <tr v-if="catalog.exercises.length === 0">
            <td colspan="3" class="empty-row">No hay ejercicios en el catálogo</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showCreateModal || showEditModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal">
        <h3>{{ showEditModal ? 'Editar ejercicio' : 'Agregar ejercicio' }}</h3>
        <div class="form-group">
          <label>Nombre del ejercicio</label>
          <input
            type="text"
            v-model="form.name"
            placeholder="Ej: Press de banca plano"
            class="form-input"
          />
        </div>
        <div class="form-group">
          <label>URLs de imágenes (JSON array)</label>
          <input
            type="text"
            v-model="form.image_urls"
            placeholder='Ej: ["https://...", "https://..."]'
            class="form-input"
          />
        </div>
        <div v-if="formError" class="form-error">{{ formError }}</div>
        <div class="modal-actions">
          <button @click="closeModal" class="btn btn-secondary">Cancelar</button>
          <button @click="saveExercise" class="btn btn-primary" :disabled="!form.name.trim()">
            {{ showEditModal ? 'Guardar cambios' : 'Crear' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirm Modal -->
    <div v-if="deleteTarget" class="modal-overlay" @click.self="deleteTarget = null">
      <div class="modal confirm-modal">
        <h3>Eliminar ejercicio</h3>
        <p>¿Estás seguro de eliminar <strong>{{ deleteTarget.name }}</strong> del catálogo?</p>
        <div v-if="deleteError" class="form-error">{{ deleteError }}</div>
        <div class="modal-actions">
          <button @click="deleteTarget = null" class="btn btn-secondary">Cancelar</button>
          <button @click="deleteExercise" class="btn btn-danger">Eliminar</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { useCatalogStore } from '../stores/catalog'

const catalog = useCatalogStore()

const searchQuery = ref('')
const showCreateModal = ref(false)
const showEditModal = ref(false)
const editingId = ref(null)
const deleteTarget = ref(null)
const formError = ref('')
const deleteError = ref('')

const form = reactive({
  name: '',
  image_urls: '',
})

onMounted(() => {
  catalog.fetchAll()
})

let searchTimer = null
function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    catalog.fetchAll(searchQuery.value || '')
  }, 300)
}

function resetForm() {
  form.name = ''
  form.image_urls = ''
  formError.value = ''
  editingId.value = null
}

function editExercise(ex) {
  editingId.value = ex.id
  form.name = ex.name
  form.image_urls = ex.image_urls || ''
  formError.value = ''
  showEditModal.value = true
}

function closeModal() {
  showCreateModal.value = false
  showEditModal.value = false
  resetForm()
}

async function saveExercise() {
  if (!form.name.trim()) return
  formError.value = ''

  try {
    if (showEditModal.value && editingId.value) {
      await catalog.update(editingId.value, {
        name: form.name.trim(),
        image_urls: form.image_urls || null,
      })
    } else {
      await catalog.create({
        name: form.name.trim(),
        image_urls: form.image_urls || null,
      })
    }
    closeModal()
  } catch (err) {
    formError.value = err.message
  }
}

function confirmDelete(ex) {
  deleteTarget.value = ex
  deleteError.value = ''
}

async function deleteExercise() {
  if (!deleteTarget.value) return
  deleteError.value = ''

  try {
    await catalog.remove(deleteTarget.value.id)
    deleteTarget.value = null
  } catch (err) {
    deleteError.value = err.message
  }
}
</script>

<style scoped>
.catalog-view {
  max-width: 900px;
}

.catalog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.catalog-header h2 {
  color: var(--color-primary);
}

.search-bar {
  margin-bottom: 1rem;
}

.search-input {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #444;
  border-radius: 4px;
  background: #2a2a2a;
  color: white;
  box-sizing: border-box;
}

.search-input:focus {
  outline: none;
  border-color: var(--color-primary);
}

.catalog-table {
  background: #363636;
  border-radius: 8px;
  padding: 0;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

table {
  width: 100%;
  border-collapse: collapse;
}

th {
  text-align: left;
  padding: 0.75rem 1rem;
  background: #2a2a2a;
  color: #888;
  font-weight: 600;
  font-size: 0.8rem;
  text-transform: uppercase;
}

td {
  padding: 0.75rem 1rem;
  border-top: 1px solid #444;
  color: #ccc;
}

.exercise-name {
  font-weight: 600;
  color: white;
}

.urls-text {
  font-size: 0.8rem;
  color: #888;
  word-break: break-all;
}

.no-urls {
  color: #555;
}

.actions {
  display: flex;
  gap: 0.5rem;
  white-space: nowrap;
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

.empty-row {
  text-align: center;
  color: #666;
  font-style: italic;
  padding: 2rem;
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

.confirm-modal p {
  color: #ccc;
  margin-bottom: 1rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  font-size: 0.75rem;
  font-weight: 600;
  color: #888;
  margin-bottom: 0.25rem;
}

.form-input {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #444;
  border-radius: 4px;
  background: #2a2a2a;
  color: white;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: var(--color-primary);
}

.form-error {
  color: #dc3545;
  font-size: 0.875rem;
  margin-bottom: 0.5rem;
}

.modal-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
  margin-top: 1rem;
}

.loading {
  text-align: center;
  padding: 2rem;
  color: #888;
}

.error {
  color: #dc3545;
  padding: 1rem;
  background: #362222;
  border-radius: 8px;
  margin-bottom: 1rem;
}
</style>
