import { useState, useEffect } from 'react'
import { Users, Search, Filter, Plus, Mail, Phone, MoreVertical, Edit2, Trash2, RefreshCw, ChevronLeft, ChevronRight } from 'lucide-react'
import ContactModal from '../components/CreateContactModal'
import { useContactStore, Contact } from '../stores/contactStore'

export default function ContactsPage() {
  const { contacts, meta, isLoading, error, fetchContacts, createContact, updateContact, deleteContact, clearError } = useContactStore()

  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingContact, setEditingContact] = useState<Contact | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [openActionId, setOpenActionId] = useState<string | null>(null)

  const [selectedContactIds, setSelectedContactIds] = useState<string[]>([])
  const [itemsPerPage, setItemsPerPage] = useState(30) // Default 30 items

  // Load contacts limits changes
  useEffect(() => {
    fetchContacts(1, itemsPerPage)
  }, [fetchContacts, itemsPerPage])

  // Clear errors when modal closes
  useEffect(() => {
    if (!isModalOpen) {
      clearError()
    }
  }, [isModalOpen, clearError])

  const filteredContacts = contacts.filter(contact => 
    contact.name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
    contact.email?.toLowerCase().includes(searchQuery.toLowerCase()) ||
    contact.phone_number?.includes(searchQuery)
  )

  const handleSaveContact = async (data: any) => {
    try {
      // Converter camelCase para snake_case se necessário
      const payload = {
        name: data.name,
        email: data.email,
        phone_number: data.phoneNumber,
        // outros campos se houver
      }

      if (editingContact) {
        await updateContact(editingContact.id, payload)
      } else {
        await createContact(payload)
      }

      // Recarregar a lista após criar/atualizar para garantir sincronia
      await fetchContacts()

      setIsModalOpen(false)
      setEditingContact(null)
    } catch (error) {
      console.error('Erro ao salvar contato:', error)
      // Modal permanece aberto em caso de erro
    }
  }

  const handleEditClick = (contact: Contact) => {
    // Adapter para o formato que o modal espera
    const modalData = {
      id: contact.id,
      name: contact.name,
      email: contact.email,
      phoneNumber: contact.phone_number, // Converte snake_case para camelCase
      company: '' // Campo extra do modal
    }
    setEditingContact(modalData as any)
    setIsModalOpen(true)
    setOpenActionId(null)
  }

  const handleDeleteClick = async (id: string) => {
    if (confirm('Tem certeza que deseja excluir este contato?')) {
      await deleteContact(id)
    }
    setOpenActionId(null)
  }

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
        setSelectedContactIds(filteredContacts.map(c => c.id))
    } else {
        setSelectedContactIds([])
    }
  }

  const handleSelectContact = (id: string, checked: boolean) => {
    if (checked) {
        setSelectedContactIds([...selectedContactIds, id])
    } else {
        setSelectedContactIds(selectedContactIds.filter(cid => cid !== id))
    }
  }

  const handleBulkDelete = async () => {
    if (confirm(`Tem certeza que deseja excluir ${selectedContactIds.length} contatos?`)) {
        for (const id of selectedContactIds) {
            await deleteContact(id)
        }
        setSelectedContactIds([])
        fetchContacts(1, itemsPerPage, true)
    }
  }

  return (
    <div className="h-full flex flex-col bg-gray-900 overflow-hidden" onClick={() => setOpenActionId(null)}>
      {/* Header */}
      <div className="px-8 py-6 border-b border-gray-800 flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white mb-1">Contatos</h1>
          <p className="text-sm text-gray-400">Gerencie sua base de clientes e leads</p>
        </div>
        
        <div className="flex gap-3">
          {selectedContactIds.length > 0 && (
            <button
              onClick={handleBulkDelete}
              className="flex items-center justify-center p-2.5 bg-red-900/50 hover:bg-red-800 text-red-200 rounded-lg transition-colors border border-red-800 mr-2"
              title="Excluir selecionados"
            >
              <Trash2 className="w-5 h-5" />
              <span className="ml-2 text-sm font-medium">{selectedContactIds.length}</span>
            </button>
          )}

          <button
            onClick={() => fetchContacts(1, itemsPerPage, true)} // force refresh
            disabled={isLoading}
            className="flex items-center justify-center p-2.5 bg-gray-800 hover:bg-gray-700 text-gray-400 hover:text-white rounded-lg transition-colors border border-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
            title="Atualizar lista"
          >
            <RefreshCw className={`w-5 h-5 ${isLoading ? 'animate-spin' : ''}`} />
          </button>
          
          <button 
            onClick={() => {
              setEditingContact(null)
              setIsModalOpen(true)
            }}
            className="flex items-center justify-center gap-2 bg-primary-600 hover:bg-primary-700 text-white px-4 py-2.5 rounded-lg font-medium transition-all shadow-lg shadow-primary-900/20 active:scale-95"
          >
            <Plus className="w-5 h-5" />
            <span>Novo Contato</span>
          </button>
        </div>
      </div>

      {/* Filters & Search */}
      <div className="px-8 py-4 bg-gray-900 border-b border-gray-800 flex gap-3">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
          <input
            type="text"
            placeholder="Buscar por nome, email ou telefone..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full bg-gray-800 border border-gray-700 rounded-lg pl-10 pr-4 py-2 text-white placeholder-gray-500 focus:outline-none focus:border-primary-500 focus:ring-1 focus:ring-primary-500"
          />
        </div>
        <button className="flex items-center gap-2 px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-gray-300 hover:bg-gray-700 hover:text-white transition-colors">
          <Filter className="w-4 h-4" />
          <span className="text-sm">Filtros</span>
        </button>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-8">
        {error && (
          <div className="mb-4 p-4 bg-red-900/50 border border-red-800 text-red-200 rounded-lg">
            Erro ao carregar contatos: {error}
          </div>
        )}
        {isLoading && contacts.length === 0 ? (
          <div className="flex justify-center py-20">
            <RefreshCw className="w-8 h-8 text-primary-500 animate-spin" />
          </div>
        ) : contacts.length === 0 ? (
          // Empty State
          <div className="h-full flex flex-col items-center justify-center text-center pb-20">
            <div className="w-20 h-20 bg-gray-800 rounded-full flex items-center justify-center mb-6 ring-4 ring-gray-800/50">
              <Users className="w-10 h-10 text-gray-600" />
            </div>
            <h2 className="text-xl font-semibold text-white mb-2">Nenhum contato encontrado</h2>
            <p className="text-gray-400 max-w-sm mb-8">
              Sua lista de contatos está vazia. Adicione contatos manualmente ou importe-os para começar.
            </p>
            <button 
              onClick={() => {
                setEditingContact(null)
                setIsModalOpen(true)
              }}
              className="px-6 py-3 bg-primary-600 hover:bg-primary-700 text-white rounded-lg font-medium transition-colors"
            >
              Adicionar Primeiro Contato
            </button>
          </div>
        ) : filteredContacts.length === 0 ? (
          // No Search Results
          <div className="text-center py-20">
            <Search className="w-12 h-12 text-gray-700 mx-auto mb-4" />
            <p className="text-gray-400">Nenhum contato corresponde à sua busca.</p>
          </div>
        ) : (
          // Contacts Table
          <div className="bg-gray-800 rounded-xl border border-gray-700 overflow-visible shadow-sm flex flex-col">
            <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="bg-gray-800/50 border-b border-gray-700">
                  <th className="px-6 py-4 w-10">
                    <input 
                      type="checkbox" 
                      className="rounded bg-gray-700 border-gray-600 text-primary-600 focus:ring-primary-500"
                      checked={filteredContacts.length > 0 && selectedContactIds.length === filteredContacts.length}
                      onChange={(e) => handleSelectAll(e.target.checked)}
                    />
                  </th>
                  <th className="px-6 py-4 text-xs font-semibold text-gray-400 uppercase tracking-wider">Nome</th>
                  <th className="px-6 py-4 text-xs font-semibold text-gray-400 uppercase tracking-wider">Email</th>
                  <th className="px-6 py-4 text-xs font-semibold text-gray-400 uppercase tracking-wider">Telefone</th>
                  <th className="px-6 py-4 text-xs font-semibold text-gray-400 uppercase tracking-wider">Criado em</th>
                  <th className="px-6 py-4 text-xs font-semibold text-gray-400 uppercase tracking-wider text-right">Ações</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-700">
                {filteredContacts.map((contact) => (
                  <tr key={contact.id} className={`hover:bg-gray-700/50 transition-colors group ${selectedContactIds.includes(contact.id) ? 'bg-gray-700/30' : ''}`}>
                    <td className="px-6 py-4">
                      <input 
                        type="checkbox" 
                        className="rounded bg-gray-700 border-gray-600 text-primary-600 focus:ring-primary-500"
                        checked={selectedContactIds.includes(contact.id)}
                        onChange={(e) => handleSelectContact(contact.id, e.target.checked)}
                      />
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        <div className="w-9 h-9 bg-primary-900/50 text-primary-400 rounded-full flex items-center justify-center text-sm font-bold border border-primary-900">
                          {contact.name?.charAt(0).toUpperCase()}
                        </div>
                        <span className="text-sm font-medium text-white">{contact.name}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2 text-sm text-gray-300">
                        {contact.email && <Mail className="w-3 h-3 text-gray-500" />}
                        {contact.email || '-'}
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2 text-sm text-gray-300">
                        {contact.phone_number && <Phone className="w-3 h-3 text-gray-500" />}
                        {contact.phone_number || '-'}
                      </div>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-400">
                      {new Date(contact.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4 text-right relative">
                      <button 
                        onClick={(e) => {
                          e.stopPropagation()
                          setOpenActionId(openActionId === contact.id ? null : contact.id)
                        }}
                        className={`p-1.5 rounded-lg transition-colors ${
                          openActionId === contact.id ? 'text-white bg-gray-600' : 'text-gray-500 hover:text-white hover:bg-gray-600'
                        }`}
                      >
                        <MoreVertical className="w-4 h-4" />
                      </button>

                      {/* Dropdown Menu */}
                      {openActionId === contact.id && (
                        <div className="absolute right-8 top-8 w-48 bg-gray-800 border border-gray-700 rounded-lg shadow-xl z-50 overflow-hidden animate-fade-in">
                          <button
                            onClick={(e) => {
                              e.stopPropagation()
                              handleEditClick(contact)
                            }}
                            className="w-full text-left px-4 py-3 text-sm text-gray-300 hover:bg-gray-700 hover:text-white flex items-center gap-2"
                          >
                            <Edit2 className="w-4 h-4" /> Editar
                          </button>
                          <button
                            onClick={(e) => {
                              e.stopPropagation()
                              handleDeleteClick(contact.id)
                            }}
                            className="w-full text-left px-4 py-3 text-sm text-red-400 hover:bg-red-900/20 hover:text-red-300 flex items-center gap-2 border-t border-gray-700"
                          >
                            <Trash2 className="w-4 h-4" /> Excluir
                          </button>
                        </div>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            </div>
            
            {/* Pagination & Limits */}
            {meta && (
              <div className="px-6 py-4 border-t border-gray-700 flex items-center justify-between">
                <div className="flex items-center gap-4">
                    <span className="text-sm text-gray-400">
                    Mostrando {(meta.current_page - 1) * itemsPerPage + 1} a {Math.min(meta.current_page * itemsPerPage, meta.count)} de {meta.count} total
                    </span>
                    <select
                        value={itemsPerPage}
                        onChange={(e) => setItemsPerPage(Number(e.target.value))}
                        className="bg-gray-800 border border-gray-600 text-gray-300 text-xs rounded-lg px-2 py-1 focus:outline-none focus:border-primary-500"
                    >
                        <option value={30}>30 por página</option>
                        <option value={60}>60 por página</option>
                        <option value={100}>100 por página</option>
                    </select>
                </div>

                <div className="flex gap-2">
                  <button
                    disabled={meta.current_page === 1}
                    onClick={() => fetchContacts(meta.current_page - 1, itemsPerPage)}
                    className="p-2 border border-gray-700 rounded-lg hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                  >
                    <ChevronLeft className="w-4 h-4 text-gray-300" />
                  </button>
                  <span className="flex items-center px-2 text-sm text-gray-300">
                    {meta.current_page} de {Math.ceil(meta.count / itemsPerPage)}
                  </span>
                  <button
                    disabled={meta.current_page >= Math.ceil(meta.count / itemsPerPage)}
                    onClick={() => fetchContacts(meta.current_page + 1, itemsPerPage)}
                    className="p-2 border border-gray-700 rounded-lg hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                  >
                    <ChevronRight className="w-4 h-4 text-gray-300" />
                  </button>
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {isModalOpen && (
        <ContactModal 
          initialData={editingContact}
          onClose={() => {
            setIsModalOpen(false)
            setEditingContact(null)
          }} 
          onSave={handleSaveContact} 
        />
      )}
    </div>
  )
}
