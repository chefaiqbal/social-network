import React, { useEffect, useState } from 'react';
import { FaCheck, FaTimes, FaUsers } from 'react-icons/fa';

const GroupInvitations = () => {
  const [isVisible, setIsVisible] = useState(false);
  const [invitations, setInvitations] = useState<Invitation[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  interface Invitation {
    id: number;
    group_id: number;
    title: string;
    status: string;
    created_at: string;
    avatar?: string;
  }

  useEffect(() => {
    async function fetchInvitations() {
      try {
        const response = await fetch(`http://localhost:8080/groups/invitation`, {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
        });

        if (!response.ok) {
          if (response.status === 204) {
            setInvitations([]);
            return;
          }
          throw new Error(`Failed to fetch invitations: ${response.status}`);
        }

        const data = await response.json();
        setInvitations(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load invitations');
      } finally {
        setLoading(false);
      }
    }

    fetchInvitations();
  }, []);

  const invitationCount = invitations.length;

  if (loading) return <p>Loading invitations...</p>;
  if (error) return <p>Error: {error}</p>;
  if (!isVisible && invitationCount === 0) return <p>No invitations found.</p>;

  const handleAccept = async (groupId: number) => {
    try {
      const response = await fetch(`http://localhost:8080/groups/invitation/accept`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({ groupId }),
      });

      if (!response.ok) throw new Error('Failed to accept invitation');

      // Update UI after accepting
      setInvitations((prev) => prev.filter((invite) => invite.group_id !== groupId));
    } catch (err) {
      alert('Error accepting invitation: ' + (err instanceof Error ? err.message : 'Unknown error'));
    }
  };

  const handleDecline = async (groupId: number) => {
    try {
      const response = await fetch(`http://localhost:8080/groups/invitation/decline`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({ groupId }),
      });

      if (!response.ok) throw new Error('Failed to decline invitation');

      // Update UI after declining
      setInvitations((prev) => prev.filter((invite) => invite.group_id !== groupId));
    } catch (err) {
      alert('Error declining invitation: ' + (err instanceof Error ? err.message : 'Unknown error'));
    }
  };



  if (loading) return <p className="text-gray-400">Loading invitations...</p>;
  if (error) return <p className="text-red-400">Error: {error}</p>;
  if (!isVisible && invitations.length === 0) return <p className="text-gray-400">No invitations found.</p>;
  return (
    <div className="mb-12">
      {/* Header */}
      <div
        className="flex items-center justify-between cursor-pointer text-gray-200 hover:text-gray-50 transition"
        onClick={() => setIsVisible(!isVisible)}
      >
        <h2 className="text-2xl font-semibold">
          Group Invitations
          {invitationCount > 0 && (
            <span className="ml-2 bg-blue-500 text-white text-sm px-2 py-1 rounded-full">
              {invitationCount}
            </span>
          )}
        </h2>
        <div className="ml-2">
          {isVisible ? (
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              fill="currentColor"
              className="bi bi-chevron-up"
              viewBox="0 0 16 16"
            >
              <path d="M8 5.293l3.5 3.5a.5.5 0 1 1-.708.707L8 6.707 4.207 9.5a.5.5 0 1 1-.707-.707L8 5.293z" />
            </svg>
          ) : (
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              fill="currentColor"
              className="bi bi-chevron-down"
              viewBox="0 0 16 16"
            >
              <path d="M8 10.707L4.207 7.5a.5.5 0 1 1 .707-.707L8 9.293l3.792-3.793a.5.5 0 1 1 .708.707L8 10.707z" />
            </svg>
          )}
        </div>
      </div>

      {/* Collapsible Content */}
      {isVisible && (
        <div className="mt-6 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {invitations.map((invite) => (
            <div
              key={invite.id}
              className="bg-white/10 backdrop-blur-lg rounded-lg shadow-lg overflow-hidden hover:shadow-xl transition-shadow duration-300 p-6"
            >
              {/* Avatar and Title */}
              <div className="flex items-center mb-4">
                {invite.avatar ? (
                  <img
                    src={invite.avatar}
                    alt={invite.title}
                    className="w-12 h-12 rounded-full mr-4"
                  />
                ) : (
                  <div className="w-12 h-12 rounded-full bg-gray-700 mr-4 flex items-center justify-center text-gray-400 font-bold">
                    {invite.title.charAt(0).toUpperCase()}
                  </div>
                )}
                <h3 className="text-lg font-bold text-gray-50">{invite.title}</h3>
              </div>

              {/* Actions */}
              <div className="flex justify-between gap-4">
                <button
                  className="flex items-center justify-center gap-2 flex-grow px-4 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600 transition"
                  onClick={() => handleAccept(invite.group_id)}
                >
                  <FaCheck />
                  Accept
                </button>
                <button
                  className="flex items-center justify-center gap-2 flex-grow px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 transition"
                  onClick={() => handleDecline(invite.group_id)}
                >
                  <FaTimes />
                  Decline
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default GroupInvitations;