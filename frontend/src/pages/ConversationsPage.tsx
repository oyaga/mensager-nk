import { useState, useEffect } from 'react'
import { MessageSquare, Search, Filter, Plus, Info, Phone, Video } from 'lucide-react'
import ContactDetailsPanel from '../components/ContactDetailsPanel'
import NewConversationModal from '../components/NewConversationModal'
import ChatPanel from '../components/ChatPanel'
import { conversationsApi } from '../lib/api'

interface Contact {
  id: string
  name: string
  phone_number: string
  avatar?: string
  email?: string
}

interface Conversation {
  id: string
  contact: Contact
  last_message?: string
  updated_at: string
  unread_count?: number
  status: 'open' | 'resolved' | 'pending'
}

export default function ConversationsPage() {
  const [selectedConversationId, setSelectedConversationId] = useState<string | null>(null)
  const [showContactInfo, setShowContactInfo] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [isNewConversationModalOpen, setIsNewConversationModalOpen] = useState(false)
  const [conversations, setConversations] = useState<Conversation[]>([])
  const [isLoading, setIsLoading] = useState(true)

  // Fetch conversations
  useEffect(() => {
    fetchConversations()
  }, [])

  const fetchConversations = async () => {
    setIsLoading(true)
    try {
      const data = await conversationsApi.list()
      // Normalize data
      const convs = Array.isArray(data) ? data : data.conversations || []
      setConversations(convs.map((c: any) => ({
        id: c.id,
        contact: {
          id: c.contact?.id || c.contact_id,
          name: c.contact?.name || 'Sem nome',
          phone_number: c.contact?.phone_number || '',
          avatar: c.contact?.avatar,
          email: c.contact?.email
        },
        last_message: c.last_message || c.messages?.[0]?.content || '',
        updated_at: c.updated_at || c.last_activity_at,
        unread_count: c.unread_count || 0,
        status: c.status || 'open'
      })))
    } catch (error) {
      console.error('Failed to fetch conversations:', error)
    } finally {
      setIsLoading(false)
    }
  }

  // Filtered conversations
  const filteredConversations = conversations.filter(conv => {
    if (!searchQuery) return true
    const lowerQ = searchQuery.toLowerCase()
    return (
      conv.contact.name.toLowerCase().includes(lowerQ) ||
      conv.contact.phone_number.includes(searchQuery) ||
      (conv.last_message && conv.last_message.toLowerCase().includes(lowerQ))
    )
  })

  const selectedConversation = conversations.find(c => c.id === selectedConversationId)

  const handleSelectConversation = (id: string) => {
    setSelectedConversationId(id)
    setShowContactInfo(true)
  }

  const handleConversationCreated = (id: string) => {
    fetchConversations()
    setSelectedConversationId(id)
  }

  const formatTime = (dateStr: string) => {
    if (!dateStr) return ''
    const date = new Date(dateStr)
    const now = new Date()
    const diffDays = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60 * 24))
    
    if (diffDays === 0) {
      return date.toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' })
    } else if (diffDays === 1) {
      return 'Ontem'
    } else if (diffDays < 7) {
      return date.toLocaleDateString('pt-BR', { weekday: 'short' })
    } else {
      return date.toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit' })
    }
  }

  return (
    <div className="h-full flex overflow-hidden bg-gray-900">
      
      {/* 1. CONVERSATIONS LIST */}
      <div className="w-96 bg-gray-800 border-r border-gray-700 flex flex-col flex-shrink-0 z-10">
        <div className="p-4 border-b border-gray-700">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-white">Conversas</h2>
            <div className="flex items-center gap-1">
              <button 
                onClick={() => setIsNewConversationModalOpen(true)}
                className="p-2 text-gray-400 hover:text-white hover:bg-primary-600 rounded-lg transition-colors"
                title="Nova Conversa"
              >
                <Plus className="w-5 h-5" />
              </button>
              <button className="p-2 text-gray-400 hover:text-white hover:bg-gray-700 rounded-lg transition-colors">
                <Filter className="w-5 h-5" />
              </button>
            </div>
          </div>
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
            <input
              type="search"
              placeholder="Buscar conversas..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-2 bg-gray-700 border border-gray-600 text-gray-100 placeholder-gray-400 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 text-sm"
            />
          </div>
        </div>

        <div className="flex-1 overflow-y-auto custom-scrollbar">
          {isLoading ? (
            <div className="p-8 text-center text-gray-400">
              <div className="w-6 h-6 border-2 border-primary-500 border-t-transparent rounded-full animate-spin mx-auto mb-2"></div>
              <p className="text-sm">Carregando...</p>
            </div>
          ) : filteredConversations.length === 0 ? (
            <div className="p-8 text-center text-gray-400 mt-10">
              <MessageSquare className="w-12 h-12 mx-auto mb-4 text-gray-600" />
              <p className="text-sm">Nenhuma conversa encontrada</p>
              <button
                onClick={() => setIsNewConversationModalOpen(true)}
                className="mt-4 text-primary-400 hover:text-primary-300 text-sm font-medium"
              >
                + Iniciar nova conversa
              </button>
            </div>
          ) : (
            filteredConversations.map((conv) => (
              <button
                key={conv.id}
                onClick={() => handleSelectConversation(conv.id)}
                className={`w-full p-4 border-b border-gray-700 hover:bg-gray-700/50 transition-colors text-left ${
                  selectedConversationId === conv.id ? 'bg-gray-700 border-l-4 border-l-primary-500' : 'border-l-4 border-l-transparent'
                }`}
              >
                <div className="flex items-start gap-3">
                  <div className="w-12 h-12 bg-primary-600 rounded-full flex items-center justify-center text-white font-semibold flex-shrink-0 overflow-hidden">
                    {conv.contact.avatar ? (
                      <img src={conv.contact.avatar} alt="" className="w-full h-full object-cover" />
                    ) : (
                      <span>{conv.contact.name.charAt(0).toUpperCase()}</span>
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex justify-between items-baseline mb-1">
                      <h3 className="font-medium text-white truncate">{conv.contact.name}</h3>
                      <span className="text-xs text-gray-400 flex-shrink-0 ml-2">
                        {formatTime(conv.updated_at)}
                      </span>
                    </div>
                    <p className="text-sm text-gray-400 truncate">
                      {conv.last_message || 'Sem mensagens'}
                    </p>
                  </div>
                  {conv.unread_count && conv.unread_count > 0 && (
                    <span className="bg-primary-600 text-white text-xs font-bold rounded-full w-5 h-5 flex items-center justify-center">
                      {conv.unread_count}
                    </span>
                  )}
                </div>
              </button>
            ))
          )}
        </div>
      </div>

      {/* 2. CHAT AREA */}
      <div className="flex-1 flex flex-col bg-gray-900 min-w-0 relative">
        {selectedConversation ? (
          <>
            {/* Chat Header */}
            <div className="h-16 px-6 border-b border-gray-700 flex items-center justify-between bg-gray-800/50 backdrop-blur-sm z-10">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-primary-600 rounded-full flex items-center justify-center text-white font-semibold overflow-hidden">
                  {selectedConversation.contact.avatar ? (
                    <img src={selectedConversation.contact.avatar} alt="" className="w-full h-full object-cover" />
                  ) : (
                    <span>{selectedConversation.contact.name.charAt(0).toUpperCase()}</span>
                  )}
                </div>
                <div>
                  <h3 className="font-semibold text-white">{selectedConversation.contact.name}</h3>
                  <p className="text-xs text-gray-400">{selectedConversation.contact.phone_number}</p>
                </div>
              </div>
              
              <div className="flex items-center gap-2">
                <button className="p-2 text-gray-400 hover:text-white hover:bg-gray-700 rounded-lg">
                  <Phone className="w-5 h-5" />
                </button>
                <button className="p-2 text-gray-400 hover:text-white hover:bg-gray-700 rounded-lg">
                  <Video className="w-5 h-5" />
                </button>
                <div className="h-6 w-px bg-gray-700 mx-1"></div>
                <button 
                  onClick={() => setShowContactInfo(!showContactInfo)}
                  className={`p-2 rounded-lg transition-colors ${
                    showContactInfo ? 'text-primary-400 bg-primary-900/20' : 'text-gray-400 hover:text-white hover:bg-gray-700'
                  }`}
                  title="Detalhes do Contato"
                >
                  <Info className="w-5 h-5" />
                </button>
              </div>
            </div>

            {/* Chat Panel */}
            <ChatPanel 
              conversationId={selectedConversation.id}
              contactName={selectedConversation.contact.name}
            />
          </>
        ) : (
          <div className="flex-1 flex flex-col items-center justify-center text-gray-500">
            <div className="w-20 h-20 bg-gray-800 rounded-full flex items-center justify-center mb-6">
              <MessageSquare className="w-10 h-10 text-gray-600" />
            </div>
            <h3 className="text-xl font-medium text-white mb-2">Selecione uma conversa</h3>
            <p className="max-w-xs text-center text-sm">Escolha uma conversa da lista ou inicie uma nova</p>
            <button
              onClick={() => setIsNewConversationModalOpen(true)}
              className="mt-6 px-4 py-2 bg-primary-600 hover:bg-primary-500 text-white rounded-lg font-medium transition-colors"
            >
              + Nova Conversa
            </button>
          </div>
        )}
      </div>

      {/* 3. CONTACT DETAILS PANEL */}
      {selectedConversation && showContactInfo && (
        <ContactDetailsPanel 
          contact={selectedConversation.contact} 
          onClose={() => setShowContactInfo(false)} 
        />
      )}

      {/* 4. NEW CONVERSATION MODAL */}
      <NewConversationModal
        isOpen={isNewConversationModalOpen}
        onClose={() => setIsNewConversationModalOpen(false)}
        onConversationCreated={handleConversationCreated}
      />

    </div>
  )
}
