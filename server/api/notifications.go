package api

import (
	 "database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
)

func GetNotifications(w http.ResponseWriter, r *http.Request) {
    userID, err := util.GetUserID(r, w)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    rows, err := sqlite.DB.Query(`
        SELECT 
            n.id, 
            n.to_user_id, 
            n.content, 
            n.from_user_id, 
            n.read, 
            n.created_at, 
            n.type, 
            n.group_id
        FROM notifications n
        WHERE n.to_user_id = ?
        ORDER BY 
            n.created_at DESC,
            CASE WHEN n.type = 'follow_request' THEN 0 ELSE 1 END
        LIMIT 50;
    `, userID)
    if err != nil {
        log.Printf("Database error: %v", err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var notifications []models.Notification
    for rows.Next() {
        var notification models.Notification
        var fromUserID sql.NullInt64 // Handle potential NULL values
        var groupID sql.NullInt64   // Handle potential NULL values

        err := rows.Scan(
            &notification.ID,        // n.id
            &notification.ToUserID,  // n.to_user_id
            &notification.Content,   // n.content
            &fromUserID,             // n.from_user_id
            &notification.Read,      // n.read
            &notification.CreatedAt, // n.created_at
            &notification.Type,      // n.type
            &groupID,                 // n.group_id
        )
        if err != nil {
            log.Printf("Error scanning notification: %v", err)
            continue
        }

        // Set FromUserID and GroupID based on NULL checks
        if fromUserID.Valid {
            notification.FromUserID = int(fromUserID.Int64)
        } else {
            notification.FromUserID = 0 // Or handle it as needed
        }

        if groupID.Valid {
            notification.GroupID = int(groupID.Int64)
        } else {
            notification.GroupID = 0 // Or handle it as needed
        }

        notifications = append(notifications, notification)
    }

    if err := rows.Err(); err != nil {
        log.Printf("Error iterating through rows: %v", err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
fmt.Println(notifications,"notifications")
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