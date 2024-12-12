import { useState, useEffect, useRef } from 'react'
import { User, Circle, MessageCircle, X } from 'lucide-react'
import { ChatWindow } from './ChatWindow'

interface ChatUser {
  id: number
  username: string
  avatar?: string
  online: boolean
  typing?: boolean
}

export function ChatList() {
  const [isOpen, setIsOpen] = useState(false)
  const [users, setUsers] = useState<ChatUser[]>([])
  const [activeChats, setActiveChats] = useState<ChatUser[]>([])
  const ws = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>()

  const fetchUsers = async () => {
    try {
      const response = await fetch('http://localhost:8080/chat/users', {
        credentials: 'include'
      })
      if (response.ok) {
        const data = await response.json()
        setUsers(Array.isArray(data) ? data : [])
      }
    } catch (error) {
      console.error('Error fetching users:', error)
      setUsers([])
    }
  }

  const connectWebSocket = () => {
    if (!ws.current || ws.current.readyState === WebSocket.CLOSED) {
      try {
        ws.current = new WebSocket('ws://localhost:8080/ws/chat')

        ws.current.onopen = () => {
          console.log('Chat WebSocket connected')
          fetchUsers()
        }

        ws.current.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data)
            console.log('Chat WebSocket message received:', data)
            
            if (data.type === 'chat') {
              // Handle chat message
              console.log('Received chat message:', data)
            } else if (data.type === 'user_status') {
              updateUserStatus(data.user_id, data.is_online)
            } else if (data.type === 'typing') {
              // Handle typing status
              const { sender_id, is_typing } = data
              setUsers(prev => prev.map(user => 
                user.id === sender_id ? { ...user, typing: is_typing } : user
              ))
            }
          } catch (error) {
            console.error('Error parsing chat WebSocket message:', error)
          }
        }

        ws.current.onerror = (error) => {
          console.error('Chat WebSocket error:', error)
          // Try to reconnect on error
          if (reconnectTimeoutRef.current) {
            clearTimeout(reconnectTimeoutRef.current)
          }
          reconnectTimeoutRef.current = setTimeout(connectWebSocket, 3000)
        }

        ws.current.onclose = (event) => {
          console.log('Chat WebSocket closed:', event.code, event.reason)
          // Only attempt to reconnect if it wasn't a normal closure
          if (event.code !== 1000) {
            if (reconnectTimeoutRef.current) {
              clearTimeout(reconnectTimeoutRef.current)
            }
            reconnectTimeoutRef.current = setTimeout(connectWebSocket, 3000)
          }
        }
      } catch (error) {
        console.error('Error creating WebSocket connection:', error)
        // Try to reconnect after error
        if (reconnectTimeoutRef.current) {
          clearTimeout(reconnectTimeoutRef.current)
        }
        reconnectTimeoutRef.current = setTimeout(connectWebSocket, 3000)
      }
    }
  }

  useEffect(() => {
    fetchUsers()
    connectWebSocket()

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
      if (ws.current) {
        const socket = ws.current
        socket.onclose = null // Prevent reconnection attempt
        socket.close(1000, 'Component unmounting')
      }
    }
  }, [])

  const updateUserStatus = (userId: number, online: boolean) => {
    setUsers(prev => prev.map(user => 
      user.id === userId ? { ...user, online } : user
    ))
    setActiveChats(prev => prev.map(chat =>
      chat.id === userId ? { ...chat, online } : chat
    ))
  }

  const startChat = (user: ChatUser) => {
    if (!activeChats.find(chat => chat.id === user.id)) {
      setActiveChats(prev => [...prev, user])
    }
    setIsOpen(false)
  }

  const closeChat = (userId: number) => {
    setActiveChats(prev => prev.filter(chat => chat.id !== userId))
  }

  return (
    <>
      <div className="fixed bottom-0 right-0 w-72">
        {/* Chat List Header */}
        <button
          onClick={() => setIsOpen(!isOpen)}
          className="w-full bg-white/10 backdrop-blur-lg p-3 rounded-t-lg border border-gray-700/50 flex items-center justify-between"
        >
          <div className="flex items-center space-x-2">
            <MessageCircle size={20} className="text-gray-200" />
            <span className="font-medium text-gray-200">Chats</span>
          </div>
          <span className="text-gray-400">
            {users.filter(user => user.online).length} online
          </span>
        </button>

        {/* Chat List Body */}
        {isOpen && (
          <div className="max-h-96 overflow-y-auto bg-white/10 backdrop-blur-lg border-x border-b border-gray-700/50 rounded-b-lg">
            {users.length > 0 ? (
              users.map(user => (
                <button
                  key={user.id}
                  onClick={() => startChat(user)}
                  className="w-full p-3 flex items-center space-x-3 hover:bg-gray-800/50 transition-colors"
                >
                  <div className="relative">
                    {user.avatar ? (
                      <img 
                        src={user.avatar} 
                        alt={user.username} 
                        className="w-10 h-10 rounded-full"
                      />
                    ) : (
                      <div className="w-10 h-10 bg-gray-600 rounded-full flex items-center justify-center">
                        <User size={20} className="text-gray-300" />
                      </div>
                    )}
                    {user.online && (
                      <Circle size={8} className="absolute bottom-0 right-0 text-green-500 fill-current" />
                    )}
                  </div>
                  <div className="flex-1 text-left">
                    <p className="text-gray-200">{user.username}</p>
                    {user.typing && (
                      <p className="text-sm text-gray-400">Typing...</p>
                    )}
                  </div>
                </button>
              ))
            ) : (
              <div className="p-4 text-center text-gray-400">
                No users available
              </div>
            )}
          </div>
        )}
      </div>

      {/* Active Chat Windows */}
      {activeChats.map((user, index) => (
        <ChatWindow
          key={user.id}
          user={user}
          websocket={ws.current}
          onClose={() => closeChat(user.id)}
          style={{
            right: `${(index * 320) + 288}px`
          }}
        />
      ))}
    </>
  )
}