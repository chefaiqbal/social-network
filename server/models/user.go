package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID          uint      `json:"id,omitempty"`
	Email       string    `json:"email,omitempty"`
	Password    string    `json:"password,omitempty"`
	FirstName   string    `json:"first_name,omitempty"`
	LastName    string    `json:"last_name,omitempty"`
	DateOfBirth time.Time `json:"date_of_birth,omitempty"`
	Avatar      string    `json:"avatar,omitempty"`
	Username    string    `json:"username,omitempty"`
	AboutMe     string    `json:"about_me,omitempty"`
	IsPrivate   bool      `json:"is_private,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

type UserResponse struct {
	ID       int64  `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
}

func DoesUserExist(userID uint, db *sql.DB) (bool, error) {
	var id int

	if err := db.QueryRow("SELECT id FROM users WHERE id = ?", userID).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
