import React, { useState, useEffect } from 'react';
import Uploader from '@/components/ui/uploadButton'


// Create Post Component
interface CreatePostProps {
    onPostCreated: () => void;
  }
  
  export function CreatePost({ onPostCreated }: CreatePostProps) {
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

  export default CreatePost;