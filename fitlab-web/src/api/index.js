const API_BASE = import.meta.env.VITE_API_URL || '/api'

class ApiError extends Error {
  constructor(message, code, status) {
    super(message)
    this.code = code
    this.status = status
  }
}

async function request(path, options = {}) {
  const url = `${API_BASE}${path}`
  
  const config = {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    credentials: 'include',
  }

  if (options.body && typeof options.body === 'object') {
    config.body = JSON.stringify(options.body)
  }

  const response = await fetch(url, config)
  
  let data
  try {
    data = await response.json()
  } catch {
    data = {}
  }

  if (!response.ok) {
    throw new ApiError(
      data.error || 'Error desconocido',
      data.code || 'UNKNOWN',
      response.status
    )
  }

  return data
}

export const api = {
  auth: {
    register: (data) => request('/auth/register', { method: 'POST', body: data }),
    login: (data) => request('/auth/login', { method: 'POST', body: data }),
    logout: () => request('/auth/logout', { method: 'POST' }),
    me: () => request('/auth/me'),
    forgotPassword: (email) => request('/auth/forgot-password', { method: 'POST', body: { email } }),
    changePassword: (currentPassword, newPassword) => request('/auth/password', { method: 'PUT', body: { current_password: currentPassword, new_password: newPassword } }),
  },
  
  routines: {
    list: () => request('/routines'),
    get: (id) => request(`/routines/${id}`),
    create: (data) => request('/routines', { method: 'POST', body: data }),
    update: (id, data) => request(`/routines/${id}`, { method: 'PUT', body: data }),
    delete: (id) => request(`/routines/${id}`, { method: 'DELETE' }),
    addExercise: (routineId, data) => request(`/routines/${routineId}/exercises`, { method: 'POST', body: data }),
    updateExercise: (exerciseId, data) => request(`/exercises/${exerciseId}`, { method: 'PUT', body: data }),
    deleteExercise: (exerciseId) => request(`/exercises/${exerciseId}`, { method: 'DELETE' }),
  },
  
  students: {
    list: () => request('/users/students'),
  },
  
  logs: {
    create: (exerciseId, data) => request(`/exercises/${exerciseId}/logs`, { method: 'POST', body: data }),
    list: (exerciseId) => request(`/exercises/${exerciseId}/logs`),
    delete: (logId) => request(`/logs/${logId}`, { method: 'DELETE' }),
  },
  
  progress: {
    get: (userId) => request(`/progress${userId ? `?user_id=${userId}` : ''}`),
  },
  
  admin: {
    login: (data) => request('/admin/login', { method: 'POST', body: data }),
    getUsers: () => request('/admin/users'),
    updateUser: (id, data) => request(`/admin/users/${id}`, { method: 'PUT', body: data }),
    deleteUser: (id) => request(`/admin/users/${id}`, { method: 'DELETE' }),
  },
}

export { ApiError }