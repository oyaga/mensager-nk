import React, { useState, useEffect, useRef } from 'react'
import { 
  User, Mail, Lock, Key, Keyboard, 
  Camera, Save, RefreshCw, Copy, Check,
  Shield, AlertCircle
} from 'lucide-react'
import { useAuthStore } from '../stores/authStore'
import { authApi, storageApi } from '../lib/api'

export default function SettingsPage() {
  const { user, updateUser, login } = useAuthStore()
  
  const [userData, setUserData] = useState({
    fullName: '',
    displayName: '',
    email: '',
    avatar: null as string | null,
    sendShortcut: 'enter',
  })
  // Token state removed, using user.access_token directly

  const [localAccounts, setLocalAccounts] = useState<any[]>([])

  useEffect(() => {
    // Refresh user profile data on load to ensure we have the latest token
    const refreshProfile = async () => {
        try {
            const profile = await authApi.getProfile()
            if (profile) {
                updateUser(profile)
                if (profile.accounts && profile.accounts.length > 0) {
                    setLocalAccounts(profile.accounts)
                    console.log('Accounts loaded:', profile.accounts)
                }
            }
        } catch (error) {
            console.error('Failed to refresh profile:', error)
        }
    }
    refreshProfile()
  }, [])

  useEffect(() => {
    if (user) {
      setUserData(prev => ({
        ...prev,
        fullName: user.name || '',
        displayName: user.display_name || user.name || '', 
        email: user.email || '',
        avatar: user.avatar || null,
        sendShortcut: user.ui_settings?.send_shortcut || 'enter'
      }))
      
      // Also try to get accounts from user store if local is empty
      if (localAccounts.length === 0 && user.accounts && user.accounts.length > 0) {
          setLocalAccounts(user.accounts)
      }
    }
  }, [user])

  // ... (rest of code)

  const handleCopyAccountId = () => {
    // Try multiple sources for account_id
    let accountId = localAccounts?.[0]?.id || user?.accounts?.[0]?.id
    
    // Fallback: decode JWT token to get account_id
    if (!accountId) {
        const { token } = useAuthStore.getState()
        if (token) {
            try {
                const payload = JSON.parse(atob(token.split('.')[1]))
                accountId = payload.account_id
            } catch (e) {
                console.error('Failed to decode JWT:', e)
            }
        }
    }
    
    if (accountId) {
        navigator.clipboard.writeText(accountId)
        setCopiedAccountId(true)
        setTimeout(() => setCopiedAccountId(false), 2000)
    } else {
        alert('Nenhuma conta encontrada. Tente fazer logout e login novamente.')
    }
  }

  const [passwords, setPasswords] = useState({ current: '', new: '', confirm: '' })
  const [isDirty, setIsDirty] = useState(false)
  const [copiedToken, setCopiedToken] = useState(false)
  const [copiedAccountId, setCopiedAccountId] = useState(false)
  const [isSaving, setIsSaving] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  // Handlers
  const handleInputChange = (field: keyof typeof userData, value: any) => {
    setUserData({ ...userData, [field]: value })
    setIsDirty(true)
  }

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    // Preview local
    const reader = new FileReader()
    reader.onloadend = () => {
      setUserData(prev => ({ ...prev, avatar: reader.result as string }))
    }
    reader.readAsDataURL(file)

     // Upload para Minio
    try {
        const { url } = await storageApi.upload(file)
        console.log("Avatar uploaded:", url)
        
        // Atualiza com a URL retornada
        setUserData(prev => ({ ...prev, avatar: url }))
        setIsDirty(true)
    } catch (error) {
        console.error("Upload Error:", error)
        alert("Erro ao fazer upload da imagem.")
    }
  }

  const handleSaveProfile = async () => {
    setIsSaving(true)
    try {
      const payload = {
        name: userData.fullName,
        display_name: userData.displayName,
        email: userData.email, 
        avatar: userData.avatar,
        ui_settings: {
          ...user?.ui_settings,
          send_shortcut: userData.sendShortcut
        }
      }

      console.log('Sending update:', payload)
      const response = await authApi.updateProfile(payload)
      console.log('Received update:', response)
      
      const updatedUser = response.user || response
      const newToken = response.token
      
      if (newToken) {
          const finalUser = {
              ...updatedUser,
              ui_settings: payload.ui_settings
          }
          login(newToken, finalUser)
      } else {
          const finalUser = {
              ...updatedUser,
              ui_settings: payload.ui_settings
          }
          updateUser(finalUser)
      }

      setIsDirty(false)
      const btn = document.getElementById('save-btn')
      if (btn) btn.innerText = 'Salvo!'
      setTimeout(() => { if(btn) btn.innerText = 'Salvar Alterações' }, 2000)
    } catch (error) {
      console.error('Failed to update profile:', error)
      alert('Erro ao atualizar perfil. Verifique se o e-mail já está em uso.')
    } finally {
      setIsSaving(false)
    }
  }

  const handlePasswordChange = async () => {
    if (passwords.new !== passwords.confirm) {
      alert('As senhas não coincidem!')
      return
    }
    
    setIsSaving(true)
    try {
      await authApi.changePassword(passwords.current, passwords.new)
      alert('Senha alterada com sucesso!')
      setPasswords({ current: '', new: '', confirm: '' })
    } catch (error: any) {
      console.error(error)
      const msg = error.response?.data?.error || 'Erro ao alterar senha. Verifique se a senha atual está correta.'
      alert(msg)
    } finally {
      setIsSaving(false)
    }
  }

  const copyToken = () => {
    navigator.clipboard.writeText(user?.access_token || '')
    setCopiedToken(true)
    setTimeout(() => setCopiedToken(false), 2000)
  }

  if (!user) return <div className="p-8 text-white">Carregando...</div>

  return (
    <div className="h-full bg-gray-900 overflow-y-auto p-8">
      <div className="max-w-4xl mx-auto space-y-8">
        
        {/* Header */}
        <div>
          <h1 className="text-2xl font-bold text-white mb-2">Configurações</h1>
          <p className="text-gray-400">Gerencie suas preferências de conta e segurança</p>
        </div>

        {/* 1. PERFIL */}
        <div className="bg-gray-800 rounded-xl border border-gray-700 overflow-hidden">
          <div className="p-6 border-b border-gray-700 flex items-center gap-3">
            <User className="w-5 h-5 text-primary-500" />
            <h2 className="text-lg font-semibold text-white">Perfil do Usuário</h2>
          </div>
          
          <div className="p-6 space-y-6">
            {/* Avatar */}
            <div className="flex items-center gap-6">
              <div className="relative group">
                <div className="w-24 h-24 bg-gray-700 rounded-full flex items-center justify-center text-3xl font-bold text-gray-400 overflow-hidden border-4 border-gray-800 ring-2 ring-gray-700 group-hover:ring-primary-500 transition-all">
                  {userData.avatar ? (
                    <img src={userData.avatar} alt="Profile" className="w-full h-full object-cover" />
                  ) : (
                    <span>{userData.displayName.charAt(0).toUpperCase()}</span>
                  )}
                </div>
                <input 
                  type="file" 
                  ref={fileInputRef} 
                  className="hidden" 
                  accept="image/png,image/jpeg"
                  onChange={handleFileChange}
                />
                <button 
                  onClick={() => fileInputRef.current?.click()}
                  className="absolute bottom-0 right-0 p-2 bg-primary-600 text-white rounded-full hover:bg-primary-500 transition-colors shadow-lg"
                >
                  <Camera className="w-4 h-4" />
                </button>
              </div>
              <div>
                <h3 className="text-white font-medium">Foto de Perfil</h3>
                <p className="text-sm text-gray-400 mt-1">Carregue uma imagem PNG ou JPG.</p>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-300">Nome Completo</label>
                <div className="relative">
                  <User className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
                  <input
                    type="text"
                    value={userData.fullName}
                    onChange={(e) => handleInputChange('fullName', e.target.value)}
                    className="w-full bg-gray-900 border border-gray-700 rounded-lg pl-10 pr-4 py-2.5 text-white focus:ring-1 focus:ring-primary-500 focus:border-primary-500"
                  />
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-300">Nome para Exibir</label>
                <div className="relative">
                  <User className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
                  <input
                    type="text"
                    value={userData.displayName}
                    onChange={(e) => handleInputChange('displayName', e.target.value)}
                    className="w-full bg-gray-900 border border-gray-700 rounded-lg pl-10 pr-4 py-2.5 text-white focus:ring-1 focus:ring-primary-500 focus:border-primary-500"
                  />
                </div>
              </div>

              <div className="space-y-2 md:col-span-2">
                <label className="text-sm font-medium text-gray-300">Endereço de E-mail</label>
                <div className="relative">
                  <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
                  <input
                    type="email"
                    value={userData.email}
                    onChange={(e) => handleInputChange('email', e.target.value)}
                    className="w-full bg-gray-900 border border-gray-700 rounded-lg pl-10 pr-4 py-2.5 text-white focus:ring-1 focus:ring-primary-500 focus:border-primary-500"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* 2. PREFERÊNCIAS */}
        <div className="bg-gray-800 rounded-xl border border-gray-700 overflow-hidden">
          <div className="p-6 border-b border-gray-700 flex items-center gap-3">
            <Keyboard className="w-5 h-5 text-purple-500" />
            <h2 className="text-lg font-semibold text-white">Preferências</h2>
          </div>
          
          <div className="p-6">
            <label className="text-sm font-medium text-gray-300 mb-4 block">Tecla de atalho para enviar mensagens</label>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <button
                onClick={() => handleInputChange('sendShortcut', 'enter')}
                className={`flex items-center justify-between p-4 rounded-lg border-2 transition-all ${
                  userData.sendShortcut === 'enter'
                    ? 'border-primary-500 bg-primary-900/10'
                    : 'border-gray-700 bg-gray-900 hover:border-gray-600'
                }`}
              >
                <span className="text-white font-medium">Enter</span>
                <span className="text-xs text-gray-400">Envia a mensagem</span>
                {userData.sendShortcut === 'enter' && <Check className="w-5 h-5 text-primary-500" />}
              </button>

              <button
                onClick={() => handleInputChange('sendShortcut', 'ctrl_enter')}
                className={`flex items-center justify-between p-4 rounded-lg border-2 transition-all ${
                  userData.sendShortcut === 'ctrl_enter'
                    ? 'border-primary-500 bg-primary-900/10'
                    : 'border-gray-700 bg-gray-900 hover:border-gray-600'
                }`}
              >
                <span className="text-white font-medium">Ctrl + Enter</span>
                <span className="text-xs text-gray-400">Nova linha com Enter</span>
                {userData.sendShortcut === 'ctrl_enter' && <Check className="w-5 h-5 text-primary-500" />}
              </button>
            </div>
          </div>
        </div>

        {/* 3. SEGURANÇA */}
        <div className="bg-gray-800 rounded-xl border border-gray-700 overflow-hidden">
          <div className="p-6 border-b border-gray-700 flex items-center gap-3">
            <Shield className="w-5 h-5 text-green-500" />
            <h2 className="text-lg font-semibold text-white">Alterar Senha</h2>
          </div>
          
          <div className="p-6 space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-gray-300">Senha Atual</label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
                <input
                  type="password"
                  value={passwords.current}
                  onChange={(e) => setPasswords({...passwords, current: e.target.value})}
                  className="w-full bg-gray-900 border border-gray-700 rounded-lg pl-10 pr-4 py-2.5 text-white focus:ring-1 focus:ring-primary-500 focus:border-primary-500"
                />
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-300">Nova Senha</label>
                <div className="relative">
                  <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
                  <input
                    type="password"
                    value={passwords.new}
                    onChange={(e) => setPasswords({...passwords, new: e.target.value})}
                    className="w-full bg-gray-900 border border-gray-700 rounded-lg pl-10 pr-4 py-2.5 text-white focus:ring-1 focus:ring-primary-500 focus:border-primary-500"
                  />
                </div>
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-300">Confirmar Senha</label>
                <div className="relative">
                  <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
                  <input
                    type="password"
                    value={passwords.confirm}
                    onChange={(e) => setPasswords({...passwords, confirm: e.target.value})}
                    className="w-full bg-gray-900 border border-gray-700 rounded-lg pl-10 pr-4 py-2.5 text-white focus:ring-1 focus:ring-primary-500 focus:border-primary-500"
                  />
                </div>
              </div>
            </div>

            <div className="flex justify-end pt-2">
              <button 
                onClick={handlePasswordChange}
                disabled={!passwords.current || !passwords.new}
                className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-lg font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Atualizar Senha
              </button>
            </div>
          </div>
        </div>

        {/* 4. API TOKEN */}
        <div className="bg-gray-800 rounded-xl border border-gray-700 overflow-hidden">
          <div className="p-6 border-b border-gray-700 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Key className="w-5 h-5 text-yellow-500" />
              <h2 className="text-lg font-semibold text-white">Token de Acesso</h2>
            </div>
            <button 
              onClick={handleCopyAccountId}
              className={`text-sm flex items-center gap-1 font-medium transition-colors ${
                copiedAccountId 
                  ? 'text-green-400' 
                  : 'text-primary-400 hover:text-primary-300'
              }`}
            >
              {copiedAccountId ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
              {copiedAccountId ? 'ID Copiado!' : 'Copiar ID da Conta'}
            </button>
          </div>
          
          <div className="p-6 space-y-4">
            <div className="bg-blue-900/20 border border-blue-800 p-4 rounded-lg flex gap-3">
              <AlertCircle className="w-5 h-5 text-blue-400 flex-shrink-0" />
              <p className="text-sm text-blue-200">
                Este token pode ser usado para integrar outras aplicações via API. Mantenha-o seguro!
              </p>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-gray-300">Seu Token de Acesso</label>
              <div className="flex gap-2">
                <code className="flex-1 bg-gray-900 border border-gray-700 rounded-lg px-4 py-3 text-gray-300 font-mono text-sm break-all">
                  {user?.access_token || 'Token não disponível'}
                </code>
                <button
                  onClick={copyToken}
                  className="flex-shrink-0 px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-lg transition-colors flex items-center gap-2"
                >
                  {copiedToken ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
                  {copiedToken ? 'Copiado!' : 'Copiar'}
                </button>
              </div>
              <p className="text-xs text-gray-500 mt-1">
                Token para autenticação Bearer em requisições API.
              </p>
            </div>
          </div>
        </div>

        {/* Floating Save Button */}
        {isDirty && (
          <div className="fixed bottom-8 right-8 animate-fade-in-up">
            <button
              id="save-btn"
              onClick={handleSaveProfile}
              disabled={isSaving}
              className="flex items-center gap-2 px-6 py-3 bg-primary-600 hover:bg-primary-500 text-white rounded-full font-bold shadow-lg shadow-primary-900/30 transform hover:scale-105 transition-all disabled:opacity-70 disabled:cursor-wait"
            >
              {isSaving ? <RefreshCw className="w-5 h-5 animate-spin" /> : <Save className="w-5 h-5" />}
              {isSaving ? 'Salvando...' : 'Salvar Alterações'}
            </button>
          </div>
        )}

      </div>
    </div>
  )
}
