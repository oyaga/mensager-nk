import axios from 'axios'
import { useAuthStore } from '../stores/authStore'

// Use relative URL when served from same origin (unified build)
// or use VITE_API_URL for development/separate deployment
const API_URL = import.meta.env.VITE_API_URL || ''

export const api = axios.create({
  baseURL: `${API_URL}/api/v1`,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().token
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor to handle errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      useAuthStore.getState().logout()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// Auth API
export const authApi = {
  login: async (email: string, password: string) => {
    const response = await api.post('/auth/login', { email, password })
    return response.data
  },

  register: async (name: string, email: string, password: string) => {
    const response = await api.post('/auth/register', { name, email, password })
    return response.data
  },

  getProfile: async () => {
    const response = await api.get('/profile')
    return response.data
  },

  updateProfile: async (data: any) => {
    const response = await api.put('/profile', data)
    return response.data
  },

  changePassword: async (currentPassword: string, newPassword: string) => {
    const response = await api.put('/profile/password', {
      current_password: currentPassword,
      new_password: newPassword
    })
    return response.data
  },

  // Access Token
  resetAccessToken: async () => {
    const response = await api.post('/profile/access_token')
    return response.data
  },
}

// Storage API
export const storageApi = {
  upload: async (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    const response = await api.post('/storage/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    return response.data
  }
}

// Conversations API
export const conversationsApi = {
  list: async (params?: any) => {
    const response = await api.get('/conversations', { params })
    return response.data
  },

  get: async (id: string) => {
    const response = await api.get(`/conversations/${id}`)
    return response.data
  },

  create: async (data: any) => {
    const response = await api.post('/conversations', data)
    return response.data
  },

  update: async (id: string, data: any) => {
    const response = await api.put(`/conversations/${id}`, data)
    return response.data
  },

  assign: async (id: string, userId: string) => {
    const response = await api.post(`/conversations/${id}/assign`, { user_id: userId })
    return response.data
  },

  resolve: async (id: string) => {
    const response = await api.post(`/conversations/${id}/resolve`)
    return response.data
  },
}

// Messages API
export const messagesApi = {
  list: async (conversationId: string) => {
    const response = await api.get(`/conversations/${conversationId}/messages`)
    return response.data
  },

  create: async (data: any) => {
    const response = await api.post('/messages', data)
    return response.data
  },
}

// Contacts API
export const contactsApi = {
  list: async (params?: any) => {
    const response = await api.get('/contacts', { params })
    return response.data
  },

  get: async (id: string) => {
    const response = await api.get(`/contacts/${id}`)
    return response.data
  },

  create: async (data: any) => {
    const response = await api.post('/contacts', data)
    return response.data
  },

  update: async (id: string, data: any) => {
    const response = await api.put(`/contacts/${id}`, data)
    return response.data
  },
}
