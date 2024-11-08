package api

import (
	"encoding/json"
	"log"
	"net/http"
	m "social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
	"time"

	"github.com/gorilla/websocket"
)

// Create a global SocketManager instance
var socketManager = makeSocketManager()

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, you should check the origin more carefully
		return true
	},
}

// Create a socket manager
func makeSocketManager() *m.SocketManager {
	return &m.SocketManager{
		Sockets: make(map[uint64]*websocket.Conn),
	}
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {

    userID, err := util.GetUserID(r, w)
    if err != nil {
        http.Error(w, "problem in getting user id", http.StatusUnauthorized)
        return
    }

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        http.Error(w, "Could not upgrade to WebSocket", http.StatusInternalServerError)
        return
    }

    log.Println("User ID:", userID)

    AddConnection(socketManager, userID, conn)

    go func() {
        HandleMessages(conn, userID)
        RemoveConnection(socketManager, userID)
    }()
}

func HandleMessages(conn *websocket.Conn, userID uint64) {
    for {
        _, msg, err := conn.ReadMessage()
        if err != nil {
            log.Printf("Error reading message: %v", err)
            break
        }

        var message struct {
            Type        string `json:"type"`
            RecipientID int64  `json:"recipient_id"`
            Content     string `json:"content"`
        }

        if err := json.Unmarshal(msg, &message); err != nil {
            log.Printf("Error unmarshalling message: %v", err)
            continue
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
                continue
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

        case "typing":
            HandleTypingStatus(conn, userID, msg)
        }
    }
}


// all functions about notification

//notfication for one user
func SendNotificationOne(sm *m.SocketManager, message []byte) {
	var notification m.Notification
	if err := json.Unmarshal(message, &notification); err != nil {
		log.Println("Error unmarshalling message:", err)
		return
	}
	
}

//function to like and deslike 
func MakeLikeDeslike(sm *m.SocketManager, message []byte) {
    var like m.Likes
    if err := json.Unmarshal(message, &like); err != nil {
        log.Println("Error unmarshalling message:", err)
        return
    }

    if like.PostID != 0 {
        if like.Like {
            _, err := sqlite.DB.Exec("INSERT INTO likes (user_id, post_id, is_like) VALUES (?, ?, ?)", like.UserID, like.PostID, like.Like)
            if err != nil {
                log.Println("Error inserting like:", err)
                return
            }
        } else {
            _, err := sqlite.DB.Exec("DELETE FROM likes WHERE user_id = ? AND post_id = ?", like.UserID, like.PostID)
            if err != nil {
                log.Println("Error removing like:", err)
                return
            }
        }

        broadcastMsgJSON, err := json.Marshal(like)
        if err != nil {
            log.Println("Error marshalling like for broadcast:", err)
            return
        }

        Broadcast(sm, broadcastMsgJSON)

    } else if like.CommentID != 0 {
        if like.Like {
            _, err := sqlite.DB.Exec("INSERT INTO likes (user_id, comment_id, is_like) VALUES (?, ?, ?)", like.UserID, like.CommentID, like.Like)
            if err != nil {
                log.Println("Error inserting like:", err)
                return
            }
        } else {
            _, err := sqlite.DB.Exec("DELETE FROM likes WHERE user_id = ? AND comment_id = ?", like.UserID, like.CommentID)
            if err != nil {
                log.Println("Error removing like:", err)
                return
            }
        }

        BroadcastMsg, err := json.Marshal(like)
        if err != nil {
            log.Println("Error marshalling like for broadcast:", err)
            return
        }

        Broadcast(sm, BroadcastMsg)
    } else {
        log.Println("Invalid like request")
        return
    }
}

// function to send message chat 
func SendMessage(sm *m.SocketManager, message []byte) {
	var chatMessage m.Chat_message
	if err := json.Unmarshal(message, &chatMessage); err != nil {
		log.Println("Error unmarshalling message:", err)
		return
	}

	chatMessage.CreatedAt = time.Now()

	// Insert the message into the database
	query := `INSERT INTO chat_messages (sender_id, recipient_id, content, created_at) VALUES (?, ?, ?, ?)`
	_, err := sqlite.DB.Exec(query, chatMessage.SenderID, chatMessage.RecipientID, chatMessage.Content, chatMessage.CreatedAt)
	if err != nil {
		log.Println("Error inserting message:", err)
		return
	}

	// Send the message to the client using userID
	responseMessage, err := json.Marshal(chatMessage)
	if err != nil {
		log.Println("Error marshalling chat message for sending:", err)
		return
	}

	// Lock the SocketManager while sending the message
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	// Send the message to the specific recipient
	if conn, exists := sm.Sockets[uint64(chatMessage.RecipientID)]; exists {
		if err := conn.WriteMessage(websocket.TextMessage, responseMessage); err != nil {
			log.Printf("Error sending message to user %d: %v", uint64(chatMessage.RecipientID), err)
			RemoveConnection(sm, uint64(chatMessage.RecipientID))
		}
	} else {
		log.Printf("No active connection for recipient ID %d", uint64(chatMessage.RecipientID))
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

func GroupChat(sm *m.SocketManager, message []byte) {
	var GroupChat m.Group_messages
	if err := json.Unmarshal(message, &GroupChat); err != nil {
		log.Println("Error unmarshalling message:", err)
		return
	}

    GroupChat.CreatedAt = time.Now();

    // Insert the message into the database
    // query := `INSERT INTO group_chat_messages (group_id, sender_id, content, created_at) VALUES (?, ?, ?, ?)`
}


// connection functions
func RemoveConnection(sm *m.SocketManager, userID uint64) {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	if conn, exists := sm.Sockets[userID]; exists {
		conn.Close()
		delete(sm.Sockets, userID)
		log.Printf("Removed connection for user ID %d", userID)

		// Broadcast that this user is now offline
		go BroadcastUserStatus(sm, userID, false)
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

// BroadcastUserStatus sends online status updates to all connected users
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

// Add a new function to get all online users
func GetOnlineUsers(sm *m.SocketManager) []uint64 {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	onlineUsers := make([]uint64, 0, len(sm.Sockets))
	for userID := range sm.Sockets {
		onlineUsers = append(onlineUsers, userID)
	}
	return onlineUsers
}
