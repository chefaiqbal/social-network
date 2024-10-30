package models

import "time"

type Chat_message struct {
	SenderID int `json:"sender_id"`
	RecipientID int `json:"recipient_id"`
	Content string `json:"content,omitempty"`
	UserName string `json:"user_name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Group_messages struct {
	ID int `json:"id"`
	GroupID int `json:"group_id"`
	UserID int `json:"user_id"`
	Content string `json:"content,omitempty"`
	Media string `json:"media,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}