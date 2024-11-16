package api

import (
    "encoding/json"
    "log"
    "net/http"
    "sync"
    "time"
    "social-network/pkg/db/sqlite"
    "social-network/util"

    "github.com/gorilla/websocket"
)

// Message types
const (
    MessageTypeChat   = "chat"
    MessageTypeTyping = "typing"
)

// Create a separate socket manager for chat
var chatSocketManager = makeChatSocketManager()

var chatUpgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
        return origin == "http://localhost:3000"
    },
}

type ChatSocketManager struct {
    Sockets map[uint64]*websocket.Conn
    Mu      sync.RWMutex
}

func makeChatSocketManager() *ChatSocketManager {
    return &ChatSocketManager{
        Sockets: make(map[uint64]*websocket.Conn),
    }
}

func ChatWebSocketHandler(w http.ResponseWriter, r *http.Request) {
    userID, err := util.GetUserID(r, w)
    if err != nil {
        log.Printf("Chat WebSocket auth error: %v", err)
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    conn, err := chatUpgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("Chat WebSocket upgrade error: %v", err)
        return
    }

    log.Printf("New chat WebSocket connection for user %d", userID)

    // Add the connection to the chat socket manager
    AddChatConnection(chatSocketManager, userID, conn)

    // Handle messages in a goroutine
    go func() {
        defer func() {
            RemoveChatConnection(chatSocketManager, userID)
            conn.Close()
        }()

        for {
            _, msg, err := conn.ReadMessage()
            if err != nil {
                if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                    log.Printf("Chat WebSocket error: %v", err)
                }
                break
            }
            HandleChatMessages(conn, userID, msg)
        }
    }()
}

func HandleChatMessages(conn *websocket.Conn, userID uint64, msg []byte) {
    var message struct {
        Type        string `json:"type"`
        RecipientID int64  `json:"recipient_id"`
        Content     string `json:"content"`
    }

    if err := json.Unmarshal(msg, &message); err != nil {
        log.Printf("Error unmarshalling chat message: %v", err)
        return
    }

    switch message.Type {
    case MessageTypeChat:
        handleChatMessage(conn, userID, message.RecipientID, message.Content)
    case MessageTypeTyping:
        handleTypingStatus(userID, message.RecipientID, message.Content == "true")
    }
}

func handleChatMessage(conn *websocket.Conn, senderID uint64, recipientID int64, content string) {
    if content == "" {
        return
    }

    now := time.Now()

    // Save message to database
    result, err := sqlite.DB.Exec(`
        INSERT INTO chat_messages (sender_id, recipient_id, content, created_at)
        VALUES (?, ?, ?, ?)
    `, senderID, recipientID, content, now)
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
        Type:        MessageTypeChat,
        ID:          msgID,
        SenderID:    int64(senderID),
        RecipientID: recipientID,
        Content:     content,
        CreatedAt:   now,
    }

    // Send to recipient if online
    if recipientConn, ok := chatSocketManager.Sockets[uint64(recipientID)]; ok {
        if err := recipientConn.WriteJSON(response); err != nil {
            log.Printf("Error sending message to recipient: %v", err)
            RemoveChatConnection(chatSocketManager, uint64(recipientID))
        }
    }

    // Send confirmation back to sender
    if err := conn.WriteJSON(response); err != nil {
        log.Printf("Error sending confirmation to sender: %v", err)
    }
}

func handleTypingStatus(senderID uint64, recipientID int64, isTyping bool) {
    response := struct {
        Type        string `json:"type"`
        SenderID    int64  `json:"sender_id"`
        RecipientID int64  `json:"recipient_id"`
        IsTyping    bool   `json:"is_typing"`
    }{
        Type:        MessageTypeTyping,
        SenderID:    int64(senderID),
        RecipientID: recipientID,
        IsTyping:    isTyping,
    }

    // Send typing status to recipient if online
    if recipientConn, ok := chatSocketManager.Sockets[uint64(recipientID)]; ok {
        if err := recipientConn.WriteJSON(response); err != nil {
            log.Printf("Error sending typing status: %v", err)
            RemoveChatConnection(chatSocketManager, uint64(recipientID))
        }
    }
}

func AddChatConnection(sm *ChatSocketManager, userID uint64, conn *websocket.Conn) {
    sm.Mu.Lock()
    defer sm.Mu.Unlock()

    if existingConn, exists := sm.Sockets[userID]; exists {
        delete(sm.Sockets, userID)
        existingConn.Close()
        log.Printf("Closed existing chat connection for user %d", userID)
    }

    sm.Sockets[userID] = conn
    log.Printf("Added new chat connection for user %d", userID)
}

func RemoveChatConnection(sm *ChatSocketManager, userID uint64) {
    sm.Mu.Lock()
    defer sm.Mu.Unlock()

    if conn, exists := sm.Sockets[userID]; exists {
        delete(sm.Sockets, userID)
        conn.Close()
        log.Printf("Removed chat connection for user %d", userID)
    }
} 