import { useState, useEffect, useRef, useCallback } from 'react'
import { User, Send, X, Circle, Smile } from 'lucide-react'
import throttle from 'lodash/throttle'
import data from '@emoji-mart/data'
import Picker from '@emoji-mart/react'
import styles from '../../styles/emoji.module.css'

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
  websocket: WebSocket | null
  onClose: () => void
}

interface EmojiData {
  id: string
  name: string
  native: string
  unified: string
  keywords: string[]
  shortcodes: string
}

interface EmojiSelectData {
  id: string;
  name: string;
  native: string;
  unified: string;
  keywords: string[];
  shortcodes: string;
}

export function ChatWindow({ user, websocket, onClose }: ChatWindowProps) {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [newMessage, setNewMessage] = useState('')
  const [isTyping, setIsTyping] = useState(false)
  const [page, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(true)
  const [isLoading, setIsLoading] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const messageContainerRef = useRef<HTMLDivElement>(null)
  const typingTimeoutRef = useRef<NodeJS.Timeout>()
  const [showEmojiPicker, setShowEmojiPicker] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

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
    if (websocket) {
      const messageHandler = (event: MessageEvent) => {
        try {
          const data = JSON.parse(event.data)
          if (data.type === 'chat') {
            if ((data.sender_id === user.id && data.recipient_id !== user.id) || 
                (data.recipient_id === user.id && data.sender_id !== user.id)) {
              setMessages(prev => [...prev, data])
              scrollToBottom('smooth')
            }
          } else if (data.type === 'typing' && data.sender_id === user.id) {
            setIsTyping(data.typing)
          }
        } catch (error) {
          console.error('Error handling WebSocket message:', error)
        }
      }

      websocket.addEventListener('message', messageHandler)

      // Initial fetch
      fetchMessages(1, true)

      return () => {
        websocket.removeEventListener('message', messageHandler)
      }
    }
  }, [user.id, websocket])

  useEffect(() => {
    const container = messageContainerRef.current
    if (container) {
      container.addEventListener('scroll', handleScroll)
      return () => container.removeEventListener('scroll', handleScroll)
    }
  }, [handleScroll])

  const handleTyping = () => {
    if (websocket?.readyState === WebSocket.OPEN) {
      websocket.send(JSON.stringify({
        type: 'typing',
        recipient_id: user.id,
        typing: true
      }))

      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current)
      }

      typingTimeoutRef.current = setTimeout(() => {
        websocket.send(JSON.stringify({
          type: 'typing',
          recipient_id: user.id,
          typing: false
        }))
      }, 2000)
    }
  }

  const sendMessage = () => {
    if (newMessage.trim() && websocket?.readyState === WebSocket.OPEN) {
      const messageData = {
        type: 'chat',
        recipient_id: user.id,
        content: newMessage.trim()
      }
      
      try {
        websocket.send(JSON.stringify(messageData))
        setNewMessage('')
      } catch (error) {
        console.error('Error sending message:', error)
      }
    }
  }

  const handleEmojiSelect = (emojiData: EmojiData) => {
    try {
      console.log('Emoji data received:', emojiData);

      const emoji = emojiData.native;
      console.log('Selected emoji:', emoji);

      if (inputRef.current) {
        const input = inputRef.current
        const start = input.selectionStart || 0
        const end = input.selectionEnd || 0
        const newText = newMessage.slice(0, start) + emoji + newMessage.slice(end)
        setNewMessage(newText)

        // Update cursor position
        setTimeout(() => {
          input.focus()
          const cursorPosition = start + emoji.length
          input.setSelectionRange(cursorPosition, cursorPosition)
        }, 0)
      }

      setShowEmojiPicker(false)
    } catch (error) {
      console.error('Error handling emoji selection:', error)
    }
  }

  // Add this effect to handle clicking outside the emoji picker
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (showEmojiPicker && !(event.target as Element).closest('.emoji-mart')) {
        setShowEmojiPicker(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [showEmojiPicker])

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
            className={`flex ${message.sender_id !== user.id ? 'justify-start' : 'justify-end'}`}
          >
            <div
              className={`max-w-[70%] rounded-lg px-4 py-2 ${styles.chatMessage} ${
                message.sender_id !== user.id 
                  ? 'bg-gray-700/50 text-gray-200' 
                  : 'bg-blue-500 text-white'
              }`}
            >
              <p style={{ wordBreak: 'break-word' }}>{message.content}</p>
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
        <div className="relative">
          {showEmojiPicker && (
            <div 
              className="absolute bottom-[100%] right-0 mb-2 z-50"
              onClick={e => e.stopPropagation()}
            >
              <Picker 
                data={data}
                onEmojiSelect={handleEmojiSelect}
                theme="dark"
                previewPosition="none"
                skinTonePosition="none"
                searchPosition="none"
                maxFrequentRows={2}
                navPosition="bottom"
                set="native"
                emojiSize={20}
                emojiButtonSize={28}
                perLine={8}
              />
            </div>
          )}
          <div className="flex items-center space-x-2">
            <button
              type="button"
              onClick={(e) => {
                e.preventDefault()
                e.stopPropagation()
                setShowEmojiPicker(!showEmojiPicker)
              }}
              className="text-gray-400 hover:text-gray-200 transition-colors"
            >
              <Smile size={20} />
            </button>
            <input
              ref={inputRef}
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
    </div>
  )
} 