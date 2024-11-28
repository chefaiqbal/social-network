import React, { useEffect, useState } from "react";
import fetchPendingUsers, { Member } from "@/lib/GetPendingMembers";

interface PendingMembersProps {
  groupId: number;
  onAccept: (memberId: number) => void;
  onReject: (memberId: number) => void;
}

const PendingMembers: React.FC<PendingMembersProps> = ({
  groupId,
  onAccept,
  onReject,
}) => {
  const [pendingMembers, setPendingMembers] = useState<Member[]>([]);
  const [error, setError] = useState<string | null>(null);

  const fetchPendingUsersList = async () => {
    try {
      const members = await fetchPendingUsers(groupId);
      setPendingMembers(members);
      setError(null);
    } catch (err: any) {
      console.error("Failed to fetch pending members:", err);
      setError(err.message || "An unknown error occurred");
    }
  };

  useEffect(() => {
    if (groupId) fetchPendingUsersList();
  }, [groupId]);

  const handleAccept = async (memberId: number) => {
    await onAccept(memberId);
    fetchPendingUsersList(); 
  };

  const handleReject = async (memberId: number) => {
    await onReject(memberId);
    fetchPendingUsersList(); 
  };

  return (
    <div>
      <h2 className="text-2xl font-semibold text-gray-200 mb-4">
        Pending Members
      </h2>
      {error ? (
        <div className="text-red-500 mb-4">{error}</div>
      ) : (
        <div className="space-y-4">
          {pendingMembers.length > 0 ? (
            pendingMembers.map((member) => (
              member && (
                <div
                  key={member.id}
                  className="flex items-center justify-between bg-gray-800/50 p-4 rounded-lg shadow-md"
                >
                  <div className="flex items-center space-x-4">
                    <div className="w-10 h-10 bg-gray-700 rounded-full flex items-center justify-center">
                      <img
                        src={member.avatar || "/default-avatar.png"}
                        alt={member.name}
                        className="w-full h-full rounded-full object-cover"
                      />
                    </div>
                    <div>
                      <h3 className="text-lg font-semibold text-gray-200">
                        {member.name}
                      </h3>
                      <div className="flex space-x-2">
                        <button
                          onClick={() => handleAccept(member.id)}
                          className="px-3 py-1 bg-green-500 text-white rounded-full hover:bg-green-600 text-sm"
                        >
                          Accept
                        </button>
                        <button
                          onClick={() => handleReject(member.id)}
                          className="px-3 py-1 bg-red-500 text-white rounded-full hover:bg-red-600 text-sm"
                        >
                          Reject
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              )
            ))
          ) : (
            <div className="text-gray-500">
              No pending members at the moment.
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default PendingMembers;