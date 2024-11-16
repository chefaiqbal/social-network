'use client';

import { useState, useEffect, useRef } from 'react';
import { Bell, User, X } from 'lucide-react';
import Link from 'next/link';

interface CurrentUser {
  id: number;
  username: string;
  avatar?: string;
}

interface Notification {
  id: number;
  content: string;
  created_at: string;
  from_username: string;
  from_avatar?: string;
  type: string;
  read: boolean;
}

export default function Header() {
  const [currentUser, setCurrentUser] = useState<CurrentUser | null>(null);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [showNotifications, setShowNotifications] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

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

  useEffect(() => {
    const fetchNotifications = async () => {
      try {
        const response = await fetch('http://localhost:8080/notifications', {
          credentials: 'include'
        })
        if (response.ok) {
          const data = await response.json()
          setNotifications(Array.isArray(data) ? data : [])
        }
      } catch (error) {
        console.error('Error fetching notifications:', error)
      }
    }

    fetchNotifications()

    wsRef.current = new WebSocket('ws://localhost:8080/ws')
    
    wsRef.current.onmessage = (event) => {
      const data = JSON.parse(event.data)
      if (data.type === 'notification') {
        setNotifications(prev => [data.data, ...prev])
      }
    }

    return () => {
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [])

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

  const handleFollowRequest = async (notificationId: number, action: 'accept' | 'reject') => {
    try {
      console.log('Handling follow request:', notificationId, action);

      // First, get the follow request ID from the notification
      const notification = notifications.find(n => n.id === notificationId);
      if (!notification) {
        console.error('Notification not found');
        return;
      }

      const response = await fetch(`http://localhost:8080/follow/request/${notificationId}`, {
        method: 'PATCH',
        credentials: 'include',
        headers: { 
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        },
        body: JSON.stringify({ 
          status: action
        }),
      });

      if (response.ok) {
        // Remove the notification from the list
        setNotifications(prev => prev.filter(n => n.id !== notificationId));
        
        // Show success message
        console.log(`Successfully ${action}ed follow request`);

        // Refresh the page if we're on the profile page
        if (window.location.pathname.includes('/profile/')) {
          window.location.reload();
        }

        // Fetch updated user data if needed
        if (window.location.pathname.includes('/follow')) {
          // Trigger a refresh of the users list
          window.dispatchEvent(new CustomEvent('refreshUsers'));
        }
      } else {
        const errorData = await response.text();
        console.error('Failed to handle follow request:', errorData);
      }
    } catch (error) {
      console.error('Error handling follow request:', error);
    }
  };

  const clearNotification = async (notificationId: number) => {
    try {
      const response = await fetch(`http://localhost:8080/notifications/${notificationId}/clear`, {
        method: 'POST',
        credentials: 'include',
      });

      if (response.ok) {
        // Remove the notification from the state
        setNotifications(prev => prev.filter(n => n.id !== notificationId));
      }
    } catch (error) {
      console.error('Error clearing notification:', error);
    }
  };

  const clearAllNotifications = async () => {
    try {
      const response = await fetch('http://localhost:8080/notifications/clear-all', {
        method: 'POST',
        credentials: 'include',
      });

      if (response.ok) {
        // Clear all notifications from the state
        setNotifications([]);
      }
    } catch (error) {
      console.error('Error clearing all notifications:', error);
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
                  <div className="flex justify-between items-center mb-4">
                    <h3 className="text-lg font-semibold text-gray-200">Notifications</h3>
                    {notifications.length > 0 && (
                      <button
                        onClick={clearAllNotifications}
                        className="text-sm text-gray-400 hover:text-gray-200"
                      >
                        Clear all
                      </button>
                    )}
                  </div>
                  {notifications.length > 0 ? (
                    <div className="space-y-2">
                      {notifications.map((notification) => (
                        <div 
                          key={notification.id}
                          className="relative p-2 rounded bg-gray-800/50 text-gray-200 group"
                        >
                          <button
                            onClick={() => clearNotification(notification.id)}
                            className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 text-gray-400 hover:text-gray-200"
                          >
                            <X size={16} />
                          </button>
                          <p>{notification.content}</p>
                          
                          {notification.type === 'follow_request' && (
                            <div className="flex space-x-2 mt-2">
                              <button
                                onClick={() => handleFollowRequest(notification.id, 'accept')}
                                className="px-3 py-1 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
                              >
                                Accept
                              </button>
                              <button
                                onClick={() => handleFollowRequest(notification.id, 'reject')}
                                className="px-3 py-1 bg-red-600 text-white rounded-md hover:bg-red-700 text-sm"
                              >
                                Decline
                              </button>
                            </div>
                          )}

                          <div className="flex items-center justify-between mt-1">
                            <span className="text-xs text-gray-400">
                              {new Date(notification.created_at).toLocaleTimeString()}
                            </span>
                            {!notification.read && (
                              <span className="text-xs text-blue-400">New</span>
                            )}
                          </div>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p className="text-gray-400 text-center">No notifications</p>
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