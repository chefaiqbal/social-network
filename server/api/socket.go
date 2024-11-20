package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	m "social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
	"github.com/gorilla/websocket"
)

// Shared upgrader for all WebSocket connections
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Message types
const (
	MessageTypeNotification = "notification"
	MessageTypeUserStatus   = "user_status"
)

// Create a global SocketManager instance
var socketManager = makeSocketManager()

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

	// Use the shared upgrader
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}

	log.Printf("New WebSocket connection for user %d", userID)

	// Add the connection to the socket manager
	AddConnection(socketManager, userID, ws)

	// Handle messages in a goroutine
	go func() {
		defer func() {
			RemoveConnection(socketManager, userID)
			ws.Close()
		}()

		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				break
			}
			HandleMessages(userID, msg)
		}
	}()
}

func HandleMessages(userID uint64, msg []byte) {
	var message struct {
		Type        string `json:"type"`
		RecipientID int64  `json:"recipient_id"`
		Content     string `json:"content"`
	}

	if err := json.Unmarshal(msg, &message); err != nil {
		log.Printf("Error unmarshalling message: %v", err)
		return
	}

	log.Printf("Received message of type: %s from user %d", message.Type, userID)

	switch message.Type {
	case MessageTypeNotification:
		handleNotification(userID, msg)
	case MessageTypeUserStatus:
		// Handle user status updates
		BroadcastUserStatus(socketManager, userID, true)
	default:
		log.Printf("Unknown message type received: %s", message.Type)
	}
}

func handleNotification(userID uint64, msg []byte) {
	var notification m.Notification
	if err := json.Unmarshal(msg, &notification); err != nil {
		log.Printf("Error unmarshaling notification: %v", err)
		return
	}

	// Broadcast the notification
	BroadcastNotification(notification)
}

// Update BroadcastNotification function
func BroadcastNotification(notification m.Notification) {
	log.Printf("Broadcasting notification: %+v", notification)

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
		log.Printf("Sending notification to user %d", notification.ToUserID)
		if err := conn.WriteMessage(websocket.TextMessage, notificationJSON); err != nil {
			log.Printf("Error sending notification: %v", err)
			RemoveConnection(socketManager, uint64(notification.ToUserID))
		}
	} else {
		log.Printf("User %d is not online", notification.ToUserID)
	}
}

func AddConnection(sm *m.SocketManager, userID uint64, conn *websocket.Conn) {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	// If there's an existing connection, close it properly
	if existingConn, exists := sm.Sockets[userID]; exists {
		// Remove from map first
		delete(sm.Sockets, userID)
		// Then close the connection
		existingConn.Close()
		log.Printf("Closed existing connection for user %d", userID)
	}

	// Add the new connection
	sm.Sockets[userID] = conn
	log.Printf("Added new connection for user %d", userID)

	// Broadcast that this user is now online
	go BroadcastUserStatus(sm, userID, true)
}

func RemoveConnection(sm *m.SocketManager, userID uint64) {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	if conn, exists := sm.Sockets[userID]; exists {
		// Remove from map first
		delete(sm.Sockets, userID)
		// Then close the connection
		conn.Close()
		log.Printf("Removed connection for user %d", userID)

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
		Type:     MessageTypeUserStatus,
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
		"status":  "success",
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
		"status":  "success",
		"message": fmt.Sprintf("Deleted %d notifications", rowsAffected),
	})
}
