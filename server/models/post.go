package models

import (
	"time"
)

type Post struct {
	ID           int       `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Media        string    `json:"media,omitempty"`
	Privacy      int       `json:"privacy"`
	Author       int       `json:"author"`
	AuthorName   string    `json:"author_name"`
	AuthorAvatar string    `json:"author_avatar,omitempty"`
	CreatedAt    string    `json:"created_at"`
	GroupID      int       `json:"group_id,omitempty"`
	LikeCount    int       `json:"like_count"`
	UserLiked    bool      `json:"user_liked"`
}

type PostResponse struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	MediaBase64  string    `json:"media,omitempty"`      // Base64 encoded for JSON response
	MediaType    string    `json:"media_type,omitempty"` // MIME type
	Privacy      int       `json:"privacy"`
	Author       int64     `json:"author"`
	AuthorName   string    `json:"author_name"`
	AuthorAvatar string    `json:"author_avatar,omitempty"`
	GroupID      *int64    `json:"group_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	LikeCount    int       `json:"like_count"`
	UserLiked    bool      `json:"user_liked"`
}
