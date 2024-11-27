import React, { useState } from 'react';
import Uploader from '@/components/ui/uploadButton';

interface Post {
  id: number;
  content: string;
  author: string;
  created_at: string;
  comments: {
    id: number;
    content: string;
    author: string;
    created_at: string;
  }[];
}

// Create Post Component
interface CreateGroupPostProps {
  onPostCreated: (groupID: number) => Promise<Post[]>;
  groupID: number;
}

export function CreateGroupPost({ onPostCreated, groupID }: CreateGroupPostProps) {
  const [content, setContent] = useState('');
  const [title, setTitle] = useState('');
  const [privacy, setPrivacy] = useState('1'); // default privacy is '1' (Public)
  const [media, setMedia] = useState<string | null>(null); // New state for media

  const handleSubmit = async (e: React.FormEvent) => {
    console.log("createGrouppost work")
    e.preventDefault();
    console.log('Media before submit:', media); // Confirm media has base64
    
    // Validate that media is a non-null and non-NaN base64 string
    if (!media || media === 'NaN') {
      console.error('Invalid media: media is null, NaN, or undefined');
      return;
    }

    try {
      const privacyInt = parseInt(privacy, 10);
      if (isNaN(privacyInt)) {
        console.error('Invalid privacy value:', privacy);
        return;
      }

      const response = await fetch(`http://localhost:8080/groups/${groupID}/posts`, {
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
          group_id: groupID,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to create post');
      }

      // Reset form fields
      setContent('');
      setTitle('');
      setPrivacy('1');
      setMedia(null);
      // Trigger the callback function
      onPostCreated(groupID);
    } catch (error) {
      console.error('Error creating post:', error);
    }
  };

  return (
    <div className="bg-white/10 backdrop-blur-lg rounded-lg shadow p-6 border border-gray-800/500 w-[900px] -ml-30 mb-4">
      <form onSubmit={handleSubmit}>
        <div>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Post title..."
            className="w-full bg-gray-800/50 border border-gray-700/50 rounded-lg p-2 mb-2 text-gray-200"
          />
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
            onUpload={(base64: string) => {
              console.log('Received base64 from Uploader:', base64); // Log base64 before setting
              // Check if base64 is valid before setting it
              if (typeof base64 === 'string' && base64.startsWith('data:image/')) {
                setMedia(base64);
              } else {
                console.error('Invalid base64 string:', base64);
              }
            }}
          />
        </div>
      </form>
    </div>
  );
}

export default CreateGroupPost;
