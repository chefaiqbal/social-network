'use client'

import { useState, useEffect, useMemo } from 'react'
import Sidebar from '@/components/layout/Sidebar'
import Header from '@/components/layout/Header'
import { Search, User as UserIcon, ExternalLink } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { useRouter } from 'next/router'

interface FollowResponse {
  status: string
  message: string
}

interface User {
  id: number
  username: string
  avatar?: string
  is_private: boolean
  about_me?: string
  first_name: string
  last_name: string
  is_following: boolean
  is_pending: boolean
}

interface FollowRequest {
  id: number
  follower_id: number
  username: string
  avatar?: string
}

const Follow = () => {
  const router = useRouter()
  const [users, setUsers] = useState<User[]>([])
  const [searchQuery, setSearchQuery] = useState('')
  const [currentIndex, setCurrentIndex] = useState(0)
  const itemsPerPage = 6
  const [loggedInUserId, setLoggedInUserId] = useState<number | null>(null)
  const [pendingFollowIds, setPendingFollowIds] = useState<Set<number>>(new Set())
  const [followRequests, setFollowRequests] = useState<FollowRequest[]>([])
  const [showFollowRequestsModal, setShowFollowRequestsModal] = useState(false)

  const fetchFollowRequests = async () => {
    try {
      const response = await fetch('http://localhost:8080/followRequest', {
        method: 'GET',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
      })
      if (response.ok) {
        const data = await response.json()
        console.log('Follow Requests:', data) // Add this line for debugging
        setFollowRequests(data)
      }
    } catch (error) {
      console.error('Error while fetching follow requests:', error)
    }
  }

  const fetchUsers = async () => {
    try {
      const [usersResponse, followStatusResponse] = await Promise.all([
        fetch('http://localhost:8080/AllUsers', {
          method: 'GET',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
        }),
        fetch('http://localhost:8080/followStatus', {
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
        })
      ]);

      if (usersResponse.ok && followStatusResponse.ok) {
        const [usersData, statusData] = await Promise.all([
          usersResponse.json(),
          followStatusResponse.json()
        ]);

        // Create maps for follow statuses
        const followStatusMap = statusData.reduce((acc: { [key: number]: string }, status: any) => {
          acc[status.followed_id] = status.status
          return acc
        }, {})

        // Update users with their current follow status
        const updatedUsers = usersData.map((user: User) => ({
          ...user,
          is_following: followStatusMap[user.id] === 'accept',
          is_pending: followStatusMap[user.id] === 'pending'
        }))

        setUsers(updatedUsers)
      }
    } catch (error) {
      console.error('Error fetching users:', error)
      setUsers([])
    }
  }

  const getCurrentUserId = async () => {
    try {
      const response = await fetch('http://localhost:8080/userIDBY', {
        credentials: 'include',
      });
      if (response.ok) {
        const data = await response.json();
        console.log('Current User ID:', data.userID) // Add this line for debugging
        return data.userID;
      }
      console.error('Failed to fetch current user ID:', response.statusText)
      return null;
    } catch (error) {
      console.error('Error getting current user ID:', error);
      return null;
    }
  };

  const followUser = async (followedId: number) => {
    if (pendingFollowIds.has(followedId)) return

    try {
      const response = await fetch('http://localhost:8080/follow', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          followed_id: followedId,
        }),
      })

      if (response.ok) {
        const data: FollowResponse = await response.json()
        
        // Update the user's status based on the response
        setUsers(prevUsers => 
          prevUsers.map(user => 
            user.id === followedId 
              ? { 
                  ...user, 
                  is_following: data.status === 'accept',
                  is_pending: data.status === 'pending'
                }
              : user
          )
        )

        // Update pending follows if needed
        if (data.status === 'pending') {
          setPendingFollowIds(prev => new Set(prev.add(followedId)))
          await fetchFollowRequests()
        }
      }
    } catch (error) {
      console.error('Error while follow request:', error)
    }
  }

  const handleAcceptFollowRequest = async (requestId: number) => {
    console.log(requestId, "accept", "this is handleAcceptFollowRequest")
    try {
      const response = await fetch(`http://localhost:8080/follow/requestF/${requestId}`, {
        method: 'PATCH',
        credentials: 'include',
        headers: { 
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        },
        body: JSON.stringify({ status: 'accept' }),
      })

      if (response.ok) {
        // Remove the request from followRequests
        setFollowRequests(prev => prev.filter(req => req.id !== requestId))

        // Update the users list to reflect the new status
        setUsers(prevUsers => 
          prevUsers.map(user => {
            const isRequestUser = user.id === requestId
            return isRequestUser
              ? {
                  ...user,
                  is_following: true,
                  is_pending: false
                }
              : user
          })
        )

        // Remove from pending IDs
        setPendingFollowIds(prev => {
          const newSet = new Set(prev)
          newSet.delete(requestId)
          return newSet
        })

        // Fetch updated user list to ensure all statuses are current
        await fetchUsers()
      }
    } catch (error) {
      console.error('Error handling follow request:', error)
    }
  }

  const handleRejectFollowRequest = async (requestId: number) => {
    console.log(requestId, "reject", "this is handleRejectFollowRequest")
    try {
      const response = await fetch(`http://localhost:8080/follow/requestF/${requestId}`, {
        method: 'PATCH',
        credentials: 'include',
        headers: { 
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        },
        body: JSON.stringify({ status: 'reject' }),
      })

      if (response.ok) {
        // Remove the request from followRequests
        setFollowRequests(prev => prev.filter(req => req.id !== requestId))

        // Update the users list to reflect the new status
        setUsers(prevUsers => 
          prevUsers.map(user => {
            const isRequestUser = user.id === requestId
            return isRequestUser
              ? {
                  ...user,
                  is_following: false,
                  is_pending: false
                }
              : user
          })
        )

        // Remove from pending IDs
        setPendingFollowIds(prev => {
          const newSet = new Set(prev)
          newSet.delete(requestId)
          return newSet
        })

        // Fetch updated user list to ensure all statuses are current
        await fetchUsers()
      }
    } catch (error) {
      console.error('Error handling follow request:', error)
    }
  }

  // Add a WebSocket listener for follow status updates
  useEffect(() => {
    const ws = new WebSocket('ws://localhost:8080/ws')

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.type === 'follow_status_update') {
          // Update users list when receiving a follow status update
          fetchUsers()
        }
      } catch (error) {
        console.error('Error processing WebSocket message:', error)
      }
    }

    return () => {
      ws.close()
    }
  }, [])

  useEffect(() => {
    const fetchInitialData = async () => {
      const userId = await getCurrentUserId();
      setLoggedInUserId(userId);
      await fetchUsers();
      await fetchFollowRequests();
    };

    fetchInitialData();
  }, [])

  const filteredUsers = useMemo(
    () => users.filter(user => user.username.toLowerCase().includes(searchQuery.toLowerCase())),
    [users, searchQuery]
  )

  const displayedUsers = useMemo(
    () => filteredUsers.slice(currentIndex, currentIndex + itemsPerPage),
    [filteredUsers, currentIndex, itemsPerPage]
  )

  const nextPage = () => {
    if (currentIndex + itemsPerPage < filteredUsers.length) {
      setCurrentIndex(currentIndex + itemsPerPage)
    }
  }

  const prevPage = () => {
    if (currentIndex - itemsPerPage >= 0) {
      setCurrentIndex(currentIndex - itemsPerPage)
    }
  }

  const closeFollowRequestsModal = () => {
    setShowFollowRequestsModal(false)
  }

  const viewProfile = (userId: number) => {
    router.push(`/profile/${userId}`)
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
      <Header />
      <div className="flex pt-16">
        <Sidebar />
        <div className="flex-1 overflow-y-auto">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
            <h1 className="text-3xl font-bold text-gray-200 mb-6">Follow</h1>

            <div className="relative mb-8">
              <input
                type="text"
                placeholder="Search user..."
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-10 pr-4 py-2 bg-white/10 backdrop-blur-lg border border-gray-700/50 rounded-full focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-200"
              />
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" size={20} />
            </div>

            <div className="grid grid-cols-3 gap-6 mb-6">
              <AnimatePresence>
                {displayedUsers.length > 0 ? (
                  displayedUsers.map(user => (
                    <motion.div
                      key={user.id}
                      initial={{ opacity: 0, y: 50 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -50 }}
                      transition={{ duration: 0.3 }}
                      className="p-6 bg-gray-800/50 backdrop-blur-lg rounded-lg shadow-lg text-center relative group hover:bg-gray-800/70 transition-all duration-300"
                    >
                      <div className="cursor-pointer">
                        <div className="relative inline-block">
                          {user.avatar ? (
                            <img 
                              src={user.avatar} 
                              alt={`${user.username}'s avatar`} 
                              className="w-24 h-24 rounded-full mx-auto mb-4 border-2 border-transparent group-hover:border-blue-500 transition-all duration-300" 
                            />
                          ) : (
                            <div className="w-24 h-24 rounded-full mx-auto mb-4 bg-gray-700 flex items-center justify-center">
                              <UserIcon size={40} className="text-gray-400" />
                            </div>
                          )}
                          {user.is_private && (
                            <span className="absolute top-0 right-0 bg-yellow-500 rounded-full p-1">
                              🔒
                            </span>
                          )}
                        </div>
                        
                        <h2 className="text-xl font-semibold text-gray-200 group-hover:text-blue-400 transition-colors">
                          {user.username}
                        </h2>
                        
                        <p className="text-gray-400 mt-2 line-clamp-2 hover:line-clamp-none">
                          {user.about_me}
                        </p>
                      </div>

                      <div className="mt-4 flex justify-center space-x-2">
                        <button
                          onClick={() => followUser(user.id)}
                          className={`px-4 py-2 rounded transition-all duration-300 ${
                            user.is_pending
                              ? 'bg-gray-600 text-gray-300 cursor-not-allowed'
                              : user.is_following
                              ? 'bg-green-600 text-white hover:bg-green-700'
                              : 'bg-blue-600 text-white hover:bg-blue-700'
                          }`}
                          disabled={user.is_pending}
                        >
                          {user.is_pending 
                            ? 'Pending' 
                            : user.is_following 
                            ? 'Following' 
                            : 'Follow'}
                        </button>
                        
                        <button
                          onClick={() => user.is_private && !user.is_following ? 
                            alert('This is a private profile. Please follow to view.') : 
                            viewProfile(user.id)}
                          className={`px-4 py-2 bg-gray-700 text-white rounded hover:bg-gray-600 transition-colors flex items-center space-x-1 ${
                            user.is_private && !user.is_following ? 'opacity-50 cursor-not-allowed' : ''
                          }`}
                        >
                          <ExternalLink size={16} />
                          <span>Profile</span>
                        </button>
                      </div>

                      {user.is_private && !user.is_following && (
                        <div className="mt-2 text-sm text-yellow-500">
                          Follow to view full profile
                        </div>
                      )}
                    </motion.div>
                  ))
                ) : (
                  <div className="col-span-3 text-center py-10">
                    <p className="text-gray-400 text-lg">No users found.</p>
                  </div>
                )}
              </AnimatePresence>
            </div>

            <div className="flex justify-center space-x-4 mb-6">
              <button
                onClick={prevPage}
                disabled={currentIndex === 0}
                className="px-4 py-2 bg-blue-600 text-white rounded disabled:opacity-50"
              >
                Previous
              </button>
              <button
                onClick={nextPage}
                disabled={currentIndex + itemsPerPage >= filteredUsers.length}
                className="px-4 py-2 bg-blue-600 text-white rounded disabled:opacity-50"
              >
                Next
              </button>
            </div>

            <button
              onClick={() => (setShowFollowRequestsModal(true), fetchFollowRequests())}
              className="px-4 py-2 bg-gray-700 text-white rounded-full mt-4 mb-6"
            >
              Show Follow Requests
            </button>

            <AnimatePresence>
              {showFollowRequestsModal && (
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  className="fixed inset-0 bg-black bg-opacity-50 flex justify-center items-center"
                  onClick={closeFollowRequestsModal} 
                >
                  <motion.div
                    initial={{ scale: 0.8 }}
                    animate={{ scale: 1 }}
                    exit={{ scale: 0.8 }}
                    className="bg-gray-900 p-6 rounded-lg shadow-lg max-w-md w-full mx-4"
                    onClick={(e) => e.stopPropagation()} 
                  >
                    <div className="flex justify-between items-center mb-4">
                      <h2 className="text-2xl font-semibold text-gray-200">Follow Requests</h2>
                      <button
                        onClick={closeFollowRequestsModal}
                        className="text-gray-400 hover:text-gray-200"
                      >
                        Close
                      </button>
                    </div>

                    {followRequests.length > 0 ? (
                      followRequests.map(request => (
                        <div key={request.id} className="flex items-center justify-between mb-4">
                          <div className="flex items-center">
                            <p className="text-lg text-gray-200">{request.username}</p>
                          </div>
                          <div className="flex space-x-2">
                            <button
                            onClick={async () => {
                              console.log("Accept button clicked");
                              await handleAcceptFollowRequest(request.id);
                            }}
                            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
                            >
                            Accept
                            </button>

                            <button
                              onClick={() => handleRejectFollowRequest(request.id)}
                              className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
                            >
                              Reject
                            </button>
                          </div>
                          
                        </div>
                      ))
                    ) : (
                      <p className="text-gray-400">No pending follow requests.</p>
                    )}
                  </motion.div>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Follow