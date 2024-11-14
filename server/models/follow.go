package models

import "time"

type Follow struct {
	ID         uint      `json:"id,omitempty"`
	FollowerID uint      `json:"follower_id,omitempty"`
	FollowedID uint      `json:"followed_id,omitempty"`
	Status     string    `json:"status,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	Username   string    `json:"username,omitempty"`  // Added Username field
}

type CloseFriends struct {
    Usernames []string `json:"selectedUsers"`
}

type FollowRequest struct {
    ID         int    `json:"id"`
    FollowedID int    `json:"followed_id"`
    Username   string `json:"username"`
    Avatar     string `json:"avatar"` // Add Avatar field
}
