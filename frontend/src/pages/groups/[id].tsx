'use client'

import { useState, useEffect, useRef } from 'react'
import { useRouter } from 'next/router'
import { motion } from 'framer-motion'
import { Users, Calendar, MessageCircle, Send, Smile, X, Image, MessageSquare } from 'lucide-react'
import Link from 'next/link'
import data from '@emoji-mart/data'
import Picker from '@emoji-mart/react'
import SearchBar from '@/components/searchbar'
import CreateGroupPost from '@/components/layout/CreateGroupPost'
import PendingMembers from '@/components/layout/PendingMembers'
import fetchPendingUsers from "@/lib/GetPendingMembers";
import Header from '@/components/layout/Header'
import GroupName from '@/components/ui/GroupName'
import { FaTimes } from 'react-icons/fa';
import GroupPost from '@/components/layout/GroupPost'




interface GroupMessage {
  id?: number
  sender_id: number
  content: string
  created_at: string
  username: string
}

interface Member {
  id: number
  name: string
  avatar?: string
  role: 'creator' | 'member'
  status: 'online' | 'offline'
}

// interface RSVP {
//   user_id: number;
//   rsvp_status: 'going' | 'not going';
// }

interface RSVP {
  user_id: number;
  username: string;
  rsvp_status: 'going' | 'not going';
}

interface EventWithRSVPs {
  event: Event;
  rsvps: RSVP[];
}

interface Event {
  id: number;
  title: string;
  description: string;
  event_date: string;
  going: string[];
  notGoing: string[];
}


interface Post {
  title: string
  id: number
  content: string
  author: string
  author_name: string
  created_at: string
  media?: string
  media_type?: string
  comments: {
      id: number
      content: string
      author: string
      author_name: string
      created_at: string
      media?: string
      media_type?: string
  }[]
}


export default function GroupDetail() {
  const router = useRouter()
  const { id } = router.query
  const groupId = parseInt(id as string, 10)
  const [currentUser, setCurrent] = useState('')
  const [members, setMembers] = useState<Member[]>([])
  const [posts, setPosts] = useState<Post[]>([])
  const [messages, setMessages] = useState<GroupMessage[]>([])
  const [newPost, setNewPost] = useState('')
  const [loggedInUserId, setLoggedInUserId] = useState<number | null>(null);
  const [events, setEvents] = useState<Event[]>([])
  const [users, setUsers] = useState([]);
  const [searchQuery, setSearchQuery] = useState<string>(""); // Search query state
  const [nonMembers, setNonMembers] = useState<{ id: number; username: string }[]>([]);
  const [filteredNonMembers, setFilteredNonMembers] = useState<{ id: number; username: string }[]>([]);
  const ITEMS_PER_PAGE = 2; // Number of users per page
  const [isMember, setIsMember] = useState<boolean | null>(null);
  const [error, setError] = useState<string | null>(null); 


  const [newEvent, setNewEvent] = useState({
    title: '',
    description: '',
    datetime: ''
  })
  const [newMessage, setNewMessage] = useState('')
  const [socket, setSocket] = useState<WebSocket | null>(null)
  const [showEmojiPicker, setShowEmojiPicker] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const messageContainerRef = useRef<HTMLDivElement>(null)
  const [pendingMembers, setPendingMembers] = useState<Member[]>([])
  // Add these state variables
const [activeCommentPost, setActiveCommentPost] = useState<number | null>(null);
const [newComment, setNewComment] = useState('');
const [commentImage, setCommentImage] = useState<File | null>(null);

const handleToggleComments = (postId: number) => {
  console.log('Toggling comments for post:', postId);
  console.log('Current post comments:', posts.find(p => p.id === postId)?.comments);
  setActiveCommentPost(activeCommentPost === postId ? null : postId);
};

const handleCommentImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
  if (e.target.files && e.target.files[0]) {
    setCommentImage(e.target.files[0]);
  }
};

