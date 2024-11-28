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

interface CreateGroupPostProps {
  onPostCreated: (groupID: number) => Promise<Post[]>;
  groupID: number;
}

export function CreateGroupPost({ onPostCreated, groupID }: CreateGroupPostProps) {
  const [content, setContent] = useState('');
  const [title, setTitle] = useState('');
  const [privacy, setPrivacy] = useState('1');
  const [media, setMedia] = useState<string | null>(null);
  const [notification, setNotification] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
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
          media,
          group_id: groupID,
        }),
      });

      if (!response.ok) {
        if (response.status === 403) {
          setNotification('You must join the group to post.');
          setTimeout(() => {
            setNotification(null); 
          }, 4000);
        } else {
          throw new Error('Failed to create post');
        }
        return;
      }

      setContent('');
      setTitle('');
      setPrivacy('1');
      setMedia(null);
      onPostCreated(groupID);
    } catch (error) {
      console.error('Error creating post:', error);
    }
  };

  return (
    <div className="bg-white/10 backdrop-blur-lg rounded-lg shadow p-6 border border-gray-800/500 w-[900px] -ml-30 mb-4">
      {notification && (
        <div className="bg-red-500 text-white p-3 rounded-lg mb-4 transition-opacity">
          {notification}
        </div>
      )}
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
