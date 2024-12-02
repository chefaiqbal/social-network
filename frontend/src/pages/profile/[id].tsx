import { useState, useEffect } from 'react'
import { useRouter } from 'next/router'
import Header from '@/components/layout/Header'
import Sidebar from '@/components/layout/Sidebar'
import { User as UserIcon, Mail, Calendar, Users, MessageCircle, Lock, Unlock, Settings } from 'lucide-react'
import UnfollowButton from '@/components/ui/Unfollow'

interface Profile {
  id: number
  username: string
  first_name: string
  last_name: string
  email: string
  date_of_birth: string
  about_me: string
  avatar: string
  is_private: boolean
  is_following: boolean
  is_pending: boolean
  posts: Post[]
  followers_count: number
  following_count: number
}

interface Post {
  id: number
  title: string
  content: string
  created_at: string
  likes_count: number
  comments_count: number
}

export default function UserProfile() {
  const router = useRouter()
  const { id } = router.query
  const [profile, setProfile] = useState<Profile | null>(null)
  const [loading, setLoading] = useState(true)
  const [isEditingPrivacy, setIsEditingPrivacy] = useState(false)
  const [isCurrentUser, setIsCurrentUser] = useState(false)
  const [media, setMedia] = useState<string | null>(null); // New state for media

  useEffect(() => {
    if (id) {
      fetchProfile()
    }
  }, [id])

  useEffect(() => {
    const checkCurrentUser = async () => {
      try {
        const response = await fetch('http://localhost:8080/userIDBY', {
          credentials: 'include',
        })
        if (response.ok) {
          const data = await response.json()
          setIsCurrentUser(data.userID === Number(id))
        }
      } catch (error) {
        console.error('Error checking current user:', error)
      }
    }

    if (id) {
      checkCurrentUser()
    }
  }, [id])

  const fetchProfile = async () => {
    try {
      const response = await fetch(`http://localhost:8080/user/${id}`, {
        credentials: 'include',
      })
      if (response.ok) {
        const data = await response.json()
        setProfile(data)
      } else {
        console.error('Failed to fetch profile')
      }
    } catch (error) {
      console.error('Error:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleFollow = async () => {
    try {
      const response = await fetch('http://localhost:8080/follow', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          followed_id: profile?.id,
          status: 'pending',
        }),
      })
      if (response.ok) {
        fetchProfile() // Refresh profile data
      }
    } catch (error) {
      console.error('Error following user:', error)
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

  if (loading) {
    return <div>Loading...</div>
  }

  if (!profile) {
    return <div>Profile not found</div>
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
      <Header />
      <div className="flex pt-16">
        <Sidebar />
        <div className="flex-1 p-8">
          <div className="max-w-4xl mx-auto bg-gray-800/50 backdrop-blur-lg rounded-lg shadow-lg overflow-hidden">
            {/* Profile Header */}
            <div className="relative h-48 bg-gradient-to-r from-blue-600 to-blue-800">
              <div className="absolute -bottom-16 left-8">
                {profile.avatar ? (
                  <img
                    src={profile.avatar}
                    alt={profile.username}
                    className="w-32 h-32 rounded-full border-4 border-gray-800"
                  />
                ) : (
                  <div className="w-32 h-32 rounded-full border-4 border-gray-800 bg-gray-700 flex items-center justify-center">
                    <UserIcon size={64} className="text-gray-400" />
                  </div>
                )}
              </div>
            </div>

            {/* l need to make sure it only appear to user  already follow */}

            {/* unfollow button */}
            <UnfollowButton fetchProfile={fetchProfile} followed_id={profile.id} />

            {isCurrentUser && (
              <div className="absolute top-4 right-4 flex items-center space-x-2">
                <button
                  onClick={() => setIsEditingPrivacy(!isEditingPrivacy)}
                  className="p-2 bg-gray-700 rounded-full hover:bg-gray-600 transition-colors"
                >
                  <Settings size={20} className="text-gray-200" />
                </button>
                
                {isEditingPrivacy && (
                  <div className="absolute right-0 top-12 bg-gray-800 rounded-lg shadow-lg p-4 w-64">
                    <h3 className="text-gray-200 font-semibold mb-4">Privacy Settings</h3>
                    <div className="space-y-4">
                      <button
                        onClick={() => updatePrivacySettings(true)}
                        className={`w-full flex items-center justify-between p-2 rounded ${
                          profile?.is_private ? 'bg-blue-600' : 'bg-gray-700'
                        }`}
                      >
                        <span className="flex items-center">
                          <Lock size={16} className="mr-2" />
                          Private Account
                        </span>
                      </button>
                      <button
                        onClick={() => updatePrivacySettings(false)}
                        className={`w-full flex items-center justify-between p-2 rounded ${
                          !profile?.is_private ? 'bg-blue-600' : 'bg-gray-700'
                        }`}
                      >
                        <span className="flex items-center">
                          <Unlock size={16} className="mr-2" />
                          Public Account
                        </span>
                      </button>
                    </div>
                  </div>
                )}
              </div>
            )}

            {/* Profile Info */}
            <div className="pt-20 px-8 pb-8">
              <div className="flex items-center space-x-2">
                <h1 className="text-3xl font-bold text-gray-200">
                  {profile.first_name} {profile.last_name}
                </h1>
                {profile.is_private && (
                  <Lock size={16} className="text-yellow-500" />
                )}
              </div>
              <p className="text-gray-400">@{profile.username}</p>

              {/* Stats */}
              <div className="flex space-x-6 mb-6">
                <div className="flex items-center text-gray-300">
                  <Users className="mr-2" size={20} />
                  <span>{profile.followers_count} followers</span>
                </div>
                <div className="flex items-center text-gray-300">
                  <Users className="mr-2" size={20} />
                  <span>{profile.following_count} following</span>
                </div>
              </div>

              {/* About */}
              <div className="mb-6">
                <h2 className="text-xl font-semibold text-gray-200 mb-2">About</h2>
                <p className="text-gray-300">{profile.about_me}</p>
              </div>

              {/* Contact Info */}
              <div className="grid grid-cols-2 gap-4 mb-8">
                <div className="flex items-center text-gray-300">
                  <Mail className="mr-2" size={20} />
                  <span>{profile.email}</span>
                </div>
                <div className="flex items-center text-gray-300">
                  <Calendar className="mr-2" size={20} />
                  <span>{new Date(profile.date_of_birth).toLocaleDateString()}</span>
                </div>
              </div>

              {/* Posts */}
              {(!profile.is_private || profile.is_following) && profile.posts && (
                <div>
                  <h2 className="text-xl font-semibold text-gray-200 mb-4">Recent Posts</h2>
                  <div className="space-y-4">
                    {profile.posts.map(post => (
                      <div key={post.id} className="bg-gray-700/50 rounded-lg p-4">
                        <h3 className="text-lg font-semibold text-gray-200 mb-2">{post.title}</h3>
                        <p className="text-gray-300 mb-4">{post.content}</p>
                        <div className="flex items-center text-gray-400 text-sm">
                          <span className="mr-4">{new Date(post.created_at).toLocaleDateString()}</span>
                          <span className="mr-4">❤️ {post.likes_count}</span>
                          <span>💬 {post.comments_count}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
} 