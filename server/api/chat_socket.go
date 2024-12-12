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
    MessageTypeChat       = "chat"
    MessageTypeGroupChat  = "groupChat"
    MessageTypeTyping     = "typing"
    MessageTypeUserStatus = "user_status" 
)


type GroupChatMessage struct {
    Type    string `json:"type"`
    Content struct {
        GroupID int    `json:"group_id"`
        Message string `json:"message"`
    } `json:"content"`
}


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

    
    AddChatConnection(chatSocketManager, userID, conn)

   
    BroadcastUserStatus(chatSocketManager, userID, true)


    go func() {
        defer func() {
            // Remove the connection from the chat socket manager
            RemoveChatConnection(chatSocketManager, userID)

            // Broadcast that the user is offline
            BroadcastUserStatus(chatSocketManager, userID, false)

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
        Type        string          `json:"type"`
        RecipientID int64          `json:"recipient_id,omitempty"`
        Content     json.RawMessage `json:"content"`
    }

    if err := json.Unmarshal(msg, &message); err != nil {
  
        return
    }


    switch message.Type {
    case "ping":
        // Just send a pong back
        conn.WriteJSON(struct {
            Type string `json:"type"`
        }{
            Type: "pong",
        })
        return
    case MessageTypeChat:
        var content string
        if err := json.Unmarshal(message.Content, &content); err != nil {

            return
        }
        handleDirectMessage(conn, userID, message.RecipientID, content)
    case MessageTypeGroupChat:
        var groupMsg GroupChatMessage
        if err := json.Unmarshal(msg, &groupMsg); err != nil {
 
            return
        }
        log.Printf("Handling group message: %+v", groupMsg)
        handleGroupMessage(conn, userID, groupMsg)
    case MessageTypeTyping:
        var isTyping bool
        if err := json.Unmarshal(message.Content, &isTyping); err != nil {

            return
        }
        handleTypingStatus(userID, message.RecipientID, isTyping)
    }
}

func handleDirectMessage(conn *websocket.Conn, senderID uint64, recipientID int64, content string) {
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

func handleGroupMessage(conn *websocket.Conn, senderID uint64, msg GroupChatMessage) {
    // Verify sender is a member of the group
    var isMember bool
    err := sqlite.DB.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM group_members 
            WHERE group_id = ? AND user_id = ? 
            AND status IN ('member', 'creator')
        )`, msg.Content.GroupID, senderID).Scan(&isMember)

    if err != nil || !isMember {
        
        return
    }

    // Get sender's username
    var username string
    err = sqlite.DB.QueryRow("SELECT username FROM users WHERE id = ?", senderID).Scan(&username)
    if err != nil {
        log.Printf("Error getting username: %v", err)
        return
    }

    now := time.Now()

    // Save message to database
    result, err := sqlite.DB.Exec(`
        INSERT INTO group_chat_messages (group_id, sender_id, content, created_at)
        VALUES (?, ?, ?, ?)
    `, msg.Content.GroupID, senderID, msg.Content.Message, now)
    if err != nil {
        log.Printf("Error saving group message: %v", err)
        return
    }

    msgID, err := result.LastInsertId()
    if err != nil {
        log.Printf("Error getting message ID: %v", err)
        return
    }

    // Prepare response
    response := struct {
        Type      string    `json:"type"`
        ID        int64     `json:"id"`
        Content   string    `json:"content"`
        GroupID   int       `json:"group_id"`
        SenderID  uint64    `json:"sender_id"`
        Username  string    `json:"username"`
        CreatedAt time.Time `json:"created_at"`
    }{
        Type:      MessageTypeGroupChat,
        ID:        msgID,
        Content:   msg.Content.Message,
        GroupID:   msg.Content.GroupID,
        SenderID:  senderID,
        Username:  username,
        CreatedAt: now,
    }



    // Get all group members
    rows, err := sqlite.DB.Query(`
        SELECT user_id FROM group_members 
        WHERE group_id = ? AND status IN ('member', 'creator')
    `, msg.Content.GroupID)
    if err != nil {
        log.Printf("Error getting group members: %v", err)
        return
    }
    defer rows.Close()

    // Send message to all online group members
    for rows.Next() {
        var memberID uint64
        if err := rows.Scan(&memberID); err != nil {
            continue
        }

        if memberConn, ok := chatSocketManager.Sockets[memberID]; ok {
            if err := memberConn.WriteJSON(response); err != nil {
                log.Printf("Error sending message to member %d: %v", memberID, err)
                RemoveChatConnection(chatSocketManager, memberID)
            }
        }
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

func BroadcastUserStatus(sm *ChatSocketManager, userID uint64, isOnline bool) {
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

    sm.Mu.RLock()
    defer sm.Mu.RUnlock()

    for id, conn := range sm.Sockets {
        if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
            log.Printf("Error sending status to user %d: %v", id, err)
          
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

}

func RemoveChatConnection(sm *ChatSocketManager, userID uint64) {
    sm.Mu.Lock()
    defer sm.Mu.Unlock()

    if conn, exists := sm.Sockets[userID]; exists {
        delete(sm.Sockets, userID)
        conn.Close()

    }
}

func GetGroupChatMessages(w http.ResponseWriter, r *http.Request) {
    groupID := r.URL.Query().Get("groupId")
    if groupID == "" {
        http.Error(w, "Missing group ID", http.StatusBadRequest)
        return
    }

    // Get messages from database
    rows, err := sqlite.DB.Query(`
        SELECT gcm.id, gcm.sender_id, gcm.content, gcm.created_at, u.username
        FROM group_chat_messages gcm
        JOIN users u ON gcm.sender_id = u.id
        WHERE gcm.group_id = ?
        ORDER BY gcm.created_at DESC
        LIMIT 50
    `, groupID)
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var messages []struct {
        ID        int64     `json:"id"`
        SenderID  int64     `json:"sender_id"`
        Content   string    `json:"content"`
        CreatedAt time.Time `json:"created_at"`
        Username  string    `json:"username"`
    }

    for rows.Next() {
        var msg struct {
            ID        int64     `json:"id"`
            SenderID  int64     `json:"sender_id"`
            Content   string    `json:"content"`
            CreatedAt time.Time `json:"created_at"`
            Username  string    `json:"username"`
        }
        if err := rows.Scan(&msg.ID, &msg.SenderID, &msg.Content, &msg.CreatedAt, &msg.Username); err != nil {
            continue
        }
        messages = append(messages, msg)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(messages)
}