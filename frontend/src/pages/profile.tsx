'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/router'
import { User, MapPin, Calendar, Mail, Link as LinkIcon } from 'lucide-react'
import { motion } from 'framer-motion'
import { Sidebar } from './feed'

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

export default function Profile() {
  const [profile, setProfile] = useState<UserProfile | null>(null)
  const [posts, setPosts] = useState<Post[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const router = useRouter()

  useEffect(() => {
    const fetchProfile = async () => {
      try {
        const userId = router.query.id || 'current'
        const response = await fetch(`http://localhost:8080/user/${userId}`, {
          credentials: 'include',
        })
        if (response.ok) {
          const data = await response.json()
          setProfile(data)
        }
      } catch (error) {
        console.error('Error fetching profile:', error)
      }
    }

    const fetchUserPosts = async () => {
      try {
        const userId = router.query.id || 'current'
        const response = await fetch(`http://localhost:8080/posts/user/${userId}`, {
          credentials: 'include',
        })
        if (response.ok) {
          const data = await response.json()
          setPosts(Array.isArray(data) ? data : [])
        }
      } catch (error) {
        console.error('Error fetching posts:', error)
        setPosts([])
      } finally {
        setIsLoading(false)
      }
    }

    if (router.isReady) {
      fetchProfile()
      fetchUserPosts()
    }
  }, [router.isReady, router.query.id])

  if (isLoading) {
    return <div>Loading...</div>
  }

  if (!profile) {
    return <div>Profile not found</div>
  }

  return (
    <div className="flex h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
      <Sidebar />
      <div className="flex-1 overflow-y-auto">
        <div className="max-w-4xl mx-auto px-4 py-8">
          {/* Profile Header */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="bg-white/10 backdrop-blur-lg rounded-lg p-6 mb-6"
          >
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
                  {profile.first_name} {profile.last_name}
                </h1>
                <p className="text-gray-400">@{profile.username}</p>
                {profile.about_me && (
                  <p className="mt-2 text-gray-300">{profile.about_me}</p>
                )}
                <div className="mt-4 flex flex-wrap gap-4 text-sm text-gray-400">
                  <div className="flex items-center">
                    <Calendar className="w-4 h-4 mr-1" />
                    Joined {new Date(profile.created_at).toLocaleDateString()}
                  </div>
                  <div className="flex items-center">
                    <Mail className="w-4 h-4 mr-1" />
                    {profile.email}
                  </div>
                </div>
              </div>
            </div>
          </motion.div>

          {/* User Posts */}
          <div className="space-y-6">
            {Array.isArray(posts) && posts.length > 0 ? (
              posts.map((post, index) => (
                <motion.div
                  key={post.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: index * 0.1 }}
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
              <div className="text-center text-gray-400">
                No posts to display
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
} 