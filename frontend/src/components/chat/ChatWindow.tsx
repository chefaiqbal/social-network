import { useState, useEffect, useRef, useCallback } from 'react'
import { User, Send, X, Circle, Smile } from 'lucide-react'
import throttle from 'lodash/throttle'
import data from '@emoji-mart/data'
import Picker from '@emoji-mart/react'

interface ChatMessage {
  id?: number
  sender_id: number
  recipient_id: number
  content: string
  created_at: string
  reactions?: MessageReaction[]
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

interface MessageReaction {
  emoji: string
  userId: number
}

export function ChatWindow({ user, websocket, onClose }: ChatWindowProps) {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [newMessage, setNewMessage] = useState('')
  const [isTyping, setIsTyping] = useState(false)
  const [page, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(true)
  const [isLoading, setIsLoading] = useState(false)
  const [showEmojiPicker, setShowEmojiPicker] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const messageContainerRef = useRef<HTMLDivElement>(null)
  const typingTimeoutRef = useRef<NodeJS.Timeout>()
  const inputRef = useRef<HTMLInputElement>(null)

  const scrollToBottom = (behavior: 'auto' | 'smooth' = 'auto') => {
    messagesEndRef.current?.scrollIntoView({ behavior })
  }

  const fetchMessages = async (pageNum: number, isInitial: boolean = false) => {
    if (isLoading || (!hasMore && !isInitial)) return

    try {
      setIsLoading(true)
      const response = await fetch(
        `http://localhost:8080/messages?userId=${user.id}&page=${pageNum}&limit=20`,
        {
          credentials: 'include'
        }
      )
      if (response.ok) {
        const data = await response.json()
        if (Array.isArray(data)) {
          if (data.length < 20) {
            setHasMore(false)
          }
          setMessages(prev => 
            isInitial ? data : [...data, ...prev]
          )
          if (isInitial) {
            setTimeout(() => scrollToBottom(), 100)
          }
        }
      }
    } catch (error) {
      console.error('Error fetching messages:', error)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchMessages(1, true)
  }, [user.id])

  useEffect(() => {
    if (websocket) {
      const handleMessage = (event: MessageEvent) => {
        const data = JSON.parse(event.data)
        if (data.type === 'chat' && 
            (data.sender_id === user.id || data.recipient_id === user.id)) {
          setMessages(prev => [...prev, data])
          scrollToBottom('smooth')
        }
      }

      websocket.addEventListener('message', handleMessage)

      return () => {
        websocket.removeEventListener('message', handleMessage)
      }
    }
  }, [websocket, user.id])

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

  const handleEmojiSelect = (emoji: any) => {
    const cursorPosition = inputRef.current?.selectionStart || newMessage.length
    const updatedMessage = 
      newMessage.slice(0, cursorPosition) + 
      emoji.native + 
      newMessage.slice(cursorPosition)
      
    setNewMessage(updatedMessage)
    setShowEmojiPicker(false)
    
    if (inputRef.current) {
      const newCursorPosition = cursorPosition + emoji.native.length
      inputRef.current.focus()
      inputRef.current.setSelectionRange(newCursorPosition, newCursorPosition)
    }
  }

  const sendMessage = () => {
    if (newMessage.trim() && websocket?.readyState === WebSocket.OPEN) {
      console.log('Sending message:', { type: 'chat', recipient_id: user.id, content: newMessage.trim() });
      
      const messageData = {
        type: 'chat',
        recipient_id: user.id,
        content: newMessage.trim()
      };

      try {
        websocket.send(JSON.stringify(messageData));
        setNewMessage('');
        setShowEmojiPicker(false);
        scrollToBottom('smooth');
      } catch (error) {
        console.error('Error sending message:', error);
      }
    }
  };

  return (
    <div className="fixed bottom-0 right-96 w-80 bg-white/10 backdrop-blur-lg rounded-t-lg shadow-lg">
      <div className="flex items-center justify-between p-3">
        <div className="flex items-center space-x-2">
          {user.avatar ? (
            <img src={user.avatar} alt={user.username} className="w-8 h-8 rounded-full" />
          ) : (
            <User className="w-8 h-8" />
          )}
          <span>{user.username}</span>
        </div>
        <button onClick={onClose}><X /></button>
      </div>

      <div 
        ref={messageContainerRef}
        className="h-96 overflow-y-auto p-4"
        onScroll={handleScroll}
      >
        {isLoading && (
          <div className="text-center py-2">
            <span className="text-gray-400">Loading messages...</span>
          </div>
        )}
        
        {messages.map((message, index) => (
          <div 
            key={message.id || index} 
            className={`flex ${message.sender_id !== user.id ? 'justify-start' : 'justify-end'}`}
          >
            <div className={`max-w-[70%] rounded-lg p-2 mb-2 ${
              message.sender_id !== user.id ? 'bg-gray-700' : 'bg-blue-600'
            }`}>
              <p className="text-gray-200">{message.content}</p>
              <span className="text-xs text-gray-400">
                {new Date(message.created_at).toLocaleTimeString()}
              </span>
            </div>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      <div className="p-3">
        <div className="relative flex items-center">
          <button
            onClick={() => setShowEmojiPicker(!showEmojiPicker)}
            className="p-2"
          >
            <Smile />
          </button>
          
          {showEmojiPicker && (
            <div className="absolute bottom-12 right-0">
              <Picker
                data={data}
                onEmojiSelect={handleEmojiSelect}
                theme="dark"
                previewPosition="none"
                skinTonePosition="none"
              />
            </div>
          )}

          <input
            ref={inputRef}
            type="text"
            value={newMessage}
            onChange={(e) => setNewMessage(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && sendMessage()}
            className="flex-1 bg-gray-700 rounded-lg px-4 py-2 mr-2"
          />
          
          <button
            onClick={sendMessage}
            disabled={!newMessage.trim()}
            className="bg-blue-600 p-2 rounded-lg"
          >
            <Send />
          </button>
        </div>
      </div>
    </div>
  )
}