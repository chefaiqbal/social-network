package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"log"
	"social-network/pkg/db/sqlite"
	"social-network/util"
	//"github.com/gorilla/websocket"
)

type ChatMessage struct {
	ID          int64     `json:"id"`
	SenderID    int64     `json:"sender_id"`
	RecipientID int64     `json:"recipient_id"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
}

func GetChatUsers(w http.ResponseWriter, r *http.Request) {
	userID, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get users that the current user is following or are following them
	rows, err := sqlite.DB.Query(`
		SELECT DISTINCT u.id, u.username, u.avatar
		FROM users u
		JOIN followers f ON (f.follower_id = ? AND f.followed_id = u.id)
			OR (f.follower_id = u.id AND f.followed_id = ?)
		WHERE f.status = 'accept'
	`, userID, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []struct {
		ID       int    `json:"id"`
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
			ID       int    `json:"id"`
			Username string `json:"username"`
			Avatar   string `json:"avatar,omitempty"`
			Online   bool   `json:"online"`
		}
		var avatar sql.NullString
		if err := rows.Scan(&user.ID, &user.Username, &avatar); err != nil {
			http.Error(w, "Error scanning users", http.StatusInternalServerError)
			return
		}
		if avatar.Valid {
			user.Avatar = avatar.String
		}
		// Check if user is online
		user.Online = onlineUsers[uint64(user.ID)]
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func GetChatMessages(w http.ResponseWriter, r *http.Request) {
	userID, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	otherUserID, err := strconv.ParseInt(r.URL.Query().Get("userId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20 // Default limit
	}

	offset := (page - 1) * limit

	// Modified query to fetch messages in both directions
	rows, err := sqlite.DB.Query(`
		SELECT id, sender_id, recipient_id, content, created_at
		FROM chat_messages
		WHERE (sender_id = ? AND recipient_id = ?)
			OR (sender_id = ? AND recipient_id = ?)
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, userID, otherUserID, otherUserID, userID, limit, offset)

	if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var msg ChatMessage
		if err := rows.Scan(&msg.ID, &msg.SenderID, &msg.RecipientID, &msg.Content, &msg.CreatedAt); err != nil {
			log.Printf("Scan error: %v", err)
			http.Error(w, "Error scanning messages", http.StatusInternalServerError)
			return
		}
		messages = append(messages, msg)
	}

	// Reverse the messages to maintain chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
