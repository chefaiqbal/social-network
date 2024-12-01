'use client';

import { useState, useEffect } from 'react';
import { Bell, User, X } from 'lucide-react';
import Link from 'next/link';
import { useRouter } from 'next/router';

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
  group_id: number;
}

export default function Header() {
  const [currentUser, setCurrentUser] = useState<CurrentUser | null>(null);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [showNotifications, setShowNotifications] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [wsConnected, setWsConnected] = useState(false);
  const router = useRouter();

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
      setIsLoading(true);
      setError(null);
      try {
        const response = await fetch('http://localhost:8080/notifications', {
          credentials: 'include'
        });
        if (!response.ok) {
          throw new Error('Failed to fetch notifications');
        }
        const data = await response.json();
        setNotifications(Array.isArray(data) ? data : []);
      } catch (error) {
        setError('Unable to load notifications');
        console.error('Error fetching notifications:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchNotifications();

    const ws = new WebSocket('ws://localhost:8080/ws');
    
    ws.onopen = () => {
      setWsConnected(true);
    };
    
    ws.onclose = () => {
      setWsConnected(false);
      // Attempt to reconnect after 5 seconds
      setTimeout(() => {
        if (!wsConnected) {
          ws.close();
          new WebSocket('ws://localhost:8080/ws');
        }
      }, 5000);
    };
    
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'notification') {
        setNotifications(prev => {
          const exists = prev.some(n => n.id === data.data.id);
          if (!exists) {
            return [data.data, ...prev];
          }
          return prev;
        });
      }
    };

    return () => {
      ws.close();
    };
  }, []);

  const handleSignOut = async () => {
    try {
      const response = await fetch('http://localhost:8080/logout', {
        method: 'POST',
        credentials: 'include',
      });

      if (response.ok) {
        window.location.href = '/';
      }
    } catch (error) {
      console.error('Error signing out:', error);
    }
  };

  const handleFollowRequest = async (notificationId: number, action: 'accept' | 'reject') => {
    const response = await fetch(`http://localhost:8080/follow/request/${notificationId}`, {
      method: 'PATCH',
      credentials: 'include',
      headers: { 
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      },
      body: JSON.stringify({ status: action }),
    });

    if (response.ok) {
      setNotifications(prev => prev.filter(n => n.id !== notificationId));
      // Refresh relevant pages if needed
    }
  };

  const clearNotification = async (notificationId: number) => {
    const response = await fetch(`http://localhost:8080/notifications/${notificationId}/clear`, {
      method: 'POST',
      credentials: 'include',
    });
    
    if (response.ok) {
      setNotifications(prev => prev.filter(n => n.id !== notificationId));
    }
  };

  const clearAllNotifications = async () => {
    const response = await fetch('http://localhost:8080/notifications/clear-all', {
      method: 'POST',
      credentials: 'include',
    });

    if (response.ok) {
      setNotifications([]);
    }
  };

  const handleNewNotification = (notification: Notification) => {
    if (Notification.permission === 'granted') {
      new Notification('New Notification', {
        body: notification.content,
        icon: '/app-icon.png'
      });
    }
    
    setNotifications(prev => {
      const exists = prev.some(n => n.id === notification.id);
      if (!exists) {
        return [notification, ...prev];
      }
      return prev;
    });
  };

  useEffect(() => {
    if (Notification.permission === 'default') {
      Notification.requestPermission();
    }
  }, []);

  const getGroupedNotifications = () => {
    const today = new Date().toDateString();
    
    return notifications.reduce((groups, notification) => {
      const date = new Date(notification.created_at).toDateString();
      const key = date === today ? 'Today' : date;
      
      if (!groups[key]) {
        groups[key] = [];
      }
      groups[key].push(notification);
      return groups;
    }, {} as Record<string, Notification[]>);
  };

  const markAsRead = async (notificationId: number) => {
    try {
      const response = await fetch(`http://localhost:8080/notifications/${notificationId}/read`, {
        method: 'POST',
        credentials: 'include',
      });
      
      if (response.ok) {
        setNotifications(prev => 
          prev.map(n => 
            n.id === notificationId ? { ...n, read: true } : n
          )
        );
      }
    } catch (error) {
      console.error('Error marking notification as read:', error);
    }
  };

  return (
<header className="fixed top-0 left-0 right-0 bg-white/10 backdrop-blur-lg shadow-sm border-b border-gray-700/50 z-50">
  <div className="w-full px-8 py-4 flex items-center justify-between">
    <Link 
      href="/feed"
      className="text-2xl font-bold bg-gradient-to-r from-blue-400 via-purple-500 to-pink-500 bg-clip-text text-transparent ml-4"
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
                      {Object.entries(getGroupedNotifications()).map(([date, notifications]) => (
                        <div key={date}>
                          <div className="text-sm text-gray-400 px-2 py-1">{date}</div>
                          {notifications.map(notification => (
                            <div 
                              key={notification.id}
                              className="relative p-2 rounded bg-gray-800/50 text-gray-200 group cursor-pointer"
                              onClick={() => {
                                if (notification.type === 'notification_event') {
                                  console.log(notification.group_id, "Navigating to group ID");
                                  router.push(`/groups/${notification.group_id}`);
                                }
                              }}
                            >
                              <button
                            onClick={(e) => {
                              e.stopPropagation(); // Prevent triggering parent click
                              clearNotification(notification.id);
                            }}
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