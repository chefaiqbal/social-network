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
	"github.com/gorilla/websocket"
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
		// Check if user is online by looking up their WebSocket connection
		_, user.Online = socketManager.Sockets[uint64(user.ID)]
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

	otherUserID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
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
		limit = 8 // Default limit
	}

	offset := (page - 1) * limit

	rows, err := sqlite.DB.Query(`
		SELECT id, sender_id, recipient_id, content, created_at
		FROM chat_messages
		WHERE (sender_id = ? AND recipient_id = ?)
			OR (sender_id = ? AND recipient_id = ?)
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, userID, otherUserID, otherUserID, userID, limit, offset)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var msg ChatMessage
		if err := rows.Scan(&msg.ID, &msg.SenderID, &msg.RecipientID, &msg.Content, &msg.CreatedAt); err != nil {
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

// HandleChatMessage processes incoming chat messages from WebSocket
func HandleChatMessage(conn *websocket.Conn, userID uint64, msg []byte) {
	var chatMsg struct {
		Type        string `json:"type"`
		RecipientID int64  `json:"recipient_id"`
		Content     string `json:"content"`
	}

	if err := json.Unmarshal(msg, &chatMsg); err != nil {
		log.Printf("Error unmarshaling chat message: %v", err)
		return
	}

	// Create timestamp for consistency
	now := time.Now()

	// Save message to database
	result, err := sqlite.DB.Exec(`
		INSERT INTO chat_messages (sender_id, recipient_id, content, created_at)
		VALUES (?, ?, ?, ?)
	`, userID, chatMsg.RecipientID, chatMsg.Content, now)
	if err != nil {
		log.Printf("Error saving chat message: %v", err)
		return
	}

	msgID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID: %v", err)
		return
	}

	// Verify the message was saved
	var savedMsg ChatMessage
	err = sqlite.DB.QueryRow(`
		SELECT id, sender_id, recipient_id, content, created_at
		FROM chat_messages
		WHERE id = ?
	`, msgID).Scan(&savedMsg.ID, &savedMsg.SenderID, &savedMsg.RecipientID, &savedMsg.Content, &savedMsg.CreatedAt)
	if err != nil {
		log.Printf("Error verifying saved message: %v", err)
		return
	}

	// Prepare response using the saved message data
	response := struct {
		Type        string    `json:"type"`
		ID          int64     `json:"id"`
		SenderID    int64     `json:"sender_id"`
		RecipientID int64     `json:"recipient_id"`
		Content     string    `json:"content"`
		CreatedAt   time.Time `json:"created_at"`
	}{
		Type:        "chat",
		ID:          savedMsg.ID,
		SenderID:    savedMsg.SenderID,
		RecipientID: savedMsg.RecipientID,
		Content:     savedMsg.Content,
		CreatedAt:   savedMsg.CreatedAt,
	}

	// Send to recipient if online
	if recipientConn, ok := socketManager.Sockets[uint64(chatMsg.RecipientID)]; ok {
		if err := recipientConn.WriteJSON(response); err != nil {
			log.Printf("Error sending message to recipient: %v", err)
		}
	}

	// Send confirmation back to sender
	if err := conn.WriteJSON(response); err != nil {
		log.Printf("Error sending confirmation to sender: %v", err)
	}

	log.Printf("Message saved and sent - ID: %d, From: %d, To: %d", msgID, userID, chatMsg.RecipientID)
}

// HandleTypingStatus processes typing status updates
func HandleTypingStatus(conn *websocket.Conn, userID uint64, msg []byte) {
	var typingStatus struct {
		Type        string `json:"type"`
		RecipientID int64  `json:"recipient_id"`
		Typing      bool   `json:"typing"`
	}

	if err := json.Unmarshal(msg, &typingStatus); err != nil {
		log.Printf("Error unmarshaling typing status: %v", err)
		return
	}

	// Forward typing status to recipient if online
	if recipientConn, ok := socketManager.Sockets[uint64(typingStatus.RecipientID)]; ok {
		if err := recipientConn.WriteJSON(typingStatus); err != nil {
			log.Printf("Error sending typing status: %v", err)
		}
	}
}
