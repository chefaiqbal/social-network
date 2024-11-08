package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"social-network/util"
	"social-network/models"
	"social-network/pkg/db/sqlite"
)

func UserProfile(w http.ResponseWriter, r *http.Request) {
	// Get the requested user ID from the URL
	userIdString := r.PathValue("userID")
	var targetUserID int

	// Check if we're requesting the current user's profile
	if userIdString == "current" {
		// Get the current user's ID from the session
		currentUserID, err := util.GetUserID(r, w)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		targetUserID = int(currentUserID)
	} else {
		// Convert id to number
		var err error
		targetUserID, err = strconv.Atoi(userIdString)
		if err != nil {
			http.Error(w, "Error processing user ID", http.StatusBadRequest)
			return
		}
	}

	var userInfo models.User
	var avatar sql.NullString
	var aboutMe sql.NullString
	if err := sqlite.DB.QueryRow(
		"SELECT id, email, username, first_name, last_name, date_of_birth, avatar, about_me, is_private, created_at FROM users WHERE id = ?",
		targetUserID).Scan(
		&userInfo.ID,
		&userInfo.Email,
		&userInfo.Username,
		&userInfo.FirstName,
		&userInfo.LastName,
		&userInfo.DateOfBirth,
		&avatar,
		&aboutMe,
		&userInfo.IsPrivate,
		&userInfo.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User does not exist", http.StatusNotFound)
			return
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error getting user info: %v", err)
		return
	}

	if avatar.Valid {
		userInfo.Avatar = avatar.String
	}

	if aboutMe.Valid {
		userInfo.AboutMe = aboutMe.String
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&userInfo); err != nil {
		http.Error(w, "Error sending data", http.StatusInternalServerError)
	}
}

func GetSuggestedUsers(w http.ResponseWriter, r *http.Request) {
	// Get current user ID from session
	userID, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Query for users that the current user is not following
	rows, err := sqlite.DB.Query(`
		SELECT DISTINCT u.id, u.username, u.avatar 
		FROM users u 
		WHERE u.id NOT IN (
			SELECT followed_id 
			FROM followers 
			WHERE follower_id = ? AND status = 'accept'
		) 
		AND u.id != ?
		AND u.id IN (
			SELECT id FROM users WHERE id != 1
		)
		LIMIT 5
	`, userID, userID)
	
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Error querying suggested users: %v", err)
		return
	}
	defer rows.Close()

	var suggestedUsers []struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
			Avatar   string `json:"avatar,omitempty"`
			Online   bool   `json:"online"`
	}

	// Get current online users
	onlineUsers := make(map[uint64]bool)
	for _, id := range GetOnlineUsers(socketManager) {
		onlineUsers[id] = true
	}

	for rows.Next() {
		var user struct {
			ID       uint   `json:"id"`
			Username string `json:"username"`
			Avatar   string `json:"avatar,omitempty"`
			Online   bool   `json:"online"`
		}
		var avatar sql.NullString
		if err := rows.Scan(&user.ID, &user.Username, &avatar); err != nil {
			http.Error(w, "Error scanning users", http.StatusInternalServerError)
			log.Printf("Error scanning user row: %v", err)
			return
		}
		if avatar.Valid {
			user.Avatar = avatar.String
		}
		// Check if user is online
		user.Online = onlineUsers[uint64(user.ID)]
		suggestedUsers = append(suggestedUsers, user)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, "Error iterating users", http.StatusInternalServerError)
		log.Printf("Error iterating users: %v", err)
		return
	}

	// If no suggested users found, return empty array instead of null
	if suggestedUsers == nil {
		suggestedUsers = make([]struct {
			ID       uint   `json:"id"`
			Username string `json:"username"`
			Avatar   string `json:"avatar,omitempty"`
			Online   bool   `json:"online"`
		}, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(suggestedUsers); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
