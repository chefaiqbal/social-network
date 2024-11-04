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
	userIdString := r.PathValue("userID")

	// Convert id to number
	userID, err := strconv.Atoi(userIdString)
	if err != nil {
		http.Error(w, "Error processing user ID", http.StatusBadRequest)
		return
	}

	var userInfo models.User
	var avatar sql.NullString // Handle nullable avatar
	var aboutMe sql.NullString
	if err := sqlite.DB.QueryRow(
		"SELECT id, email, username, first_name, last_name, date_of_birth, avatar, about_me, is_private, created_at FROM users WHERE id = ?",
		userID).Scan(
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
			http.Error(w, "User does not exist", http.StatusBadRequest)
			return
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error getting user info: %v", err)
		return
	}

	// Check if the avatar is valid (i.e., not null)
	if avatar.Valid {
		userInfo.Avatar = avatar.String // Use the value if it's valid
	}

	if aboutMe.Valid {
		userInfo.AboutMe = aboutMe.String
	}

	// Encode the userInfo as JSON and send it as a response
	if err := json.NewEncoder(w).Encode(&userInfo); err != nil {
		http.Error(w, "Error sending data", http.StatusInternalServerError)
	}
}

func GetSuggestedUsers(w http.ResponseWriter, r *http.Request) {
	// Get current user ID from session
	cookie, err := r.Cookie("AccessToken")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUserID := util.UserSession[cookie.Value]

	// Query for users that the current user is not following
	rows, err := sqlite.DB.Query(`
		SELECT u.id, u.username, u.avatar 
		FROM users u 
		WHERE u.id NOT IN (
			SELECT followed_id 
			FROM followers 
			WHERE follower_id = ? AND status = 'accept'
		) 
		AND u.id != ?
		LIMIT 5`, currentUserID, currentUserID)
	
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var suggestedUsers []struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Avatar   string `json:"avatar,omitempty"`
	}

	for rows.Next() {
		var user struct {
			ID       uint   `json:"id"`
			Username string `json:"username"`
			Avatar   string `json:"avatar,omitempty"`
		}
		var avatar sql.NullString
		if err := rows.Scan(&user.ID, &user.Username, &avatar); err != nil {
			http.Error(w, "Error scanning users", http.StatusInternalServerError)
			return
		}
		if avatar.Valid {
			user.Avatar = avatar.String
		}
		suggestedUsers = append(suggestedUsers, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestedUsers)
}
