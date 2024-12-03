package api

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	m "social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
	"strconv"
	"strings"
	"time"
	// "golang.org/x/mod/module"
)

func CreateGroup(w http.ResponseWriter, r *http.Request) {
	var group m.Group

	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	userID, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if strings.TrimSpace(group.Description) == "" || strings.TrimSpace(group.Title) == "" {
		http.Error(w, "Please Provide all fields", http.StatusBadRequest)
		return
	}

	if exist, err := m.DoesUserExist(uint(userID), sqlite.DB); !exist {
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			log.Printf("error: %v", err)
			return
		}
		http.Error(w, "User does not exist", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO groups (title, description, creator_id, created_at) 
              VALUES (?, ?, ?, ?)`
	result, err := sqlite.DB.Exec(query, group.Title, group.Description, userID, time.Now())
	if err != nil {
		http.Error(w, "Error creating group", http.StatusInternalServerError)
		log.Printf("error: %v", err)
		return
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Error retrieving group ID", http.StatusInternalServerError)
		log.Printf("error: %v", err)
		return
	}

	memberQuery := `INSERT INTO group_members (group_id, user_id, status, created_at) 
                    VALUES (?, ?, ?, ?)`
	_, err = sqlite.DB.Exec(memberQuery, groupID, userID, "creator", time.Now())
	if err != nil {
		http.Error(w, "Error adding creator as group member", http.StatusInternalServerError)
		log.Printf("error: %v", err)
		return
	}

	group.ID = uint(groupID)
	group.CreatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(group); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		log.Printf("error: %v", err)
	}
}

func CreateGroupPost(w http.ResponseWriter, r *http.Request) {
    var postInput struct {
        Title   string `json:"title"`
        Content string `json:"content"`
        Media   string `json:"media"`      // Base64 string from frontend
        Privacy int    `json:"privacy"`
        GroupID *int64 `json:"group_id"`   // Group ID passed in the body
    }

    if err := json.NewDecoder(r.Body).Decode(&postInput); err != nil {
        http.Error(w, "Error reading data", http.StatusBadRequest)
        return
    }

    // Get the current user's ID
    userID, err := util.GetUserID(r, w)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

	//  validate if the user is a member of the group
query := `SELECT 1 FROM group_members WHERE group_id = ? AND user_id = ? AND (status = 'member' OR status = 'creator')`
	var exists bool
	err = sqlite.DB.QueryRow(query, postInput.GroupID, userID).Scan(&exists)
	if err == sql.ErrNoRows {
		http.Error(w, "You are not a member of this group. Join the group to post.", http.StatusForbidden)
		return
	} else if err != nil {
		http.Error(w, "An error occurred while verifying membership", http.StatusInternalServerError)
		log.Printf("Membership query error: %v", err)
		return
	}	

    // Validate privacy value
    if postInput.Privacy != 1 && postInput.Privacy != 2 && postInput.Privacy != 3 {
        http.Error(w, "Invalid privacy type", http.StatusBadRequest)
        return
    }

    var mediaBytes []byte
    var mediaType string

    // Process media if provided
    if postInput.Media != "" {
        // Split the base64 string to get the media type
        parts := strings.Split(postInput.Media, ";base64,")
        if len(parts) != 2 {
            http.Error(w, "Invalid media format", http.StatusBadRequest)
            return
        }

        mediaType = strings.TrimPrefix(parts[0], "data:")
        mediaBytes, err = base64.StdEncoding.DecodeString(parts[1])
        if err != nil {
            http.Error(w, "Invalid media encoding", http.StatusBadRequest)
            return
        }
    }

    // Insert the post into the group_posts table instead
    result, err := sqlite.DB.Exec(
        `INSERT INTO group_posts (title, content, media, media_type, author, group_id, created_at) 
         VALUES (?, ?, ?, ?, ?, ?, ?)`,
        postInput.Title,
        postInput.Content,
        mediaBytes,
        mediaType,
        userID,
        postInput.GroupID,
        time.Now(),
    )
    if err != nil {
        http.Error(w, "Failed to create post", http.StatusInternalServerError)
        log.Printf("create group post error: %v", err)
        return
    }

    postID, _ := result.LastInsertId()

    // Return the created post
    response := m.PostResponse{
        ID:        postID,
        Title:     postInput.Title,
        Content:   postInput.Content,
        MediaBase64: postInput.Media,
        MediaType: mediaType,
        Privacy:   postInput.Privacy,
        Author:    int64(userID),
        GroupID:   postInput.GroupID,
        CreatedAt: time.Now(),
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(response)
}

	func GetGroupPost(w http.ResponseWriter, r *http.Request) {
		var groupPosts []m.Post
		groupIDString := r.PathValue("id")

		// convert the string into a number
		if groupIDString == "" {
			http.Error(w, "group id is null", http.StatusBadRequest)
			return
		}

		groupID, err := strconv.Atoi(groupIDString)
		if err != nil {
			http.Error(w, "Invalid number", http.StatusBadRequest)
			return
		}

		// the value of the group id can't be less than 1
		if groupID < 1 {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		rows, err := sqlite.DB.Query(`
			SELECT gp.id, gp.title, gp.content, 
				   COALESCE(gp.media, '') AS media,
				   COALESCE(gp.media_type, '') AS media_type,
				   gp.author, gp.created_at, gp.group_id,
				   u.username as author_name
			FROM group_posts gp
			LEFT JOIN users u ON gp.author = u.id
			WHERE gp.group_id = ? 
			ORDER BY gp.created_at DESC`,
			groupID,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Group does not exist", http.StatusBadRequest)
				return
			}
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			log.Printf("Error: %v", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var post m.Post
			var mediaBytes []byte
			if err := rows.Scan(
				&post.ID,
				&post.Title,
				&post.Content,
				&mediaBytes,
				&post.MediaType,
				&post.Author,
				&post.CreatedAt,
				&post.GroupID,
				&post.AuthorName,
			); err != nil {
				http.Error(w, "Error getting post", http.StatusInternalServerError)
				log.Printf("Error scanning: %v", err)
				return
			}

			if len(mediaBytes) > 0 {
				post.Media = base64.StdEncoding.EncodeToString(mediaBytes)
			}

			groupPosts = append(groupPosts, post)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(groupPosts); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			log.Printf("Error encoding: %v", err)
		}
	}

func VeiwGorups(w http.ResponseWriter, r *http.Request) {
	var groups []m.Group

	userID, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := `SELECT g.* FROM groups g
	LEFT JOIN group_members gm
	ON g.id = gm.group_id AND gm.user_id = ?
	WHERE gm.id IS NULL OR gm.status = ?;`

	rows, err := sqlite.DB.Query(query, userID, "pending")
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}

	for rows.Next() {
		var group m.Group
		if err := rows.Scan(&group.ID, &group.Title, &group.Description, &group.CreatorID, &group.CreatedAt); err != nil {
			http.Error(w, "Error getting group", http.StatusInternalServerError)
			log.Printf("Error: %v", err)
			return
		}

		groups = append(groups, group)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&groups); err != nil {
		http.Error(w, "Error sending json", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}
}

func GroupInvitation(w http.ResponseWriter, r *http.Request) {
	var inviteRequest struct {
		GroupID int `json:"groupId"`
	}

	// Get the user ID from the request context
	userID, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode the JSON request body
	if err := json.NewDecoder(r.Body).Decode(&inviteRequest); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Check if the user is already a member of the group
	var existingStatus string
	err = sqlite.DB.QueryRow(
		"SELECT status FROM group_members WHERE group_id = ? AND user_id = ?",
		inviteRequest.GroupID, userID,
	).Scan(&existingStatus)

	if err != nil && err != sql.ErrNoRows {
		// Handle database error
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Error checking existing membership: %v", err)
		return
	}

	// If the user is already in the group
	if err == nil {
		// Customize the response based on the current status
		message := "You are already part of this group."
		if existingStatus == "pending" {
			message = "Your invitation is still pending."
		}
		w.WriteHeader(http.StatusConflict) // 409 Conflict
		json.NewEncoder(w).Encode(map[string]string{"message": message})
		return
	}

	// Insert the new member record
	_, err = sqlite.DB.Exec(
		"INSERT INTO group_members (group_id, user_id, status) VALUES (?, ?, ?)",
		inviteRequest.GroupID, userID, "pending",
	)

	if err != nil {
		http.Error(w, "Failed to join the group", http.StatusInternalServerError)
		log.Printf("Error inserting group member: %v", err)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Group invitation sent successfully"})
}



func GroupJoinRequest(w http.ResponseWriter, r *http.Request) {
	var inviteRequest struct {
		GroupID int `json:"groupId"`
	}

	// Get the user ID from the request context
	userID, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode the JSON request body
	if err := json.NewDecoder(r.Body).Decode(&inviteRequest); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Check if the user is already a member of the group
	var existingStatus string
	err = sqlite.DB.QueryRow(
		"SELECT status FROM group_members WHERE group_id = ? AND user_id = ?",
		inviteRequest.GroupID, userID,
	).Scan(&existingStatus)

	if err != nil && err != sql.ErrNoRows {
		// Handle database error
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Error checking existing membership: %v", err)
		return
	}

	// If the user is already in the group
	if err == nil {
		// Customize the response based on the current status
		message := "You are already part of this group."
		if existingStatus == "pending" {
			message = "Your invitation is still pending."
		}
		w.WriteHeader(http.StatusConflict) // 409 Conflict
		json.NewEncoder(w).Encode(map[string]string{"message": message})
		return
	}

	// Insert the new member record
	_, err = sqlite.DB.Exec(
		"INSERT INTO group_members (group_id, user_id, status) VALUES (?, ?, ?)",
		inviteRequest.GroupID, userID, "pending",
	)

	if err != nil {
		http.Error(w, "Failed to join the group", http.StatusInternalServerError)
		log.Printf("Error inserting group member: %v", err)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Group invitation sent successfully"})
}


func GroupAccept(w http.ResponseWriter, r *http.Request) {
	var inviteRequest m.GroupJoinRequest

	if err := json.NewDecoder(r.Body).Decode(&inviteRequest); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Validate input
	if inviteRequest.GroupID <= 0 || inviteRequest.UserID <= 0 {
		http.Error(w, "Invalid group ID or receiver ID", http.StatusBadRequest)
		return
	}

	// Check if the invitation exists and is pending
	var status string
	err := sqlite.DB.QueryRow("SELECT status FROM group_members WHERE group_id = ? AND user_id = ?", inviteRequest.GroupID, inviteRequest.UserID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invitation not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error checking invitation status", http.StatusInternalServerError)
			log.Printf("Error: %v", err)
		}
		return
	}

	if status != "pending" {
		http.Error(w, "Invitation is not pending", http.StatusBadRequest)
		return
	}

	// Update the status to "accepted"
	result, err := sqlite.DB.Exec("UPDATE group_members SET status = 'member' WHERE group_id = ? AND user_id = ?", inviteRequest.GroupID, inviteRequest.UserID)
	if err != nil {
		http.Error(w, "Error updating invitation status", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "No invitation updated", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Group invitation accepted successfully"})
}

func GroupReject(w http.ResponseWriter, r *http.Request) {
	var inviteRequest m.GroupJoinRequest

	if err := json.NewDecoder(r.Body).Decode(&inviteRequest); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Validate input
	if inviteRequest.GroupID <= 0 || inviteRequest.UserID <= 0 {
		http.Error(w, "Invalid group ID or receiver ID", http.StatusBadRequest)
		return
	}

	// Check if the invitation exists
	var status string
	err := sqlite.DB.QueryRow("SELECT status FROM group_members WHERE group_id = ? AND user_id = ?", inviteRequest.GroupID, inviteRequest.UserID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invitation not found", http.StatusNotFound)
			return
		} else {
			http.Error(w, "Error checking invitation status", http.StatusInternalServerError)
			log.Printf("Error: %v", err)
			return
		}
	}

	if status != "pending" {
		http.Error(w, "Invitation is not pending", http.StatusBadRequest)
		return
	}

	// Delete the invitation
	result, err := sqlite.DB.Exec("DELETE FROM group_members WHERE group_id = ? AND user_id = ?", inviteRequest.GroupID, inviteRequest.UserID)
	if err != nil {
		http.Error(w, "Error deleting invitation", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error checking rows affected", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "No invitation deleted", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Group invitation rejected successfully"})
}

func GroupLeave(w http.ResponseWriter, r *http.Request) {
	// neded Get the group
	var Leave m.GroupLeave

	if err := json.NewDecoder(r.Body).Decode(&Leave); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var exists bool
	err := sqlite.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = ? AND user_id = ?)", Leave.GroupID, Leave.UserID).
		Scan(&exists)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "User is not a member of the group", http.StatusNotFound)
		return
	}

	// leave logic
	result, err := sqlite.DB.Exec("DELETE FROM group_members WHERE group_id = ? AND user_id = ?", Leave.GroupID, Leave.UserID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "User is not a member of the group", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User successfully removed from the group"})

}

// we need group delete

// get our group
func MyGroups(w http.ResponseWriter, r *http.Request) {
	// Get the user ID from the request
	userID, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Query to fetch group IDs
	groupIDsQuery := `SELECT group_id FROM group_members WHERE user_id = ? AND status != 'pending'`
	rows, err := sqlite.DB.Query(groupIDsQuery, userID)
	if err != nil {
		http.Error(w, "Database error while fetching group IDs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var groupIDs []string
	for rows.Next() {
		var groupID uint
		if err := rows.Scan(&groupID); err != nil {
			http.Error(w, "Database error while scanning group IDs", http.StatusInternalServerError)
			return
		}
		groupIDs = append(groupIDs, fmt.Sprint(groupID))
	}

	// If no groups are found, exit early
	if len(groupIDs) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Construct the query for fetching group details
	idList := strings.Join(groupIDs, ",")
	groupsQuery := fmt.Sprintf(`SELECT id, title, description, creator_id, created_at FROM groups WHERE id IN (%s)`, idList)

	rows, err = sqlite.DB.Query(groupsQuery)
	if err != nil {
		http.Error(w, "Database error while fetching group details", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var groups []m.Group
	for rows.Next() {
		var group m.Group
		if err := rows.Scan(&group.ID, &group.Title, &group.Description, &group.CreatorID, &group.CreatedAt); err != nil {
			http.Error(w, "Database error while scanning group details", http.StatusInternalServerError)
			return
		}
		groups = append(groups, group)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(groups); err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		return
	}
}

func Members(w http.ResponseWriter, r *http.Request) {
	GetMembers(w, r)
}

// to get all members for a praticular group
func GetMembers(w http.ResponseWriter, r *http.Request) {
	type GroupID struct {
		ID string `json:"group_id"`
	}

	var groupID GroupID

	if err := json.NewDecoder(r.Body).Decode(&groupID); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Printf("Error decoding request body: %v", err)
		return
	}

	groupIDInt, err := strconv.Atoi(groupID.ID)
	if err != nil {
		http.Error(w, "group_id must be a number", http.StatusBadRequest)
		log.Printf("Error converting group_id to int: %v", err)
		return
	}

	log.Printf("Group ID: %d", groupIDInt)

	query := `
		SELECT gm.user_id, gm.status, u.username
		FROM group_members gm
		INNER JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = ? AND (gm.status = ? OR gm.status = ?)
	`

	rows, err := sqlite.DB.Query(query, groupIDInt, "member", "creator")
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error querying database: %v", err)
		return
	}
	defer rows.Close()

	var members []m.GroupMemebers
	for rows.Next() {
		var member m.GroupMemebers
		if err := rows.Scan(&member.UserID, &member.Status, &member.Username); err != nil {
			http.Error(w, "Error getting group members", http.StatusInternalServerError)
			log.Printf("Error scanning row: %v", err)
			return
		}
		members = append(members, member)
	}

	log.Printf("Members: %+v", members)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(members); err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		log.Printf("Error encoding response: %v", err)
		return
	}
}



func GetPendingUsers(w http.ResponseWriter, r *http.Request) {
    type GroupID struct {
        ID int `json:"group_id"` // Change to int
    }

    var groupID GroupID

    if err := json.NewDecoder(r.Body).Decode(&groupID); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        log.Printf("Error decoding request body: %v", err)
        return
    }

    log.Printf("Group ID: %d", groupID.ID)

    query := `
        SELECT gm.user_id, gm.status, u.username
        FROM group_members gm
        INNER JOIN users u ON gm.user_id = u.id
        WHERE gm.group_id = ? AND gm.status = ?
    `

    rows, err := sqlite.DB.Query(query, groupID.ID, "pending")
    if err != nil {
        http.Error(w, "Something went wrong", http.StatusInternalServerError)
        log.Printf("Error querying database: %v", err)
        return
    }
    defer rows.Close()

    var members []m.GroupMemebers
    for rows.Next() {
        var member m.GroupMemebers
        if err := rows.Scan(&member.UserID, &member.Status, &member.Username); err != nil {
            http.Error(w, "Error getting group members", http.StatusInternalServerError)
            log.Printf("Error scanning row: %v", err)
            return
        }
        members = append(members, member)
    }

    log.Printf("Members: %+v", members)

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(members); err != nil {
        http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
        log.Printf("Error encoding response: %v", err)
        return
    }
}

func DelGroup(r *http.Request, w http.ResponseWriter) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	log.Println(id)

	query := `DELETE FROM groups 
	WHERE id = ?`

	query2 := `DELETE FROM group_members
	WHERE group_id = ?`

	_, err = sqlite.DB.Exec(query2, id)
	if err != nil {
		http.Error(w, "Error deleting group", http.StatusInternalServerError)
		log.Printf("Error deleting group: %v", err)
		return
	}

	_, err = sqlite.DB.Exec(query, id)
	if err != nil {
		http.Error(w, "Error deleting group", http.StatusInternalServerError)
		log.Printf("Error deleting group: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func CreateEvent(w http.ResponseWriter, r *http.Request) {
	var event m.GroupEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if group exists
	var exists bool
	err := sqlite.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM groups WHERE id = ?)", event.GroupID).Scan(&exists)
	if err != nil || !exists {

		http.Error(w, "Group does not exist", http.StatusBadRequest)
		return
	}
	// Assuming event.GroupID is an int
	groupTitle, err := getGroupTitleByID(uint(event.GroupID))
	if err != nil {
		fmt.Println("Error fetching group title:", err)
	} else {
		fmt.Println("Group Title:", groupTitle)
	}

	// Check if user is a member
	var isMember bool
	err = sqlite.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM group_members WHERE user_id = ? AND group_id = ?)", event.CreatorID, event.GroupID).Scan(&isMember)
	if err != nil || !isMember {
		http.Error(w, "User is not a member of the group", http.StatusForbidden)
		return
	}

	// Insert the event
	query := `INSERT INTO group_events (group_id, creator_id, title, description, event_date) 
              VALUES (?, ?, ?, ?, ?)`
	result, err:= sqlite.DB.Exec(query, event.GroupID, event.CreatorID, event.Title, event.Description, event.EventDate)
	if err != nil {
		http.Error(w, "Error creating event", http.StatusInternalServerError)
		return
	}
	// Retrieve the auto-generated ID
	eventID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Error retrieving event ID", http.StatusInternalServerError)
		return
	}
	event.ID = int(eventID) 

	// Batch notifications
	rows, err := sqlite.DB.Query("SELECT user_id FROM group_members WHERE group_id = ?", event.GroupID)
	if err != nil {
		log.Printf("Error fetching group members: %v", err)
		return
	}
	defer rows.Close()

	var notifications []m.Notification
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			log.Printf("Error scanning user_id: %v", err)
			continue
		}

		// Construct the message
		notifications = append(notifications, m.Notification{
			ToUserID:  userID,
			GroupID:   int(event.GroupID),
			Content:   fmt.Sprintf("%s invites you to join %s ! RSVP now to save your spot!", groupTitle, event.Title),
			Type:      m.NotificationEvent,
			CreatedAt: time.Now(),
			Read:      false,
		})
	}

	tx, err := sqlite.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return
	}
	stmt, err := tx.Prepare(`
        INSERT INTO notifications (to_user_id, group_id, content, type, read, created_at) 
        VALUES (?, ?, ?, ?, ?, ?)
    `)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		tx.Rollback()
		return
	}
	defer stmt.Close()

	for _, notification := range notifications {
		_, err := stmt.Exec(
			notification.ToUserID,
			notification.GroupID,
			notification.Content,
			notification.Type,
			notification.Read,
			notification.CreatedAt,
		)
		if err != nil {
			log.Printf("Error inserting notification: %v", err)
			tx.Rollback()
			return
		}
		BroadcastNotification(notification)

	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return
	}


	
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(event)
}

func RSVPEvent(w http.ResponseWriter, r *http.Request) {
	var rsvp m.GroupEventRSVP
	// Decode the incoming RSVP data from the request body
	if err := json.NewDecoder(r.Body).Decode(&rsvp); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set CreatedAt to the current time if it's not provided
	if rsvp.CreatedAt.IsZero() {
		rsvp.CreatedAt = time.Now()
	}

	// Check if an RSVP already exists for this user and event
	var existingRSVP int
	checkQuery := `SELECT id FROM group_event_RSVP WHERE event_id = ? AND user_id = ?`
	err := sqlite.DB.QueryRow(checkQuery, rsvp.EventID, rsvp.UserID).Scan(&existingRSVP)

	if err != nil && err != sql.ErrNoRows {
		// If the query failed for reasons other than "no rows found"
		http.Error(w, "Error checking RSVP", http.StatusInternalServerError)
		return
	}

	if existingRSVP > 0 {
		// If an RSVP exists, update it
		updateQuery := `UPDATE group_event_RSVP SET rsvp_status = ?, created_at = ? WHERE event_id = ? AND user_id = ?`
		_, err := sqlite.DB.Exec(updateQuery, rsvp.RSVPStatus, rsvp.CreatedAt, rsvp.EventID, rsvp.UserID)
		if err != nil {
			http.Error(w, "Error updating RSVP", http.StatusInternalServerError)
			return
		}
	} else {
		// If no RSVP exists, insert a new one
		insertQuery := `INSERT INTO group_event_RSVP (event_id, user_id, rsvp_status, created_at) VALUES (?, ?, ?, ?)`
		_, err := sqlite.DB.Exec(insertQuery, rsvp.EventID, rsvp.UserID, rsvp.RSVPStatus, rsvp.CreatedAt)
		if err != nil {
			http.Error(w, "Error adding RSVP", http.StatusInternalServerError)
			return
		}
	}

	// Send response indicating the RSVP was successfully recorded or updated
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "RSVP recorded"})
}

func GetGroupEvents(w http.ResponseWriter, r *http.Request) {
	groupID := r.URL.Path[len("/event/getGroupEvents/"):]
	fmt.Println("Received groupID:", groupID)

	var events []m.GroupEvent

	query := `SELECT id, group_id, creator_id, title, description, event_date, created_at 
              FROM group_events WHERE group_id = ?`
	rows, err := sqlite.DB.Query(query, groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var event m.GroupEvent
		err := rows.Scan(&event.ID, &event.GroupID, &event.CreatorID, &event.Title, &event.Description, &event.EventDate, &event.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(events)
}







// func GetRSVPs(w http.ResponseWriter, r *http.Request) {
// 	eventID := r.URL.Path[len("/event/rsvps/"):]

// 	fmt.Println("Received groupID:", eventID)

// 	var rsvps []m.GroupEventRSVP

// 	query := `
// 		SELECT id, event_id, user_id, rsvp_status, created_at
// 		FROM group_event_RSVP
// 		WHERE event_id = ?
// 	`
// 	rows, err := sqlite.DB.Query(query, eventID)
// 	if err != nil {
// 		http.Error(w, "Failed to query the database", http.StatusInternalServerError)
// 		return
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var rsvp m.GroupEventRSVP
// 		if err := rows.Scan(&rsvp.ID, &rsvp.EventID, &rsvp.UserID, &rsvp.RSVPStatus, &rsvp.CreatedAt); err != nil {
// 			http.Error(w, "Failed to scan the row", http.StatusInternalServerError)
// 			return
// 		}
// 		rsvps = append(rsvps, rsvp)
// 	}

// 	if err := rows.Err(); err != nil {
// 		http.Error(w, "Error reading rows", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	if err := json.NewEncoder(w).Encode(rsvps); err != nil {
// 		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
// 		return
// 	}
// }






func GetRSVPs(w http.ResponseWriter, r *http.Request) {
    // Extract event ID from the URL
    eventID := r.URL.Path[len("/event/rsvps/"):]

    // Fetch the event details
    var event m.GroupEvent
    err := sqlite.DB.QueryRow(
        "SELECT id, group_id, creator_id, title, description, event_date, created_at FROM group_events WHERE id = ?",
        eventID,
    ).Scan(
        &event.ID, &event.GroupID, &event.CreatorID, &event.Title, &event.Description, &event.EventDate, &event.CreatedAt,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Event not found", http.StatusNotFound)
            return
        }
        http.Error(w, "Failed to fetch event", http.StatusInternalServerError)
        return
    }

    // Fetch RSVPs with usernames
    rows, err := sqlite.DB.Query(`
        SELECT rsvp.rsvp_status, u.username
        FROM group_event_RSVP rsvp
        INNER JOIN users u ON rsvp.user_id = u.id
        WHERE rsvp.event_id = ?`,
        eventID,
    )
    if err != nil {
        http.Error(w, "Failed to fetch RSVPs", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var rsvps []m.RSVPWithUsername
    for rows.Next() {
        var rsvp m.RSVPWithUsername
        if err := rows.Scan(&rsvp.RSVPStatus, &rsvp.Username); err != nil {
            http.Error(w, "Failed to scan RSVP data", http.StatusInternalServerError)
            return
        }
        rsvps = append(rsvps, rsvp)
    }

    if err := rows.Err(); err != nil {
        http.Error(w, "Error reading RSVPs", http.StatusInternalServerError)
        return
    }

    // Construct the final response
    response := m.EventWithRSVPs{
        Event: event,
        RSVPs: rsvps,
    }
fmt.Println(" gbrg ",response)
    // Send the response as JSON
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
        return
    }
}



func getGroupTitleByID(groupID uint) (string, error) {
	var group m.Group
	// Prepare the SQL query
	row := sqlite.DB.QueryRow("SELECT id, title FROM groups WHERE id = ?", groupID)
	err := row.Scan(&group.ID, &group.Title)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("group not found")
		}
		return "", err
	}
	return group.Title, nil
}


