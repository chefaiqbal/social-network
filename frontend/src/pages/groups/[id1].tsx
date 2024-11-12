'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/router'
import { motion } from 'framer-motion'
import { Users, Calendar, MessageSquare, Send } from 'lucide-react'

interface Event {
  id: number
  title: string
  description: string
  datetime: string
  going: string[]
  notGoing: string[]
}

interface Post {
  id: number
  content: string
  author: string
  created_at: string
  comments: {
    id: number
    content: string
    author: string
    created_at: string
  }[]
}

export default function GroupDetail() {
  const router = useRouter()
  const { id } = router.query
  
  const [posts, setPosts] = useState<Post[]>([])
  const [events, setEvents] = useState<Event[]>([])
  const [messages, setMessages] = useState<string[]>([])
  const [newPost, setNewPost] = useState('')
  const [newEvent, setNewEvent] = useState({
    title: '',
    description: '',
    datetime: ''
  })
  const [newMessage, setNewMessage] = useState('')
  const [socket, setSocket] = useState<WebSocket | null>(null)

  useEffect(() => {
    // Initialize WebSocket connection
    const ws = new WebSocket('ws://localhost:8080/ws')
    setSocket(ws)

    ws.onmessage = (event) => {
      const message = JSON.parse(event.data)
      if (message.type === 'groupChat') {
        setMessages(prev => [...prev, message.content])
      }
    }

    return () => ws?.close()
  }, [])

  const handleCreatePost = (e: React.FormEvent) => {
    e.preventDefault()
    if (!newPost.trim()) return

    const post: Post = {
      id: Date.now(),
      content: newPost,
      author: 'Current User',
      created_at: new Date().toISOString(),
      comments: []
    }
    setPosts(prev => [post, ...prev])
    setNewPost('')
  }

  const handleCreateEvent = (e: React.FormEvent) => {
    e.preventDefault()
    const event: Event = {
      id: Date.now(),
      ...newEvent,
      going: [],
      notGoing: []
    }
    setEvents(prev => [event, ...prev])
    setNewEvent({ title: '', description: '', datetime: '' })
  }

  const handleSendMessage = (e: React.FormEvent) => {
    e.preventDefault()
    if (!newMessage.trim() || !socket) return

    socket.send(JSON.stringify({
      type: 'groupChat',
      content: {
        group_id: id,
        message: newMessage
      }
    }))
    setNewMessage('')
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Posts Section */}
          <div className="lg:col-span-2 space-y-6">
            <div className="bg-white/10 backdrop-blur-lg rounded-lg p-6">
              <h2 className="text-2xl font-semibold text-gray-200 mb-4">Posts</h2>
              <form onSubmit={handleCreatePost} className="mb-6">
                <textarea
                  value={newPost}
                  onChange={(e) => setNewPost(e.target.value)}
                  className="w-full px-4 py-2 bg-gray-700 rounded-lg text-gray-200 mb-2"
                  placeholder="Write a post..."
                  rows={3}
                />
                <button
                  type="submit"
                  className="px-4 py-2 bg-blue-500 text-white rounded-lg"
                >
                  Post
                </button>
              </form>

              <div className="space-y-4">
                {posts.map(post => (
                  <div key={post.id} className="bg-gray-800 rounded-lg p-4">
                    <p className="text-gray-200">{post.content}</p>
                    <div className="mt-2 text-sm text-gray-400">
                      Posted by {post.author}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Events and Chat Section */}
          <div className="space-y-8">
            {/* Events */}
            <div className="bg-white/10 backdrop-blur-lg rounded-lg p-6">
              <h2 className="text-2xl font-semibold text-gray-200 mb-4">Events</h2>
              <form onSubmit={handleCreateEvent} className="mb-6">
                <input
                  type="text"
                  value={newEvent.title}
                  onChange={(e) => setNewEvent({...newEvent, title: e.target.value})}
                  className="w-full px-4 py-2 bg-gray-700 rounded-lg text-gray-200 mb-2"
                  placeholder="Event Title"
                />
                <textarea
                  value={newEvent.description}
                  onChange={(e) => setNewEvent({...newEvent, description: e.target.value})}
                  className="w-full px-4 py-2 bg-gray-700 rounded-lg text-gray-200 mb-2"
                  placeholder="Event Description"
                  rows={2}
                />
                <input
                  type="datetime-local"
                  value={newEvent.datetime}
                  onChange={(e) => setNewEvent({...newEvent, datetime: e.target.value})}
                  className="w-full px-4 py-2 bg-gray-700 rounded-lg text-gray-200 mb-2"
                />
                <button
                  type="submit"
                  className="px-4 py-2 bg-blue-500 text-white rounded-lg"
                >
                  Create Event
                </button>
              </form>

              <div className="space-y-4">
                {events.map(event => (
                  <div key={event.id} className="bg-gray-800 rounded-lg p-4">
                    <h3 className="text-xl font-semibold text-gray-200">{event.title}</h3>
                    <p className="text-gray-300 mt-2">{event.description}</p>
                    <p className="text-sm text-gray-400 mt-2">
                      {new Date(event.datetime).toLocaleString()}
                    </p>
                    <div className="mt-4 flex space-x-4">
                      <button className="px-4 py-2 bg-green-500 text-white rounded-lg">
                        Going ({event.going.length})
                      </button>
                      <button className="px-4 py-2 bg-red-500 text-white rounded-lg">
                        Not Going ({event.notGoing.length})
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Chat Room */}
            <div className="bg-white/10 backdrop-blur-lg rounded-lg p-6">
              <h2 className="text-2xl font-semibold text-gray-200 mb-4">Group Chat</h2>
              <div className="h-96 overflow-y-auto mb-4 space-y-2">
                {messages.map((message, index) => (
                  <div key={index} className="bg-gray-800 rounded-lg p-3">
                    <p className="text-gray-200">{message}</p>
                  </div>
                ))}
              </div>
              <form onSubmit={handleSendMessage} className="flex space-x-2">
                <input
                  type="text"
                  value={newMessage}
                  onChange={(e) => setNewMessage(e.target.value)}
                  className="flex-1 px-4 py-2 bg-gray-700 rounded-lg text-gray-200"
                  placeholder="Type a message..."
                />
                <button
                  type="submit"
                  className="px-4 py-2 bg-blue-500 text-white rounded-lg"
                >
                  <Send size={20} />
                </button>
              </form>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
