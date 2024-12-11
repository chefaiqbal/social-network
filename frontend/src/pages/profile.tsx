'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/router'
import { User, MapPin, Calendar, Mail, Users, UserPlus, UserMinus, Lock, Unlock, Settings } from 'lucide-react'
import { motion } from 'framer-motion'
import Sidebar from '@/components/layout/Sidebar'
import Header from '@/components/layout/Header'

interface UserProfile {
  id: number
  username: string
  first_name: string
  last_name: string
  email: string
  avatar?: string
  about_me?: string
  date_of_birth: string
  created_at: string
  is_private: boolean
  nickName: string
}

interface Post {
  id: number
  title: string
  content: string
  media?: string
  privacy: number
  author: number
  created_at: string
}

interface Follower {
  id: number
  username: string
  avatar?: string
  status: string
  is_following: boolean
}

type TabType = 'posts' | 'followers' | 'following'

export default function Profile() {
  const [profile, setProfile] = useState<UserProfile | null>(null)
  const [activeTab, setActiveTab] = useState<TabType>('posts')
  const [posts, setPosts] = useState<Post[]>([])
  const [followers, setFollowers] = useState<Follower[]>([])
  const [following, setFollowing] = useState<Follower[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isEditingPrivacy, setIsEditingPrivacy] = useState(false)
  const router = useRouter()

  useEffect(() => {
    const fetchData = async () => {
      if (!router.isReady) return

      const userId = router.query.id || 'current'
      setIsLoading(true)
      setError(null)

      try {
        // Fetch profile
        const profileRes = await fetch(`http://localhost:8080/user/${userId}`, {
          credentials: 'include',
        })
        if (profileRes.ok) {
          const profileData = await profileRes.json()
          console.log('profileData:', profileData)
          setProfile(profileData)
        } else {
          throw new Error('Failed to fetch profile')
        }

        // Fetch data based on active tab
        switch (activeTab) {
          case 'posts':
            const postsRes = await fetch(`http://localhost:8080/posts/user/${userId}`, {
              credentials: 'include',
            })
            if (postsRes.ok) {
              const postsData = await postsRes.json()
              setPosts(Array.isArray(postsData) ? postsData : [])
            }
            break

          case 'followers':
            const followersRes = await fetch(`http://localhost:8080/followers`, {
              credentials: 'include',
            })
            if (followersRes.ok) {
              const followersData = await followersRes.json()
              setFollowers(Array.isArray(followersData) ? followersData : [])
            }
            break

          case 'following':
            try {
              const followingRes = await fetch(`http://localhost:8080/following/${userId}`, {
                credentials: 'include',
              })
              if (followingRes.ok) {
                const followingData = await followingRes.json()
                setFollowing(
                  Array.isArray(followingData) 
                    ? followingData
                        .filter(follow => follow.status === 'accept')
                        .map(follow => ({
                          id: follow.id,
                          username: follow.username,
                          avatar: follow.avatar,
                          status: follow.status,
                          is_following: true
                        }))
                    : []
                )
              }
            } catch (error) {
              console.error('Error fetching following:', error)
            }
            break
        }
      } catch (error) {
        console.error('Error fetching data:', error)
        setError(error instanceof Error ? error.message : 'Failed to load data')
      } finally {
        setIsLoading(false)
      }
    }

    fetchData()
  }, [router.isReady, router.query.id, activeTab])

  const TabButton = ({ tab, label }: { tab: TabType; label: string }) => (
    <button
      onClick={() => setActiveTab(tab)}
      className={`px-4 py-2 rounded-lg transition-colors ${
        activeTab === tab
          ? 'bg-blue-500 text-white'
          : 'text-gray-400 hover:bg-gray-800/50'
      }`}
    >
      {label}
    </button>
  )

  const renderTabContent = () => {
    if (isLoading) {
      return (
        <div className="flex items-center justify-center p-8">
          <div className="text-gray-400">Loading...</div>
        </div>
      )
    }

    if (error) {
      return (
        <div className="flex items-center justify-center p-8">
          <div className="text-red-400">{error}</div>
        </div>
      )
    }

    switch (activeTab) {
      case 'posts':
        return (
          <div className="space-y-6">
            {posts.length > 0 ? (
              posts.map((post) => (
                <motion.div
                  key={post.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  className="bg-white/10 backdrop-blur-lg rounded-lg p-6"
                >
                  <h2 className="text-xl font-semibold text-gray-200 mb-2">
                    {post.title}
                  </h2>
                  <p className="text-gray-300 mb-4">{post.content}</p>
                  {post.media && (
                    <img
                      src={post.media}
                      alt="Post media"
                      className="rounded-lg max-h-96 object-cover mb-4"
                    />
                  )}
                  <div className="text-sm text-gray-400">
                    Posted on {new Date(post.created_at).toLocaleString()}
                  </div>
                </motion.div>
              ))
            ) : (
              <div className="text-center text-gray-400">No posts to display</div>
            )}
          </div>
        )

      case 'followers':
        return (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {followers && followers.length > 0 ? (
              followers.map((follower) => (
                <motion.div
                  key={follower.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  className="bg-white/10 backdrop-blur-lg rounded-lg p-4 flex items-center space-x-4"
                >
                  <div className="relative w-12 h-12">
                    {follower.avatar ? (
                      <img
                        src={follower.avatar}
                        alt={follower.username}
                        className="rounded-full w-full h-full object-cover"
                      />
                    ) : (
                      <div className="w-full h-full bg-gray-700 rounded-full flex items-center justify-center">
                        <User size={24} className="text-gray-400" />
                      </div>
                    )}
                  </div>
                  <div className="flex-1">
                    <p className="text-gray-200 font-medium">{follower.username}</p>
                    <p className="text-gray-400 text-sm">{follower.status}</p>
                  </div>
                </motion.div>
              ))
            ) : (
              <div className="text-center text-gray-400 col-span-2">No followers yet</div>
            )}
          </div>
        )

      case 'following':
        return (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {following && following.length > 0 ? (
              following.map((user) => (
                <motion.div
                  key={user.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  className="bg-white/10 backdrop-blur-lg rounded-lg p-4 flex items-center space-x-4"
                >
                  <div className="relative w-12 h-12">
                    {user.avatar ? (
                      <img
                        src={user.avatar}
                        alt={user.username}
                        className="rounded-full w-full h-full object-cover"
                      />
                    ) : (
                      <div className="w-full h-full bg-gray-700 rounded-full flex items-center justify-center">
                        <User size={24} className="text-gray-400" />
                      </div>
                    )}
                  </div>
                  <div className="flex-1">
                    <p className="text-gray-200 font-medium">{user.username}</p>
                    <p className="text-green-400 text-sm">Following</p>
                  </div>
                </motion.div>
              ))
            ) : (
              <div className="text-center text-gray-400 col-span-2">
                Not following anyone yet
              </div>
            )}
          </div>
        )
    }
  }

  const updatePrivacySettings = async (isPrivate: boolean) => {
    try {
      const response = await fetch('http://localhost:8080/user/privacy', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ is_private: isPrivate }),
      })
      if (response.ok) {
        setProfile(prev => prev ? { ...prev, is_private: isPrivate } : null)
        setIsEditingPrivacy(false)
      }
    } catch (error) {
      console.error('Error updating privacy settings:', error)
    }
  }

  if (!profile) {
    return <div>Profile not found</div>
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
      <Header />
      <div className="flex pt-16">
        <Sidebar />
        <div className="flex-1 overflow-y-auto">
          <div className="max-w-4xl mx-auto px-4 py-8">
            {/* Profile Header */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="bg-white/10 backdrop-blur-lg rounded-lg p-6 mb-6"
            >
              <div className="flex items-start justify-between">
                <div className="flex items-start space-x-6">
                  <div className="relative w-32 h-32">
                    {profile.avatar ? (
                      <img
                        src={profile.avatar}
                        alt={profile.username}
                        className="w-full h-full rounded-full object-cover"
                      />
                    ) : (
                      <div className="w-full h-full bg-gray-700 rounded-full flex items-center justify-center">
                        <User size={48} className="text-gray-400" />
                      </div>
                    )}
                  </div>
                  <div className="flex-1">
                    <h1 className="text-2xl font-bold text-gray-200">
                      {profile.first_name} {profile.last_name} {profile.nickName && `(${profile.nickName})`}
                    </h1>
                    <p className="text-gray-400">@{profile.username}</p>
                    {profile.about_me && (
                      <p className="mt-2 text-gray-300">{profile.about_me}</p>
                    )}
                    <div className="mt-4 flex flex-wrap gap-4 text-sm text-gray-400">
                      <div className="flex items-center">
                        <Calendar className="w-4 h-4 mr-1" />
                        Joined {profile.created_at ? new Date(profile.created_at).toLocaleDateString('en-US', {
                          year: 'numeric',
                          month: 'long',
                          day: 'numeric'
                        }) : 'Not available'}
                      </div>
                      <div className="flex items-center">
                        <Mail className="w-4 h-4 mr-1" />
                        {profile.email}
                      </div>
                      <div className="flex items-center">
                        <Calendar className="w-4 h-4 mr-1 text-blue-400" />
                        Born {profile.date_of_birth ? new Date(profile.date_of_birth).toLocaleDateString('en-US', {
                          year: 'numeric',
                          month: 'long',
                          day: 'numeric'
                        }) : 'Not available'}
                      </div>
                    </div>
                  </div>
                </div>

                <div className="relative">
                  <button
                    onClick={() => setIsEditingPrivacy(!isEditingPrivacy)}
                    className="p-2 bg-gray-700 rounded-full hover:bg-gray-600 transition-colors flex items-center space-x-2"
                  >
                    <Settings size={20} className="text-gray-200" />
                    <span className="text-gray-200">Privacy</span>
                  </button>
                  
                  {isEditingPrivacy && (
                    <div className="absolute right-0 top-12 bg-gray-800 rounded-lg shadow-lg p-4 w-64 z-50">
                      <h3 className="text-gray-200 font-semibold mb-4">Privacy Settings</h3>
                      <div className="space-y-4">
                        <button
                          onClick={() => updatePrivacySettings(true)}
                          className={`w-full flex items-center justify-between p-2 rounded ${
                            profile?.is_private ? 'bg-blue-600' : 'bg-gray-700'
                          } hover:bg-opacity-80 transition-colors`}
                        >
                          <span className="flex items-center text-gray-200">
                            <Lock size={16} className="mr-2" />
                            Private Account
                          </span>
                        </button>
                        <button
                          onClick={() => updatePrivacySettings(false)}
                          className={`w-full flex items-center justify-between p-2 rounded ${
                            !profile?.is_private ? 'bg-blue-600' : 'bg-gray-700'
                          } hover:bg-opacity-80 transition-colors`}
                        >
                          <span className="flex items-center text-gray-200">
                            <Unlock size={16} className="mr-2" />
                            Public Account
                          </span>
                        </button>
                      </div>
                    </div>
                  )}
                </div>
              </div>

              <div className="flex items-center mt-4">
                <h1 className="text-2xl font-bold text-gray-200">
                  {profile?.first_name} {profile?.last_name}
                </h1>
                {profile?.is_private && (
                  <Lock size={16} className="ml-2 text-yellow-500" />
                )}
              </div>
            </motion.div>

            {/* Tabs */}
            <div className="flex space-x-4 mb-6">
              <TabButton tab="posts" label="Posts" />
              <TabButton tab="followers" label="Followers" />
              <TabButton tab="following" label="Following" />
            </div>

            {/* Tab Content */}
            {renderTabContent()}
          </div>
        </div>
      </div>
    </div>
  )
} 