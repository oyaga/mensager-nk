import { useEffect, useRef, useCallback } from 'react'
import { useAuthStore } from '../stores/authStore'
import { useNotificationStore } from '../stores/notificationStore'

// Build WebSocket URL dynamically based on current origin
const getWebSocketUrl = () => {
  if (import.meta.env.VITE_WS_URL) {
    return import.meta.env.VITE_WS_URL
  }
  // Use current origin for unified deployment
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${protocol}//${window.location.host}/cable`
}

interface WebSocketMessage {
  type: string
  payload: any
  room?: string
}

export function useWebSocket() {
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>()
  const { token, isAuthenticated } = useAuthStore()
  const { addNotification } = useNotificationStore()

  const connect = useCallback(() => {
    if (!isAuthenticated || !token) return

    // Close existing connection
    if (wsRef.current) {
      wsRef.current.close()
    }

    const ws = new WebSocket(`${getWebSocketUrl()}?token=${token}`)

    ws.onopen = () => {
      console.log('🔌 WebSocket connected')
      
      // Subscribe to user's notification channel
      ws.send(JSON.stringify({
        type: 'subscribe',
        payload: 'notifications'
      }))
    }

    ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data)
        handleMessage(message)
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err)
      }
    }

    ws.onclose = () => {
      console.log('🔌 WebSocket disconnected')
      // Reconnect after 3 seconds
      reconnectTimeoutRef.current = setTimeout(() => {
        if (isAuthenticated) {
          connect()
        }
      }, 3000)
    }

    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
    }

    wsRef.current = ws
  }, [isAuthenticated, token])

  const handleMessage = useCallback((message: WebSocketMessage) => {
    switch (message.type) {
      case 'message.created':
        // New message received
        const msgPayload = message.payload
        addNotification({
          type: 'message',
          title: 'Nova mensagem',
          body: msgPayload.content?.substring(0, 100) || 'Nova mensagem recebida',
          conversationId: msgPayload.conversation_id,
        })
        break

      case 'conversation.created':
        addNotification({
          type: 'conversation',
          title: 'Nova conversa',
          body: 'Uma nova conversa foi iniciada',
          conversationId: message.payload.id,
        })
        break

      case 'conversation.updated':
        // Conversation status changed (e.g., replied)
        if (message.payload.status === 'open') {
          addNotification({
            type: 'conversation',
            title: 'Conversa atualizada',
            body: 'Uma conversa foi reaberta ou respondida',
            conversationId: message.payload.id,
          })
        }
        break

      default:
        console.log('Unknown message type:', message.type)
    }
  }, [addNotification])

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
    }
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
  }, [])

  const send = useCallback((message: WebSocketMessage) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message))
    }
  }, [])

  const subscribeToConversation = useCallback((conversationId: string) => {
    send({
      type: 'subscribe',
      payload: conversationId
    })
  }, [send])

  const unsubscribeFromConversation = useCallback((conversationId: string) => {
    send({
      type: 'unsubscribe',
      payload: conversationId
    })
  }, [send])

  useEffect(() => {
    if (isAuthenticated) {
      connect()
    } else {
      disconnect()
    }

    return () => {
      disconnect()
    }
  }, [isAuthenticated, connect, disconnect])

  return {
    send,
    subscribeToConversation,
    unsubscribeFromConversation,
  }
}