func GetnonMembers(w http.ResponseWriter, r *http.Request) {
	// Check if it's a POST request
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	var reqBody struct {
		GroupID int `json:"group_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Check if group_id was successfully extracted
	if reqBody.GroupID == 0 {
		http.Error(w, "group_id is required", http.StatusBadRequest)
		return
	}

	// Log the parsed group_id for debugging
	fmt.Println("Parsed group_id:", reqBody.GroupID)

	// Query to get non-members
	query := `
		SELECT u.id, u.username
		FROM users u
		WHERE u.id NOT IN (
			SELECT gm.user_id
			FROM group_members gm
			WHERE gm.group_id = ?
		);
	`

	// Execute the query
	rows, err := sqlite.DB.Query(query, reqBody.GroupID)
	if err != nil {
		http.Error(w, "Database query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Parse results
	var users []m.User
	for rows.Next() {
		var user m.User
		err := rows.Scan(
			&user.ID,&user.Username,
		)
		if err != nil {
			http.Error(w, "Failed to parse database results: "+err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	// Check for errors after iterating rows
	if err := rows.Err(); err != nil {
		http.Error(w, "Database iteration error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return results as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}


func GetGroupName(w http.ResponseWriter, r *http.Request){
	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var GroupName string
	
	err = sqlite.DB.QueryRow("SELECT title FROM groups WHERE id = ?", id).Scan(&GroupName)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Group does not exist", http.StatusBadRequest)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"group_name": GroupName})
}