'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/router'

const Post = () => {
  const router = useRouter()
  const { id } = router.query
  const [post, setPost] = useState<any | null>(null)

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

  if (!post) {
    return <div>Loading...</div>
  }

  return (
    <div className="flex h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
      <div className="post-content">
        <h1>{post.title}</h1>
        <p>{post.content}</p>
      </div>
    </div>
  )
}

export default Post
