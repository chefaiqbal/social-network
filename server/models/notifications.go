package models

import "time"

type Notification struct {
    ID         int       `json:"id"`          
    ToUserID   int       `json:"to_user_id"`  
    Content    string    `json:"content"`     
    FromUserID int       `json:"from_user_id"`
    Read       bool      `json:"read"`       
    GroupID    int      `json:"group_id,omitempty"` 
	CreatedAt  time.Time `json:"created_at"`  

}
