'use client'

import { useState, useRef } from 'react'
import { Image as ImageIcon, X } from 'lucide-react'
import { GROUP_LIMITS } from '@/constants/limits'
import toast, { Toaster } from 'react-hot-toast'

interface Post {
  title: string
  id: number
  content: string
  author: number
  author_name : string
  created_at: string
  media?: string
  media_type?: string
  comments: {
    id: number
    content: string
    author: number
    created_at: string
  }[]
}


interface CreateGroupPostProps {
  groupID: number
  onPostCreated: (groupId: number) => Promise<Post[]>
}

export default function CreateGroupPost({ groupID, onPostCreated }: CreateGroupPostProps) {
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [selectedImage, setSelectedImage] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleImageSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
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
          setSelectedImage(reader.result as string);
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
        setSelectedImage(compressedDataUrl);

      } catch (error) {
        console.error('Error processing image:', error);
        toast.error('Failed to process image. Please try again.');
      }
    };

    processImage(file);
  };

  const removeImage = () => {
    setSelectedImage(null)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!title.trim() || !content.trim()) return

    setIsSubmitting(true)
    try {
      const response = await fetch(`http://localhost:8080/groups/${groupID}/posts`, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          title,
          content,
          media: selectedImage,
          privacy: 1,
          group_id: groupID,
        }),
      })

      if (!response.ok) {
        throw new Error('Failed to create post')
      }

      setTitle('')
      setContent('')
      setSelectedImage(null)
      if (fileInputRef.current) {
        fileInputRef.current.value = ''
      }
      
      await onPostCreated(groupID)
    } catch (error) {
      console.error('Error creating post:', error)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleTitleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newTitle = e.target.value;
    if (newTitle.length <= GROUP_LIMITS.POST_TITLE_MAX_LENGTH) {
      setTitle(newTitle);
    }
  };

  const handleContentChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newContent = e.target.value;
    if (newContent.length <= GROUP_LIMITS.POST_CONTENT_MAX_LENGTH) {
      setContent(newContent);
    }
  };

  return (
    <div className="bg-white/10 backdrop-blur-lg rounded-lg shadow p-6 border border-gray-800/50 mb-4">
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
      <form onSubmit={handleSubmit}>
        <div className="space-y-4">
          <div>
            <input
              type="text"
              value={title}
              onChange={handleTitleChange}
              placeholder="Post title..."
              maxLength={GROUP_LIMITS.POST_TITLE_MAX_LENGTH}
              className="w-full bg-gray-800/50 border border-gray-700/50 rounded-lg p-2 text-gray-200"
            />
            <div className="text-sm text-gray-400 mt-1">
              {title.length}/{GROUP_LIMITS.POST_TITLE_MAX_LENGTH} characters
            </div>
          </div>

          <div>
            <textarea
              value={content}
              onChange={handleContentChange}
              placeholder="What's on your mind?"
              maxLength={GROUP_LIMITS.POST_CONTENT_MAX_LENGTH}
              className="w-full bg-gray-800/50 border border-gray-700/50 rounded-lg p-2 text-gray-200 min-h-[100px]"
            />
            <div className="text-sm text-gray-400 mt-1">
              {content.length}/{GROUP_LIMITS.POST_CONTENT_MAX_LENGTH} characters
            </div>
          </div>

          {selectedImage && (
            <div className="relative mb-2">
              <img
                src={selectedImage}
                alt="Selected"
                className="max-h-96 w-auto max-w-full h-auto object-contain rounded-lg"
              />
              <button
                type="button"
                onClick={removeImage}
                className="absolute top-2 right-2 p-1 bg-gray-900/50 rounded-full hover:bg-gray-900"
              >
                <X className="w-4 h-4 text-white" />
              </button>
            </div>
          )}

          <div className="flex justify-between items-center">
            <div>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                onChange={handleImageSelect}
                className="hidden"
                id="image-upload"
              />
              <label
                htmlFor="image-upload"
                className="flex items-center space-x-2 px-4 py-2 bg-gray-700 text-gray-200 rounded-lg cursor-pointer hover:bg-gray-600"
              >
                <ImageIcon className="w-5 h-5" />
                <span>Add Image</span>
              </label>
            </div>

            <button
              type="submit"
              disabled={isSubmitting || !title.trim() || !content.trim()}
              className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isSubmitting ? 'Posting...' : 'Post'}
            </button>
          </div>
        </div>
      </form>
    </div>
  )
}
