package middleware

import (
    "net/http"
    //"database/sql"
    "strconv"
    "social-network/pkg/db/sqlite"
    "social-network/util"
)

// GroupAuthMiddleware checks if the user is a member of the group before allowing access
func GroupAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Get user ID from session
        userID, err := util.GetUserID(r, w)
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        // Get group ID from URL path or request body
        var groupID int64
        if id := r.PathValue("id"); id != "" {
            groupID, err = strconv.ParseInt(id, 10, 64)
        } else if id := r.PathValue("groupId"); id != "" {
            groupID, err = strconv.ParseInt(id, 10, 64)
        }

        if err != nil {
            http.Error(w, "Invalid group ID", http.StatusBadRequest)
            return
        }

        // Check if user is a member
        var isMember bool
        query := `SELECT EXISTS (
            SELECT 1 FROM group_members 
            WHERE group_id = ? AND user_id = ? 
            AND (status = 'member' OR status = 'creator')
        )`
        
        err = sqlite.DB.QueryRow(query, groupID, userID).Scan(&isMember)
        if err != nil {
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }

        if !isMember {
            http.Error(w, "Access denied: Not a group member", http.StatusForbidden)
            return
        }

        // If user is a member, proceed to the next handler
        next(w, r)
    }
} 