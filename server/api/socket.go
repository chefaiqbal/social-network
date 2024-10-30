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
}

// Create a socket manager
func makeSocketManager() *m.SocketManager {
	return &m.SocketManager{
		Sockets: make(map[uint64]*websocket.Conn),
	}
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {

    username, err := util.GetUsernameFromSession(r)
    if err != nil {
        http.Error(w, "Unauthorized: no session cookie", http.StatusUnauthorized)
        return
    }

    var userID uint64
    err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
    if err != nil {
        http.Error(w, "Unauthorized: user not found", http.StatusUnauthorized)
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
        var connectionType m.ConnectionType

        _, message, err := conn.ReadMessage()
        if err != nil {
            log.Println("Error reading message:", err)
            break
        }

        // Handle the message here
        if err := json.Unmarshal(message, &connectionType); err != nil {
            log.Println("Error unmarshalling message:", err)
            continue
        }

        switch connectionType.Type {
        case "notification":
            // Handle notification message
        case "chat":
            // Handle chat message
            SendMessage(socketManager, message)
        case "groupChat":
            // Handle group chat message
        case "like":
            // Handle like message
            MakeLikeDeslike(socketManager, message)
        default:
            log.Printf("Unknown message type: %s", connectionType.Type)
        }
    }
    // Ensure the connection is closed when the loop exits
    RemoveConnection(socketManager, userID)
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

	sm.Sockets[userID] = conn
	log.Printf("Added new connection for user ID %d", userID)
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
