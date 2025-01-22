import Uploader from '@/components/ui/uploadButton'
import CommentUploader from '@/components/ui/commentUploader'
import { useState, useEffect, FormEvent, useRef, useCallback } from 'react'
import Link from 'next/link'
import { ThumbsUp, MessageCircle, User } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { ChatList } from '@/components/chat/ChatList'
import { Toaster } from 'react-hot-toast'
import { POST_LIMITS, COMMENT_LIMITS } from '@/constants/limits'
import { Image as ImageIcon } from 'lucide-react'
import toast from 'react-hot-toast'


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
    const fileInputRef = useRef<HTMLInputElement>(null);

    /*const connectWebSocket = useCallback(() => {
        if (wsRef.current?.readyState === WebSocket.OPEN) {
            return;
        }

        wsRef.current = new WebSocket('ws://localhost:8080/ws/likes');
        
        wsRef.current.onopen = () => {

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

            reconnectTimeoutRef.current = setTimeout(connectWebSocket, 3000);
        };

        wsRef.current.onerror = (error) => {
            //console.error('Like WebSocket error:', error);
        };
    }, [post.id]);*/

    useEffect(() => {
        // Get current user ID from localStorage when component mounts
        const userId = localStorage.getItem('userId');
        if (!userId) {
            fetch('http://localhost:8080/userIDBY', {
                credentials: 'include',
            })
            .then(res => res.json())
            .then(data => {
                // Fix: Changed to data.id to match the backend response
                if (data && data.id) {
                    localStorage.setItem('userId', data.id.toString());
                }
            })
            .catch(() => {
                // Silently handle error to prevent console spam
            });
        }

        // connectWebSocket();

        return () => {
            if (reconnectTimeoutRef.current) {
                clearTimeout(reconnectTimeoutRef.current);
            }
            if (wsRef.current) {
                wsRef.current.close();
            }
        };
    }, [/*connectWebSocket]*/]);

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
                if (response.status === 401) {
                    // Handle unauthorized - maybe redirect to login
                    window.location.href = '/';
                    return;
                }
                throw new Error('Failed to update like');
            }

            const data = await response.json();
            setLikeCount(data.like_count);
            setUserLiked(!userLiked);
        } catch (error) {
            console.error('Error liking post:', error);
            toast.error('Failed to update like');
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

////console.log here for debugging here for debugging

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
////console.log here for debugging here for debugging

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

    const handleCommentChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const newComment = e.target.value;
        if (newComment.length <= COMMENT_LIMITS.CONTENT_MAX_LENGTH) {
            setNewComment(newComment);
        }
    };

    const handleImageSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        // Validate file type
        const validTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
        if (!validTypes.includes(file.type)) {
            toast.error('Please upload a valid image file (JPEG, PNG, GIF, or WebP)');
            if (fileInputRef.current) {
                fileInputRef.current.value = '';
            }
            return;
        }

        // Validate file size (200kB limit)
        const maxSize = 200 * 1024; // 200kB in bytes
        if (file.size > maxSize) {
            toast.error('Image size must be 200kB or less');
            if (fileInputRef.current) {
                fileInputRef.current.value = '';
            }
            return;
        }

        // Process image
        const processImage = async (file: File) => {
            // Don't compress GIFs to preserve animation
            if (file.type === 'image/gif') {
                const reader = new FileReader();
                reader.onloadend = () => {
                    setCommentMedia(reader.result as string);
                };
                reader.readAsDataURL(file);
                return;
            }

            // Compress other image types
            try {
                const img = new Image();
                const imgPromise = new Promise((resolve, reject) => {
                    img.onload = () => resolve(img);
                    img.onerror = reject;
                });
                img.src = URL.createObjectURL(file);
                await imgPromise;

                const canvas = document.createElement('canvas');
                let width = img.width;
                let height = img.height;

                // Maximum dimensions
                const MAX_WIDTH = 1200;
                const MAX_HEIGHT = 1200;

                if (width > height) {
                    if (width > MAX_WIDTH) {
                        height = Math.round((height * MAX_WIDTH) / width);
                        width = MAX_WIDTH;
                    }
                } else {
                    if (height > MAX_HEIGHT) {
                        width = Math.round((width * MAX_HEIGHT) / height);
                        height = MAX_HEIGHT;
                    }
                }

                canvas.width = width;
                canvas.height = height;
                const ctx = canvas.getContext('2d');
                if (!ctx) throw new Error('Failed to get canvas context');

                ctx.drawImage(img, 0, 0, width, height);
                const compressedDataUrl = canvas.toDataURL('image/jpeg', 0.8);
                setCommentMedia(compressedDataUrl);

            } catch (error) {
                console.error('Error processing image:', error);
                toast.error('Failed to process image. Please try again.');
            }
        };

        processImage(file);
    };

    return (
        <>
            <Toaster 
                position="top-center"
                toastOptions={{
                    duration: 3000,
                    style: {
                        background: '#1f2937',
                        color: '#fff',
                    },
                    error: {
                        style: {
                            background: '#991b1b',
                        },
                    },
                }}
            />
            <div className="bg-white/10 backdrop-blur-lg rounded-lg shadow p-6 border border-gray-800/50 mb-4">
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
                <h2 className="text-xl font-semibold text-gray-200 mb-2 break-words">
                    {post.title}
                </h2>
                <div className="text-gray-300 whitespace-pre-wrap break-words overflow-hidden">
                    {post.content}
                </div>
                {post.media && (
                    <div className="mt-4">
                        <img 
                            src={post.media} 
                            alt="Post media" 
                            className="max-w-full h-auto rounded-lg"
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
                                    <div className="flex-1 relative">
                                        <input
                                            type="text"
                                            value={newComment}
                                            onChange={handleCommentChange}
                                            placeholder="Write a comment..."
                                            maxLength={COMMENT_LIMITS.CONTENT_MAX_LENGTH}
                                            className="w-full bg-gray-800/50 border border-gray-700/50 rounded-lg p-2 text-gray-200 pr-16"
                                        />
                                        <div className="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-gray-400">
                                            {newComment.length}/{COMMENT_LIMITS.CONTENT_MAX_LENGTH}
                                        </div>
                                    </div>
                                    <button
                                        type="submit"
                                        className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
                                        disabled={!newComment.trim() && !commentMedia}
                                    >
                                        Comment
                                    </button>
                                </div>
                                <div className="flex items-center space-x-3">
                                    <input
                                        ref={fileInputRef}
                                        type="file"
                                        accept="image/*"
                                        onChange={handleImageSelect}
                                        className="hidden"
                                        id={`comment-image-${post.id}`}
                                    />
                                    <label
                                        htmlFor={`comment-image-${post.id}`}
                                        className="cursor-pointer text-gray-400 hover:text-gray-200"
                                    >
                                        <ImageIcon size={20} />
                                    </label>
                                    {commentMedia && (
                                        <div className="relative">
                                            <img 
                                                src={commentMedia} 
                                                alt="Comment upload preview" 
                                                className="h-20 w-auto object-contain rounded-lg"
                                            />
                                            <button
                                                type="button"
                                                onClick={(e) => {
                                                    e.preventDefault();
                                                    e.stopPropagation();
                                                    setCommentMedia(null);
                                                    if (fileInputRef.current) {
                                                        fileInputRef.current.value = '';
                                                    }
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
                                            <div className="flex-1 bg-gray-800/30 rounded-lg p-3 max-w-[calc(100%-2rem)]">
                                                <p className="text-sm font-medium text-gray-200">
                                                    {comment.author_name || `User ${comment.author}`}
                                                </p>
                                                <div className="text-sm text-gray-300 break-all whitespace-normal">
                                                    {comment.content}
                                                </div>
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
        </>
    )
}

export default Post

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

