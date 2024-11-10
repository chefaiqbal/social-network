'use client'

import { useState, useEffect, FormEvent } from 'react'
import Link from 'next/link'
import { Heart, MessageCircle, Share2, Home, Users, User, Bell, Search } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import DropDownCheck from '@/components/ui/DropDownCheck'
import Uploader from '@/components/ui/uploadButton'
import { ChatList } from '@/components/chat/ChatList'
import { Toaster } from 'react-hot-toast'


// Create Post Component
function CreatePost({ onPostCreated }: { onPostCreated: () => void }) {
  const [content, setContent] = useState('');
  const [title, setTitle] = useState('');
  const [privacy, setPrivacy] = useState('1'); // default privacy is '1' (Public)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();

    try {
      const privacyInt = parseInt(privacy, 10);
      if (isNaN(privacyInt)) {
        console.error('Invalid privacy value:', privacy);
        return;
      }

      const response = await fetch('http://localhost:8080/posts', {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          title,
          content,
          privacy: privacyInt,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to create post');
      }

      // Clear form
      setContent('');
      setTitle('');
      setPrivacy('1');
      
      // Refresh posts
      onPostCreated();
    } catch (error) {
      console.error('Error creating post:', error);
    }
  };

  return (
    <div className="bg-white/10 backdrop-blur-lg rounded-lg shadow p-6 border border-gray-800/500 w-[1155px] -ml-40 mb-4">
      <form onSubmit={handleSubmit}>
        <div>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Post title..."
            className="w-full bg-gray-800/50 border border-gray-700/50 rounded-lg p-2 mb-2 text-gray-200"
          />

          <div className="mb-2">
            <select
              name="category"
              id="category"
              value={privacy}
              onChange={(e) => setPrivacy(e.target.value)}
              className="w-full bg-gray-800/50 border border-gray-700/50 rounded-lg p-2 text-gray-200"
            >
              <option value="1">Public</option>
              <option value="2">Follower</option>
              <option value="3">Close friend</option>
            </select>
          </div>

        </div>
        <textarea
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="What's on your mind?"
          className="w-full bg-gray-800/50 border border-gray-700/50 rounded-lg p-2 mb-2 text-gray-200"
          rows={3}
        />
      <div className="flex items-center space-x-4">
        <button
          type="submit"
          className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg transition-colors"
        >
          Post
        </button>
        <Uploader/>
      </div>
      </form>
    </div>
  );
}



// Post Component
function Post({ post }: { post: PostType }) {

  

  return (
    <div className="bg-white/10 backdrop-blur-lg rounded-lg shadow p-6 border border-gray-800/500 w-[1155px] -ml-40">
      <div className="flex items-center mb-4">
        <div className="w-10 h-10 bg-gray-700 rounded-full mr-3 overflow-hidden">
          {post.author_avatar ? (
            <img 
              src={post.author_avatar} 
              alt={post.author_name} 
              className="w-full h-full object-cover"
            />
          ) : (
            <div className="w-full h-full flex items-center justify-center bg-gray-600">
              <User size={20} className="text-gray-300" />
            </div>
          )}
        </div>
        <div>
          <h3 className="font-semibold text-gray-200">{post.author_name}</h3>
          <p className="text-sm text-gray-400">
            {new Date(post.created_at).toLocaleString()}
          </p>
        </div>
      </div>
      <h2 className="text-xl font-semibold text-gray-200 mb-2">{post.title}</h2>
      <p className="mb-9 mr-10 text-gray-200">{post.content}</p>
      {post.media && (
        <img 
          src={post.media} 
          alt="Post media" 
          className="mb-4 rounded-lg max-h-96 object-cover"
        />
      )}
      <div className="flex items-center space-x-4 text-gray-400">
        <button className="flex items-center space-x-1 hover:text-pink-500 transition-colors">
          <Heart size={20} />
          <span>0</span>
        </button>
        <button className="flex items-center space-x-1 hover:text-blue-500 transition-colors">
          <MessageCircle size={20} />
          <span>0</span>
        </button>
        <button className="flex items-center space-x-1 hover:text-green-500 transition-colors">
          <Share2 size={20} />
          <span>Share</span>
        </button>
      </div>
    </div>
  )
}

// Update the Sidebar navigation items
const navItems = [
  { href: '/feed', icon: Home, label: 'Feed' },
  { href: '/profile', icon: User, label: 'Profile' },
  { href: '/groups', icon: Users, label: 'Groups' },
  { href: '/notifications', icon: Bell, label: 'Notifications' },
  // Removed the Chats link since we have the chat panel
];

