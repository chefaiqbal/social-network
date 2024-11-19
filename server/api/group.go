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

