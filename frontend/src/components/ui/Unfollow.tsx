'use client'
// A red button to unfollow

import { useEffect, useState } from 'react';

interface UnfollowButtonProps {
    fetchProfile: () => void;
    followed_id: number;
}

const UnfollowAction = async (fetchProfile: () => void, followed_id: number) => {
    try {
        const response = await fetch('http://localhost:8080/Unfollow', {
            method: 'POST',
            credentials: 'include',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                followed_id: followed_id,
            }),
        });
        if (response.ok) {
            fetchProfile();
            window.location.reload(); // Refresh the page
        }
    } catch (error) {
        console.error('Error unfollowing user:', error);
    }
};

const UnfollowButton = ({ fetchProfile, followed_id }: UnfollowButtonProps) => {
    const [isFollower, setIsFollower] = useState(false);

    useEffect(() => {
        const fetchFollowers = async () => {
            try {
                const response = await fetch('http://localhost:8080/following', {
                    method: 'GET',
                    credentials: 'include',
                });
                if (response.ok) {
                    const followers = await response.json();
                    const isFollowing = followers.some((follower: { followed_id: number }) => follower.followed_id === followed_id);
                    setIsFollower(isFollowing);
                }
            } catch (error) {
                console.error('Error fetching followers:', error);
            }
        };

        fetchFollowers();
    }, [followed_id]);

    if (!isFollower) {
        return null;
    }

    return (
        <button
            className="w-full bg-red-500 p-2 rounded hover:bg-red-700"
            onClick={() => UnfollowAction(fetchProfile, followed_id)}
        >
            Unfollow
        </button>
    );
};

export default UnfollowButton;
