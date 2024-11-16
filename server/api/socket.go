package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	m "social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"

	"github.com/gorilla/websocket"
)

// Create a global SocketManager instance
var socketManager = makeSocketManager()

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:3000"
	},
}

// Create a socket manager
func makeSocketManager() *m.SocketManager {
	return &m.SocketManager{
		Sockets: make(map[uint64]*websocket.Conn),
	}
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Get the user ID from the session
	userID, err := util.GetUserID(r, w)
	if err != nil {
		log.Printf("WebSocket auth error: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade the connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	log.Printf("New WebSocket connection for user %d", userID)

	// Add the connection to the socket manager
	AddConnection(socketManager, userID, conn)

	// Handle messages in a goroutine
	go func() {
		defer func() {
			RemoveConnection(socketManager, userID)
			conn.Close()
		}()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				break
			}
			HandleMessages(conn, userID, msg)
		}
	}()
}

func HandleMessages(conn *websocket.Conn, userID uint64, msg []byte) {
	// Validate UTF-8
	if !utf8.Valid(msg) {
		log.Println("Received invalid UTF-8 message")
		return
	}

	var message struct {
		Type        string `json:"type"`
		RecipientID int64  `json:"recipient_id"`
		Content     string `json:"content"`
	}

	if err := json.Unmarshal(msg, &message); err != nil {
		log.Printf("Error unmarshalling message: %v", err)
		return
	}

	// Ensure the content is valid UTF-8 (which includes emojis)
	if !utf8.Valid([]byte(message.Content)) {
		log.Printf("Invalid UTF-8 in message content")
		return
	}

	switch message.Type {
	case "chat":
		// Create timestamp for consistency
		now := time.Now()

		// Save message to database
		result, err := sqlite.DB.Exec(`
			INSERT INTO chat_messages (sender_id, recipient_id, content, created_at)
			VALUES (?, ?, ?, ?)
		`, userID, message.RecipientID, message.Content, now)
		if err != nil {
			log.Printf("Error saving chat message: %v", err)
			return
		}

		msgID, _ := result.LastInsertId()

		// Prepare response
		response := struct {
			Type        string    `json:"type"`
			ID          int64     `json:"id"`
			SenderID    int64     `json:"sender_id"`
			RecipientID int64     `json:"recipient_id"`
			Content     string    `json:"content"`
			CreatedAt   time.Time `json:"created_at"`
		}{
			Type:        "chat",
			ID:          msgID,
			SenderID:    int64(userID),
			RecipientID: message.RecipientID,
			Content:     message.Content,
			CreatedAt:   now,
		}

		// Send to recipient if online
		if recipientConn, ok := socketManager.Sockets[uint64(message.RecipientID)]; ok {
			if err := recipientConn.WriteJSON(response); err != nil {
				log.Printf("Error sending message to recipient: %v", err)
			}
		}

		// Send confirmation back to sender
		if err := conn.WriteJSON(response); err != nil {
			log.Printf("Error sending confirmation to sender: %v", err)
		}

		log.Printf("Message sent - From: %d, To: %d, Content: %s", userID, message.RecipientID, message.Content)

	case "typing":
		HandleTypingStatus(conn, userID, msg)
	}
}

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

// Add this function to broadcast notifications
func BroadcastNotification(notification m.Notification) {
	// Convert notification to JSON
	notificationJSON, err := json.Marshal(struct {
		Type string         `json:"type"`
		Data m.Notification `json:"data"`
	}{
		Type: "notification",
		Data: notification,
	})
	if err != nil {
		log.Printf("Error marshaling notification: %v", err)
		return
	}

	// Send to recipient if online
	if conn, ok := socketManager.Sockets[uint64(notification.ToUserID)]; ok {
		if err := conn.WriteMessage(websocket.TextMessage, notificationJSON); err != nil {
			log.Printf("Error sending notification: %v", err)
			RemoveConnection(socketManager, uint64(notification.ToUserID))
		}
	}
}

func AddConnection(sm *m.SocketManager, userID uint64, conn *websocket.Conn) {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	// If there's an existing connection, close it
	if existingConn, exists := sm.Sockets[userID]; exists {
		existingConn.Close()
	}

	sm.Sockets[userID] = conn
	log.Printf("Added new connection for user ID %d", userID)

	// Broadcast that this user is now online
	go BroadcastUserStatus(sm, userID, true)
}

func RemoveConnection(sm *m.SocketManager, userID uint64) {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	if conn, exists := sm.Sockets[userID]; exists {
		// Remove from map first to prevent any new messages
		delete(sm.Sockets, userID)
		
		// Close the connection
		conn.Close()
		
		log.Printf("Removed connection for user ID %d", userID)

		// Broadcast that this user is now offline
		go BroadcastUserStatus(sm, userID, false)
	}
}

func BroadcastUserStatus(sm *m.SocketManager, userID uint64, isOnline bool) {
	statusUpdate := struct {
		Type     string `json:"type"`
		UserID   uint64 `json:"user_id"`
		IsOnline bool   `json:"is_online"`
	}{
		Type:     "user_status",
		UserID:   userID,
		IsOnline: isOnline,
	}

	message, err := json.Marshal(statusUpdate)
	if err != nil {
		log.Printf("Error marshalling status update: %v", err)
		return
	}

	Broadcast(sm, message)
}

func Broadcast(sm *m.SocketManager, message []byte) {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	for userID, conn := range sm.Sockets {
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Error broadcasting to user ID %d: %v", userID, err)
			RemoveConnection(sm, userID)
		}
	}
}

func GetOnlineUsers(sm *m.SocketManager) []uint64 {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	onlineUsers := make([]uint64, 0, len(sm.Sockets))
	for userID := range sm.Sockets {
		onlineUsers = append(onlineUsers, userID)
	}
	return onlineUsers
}

// Update the ClearNotification function
func ClearNotification(w http.ResponseWriter, r *http.Request) {
	// Get the notification ID from the URL
	notificationIDStr := r.PathValue("id")
	notificationID, err := strconv.Atoi(notificationIDStr)
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	// Get the current user's ID
	userID, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete the notification from the database
	result, err := sqlite.DB.Exec(`
		DELETE FROM notifications 
		WHERE id = ? AND to_user_id = ?
	`, notificationID, userID)
	if err != nil {
		log.Printf("Error deleting notification: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Notification not found or unauthorized", http.StatusNotFound)
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": "Notification deleted",
	})
}

// Update the ClearAllNotifications function
func ClearAllNotifications(w http.ResponseWriter, r *http.Request) {
	userID, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete all notifications for the user
	result, err := sqlite.DB.Exec(`
		DELETE FROM notifications 
		WHERE to_user_id = ?
	`, userID)
	if err != nil {
		log.Printf("Error deleting all notifications: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": fmt.Sprintf("Deleted %d notifications", rowsAffected),
	})
}
