import React, { useState, useRef } from 'react';
import { Upload, X } from 'lucide-react';
import { POST_LIMITS } from '@/constants/limits';


interface CreatePostProps {
    onPostCreated: () => void;
}

export function CreatePost({ onPostCreated }: CreatePostProps) {
    const [content, setContent] = useState('');
    const [title, setTitle] = useState('');
    const [privacy, setPrivacy] = useState('1');
    const [media, setMedia] = useState<string | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [isUploading, setIsUploading] = useState(false);
    const fileInputRef = useRef<HTMLInputElement>(null);

    const handleMediaChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        // Validate file type
        const validTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
        if (!validTypes.includes(file.type)) {
            setError('Please upload an image file (JPEG, PNG, GIF, or WEBP)');
            return;
        }

        // Validate file size (e.g., 5MB limit)
        const MAX_FILE_SIZE = 5 * 1024 * 1024; // 5MB in bytes
        if (file.size > MAX_FILE_SIZE) {
            setError('File size should be less than 5MB');
            return;
        }

        try {
            setIsUploading(true);
            setError(null);

            // Convert file to base64
            const base64 = await fileToBase64(file);
            setMedia(base64);
        } catch (error) {
            setError('Error processing image');
            console.error('Error processing image:', error);
        } finally {
            setIsUploading(false);
        }
    };

    const fileToBase64 = (file: File): Promise<string> => {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.readAsDataURL(file);
            reader.onload = () => resolve(reader.result as string);
            reader.onerror = (error) => reject(error);
        });
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
        
        if (!title.trim() && !content.trim() && !media) {
            setError('Post must have either title, content, or media');
            return;
        }

        try {
            const privacyInt = parseInt(privacy, 10);
            if (isNaN(privacyInt)) {
                setError('Invalid privacy value');
                return;
            }

            const payload = {
                title: title.trim(),
                content: content.trim(),
                privacy: privacyInt,
                media,
            };

            const response = await fetch('http://localhost:8080/posts', {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(payload),
            });

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(`Failed to create post: ${errorText}`);
            }

            // Reset form
            setContent('');
            setTitle('');
            setPrivacy('1');
            setMedia(null);
            setError(null);
            if (fileInputRef.current) {
                fileInputRef.current.value = '';
            }
            onPostCreated();
        } catch (error) {
            console.error('Error creating post:', error);
            setError(error instanceof Error ? error.message : 'Failed to create post');
        }
    };

    return (
        <div className="bg-white/10 backdrop-blur-lg rounded-lg shadow p-6 border border-gray-800/50 mb-4">
            {/* Modified Header */}
            <div className="flex items-center mb-4">
                <div>
                    <h3 className="font-semibold text-gray-200">Select Post Type</h3>
                    <select
                        value={privacy}
                        onChange={(e) => setPrivacy(e.target.value)}
                        className="text-sm bg-gray-800/50 border border-gray-700/50 rounded px-2 py-1 text-gray-400"
                    >
                        <option value="1">Public</option>
                        <option value="2">Followers</option>
                        <option value="3">Close Friends</option>
                    </select>
                </div>
            </div>

            <form onSubmit={handleSubmit}>
                {error && (
                    <div className="mb-4 p-3 bg-red-500/10 border border-red-500/50 rounded-lg text-red-500">
                        {error}
                    </div>
                )}

                <input
                    type="text"
                    value={title}
                    onChange={(e) => {
                        if (e.target.value.length <= POST_LIMITS.TITLE_MAX_LENGTH) {
                            setTitle(e.target.value);
                        }
                    }}
                    placeholder="Post title..."
                    className="w-full bg-gray-800/50 border border-gray-700/50 rounded-lg p-2 mb-2 text-gray-200"
                    maxLength={POST_LIMITS.TITLE_MAX_LENGTH}
                />
                <div className="text-sm text-gray-400 mb-2">
                    {title.length}/{POST_LIMITS.TITLE_MAX_LENGTH} characters
                </div>

                <textarea
                    value={content}
                    onChange={(e) => {
                        if (e.target.value.length <= POST_LIMITS.CONTENT_MAX_LENGTH) {
                            setContent(e.target.value);
                        }
                    }}
                    placeholder="What's on your mind?"
                    className="w-full bg-gray-800/50 border border-gray-700/50 rounded-lg p-2 mb-2 text-gray-200 min-h-[100px]"
                    maxLength={POST_LIMITS.CONTENT_MAX_LENGTH}
                />
                <div className="text-sm text-gray-400 mb-4">
                    {content.length}/{POST_LIMITS.CONTENT_MAX_LENGTH} characters
                </div>

                <div className="mb-4">
                    <input
                        type="file"
                        ref={fileInputRef}
                        onChange={handleMediaChange}
                        accept="image/*,.gif"
                        className="hidden"
                        id="media-upload"
                    />
                    <div className="flex items-center space-x-4">
                        <label
                            htmlFor="media-upload"
                            className="cursor-pointer flex items-center space-x-2 bg-gray-800/50 hover:bg-gray-700/50 text-gray-200 px-4 py-2 rounded-lg transition-colors"
                        >
                            <Upload size={20} />
                            <span>Add Media</span>
                        </label>
                        
                        {isUploading && (
                            <span className="text-gray-400">Uploading...</span>
                        )}

                        {media && (
                            <div className="relative">
                                <img 
                                    src={media} 
                                    alt="Upload preview" 
                                    className="h-20 w-20 object-cover rounded-lg"
                                />
                                <button
                                    type="button"
                                    onClick={() => {
                                        setMedia(null);
                                        if (fileInputRef.current) {
                                            fileInputRef.current.value = '';
                                        }
                                    }}
                                    className="absolute -top-2 -right-2 bg-red-500 hover:bg-red-600 text-white rounded-full w-6 h-6 flex items-center justify-center transition-colors"
                                >
                                    <X size={16} />
                                </button>
                            </div>
                        )}
                    </div>
                </div>

                <button
                    type="submit"
                    disabled={isUploading}
                    className={`bg-blue-500 hover:bg-blue-600 text-white px-6 py-2 rounded-lg transition-colors ${
                        isUploading ? 'opacity-50 cursor-not-allowed' : ''
                    }`}
                >
                    {isUploading ? 'Uploading...' : 'Post'}
                </button>
            </form>
        </div>
    );
}

export default CreatePost;