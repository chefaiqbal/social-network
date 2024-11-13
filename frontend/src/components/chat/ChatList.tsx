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

  useEffect(() => {
    // Create a single WebSocket connection
    if (!ws.current || ws.current.readyState === WebSocket.CLOSED) {
      ws.current = new WebSocket('ws://localhost:8080/ws')

      ws.current.onopen = () => {
        console.log('WebSocket connected');
        fetchUsers()
      }

      ws.current.onmessage = (event) => {
        const data = JSON.parse(event.data)
        if (data.type === 'user_status') {
          updateUserStatus(data.user_id, data.is_online)
        }
      }

      ws.current.onerror = (error) => {
        console.error('WebSocket error:', error)
      }

      ws.current.onclose = () => {
        console.log('WebSocket closed')
        // Attempt to reconnect after a delay
        setTimeout(() => {
          if (!ws.current || ws.current.readyState === WebSocket.CLOSED) {
            console.log('Attempting to reconnect...')
            ws.current = new WebSocket('ws://localhost:8080/ws')
          }
        }, 3000)
      }
    }

    return () => {
      // Don't close the WebSocket on component unmount
      // as it's shared across chat windows
    }
  }, [])

  const fetchUsers = async () => {
    try {
      const response = await fetch('http://localhost:8080/chat/users', {
        credentials: 'include'
      })
      if (response.ok) {
        const data = await response.json()
        if (Array.isArray(data)) {
          setUsers(data)
        } else {
          setUsers([])
        }
      }
    } catch (error) {
      console.error('Error fetching users:', error)
      setUsers([])
    }
  }

  const updateUserStatus = (userId: number, online: boolean) => {
    setUsers(prev => prev.map(user => 
      user.id === userId ? { ...user, online } : user
    ))
    // Also update active chats
    setActiveChats(prev => prev.map(chat =>
      chat.id === userId ? { ...chat, online } : chat
    ))
  }

  const startChat = (user: ChatUser) => {
    if (!activeChats.find(chat => chat.id === user.id)) {
      setActiveChats(prev => [...prev, user])
    }
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
            {Array.isArray(users) ? users.filter(user => user.online).length : 0} online
          </span>
        </button>

        {/* Chat List Body */}
        {isOpen && Array.isArray(users) && users.length > 0 && (
          <div className="max-h-96 overflow-y-auto bg-white/10 backdrop-blur-lg border-x border-b border-gray-700/50 rounded-b-lg">
            {users.map(user => (
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
            ))}
          </div>
        )}
      </div>

      {/* Active Chat Windows */}
      {Array.isArray(activeChats) && activeChats.map((user, index) => (
        <ChatWindow
          key={user.id}
          user={user}
          websocket={ws.current}
          onClose={() => closeChat(user.id)}
        />
      ))}
    </>
  )
} 