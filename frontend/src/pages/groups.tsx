'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Users, Plus, Search } from 'lucide-react'
import Link from 'next/link'

type Group = {
  id: number
  name: string
  description: string
  memberCount: number
  isPrivate: boolean
  createdAt: string
  avatar?: string
  isMember?: boolean
}

export default function Groups() {
  const [groups, setGroups] = useState<Group[]>([])
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [newGroup, setNewGroup] = useState({
    name: '',
    description: '',
    isPrivate: false
  })

  const handleJoinGroup = (groupId: number) => {
    setGroups(prevGroups => 
      prevGroups.map(group => 
        group.id === groupId 
          ? { ...group, isMember: true, memberCount: group.memberCount + 1 }
          : group
      )
    )
  }

  const handleCreateGroup = (e: React.FormEvent) => {
    e.preventDefault()
    const createdGroup = {
      id: Date.now(),
      name: newGroup.name,
      description: newGroup.description,
      isPrivate: newGroup.isPrivate,
      memberCount: 1,
      createdAt: 'Just now',
      isMember: true
    }

    setGroups(prevGroups => [createdGroup, ...prevGroups])
    setNewGroup({ name: '', description: '', isPrivate: false })
    setShowCreateModal(false)
  }

  useEffect(() => {
    const fetchGroups = async () => {
      try {
        setGroups([
          {
            id: 1,
            name: 'Barcelona FC Fans',
            description: 'Mes Que Un Club',
            memberCount: 150,
            isPrivate: false,
            createdAt: 'January 2024',
            isMember: true
          },
          {
            id: 2,
            name: 'VARDRID Club',
            description: 'Share your coding journey',
            memberCount: 75,
            isPrivate: true,
            createdAt: 'February 2024',
            isMember: false
          },
          {
            id: 3,
            name: 'Byte Area',
            description: 'pizza lovers',
            memberCount: 120,
            isPrivate: false,
            createdAt: 'March 2024',
            isMember: true
          }
        ])
      } catch (error) {
        console.error('Error fetching groups:', error)
      }
    }

    fetchGroups()
  }, [])

  const filteredGroups = groups.filter(group =>
    group.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    group.description.toLowerCase().includes(searchQuery.toLowerCase())
  )

  const myGroups = filteredGroups.filter(group => group.isMember)
  const availableGroups = filteredGroups.filter(group => !group.isMember)

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <Link href="/feed">
          <button className="mb-6 px-4 py-2 bg-gray-700 text-white rounded-lg hover:bg-gray-600 transition-colors">
            ← Back to Feed
          </button>
        </Link>

        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="flex justify-between items-center mb-8"
        >
          <h1 className="text-3xl font-bold text-gray-200">Groups</h1>
          <button
            onClick={() => setShowCreateModal(true)}
            className="flex items-center px-4 py-2 bg-blue-500 text-white rounded-full hover:bg-blue-600 transition-colors"
          >
            <Plus size={20} className="mr-2" />
            Create Group
          </button>
        </motion.div>

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="mb-8"
        >
          <div className="relative">
            <input
              type="text"
              placeholder="Search groups..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-2 bg-white/10 backdrop-blur-lg border border-gray-700/50 rounded-full focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-200"
            />
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" size={20} />
          </div>
        </motion.div>

        {/* My Groups Section */}
        <div className="mb-12">
          <h2 className="text-2xl font-semibold text-gray-200 mb-6">My Groups</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {myGroups.map((group) => (
              <motion.div
                key={group.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                className="bg-white/10 backdrop-blur-lg rounded-lg shadow-lg overflow-hidden"
              >
                <div className="p-6">
                  <div className="flex items-center space-x-4">
                    <div className="w-16 h-16 bg-gray-700 rounded-full flex items-center justify-center">
                      {group.avatar ? (
                        <img
                          src={group.avatar}
                          alt={group.name}
                          className="w-full h-full rounded-full object-cover"
                        />
                      ) : (
                        <Users size={24} className="text-gray-400" />
                      )}
                    </div>
                    <div>
                      <h3 className="text-xl font-semibold text-gray-200">{group.name}</h3>
                      <p className="text-gray-400">{group.memberCount} members</p>
                    </div>
                  </div>
                  <p className="mt-4 text-gray-300">{group.description}</p>
                  <div className="mt-6 flex justify-between items-center">
                    <span className="text-sm text-gray-400">Created {group.createdAt}</span>
                    <Link href={`/groups/${group.id}`}>
                      <button className="px-4 py-2 bg-green-500 text-white rounded-full hover:bg-green-600 transition-colors">
                        View Group
                      </button>
                    </Link>
                  </div>
                </div>
              </motion.div>
            ))}
          </div>
        </div>

        {/* Available Groups Section */}
        <div>
          <h2 className="text-2xl font-semibold text-gray-200 mb-6">Available Groups</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {availableGroups.map((group) => (
              <motion.div
                key={group.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                className="bg-white/10 backdrop-blur-lg rounded-lg shadow-lg overflow-hidden"
              >
                <div className="p-6">
                  <div className="flex items-center space-x-4">
                    <div className="w-16 h-16 bg-gray-700 rounded-full flex items-center justify-center">
                      {group.avatar ? (
                        <img
                          src={group.avatar}
                          alt={group.name}
                          className="w-full h-full rounded-full object-cover"
                        />
                      ) : (
                        <Users size={24} className="text-gray-400" />
                      )}
                    </div>
                    <div>
                      <h3 className="text-xl font-semibold text-gray-200">{group.name}</h3>
                      <p className="text-gray-400">{group.memberCount} members</p>
                    </div>
                  </div>
                  <p className="mt-4 text-gray-300">{group.description}</p>
                  <div className="mt-6 flex justify-between items-center">
                    <span className="text-sm text-gray-400">Created {group.createdAt}</span>
                    <button 
                      onClick={() => handleJoinGroup(group.id)}
                      className="px-4 py-2 bg-blue-500 text-white rounded-full hover:bg-blue-600 transition-colors"
                    >
                      Join Group
                    </button>
                  </div>
                </div>
              </motion.div>
            ))}
          </div>
        </div>
      </div>

      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center">
          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            className="bg-gray-800 rounded-lg p-8 w-full max-w-md"
          >
            <h2 className="text-2xl font-bold text-gray-200 mb-6">Create New Group</h2>
            <form onSubmit={handleCreateGroup} className="space-y-6">
              <div>
                <label className="block text-gray-300 mb-2">Group Name</label>
                <input
                  type="text"
                  value={newGroup.name}
                  onChange={(e) => setNewGroup({...newGroup, name: e.target.value})}
                  className="w-full px-4 py-2 bg-gray-700 rounded-lg text-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  required
                />
              </div>
              <div>
                <label className="block text-gray-300 mb-2">Description</label>
                <textarea
                  value={newGroup.description}
                  onChange={(e) => setNewGroup({...newGroup, description: e.target.value})}
                  className="w-full px-4 py-2 bg-gray-700 rounded-lg text-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  rows={4}
                  required
                />
              </div>
              <div>
                <label className="flex items-center space-x-3 text-gray-300">
                  <input
                    type="checkbox"
                    checked={newGroup.isPrivate}
                    onChange={(e) => setNewGroup({...newGroup, isPrivate: e.target.checked})}
                    className="w-5 h-5 rounded border-gray-600 text-blue-500 focus:ring-blue-500 focus:ring-offset-gray-800"
                  />
                  <span>Make this group private</span>
                </label>
                <p className="mt-1 text-sm text-gray-400 ml-8">
                  Private groups require admin approval to join and are only visible to members
                </p>
              </div>
              <div className="flex justify-end space-x-4">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-2 text-gray-300 hover:text-white"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
                >
                  Create Group
                </button>
              </div>
            </form>

          </motion.div>
        </div>
      )}
    </div>
  )
}
