package api

import (
	 "database/sql"
	"encoding/json"
	//"fmt"
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
            n.group_id,
            CASE 
                WHEN n.type = 'follow_request' THEN EXISTS(
                    SELECT 1 FROM followers 
                    WHERE follower_id = n.from_user_id 
                    AND followed_id = n.to_user_id 
                    AND status = 'pending'
                )
                ELSE false
            END as has_pending_request
        FROM notifications n
        WHERE n.to_user_id = ?
        AND (
            n.type != 'follow_request' 
            OR (
                n.type = 'follow_request' 
                AND EXISTS(
                    SELECT 1 FROM followers 
                    WHERE follower_id = n.from_user_id 
                    AND followed_id = n.to_user_id 
                    AND status = 'pending'
                )
            )
        )
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
        var fromUserID sql.NullInt64
        var groupID sql.NullInt64
        var hasPendingRequest bool

        err := rows.Scan(
            &notification.ID,
            &notification.ToUserID,
            &notification.Content,
            &fromUserID,
            &notification.Read,
            &notification.CreatedAt,
            &notification.Type,
            &groupID,
            &hasPendingRequest,
        )
        if err != nil {
            log.Printf("Error scanning notification: %v", err)
            continue
        }

        // Only include follow request notifications if they have a pending request
        if notification.Type == models.NotificationTypeFollow && !hasPendingRequest {
            continue
        }

        if fromUserID.Valid {
            notification.FromUserID = int(fromUserID.Int64)
        }
        if groupID.Valid {
            notification.GroupID = int(groupID.Int64)
        }

        notifications = append(notifications, notification)
    }

    if err := rows.Err(); err != nil {
        log.Printf("Error iterating through rows: %v", err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
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