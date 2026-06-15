import { defineStore } from 'pinia'
import { api } from '../api'

export const useCatalogStore = defineStore('catalog', {
  state: () => ({
    exercises: [],
    loading: false,
    error: null,
  }),

  actions: {
    async fetchAll(q) {
      this.loading = true
      this.error = null

      try {
        const response = await api.catalog.list(q)
        this.exercises = response.data || []
      } catch (err) {
        this.error = err.message
      } finally {
        this.loading = false
      }
    },

    async create(data) {
      this.error = null
      try {
        const response = await api.catalog.create(data)
        await this.fetchAll()
        return response.data
      } catch (err) {
        this.error = err.message
        throw err
      }
    },

    async update(id, data) {
      this.error = null
      try {
        await api.catalog.update(id, data)
        await this.fetchAll()
      } catch (err) {
        this.error = err.message
        throw err
      }
    },

    async remove(id) {
      this.error = null
      try {
        await api.catalog.delete(id)
        await this.fetchAll()
      } catch (err) {
        this.error = err.message
        throw err
      }
    },
  },
})
