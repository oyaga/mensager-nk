import { create } from 'zustand'
import { useAuthStore } from './authStore'

// Use relative URL when served from same origin (unified build)
const API_BASE = `${import.meta.env.VITE_API_URL || ''}/api/v1`

export interface Contact {
  id: string
  name: string
  email: string
  phone_number: string
  avatar: string | null
  created_at: string
  custom_attributes?: any
  additional_attributes?: any
}

interface ContactState {
  contacts: Contact[]
  meta?: {
    count: number
    current_page: number
  }
  isLoading: boolean
  error: string | null
  lastFetch: number | null // Timestamp da última busca
  isFetching: boolean // Flag para evitar múltiplas chamadas simultâneas
  fetchContacts: (page?: number, limit?: number, force?: boolean) => Promise<void>
  createContact: (data: Partial<Contact>) => Promise<void>
  updateContact: (id: string, data: Partial<Contact>) => Promise<void>
  deleteContact: (id: string) => Promise<void>
  clearError: () => void
}

const CACHE_TIME = 30000 // 30 segundos

export const useContactStore = create<ContactState>((set, get) => ({
  contacts: [],
  meta: {
    count: 0,
    current_page: 1,
  },
  isLoading: false,
  error: null,
  lastFetch: null,
  isFetching: false,

  fetchContacts: async (page = 1, limit = 15, force = false) => {
    const state = get()

    // Prevenir múltiplas chamadas simultâneas
    if (state.isFetching) {
      console.log('Fetch já em andamento, ignorando...')
      return
    }

    // Usar cache se dados forem recentes (a menos que force=true)
    const now = Date.now()
    if (!force && state.lastFetch && (now - state.lastFetch) < CACHE_TIME && state.contacts.length > 0) {
      console.log('Usando dados em cache')
      return
    }

    set({ isLoading: true, error: null, isFetching: true })
    try {
      console.log(`Fetching contacts page: ${page}, limit: ${limit}`)
      const token = useAuthStore.getState().token
      console.log('Token used:', token ? 'Found' : 'Missing')
      
      const url = `${API_BASE}/contacts?page=${page}&limit=${limit}`
      console.log('Fetch URL:', url)

      const response = await fetch(url, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })
      
      console.log('Response status:', response.status)
      if (!response.ok) {
        const text = await response.text()
        console.error('Fetch error body:', text)
        throw new Error(`Failed to fetch contacts: ${response.status} ${text}`)
      }
      
      const data = await response.json()
      console.log('Fetch data received:', data)

      // Handle both old format (array) and new format ({ meta, payload })
      if (Array.isArray(data)) {
         console.log('Data is array, setting contacts directly')
         set({
           contacts: data,
           isLoading: false,
           isFetching: false,
           lastFetch: Date.now()
         })
      } else {
         console.log('Data is object with payload:', data.payload?.length)
         set({
           contacts: data.payload,
           meta: data.meta,
           isLoading: false,
           isFetching: false,
           lastFetch: Date.now()
         })
      }
    } catch (error) {
      console.error('Fetch contacts error:', error)
      set({ error: (error as Error).message, isLoading: false, isFetching: false })
    }
  },

  createContact: async (contactData: Partial<Contact>) => {
    set({ isLoading: true, error: null })
    try {
      const token = useAuthStore.getState().token
      const response = await fetch(`${API_BASE}/contacts`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(contactData),
      })

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(`Failed to create contact: ${errorText}`)
      }

      const newContact = await response.json()

      // Verificar se o contato já existe na lista (evitar duplicação)
      set((state: ContactState) => {
        const exists = state.contacts.some(c => c.id === newContact.id)
        if (exists) {
          // Se já existe, apenas atualizar
          return {
            contacts: state.contacts.map(c => c.id === newContact.id ? newContact : c),
            isLoading: false
          }
        }
        // Se não existe, adicionar ao início
        return {
          contacts: [newContact, ...state.contacts],
          isLoading: false
        }
      })
    } catch (error) {
      set({ error: (error as Error).message, isLoading: false })
      throw error // Re-throw para que o componente possa tratar
    }
  },

  updateContact: async (id: string, contactData: Partial<Contact>) => {
    set({ isLoading: true, error: null })
    try {
      const token = useAuthStore.getState().token
      const response = await fetch(`${API_BASE}/contacts/${id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(contactData),
      })

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(`Failed to update contact: ${errorText}`)
      }

      const updatedContact = await response.json()
      set((state: ContactState) => ({
        contacts: state.contacts.map((c) => (c.id === id ? updatedContact : c)),
        isLoading: false
      }))
    } catch (error) {
      set({ error: (error as Error).message, isLoading: false })
      throw error // Re-throw para que o componente possa tratar
    }
  },

  deleteContact: async (id: string) => {
    set({ isLoading: true, error: null })
    try {
      const token = useAuthStore.getState().token
      const response = await fetch(`${API_BASE}/contacts/${id}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(`Failed to delete contact: ${errorText}`)
      }

      set((state: ContactState) => ({
        contacts: state.contacts.filter((c) => c.id !== id),
        isLoading: false
      }))
    } catch (error) {
      set({ error: (error as Error).message, isLoading: false })
      throw error // Re-throw para que o componente possa tratar
    }
  },

  clearError: () => {
    set({ error: null })
  },
}))
