'use client'

import { useState, useEffect, FormEvent, useRef, useCallback } from 'react'
import { useRouter } from 'next/router'
import { motion, AnimatePresence } from 'framer-motion'
import { ChatList } from '@/components/chat/ChatList'
import { Toaster } from 'react-hot-toast'
import Header from '@/components/layout/Header'
import Sidebar from '@/components/layout/Sidebar'
import CreatePost from '@/components/layout/createPost'
import Post from '@/components/layout/Post'
import DropDownCheck from '@/components/ui/DropDownCheck'

// Feed Component (Main Component)
const Feed = () => {
  const [posts, setPosts] = useState<PostType[]>([]) // Initialize with empty array
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false)
  const [isLoading, setIsLoading] = useState<boolean>(true) // Add loading state
  const router = useRouter()

  const fetchPosts = async () => {
    try {
      setIsLoading(true) // Set loading true before fetch
      const response = await fetch('http://localhost:8080/posts', {
        credentials: 'include',
      })
      if (!response.ok) {
        throw new Error('Failed to fetch posts')
      }
      const data = await response.json()
      setPosts(data || []) // Ensure we always set an array
    } catch (error) {
      console.error('Error fetching posts:', error)
      setPosts([]) // Reset to empty array on error
    } finally {
      setIsLoading(false) // Set loading false after fetch
    }
  }

  const checkAuthentication = async () => {
    try {
      const response = await fetch('http://localhost:8080/checkAuth', {
        credentials: 'include', 
      })

      if (response.ok) {
        setIsAuthenticated(true)
        fetchPosts() 
      } else {
        setIsAuthenticated(false)
        router.push('/') 
      }
    } catch (error) {
      console.error('Error checking authentication:', error)
      setIsAuthenticated(false)
      router.push('/') 
    }
  }


  useEffect(() => {
    checkAuthentication();
  }, [])

  return (
    <>
      <Toaster 
        position="top-right"
        toastOptions={{
          duration: 3000,
          style: {
            background: '#1f2937',
            color: '#fff',
          },
          success: {
            iconTheme: {
              primary: '#10B981',
              secondary: '#fff',
            },
          },
          error: {
            iconTheme: {
              primary: '#EF4444',
              secondary: '#fff',
            },
          },
        }}
      />
      <div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
        <Header />
        <div className="flex pt-16">
          <Sidebar />
          <div className="flex-1 overflow-hidden">
            <main className="overflow-y-auto h-[calc(100vh-64px)]">
              <div className="flex max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                <div className="flex-1">
                      <div className="flex items-start space-x-4">
                      <div className="flex-1">
                        <CreatePost onPostCreated={fetchPosts} />
                      </div>
                      <DropDownCheck />
                      </div>
                      <div className="mt-8">
                      <h2 className="text-2xl font-bold text-white">Recent Posts</h2>
                      </div>
                  <AnimatePresence>
                    <div className="space-y-4">
                      {isLoading ? (
                        <div className="text-white text-center">Loading posts...</div>
                      ) : posts && posts.length > 0 ? (
                        posts.map((post, index) => (
                          <motion.div
                            key={post.id}
                            initial={{ opacity: 0, y: 20 }}
                            animate={{ opacity: 1, y: 0 }}
                            transition={{ delay: index * 0.1 }}
                          >
                            <Post post={post} />
                          </motion.div>
                        ))
                      ) : (
                        <div className="text-white text-center">No posts found</div>
                      )}
                    </div>
                  </AnimatePresence>
                </div>
              </div>
            </main>
          </div>
          <ChatList />
        </div>
      </div>
    </>
  )
}

// Add PostType interface
interface PostType {
  id: number
  title: string
  content: string
  media?: string
  privacy: number
  author: number
  author_name: string
  author_avatar?: string
  created_at: string
  group_id?: number
  like_count?: number
  user_liked?: boolean
}

// Add this interface for comments
interface Comment {
  id: number;
  content: string;
  media?: string;
  media_type?: string;
  post_id: number;
  author: number;
  author_name?: string;
  author_avatar?: string;
  created_at: string;
}

// Export the Feed component as default
export default Feed
