import { useState, useEffect, FormEvent, useRef, useCallback } from 'react'
import { ThumbsUp, MessageCircle, User, Image as ImageIcon, X } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { GROUP_LIMITS } from '@/constants/limits'
import toast, { Toaster } from 'react-hot-toast'

interface GroupPostProps {
    post: GroupPostType;
    groupId: number;
}

interface GroupPostType {
    id: number;
    title: string;
    content: string;
    media?: string;
    media_type?: string;
    author: number;
    author_name: string;
    author_avatar?: string;
    created_at: string;
    group_id: number;
    like_count?: number;
    user_liked?: boolean;
}

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

export default function GroupPost({ post, groupId }: GroupPostProps) {
    const [showComments, setShowComments] = useState(false);
    const [comments, setComments] = useState<Comment[]>([]);
    const [newComment, setNewComment] = useState('');
    const [commentMedia, setCommentMedia] = useState<string | null>(null);
    const [isLoadingComments, setIsLoadingComments] = useState(false);
    const [commentCount, setCommentCount] = useState(0);
    const fileInputRef = useRef<HTMLInputElement>(null);

    const fetchComments = async () => {
        setIsLoadingComments(true);
        try {
////console.log here for debugging here for debugging
            const response = await fetch(`http://localhost:8080/groups/${groupId}/posts/${post.id}/comments`, {
                credentials: 'include',
            });
            if (response.ok) {
                const data = await response.json();
////console.log here for debugging here for debugging
                setComments(data || []);
                setCommentCount(data?.length || 0);
            } else {
                const errorText = await response.text();
                console.error('Error response:', errorText);
            }
        } catch (error) {
            console.error('Error fetching comments:', error);
            setComments([]);
            setCommentCount(0);
        } finally {
            setIsLoadingComments(false);
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

    const handleCommentSubmit = async (e: FormEvent) => {
        e.preventDefault();
        if (!newComment.trim() && !commentMedia) return;

        try {
            const payload = {
                content: newComment.trim(),
                post_id: post.id,
                group_id: groupId,
                media: commentMedia || ""
            };

////console.log here for debugging here for debugging

            const response = await fetch(`http://localhost:8080/groups/${groupId}/posts/${post.id}/comments`, {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(payload),
            });

            if (!response.ok) {
                throw new Error(await response.text());
            }

            const newCommentData = await response.json();
////console.log here for debugging here for debugging
            setComments(prev => [newCommentData, ...prev]);
            setCommentCount(prev => prev + 1);
            setNewComment('');
            setCommentMedia(null);
            if (fileInputRef.current) {
                fileInputRef.current.value = '';
            }
        } catch (error) {
            console.error('Error posting comment:', error);
        }
    };

    const handleCommentChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        const newComment = e.target.value;
        if (newComment.length <= GROUP_LIMITS.COMMENT_MAX_LENGTH) {
            setNewComment(newComment);
        }
    };

    useEffect(() => {
        fetchComments();
    }, [post.id, groupId]);

    return (
        <div className="bg-gray-800 rounded-lg p-6 mb-4">
            <Toaster 
                position="top-right"
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
            {/* Post Header */}
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

            {/* Post Content - Updated with overflow handling */}
            <h2 className="text-xl font-semibold text-gray-200 mb-2 break-words whitespace-pre-wrap overflow-hidden">
                {post.title}
            </h2>
            <div className="text-gray-200 mb-4 break-words whitespace-pre-wrap overflow-hidden">
                {post.content}
            </div>

            {/* Media section - Updated for left alignment */}
            {post.media && (
                <div className="mb-4">
                    <img 
                        src={post.media} 
                        alt="Post media" 
                        className="rounded-lg max-h-96 w-auto max-w-full h-auto object-contain"
                    />
                </div>
            )}

            {/* Actions */}
            <div className="flex items-center space-x-4 text-gray-400 mb-4">
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

            {/* Comments Section - Updated with overflow handling */}
            <AnimatePresence>
                {showComments && (
                    <motion.div
                        initial={{ height: 0, opacity: 0 }}
                        animate={{ height: 'auto', opacity: 1 }}
                        exit={{ height: 0, opacity: 0 }}
                        className="space-y-4"
                    >
                        {/* Comment Form */}
                        <form onSubmit={handleCommentSubmit} className="space-y-3">
                            <div className="relative">
                                <textarea
                                    value={newComment}
                                    onChange={handleCommentChange}
                                    placeholder="Write a comment..."
                                    maxLength={GROUP_LIMITS.COMMENT_MAX_LENGTH}
                                    className="w-full px-3 py-2 bg-gray-700 rounded-lg text-gray-200 pr-16"
                                    rows={2}
                                />
                                <div className="absolute right-2 top-2 text-xs text-gray-400">
                                    {newComment.length}/{GROUP_LIMITS.COMMENT_MAX_LENGTH}
                                </div>
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
                                            onClick={() => {
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
                                <button
                                    type="submit"
                                    className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
                                    disabled={!newComment.trim() && !commentMedia}
                                >
                                    Comment
                                </button>
                            </div>
                        </form>

                        {/* Comments List - Updated with overflow handling */}
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
                                        <div className="flex-1 bg-gray-700/30 rounded-lg p-3">
                                            <p className="text-sm font-medium text-gray-200 break-words">
                                                {comment.author_name}
                                            </p>
                                            <div className="text-sm text-gray-300 break-words whitespace-pre-wrap overflow-hidden">
                                                {comment.content}
                                            </div>
                                            {comment.media && (
                                                <div className="mt-2">
                                                    <img
                                                        src={comment.media}
                                                        alt="Comment media"
                                                        className="max-h-60 w-auto max-w-full h-auto object-contain rounded-lg"
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
    );
} 