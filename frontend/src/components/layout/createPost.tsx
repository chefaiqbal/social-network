import React, { useState } from 'react';
import Uploader from '@/components/ui/uploadButton';

interface CreatePostProps {
    onPostCreated: () => void;
}

export function CreatePost({ onPostCreated }: CreatePostProps) {
    const [content, setContent] = useState('');
    const [title, setTitle] = useState('');
    const [privacy, setPrivacy] = useState('1');
    const [media, setMedia] = useState<string | null>(null);
    const [error, setError] = useState<string | null>(null);

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

            console.log('Submitting post with payload:', payload);

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
                console.error('Server error response:', {
                    status: response.status,
                    statusText: response.statusText,
                    error: errorText
                });
                throw new Error(`Failed to create post: ${errorText}`);
            }

            const result = await response.json();
            console.log('Post created successfully:', result);

            setContent('');
            setTitle('');
            setPrivacy('1');
            setMedia(null);
            setError(null);
            onPostCreated();
        } catch (error) {
            console.error('Error creating post:', error);
            setError(error instanceof Error ? error.message : 'Failed to create post');
        }
    };

    return (
        <div className="bg-white/10 backdrop-blur-lg rounded-lg shadow p-6 border border-gray-800/500 w-[1155px] -ml-40 mb-4">
            <form onSubmit={handleSubmit}>
                {error && (
                    <div className="mb-4 p-3 bg-red-500/10 border border-red-500/50 rounded-lg text-red-500">
                        {error}
                    </div>
                )}
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
                    <Uploader 
                        onUpload={(base64: string, mediaType: string) => {
                            console.log('Received media:', { type: mediaType, preview: base64.substring(0, 100) });
                            setMedia(base64);
                        }} 
                        acceptedTypes="image/*,.gif"
                        buttonText="Add Media"
                    />
                    {media && (
                        <div className="relative">
                            <img 
                                src={media} 
                                alt="Upload preview" 
                                className="h-20 w-20 object-cover rounded-lg"
                            />
                            <button
                                type="button"
                                onClick={() => setMedia(null)}
                                className="absolute -top-2 -right-2 bg-red-500 text-white rounded-full w-5 h-5 flex items-center justify-center hover:bg-red-600"
                            >
                                ×
                            </button>
                        </div>
                    )}
                </div>
            </form>
        </div>
    );
}

export default CreatePost;