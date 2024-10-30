package models

type Likes struct {
	ID int `json:"id"`
	UserID int `json:"user_id"`
	PostID int `json:"post_id,omitempty"`
	CommentID  int `json:"comment_id,omitempty"`	
	Like bool `json:"like"`
}