// RightSidebar Component
import { motion } from 'framer-motion';
import DropDownCheck from '@/components/ui/DropDownCheck'
import {useEffect, useState} from 'react'

const RightSidebar = () => {
    const [suggestedUsers, setSuggestedUsers] = useState<any[]>([]);
    const [groups, setGroups] = useState<any[]>([]);
    const [loggedInUserId, setLoggedInUserId] = useState<number | null>(null);
    const [loading, setLoading] = useState<boolean>(true);
  
    // Fetch the logged-in user's ID
    const loginUserID = async () => {
      try {
        const response = await fetch('http://localhost:8080/userIDBY', {
          method: 'GET',
          credentials: 'include', 
          headers: {
            'Content-Type': 'application/json',
          },
        });
  
        if (!response.ok) {
          throw new Error(`Failed to fetch user ID, Status: ${response.status}`);
        }
  
        const data = await response.json();
        console.log('Fetched user ID:', data);
  
        if (data.userID) {
          setLoggedInUserId(data.userID);
        } else {
          throw new Error('User ID not found in response');
        }
      } catch (error) {
        console.error('Error fetching user ID:', error);
        setLoggedInUserId(null);
      }
    };
  
    // fetch the user ID for the given username
    const fetchFollowedID = async (username: string): Promise<number | null> => {
      try {
        const response = await fetch('http://localhost:8080/userID', {
          method: 'POST',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ username }),
        });
  
        if (!response.ok) {
          throw new Error(`Failed to fetch user ID, Status: ${response.status}`);
        }
  
        const data = await response.json();
        console.log('Fetched user ID:', data);
        return data.userID || null;
      } catch (error) {
        console.error('Error fetching user ID:', error);
        return null;
      }
    };
  
    // Handle following a user
    const handleFollow = async (followedName: string) => {
      if (!loggedInUserId) {
        console.error('Logged-in user ID is not available');
        return;
      }
  
      const followedID = await fetchFollowedID(followedName);
  
      if (followedID === null) {
        console.error('Failed to get followed user ID');
        return;
      }
  
      console.log(`User ID: ${loggedInUserId}, Followed ID: ${followedID}`);
  
      try {
        const response = await fetch('http://localhost:8080/follow', {
          method: 'POST',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            follower_id: loggedInUserId,
            followed_id: followedID,
            status: 'pending',
          }),
        });
  
        if (!response.ok) {
          throw new Error('Failed to follow user');
        }
  
        // Optionally refresh the suggested users list
        fetchSuggestedUsers();
      } catch (error) {
        console.error('Error following user:', error);
      }
    };
  
    // Fetch suggested users
    const fetchSuggestedUsers = async () => {
      try {
        const response = await fetch('http://localhost:8080/users/suggested', {
          credentials: 'include',
        });
        if (response.ok) {
          const data = await response.json();
          setSuggestedUsers(data);
        }
      } catch (error) {
        console.error('Error fetching suggested users:', error);
      }
    };
  
    // Fetch groups
    const fetchGroups = async () => {
      try {
        const response = await fetch('http://localhost:8080/groups', {
          credentials: 'include',
        });
        if (response.ok) {
          const data = await response.json();
          setGroups(data);
        }
      } catch (error) {
        console.error('Error fetching groups:', error);
      }
    };
  
    useEffect(() => {
      const fetchData = async () => {
        await loginUserID(); // Fetch the logged-in user's ID
        await fetchSuggestedUsers(); // Fetch suggested users
        await fetchGroups(); // Fetch groups
        setLoading(false); // Set loading to false after all data is fetched
      };
      fetchData();
    }, []);
  
    if (loading) {
      return <div>Loading...</div>; // Optional: Show loading state while data is being fetched
    }
  
    return (
      <div className="fixed right-0 top-0 h-screen w-96 pt-20 border-l border-gray-700/50 overflow-y-auto">
        <div className="p-6">
          <div className="bg-white/10 backdrop-blur-lg shadow-lg rounded-lg p-6 border border-gray-700/50">
            <h2 className="text-lg font-semibold mb-6 text-gray-200">Suggested Followers</h2>
            <div className="space-y-6 mb-8">
              {suggestedUsers.map((user: any, index) => (
                <motion.div
                  key={user.id}
                  initial={{ opacity: 0, x: 20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: index * 0.1 }}
                  className="flex items-center justify-between"
                >
                  <div className="flex items-center flex-1 min-w-0">
                    <div className="w-10 h-10 bg-gray-700 rounded-full mr-3 flex-shrink-0">
                      {user.avatar && (
                        <img
                          src={user.avatar}
                          alt={user.username}
                          className="w-full h-full rounded-full object-cover"
                        />
                      )}
                    </div>
                    <p className="font-medium text-gray-300 truncate">{user.username}</p>
                  </div>
                  <button
                    onClick={() => handleFollow(user.username)}
                    className="ml-4 px-4 py-1 text-sm text-blue-400 hover:text-blue-300 border border-blue-400/50 rounded-full hover:bg-blue-400/10 transition-colors flex-shrink-0"
                  >
                    Follow
                  </button>
                </motion.div>
              ))}
            </div>
  
            <h2 className="text-lg font-semibold mb-6 text-gray-200">Active Groups</h2>
            <div className="space-y-6">
              {groups.map((group: any, index) => (
                <motion.div
                  key={group.id}
                  initial={{ opacity: 0, x: 20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: (index + 3) * 0.1 }}
                  className="flex items-center justify-between"
                >
                  <div className="flex-1 min-w-0">
                    <p className="font-medium text-gray-300 truncate">{group.title}</p>
                    <p className="text-sm text-gray-400">{group.member_count} members</p>
                  </div>
                  <button className="ml-4 px-4 py-1 text-sm text-green-400 hover:text-green-300 border border-green-400/50 rounded-full hover:bg-green-400/10 transition-colors flex-shrink-0">
                    Join
                  </button>
                </motion.div>
              ))}
            </div>
          </div>
          <div className="mt-6">
            <DropDownCheck />
          </div>
        </div>
      </div>
    );
  };
  
  export default RightSidebar;