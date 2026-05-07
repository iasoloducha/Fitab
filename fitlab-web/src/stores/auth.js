import { defineStore } from 'pinia'
import { api } from '../api'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: JSON.parse(sessionStorage.getItem('user') || 'null'),
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
    _persistUser() {
      sessionStorage.setItem('user', JSON.stringify(this.user))
    },
    async login(email, password) {
      this.loading = true
      this.error = null
      
      try {
        const response = await api.auth.login({ email, password })
        this.user = response.data
        this._persistUser()
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
        this._persistUser()
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
        sessionStorage.removeItem('user')
      }
    },
    
    async fetchMe() {
      try {
        const response = await api.auth.me()
        this.user = response.data
        this._persistUser()
      } catch (err) {
        // Only clear user on 401 (not authenticated), keep existing user on network errors
        if (err.status === 401 || err.message?.includes('401')) {
          this.user = null
          sessionStorage.removeItem('user')
        }
        // Other errors: don't clear user, keep session alive
      }
    },
    
    async adminLogin(email, password) {
      this.loading = true
      this.error = null
      
      try {
        const response = await api.admin.login({ email, password })
        this.user = response.data
        this._persistUser()
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