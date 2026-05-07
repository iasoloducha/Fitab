import { defineStore } from 'pinia'
import { api } from '../api'

export const useRoutineStore = defineStore('routines', {
  state: () => ({
    routines: [],
    currentRoutine: null,
    students: [],
    loading: false,
    error: null,
  }),
  
  actions: {
    async fetchRoutines() {
      this.loading = true
      this.error = null
      
      try {
        const response = await api.routines.list()
        this.routines = response.data || []
      } catch (err) {
        this.error = err.message
      } finally {
        this.loading = false
      }
    },
    
    async fetchRoutine(id) {
      this.loading = true
      this.error = null
      
      try {
        const response = await api.routines.get(id)
        this.currentRoutine = response.data
      } catch (err) {
        this.error = err.message
      } finally {
        this.loading = false
      }
    },
    
    async createRoutine(data) {
      this.loading = true
      this.error = null
      
      try {
        const response = await api.routines.create(data)
        await this.fetchRoutines()
        return response.data.id
      } catch (err) {
        this.error = err.message
        return null
      } finally {
        this.loading = false
      }
    },
    
    async updateRoutine(id, data) {
      try {
        await api.routines.update(id, data)
        await this.fetchRoutines()
      } catch (err) {
        this.error = err.message
      }
    },
    
    async deleteRoutine(id) {
      try {
        await api.routines.delete(id)
        await this.fetchRoutines()
      } catch (err) {
        this.error = err.message
      }
    },
    
    async copyRoutine(id, data) {
      try {
        const response = await api.routines.copy(id, data)
        await this.fetchRoutines()
        return response.data
      } catch (err) {
        this.error = err.message
        return null
      }
    },
    
    async toggleActive(id, isActive) {
      try {
        await api.routines.update(id, { is_active: isActive })
        const routine = this.routines.find(r => r.id === id)
        if (routine) {
          routine.is_active = isActive
        }
      } catch (err) {
        this.error = err.message
      }
    },
    
    async addExercise(routineId, exerciseData) {
      try {
        await api.routines.addExercise(routineId, exerciseData)
        await this.fetchRoutine(routineId)
      } catch (err) {
        this.error = err.message
      }
    },
    
    async fetchStudents() {
      try {
        const response = await api.students.list()
        this.students = response.data || []
      } catch (err) {
        this.error = err.message
      }
    },
    
    async logExercise(exerciseId, data) {
      try {
        await api.logs.create(exerciseId, data)
      } catch (err) {
        this.error = err.message
      }
    },
    
    async getProgress(userId) {
      try {
        const response = await api.progress.get(userId)
        return response.data
      } catch (err) {
        this.error = err.message
        return []
      }
    },
  },
})