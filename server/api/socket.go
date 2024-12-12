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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Define Client type
type Client struct {
	Conn   *websocket.Conn
	UserID int
}

// Define a map to store clients
var clients = make(map[*Client]bool)

// Message types
const (
	MessageTypeNotification = "notification"
)

// Create a global SocketManager instance
var socketManager = makeSocketManager()

func makeSocketManager() *m.SocketManager {
	return &m.SocketManager{
		Sockets: make(map[uint64]*websocket.Conn),
	}
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := util.GetUserID(r, w)
	if err != nil {
		log.Printf("WebSocket auth error: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}

	client := &Client{
		Conn:   ws,
		UserID: int(userID),
	}

	// Add client to both maps
	clients[client] = true
	AddConnection(socketManager, uint64(userID), ws)


	go func() {
		defer func() {
			delete(clients, client)
			RemoveConnection(socketManager, uint64(userID))
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
			HandleMessages(uint64(userID), msg)
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

func BroadcastNotification(notification m.Notification) {
	message := struct {
		Type string         `json:"type"`
		Data m.Notification `json:"data"`
	}{
		Type: "notification",
		Data: notification,
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling notification: %v", err)
		return
	}

	log.Printf("Broadcasting notification message: %s", string(messageJSON))

	// Broadcast using both methods to ensure delivery
	// Method 1: Using clients map
	for client := range clients {
		if client.UserID == notification.ToUserID {
			err := client.Conn.WriteMessage(websocket.TextMessage, messageJSON)
			if err != nil {
				log.Printf("Error sending notification to client: %v", err)
				client.Conn.Close()
				delete(clients, client)
			}
		}
	}

	// Method 2: Using socket manager
	if conn, exists := socketManager.Sockets[uint64(notification.ToUserID)]; exists {
		err := conn.WriteMessage(websocket.TextMessage, messageJSON)
		if err != nil {
			log.Printf("Error sending notification via socket manager: %v", err)
			RemoveConnection(socketManager, uint64(notification.ToUserID))
		}
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

	}

	// Add the new connection
	sm.Sockets[userID] = conn

}

func RemoveConnection(sm *m.SocketManager, userID uint64) {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	if conn, exists := sm.Sockets[userID]; exists {
		// Remove from map first
		delete(sm.Sockets, userID)
		// Then close the connection
		conn.Close()

	}
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
