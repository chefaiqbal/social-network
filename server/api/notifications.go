package api

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "social-network/pkg/db/sqlite"
    "social-network/util"
    "social-network/models"
    "log"
)

func GetNotifications(w http.ResponseWriter, r *http.Request) {
    userID, err := util.GetUserID(r, w)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    rows, err := sqlite.DB.Query(`
        SELECT n.id, n.to_user_id, n.content, n.from_user_id, n.read, 
               n.created_at, n.type, u.username, u.avatar
        FROM notifications n
        JOIN users u ON u.id = n.from_user_id
        WHERE n.to_user_id = ?
        ORDER BY n.created_at DESC, 
                 CASE WHEN n.type = 'follow_request' THEN 0 ELSE 1 END
        LIMIT 50
    `, userID)
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    type NotificationResponse struct {
        models.Notification
        FromUsername string `json:"from_username"`
        FromAvatar   string `json:"from_avatar"`
    }

    var notifications []NotificationResponse
    for rows.Next() {
        var n NotificationResponse
        var avatar sql.NullString
        err := rows.Scan(&n.ID, &n.ToUserID, &n.Content, &n.FromUserID, 
            &n.Read, &n.CreatedAt, &n.Type, &n.FromUsername, &avatar)
        if err != nil {
            log.Printf("Error scanning notification: %v", err)
            continue
        }
        if avatar.Valid {
            n.FromAvatar = avatar.String
        }
        notifications = append(notifications, n)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(notifications)
}

func MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
    userID, err := util.GetUserID(r, w)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    notificationID := r.PathValue("id")
    
    _, err = sqlite.DB.Exec(`
        UPDATE notifications 
        SET read = true 
        WHERE id = ? AND to_user_id = ?
    `, notificationID, userID)
    
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
} 