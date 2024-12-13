'use client'

import { useState, useRef } from 'react'
import { Image as ImageIcon, X } from 'lucide-react'

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
    if (file) {
      if (file.size > 5 * 1024 * 1024) { // 5MB limit
        alert('Image size should be less than 5MB')
        return
      }

      const reader = new FileReader()
      reader.onloadend = () => {
        setSelectedImage(reader.result as string)
      }
      reader.readAsDataURL(file)
    }
  }

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

  return (
    <div className="bg-gray-800 rounded-lg p-4 mb-6">
      <form onSubmit={handleSubmit}>
        <input
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Post title"
          className="w-full px-4 py-2 bg-gray-700 rounded-lg text-gray-200 mb-2"
          required
        />
        <textarea
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="What's on your mind?"
          className="w-full px-4 py-2 bg-gray-700 rounded-lg text-gray-200 mb-2"
          rows={3}
          required
        />

        {selectedImage && (
          <div className="relative mb-2">
            <img
              src={selectedImage}
              alt="Selected"
              className="max-h-48 rounded-lg"
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
      </form>
    </div>
  )
}
