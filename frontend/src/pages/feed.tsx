'use client'

import { useState, useEffect, FormEvent, useRef, useCallback } from 'react'
import Link from 'next/link'
import { ThumbsUp, MessageCircle, Share2, User } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import DropDownCheck from '@/components/ui/DropDownCheck'
import Uploader from '@/components/ui/uploadButton'
import { ChatList } from '@/components/chat/ChatList'
import { Toaster } from 'react-hot-toast'
import Header from '@/components/layout/Header'
import Sidebar from '@/components/layout/Sidebar'

// Create Post Component
interface CreatePostProps {
  onPostCreated: () => void;
}

function CreatePost({ onPostCreated }: CreatePostProps) {
  const [content, setContent] = useState('');
  const [title, setTitle] = useState('');
  const [privacy, setPrivacy] = useState('1'); // default privacy is '1' (Public)
  const [media, setMedia] = useState<string | null>(null); // New state for media

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    console.log('Media before submit:', media); // Confirm media has base64
  
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
          media, // Send the base64 string as media
        }),
      });
  
      if (!response.ok) {
        throw new Error('Failed to create post');
      }
  
      setContent('');
      setTitle('');
      setPrivacy('1');
      setMedia(null);
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
          <Uploader onUpload={(base64: string) => {
            console.log('Received base64 from Uploader:', base64); // Log base64 in CreatePost
            setMedia(base64);
          }} />
        </div>
      </form>
    </div>
  );
}


// Post Component
function Post({ post }: { post: PostType }) {
    const [likeCount, setLikeCount] = useState(post.like_count || 0);
    const [userLiked, setUserLiked] = useState(post.user_liked || false);
    const [isLoading, setIsLoading] = useState(false);
    const wsRef = useRef<WebSocket | null>(null);
    const reconnectTimeoutRef = useRef<NodeJS.Timeout>();

    const connectWebSocket = useCallback(() => {
        if (wsRef.current?.readyState === WebSocket.OPEN) {
            return;
        }

        wsRef.current = new WebSocket('ws://localhost:8080/ws/likes');
        
        wsRef.current.onopen = () => {
            console.log('Like WebSocket connected');
        };

        wsRef.current.onmessage = (event) => {
            try {
                const update = JSON.parse(event.data);
                if (update.post_id === post.id) {
                    setLikeCount(update.like_count);
                    // Update the heart fill state for all users
                    const currentUserId = parseInt(localStorage.getItem('userId') || '0');
                    if (update.user_id === currentUserId) {
                        setUserLiked(update.user_liked);
                    }
                }
            } catch (error) {
                console.error('Error processing WebSocket message:', error);
            }
        };

        wsRef.current.onclose = () => {
            console.log('Like WebSocket closed, attempting to reconnect...');
            reconnectTimeoutRef.current = setTimeout(connectWebSocket, 3000);
        };

        wsRef.current.onerror = (error) => {
            console.error('Like WebSocket error:', error);
        };
    }, [post.id]);

    useEffect(() => {
        // Get current user ID from localStorage when component mounts
        const userId = localStorage.getItem('userId');
        if (!userId) {
            fetch('http://localhost:8080/userIDBY', {
                credentials: 'include',
            })
            .then(res => res.json())
            .then(data => {
                localStorage.setItem('userId', data.userID.toString());
            })
            .catch(console.error);
        }

        connectWebSocket();

        return () => {
            if (reconnectTimeoutRef.current) {
                clearTimeout(reconnectTimeoutRef.current);
            }
            if (wsRef.current) {
                wsRef.current.close();
            }
        };
    }, [connectWebSocket]);

    const handleLike = async () => {
        if (isLoading) return;
        setIsLoading(true);

        try {
            const response = await fetch('http://localhost:8080/likes', {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    post_id: post.id,
                    is_like: !userLiked,
                }),
            });

            if (!response.ok) {
                throw new Error('Failed to update like');
            }

            // Let the WebSocket handle the state update
        } catch (error) {
            console.error('Error liking post:', error);
        } finally {
            setIsLoading(false);
        }
    };

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
                <motion.button 
                    onClick={handleLike}
                    className={`flex items-center space-x-1 transition-colors ${
                        userLiked ? 'text-blue-500' : 'hover:text-blue-500'
                    }`}
                    disabled={isLoading}
                    whileTap={{ scale: 0.95 }}
                >
                    <ThumbsUp 
                        size={20} 
                        fill={userLiked ? 'currentColor' : 'none'} 
                        className={`transition-all duration-200 ${isLoading ? 'opacity-50' : ''}`}
                    />
                    <motion.span
                        key={likeCount}
                        initial={{ y: -10, opacity: 0 }}
                        animate={{ y: 0, opacity: 1 }}
                        transition={{ duration: 0.2 }}
                    >
                        {likeCount}
                    </motion.span>
                </motion.button>
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

