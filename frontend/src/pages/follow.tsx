import { useState, useEffect, useMemo } from 'react'
import Sidebar from '@/components/layout/Sidebar'
import Header from '@/components/layout/Header'
import { Search } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'

interface User {
  id: number
  username: string
  avatar: string
  is_private: boolean
  about_me: string
}

const Follow = () => {
  const [users, setUsers] = useState<User[]>([])
  const [searchQuery, setSearchQuery] = useState('')
  const [currentIndex, setCurrentIndex] = useState(0)
  const itemsPerPage = 6
  const [loggedInUserId, setLoggedInUserId] = useState<number | null>(null)
  const [pendingFollowIds, setPendingFollowIds] = useState<Set<number>>(new Set())

  const followRequest = async (followedId: number) => {
    try {
      const response = await fetch('http://localhost:8080/follow', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          follower_id: loggedInUserId,
          followed_id: followedId,
          status: 'pending',
        }),
      })
      if (response.ok) {
        setPendingFollowIds(prev => new Set(prev.add(followedId)))
      } else {
        console.error(`Failed to follow user ID ${followedId}`)
      }
    } catch (error) {
      console.error('Error while follow request:', error)
    }
  }

  // Fetches pending follow statuses and user ID
  useEffect(() => {
    const fetchInitialData = async () => {
      try {
        const [statusResponse, userIdResponse] = await Promise.all([
          fetch('http://localhost:8080/followStatus', {
            method: 'GET',
            credentials: 'include',
            headers: { 'Content-Type': 'application/json' },
          }),
          fetch('http://localhost:8080/userIDBY', {
            method: 'GET',
            credentials: 'include',
            headers: { 'Content-Type': 'application/json' },
          })
        ])

        if (statusResponse.ok && userIdResponse.ok) {
          const statusData = await statusResponse.json()
          const userIdData = await userIdResponse.json()

          const pendingIds = new Set<number>(statusData.map((follow: { followed_id: number }) => follow.followed_id))
          setPendingFollowIds(pendingIds)
          setLoggedInUserId(userIdData.userID)
        } else {
          console.error('Failed to fetch initial data')
        }
      } catch (error) {
        console.error('Error fetching initial data:', error)
      }
    }

    fetchInitialData()
  }, [])

  // Fetch all users
  useEffect(() => {
    const fetchUsers = async () => {
      try {
        const response = await fetch('http://localhost:8080/AllUsers', {
          method: 'GET',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
        })
        if (response.ok) {
          const data = await response.json()
          setUsers(data)
        } else {
          console.error('Failed to fetch users')
        }
      } catch (error) {
        console.error('Error fetching users:', error)
      }
    }

    fetchUsers()
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

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-6">
              <AnimatePresence>
                {displayedUsers.length > 0 ? (
                  displayedUsers.map(user => (
                    <motion.div
                      key={user.id}
                      initial={{ opacity: 0, y: 50 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -50 }}
                      transition={{ duration: 0.3 }}
                      className="p-4 bg-gray-800 rounded-lg shadow-lg text-center"
                    >
                      <img src={user.avatar} alt={`${user.username}'s avatar`} className="w-20 h-20 rounded-full mx-auto mb-4" />
                      <h2 className="text-xl font-semibold text-gray-200">{user.username}</h2>
                      <p className="text-gray-400">{user.about_me}</p>
                      <p className="text-gray-400">Private: {user.is_private ? 'Yes' : 'No'}</p>
                      <button
                        onClick={() => followRequest(user.id)}
                        className="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
                      >
                        {pendingFollowIds.has(user.id) ? 'Pending' : 'Follow'}
                      </button>
                    </motion.div>
                  ))
                ) : (
                  <p className="text-gray-400 text-center col-span-3">No users found.</p>
                )}
              </AnimatePresence>
            </div>

            <div className="flex justify-center space-x-4">
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
          </div>
        </div>
      </div>
    </div>
  )
}

export default Follow
