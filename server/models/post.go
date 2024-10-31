package models

import "time"

type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Media     string    `json:"media,omitempty"`
	Privacy   int       `json:"privacy"`
	CreatedAt time.Time `json:"created_at"`
	Author    uint      `json:"author"`
	GroupID   *int64    `json:"group_id,omitempty"`
}

type PostPrivateView struct {
    ID     int `json:"id"`      
    PostID int `json:"post_id"`  
    UserID int `json:"user_id"`  
}
