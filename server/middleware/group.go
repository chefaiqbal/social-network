package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"social-network/pkg/db/sqlite"
	"social-network/util"
	"strconv"
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

		// First try to get from URL parameters
		if id := r.PathValue("id"); id != "" {
			groupID, err = strconv.ParseInt(id, 10, 64)
		} else if id := r.PathValue("groupId"); id != "" {
			groupID, err = strconv.ParseInt(id, 10, 64)
		} else if r.Method == "POST" {
			// If not in URL, try to get from request body for POST requests
			var bodyBytes []byte
			if r.Body != nil {
				bodyBytes, _ = io.ReadAll(r.Body)
				// Restore the body for subsequent handlers
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			if len(bodyBytes) > 0 {
				var requestData struct {
					GroupID int64 `json:"group_id"`
				}
				if err := json.Unmarshal(bodyBytes, &requestData); err == nil {
					groupID = requestData.GroupID
				}
			}
		}

		if groupID == 0 {
			http.Error(w, "Group ID not found", http.StatusBadRequest)
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
			log.Printf("Database error checking membership: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		if !isMember {
			log.Printf("Access denied: User %d is not a member of group %d", userID, groupID)
			http.Error(w, "Access denied: Not a group member", http.StatusForbidden)
			return
		}

		// If user is a member, proceed to the next handler
		next(w, r)
	}
}
