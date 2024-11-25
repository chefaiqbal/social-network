package models

import "time"

type Comment struct {
	ID        uint      `json:"id,omitempty"`
	Content   string    `json:"content,omitempty"`
	Media     []byte    `json:"-"`
	MediaType string    `json:"media_type,omitempty"`
	MediaBase64 string  `json:"media,omitempty"`
	Post_ID   uint      `json:"post_id,omitempty"`
	Author    uint      `json:"author,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	AuthorName   string `json:"author_name,omitempty"`
	AuthorAvatar string `json:"author_avatar,omitempty"`
}

type CommentResponse struct {
	ID           uint      `json:"id"`
	Content      string    `json:"content"`
	MediaBase64  string    `json:"media,omitempty"`
	MediaType    string    `json:"media_type,omitempty"`
	PostID       uint      `json:"post_id"`
	Author       uint      `json:"author"`
	CreatedAt    time.Time `json:"created_at"`
	AuthorName   string    `json:"author_name"`
	AuthorAvatar string    `json:"author_avatar,omitempty"`
}
