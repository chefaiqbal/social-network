'use client';

import { useEffect, useState } from 'react';

interface GroupNameProps {
    groupID: number;
}

const GroupName = ({ groupID }: GroupNameProps) => {
    const [groupName, setGroupName] = useState('');

    useEffect(() => {
        const fetchGroupName = async () => {
            try {
                const response = await fetch(`http://localhost:8080/groups/${groupID}`, {
                    credentials: 'include',
                });
                if (response.ok) {
                    const data = await response.json();
                    setGroupName(data.group_name); 
                    console.log(data.group_name);
                } else {
                    console.error('Failed to fetch group name');
                }
            } catch (error) {
                console.error('Error fetching group name:', error);
            }
        };

        fetchGroupName();
    }, [groupID]);

    return (
        <div>
            <h1
                className="text-2xl font-bold bg-gradient-to-r from-blue-400 via-purple-500 to-pink-500 bg-clip-text text-transparent ml-4"
            >{groupName}</h1>
        </div>
    );
}

export default GroupName;