// Sidebar Component
export const Sidebar = () => {
  return (
    <div className="bg-white/10 backdrop-blur-lg w-64 h-full shadow-lg border border-gray-700/50">
      <div className="p-4">
        <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-400 via-purple-500 to-pink-500 bg-clip-text text-transparent">
          Reboot Network
        </h1>
      </div>
      <nav className="mt-8">
        {navItems.map((item, index) => (
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

// Header Component
const Header = () => {
  const handleSignOut = async () => {
    try {
      const response = await fetch('http://localhost:8080/logout', {
        method: 'POST',
        credentials: 'include',
      })

      if (response.ok) {
        // Redirect to login page after successful logout
        window.location.href = '/'
      } else {
        console.error('Failed to sign out')
      }
    } catch (error) {
      console.error('Error signing out:', error)
    }
  }

  return (
    <header className="bg-white/10 backdrop-blur-lg shadow-sm border-b border-gray-700/50 relative z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4 flex items-center justify-between">
        <div className="relative w-96">
          <input
            type="text"
            placeholder="Search..."
            className="pl-10 pr-4 py-2 bg-gray-800/50 border border-gray-700/50 rounded-full focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-300 w-full"
          />
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" size={20} />
        </div>
        <div className="flex items-center gap-6">
          <button className="text-gray-300 hover:text-white">
            <Bell size={20} />
          </button>
          <div className="relative group">
            <button className="w-8 h-8 bg-white rounded-full flex items-center justify-center hover:ring-2 hover:ring-blue-500 transition-all">
              <User size={18} className="text-gray-800" />
            </button>
            <div className="absolute right-0 mt-2 w-48 rounded-md shadow-lg bg-white/10 backdrop-blur-lg border border-gray-700/50 opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 transform origin-top-right">
              <div className="py-1">
                <Link 
                  href="/profile" 
                  className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-800/50 transition-colors"
                >
                  Profile
                </Link>
                <button
                  onClick={handleSignOut}
                  className="block w-full text-left px-4 py-2 text-sm text-red-400 hover:bg-gray-800/50 transition-colors"
                >
                  Sign out
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </header>
  )
}

// RightSidebar Component
const RightSidebar = () => {
  const [suggestedUsers, setSuggestedUsers] = useState([])
  const [groups, setGroups] = useState([])

  const handleFollow = async (userId: number) => {
    try {
      const response = await fetch('http://localhost:8080/follow', {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          follower_id: userId, // The logged-in user's ID
          followed_id: userId, // The user to follow
        }),
      })

      if (!response.ok) {
        throw new Error('Failed to follow user')
      }

      // Refresh the suggested users list
      fetchSuggestedUsers()
    } catch (error) {
      console.error('Error following user:', error)
    }
  }

  const fetchSuggestedUsers = async () => {
    try {
      const response = await fetch('http://localhost:8080/users/suggested', {
        credentials: 'include',
      })
      if (response.ok) {
        const data = await response.json()
        setSuggestedUsers(data)
      }
    } catch (error) {
      console.error('Error fetching suggested users:', error)
    }
  }

  const fetchGroups = async () => {
    try {
      const response = await fetch('http://localhost:8080/groups', {
        credentials: 'include',
      })
      if (response.ok) {
        const data = await response.json()
        setGroups(data)
      }
    } catch (error) {
      console.error('Error fetching groups:', error)
    }
  }

  useEffect(() => {
    fetchSuggestedUsers()
    fetchGroups()
  }, [])

  return (
    <div className="fixed right-0 top-0 h-screen w-96 pt-20 border-l border-gray-700/50 overflow-y-auto">
      <div className="p-6">
        <div className="bg-white/10 backdrop-blur-lg shadow-lg rounded-lg p-6 border border-gray-700/50">
          <h2 className="text-lg font-semibold mb-6 text-gray-200">Suggested Followers</h2>
          <div className="space-y-6 mb-8">
            {suggestedUsers.map((user: any, index) => (
              <motion.div
                key={user.id}
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: index * 0.1 }}
                className="flex items-center justify-between"
              >
                <div className="flex items-center flex-1 min-w-0">
                  <div className="w-10 h-10 bg-gray-700 rounded-full mr-3 flex-shrink-0">
                    {user.avatar && (
                      <img 
                        src={user.avatar} 
                        alt={user.username} 
                        className="w-full h-full rounded-full object-cover"
                      />
                    )}
                  </div>
                  <p className="font-medium text-gray-300 truncate">{user.username}</p>
                </div>
                <button 
                  onClick={() => handleFollow(user.id)}
                  className="ml-4 px-4 py-1 text-sm text-blue-400 hover:text-blue-300 border border-blue-400/50 rounded-full hover:bg-blue-400/10 transition-colors flex-shrink-0"
                >
                  Follow
                </button>
              </motion.div>
            ))}
          </div>

          <h2 className="text-lg font-semibold mb-6 text-gray-200">Active Groups</h2>
          <div className="space-y-6">
            {groups.map((group: any, index) => (
              <motion.div
                key={group.id}
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: (index + 3) * 0.1 }}
                className="flex items-center justify-between"
              >
                <div className="flex-1 min-w-0">
                  <p className="font-medium text-gray-300 truncate">{group.title}</p>
                  <p className="text-sm text-gray-400">{group.member_count} members</p>
                </div>
                <button className="ml-4 px-4 py-1 text-sm text-green-400 hover:text-green-300 border border-green-400/50 rounded-full hover:bg-green-400/10 transition-colors flex-shrink-0">
                  Join
                </button>
              </motion.div>
            ))}
          </div>
        </div>
        <div className="mt-6">
          <DropDownCheck />
        </div>
      </div>
    </div>
  )
}
// Feed Component (Main Component)
const Feed = () => {
  const [posts, setPosts] = useState<PostType[]>([])

  const fetchPosts = async () => {
    try {
      const response = await fetch('http://localhost:8080/posts', {
        credentials: 'include',
      })
      if (!response.ok) {
        throw new Error('Failed to fetch posts')
      }
      const data = await response.json()
      setPosts(data)
    } catch (error) {
      console.error('Error fetching posts:', error)
    }
  }

  useEffect(() => {
    fetchPosts()
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
                <CreatePost onPostCreated={fetchPosts} />
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
        <ChatList />
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
}

// Export the Feed component as default
export default Feed
