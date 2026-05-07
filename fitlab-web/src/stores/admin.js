import { defineStore } from 'pinia'
import { api } from '../api'

export const useAdminStore = defineStore('admin', {
  state: () => ({
    users: [],
    loading: false,
    error: null,
    success: null,
  }),

  actions: {
    async fetchUsers() {
      this.loading = true
      this.error = null
      this.success = null
      
      try {
        const response = await api.admin.getUsers()
        this.users = response.data || []
      } catch (err) {
        this.error = err.message
      } finally {
        this.loading = false
      }
    },

    async updateUser(userId, newName) {
      this.loading = true
      this.error = null
      this.success = null
      
      try {
        await api.admin.updateUser(userId, { name: newName })
        this.success = 'Nombre actualizado correctamente'
        await this.fetchUsers()
        return true
      } catch (err) {
        this.error = err.message
        return false
      } finally {
        this.loading = false
      }
    },

    async deleteUser(userId) {
      this.loading = true
      this.error = null
      this.success = null
      
      try {
        await api.admin.deleteUser(userId)
        this.success = 'Usuario eliminado correctamente'
        await this.fetchUsers()
        return true
      } catch (err) {
        this.error = err.message
        return false
      } finally {
        this.loading = false
      }
    },

    clearMessages() {
      this.error = null
      this.success = null
    },

    async restoreDatabase(file) {
      this.loading = true
      this.error = null
      this.success = null

      try {
        const response = await api.admin.restore(file)
        this.success = 'Base de datos restaurada correctamente'
        return true
      } catch (err) {
        this.error = err.message
        return false
      } finally {
        this.loading = false
      }
    },
  },
})