// RightSidebar Component
const RightSidebar = () => {
  const [suggestedUsers, setSuggestedUsers] = useState<any[]>([]);
  const [groups, setGroups] = useState<any[]>([]);
  const [loggedInUserId, setLoggedInUserId] = useState<number | null>(null);
  const [loading, setLoading] = useState<boolean>(true);

  // Fetch the logged-in user's ID
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
      console.log('Fetched user ID:', data);

      if (data.userID) {
        setLoggedInUserId(data.userID);
      } else {
        throw new Error('User ID not found in response');
      }
    } catch (error) {
      console.error('Error fetching user ID:', error);
      setLoggedInUserId(null);
    }
  };

  // fetch the user ID for the given username
  const fetchFollowedID = async (username: string): Promise<number | null> => {
    try {
      const response = await fetch('http://localhost:8080/userID', {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username }),
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch user ID, Status: ${response.status}`);
      }

      const data = await response.json();
      console.log('Fetched user ID:', data);
      return data.userID || null;
    } catch (error) {
      console.error('Error fetching user ID:', error);
      return null;
    }
  };

  // Handle following a user
  const handleFollow = async (followedName: string) => {
    if (!loggedInUserId) {
      console.error('Logged-in user ID is not available');
      return;
    }

    const followedID = await fetchFollowedID(followedName);

    if (followedID === null) {
      console.error('Failed to get followed user ID');
      return;
    }

    console.log(`User ID: ${loggedInUserId}, Followed ID: ${followedID}`);

    try {
      const response = await fetch('http://localhost:8080/follow', {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          follower_id: loggedInUserId,
          followed_id: followedID,
          status: 'pending',
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to follow user');
      }

      // Optionally refresh the suggested users list
      fetchSuggestedUsers();
    } catch (error) {
      console.error('Error following user:', error);
    }
  };

  // Fetch suggested users
  const fetchSuggestedUsers = async () => {
    try {
      const response = await fetch('http://localhost:8080/users/suggested', {
        credentials: 'include',
      });
      if (response.ok) {
        const data = await response.json();
        setSuggestedUsers(data);
      }
    } catch (error) {
      console.error('Error fetching suggested users:', error);
    }
  };

  // Fetch groups
  const fetchGroups = async () => {
    try {
      const response = await fetch('http://localhost:8080/groups', {
        credentials: 'include',
      });
      if (response.ok) {
        const data = await response.json();
        setGroups(data);
      }
    } catch (error) {
      console.error('Error fetching groups:', error);
    }
  };

  useEffect(() => {
    const fetchData = async () => {
      await loginUserID(); // Fetch the logged-in user's ID
      await fetchSuggestedUsers(); // Fetch suggested users
      await fetchGroups(); // Fetch groups
      setLoading(false); // Set loading to false after all data is fetched
    };
    fetchData();
  }, []);

  if (loading) {
    return <div>Loading...</div>; // Optional: Show loading state while data is being fetched
  }

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
                  onClick={() => handleFollow(user.username)}
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
  );
};


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
      <div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
        <Header />
        <div className="flex pt-16">
          <Sidebar />
          <div className="flex-1 overflow-hidden">
            <main className="overflow-y-auto h-[calc(100vh-64px)]">
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

// Export the Feed component as default
export default Feed
