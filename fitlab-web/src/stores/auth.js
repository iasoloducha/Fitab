import { defineStore } from 'pinia'
import { api } from '../api'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null,
    loading: false,
    error: null,
  }),
  
  getters: {
    isLoggedIn: (state) => !!state.user,
    isProfessor: (state) => state.user?.role === 'professor',
    isStudent: (state) => state.user?.role === 'student',
    isAdmin: (state) => state.user?.role === 'admin',
  },
  
  actions: {
    async login(email, password) {
      this.loading = true
      this.error = null
      
      try {
        const response = await api.auth.login({ email, password })
        this.user = response.data
        return true
      } catch (err) {
        this.error = err.message
        return false
      } finally {
        this.loading = false
      }
    },
    
    async register(data) {
      this.loading = true
      this.error = null
      
      try {
        const response = await api.auth.register(data)
        this.user = response.data
        return true
      } catch (err) {
        this.error = err.message
        return false
      } finally {
        this.loading = false
      }
    },
    
    async logout() {
      try {
        await api.auth.logout()
      } catch (err) {
        console.error('Logout error:', err)
      } finally {
        this.user = null
      }
    },
    
    async fetchMe() {
      try {
        const response = await api.auth.me()
        this.user = response.data
      } catch (err) {
        this.user = null
      }
    },
    
    async adminLogin(email, password) {
      this.loading = true
      this.error = null
      
      try {
        const response = await api.admin.login({ email, password })
        this.user = response.data
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