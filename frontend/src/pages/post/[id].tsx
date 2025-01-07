'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/router'
import { POST_LIMITS } from '@/constants/limits'

interface Post {
  id: number
  title: string
  content: string
  author: number
  authorName: string
  createdAt: string
  // ... other fields
}

const Post = () => {
  const router = useRouter()
  const { id } = router.query
  const [post, setPost] = useState<Post | null>(null)
  const [editTitle, setEditTitle] = useState('')
  const [editContent, setEditContent] = useState('')

  useEffect(() => {
    if (!id) return

    const fetchPost = async () => {
      try {
        const response = await fetch(`http://localhost:8080/posts/${id}`, {
          credentials: 'include',
        })
        if (!response.ok) {
          throw new Error('Failed to fetch post')
        }
        const data = await response.json()
        setPost(data)
      } catch (error) {
        console.error('Error fetching post:', error)
      }
    }

    fetchPost()
  }, [id])

  const handleTitleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newTitle = e.target.value
    if (newTitle.length <= POST_LIMITS.TITLE_MAX_LENGTH) {
      setEditTitle(newTitle)
    }
  }

  const handleContentChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newContent = e.target.value
    if (newContent.length <= POST_LIMITS.CONTENT_MAX_LENGTH) {
      setEditContent(newContent)
    }
  }

  if (!post) {
    return <div>Loading...</div>
  }

  return (
    <div className="flex h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
      <div className="post-content">
        {/* If editing */}
        <input
          type="text"
          value={editTitle}
          onChange={handleTitleChange}
          maxLength={POST_LIMITS.TITLE_MAX_LENGTH}
          className="..."
        />
        <textarea
          value={editContent}
          onChange={handleContentChange}
          maxLength={POST_LIMITS.CONTENT_MAX_LENGTH}
          className="..."
        />
        {/* If not editing */}
        <h1>{post.title}</h1>
        <p>{post.content}</p>
      </div>
    </div>
  )
}

export default Post
