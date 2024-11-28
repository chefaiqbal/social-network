'use client'

import { useState, useEffect, FormEvent, useRef, useCallback } from 'react'
import Link from 'next/link'
import { ThumbsUp, MessageCircle, Share2, User } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import DropDownCheck from '@/components/ui/DropDownCheck'
import { ChatList } from '@/components/chat/ChatList'
import { Toaster } from 'react-hot-toast'
import Header from '@/components/layout/Header'
import Sidebar from '@/components/layout/Sidebar'
import CreatePost from '@/components/layout/createPost'
import Uploader from '@/components/ui/uploadButton'
import CommentUploader from '@/components/ui/commentUploader'

// Post Component
function Post({ post }: { post: PostType }) {
    const [likeCount, setLikeCount] = useState(post.like_count || 0);
    const [userLiked, setUserLiked] = useState(post.user_liked || false);
    const [isLoading, setIsLoading] = useState(false);
    const wsRef = useRef<WebSocket | null>(null);
    const reconnectTimeoutRef = useRef<NodeJS.Timeout>();
    const [showComments, setShowComments] = useState(false);
    const [comments, setComments] = useState<Comment[]>([]);
    const [commentCount, setCommentCount] = useState(0);
    const [newComment, setNewComment] = useState('');
    const [isLoadingComments, setIsLoadingComments] = useState(false);
    const [commentMedia, setCommentMedia] = useState<string | null>(null);

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

    // Update fetchComments to set comment count
    const fetchComments = async () => {
        setIsLoadingComments(true);
        try {
            const response = await fetch(`http://localhost:8080/comments/${post.id}`, {
                credentials: 'include',
            });
            if (response.ok) {
                const data = await response.json();
                setComments(data || []); // Ensure it's an array
                setCommentCount(data?.length || 0); // Update comment count
            }
        } catch (error) {
            console.error('Error fetching comments:', error);
            setComments([]); // Set empty array on error
            setCommentCount(0);
        } finally {
            setIsLoadingComments(false);
        }
    };

    // Add useEffect to fetch initial comment count
    useEffect(() => {
        const fetchCommentCount = async () => {
            try {
                const response = await fetch(`http://localhost:8080/comments/${post.id}/count`, {
                    credentials: 'include',
                });
                if (response.ok) {
                    const data = await response.json();
                    setCommentCount(data.count || 0);
                }
            } catch (error) {
                console.error('Error fetching comment count:', error);
                setCommentCount(0);
            }
        };

        fetchCommentCount();
    }, [post.id]);

    // Add function to handle comment submission
    const handleCommentSubmit = async (e: FormEvent) => {
        e.preventDefault();
        if (!newComment.trim() && !commentMedia) return;

        try {
            const payload = {
                content: newComment.trim(),
                post_id: Number(post.id),
                media: commentMedia || ""
            };

            console.log("Submitting comment with payload:", payload);

            const response = await fetch('http://localhost:8080/comments', {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(payload),
            });

            if (!response.ok) {
                const errorText = await response.text();
                console.error('Server error response:', {
                    status: response.status,
                    statusText: response.statusText,
                    error: errorText
                });
                throw new Error(`Failed to post comment: ${errorText}`);
            }

            const result = await response.json();
            console.log("Comment created successfully:", result);

            setNewComment('');
            setCommentMedia(null);
            fetchComments(); // Refresh comments after posting
            
            // Update comment count
            setCommentCount(prev => prev + 1);
        } catch (error) {
            console.error('Error posting comment:', error);
            // You might want to show an error message to the user here
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
                <div className="mb-4">
                    <img 
                        src={post.media} 
                        alt="Post media" 
                        className="rounded-lg max-h-96 object-cover"
                        style={{
                            objectFit: post.media_type === 'image/gif' ? 'contain' : 'cover'
                        }}
                    />
                </div>
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
                <button 
                    onClick={() => {
                        setShowComments(!showComments);
                        if (!showComments) {
                            fetchComments();
                        }
                    }}
                    className="flex items-center space-x-1 hover:text-blue-500 transition-colors"
                >
                    <MessageCircle size={20} />
                    <span>{commentCount}</span>
                </button>
                <button className="flex items-center space-x-1 hover:text-green-500 transition-colors">
                    <Share2 size={20} />
                    <span>Share</span>
                </button>
            </div>

            {/* Comments Section */}
            <AnimatePresence>
                {showComments && (
                    <motion.div
                        initial={{ height: 0, opacity: 0 }}
                        animate={{ height: 'auto', opacity: 1 }}
                        exit={{ height: 0, opacity: 0 }}
                        className="mt-4 space-y-4"
                    >
                        {/* Comment Form */}
                        <form onSubmit={handleCommentSubmit} className="flex flex-col space-y-2" onClick={(e) => e.stopPropagation()}>
                            <div className="flex space-x-2">
                                <input
                                    type="text"
                                    value={newComment}
                                    onChange={(e) => setNewComment(e.target.value)}
                                    placeholder="Write a comment..."
                                    className="flex-1 bg-gray-800/50 border border-gray-700/50 rounded-lg p-2 text-gray-200"
                                />
                                <button
                                    type="submit"
                                    className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
                                >
                                    Comment
                                </button>
                            </div>
                            <div className="flex items-center space-x-2">
                                <CommentUploader 
                                    onUpload={(base64: string) => {
                                        console.log("Comment media uploaded:", base64.substring(0, 50));
                                        setCommentMedia(base64);
                                    }}
                                    buttonText="Add Image"
                                />
                                {commentMedia && (
                                    <div className="relative">
                                        <img 
                                            src={commentMedia} 
                                            alt="Comment upload preview" 
                                            className="h-20 w-20 object-cover rounded-lg"
                                        />
                                        <button
                                            type="button"
                                            onClick={(e) => {
                                                e.preventDefault();
                                                e.stopPropagation();
                                                setCommentMedia(null);
                                            }}
                                            className="absolute -top-2 -right-2 bg-red-500 text-white rounded-full w-5 h-5 flex items-center justify-center hover:bg-red-600"
                                        >
                                            ×
                                        </button>
                                    </div>
                                )}
                            </div>
                        </form>

                        {/* Comments List */}
                        <div className="space-y-4">
                            {isLoadingComments ? (
                                <div className="text-center text-gray-400">Loading comments...</div>
                            ) : comments.length > 0 ? (
                                comments.map((comment) => (
                                    <motion.div
                                        key={comment.id}
                                        initial={{ opacity: 0, y: 20 }}
                                        animate={{ opacity: 1, y: 0 }}
                                        className="flex space-x-3"
                                    >
                                        <div className="flex-shrink-0 w-8 h-8">
                                            {comment.author_avatar ? (
                                                <img
                                                    src={comment.author_avatar}
                                                    alt=""
                                                    className="rounded-full"
                                                />
                                            ) : (
                                                <div className="w-8 h-8 bg-gray-700 rounded-full flex items-center justify-center">
                                                    <User size={16} className="text-gray-400" />
                                                </div>
                                            )}
                                        </div>
                                        <div className="flex-1 bg-gray-800/30 rounded-lg p-3">
                                            <p className="text-sm font-medium text-gray-200">
                                                {comment.author_name || `User ${comment.author}`}
                                            </p>
                                            <p className="text-sm text-gray-300">{comment.content}</p>
                                            {comment.media && (
                                                <div className="mt-2">
                                                    <img
                                                        src={comment.media}
                                                        alt="Comment media"
                                                        className="max-h-60 rounded-lg object-cover"
                                                    />
                                                </div>
                                            )}
                                            <p className="text-xs text-gray-400 mt-1">
                                                {new Date(comment.created_at).toLocaleString()}
                                            </p>
                                        </div>
                                    </motion.div>
                                ))
                            ) : (
                                <div className="text-center text-gray-400">No comments yet</div>
                            )}
                        </div>
                    </motion.div>
                )}
            </AnimatePresence>
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