const handleCommentSubmit = async (e: React.FormEvent, postId: number) => {
  e.preventDefault();
  
  if (!newComment.trim() && !commentImage) return;

  const formData = new FormData();
  formData.append('content', newComment);
  formData.append('post_id', postId.toString());
  formData.append('group_id', groupId.toString());

  // Handle image if present
  if (commentImage) {
      // Convert image file to base64
      const reader = new FileReader();
      reader.readAsDataURL(commentImage);
      
      reader.onload = async () => {
          const base64Image = reader.result as string;
          formData.append('media', base64Image);
          
          try {
              const response = await fetch(`http://localhost:8080/groups/${groupId}/posts/${postId}/comments`, {
                  method: 'POST',
                  credentials: 'include',
                  headers: {
                      'Content-Type': 'application/json',
                  },
                  body: JSON.stringify({
                      content: newComment,
                      media: base64Image,
                      post_id: postId,
                      group_id: groupId
                  }),
              });

              if (!response.ok) {
                  const errorData = await response.text();
                  throw new Error(errorData);
              }

              // Refresh posts to show new comment
              const postsData = await fetchGroupPosts(groupId);
              setPosts(postsData);

              // Reset form
              setNewComment('');
              setCommentImage(null);
              setActiveCommentPost(null);

          } catch (error) {
              console.error('Error posting comment:', error);
              alert(error instanceof Error ? error.message : 'Failed to post comment');
          }
      };

      reader.onerror = (error) => {
          console.error('Error reading file:', error);
          alert('Error processing image');
      };
  } else {
      // Submit without image
      try {
          const response = await fetch(`http://localhost:8080/groups/${groupId}/posts/${postId}/comments`, {
              method: 'POST',
              credentials: 'include',
              headers: {
                  'Content-Type': 'application/json',
              },
              body: JSON.stringify({
                  content: newComment,
                  post_id: postId,
                  group_id: groupId
              }),
          });

          if (!response.ok) {
              const errorData = await response.text();
              throw new Error(errorData);
          }

          // Refresh posts to show new comment
          const postsData = await fetchGroupPosts(groupId);
          setPosts(postsData);

          // Reset form
          setNewComment('');
          setCommentImage(null);
          setActiveCommentPost(null);

      } catch (error) {
          console.error('Error posting comment:', error);
          alert(error instanceof Error ? error.message : 'Failed to post comment');
      }
  }
};
  
  const handleAcceptMember = async (memberId: number) => {
    try {
      const payload = {
        group_id: groupId,
        user_id: memberId,
      };
      console.log("Sending payload:", payload);
  
      const response = await fetch("http://localhost:8080/groups/accept", {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });
  
      if (!response.ok) {
        const errorText = await response.text();
        console.error("Error accepting member:", errorText);
        return;
      }
  
      const data = await response.json();
      console.log("Member accepted:", data.message);
  
      // Move the member from pendingMembers to members
      setPendingMembers((prev) =>
        prev.filter((member) => member.id !== memberId)
      );
      setMembers((prev) => {
        const acceptedMember = pendingMembers.find((member) => member.id === memberId);
        if (acceptedMember) {
          return [...prev, acceptedMember];
        }
        return prev;
      });
  
      fetchMembers();
      fetchPendingUsers(groupId);
    } catch (error) {
      console.error("Network error accepting member:", error);
    }
  };
  
  const handleRejectMember = async (memberId: number) => {
    try {
      const payload = {
        group_id: groupId,
        user_id: memberId,
      };
      console.log("Sending payload:", payload);
  
      const response = await fetch("http://localhost:8080/groups/reject", {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });
  
      if (!response.ok) {
        const errorText = await response.text();
        console.error("Error rejecting member:", errorText);
        return;
      }
  
      const data = await response.json();
      console.log("Member rejected:", data.message);
  
      setPendingMembers((prev) =>
        prev.filter((member) => member.id !== memberId)
      );
  
      fetchPendingUsers(groupId);
    } catch (error) {
      console.error("Network error rejecting member:", error);
    }
  };

  
    
  const fetchCurrentUsername = async () => {
    const response = await fetch('http://localhost:8080/userName', {
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      }
    })
    if (!response.ok) throw new Error(`Error fetching data: ${response.statusText}`)
    const data = await response.json()
    setCurrent(data.username)
  }

  const fetchMembers = async () => {
    try {
      const response = await fetch(`http://localhost:8080/groups/Members`, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ group_id: id })
      })
      const data = await response.json()
      const formattedMembers = data.map((member: any) => ({
        id: member.user_id,
        name: member.username,
        avatar: member.avatar || null,
        role: member.status,
        status: member.status === 'creator' ? 'online' : 'offline',
      }))
      setMembers(formattedMembers)
    } catch (error) {
      console.error(error)
    }
  }

  const fetchGroupPosts = async (groupId: number): Promise<Post[]> => {
    const response = await fetch(`http://localhost:8080/groups/${groupId}/posts`, {
        method: 'GET',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json',
        },
    });
    
    if (!response.ok) throw new Error(`Error fetching posts: ${response.statusText}`);
    
    const data = await response.json();
    console.log('Posts with comments:', data); // Check this in browser console
    return data;
};


  const DeleteGroup = async () => {
    try {
      const response = await fetch(`http://localhost:8080/groups/${id}`, {
        method: 'DELETE',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        }
      })
      if (response.ok) {
        router.push('/groups')
      }
    } catch (error) {
      console.error(error)
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

  const scrollToBottom = (behavior: 'auto' | 'smooth' = 'auto') => {
    messageContainerRef.current?.scrollIntoView({ behavior })
  }

  const handleSendMessage = (e: React.FormEvent) => {
    e.preventDefault()
    if (!newMessage.trim() || !socket || socket.readyState !== WebSocket.OPEN) {
      console.log('Cannot send message:', {
        hasContent: Boolean(newMessage.trim()),
        hasSocket: Boolean(socket),
        socketState: socket?.readyState
      })
      return
    }

    const message = {
      type: 'groupChat',
      content: {
        group_id: groupId,
        message: newMessage.trim()
      }
    }

    console.log('Sending message:', message)

    try {
      socket.send(JSON.stringify(message))
      setNewMessage('')
      // Don't scroll yet - wait for the message to be confirmed via WebSocket
    } catch (error) {
      console.error('Error sending message:', error)
      // Optionally show an error message to the user
      alert('Failed to send message. Please try again.')
    }
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return new Intl.DateTimeFormat('en-US', {
      weekday: 'long', // Day of the week
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      hour12: true, // 12-hour clock
    }).format(date);
  };


  useEffect(() => {
    let ws: WebSocket | null = null;
    let reconnectAttempt = 0;
    const maxReconnectAttempts = 5;
    const reconnectDelay = 3000; // 3 seconds

    const connectWebSocket = () => {
      ws = new WebSocket('ws://localhost:8080/ws/group-chat')
      
      ws.onopen = () => {
        console.log('Group Chat WebSocket Connected')
        setSocket(ws)
        reconnectAttempt = 0 // Reset reconnect attempts on successful connection
        
        // Fetch existing messages when connection is established
        fetchGroupMessages()
      }

      ws.onmessage = (event) => {
        console.log('Group Chat message received:', event.data)
        try {
          const data = JSON.parse(event.data)
          if (data.type === 'groupChat') {
            const newMessage = {
              id: data.id,
              content: data.content,
              sender_id: data.sender_id,
              username: data.username,
              created_at: data.created_at
            }
            setMessages(prev => [...prev, newMessage])
            scrollToBottom('smooth')
          }
        } catch (error) {
          console.error('Error parsing message:', error)
        }
      }

      ws.onerror = (error) => {
        console.log('WebSocket error:', error)
      }

      ws.onclose = (event) => {
        console.log('WebSocket closed:', event)
        setSocket(null)

        // Attempt to reconnect if not a normal closure and within max attempts
        if (event.code !== 1000 && reconnectAttempt < maxReconnectAttempts) {
          reconnectAttempt++
          console.log(`Attempting to reconnect (${reconnectAttempt}/${maxReconnectAttempts})...`)
          setTimeout(connectWebSocket, reconnectDelay)
        }
      }
    }

    // Initial connection
    connectWebSocket()

    // Cleanup function
    return () => {
      if (ws) {
        ws.onclose = null // Prevent reconnection attempt on intentional closure
        ws.close(1000, 'Component unmounting')
      }
    }
  }, [groupId]) // Only reconnect if groupId changes

  // Add this function to fetch existing messages
  const fetchGroupMessages = async () => {
    try {
      const response = await fetch(`http://localhost:8080/groups/messages?groupId=${groupId}`, {
        credentials: 'include'
      })
      if (response.ok) {
        const data = await response.json()
        setMessages(data.reverse()) // Reverse to show newest messages at bottom
      }
    } catch (error) {
      console.error('Error fetching group messages:', error)
    }
  }



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

const handleEventResponse = async (eventId: number, response: 'going' | 'not going') => {
  if (loggedInUserId === null) {
    console.error('User is not logged in');
    return;
  }

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
    console.log('Fetched user ID:', data.id);

    if (data.id) {
      setLoggedInUserId(data.id);
    } else {
      throw new Error('User ID not found in response');
    }
  } catch (error) {
    console.error('Error fetching user ID:', error);

    setLoggedInUserId(null);
  }
};



