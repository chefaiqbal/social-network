'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/router'
import { motion } from 'framer-motion'
import { Settings, MapPin, Calendar, Lock, Globe, Edit2 } from 'lucide-react'
import Link from 'next/link'

type UserProfile = {
  id: string
  firstName: string
  lastName: string
  username: string
  email: string
  dateOfBirth: string
  avatar?: string
  bio?: string
  location?: string
  isPrivate: boolean
  joinedDate: string
  followers: number
  following: number
}

type UserPost = {
  id: number
  content: string
  likes: number
  comments: number
  timestamp: string
}

type FollowerType = {
  id: string
  username: string
  avatar?: string
}

export default function Profile() {
  const router = useRouter()
  const [profile, setProfile] = useState<UserProfile | null>(null)
  const [posts, setPosts] = useState<UserPost[]>([])
  const [followers, setFollowers] = useState<FollowerType[]>([])
  const [following, setFollowing] = useState<FollowerType[]>([])
  const [activeTab, setActiveTab] = useState<'posts' | 'followers' | 'following'>('posts')
  const [isOwnProfile, setIsOwnProfile] = useState(false)
  const [isEditing, setIsEditing] = useState(false)
  const [isFollowing, setIsFollowing] = useState(false)
  const [canViewProfile, setCanViewProfile] = useState(true)

  useEffect(() => {
    // Fetch profile data
    const fetchProfile = async () => {
      try {
        // In real implementation, use the actual API endpoint
        // const response = await fetch(`/api/profile/${router.query.username}`)
        // const data = await response.json()
        
        // Dummy data for now
        setProfile({
          id: '1',
          firstName: 'John',
          lastName: 'Doe',
          username: 'johndoe',
          email: 'john@example.com',
          dateOfBirth: '1990-01-01',
          bio: 'Software developer | Coffee enthusiast',
          location: 'Dubai, UAE',
          isPrivate: false,
          joinedDate: 'January 2024',
          followers: 150,
          following: 120
        })

        // Set dummy posts
        setPosts([
          {
            id: 1,
            content: 'Just launched my new project! #coding',
            likes: 25,
            comments: 10,
            timestamp: '2h ago'
          },
          // Add more dummy posts...
        ])

        // Determine if this is the user's own profile
        setIsOwnProfile(true) // This should be based on actual auth state
      } catch (error) {
        console.error('Error fetching profile:', error)
      }
    }

    fetchProfile()
  }, [router.query.username])

  const handlePrivacyToggle = async () => {
    if (!profile) return
    
    try {
      // In real implementation, make API call to update privacy setting
      // await fetch('/api/profile/privacy', {
      //   method: 'PUT',
      //   body: JSON.stringify({ isPrivate: !profile.isPrivate })
      // })
      
      setProfile(prev => prev ? { ...prev, isPrivate: !prev.isPrivate } : null)
    } catch (error) {
      console.error('Error updating privacy:', error)
    }
  }

  const handleFollow = async () => {
    try {
      // In real implementation, make API call to follow/unfollow
      // await fetch('/api/follow', {
      //   method: 'POST',
      //   body: JSON.stringify({ userId: profile?.id })
      // })
      
      setIsFollowing(prev => !prev)
    } catch (error) {
      console.error('Error following user:', error)
    }
  }

  if (!profile) return <div>Loading...</div>

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Profile Header */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-white/10 backdrop-blur-lg rounded-lg shadow-lg p-6 mb-6"
        >
          <div className="flex items-start justify-between">
            <div className="flex items-center space-x-6">
              <div className="w-32 h-32 rounded-full bg-gray-700">
                {profile.avatar && (
                  <img
                    src={profile.avatar}
                    alt={profile.username}
                    className="w-full h-full rounded-full object-cover"
                  />
                )}
              </div>
              <div>
                <h1 className="text-2xl font-bold text-gray-200">
                  {profile.firstName} {profile.lastName}
                </h1>
                <p className="text-gray-400">@{profile.username}</p>
                <div className="flex items-center space-x-4 mt-2 text-gray-300">
                  <span className="flex items-center">
                    <MapPin size={16} className="mr-1" />
                    {profile.location}
                  </span>
                  <span className="flex items-center">
                    <Calendar size={16} className="mr-1" />
                    Joined {profile.joinedDate}
                  </span>
                </div>
                {profile.bio && (
                  <p className="mt-4 text-gray-300">{profile.bio}</p>
                )}
              </div>
            </div>
            <div className="flex items-center space-x-4">
              {isOwnProfile ? (
                <>
                  <button
                    onClick={() => setIsEditing(true)}
                    className="flex items-center px-4 py-2 rounded-full border border-gray-600 text-gray-300 hover:bg-gray-800 transition-colors"
                  >
                    <Edit2 size={16} className="mr-2" />
                    Edit Profile
                  </button>
                  <button
                    onClick={handlePrivacyToggle}
                    className="flex items-center px-4 py-2 rounded-full border border-gray-600 text-gray-300 hover:bg-gray-800 transition-colors"
                  >
                    {profile.isPrivate ? <Lock size={16} /> : <Globe size={16} />}
                    <span className="ml-2">
                      {profile.isPrivate ? 'Private' : 'Public'}
                    </span>
                  </button>
                </>
              ) : (
                <button
                  onClick={handleFollow}
                  className={`px-6 py-2 rounded-full ${
                    isFollowing
                      ? 'border border-blue-500 text-blue-500'
                      : 'bg-blue-500 text-white'
                  }`}
                >
                  {isFollowing ? 'Following' : 'Follow'}
                </button>
              )}
            </div>
          </div>

          <div className="flex justify-center space-x-8 mt-8">
            <button
              onClick={() => setActiveTab('posts')}
              className={`text-lg ${
                activeTab === 'posts'
                  ? 'text-blue-500 border-b-2 border-blue-500'
                  : 'text-gray-400'
              }`}
            >
              Posts
            </button>
            <button
              onClick={() => setActiveTab('followers')}
              className={`text-lg ${
                activeTab === 'followers'
                  ? 'text-blue-500 border-b-2 border-blue-500'
                  : 'text-gray-400'
              }`}
            >
              Followers ({profile.followers})
            </button>
            <button
              onClick={() => setActiveTab('following')}
              className={`text-lg ${
                activeTab === 'following'
                  ? 'text-blue-500 border-b-2 border-blue-500'
                  : 'text-gray-400'
              }`}
            >
              Following ({profile.following})
            </button>
          </div>
        </motion.div>

        {/* Profile Content */}
        {(!profile.isPrivate || isOwnProfile || isFollowing) ? (
          <div className="space-y-6">
            {activeTab === 'posts' && (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="space-y-4"
              >
                {posts.map((post) => (
                  <div
                    key={post.id}
                    className="bg-white/10 backdrop-blur-lg rounded-lg p-6"
                  >
                    <p className="text-gray-200">{post.content}</p>
                    <div className="flex items-center space-x-4 mt-4 text-gray-400">
                      <span>{post.likes} likes</span>
                      <span>{post.comments} comments</span>
                      <span>{post.timestamp}</span>
                    </div>
                  </div>
                ))}
              </motion.div>
            )}

            {activeTab === 'followers' && (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="grid grid-cols-1 md:grid-cols-2 gap-4"
              >
                {/* Followers list */}
              </motion.div>
            )}

            {activeTab === 'following' && (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="grid grid-cols-1 md:grid-cols-2 gap-4"
              >
                {/* Following list */}
              </motion.div>
            )}
          </div>
        ) : (
          <div className="text-center py-12">
            <Lock className="mx-auto h-12 w-12 text-gray-400" />
            <h3 className="mt-4 text-lg font-medium text-gray-300">
              This profile is private
            </h3>
            <p className="mt-2 text-gray-400">
              Follow this user to see their posts and activity
            </p>
          </div>
        )}
      </div>
    </div>
  )
} 