import { useState } from 'react'
import { contactsApi } from '../lib/api'
import { 
  User, History, FileText, GitMerge, 
  MapPin, Mail, Phone, Globe, Building, 
  Linkedin, Facebook, Instagram, Github, X,
  Search, AlertTriangle, Save, Loader2 
} from 'lucide-react'

type Tab = 'attributes' | 'history' | 'notes' | 'merge'

interface ContactDetailsPanelProps {
  contact: any
  onClose: () => void
}

export default function ContactDetailsPanel({ contact = {}, onClose }: ContactDetailsPanelProps) {
  const [activeTab, setActiveTab] = useState<Tab>('attributes')
  
  // Mock data para exemplos de UI (já que não temos dados reais ainda)
  const [notes, setNotes] = useState([
    { id: 1, text: 'Cliente interessado no plano Enterprise.', date: '2 dias atrás', author: 'Admin' }
  ])
  const [newNote, setNewNote] = useState('')

  // Merge Logic States
  const [mergeSearchQuery, setMergeSearchQuery] = useState('')
  const [mergeSearchResults, setMergeSearchResults] = useState<any[]>([])
  const [isSearchingMerge, setIsSearchingMerge] = useState(false)
  const [targetContact, setTargetContact] = useState<any>(null)

  const handleMergeSearch = async (query: string) => {
    if (!query.trim()) {
        setMergeSearchResults([])
        return
    }
    
    setIsSearchingMerge(true)
    try {
        const data = await contactsApi.list({ search: query })
        // Garantir que seja array, tratando paginação se houver
        const results = Array.isArray(data) ? data : (data.payload || data.contacts || [])
        
        // Remover o contato atual da lista de resultados (não pode mesclar com si mesmo)
        const filtered = results.filter((c: any) => c.id !== contact.id)
        setMergeSearchResults(filtered)
    } catch (error) {
        console.error('Failed to search contacts for merge', error)
    } finally {
        setIsSearchingMerge(false)
    }
  }

  return (
    <div className="w-96 bg-gray-800 border-l border-gray-700 flex flex-col h-full">
      {/* Header com Tabs */}
      <div className="flex border-b border-gray-700">
        <button
          onClick={() => setActiveTab('attributes')}
          className={`flex-1 py-3 text-sm font-medium border-b-2 transition-colors ${
            activeTab === 'attributes' 
              ? 'border-primary-500 text-primary-500' 
              : 'border-transparent text-gray-400 hover:text-gray-300'
          }`}
        >
          Atributos
        </button>
        <button
          onClick={() => setActiveTab('history')}
          className={`flex-1 py-3 text-sm font-medium border-b-2 transition-colors ${
            activeTab === 'history' 
              ? 'border-primary-500 text-primary-500' 
              : 'border-transparent text-gray-400 hover:text-gray-300'
          }`}
        >
          Histórico
        </button>
        <button
          onClick={() => setActiveTab('notes')}
          className={`flex-1 py-3 text-sm font-medium border-b-2 transition-colors ${
            activeTab === 'notes' 
              ? 'border-primary-500 text-primary-500' 
              : 'border-transparent text-gray-400 hover:text-gray-300'
          }`}
        >
          Notas
        </button>
        <button
          onClick={() => setActiveTab('merge')}
          className={`flex-1 py-3 text-sm font-medium border-b-2 transition-colors ${
            activeTab === 'merge' 
              ? 'border-primary-500 text-primary-500' 
              : 'border-transparent text-gray-400 hover:text-gray-300'
          }`}
        >
          Mesclar
        </button>
      </div>

      {/* Content Area */}
      <div className="flex-1 overflow-y-auto p-4 scrollbar-hide">
        
        {/* === TAB: ATRIBUTOS === */}
        {activeTab === 'attributes' && (
          <div className="space-y-6">
            {/* Profile Info */}
            <div className="flex flex-col items-center mb-6">
              <div className="w-20 h-20 bg-primary-600 rounded-full flex items-center justify-center text-white text-2xl font-bold mb-3">
                {contact?.name?.charAt(0) || 'U'}
              </div>
              <h3 className="text-xl font-bold text-white">{contact?.name || 'Unknown User'}</h3>
              <p className="text-sm text-gray-400">{contact?.email || 'No email'}</p>
              <p className="text-sm text-gray-400">{contact?.phoneNumber || 'No phone'}</p>
            </div>

            {/* Form Fields */}
            <div className="space-y-4">
              <div>
                <label className="block text-xs text-gray-400 mb-1">Nome Completo</label>
                <input 
                  type="text" 
                  defaultValue={contact?.name}
                  className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-primary-500"
                />
              </div>

              <div>
                <label className="block text-xs text-gray-400 mb-1">Email</label>
                <input 
                  type="email" 
                  defaultValue={contact?.email}
                  className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-primary-500"
                />
              </div>

              <div>
                <label className="block text-xs text-gray-400 mb-1">Telefone</label>
                <div className="flex">
                  <span className="bg-gray-600 border border-gray-600 border-r-0 rounded-l-lg px-2 py-2 text-gray-300 text-sm flex items-center">BR +55</span>
                  <input 
                    type="tel" 
                    defaultValue={contact?.phoneNumber}
                    className="flex-1 bg-gray-700 border border-gray-600 rounded-r-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-primary-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs text-gray-400 mb-1">Bio</label>
                <textarea 
                  rows={2}
                  className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-primary-500"
                  placeholder="Adicione uma biografia..."
                />
              </div>

              <div>
                <label className="block text-xs text-gray-400 mb-1">Cidade</label>
                <input 
                  type="text" 
                  className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-primary-500"
                  placeholder="Cidade"
                />
              </div>
            </div>

            {/* Social Media */}
            <div className="pt-4 border-t border-gray-700">
              <h4 className="text-sm font-semibold text-gray-300 mb-3">Redes Sociais</h4>
              <div className="grid grid-cols-2 gap-2">
                <button className="flex items-center gap-2 px-3 py-2 bg-gray-700 rounded-lg text-xs text-gray-300 hover:bg-gray-600 transition-colors">
                  <Linkedin className="w-3 h-3" /> LinkedIn
                </button>
                <button className="flex items-center gap-2 px-3 py-2 bg-gray-700 rounded-lg text-xs text-gray-300 hover:bg-gray-600 transition-colors">
                  <Facebook className="w-3 h-3" /> Facebook
                </button>
                <button className="flex items-center gap-2 px-3 py-2 bg-gray-700 rounded-lg text-xs text-gray-300 hover:bg-gray-600 transition-colors">
                  <Instagram className="w-3 h-3" /> Instagram
                </button>
                <button className="flex items-center gap-2 px-3 py-2 bg-gray-700 rounded-lg text-xs text-gray-300 hover:bg-gray-600 transition-colors">
                  <Github className="w-3 h-3" /> Github
                </button>
              </div>
            </div>

            <button className="w-full mt-4 bg-primary-600 text-white py-2 rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors flex items-center justify-center gap-2">
              <Save className="w-4 h-4" />
              Atualizar Contato
            </button>
          </div>
        )}

        {/* === TAB: HISTÓRICO === */}
        {activeTab === 'history' && (
          <div className="space-y-4">
            <h4 className="text-sm font-semibold text-gray-300">Conversas Anteriores</h4>
            <div className="space-y-3">
              {[1, 2, 3].map((_, i) => (
                <div key={i} className="bg-gray-700 p-3 rounded-lg border border-gray-600 cursor-pointer hover:border-gray-500">
                  <div className="flex justify-between items-start mb-1">
                    <span className="text-xs font-medium text-primary-400">#123{i} - Suporte</span>
                    <span className="text-[10px] text-gray-400">2 dias atrás</span>
                  </div>
                  <p className="text-xs text-gray-300 line-clamp-2">
                    O cliente entrou em contato para resolver um problema de conexão com a API...
                  </p>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* === TAB: NOTAS === */}
        {activeTab === 'notes' && (
          <div className="flex flex-col h-full">
            <div className="flex-1 space-y-4 mb-4">
              {notes.map((note) => (
                <div key={note.id} className="bg-yellow-900/20 border border-yellow-800/30 p-3 rounded-lg">
                  <p className="text-sm text-gray-200 mb-2">{note.text}</p>
                  <div className="flex items-center justify-between text-[10px] text-gray-500">
                    <span className="font-medium text-yellow-600">{note.author}</span>
                    <span>{note.date}</span>
                  </div>
                </div>
              ))}
            </div>
            
            <div className="mt-auto">
              <textarea
                value={newNote}
                onChange={(e) => setNewNote(e.target.value)}
                placeholder="Adicionar uma nota interna..."
                className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-primary-500 mb-2"
                rows={3}
              />
              <button 
                onClick={() => {
                  if (newNote.trim()) {
                    setNotes([...notes, { id: Date.now(), text: newNote, date: 'Agora', author: 'Você' }])
                    setNewNote('')
                  }
                }}
                className="w-full bg-gray-700 border border-gray-600 text-white py-2 rounded-lg text-sm hover:bg-gray-600 transition-colors"
              >
                Adicionar Nota
              </button>
            </div>
          </div>
        )}

        {/* === TAB: MESCLAR === */}
        {activeTab === 'merge' && (
          <div className="space-y-6">
            <div className="bg-blue-900/20 border border-blue-800 p-3 rounded-lg flex gap-3">
              <AlertTriangle className="w-5 h-5 text-blue-400 flex-shrink-0" />
              <p className="text-xs text-blue-200">
                Mescle contatos para combinar dois perfis em um. Os atributos do contato principal terão prioridade.
              </p>
            </div>

            <div className="space-y-4">
              {/* Target Contact Search/Display */}
              <div>
                <div className="flex justify-between items-center mb-1">
                  <label className="text-xs text-gray-400">Contato Principal</label>
                  <span className="text-[10px] bg-green-900/50 text-green-400 px-1.5 py-0.5 rounded border border-green-800">A ser salvo</span>
                </div>
                
                {targetContact ? (
                  <div className="bg-gray-700 border border-green-700/50 rounded-lg p-3 flex items-center justify-between group">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 bg-green-900/50 text-green-400 rounded-full flex items-center justify-center text-xs font-bold border border-green-900">
                        {targetContact.name?.charAt(0) || 'U'}
                      </div>
                      <div>
                        <p className="text-sm font-medium text-white">{targetContact.name}</p>
                        <p className="text-xs text-gray-400">{targetContact.email || targetContact.phoneNumber}</p>
                      </div>
                    </div>
                    <button 
                      onClick={() => {
                        setTargetContact(null)
                        setMergeSearchQuery('')
                      }}
                      className="p-1 text-gray-400 hover:text-white hover:bg-gray-600 rounded opacity-0 group-hover:opacity-100 transition-all"
                    >
                      <X className="w-4 h-4" />
                    </button>
                  </div>
                ) : (
                  <div className="relative">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
                    <input
                      type="text"
                      placeholder="Pesquisar contato principal..."
                      value={mergeSearchQuery}
                      onChange={(e) => {
                        setMergeSearchQuery(e.target.value)
                        handleMergeSearch(e.target.value)
                      }}
                      className="w-full bg-gray-700 border border-gray-600 rounded-lg pl-9 pr-3 py-2 text-white text-sm focus:outline-none focus:border-primary-500"
                    />
                    {isSearchingMerge && (
                       <div className="absolute right-3 top-1/2 -translate-y-1/2">
                         <Loader2 className="w-3 h-3 text-gray-400 animate-spin" />
                       </div>
                    )}

                    {/* Search Results Dropdown */}
                    {mergeSearchResults.length > 0 && (
                      <div className="absolute top-full left-0 right-0 mt-1 bg-gray-700 border border-gray-600 rounded-lg shadow-xl z-20 max-h-48 overflow-y-auto">
                        {mergeSearchResults.map(result => (
                          <button
                            key={result.id}
                            onClick={() => {
                                setTargetContact(result)
                                setMergeSearchResults([])
                            }}
                            className="w-full text-left px-3 py-2 hover:bg-gray-600 flex items-center gap-2 border-b border-gray-600/50 last:border-0"
                          >
                             <div className="w-6 h-6 bg-gray-500 rounded-full flex items-center justify-center text-[10px] text-white font-bold shrink-0">
                                {result.name?.charAt(0) || 'U'}
                             </div>
                             <div className="min-w-0">
                               <p className="text-sm text-gray-200 truncate">{result.name}</p>
                               <p className="text-[10px] text-gray-400 truncate">{result.email || result.phoneNumber}</p>
                             </div>
                          </button>
                        ))}
                      </div>
                    )}
                  </div>
                )}
              </div>

              <div className="flex justify-center my-2">
                <div className="flex flex-col items-center">
                  <span className="text-xs text-gray-500 mb-1">Para ser excluído</span>
                  <div className="w-0.5 h-6 bg-gray-700"></div>
                  <div className="w-2 h-2 rounded-full bg-gray-700"></div>
                </div>
              </div>

              <div className="bg-gray-700/50 border border-gray-700 rounded-lg p-3 flex items-center gap-3 opacity-60">
                <div className="w-8 h-8 bg-primary-600 rounded-full flex items-center justify-center text-white text-xs font-bold">
                  {contact?.name?.charAt(0) || 'U'}
                </div>
                <div>
                  <p className="text-sm font-medium text-white">{contact?.name || 'Contato Atual'}</p>
                  <p className="text-xs text-gray-400">Este contato será excluído</p>
                </div>
              </div>

              <div className="flex gap-2 pt-4">
                <button 
                  onClick={() => setActiveTab('attributes')}
                  className="flex-1 py-2 bg-transparent border border-gray-600 text-gray-300 rounded-lg text-sm hover:bg-gray-700 transition-colors"
                >
                  Cancelar
                </button>
                <button 
                  disabled={!targetContact}
                  className="flex-1 py-2 bg-primary-600 text-white rounded-lg text-sm hover:bg-primary-700 transition-colors flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <GitMerge className="w-4 h-4" />
                  Mesclar
                </button>
              </div>
            </div>
          </div>
        )}

      </div>
    </div>
  )
}
