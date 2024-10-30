package models

import "time"

type Post struct {
	ID        uint      `json:"id,omitempty"`
	Title     string    `json:"title,omitempty"`
	Content   string    `json:"content,omitempty"`
	Media     string    `json:"media,omitempty"`
	Privay    uint      `json:"privay,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	Author    uint      `json:"author,omitempty"`
	GroupID   uint      `json:"group_id,omitempty"`
}

type PostPrivateView struct {
    ID     int `json:"id"`      
    PostID int `json:"post_id"`  
    UserID int `json:"user_id"`  
}