useEffect(() => {
  if (!groupId) return; 
  const checkMembership = async () => {
    try {
      const response = await fetch("http://localhost:8080/groups/ismember", {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ group_id: groupId }),
      });

      if (response.status === 403) {
        console.log("NOT MEM");
        setIsMember(false);
        setError("You are not a member of this group.");
        return;
      }

      if (response.ok) {
        const data = await response.json();  // Parse the JSON response
        console.log("response data:", data);

        if (data.message === "User is a member") {
          setIsMember(true);
          console.log("MEM");
        } else {
          console.log("NOT MEM");
        }
      }
    } catch (error) {
      console.error("Error fetching membership status:", error);
      setError("An error occurred while checking membership.");
    }
  };

  checkMembership();
}, [groupId]);


useEffect(() => {
  const fetchEventsWithRSVPs = async () => {
    try {
      const eventsResponse = await fetch(`http://localhost:8080/event/getGroupEvents/${groupId}`, {
        method: "GET",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
      });

      if (!eventsResponse.ok) throw new Error("Failed to fetch events");
      const eventsData: Event[] = await eventsResponse.json();

      const eventsWithRSVPs = await Promise.all(
        eventsData.map(async (event) => {
          const rsvpResponse = await fetch(`http://localhost:8080/event/rsvps/${event.id}`, {
            method: "GET",
            credentials: "include",
            headers: { "Content-Type": "application/json" },
          });

          let rsvpData: RSVP[] = [];
          if (rsvpResponse.ok) {
            const response = await rsvpResponse.json();
            console.log("RSVP Response for Event ID", event.id, ":", response);

            // Extract the `rsvps` array from the response
            rsvpData = response.rsvps || [];
          }

          // Map the RSVP data to extract usernames
          const going = rsvpData
            .filter((rsvp) => rsvp.rsvp_status === "going")
            .map((rsvp) => rsvp.username);

          const notGoing = rsvpData
            .filter((rsvp) => rsvp.rsvp_status === "not going")
            .map((rsvp) => rsvp.username);

          return { ...event, going, notGoing };
        })
      );

      console.log("Final Events with RSVP Data:", eventsWithRSVPs);
      setEvents(eventsWithRSVPs);
    } catch (err) {
      console.error("Failed to fetch events or RSVPs", err);
    }
  };

  if (groupId) {
    fetchEventsWithRSVPs();
  }
}, [groupId]);

  useEffect(() => {
    const fetchData = async () => {
      await loginUserID(); // Fetch the logged-in user's ID
    };
    fetchData();
  }, []);


  useEffect(() => {
    const fetchData = async () => {
      try {
        await fetchCurrentUsername();
        await fetchMembers();
        
        if (!isNaN(groupId)) {
          const postsData = await fetchGroupPosts(groupId);
          setPosts(postsData);
        }
      } catch (error) {
        console.error('Error fetching data:', error);
      }
    };

    if (groupId) {
      fetchData();
    }
  }, [groupId]); // Only re-run when groupId changes



  
    const handleSearch = async (query: string) => {
      try {
        const res = await fetch(`/api/users?query=${query}`);
        const data = await res.json();
        setUsers(data.users);
      } catch (error) {
        console.error('Error fetching users:', error);
      }
    };
  
    useEffect(() => {
      const fetchNonMembers = async () => {
        try {
          const response = await fetch("http://localhost:8080/groups/getnonmembers", {
            method: "POST",
            credentials: "include",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify({ group_id: groupId }),
          });
    
          if (!response.ok) {
            throw new Error("Failed to fetch non-members");
          }
    
          const data: { id: number; username: string }[] = await response.json();
          console.log("Non-members data:", data);
          setNonMembers(data);
          setFilteredNonMembers(data); // Initially display all non-members
        } catch (err) {
          console.error("Error fetching non-members:", err);
        }
      };
    
      if (groupId) {
        fetchNonMembers();
      }
    }, [groupId]);
    
    useEffect(() => {
      setFilteredNonMembers(
        nonMembers.filter((user) =>
          user.username.toLowerCase().includes(searchQuery.toLowerCase())
        )
      );
    }, [searchQuery, nonMembers]);


      const [currentPage, setCurrentPage] = useState(1);
    
      const totalPages = Math.ceil(filteredNonMembers.length / ITEMS_PER_PAGE);
    
      // Get the users for the current page
      const paginatedUsers = filteredNonMembers.slice(
        (currentPage - 1) * ITEMS_PER_PAGE,
        currentPage * ITEMS_PER_PAGE
      );
    
      const handlePageChange = (newPage:number) => {
        setCurrentPage(newPage);
      };

      // Add a ping function to keep the connection alive
      useEffect(() => {
        const pingInterval = setInterval(() => {
          if (socket && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify({ type: 'ping' }))
          }
        }, 30000) // Send ping every 30 seconds

        return () => clearInterval(pingInterval)
      }, [socket])


      const handleInvite = async (userId:number) => {
        console.log("idddddddd",userId)
        console.log("groupId",groupId)
        
      
        try {
          const response = await fetch('http://localhost:8080/groups/invitation', {
            method: 'POST',
            credentials: 'include', 
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({ reciver_id:userId, groupId }), // Sending userId and groupId in the body
          });
      
          if (response.ok) {
            const data = await response.json();
            alert(data.message); // Show success message from the backend
          } else {
            const errorData = await response.json();
            alert(errorData.message || 'Failed to send group invitation.');
          }
        } catch (error) {
          console.error('Error in handleInvite:', error);
          alert('An unexpected error occurred. Please try again.');
        }
      };
      return isMember ?  (
        <div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
          <Header />
          <div className="pt-20"></div>
          <div className="max-w-[1920px] mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className='flex justify-between items-center mt-4'>
          <Link href="/groups">
            <button className="mb-6 px-4 py-2 bg-gray-700 text-white rounded-lg hover:bg-gray-600 transition-colors">
              ← Back to Groups
            </button>          
          </Link>

          {/* here should be the title */}
          <GroupName groupID={groupId} />

          <div className="w-1/4 space-y-6">
          {members && members[0]?.name === currentUser && members[0]?.role === 'creator' && (
            <button className="ml-60 mb-3 px-4 py-2 bg-red-700 text-white rounded-lg hover:bg-red-600 transition-colors" onClick={DeleteGroup}>
              Delete Group
            </button>
          )}
        </div>
        </div>

        <div className="flex flex-row space-x-6">
        <div className="w-1/6">
  <div className="bg-white/10 backdrop-blur-lg rounded-lg p-6 sticky top-8">
    <h2 className="text-2xl font-semibold text-gray-200 mb-4">Members</h2>
    <div className="space-y-4">
      {members && members.length > 0 ? (
        members.map((member) => (
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
                    member.status === "online" ? "bg-green-500" : "bg-gray-500"
                  }`}
                />
                <span className="text-sm text-gray-400">{member.role}</span>
              </div>
            </div>
          </div>
        ))
      ) : (
        <div className="text-gray-500">
        No members at the moment.
        </div>     
         )
      }
    </div>

    <hr className="border-gray-700 my-6" />
    {members && members[0]?.name === currentUser && members[0]?.role === 'creator' && (

      <PendingMembers
          groupId={groupId}
          onAccept={handleAcceptMember}
          onReject={handleRejectMember}
       />
    )}
  </div>

  <div className="mt-8 bg-white/10 backdrop-blur-lg rounded-lg p-6 sticky top-8">
      {/* Search Bar */}
      <form className="flex items-center space-x-4 bg-gray-800/50 p-4 rounded-lg shadow-md">
        <div className="relative w-full">
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search for users..."
            className="flex-1 w-full px-4 py-2 text-gray-200 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 placeholder-gray-400"
          />
        </div>
      </form>

      {/* Non-Members List */}
      <ul className="mt-4 bg-gray-800/50 p-4 rounded-lg shadow-md">
        {paginatedUsers.length > 0 ? (
          paginatedUsers.map((user, index) => (
            <li
              key={user.id}
              className={`flex items-center justify-between py-2 text-gray-200 ${
                index === 0 ? "" : "border-t"
              } border-gray-700`}
            >
              <span>{user.username}</span>
              <button
                onClick={() => handleInvite(user.id)}
                className="px-3 py-1 text-sm text-white bg-blue-500 rounded-lg hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-300"
              >
                Invite
              </button>
            </li>
          ))
        ) : (
          <li className="py-2 text-gray-400">No users found</li>
        )}
      </ul>

      {/* Pagination Controls */}
              <div className="mt-4 flex justify-center space-x-4">
          {/* Previous Arrow */}
          <button
            onClick={() => handlePageChange(currentPage - 1)}
            disabled={currentPage === 1}
            className="p-2 text-gray-400 bg-gray-700 rounded-lg hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span className="sr-only">Previous</span>
            <svg
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              strokeWidth="2"
              stroke="currentColor"
              className="w-5 h-5"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M15 19l-7-7 7-7"
              />
            </svg>
          </button>

          <span className="px-3 py-1 text-gray-200">
            Page {currentPage} of {totalPages}
          </span>

          {/* Next Arrow */}
          <button
            onClick={() => handlePageChange(currentPage + 1)}
            disabled={currentPage === totalPages}
            className="p-2 text-gray-400 bg-gray-700 rounded-lg hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span className="sr-only">Next</span>
            <svg
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              strokeWidth="2"
              stroke="currentColor"
              className="w-5 h-5"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M9 5l7 7-7 7"
              />
            </svg>
          </button>
        </div>


    </div>

        </div>
        <div className="w-1/2">
  <>
    <CreateGroupPost onPostCreated={fetchGroupPosts} groupID={groupId} />
    <div className="space-y-4 overflow-y-auto max-h-[500px]">
      {posts && posts.length > 0 ? (
        posts.map(post => (
          <GroupPost 
            key={post.id} 
            post={post} 
            groupId={groupId} 
          />
        ))
      ) : (
        <div className="text-gray-500">
          No posts at the moment.
        </div>
      )}
    </div>
  </>
</div>
          
          {/* Events Section */}
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
                      <div className="text-xs text-gray-400">Time: {formatDate(event.event_date)}</div>
                      <div className="flex space-x-4 mt-2">
                        {/* Button for Going */}
                        <button
                          onClick={() => handleEventResponse(event.id, 'going')}
                          className={`px-3 py-1 rounded-lg ${
                            loggedInUserId !== null && event.going.includes(currentUser)
                              ? 'bg-green-700' // If user is already going, highlight the button
                              : 'bg-green-500'
                          } text-white flex items-center space-x-2`}
                        >
                          <span>Going</span>
                          <span className=" px-2 py-1 rounded text-sm">
                            {event.going.length}
                          </span>
                        </button>

                        {/* Button for Not Going */}
                        <button
                          onClick={() => handleEventResponse(event.id, 'not going')}
                          className={`px-3 py-1 rounded-lg ${
                            loggedInUserId !== null && event.notGoing.includes(currentUser)
                              ? 'bg-red-700' // If user is already not going, highlight the button
                              : 'bg-red-500'
                          } text-white flex items-center space-x-2`}
                        >
                          <span>Not Going</span>
                          <span className=" px-2 py-1 rounded text-sm">
                            {event.notGoing.length}
                          </span>
                        </button>
                      </div>


                      <div className="mt-2 text-sm text-gray-400">
                        {event.going.length > 0 && (
                          <div>Going: {event.going.join(', ')}</div> // Display usernames
                        )}
                        {event.notGoing.length > 0 && (
                          <div>Not Going: {event.notGoing.join(', ')}</div> // Display usernames
                        )}
                      </div>
                    </div>
                  ))}
              </div>
            </div>
            <div className="bg-white/10 backdrop-blur-lg rounded-lg">
              <div className="flex items-center justify-between p-3">
                <div className="flex items-center space-x-2">
                  <MessageCircle className="w-8 h-8" />
                  <span className="text-gray-200">Group Chat</span>
                </div>
              </div>

              <div ref={messageContainerRef} className="h-96 overflow-y-auto p-4">
                {messages.map((message, index) => (
                  <div 
                    key={message.id || index} 
                    className={`flex ${message.username !== currentUser ? 'justify-start' : 'justify-end'}`}
                  >
                    <div className={`max-w-[70%] rounded-lg p-2 mb-2 ${
                      message.username !== currentUser ? 'bg-gray-700' : 'bg-blue-600'
                    }`}>
                      <p className="text-gray-200">{message.content}</p>
                      <span className="text-xs text-gray-400">
                        {new Date(message.created_at).toLocaleTimeString()}
                      </span>
                    </div>
                  </div>
                ))}
              </div>

              <div className="p-3">
                <div className="relative flex items-center">
                  <button
                    onClick={() => setShowEmojiPicker(!showEmojiPicker)}
                    className="p-2"
                  >
                    <Smile className="text-gray-400" />
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
                    onKeyPress={(e) => {
                      if (e.key === 'Enter' && !e.shiftKey) {
                        e.preventDefault()
                        handleSendMessage(e)
                      }
                    }}
                    className="flex-1 bg-gray-700 rounded-lg px-4 py-2 mr-2 text-gray-200"
                    placeholder="Type a message..."
                  />
                  
                  <button
                    onClick={handleSendMessage}
                    disabled={!newMessage.trim()}
                    className="bg-blue-600 p-2 rounded-lg"
                  >
                    <Send className="text-white" />
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  ) : (
    // <div>You are not a member of this group.</div>


    <div className="min-h-screen flex flex-col items-center justify-center bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
    {/* Profile Card */}
    <div className="max-w-lg w-full bg-gray-800/60 backdrop-blur-lg rounded-lg shadow-lg overflow-hidden">
      {/* Header */}
      <div className="relative h-24 bg-gradient-to-r from-blue-600 to-purple-600 flex items-center justify-center">
        <h2 className="text-white text-3xl font-bold flex items-center gap-2">
          <FaTimes />
          Group Unavailable
        </h2>
      </div>

      {/* Content */}
      <div className="p-8 text-center space-y-6">
        <p className="text-gray-300 text-lg">
          {
            <>
              <span className="block font-medium text-white">
                Oops! This Group is missing.
              </span>
               Or you are not a member of this group
            </>
          }
        </p>
      </div>
    </div>
  </div>
  );
}