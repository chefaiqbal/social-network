import { useState, useEffect, useRef, useCallback } from 'react'
import { User, Send, X, Circle } from 'lucide-react'
import throttle from 'lodash/throttle'

interface ChatMessage {
  id?: number
  sender_id: number
  recipient_id: number
  content: string
  created_at: string
}

interface ChatUser {
  id: number
  username: string
  avatar?: string
  online: boolean
  typing?: boolean
}

interface ChatWindowProps {
  user: ChatUser
  onClose: () => void
}

export function ChatWindow({ user, onClose }: ChatWindowProps) {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [newMessage, setNewMessage] = useState('')
  const [isTyping, setIsTyping] = useState(false)
  const [page, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(true)
  const [isLoading, setIsLoading] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const messageContainerRef = useRef<HTMLDivElement>(null)
  const ws = useRef<WebSocket | null>(null)
  const typingTimeoutRef = useRef<NodeJS.Timeout>()

  const scrollToBottom = (behavior: 'auto' | 'smooth' = 'auto') => {
    messagesEndRef.current?.scrollIntoView({ behavior })
  }

  const fetchMessages = async (pageNum: number, isInitial: boolean = false) => {
    if (isLoading || (!hasMore && !isInitial)) return

    try {
      setIsLoading(true)
      const response = await fetch(
        `http://localhost:8080/messages/${user.id}?page=${pageNum}&limit=8`,
        {
          credentials: 'include'
        }
      )
      if (response.ok) {
        const data = await response.json()
        if (Array.isArray(data)) {
          if (data.length < 8) {
            setHasMore(false)
          }
          setMessages(prev => 
            isInitial ? data : [...data, ...prev]
          )
          if (isInitial) {
            scrollToBottom()
          }
        }
      }
    } catch (error) {
      console.error('Error fetching messages:', error)
    } finally {
      setIsLoading(false)
    }
  }

  // Throttled scroll handler
  const handleScroll = useCallback(
    throttle(() => {
      const container = messageContainerRef.current
      if (container && container.scrollTop === 0 && hasMore && !isLoading) {
        setPage(prev => {
          const newPage = prev + 1
          fetchMessages(newPage)
          return newPage
        })
      }
    }, 500),
    [hasMore, isLoading]
  )

  useEffect(() => {
    // Connect to WebSocket
    ws.current = new WebSocket('ws://localhost:8080/ws')

    ws.current.onopen = () => {
      console.log('WebSocket connected')
    }

    ws.current.onerror = (error) => {
      console.error('WebSocket error:', error)
    }

    ws.current.onmessage = (event) => {
      console.log('Received message:', event.data)
      
      const data = JSON.parse(event.data)
      if (data.type === 'chat') {
        if (data.sender_id === user.id || data.recipient_id === user.id) {
          setMessages(prev => [...prev, data])
          scrollToBottom('smooth')
        }
      } else if (data.type === 'typing' && data.sender_id === user.id) {
        setIsTyping(data.typing)
      }
    }

    // Initial fetch
    fetchMessages(1, true)

    // Cleanup
    return () => {
      if (ws.current) {
        ws.current.close()
      }
      handleScroll.cancel()
    }
  }, [user.id])

  useEffect(() => {
    const container = messageContainerRef.current
    if (container) {
      container.addEventListener('scroll', handleScroll)
      return () => container.removeEventListener('scroll', handleScroll)
    }
  }, [handleScroll])

  const handleTyping = () => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({
        type: 'typing',
        recipient_id: user.id,
        typing: true
      }))

      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current)
      }

      typingTimeoutRef.current = setTimeout(() => {
        ws.current?.send(JSON.stringify({
          type: 'typing',
          recipient_id: user.id,
          typing: false
        }))
      }, 2000)
    }
  }

  const sendMessage = () => {
    if (newMessage.trim() && ws.current?.readyState === WebSocket.OPEN) {
      const messageData = {
        type: 'chat',
        recipient_id: user.id,
        content: newMessage.trim()
      }
      
      ws.current.send(JSON.stringify(messageData))
      setNewMessage('')
    }
  }

  return (
    <div className="fixed bottom-0 right-96 w-80 bg-white/10 backdrop-blur-lg rounded-t-lg shadow-lg border border-gray-700/50">
      {/* Chat Header */}
      <div className="flex items-center justify-between p-3 border-b border-gray-700/50">
        <div className="flex items-center space-x-2">
          <div className="relative">
            {user.avatar ? (
              <img 
                src={user.avatar} 
                alt={user.username} 
                className="w-8 h-8 rounded-full"
              />
            ) : (
              <div className="w-8 h-8 bg-gray-600 rounded-full flex items-center justify-center">
                <User size={16} className="text-gray-300" />
              </div>
            )}
            {user.online && (
              <Circle size={8} className="absolute bottom-0 right-0 text-green-500 fill-current" />
            )}
          </div>
          <span className="font-medium text-gray-200">{user.username}</span>
        </div>
        <button 
          onClick={onClose}
          className="text-gray-400 hover:text-gray-200 transition-colors"
        >
          <X size={20} />
        </button>
      </div>

      {/* Messages */}
      <div 
        ref={messageContainerRef}
        className="h-96 overflow-y-auto p-4 space-y-4"
      >
        {isLoading && page > 1 && (
          <div className="text-center text-gray-400 text-sm">
            Loading older messages...
          </div>
        )}
        {Array.isArray(messages) && messages.map((message, index) => (
          <div
            key={message.id || index}
            className={`flex ${message.sender_id === user.id ? 'justify-start' : 'justify-end'}`}
          >
            <div
              className={`max-w-[70%] rounded-lg px-4 py-2 ${
                message.sender_id === user.id 
                  ? 'bg-gray-700/50 text-gray-200' 
                  : 'bg-blue-500 text-white'
              }`}
            >
              <p>{message.content}</p>
              <span className="text-xs opacity-70">
                {new Date(message.created_at).toLocaleTimeString()}
              </span>
            </div>
          </div>
        ))}
        {isTyping && (
          <div className="flex justify-start items-center space-x-2">
            <div className="bg-gray-700/50 rounded-lg px-4 py-2 text-gray-200">
              <div className="flex space-x-1">
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }}></div>
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }}></div>
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }}></div>
              </div>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="p-3 border-t border-gray-700/50">
        <div className="flex items-center space-x-2">
          <input
            type="text"
            value={newMessage}
            onChange={(e) => setNewMessage(e.target.value)}
            onKeyUp={handleTyping}
            onKeyPress={(e) => e.key === 'Enter' && sendMessage()}
            placeholder="Type a message..."
            className="flex-1 bg-gray-800/50 border border-gray-700/50 rounded-lg px-4 py-2 text-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <button
            onClick={sendMessage}
            disabled={!newMessage.trim()}
            className="bg-blue-500 hover:bg-blue-600 disabled:bg-blue-500/50 disabled:cursor-not-allowed text-white rounded-lg p-2 transition-colors"
          >
            <Send size={20} />
          </button>
        </div>
      </div>
    </div>
  )
} 