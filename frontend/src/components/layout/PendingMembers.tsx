import React, { useEffect, useState } from "react";

interface Member {
  id: number;
  name: string;
  avatar?: string;
  role: "creator" | "member";
  status: "online" | "offline";
}

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

  const fetchPendingUsers = async () => {
    try {
      const payload = { group_id: groupId }; // Send groupId as an integer
      console.log("Payload being sent:", JSON.stringify(payload));
  
      const response = await fetch("http://localhost:8080/groups/pendingUsers", {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });
  
      console.log("Response Status:", response.status);
  
      if (!response.ok) {
        const errorText = await response.text();
        console.error("Error fetching pending members:", errorText);
        throw new Error(
          `Error fetching pending members: ${response.statusText} - ${errorText}`
        );
      }
  
      const data: any[] = await response.json();
      console.log("Fetched Data:", data);
  
      
      const formattedMembers: Member[] = Array.isArray(data) ? data.map((member) => ({
        id: member.user_id,
        name: member.username,
        avatar: member.avatar || null,
        role: member.status,
        status: member.status === "creator" ? "online" : "offline",
      })) : [];
  
      setPendingMembers(formattedMembers);
      setError(null);
    } catch (err: any) {
      console.error("Failed to fetch pending members:", err);
      setError(err.message || "An unknown error occurred");
    }
  };

  useEffect(() => {
    if (groupId) fetchPendingUsers();
  }, [groupId]);

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
                          onClick={() => onAccept(member.id)}
                          className="px-3 py-1 bg-green-500 text-white rounded-full hover:bg-green-600 text-sm"
                        >
                          Accept
                        </button>
                        <button
                          onClick={() => onReject(member.id)}
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