'use client';

import { useState, useEffect } from 'react';
import { Bell, User } from 'lucide-react';
import Link from 'next/link';

interface CurrentUser {
  id: number;
  username: string;
  avatar?: string;
}

export default function Header() {
  const [currentUser, setCurrentUser] = useState<CurrentUser | null>(null);
  const [notifications, setNotifications] = useState([]);
  const [showNotifications, setShowNotifications] = useState(false);

  useEffect(() => {
    const fetchCurrentUser = async () => {
      try {
        const response = await fetch('http://localhost:8080/user/current', {
          credentials: 'include',
        });
        if (response.ok) {
          const userData = await response.json();
          setCurrentUser(userData);
        }
      } catch (error) {
        console.error('Error fetching user data:', error);
      }
    };

    fetchCurrentUser();
  }, []);

  const handleSignOut = async () => {
    try {
      const response = await fetch('http://localhost:8080/logout', {
        method: 'POST',
        credentials: 'include',
      });

      if (response.ok) {
        window.location.href = '/';
      } else {
        console.error('Failed to sign out');
      }
    } catch (error) {
      console.error('Error signing out:', error);
    }
  };

  return (
    <header className="fixed top-0 left-0 right-0 bg-white/10 backdrop-blur-lg shadow-sm border-b border-gray-700/50 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4 flex items-center justify-between">
        <Link 
          href="/feed"
          className="text-2xl font-bold bg-gradient-to-r from-blue-400 via-purple-500 to-pink-500 bg-clip-text text-transparent"
        >
          Reboot Network
        </Link>

        <div className="flex items-center gap-6">
          <div className="relative">
            <button 
              onClick={() => setShowNotifications(!showNotifications)}
              className="text-gray-300 hover:text-white transition-colors relative"
            >
              <Bell size={20} />
              {notifications.length > 0 && (
                <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full w-4 h-4 flex items-center justify-center">
                  {notifications.length}
                </span>
              )}
            </button>
            
            {showNotifications && (
              <div className="absolute right-0 mt-2 w-80 rounded-lg shadow-lg bg-white/10 backdrop-blur-lg border border-gray-700/50">
                <div className="p-4">
                  {notifications.length > 0 ? (
                    <div className="space-y-2">
                      {/* Render notifications here */}
                    </div>
                  ) : (
                    <p className="text-gray-400 text-center">No new notifications</p>
                  )}
                </div>
              </div>
            )}
          </div>

          <div className="relative group">
            <button className="flex items-center space-x-2 hover:bg-gray-800/50 rounded-lg p-2 transition-colors">
              {currentUser?.avatar ? (
                <img
                  src={currentUser.avatar}
                  alt={currentUser.username}
                  className="w-8 h-8 rounded-full object-cover border border-gray-700/50"
                />
              ) : (
                <div className="w-8 h-8 bg-gray-700 rounded-full flex items-center justify-center">
                  <User size={18} className="text-gray-300" />
                </div>
              )}
              <span className="text-gray-200">{currentUser?.username}</span>
            </button>
            <div className="absolute right-0 mt-2 w-48 rounded-md shadow-lg bg-white/10 backdrop-blur-lg border border-gray-700/50 opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 transform origin-top-right">
              <div className="py-1">
                <Link 
                  href="/profile" 
                  className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-800/50 transition-colors"
                >
                  Profile
                </Link>
                <button
                  onClick={handleSignOut}
                  className="block w-full text-left px-4 py-2 text-sm text-red-400 hover:bg-gray-800/50 transition-colors"
                >
                  Sign out
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
} 