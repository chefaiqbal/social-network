export interface Member {
  id: number;
  name: string;
  avatar?: string;
  role: "creator" | "member";
  status: "online" | "offline";
}

const fetchPendingUsers = async (groupId: number): Promise<Member[]> => {
  const payload = { group_id: groupId };
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

  return Array.isArray(data)
    ? data.map((member) => ({
        id: member.user_id,
        name: member.username,
        avatar: member.avatar || null,
        role: member.status,
        status: member.status === "creator" ? "online" : "offline",
      }))
    : [];
};

export default fetchPendingUsers;
