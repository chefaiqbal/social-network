package api

import (
    "log"
    "sync"
    "time"
    "net/http"
    "github.com/gorilla/websocket"
)

var (
    likeClients = make(map[*websocket.Conn]bool)
    likeMutex   sync.Mutex
)

type LikeUpdate struct {
    PostID    int  `json:"post_id"`
    LikeCount int  `json:"like_count"`
    UserLiked bool `json:"user_liked"`
    UserID    int  `json:"user_id"`
}

func LikeWebSocketHandler(w http.ResponseWriter, r *http.Request) {
    // Use the shared upgrader from socket.go
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("Error upgrading to WebSocket: %v", err)
        return
    }

    // Register new client
    likeMutex.Lock()
    likeClients[ws] = true
    likeMutex.Unlock()

    log.Printf("New WebSocket client connected. Total clients: %d", len(likeClients))

    // Cleanup on disconnect
    defer func() {
        likeMutex.Lock()
        delete(likeClients, ws)
        likeMutex.Unlock()
        ws.Close()
        log.Printf("WebSocket client disconnected. Remaining clients: %d", len(likeClients))
    }()

    // Keep the connection alive and handle incoming messages
    for {
        messageType, _, err := ws.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("WebSocket error: %v", err)
            }
            break
        }

        // Echo back ping messages
        if messageType == websocket.PingMessage {
            deadline := time.Now().Add(10 * time.Second)
            if err := ws.WriteControl(websocket.PongMessage, nil, deadline); err != nil {
                log.Printf("Error sending pong: %v", err)
                break
            }
        }
    }
}

func broadcastLikeUpdate(update LikeUpdate) {
    likeMutex.Lock()
    defer likeMutex.Unlock()

    log.Printf("Broadcasting like update to %d clients", len(likeClients))
    for client := range likeClients {
        err := client.WriteJSON(update)
        if err != nil {
            log.Printf("Error broadcasting like update: %v", err)
            client.Close()
            delete(likeClients, client)
            continue
        }
        log.Printf("Successfully sent update to a client")
    }
} 