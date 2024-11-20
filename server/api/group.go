package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"social-network/models"
	m "social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
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
	var post m.Post

	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	groupIDString := r.PathValue("id")

	// convert the string into a number
	groupID, err := strconv.Atoi(groupIDString)
	if err != nil {
		http.Error(w, "Invalid number", http.StatusBadRequest)
		return
	}

	// post will always be public for the group members
	post.Privacy = 1

	// check if the passed privacy is within the allowed range
	if post.Privacy != 1 && post.Privacy != 2 && post.Privacy != 3 {
		http.Error(w, "invalid privacy type", http.StatusBadRequest)
		return
	}
	

	if _, err := sqlite.DB.Exec("INSERT INTO posts (title, content, media, privacy, author, group_id) VALUES (?, ?, ?, ?, ?, ?)", post.Title, post.Content, post.Media, post.Privacy, post.Author, groupID); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("create post: %v", err)
		return
	}

	w.Write([]byte("Post created successfully"))
}

func GetGroupPost(w http.ResponseWriter, r *http.Request) {
	var groupPosts []m.Post
	groupIDString := r.PathValue("id")

	// convert the string into a number
	groupID, err := strconv.Atoi(groupIDString)
	if err != nil {
		http.Error(w, "Invalid number", http.StatusBadRequest)
		return
	}

	// the value of the group id can't be less then 1
	if groupID < 1 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	rows, err := sqlite.DB.Query("SELECT id, title, content, media, privacy, author, created_at, group_id FROM posts WHERE group_id = ?", groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Group does not exists", http.StatusBadRequest)
			return
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var post m.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Media, &post.Privacy, &post.Author, &post.CreatedAt, &post.GroupID); err != nil {
			http.Error(w, "Error getting post", http.StatusInternalServerError)
			log.Printf("Error scanning: %v", err)
			return
		}

		groupPosts = append(groupPosts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&groupPosts); err != nil {
		http.Error(w, "Error sending json", http.StatusInternalServerError)
	}
}


func VeiwGorups(w http.ResponseWriter, r *http.Request) {
	var groups []m.Group

	rows, err := sqlite.DB.Query("SELECT * FROM groups")
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

	userID, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&inviteRequest); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	_, err = sqlite.DB.Exec(
		"INSERT INTO group_members (group_id, user_id, status) VALUES (?, ?, ?)",
		inviteRequest.GroupID, userID, "pending",
	)

	if err != nil {
		http.Error(w, "Failed to join the group", http.StatusInternalServerError)
		log.Printf("Error inserting group member: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Group invitation sent successfully"})
}



func GroupAccept(w http.ResponseWriter, r *http.Request) {
    var inviteRequest m.GroupInvaitation

    if err := json.NewDecoder(r.Body).Decode(&inviteRequest); err != nil {
        http.Error(w, "Invalid JSON data", http.StatusBadRequest)
        return
    }

    // Validate input
    if inviteRequest.GroupID <= 0 || inviteRequest.ReciverID <= 0 {
        http.Error(w, "Invalid group ID or receiver ID", http.StatusBadRequest)
        return
    }

    // Check if the invitation exists and is pending
    var status string
    err := sqlite.DB.QueryRow("SELECT status FROM group_members WHERE group_id = ? AND user_id = ?", inviteRequest.GroupID, inviteRequest.ReciverID).Scan(&status)
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
    result, err := sqlite.DB.Exec("UPDATE group_members SET status = 'member' WHERE group_id = ? AND user_id = ?", inviteRequest.GroupID, inviteRequest.ReciverID)
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
    var inviteRequest m.GroupInvaitation

    if err := json.NewDecoder(r.Body).Decode(&inviteRequest); err != nil {
        http.Error(w, "Invalid JSON data", http.StatusBadRequest)
        return
    }

    // Validate input
    if inviteRequest.GroupID <= 0 || inviteRequest.ReciverID <= 0 {
        http.Error(w, "Invalid group ID or receiver ID", http.StatusBadRequest)
        return
    }

    // Check if the invitation exists
    var status string
    err := sqlite.DB.QueryRow("SELECT status FROM group_members WHERE group_id = ? AND user_id = ?", inviteRequest.GroupID, inviteRequest.ReciverID).Scan(&status)
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
    result, err := sqlite.DB.Exec("DELETE FROM group_members WHERE group_id = ? AND user_id = ?", inviteRequest.GroupID, inviteRequest.ReciverID)
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


//get our group
func MyGroups(w http.ResponseWriter, r *http.Request) {
	// Get the user ID from the request
	userID, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Query to fetch group IDs
	groupIDsQuery := `SELECT group_id FROM group_members WHERE user_id = ?`
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

func Members (w http.ResponseWriter, r *http.Request) {
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

	_, err = sqlite.DB.Exec(query, id)
	if err != nil {
		http.Error(w, "Error deleting group", http.StatusInternalServerError)
		log.Printf("Error deleting group: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}





///note:cheack if the group exist or not, is he a member or not?
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
    _,err = sqlite.DB.Exec(query, event.GroupID, event.CreatorID, event.Title, event.Description, event.EventDate)
    if err != nil {
        http.Error(w, "Error creating event", http.StatusInternalServerError)
        return
    }

    // Batch notifications
    rows, err := sqlite.DB.Query("SELECT user_id FROM group_members WHERE group_id = ?", event.GroupID)
    if err != nil {
        log.Printf("Error fetching group members: %v", err)
        return
    }
    defer rows.Close()

    var notifications []models.Notification
    for rows.Next() {
        var userID int
        if err := rows.Scan(&userID); err != nil {
            log.Printf("Error scanning user_id: %v", err)
            continue
        }
        notifications = append(notifications, models.Notification{
            ToUserID:   userID,
            GroupID:    int(event.GroupID),
            Content:    fmt.Sprintf("%s ..Join To Us", event.Title),
            Type:       models.NotificationEvent,
            CreatedAt:  time.Now(),
            Read:       false,
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






func GetGroupEvents(w http.ResponseWriter, r *http.Request) {
	groupID := r.URL.Query().Get("group_id")
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


func GetRSVPs(w http.ResponseWriter, r *http.Request) {
	eventID := r.URL.Query().Get("event_id")
	rows, err := sqlite.DB.Query(`SELECT id, event_id, user_id, rsvp_status, created_at 
	                                FROM group_event_RSVP WHERE event_id = ?`, eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var rsvps []m.GroupEventRSVP
	for rows.Next() {
		var rsvp m.GroupEventRSVP
		if err := rows.Scan(&rsvp.ID, &rsvp.EventID, &rsvp.UserID, &rsvp.RSVPStatus, &rsvp.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		rsvps = append(rsvps, rsvp)
	}

	json.NewEncoder(w).Encode(rsvps)
}