package api

import (
    "encoding/json"
    "log"
    "net/http"
    "time"
    "social-network/pkg/db/sqlite"
    "social-network/util"
    "github.com/gorilla/websocket"
    "database/sql"
)

// GroupChatHandler handles WebSocket connections for group chat
func GroupChatHandler(w http.ResponseWriter, r *http.Request) {
    userID, err := util.GetUserID(r, w)
    if err != nil {
        log.Printf("Group Chat WebSocket auth error: %v", err)
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    conn, err := chatUpgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("Group Chat WebSocket upgrade error: %v", err)
        return
    }

    log.Printf("New group chat WebSocket connection for user %d", userID)

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
                    log.Printf("Group Chat WebSocket error: %v", err)
                }
                break
            }
            HandleGroupChatMessages(conn, userID, msg)
        }
    }()
}

// HandleGroupChatMessages specifically handles group chat messages
func HandleGroupChatMessages(conn *websocket.Conn, userID uint64, msg []byte) {
    log.Printf("Received group message: %s", string(msg))
    
    var groupMsg GroupChatMessage
    if err := json.Unmarshal(msg, &groupMsg); err != nil {
        log.Printf("Error unmarshalling group message: %v", err)
        return
    }

    // Debug log to check the values
    log.Printf("Checking membership for user %d in group %d", userID, groupMsg.Content.GroupID)

    // Modified query to check for both member and creator status
    var status string
    err := sqlite.DB.QueryRow(`
        SELECT status 
        FROM group_members 
        WHERE group_id = ? 
        AND user_id = ?
    `, groupMsg.Content.GroupID, userID).Scan(&status)

    if err != nil {
        if err == sql.ErrNoRows {
            log.Printf("User %d is not a member of group %d", userID, groupMsg.Content.GroupID)
            return
        }
        log.Printf("Database error checking membership: %v", err)
        return
    }

    log.Printf("User %d status in group %d: %s", userID, groupMsg.Content.GroupID, status)

    // Check if user is either a member or creator
    if status != "member" && status != "creator" {
        log.Printf("User %d is not an active member of group %d (status: %s)", userID, groupMsg.Content.GroupID, status)
        return
    }

    // Get sender's username
    var username string
    err = sqlite.DB.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
    if err != nil {
        log.Printf("Error getting username: %v", err)
        return
    }

    now := time.Now()

    // Save message to database
    result, err := sqlite.DB.Exec(`
        INSERT INTO group_chat_messages (group_id, sender_id, content, created_at)
        VALUES (?, ?, ?, ?)
    `, groupMsg.Content.GroupID, userID, groupMsg.Content.Message, now)
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
        Content:   groupMsg.Content.Message,
        GroupID:   groupMsg.Content.GroupID,
        SenderID:  userID,
        Username:  username,
        CreatedAt: now,
    }

    log.Printf("Broadcasting group message: %+v", response)

    // Get all group members
    rows, err := sqlite.DB.Query(`
        SELECT user_id 
        FROM group_members 
        WHERE group_id = ? 
        AND (status = 'member' OR status = 'creator')
    `, groupMsg.Content.GroupID)
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