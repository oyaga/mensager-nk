import React, { useState, useRef, useEffect } from 'react'
import { 
  Send, Paperclip, Mic, Image, FileText, X, 
  Play, Pause, Download, File, Volume2
} from 'lucide-react'
import { messagesApi, storageApi } from '../lib/api'

interface Attachment {
  id: string
  file_type: 'image' | 'audio' | 'video' | 'file'
  file_url: string
  file_name: string
  file_size?: number
}

interface Message {
  id: string
  content: string
  content_type: string
  message_type: 'incoming' | 'outgoing' | 'activity'
  created_at: string
  sender?: { id: string; name: string; avatar?: string }
  attachments?: Attachment[]
}

interface ChatPanelProps {
  conversationId: string
  contactName: string
}

export default function ChatPanel({ conversationId, contactName }: ChatPanelProps) {
  const [messages, setMessages] = useState<Message[]>([])
  const [newMessage, setNewMessage] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [isSending, setIsSending] = useState(false)
  const [pendingFiles, setPendingFiles] = useState<File[]>([])
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  // Fetch messages
  useEffect(() => {
    if (conversationId) {
      fetchMessages()
    }
  }, [conversationId])

  // Auto scroll to bottom
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const fetchMessages = async () => {
    setIsLoading(true)
    try {
      const data = await messagesApi.list(conversationId)
      setMessages(Array.isArray(data) ? data : data.messages || [])
    } catch (error) {
      console.error('Failed to fetch messages:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const handleSend = async () => {
    if (!newMessage.trim() && pendingFiles.length === 0) return

    setIsSending(true)
    try {
      // Upload files first if any
      const attachmentUrls: { file_type: string; file_url: string; file_name: string }[] = []
      
      for (const file of pendingFiles) {
        const { url } = await storageApi.upload(file)
        const fileType = getFileType(file.type)
        attachmentUrls.push({
          file_type: fileType,
          file_url: url,
          file_name: file.name
        })
      }

      // Send message
      const messageData = {
        conversation_id: conversationId,
        content: newMessage,
        content_type: attachmentUrls.length > 0 ? attachmentUrls[0].file_type : 'text',
        message_type: 'outgoing',
        attachments: attachmentUrls
      }

      await messagesApi.create(messageData)
      
      // Clear and refresh
      setNewMessage('')
      setPendingFiles([])
      fetchMessages()
    } catch (error) {
      console.error('Failed to send message:', error)
      alert('Erro ao enviar mensagem')
    } finally {
      setIsSending(false)
    }
  }

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || [])
    setPendingFiles(prev => [...prev, ...files])
  }

  const removePendingFile = (index: number) => {
    setPendingFiles(prev => prev.filter((_, i) => i !== index))
  }

  const getFileType = (mimeType: string): 'image' | 'audio' | 'video' | 'file' => {
    if (mimeType.startsWith('image/')) return 'image'
    if (mimeType.startsWith('audio/')) return 'audio'
    if (mimeType.startsWith('video/')) return 'video'
    return 'file'
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  return (
    <div className="flex flex-col h-full bg-gray-900">
      {/* Messages Area */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4 custom-scrollbar">
        {isLoading ? (
          <div className="flex items-center justify-center h-full text-gray-500">
            <div className="w-6 h-6 border-2 border-primary-500 border-t-transparent rounded-full animate-spin"></div>
          </div>
        ) : messages.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-gray-500">
            <p>Nenhuma mensagem ainda</p>
            <p className="text-sm">Envie a primeira mensagem para {contactName}</p>
          </div>
        ) : (
          messages.map((msg) => (
            <MessageBubble key={msg.id} message={msg} />
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Pending Files Preview */}
      {pendingFiles.length > 0 && (
        <div className="px-4 py-2 bg-gray-800 border-t border-gray-700 flex gap-2 overflow-x-auto">
          {pendingFiles.map((file, index) => (
            <div key={index} className="relative flex-shrink-0 w-20 h-20 bg-gray-700 rounded-lg overflow-hidden group">
              {file.type.startsWith('image/') ? (
                <img 
                  src={URL.createObjectURL(file)} 
                  alt={file.name} 
                  className="w-full h-full object-cover"
                />
              ) : (
                <div className="w-full h-full flex flex-col items-center justify-center text-gray-400">
                  <File className="w-6 h-6" />
                  <span className="text-[10px] truncate w-full text-center px-1">{file.name}</span>
                </div>
              )}
              <button
                onClick={() => removePendingFile(index)}
                className="absolute top-1 right-1 p-0.5 bg-red-600 rounded-full opacity-0 group-hover:opacity-100 transition-opacity"
              >
                <X className="w-3 h-3 text-white" />
              </button>
            </div>
          ))}
        </div>
      )}

      {/* Input Area */}
      <div className="p-4 bg-gray-800 border-t border-gray-700">
        <div className="flex items-end gap-2">
          {/* Attachment Button */}
          <input
            type="file"
            ref={fileInputRef}
            className="hidden"
            multiple
            accept="image/*,audio/*,video/*,.pdf,.doc,.docx,.xls,.xlsx,.txt"
            onChange={handleFileSelect}
          />
          <button
            onClick={() => fileInputRef.current?.click()}
            className="p-3 text-gray-400 hover:text-white hover:bg-gray-700 rounded-lg transition-colors"
            title="Anexar arquivo"
          >
            <Paperclip className="w-5 h-5" />
          </button>

          {/* Text Input */}
          <div className="flex-1 relative">
            <textarea
              value={newMessage}
              onChange={(e) => setNewMessage(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Digite sua mensagem..."
              rows={1}
              className="w-full bg-gray-700 border border-gray-600 rounded-lg px-4 py-3 text-white placeholder-gray-400 focus:outline-none focus:border-primary-500 resize-none max-h-32"
              style={{ minHeight: '48px' }}
            />
          </div>

          {/* Send Button */}
          <button
            onClick={handleSend}
            disabled={isSending || (!newMessage.trim() && pendingFiles.length === 0)}
            className="p-3 bg-primary-600 hover:bg-primary-500 disabled:bg-gray-700 disabled:text-gray-500 text-white rounded-lg transition-colors"
          >
            {isSending ? (
              <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
            ) : (
              <Send className="w-5 h-5" />
            )}
          </button>
        </div>
      </div>
    </div>
  )
}

// Message Bubble Component
function MessageBubble({ message }: { message: Message }) {
  const isOutgoing = message.message_type === 'outgoing'
  const isActivity = message.message_type === 'activity'

  if (isActivity) {
    return (
      <div className="flex justify-center">
        <span className="px-3 py-1 bg-gray-700 rounded-full text-xs text-gray-400">
          {message.content}
        </span>
      </div>
    )
  }

  return (
    <div className={`flex ${isOutgoing ? 'justify-end' : 'justify-start'}`}>
      <div 
        className={`max-w-[70%] rounded-2xl px-4 py-2 ${
          isOutgoing 
            ? 'bg-primary-600 text-white rounded-br-md' 
            : 'bg-gray-700 text-white rounded-bl-md'
        }`}
      >
        {/* Attachments */}
        {message.attachments && message.attachments.length > 0 && (
          <div className="mb-2 space-y-2">
            {message.attachments.map((att, index) => (
              <AttachmentRenderer key={index} attachment={att} />
            ))}
          </div>
        )}

        {/* Text Content */}
        {message.content && (
          <p className="whitespace-pre-wrap break-words">{message.content}</p>
        )}

        {/* Timestamp */}
        <div className={`text-[10px] mt-1 ${isOutgoing ? 'text-primary-200' : 'text-gray-400'}`}>
          {new Date(message.created_at).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' })}
        </div>
      </div>
    </div>
  )
}

// Attachment Renderer
function AttachmentRenderer({ attachment }: { attachment: Attachment }) {
  const [isPlaying, setIsPlaying] = useState(false)
  const audioRef = useRef<HTMLAudioElement>(null)

  const toggleAudio = () => {
    if (audioRef.current) {
      if (isPlaying) {
        audioRef.current.pause()
      } else {
        audioRef.current.play()
      }
      setIsPlaying(!isPlaying)
    }
  }

  switch (attachment.file_type) {
    case 'image':
      return (
        <div className="rounded-lg overflow-hidden">
          <img 
            src={attachment.file_url} 
            alt={attachment.file_name}
            className="max-w-full max-h-64 object-contain cursor-pointer hover:opacity-90"
            onClick={() => window.open(attachment.file_url, '_blank')}
          />
        </div>
      )

    case 'audio':
      return (
        <div className="flex items-center gap-2 bg-black/20 rounded-lg p-2 min-w-[200px]">
          <button
            onClick={toggleAudio}
            className="p-2 bg-white/10 rounded-full hover:bg-white/20 transition-colors"
          >
            {isPlaying ? <Pause className="w-4 h-4" /> : <Play className="w-4 h-4" />}
          </button>
          <div className="flex-1">
            <div className="h-1 bg-white/20 rounded-full">
              <div className="h-full w-0 bg-white/60 rounded-full" />
            </div>
          </div>
          <Volume2 className="w-4 h-4 opacity-50" />
          <audio 
            ref={audioRef} 
            src={attachment.file_url}
            onEnded={() => setIsPlaying(false)}
          />
        </div>
      )

    case 'video':
      return (
        <div className="rounded-lg overflow-hidden">
          <video 
            src={attachment.file_url}
            controls
            className="max-w-full max-h-64"
          />
        </div>
      )

    case 'file':
    default:
      return (
        <a 
          href={attachment.file_url}
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center gap-2 bg-black/20 rounded-lg p-3 hover:bg-black/30 transition-colors"
        >
          <FileText className="w-8 h-8 text-primary-300" />
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium truncate">{attachment.file_name}</p>
            <p className="text-xs opacity-70">Documento</p>
          </div>
          <Download className="w-4 h-4 opacity-70" />
        </a>
      )
  }
}
