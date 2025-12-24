import React, { useState, useEffect } from 'react'
import { X, Search, MessageSquare } from 'lucide-react'
import { contactsApi, conversationsApi } from '../lib/api'

interface Contact {
  id: string
  name: string
  email?: string
  phone_number: string
  avatar_url?: string
}

interface NewConversationModalProps {
  isOpen: boolean
  onClose: () => void
  onConversationCreated: (conversationId: string) => void
}

export default function NewConversationModal({ isOpen, onClose, onConversationCreated }: NewConversationModalProps) {
  const [query, setQuery] = useState('')
  const [contacts, setContacts] = useState<Contact[]>([])
  const [loading, setLoading] = useState(false)
  const [creating, setCreating] = useState(false)

  // Fetch contacts on open or search debounced
  useEffect(() => {
    if (isOpen) {
      const timer = setTimeout(() => {
        fetchContacts()
      }, 300)
      return () => clearTimeout(timer)
    }
  }, [isOpen, query])

  const fetchContacts = async () => {
    setLoading(true)
    try {
      // Assumindo que a API suporta filtro simples ou retorna todos
      // O backend atual retorna array de contatos
      const data = await contactsApi.list({ search: query }) 
      let loadedContacts = Array.isArray(data) ? data : (data.payload || data.contacts || [])

      // Filtro client-side simples se a API não filtrar
      if (query && !data.filtered) {
          const lowerQ = query.toLowerCase()
          loadedContacts = loadedContacts.filter((c: Contact) => 
            c.name.toLowerCase().includes(lowerQ) || 
            c.phone_number.includes(lowerQ)
          )
      }

      setContacts(loadedContacts)
    } catch (error) {
      console.error(error)
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = async (contactId: string) => {
    setCreating(true)
    try {
      console.log('Creating conversation for contact:', contactId)
      const conv = await conversationsApi.create({ contact_id: contactId })
      console.log('Conversation created:', conv)
      onConversationCreated(conv.id)
      onClose()
    } catch (error) {
      console.error('Failed to create conversation', error)
      alert("Erro ao criar conversa. Verifique o console.")
    } finally {
      setCreating(false)
    }
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-fade-in">
      <div 
        className="fixed inset-0" 
        onClick={onClose}
      />
      <div className="bg-gray-800 border border-gray-700 rounded-xl w-full max-w-md shadow-2xl flex flex-col max-h-[80vh] relative animate-scale-up z-10">
        
        {/* Header */}
        <div className="p-4 border-b border-gray-700 flex justify-between items-center bg-gray-800 rounded-t-xl">
          <h2 className="text-lg font-semibold text-white flex items-center gap-2">
            <MessageSquare className="w-5 h-5 text-primary-500" />
            Nova Conversa
          </h2>
          <button onClick={onClose} className="text-gray-400 hover:text-white transition-colors">
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Search */}
        <div className="p-4 border-b border-gray-700 bg-gray-800/50">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
            <input
              type="text"
              placeholder="Buscar contato por nome ou telefone..."
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              className="w-full bg-gray-900 border border-gray-700 rounded-lg pl-10 pr-4 py-2.5 text-white placeholder-gray-500 focus:outline-none focus:ring-1 focus:ring-primary-500 text-sm"
              autoFocus
            />
          </div>
        </div>

        {/* List */}
        <div className="flex-1 overflow-y-auto p-2 space-y-1 custom-scrollbar">
          {loading ? (
             <div className="p-8 text-center text-gray-500 flex flex-col items-center gap-2">
                <div className="w-5 h-5 border-2 border-primary-500 border-t-transparent rounded-full animate-spin"></div>
                <span className="text-sm">Buscando...</span>
             </div>
          ) : contacts.length === 0 ? (
             <div className="p-8 text-center text-gray-500">
               <p>Nenhum contato encontrado.</p>
               <p className="text-xs mt-1">Tente buscar por outro nome.</p>
             </div>
          ) : (
             contacts.map(contact => (
               <button
                 key={contact.id}
                 onClick={() => handleCreate(contact.id)}
                 disabled={creating}
                 className="w-full flex items-center gap-3 p-3 rounded-lg hover:bg-gray-700 transition-colors text-left group disabled:opacity-50"
               >
                 <div className="w-10 h-10 rounded-full bg-primary-900/50 text-primary-200 flex items-center justify-center font-semibold group-hover:bg-primary-600 group-hover:text-white transition-colors border border-primary-900/50 overflow-hidden shrink-0">
                    {contact.avatar_url ? (
                        <img src={contact.avatar_url} alt="" className="w-full h-full object-cover" />
                    ) : (
                        <span>{contact.name.charAt(0).toUpperCase()}</span>
                    )}
                 </div>
                 <div className="min-w-0 flex-1">
                   <h3 className="text-sm font-medium text-white truncate">{contact.name}</h3>
                   <div className="flex items-center gap-2 text-xs text-gray-400">
                      <span>{contact.phone_number}</span>
                      {contact.email && (
                          <>
                            <span className="w-1 h-1 bg-gray-600 rounded-full"></span>
                            <span className="truncate">{contact.email}</span>
                          </>
                      )}
                   </div>
                 </div>
               </button>
             ))
          )}
        </div>

      </div>
    </div>
  )
}
