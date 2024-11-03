'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { Heart, MessageCircle, Share2, Home, Users, User, Bell, Search } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'

type PostType = {
  id: number
  author: string
  content: string
  likes: number
  comments: number
  timestamp: string
}

export default function Feed() {
  const [posts, setPosts] = useState<PostType[]>([])

  useEffect(() => {
    // Here we'll fetch real posts from the API
    const fetchPosts = async () => {
      try {
        const response = await fetch('/api/posts')
        const data = await response.json()
        setPosts(data)
      } catch (error) {
        console.log('Error fetching posts:', error)
        // Fallback data in case API fails
        setPosts([
          {
            id: 1,
            author: 'Abdulla Aljuffairi',
            content: 'Just finished working on the frontend of Feed #programming',
            likes: 15,
            comments: 3,
            timestamp: '2h ago'
          },
          {
            id: 2,
            author: 'Ahmed Abdeen',
            content: 'looking for group? Discord!',
            likes: 22,
            comments: 7,
            timestamp: '4h ago'
          }
        ])
      }
    }

    fetchPosts()
  }, [])

  return (
    <div className="flex h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
      <motion.div
        initial={{ opacity: 0, x: -20 }}
        animate={{ opacity: 1, x: 0 }}
        transition={{ duration: 0.5 }}
      >
        <Sidebar />
      </motion.div>
      <div className="flex-1 flex flex-col overflow-hidden">
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
        >
          <Header />
        </motion.div>
        <main className="flex-1 overflow-x-hidden overflow-y-auto">
          <div className="flex max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
            <div className="flex-1 max-w-3xl">
              <AnimatePresence>
                <div className="space-y-4">
                  {posts.map((post, index) => (
                    <motion.div
                      key={post.id}
                      initial={{ opacity: 0, y: 20 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: index * 0.1 }}
                    >
                      <Post post={post} />
                    </motion.div>
                  ))}
                </div>
              </AnimatePresence>
            </div>
            <motion.div
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.5 }}
            >
              <RightSidebar />
            </motion.div>
          </div>
        </main>
      </div>
    </div>
  )
}

function Sidebar() {
  return (
    <div className="bg-white/10 backdrop-blur-lg w-64 h-full shadow-lg border border-gray-700/50">
      <div className="p-4">
        <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-400 via-purple-500 to-pink-500 bg-clip-text text-transparent">
          Reboot Network
        </h1>
      </div>
      <nav className="mt-8">
        {[
          { href: '/feed', icon: Home, label: 'Feed' },
          { href: '/profile', icon: User, label: 'Profile' },
          { href: '/groups', icon: Users, label: 'Groups' },
          { href: '/notifications', icon: Bell, label: 'Notifications' },
          { href: '/chats', icon: MessageCircle, label: 'Chats' },
        ].map((item, index) => (
          <motion.div
            key={item.href}
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: index * 0.1 }}
          >
            <Link href={item.href} className="flex items-center px-3 py-2 text-gray-300 hover:bg-gray-800/50">
              <item.icon className="mr-3" size={20} />
              {item.label}
            </Link>
          </motion.div>
        ))}
      </nav>
    </div>
  )
}

function Header() {
    return (
      <header className="bg-white/10 backdrop-blur-lg shadow-sm border-b border-gray-700/50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4 flex items-center">
          <div className="relative w-96">
            <input
              type="text"
              placeholder="Search..."
              className="pl-10 pr-4 py-2 bg-gray-800/50 border border-gray-700/50 rounded-full focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-300 w-full"
            />
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" size={20} />
          </div>
          <div className="flex items-center gap-6 fixed right-8">
            <button className="text-gray-300 hover:text-white">
              <Bell size={20} />
            </button>
            <div className="w-8 h-8 bg-white rounded-full flex items-center justify-center"> <User size={18} className="text-gray-800" /></div>
          </div>
        </div>
      </header>
    )
  }  

  function Post({ post }: { post: PostType }) {
    return (
      <div className="bg-white/10 backdrop-blur-lg rounded-lg shadow p-6 border border-gray-800/500 w-[1155px] -ml-40">
        <div className="flex items-center mb-4">
          <div className="w-10 h-10 bg-gray-700 rounded-full mr-3"></div>
          <div>
            <h3 className="font-semibold text-gray-200">{post.author}</h3>
            <p className="text-sm text-gray-400">{post.timestamp}</p>
          </div>
        </div>
        <p className="mb-9 mr-10 text-gray-200">{post.content}</p>
        <div className="flex items-center space-x-4 text-gray-400">
          <button className="flex items-center space-x-1 hover:text-pink-500 transition-colors">
            <Heart size={20} />
            <span>{post.likes}</span>
          </button>
          <button className="flex items-center space-x-1 hover:text-blue-500 transition-colors">
            <MessageCircle size={20} />
            <span>{post.comments}</span>
          </button>
          <button className="flex items-center space-x-1 hover:text-green-500 transition-colors">
            <Share2 size={20} />
            <span>Share</span>
          </button>
        </div>
      </div>
    )
  }  

function RightSidebar() {
    return (
      <div className="fixed right-0 top-0 h-screen w-96 pt-20 border-l border-gray-700/50 overflow-y-auto">
        <div className="p-6">
          <div className="bg-white/10 backdrop-blur-lg shadow-lg rounded-lg p-6 border border-gray-700/50">
            <h2 className="text-lg font-semibold mb-6 text-gray-200">Suggested Followers</h2>
            <div className="space-y-6 mb-8">
              {['aaljuffa', 'aiqbal', 'hhelal', 'zfadhel', 'Famohamed'].map((name, index) => (
                <motion.div
                  key={index}
                  initial={{ opacity: 0, x: 20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: index * 0.1 }}
                  className="flex items-center justify-between"
                >
                  <div className="flex items-center flex-1 min-w-0">
                    <div className="w-10 h-10 bg-gray-700 rounded-full mr-3 flex-shrink-0"></div>
                    <p className="font-medium text-gray-300 truncate">{name}</p>
                  </div>
                  <button className="ml-4 px-4 py-1 text-sm text-blue-400 hover:text-blue-300 border border-blue-400/50 rounded-full hover:bg-blue-400/10 transition-colors flex-shrink-0">
                    Follow
                  </button>
                </motion.div>
              ))}
            </div>
  
            <h2 className="text-lg font-semibold mb-6 text-gray-200">Active Groups</h2>
            <div className="space-y-6">
              {[
                { name: 'Football Club', members: '1.2k members' },
                { name: 'Pizza Club', members: '892 members' },
                { name: 'Barcelona club', members: '654 members' }
              ].map((group, index) => (
                <motion.div
                  key={index}
                  initial={{ opacity: 0, x: 20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: (index + 3) * 0.1 }}
                  className="flex items-center justify-between"
                >
                  <div className="flex-1 min-w-0">
                    <p className="font-medium text-gray-300 truncate">{group.name}</p>
                    <p className="text-sm text-gray-400">{group.members}</p>
                  </div>
                  <button className="ml-4 px-4 py-1 text-sm text-green-400 hover:text-green-300 border border-green-400/50 rounded-full hover:bg-green-400/10 transition-colors flex-shrink-0">
                    Join
                  </button>
                </motion.div>
              ))}
            </div>
          </div>
        </div>
      </div>
    )
  }  
  