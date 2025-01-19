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

// validation function
func validateGroupAccess(userID uint64, groupID int) bool {
    var status string
    err := sqlite.DB.QueryRow(`
        SELECT status FROM group_members 
        WHERE group_id = ? AND user_id = ? 
        AND status IN ('member', 'creator')
    `, groupID, userID).Scan(&status)
    return err == nil
}
// GroupChatHandler handles WebSocket connections for group chat
func GroupChatHandler(w http.ResponseWriter, r *http.Request) {
    userID, err := util.GetUserID(r, w)
    if err != nil {

        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    conn, err := chatUpgrader.Upgrade(w, r, nil)
    if err != nil {

        return
    }


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

                }
                break
            }
            HandleGroupChatMessages(conn, userID, msg)
        }
    }()
}

// HandleGroupChatMessages specifically handles group chat messages
func HandleGroupChatMessages(conn *websocket.Conn, userID uint64, msg []byte) {
    var groupMsg GroupChatMessage
    if err := json.Unmarshal(msg, &groupMsg); err != nil {
        return
    }

    // Validate group access
    if !validateGroupAccess(userID, groupMsg.Content.GroupID) {
        conn.WriteJSON(struct {
            Type    string `json:"type"`
            Error   string `json:"error"`
        }{
            Type:  "error",
            Error: "Unauthorized group access",
        })
        return
    }

    var status string
    err := sqlite.DB.QueryRow(`
        SELECT status 
        FROM group_members 
        WHERE group_id = ? 
        AND user_id = ?
    `, groupMsg.Content.GroupID, userID).Scan(&status)

    if err != nil {
        if err == sql.ErrNoRows {
            return
        }
        log.Printf("Database error checking membership: %v", err)
        return
    }

    log.Printf("User %d status in group %d: %s", userID, groupMsg.Content.GroupID, status)

    if status != "member" && status != "creator" {
        return
    }

    var username string
    err = sqlite.DB.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
    if err != nil {
        return
    }

    now := time.Now()

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
