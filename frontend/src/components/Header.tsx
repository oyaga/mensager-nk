import { Bell, Search, X } from 'lucide-react'
import { useAuthStore } from '../stores/authStore'
import { useNotificationStore } from '../stores/notificationStore'
import { useWebSocket } from '../hooks/useWebSocket'
import clsx from 'clsx'
import { useState, useRef, useEffect } from 'react'
import { NavLink } from 'react-router-dom'

export default function Header() {
  const { user } = useAuthStore()
  const { unreadCount, notifications, markAllAsRead } = useNotificationStore()
  const [showNotifications, setShowNotifications] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  // Initialize WebSocket connection
  useWebSocket()
  
  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setShowNotifications(false)
      }
    }

    if (showNotifications) {
      document.addEventListener('mousedown', handleClickOutside)
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [showNotifications])

  return (
    <header className="h-16 bg-gray-800 border-b border-gray-700 flex items-center justify-between px-6">
      {/* Search */}
      <div className="flex-1 max-w-2xl">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
          <input
            type="search"
            placeholder="Buscar conversas..."
            className="w-full pl-10 pr-4 py-2 bg-gray-700 border border-gray-600 text-gray-100 placeholder-gray-400 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
          />
        </div>
      </div>

      {/* Right Section */}
      <div className="flex items-center gap-4">
        {/* Notifications */}
        <div className="relative" ref={dropdownRef}>
          <button
            onClick={() => setShowNotifications(!showNotifications)}
            className="relative p-2 text-gray-300 hover:text-white hover:bg-gray-700 rounded-lg transition-colors"
            title="Notificações"
            type="button"
          >
            <Bell className="w-5 h-5" />
            {unreadCount > 0 && (
              <span className="absolute -top-1 -right-1 min-w-[18px] h-[18px] bg-red-500 text-white text-[10px] font-bold rounded-full flex items-center justify-center px-1 animate-pulse border-2 border-gray-800">
                {unreadCount > 99 ? '99+' : unreadCount}
              </span>
            )}
          </button>

          {/* Notification Dropdown */}
          {showNotifications && (
            <div className="absolute right-0 top-full mt-2 w-80 bg-gray-900 border border-gray-700 rounded-xl shadow-2xl z-[100]">
              <div className="flex items-center justify-between p-4 border-b border-gray-700">
                <h3 className="text-white font-semibold">Notificações</h3>
                <div className="flex items-center gap-2">
                  {unreadCount > 0 && (
                    <button
                      onClick={markAllAsRead}
                      className="text-xs text-primary-400 hover:text-primary-300"
                      type="button"
                    >
                      Marcar lidas
                    </button>
                  )}
                  <button
                    onClick={() => setShowNotifications(false)}
                    className="text-gray-400 hover:text-white"
                    type="button"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              </div>
              <div className="max-h-80 overflow-y-auto">
                {notifications.length === 0 ? (
                  <div className="p-8 text-center text-gray-500">
                    <Bell className="w-8 h-8 mx-auto mb-2 opacity-50" />
                    <p className="text-sm">Nenhuma notificação</p>
                  </div>
                ) : (
                  notifications.slice(0, 10).map((notification) => (
                    <NavLink
                      key={notification.id}
                      to={`/conversations/${notification.conversationId}`}
                      className={clsx(
                        'block p-3 border-b border-gray-800 hover:bg-gray-800 cursor-pointer transition-colors',
                        !notification.read && 'bg-gray-800/50'
                      )}
                      onClick={() => setShowNotifications(false)}
                    >
                      <div className="flex items-start gap-3">
                        <div className={clsx(
                          'w-2 h-2 rounded-full mt-2 flex-shrink-0',
                          notification.read ? 'bg-gray-600' : 'bg-primary-500'
                        )} />
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium text-white">
                            {notification.title}
                          </p>
                          <p className="text-xs text-gray-400 truncate">
                            {notification.body}
                          </p>
                          <p className="text-xs text-gray-500 mt-1">
                            {new Date(notification.createdAt).toLocaleTimeString('pt-BR', {
                              hour: '2-digit',
                              minute: '2-digit'
                            })}
                          </p>
                        </div>
                      </div>
                    </NavLink>
                  ))
                )}
              </div>
            </div>
          )}
        </div>

        {/* Availability Status */}
        <div className="flex items-center gap-2 px-3 py-1.5 bg-green-900/30 text-green-400 rounded-lg text-sm font-medium border border-green-800">
          <span className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
          {user?.availability || 'Online'}
        </div>
      </div>
    </header>
  )
}
