'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/router'
import { motion } from 'framer-motion'
import { Users, Calendar, MessageSquare, Send } from 'lucide-react'
import Link from 'next/link'

interface Member {
  id: number
  name: string
  avatar?: string
  role: 'creator' | 'member'
  status: 'online' | 'offline'
}

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
  console.log(id) 
  const [currentUser, setCurrent] = useState('')
  const groupId = parseInt(id as string, 10);

  // Set initial state to an empty array to prevent null errors
  const [members, setMembers] = useState<Member[]>([])

  const [posts, setPosts] = useState<Post[]>([])
  const [events, setEvents] = useState<Event[]>([])
  const [messages, setMessages] = useState<string[]>([])
  const [newPost, setNewPost] = useState('')
  const [loggedInUserId, setLoggedInUserId] = useState<number | null>(null);

  const [newEvent, setNewEvent] = useState({
    title: '',
    description: '',
    datetime: ''
  })
  const [newMessage, setNewMessage] = useState('')
  const [socket, setSocket] = useState<WebSocket | null>(null)

  // fetch current username 
  const fetchCurrentUsername = async () => {
    const response = await fetch('http://localhost:8080/userName', {
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      }
    });
    
    if (!response.ok) {
      throw new Error(`Error fetching data: ${response.statusText}`);
    }
    
    const data = await response.json();
    setCurrent(data.username);
  }

  // Fetch members on mount
  const fetchMembers = async () => {
    try {
      const response = await fetch(`http://localhost:8080/groups/Members`, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ group_id: id })
      });
      const data = await response.json();
      const formattedMembers = data.map((member: any) => ({
        id: member.user_id,
        name: member.username,
        avatar: member.avatar || null, // Handle avatar if it's optional
        role: member.status, // Assuming 'status' maps to 'role'
        status: member.status === 'creator' ? 'online' : 'offline', // Adjust as needed
      }));
      setMembers(formattedMembers);
    } catch (error) {
      console.error(error);
    }
  };
  
  const DeleteGroup = async () => {
    try {
      const response = await fetch(`http://localhost:8080/groups/${id}`, {
        method: 'DELETE',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        }
      });
      if (response.ok) {
        router.push('/groups')
      } else {
        console.error('Failed to delete group');
      }
    } catch (error) {
      console.error(error);
    }
  }

  const handleEventResponse = async (eventId: number, response: 'going' | 'not going') => {
    try {
      // Send RSVP status to the backend
      const res = await fetch('http://localhost:8080/event/rsvp', {
        method: 'POST',
        credentials: 'include',

        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          event_id: eventId,
          user_id: loggedInUserId, 
          rsvp_status: response, 
        }),
      });
  
      // Check if the request was successful
      if (!res.ok) {
        console.error('Failed to update RSVP status');
        return;
      }
  
      // If successful, update the event in the state
      setEvents(prevEvents => 
        prevEvents.map(event => {
          if (event.id === eventId) {
            const updatedEvent = {
              ...event,
              going: event.going.filter(user => user !== currentUser),
              notGoing: event.notGoing.filter(user => user !== currentUser),
            };
  
            // Add current user to the appropriate list based on response
            if (response === 'going') {
              updatedEvent.going.push(currentUser);
            } else {
              updatedEvent.notGoing.push(currentUser);
            }
  
            return updatedEvent;
          }
          return event;
        })
      );
    } catch (error) {
      console.error('Error handling RSVP:', error);
    }
  };
  useEffect(() => {
    const fetchData = async () => {
      await loginUserID(); // Fetch the logged-in user's ID
    };
    fetchData();
  }, []);
  useEffect(() => {
    fetchCurrentUsername()
    fetchMembers()
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
      author: currentUser,
      created_at: new Date().toISOString(),
      comments: []
    }
    setPosts(prev => [post, ...prev])
    setNewPost('')
  }

  // const handleCreateEvent = (e: React.FormEvent) => {
  //   e.preventDefault()
  //   const event: Event = {
  //     id: Date.now(),
  //     ...newEvent,
  //     going: [],
  //     notGoing: []
  //   }
  //   setEvents(prev => [event, ...prev])
  //   setNewEvent({ title: '', description: '', datetime: '' })
  // }
  const handleCreateEvent = async (e: React.FormEvent) => {
    e.preventDefault();
    const formattedDate = new Date(newEvent.datetime).toISOString(); // Convert to ISO 8601
  
    const event = {
      group_id: groupId,
      creator_id: loggedInUserId,
      title: newEvent.title,
      description: newEvent.description,
      event_date: formattedDate, // Use formatted date
    };
  
    console.log(event,"Formatted event_date:", formattedDate);
  
    try {
      const response = await fetch('http://localhost:8080/event/create', {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(event),
      });
  
      if (response.ok) {
        const createdEvent = await response.json();
        setEvents((prev) => [
          {
            ...createdEvent,
            going: [],
            notGoing: [],
          },
          ...prev,
        ]);
        setNewEvent({ title: '', description: '', datetime: '' });
      } else {
        console.error('Failed to create event:', await response.text());
      }
    } catch (error) {
      console.error('Error creating event:', error);
    }
  };
  
  const loginUserID = async () => {
    try {
      const response = await fetch('http://localhost:8080/userIDBY', {
        method: 'GET',
        credentials: 'include', 
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch user ID, Status: ${response.status}`);
      }

      const data = await response.json();
      console.log('Fetched user ID:', data);

      if (data.userID) {
        setLoggedInUserId(data.userID);
      } else {
        throw new Error('User ID not found in response');
      }
    } catch (error) {
      console.error('Error fetching user ID:', error);
      setLoggedInUserId(null);
    }
  };
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
      <div className="max-w-[1920px] mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className='flex justify-between items-center'>
          <Link href="/groups">
            <button className="mb-6 px-4 py-2 bg-gray-700 text-white rounded-lg hover:bg-gray-600 transition-colors">
              ← Back to Groups
            </button>          
          </Link>

          {members && members[0]?.name === currentUser && members[0]?.role === 'creator' && (
            <button className="mb-6 px-4 py-2 bg-red-700 text-white rounded-lg hover:bg-red-600 transition-colors" onClick={DeleteGroup}>
              Delete Group
            </button>
          )}
        </div>

        <div className="flex flex-row space-x-6">
          {/* Members Section */}
          <div className="w-1/6">
            <div className="bg-white/10 backdrop-blur-lg rounded-lg p-6 sticky top-8">
              <h2 className="text-2xl font-semibold text-gray-200 mb-4">Members</h2>
              <div className="space-y-4">
                {members && members.map(member => (
                  <div key={member.id} className="flex items-center space-x-3">
                    <div className="w-10 h-10 bg-gray-700 rounded-full flex items-center justify-center">
                      {member.avatar ? (
                        <img
                          src={member.avatar}
                          alt={member.name}
                          className="w-full h-full rounded-full object-cover"
                        />
                      ) : (
                        <Users size={20} className="text-gray-400" />
                      )}
                    </div>
                    <div>
                      <p className="text-gray-200">{member.name}</p>
                      <div className="flex items-center space-x-2">
                        <span
                          className={`w-2 h-2 rounded-full ${
                            member.status === 'online' ? 'bg-green-500' : 'bg-gray-500'
                          }`}
                        />
                        <span className="text-sm text-gray-400">{member.role}</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
  
          {/* Posts Section */}
          <div className="w-1/2">
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
          <div className="w-1/4 space-y-6">
            <div className="bg-white/10 backdrop-blur-lg rounded-lg p-6">
              <h2 className="text-2xl font-semibold text-gray-200 mb-4">Events</h2>
              <form onSubmit={handleCreateEvent} className="mb-6">
                <input
                  type="text"
                  value={newEvent.title}
                  onChange={(e) => setNewEvent({ ...newEvent, title: e.target.value })}
                  className="w-full px-4 py-2 bg-gray-700 rounded-lg text-gray-200 mb-2"
                  placeholder="Event Title"
                />
                <textarea
                  value={newEvent.description}
                  onChange={(e) => setNewEvent({ ...newEvent, description: e.target.value })}
                  className="w-full px-4 py-2 bg-gray-700 rounded-lg text-gray-200 mb-2"
                  placeholder="Event Description"
                  rows={2}
                />
                <input
                  type="datetime-local"
                  value={newEvent.datetime}
                  onChange={(e) => setNewEvent({ ...newEvent, datetime: e.target.value })}
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
                    <h3 className="text-gray-200">{event.title}</h3>
                    <p className="text-sm text-gray-400">{event.description}</p>
                    <div className="text-xs text-gray-400">Time: {event.datetime}</div>
                    <div className="flex space-x-4 mt-2">
                      <button
                        onClick={() => handleEventResponse(event.id, 'going')}
                        className="px-3 py-1 bg-green-500 text-white rounded-lg"
                      >
                        Going
                      </button>
                      <button
                        onClick={() => handleEventResponse(event.id, 'not going')}
                        className="px-3 py-1 bg-red-500 text-white rounded-lg"
                      >
                        Not Going
                      </button>
                    </div>
                    <div className="mt-2 text-sm text-gray-400">
                      {event.going.length > 0 && (
                        <div>Going: {event.going.join(', ')}</div>
                      )}
                      {event.notGoing.length > 0 && (
                        <div>Not Going: {event.notGoing.join(', ')}</div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
  
            {/* Chat Section */}
            <div className="bg-white/10 backdrop-blur-lg rounded-lg p-6">
              <h2 className="text-2xl font-semibold text-gray-200 mb-4">Group Chat</h2>
              <div className="space-y-4">
                {messages.map((message, index) => (
                  <div key={index} className="bg-gray-700 rounded-lg p-4">
                    <p className="text-gray-200">{message}</p>
                  </div>
                ))}
              </div>
              <form onSubmit={handleSendMessage} className="mt-6 flex">
                <input
                  type="text"
                  value={newMessage}
                  onChange={(e) => setNewMessage(e.target.value)}
                  className="flex-grow px-4 py-2 bg-gray-700 rounded-lg text-gray-200"
                  placeholder="Type a message"
                />
                <button
                  type="submit"
                  className="ml-2 px-4 py-2 bg-blue-500 text-white rounded-lg"
                >
                  Send
                </button>
              </form>
            </div>
          </div>
        </div>
      </div>
    </div>
  )  
}
