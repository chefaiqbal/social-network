package models

import (
	//"database/sql"
	"time"
)

type Post struct {
	ID        int64          `json:"id"`
	Title     string         `json:"title"`
	Content   string         `json:"content"`
	Media     []byte         `json:"media,omitempty"`
	MediaType string         `json:"media_type,omitempty"`
	Privacy   int            `json:"privacy"`
	Author    int64          `json:"author"`
	GroupID   *int64         `json:"group_id,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
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
}
