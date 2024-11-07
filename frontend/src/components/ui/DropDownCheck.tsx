'use client';

import { useState, useEffect } from 'react';
import Button from '@/components/ui/Button';

function DropDownCheck() {
  const [followedUsers, setFollowedUsers] = useState<{ username: string }[]>([]);
  const [selectedItems, setSelectedItems] = useState<string[]>([]);
  const [isExpanded, setIsExpanded] = useState(false); // Track if the dropdown is expanded

  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await fetch('http://localhost:8080/followers', {
          credentials: 'include',
        });

        if (!response.ok) {
          throw new Error('Failed to fetch followed users');
        }

        const data = await response.json();
        console.log('Fetched followed users:', data);

        if (Array.isArray(data)) {
          setFollowedUsers(data);
        } else {
          console.error('Data is not an array:', data);
        }
      } catch (error) {
        console.error('Error fetching followed users:', error);
      }
    };

    fetchData();
  }, []);

  const handleCheckboxChange = (name: string) => {
    setSelectedItems((prev) => {
      if (prev.includes(name)) {
        return prev.filter(item => item !== name);
      } else {
        return [...prev, name];
      }
    });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    console.log('Selected items:', selectedItems);
  
    try {
      const response = await fetch('http://localhost:8080/CloseFriend', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include', // Include cookies if necessary
        body: JSON.stringify({ selectedUsers: selectedItems }), // Send selected users
      });
  
      // Log the response as text first to check its contents
      const textResponse = await response.text();
      console.log('Response text:', textResponse); // Log the response as text to check for issues
  
      if (response.ok) {
        try {
          const result = JSON.parse(textResponse); // Try to parse as JSON
          console.log('Response from backend:', result);
          alert(result.message || 'Success!');
        } catch (jsonError) {
          console.error('Error parsing JSON:', jsonError);
          alert('Failed to parse response. Please try again.');
        }
      } else {
        alert(`Error: ${textResponse}`);
      }
    } catch (error) {
      console.error('Error submitting selected users:', error);
      alert('Failed to submit selected users.');
    }
  };
  

  const toggleDropdown = () => {
    setIsExpanded((prev) => !prev);
  };

  // Display the list of selected close friends (optional, for user feedback)
  const selectedFriendsDisplay = selectedItems.join(', ');

  return (
    <form onSubmit={handleSubmit} className="relative">
      <div onClick={toggleDropdown} className="cursor-pointer bg-gray-800/50 border border-gray-700/50 rounded-lg p-2 text-gray-200 w-full">
        Choose close friends <span className="text-xs ml-1">{isExpanded ? '▲' : '▼'}</span>
      </div>
      {isExpanded && (
        <div className="mt-2 bg-gray-800/50 border border-gray-700/50 rounded-lg p-2 text-gray-200 shadow-lg max-h-64 overflow-y-auto z-10">
          <ul className="w-full">
            {followedUsers.length > 0 ? (
              followedUsers.map((user, index) => (
                <li key={index} className="p-2 hover:bg-gray-700 cursor-pointer">
                  <label className="flex items-center w-full">
                    <input
                      type="checkbox"
                      checked={selectedItems.includes(user.username)}
                      onChange={() => handleCheckboxChange(user.username)}
                      className="mr-2"
                    />
                    {user.username}
                  </label>
                </li>
              ))
            ) : (
              <li className="p-2 text-gray-400">No followers available.</li>
            )}
          </ul>
        </div>
      )}

      {/* Display selected friends */}
      {selectedItems.length > 0 && (
        <div className="mt-2 text-gray-200">
          <h3>Selected Close Friends:</h3>
          <p>{selectedFriendsDisplay}</p>
        </div>
      )}

      <div className="mt-4">
        <Button type="submit" className="w-full">Submit</Button>
      </div>
    </form>
  );
}

export default DropDownCheck;
