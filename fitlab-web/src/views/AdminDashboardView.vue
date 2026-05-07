<template>
  <div class="admin-dashboard">
    <h1>Panel de Admin</h1>
    <p class="subtitle">Gestión de Usuarios</p>

    <div class="toolbar">
      <button class="btn btn-tool" @click="handleBackup">📥 Backup</button>
      <button class="btn btn-tool" @click="openRestoreModal">📤 Restore</button>
    </div>

    <div v-if="adminStore.error" class="error">{{ adminStore.error }}</div>
    <div v-if="adminStore.success" class="success">{{ adminStore.success }}</div>

    <div class="users-table">
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>Email</th>
            <th>Nombre</th>
            <th>Rol</th>
            <th>Fecha Creación</th>
            <th>Acciones</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in adminStore.users" :key="user.id">
            <td>{{ user.id }}</td>
            <td>{{ user.email }}</td>
            <td>{{ user.name }}</td>
            <td><span class="role-badge" :class="user.role">{{ user.role }}</span></td>
            <td>{{ formatDate(user.created_at) }}</td>
            <td class="actions">
              <button class="btn-sm btn-edit" @click="openEditModal(user)">Editar</button>
              <button class="btn-sm btn-delete" @click="confirmDelete(user)">Eliminar</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Edit Modal -->
    <div v-if="showEditModal" class="modal-overlay" @click.self="closeEditModal">
      <div class="modal-card">
        <h3>Editar Nombre de Usuario</h3>
        <div class="form-group">
          <label for="edit-name">Nombre</label>
          <input 
            type="text" 
            id="edit-name" 
            v-model="editName" 
            required 
            placeholder="Nuevo nombre"
          />
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="closeEditModal">Cancelar</button>
          <button class="btn btn-primary" @click="saveEdit" :disabled="!editName.trim()">
            Guardar
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation -->
    <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="closeDeleteConfirm">
      <div class="modal-card">
        <h3>Confirmar Eliminación</h3>
        <p>¿Estás seguro de eliminar al usuario <strong>{{ userToDelete?.name }}</strong> ({{ userToDelete?.email }})?</p>
        <p class="warning">Esta acción no se puede deshacer.</p>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="closeDeleteConfirm">Cancelar</button>
          <button class="btn btn-danger" @click="deleteUser">Eliminar</button>
        </div>
      </div>
    </div>

    <!-- Restore Modal -->
    <div v-if="showRestoreModal" class="modal-overlay" @click.self="closeRestoreModal">
      <div class="modal-card">
        <h3>Restaurar Base de Datos</h3>
        <p class="warning-message">
          ⚠️ Recomendamos hacer un backup antes de restaurar. Esto reemplazará todos los datos actuales.
        </p>
        <div class="form-group">
          <label for="restore-file">Seleccionar archivo</label>
          <input 
            type="file" 
            id="restore-file" 
            accept=".db"
            @change="handleFileSelect"
          />
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="closeRestoreModal">Cancelar</button>
          <button 
            class="btn btn-primary" 
            @click="handleRestore" 
            :disabled="!selectedFile || restoring"
          >
            {{ restoring ? 'Restaurando...' : 'Restaurar' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useAdminStore } from '../stores/admin'
import { api } from '../api'

const adminStore = useAdminStore()

const showEditModal = ref(false)
const showDeleteConfirm = ref(false)
const showRestoreModal = ref(false)
const editName = ref('')
const userToEdit = ref(null)
const userToDelete = ref(null)
const selectedFile = ref(null)
const restoring = ref(false)

onMounted(async () => {
  await adminStore.fetchUsers()
})

function openEditModal(user) {
  userToEdit.value = user
  editName.value = user.name
  showEditModal.value = true
}

function closeEditModal() {
  showEditModal.value = false
  userToEdit.value = null
  editName.value = ''
}

async function saveEdit() {
  if (!editName.value.trim()) return
  
  const success = await adminStore.updateUser(userToEdit.value.id, editName.value.trim())
  if (success) {
    closeEditModal()
  }
}

function confirmDelete(user) {
  userToDelete.value = user
  showDeleteConfirm.value = true
}

function closeDeleteConfirm() {
  showDeleteConfirm.value = false
  userToDelete.value = null
}

async function deleteUser() {
  const success = await adminStore.deleteUser(userToDelete.value.id)
  if (success) {
    closeDeleteConfirm()
  }
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString('es-AR')
}

function handleBackup() {
  api.admin.backup()
}

function openRestoreModal() {
  selectedFile.value = null
  showRestoreModal.value = true
}

function closeRestoreModal() {
  showRestoreModal.value = false
  selectedFile.value = null
}

function handleFileSelect(event) {
  const file = event.target.files[0]
  if (file) {
    selectedFile.value = file
  }
}

async function handleRestore() {
  if (!selectedFile.value) return

  restoring.value = true
  const success = await adminStore.restoreDatabase(selectedFile.value)
  restoring.value = false

  if (success) {
    closeRestoreModal()
  }
}
</script>

<style scoped>
.admin-dashboard {
  max-width: 1200px;
  margin: 0 auto;
  padding: 1rem;
}

h1 {
  color: var(--color-primary);
  margin-bottom: 0.25rem;
}

.subtitle {
  color: #aaa;
  margin-bottom: 1.5rem;
}

.toolbar {
  display: flex;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.btn-tool {
  background: #363636;
  color: #fff;
  border: 1px solid #555;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-tool:hover {
  background: #404040;
  border-color: #666;
}

.warning-message {
  background: rgba(255, 152, 0, 0.1);
  border: 1px solid rgba(255, 152, 0, 0.3);
  color: #ff9800;
  padding: 0.75rem;
  border-radius: 8px;
  margin-bottom: 1rem;
  font-size: 0.9rem;
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

.users-table {
  overflow-x: auto;
  background: #2d2d2d;
  border-radius: 12px;
  padding: 1rem;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th {
  text-align: left;
  padding: 0.75rem;
  color: #aaa;
  font-weight: 600;
  border-bottom: 1px solid #444;
}

td {
  padding: 0.75rem;
  border-bottom: 1px solid #3a3a3a;
  color: white;
}

.role-badge {
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.85rem;
  font-weight: 600;
}

.role-badge.professor {
  background: rgba(255, 152, 0, 0.2);
  color: #ff9800;
}

.role-badge.student {
  background: rgba(33, 150, 243, 0.2);
  color: #2196f3;
}

.actions {
  display: flex;
  gap: 0.5rem;
}

.btn-sm {
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.85rem;
  cursor: pointer;
  border: none;
}

.btn-edit {
  background: rgba(255, 152, 0, 0.2);
  color: #ff9800;
}

.btn-edit:hover {
  background: rgba(255, 152, 0, 0.3);
}

.btn-delete {
  background: rgba(255, 68, 68, 0.2);
  color: #ff4444;
}

.btn-delete:hover {
  background: rgba(255, 68, 68, 0.3);
}

/* Modal styles */
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

.modal-card {
  background: #2d2d2d;
  padding: 2rem;
  border-radius: 12px;
  width: 90%;
  max-width: 500px;
}

.modal-card h3 {
  color: var(--color-primary);
  margin-bottom: 1rem;
}

.modal-card .warning {
  color: #ff9800;
  font-size: 0.9rem;
  margin-top: 0.5rem;
}

.modal-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
  margin-top: 1.5rem;
}

.btn-danger {
  background: #ff4444;
  color: white;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  cursor: pointer;
}

.btn-danger:hover {
  background: #cc0000;
}
</style>
