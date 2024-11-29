'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Users, Plus, Search } from 'lucide-react'
import Link from 'next/link'
import Header from '@/components/layout/Header'
import Sidebar from '@/components/layout/Sidebar'

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
  const [loading, setLoading] = useState(false)
  const [userGroups, setUserGroups] = useState<Group[]>([]) // State for the user's groups
  const [error, setError] = useState<string | null>(null)

  const fetchGroups = async () => {
    setLoading(true)
    setError(null)
    try {
      const response = await fetch('http://localhost:8080/groups', {
        credentials: 'include',
      });
      if (response.ok) {
        const data = await response.json();
        const formattedGroups = data.map((group: any) => ({
          id: group.id,
          name: group.title,
          description: group.description,
          isPrivate: group.is_private || false,
          createdAt: new Date(group.created_at).toLocaleString(),
        }));
        setGroups(formattedGroups);
      } else {
        setError('Failed to fetch groups.');
      }
    } catch (error) {
      setError('Error fetching groups.');
    } finally {
      setLoading(false)
    }
  };

  const handleJoinGroup = async (groupId: number) => {
    try {
      const response = await fetch('http://localhost:8080/groups/JoinRequest', {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ groupId }),
      });
      if (response.ok) {
        const data = await response.json();
        console.log('Joined group:', data.message);
      } else {
        console.error("Error joining group");
      }
    } catch (error) {
      console.error('Request failed:', error);
    }

    fetchGroups();
  };

  const handleCreateGroup = async (e: React.FormEvent) => {
    e.preventDefault();
  
    const createdGroup: Group = {
      id: Date.now(),
      name: newGroup.name, 
      description: newGroup.description,
      isPrivate: false, 
      memberCount: 1, 
      createdAt: new Date().toISOString(),
    };
  
    try {
      const response = await fetch('http://localhost:8080/groups', {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          title: newGroup.name, 
          description: newGroup.description,
        }),
      });
  
      if (response.ok) {
        setGroups(prevGroups => [createdGroup, ...prevGroups]); 
        setNewGroup({ name: '', description: '', isPrivate: false });
        setShowCreateModal(false);
      } else {
        setError('Failed to create group.');
      }
    } catch (error) {
      setError('Error creating group.');
    }
    fetchGroups();
    Mygroups(); 
  };

  // Function to fetch groups the user is a member of
  const Mygroups = async () => {
    try {
        const response = await fetch('http://localhost:8080/groups/myGroup', {
          credentials: 'include',
        });

        if (response.ok) {
          const data = await response.json();
          console.log('My groups:', data);
          const formattedGroups = data.map((group: any) => ({
            id: group.id,
            name: group.title,
            description: group.description,
            isPrivate: group.is_private || false,
            createdAt: new Date(group.created_at).toLocaleString(),
          }));
          setUserGroups(formattedGroups); // Store in state
        } else {
          console.error("Error fetching my groups");
        }
    } catch (error) {
      console.error('Error fetching groups:', error);
    }
  }

  useEffect(() => {
    fetchGroups();
    Mygroups(); // Call the Mygroups function to fetch user-specific groups
  }, []);

  const filteredGroups = groups.filter(group =>
    (group.name?.toLowerCase().includes(searchQuery.toLowerCase()) || 
     group.description?.toLowerCase().includes(searchQuery.toLowerCase()))
  );  

  const availableGroups = filteredGroups.filter(group => !group.isMember);

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
      <Header />
      <div className="flex pt-16">
        <Sidebar />
        <div className="flex-1 overflow-y-auto">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
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

             {/* Create Group Modal */}
             {showCreateModal && (
              <div className="fixed inset-0 flex items-center justify-center bg-black/50 z-50">
                <div className="bg-gray-800 p-8 rounded-lg max-w-sm w-full">
                  <h3 className="text-xl text-gray-200 mb-4">Create Group</h3>
                  <form onSubmit={handleCreateGroup}>
                    <input
                      type="text"
                      placeholder="Group Name"
                      value={newGroup.name}
                      onChange={(e) => setNewGroup({ ...newGroup, name: e.target.value })}
                      className="w-full mb-4 px-4 py-2 bg-gray-700 text-white rounded-md"
                      required
                    />
                    <textarea
                      placeholder="Group Description"
                      value={newGroup.description}
                      onChange={(e) => setNewGroup({ ...newGroup, description: e.target.value })}
                      className="w-full mb-4 px-4 py-2 bg-gray-700 text-white rounded-md"
                      required
                    />
                    <div className="flex justify-between">
                      <button
                        type="button"
                        onClick={() => setShowCreateModal(false)}
                        className="px-4 py-2 bg-gray-600 text-white rounded-full hover:bg-gray-700"
                      >
                        Cancel
                      </button>
                      <button
                        type="submit"
                        className="px-4 py-2 bg-blue-500 text-white rounded-full hover:bg-blue-600"
                      >
                        Create
                      </button>
                    </div>
                  </form>
                </div>
              </div>
            )}
            

            {/* Search Bar */}
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

            {/* Loading or Error State */}
            {loading && <div className="text-gray-200">Loading...</div>}
            {error && <div className="text-red-500">{error}</div>}

            {/* My Groups Section */}
            <div className="mb-12">
              <h2 className="text-2xl font-semibold text-gray-200 mb-6">My Groups</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {userGroups.length > 0 ? (
                  userGroups.map((group) => (
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
                        <div className="mt-4 flex justify-between">
                          <Link
                            href={`/groups/${group.id}`}
                            className="text-blue-500 hover:text-blue-600"
                          >
                            View Group
                          </Link>
                          {/* {!group.isMember && (
                            <button
                              onClick={() => handleJoinGroup(group.id)}
                              className="px-4 py-2 bg-blue-500 text-white rounded-full hover:bg-blue-600"
                            >
                              Join Group
                            </button>
                          )} */}
                        </div>
                      </div>
                    </motion.div>
                  ))
                ) : (
                  <div className="text-gray-500">You are not a member of any groups yet.</div>
                )}
              </div>
            </div>

            {/* Available Groups */}
            <div className="mb-12">
              <h2 className="text-2xl font-semibold text-gray-200 mb-6">Available Groups</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {availableGroups.length > 0 ? (
                  availableGroups.map((group) => (
                    <motion.div
                      key={group.id}
                      initial={{ opacity: 0, y: 20 }}
                      animate={{ opacity: 1, y: 0 }}
                      className="bg-white/10 backdrop-blur-lg rounded-lg shadow-lg overflow-hidden"
                    >
                      <div className="p-6">
                        <h3 className="text-xl font-semibold text-gray-200">{group.name}</h3>
                        <p className="text-gray-300">{group.description}</p>
                        <div className="mt-4 flex justify-between">
                          <Link
                            href={`/groups/${group.id}`}
                            className="text-blue-500 hover:text-blue-600"
                          >
                            View Group
                          </Link>
                          <button
                            onClick={() => handleJoinGroup(group.id)}
                            className="px-4 py-2 bg-blue-500 text-white rounded-full hover:bg-blue-600"
                          >
                            Join Group
                          </button>
                        </div>
                      </div>
                    </motion.div>
                  ))
                ) : (
                  <div className="text-gray-500">No available groups to join.</div>
                )}
              </div>
              
            </div>

          </div>
        </div>
      </div>
    </div>
  );
}
