package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
)

// subject to change might be changed to websockets
func RequestFollowUser(w http.ResponseWriter, r *http.Request) {
	log.Println("RequestFollowUser called")
	
	// Get the current user's ID from the session
	currentUserID, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode the request body
	var request struct {
		FollowedID uint `json:"followed_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error reading json", http.StatusBadRequest)
		return
	}

	// Check if followed user exists and get their privacy status
	var isPrivate bool
	err = sqlite.DB.QueryRow(`
		SELECT is_private FROM users WHERE id = ?
	`, request.FollowedID).Scan(&isPrivate)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User you are trying to follow does not exist", http.StatusBadRequest)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get follower's username for notifications
	var followerUsername string
	err = sqlite.DB.QueryRow(`
		SELECT username FROM users WHERE id = ?
	`, currentUserID).Scan(&followerUsername)
	if err != nil {
		log.Printf("Error getting username: %v", err)
		followerUsername = "Someone" // Fallback username
	}

	// Check if follow relationship already exists
	var existingStatus string
	err = sqlite.DB.QueryRow(`
			SELECT status FROM followers 
			WHERE follower_id = ? AND followed_id = ?
	`, currentUserID, request.FollowedID).Scan(&existingStatus)
	if err == nil {
		// Follow relationship exists
		response := models.FollowResponse{
			Status:  existingStatus,
			Message: fmt.Sprintf("Follow request already exists with status: %s", existingStatus),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	} else if err != sql.ErrNoRows {
		// Unexpected error
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Set initial status based on privacy
	status := "accept"
	if isPrivate {
		status = "pending"
	}

	// Insert follow request
	_, err = sqlite.DB.Exec(`
		INSERT INTO followers (follower_id, followed_id, status, created_at) 
		VALUES (?, ?, ?, ?)
	`, currentUserID, request.FollowedID, status, time.Now())
	if err != nil {
		http.Error(w, "Error creating follow request", http.StatusInternalServerError)
		return
	}

	// If the profile is private, create a notification
	if isPrivate {
		notification := models.Notification{
			ToUserID:   int(request.FollowedID),
			FromUserID: int(currentUserID),
			Content:    fmt.Sprintf("%s wants to follow you", followerUsername),
			Type:       models.NotificationTypeFollow,
			CreatedAt:  time.Now(),
			Read:       false,
		}

		result, err := sqlite.DB.Exec(`
			INSERT INTO notifications 
			(to_user_id, from_user_id, content, type, read, created_at) 
			VALUES (?, ?, ?, ?, ?, ?)
		`, 
		notification.ToUserID, 
		notification.FromUserID, 
		notification.Content,
		notification.Type,
		notification.Read,
		notification.CreatedAt)

		if err != nil {
			log.Printf("Error creating notification: %v", err)
		} else {
			id, _ := result.LastInsertId()
			notification.ID = int(id)
			// Broadcast the notification
			BroadcastNotification(notification)
		}
	}

	// Send response
	response := models.FollowResponse{
		Status:  status,
		Message: fmt.Sprintf("Follow request %s", status),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func AcceptOrRejectRequest(w http.ResponseWriter, r *http.Request) {
	requestIdStr := r.PathValue("id")
	log.Printf("Handling follow request: %s", requestIdStr)

	requestId, err := strconv.Atoi(requestIdStr)
	if err != nil {
		log.Printf("Invalid request ID: %v", err)
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	// Get the current user's ID
	currentUserID, err := util.GetUserID(r, w)
	if err != nil {
		log.Printf("Auth error: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start a transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Get the follower ID and notification ID
	var followerID int
	var notificationID int
	err = tx.QueryRow(`
		SELECT from_user_id, id
		FROM notifications 
		WHERE id = ? AND to_user_id = ? AND type = 'follow_request' AND read = false
	`, requestId, currentUserID).Scan(&followerID, &notificationID)
	if err != nil {
		log.Printf("Error getting follower ID: %v", err)
		http.Error(w, "Follow request not found", http.StatusNotFound)
		return
	}

	// Get usernames for notification before creating notifications
	var followerUsername, currentUsername string
	err = tx.QueryRow("SELECT username FROM users WHERE id = ?", followerID).Scan(&followerUsername)
	if err != nil {
		log.Printf("Error getting follower username: %v", err)
		followerUsername = "Someone"
	}
	err = tx.QueryRow("SELECT username FROM users WHERE id = ?", currentUserID).Scan(&currentUsername)
	if err != nil {
		log.Printf("Error getting current username: %v", err)
		currentUsername = "Someone"
	}

	// Update the follow request status in followers table for both users
	if req.Status == "accept" {
		// Update the original follow request status
		result, err := tx.Exec(`
			UPDATE followers 
			SET status = ? 
			WHERE follower_id = ? AND followed_id = ? AND status = 'pending'
		`, req.Status, followerID, currentUserID)
		if err != nil {
			log.Printf("Error updating follow status: %v", err)
			http.Error(w, "Error updating follow status", http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		log.Printf("Updated %d rows in followers table for original request", rowsAffected)

		// Create reciprocal follow relationship
		_, err = tx.Exec(`
			INSERT OR IGNORE INTO followers (follower_id, followed_id, status, created_at)
			VALUES (?, ?, 'accept', ?)
		`, currentUserID, followerID, time.Now())
		if err != nil {
			log.Printf("Error creating reciprocal follow: %v", err)
			http.Error(w, "Error creating reciprocal follow", http.StatusInternalServerError)
			return
		}

		// Create acceptance notification
		_, err = tx.Exec(`
			INSERT INTO notifications (
				to_user_id, 
				from_user_id, 
				content, 
				type, 
				read, 
				created_at
			) VALUES (?, ?, ?, ?, false, ?)
		`, followerID, currentUserID, 
		   fmt.Sprintf("%s accepted your follow request", currentUsername),
		   models.NotificationTypeAccept, time.Now())

		if err != nil {
			log.Printf("Error creating acceptance notification: %v", err)
		}
	} else {
		// If rejecting, just delete the follow request
		_, err = tx.Exec(`
			DELETE FROM followers 
			WHERE follower_id = ? AND followed_id = ? AND status = 'pending'
		`, followerID, currentUserID)
		if err != nil {
			log.Printf("Error deleting follow request: %v", err)
			http.Error(w, "Error deleting follow request", http.StatusInternalServerError)
			return
		}
	}

	// Delete all follow request notifications
	_, err = tx.Exec(`
		DELETE FROM notifications 
		WHERE to_user_id = ? 
		AND from_user_id = ? 
		AND type = 'follow_request'
	`, currentUserID, followerID)
	if err != nil {
		log.Printf("Error deleting follow request notifications: %v", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": fmt.Sprintf("Follow request %sed", req.Status),
	})
}

func GetFollowers(w http.ResponseWriter, r *http.Request) {
	// Get the user ID from the request
	userId, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "problem in getting user id", http.StatusUnauthorized)
		return
	}

	var followers []models.Follow

	// Modify the query to join the `followers` table with the `users` table
	query := `
        SELECT f.id, f.follower_id, f.followed_id, f.status, f.created_at, u.username
        FROM followers f
        JOIN users u ON u.id = f.follower_id
        WHERE f.followed_id = ? AND f.status = 'accept'`

	rows, err := sqlite.DB.Query(query, userId)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No followers found", http.StatusBadRequest)
			return
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error getting followers: %v", err)
		return
	}

	// Iterate over the rows to fetch the follower information
	for rows.Next() {
		var follower models.Follow
		if err := rows.Scan(&follower.ID, &follower.FollowerID, &follower.FollowedID, &follower.Status, &follower.CreatedAt, &follower.Username); err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			log.Printf("Error scanning follower: %v", err)
			return
		}

		followers = append(followers, follower)
	}

	// Send the followers data back as a JSON response
	if err := json.NewEncoder(w).Encode(&followers); err != nil {
		http.Error(w, "Error sending data", http.StatusInternalServerError)
	}
}

func CloseFriend(w http.ResponseWriter, r *http.Request) {
	userId, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "problem in getting user id", http.StatusUnauthorized)
		return
	}

	var closeFriend models.CloseFriends

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &closeFriend); err != nil {
		http.Error(w, "failed to unmarshal request body", http.StatusBadRequest)
		return
	}

	closeFriendsStr := strings.Join(closeFriend.Usernames, ",")

	// Check if the user has an existing entry in post_PrivateViews
	var existingCloseFriends string
	query := `SELECT close_friends FROM post_PrivateViews WHERE user_id = ?`
	err = sqlite.DB.QueryRow(query, userId).Scan(&existingCloseFriends)

	if err == sql.ErrNoRows {
		log.Printf("No existing entry found for user_id %d. Inserting new entry.", userId)

		// Insert a new entry for the user if no row exists
		insertQuery := `INSERT INTO post_PrivateViews (user_id, close_friends) VALUES (?, ?)`
		_, insertErr := sqlite.DB.Exec(insertQuery, userId, closeFriendsStr)
		if insertErr != nil {
			log.Printf("Failed to insert close friends for new user_id %d: %v", userId, insertErr)
			http.Error(w, "failed to insert close friends for new user", http.StatusInternalServerError)
			return
		}

		response := map[string]string{"message": "Close friends added successfully"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return

	} else if err != nil {
		log.Printf("Error retrieving close friends for user_id %d: %v", userId, err)
		http.Error(w, "failed to retrieve existing close friends", http.StatusInternalServerError)
		return
	}

	// updating if existing close friends are found
	existingList := strings.Split(existingCloseFriends, ",")
	for i, user := range existingList {
		existingList[i] = strings.TrimSpace(user)
	}

	// check selected usernames and update list
	friendMap := make(map[string]bool)
	for _, user := range closeFriend.Usernames {
		friendMap[user] = true
	}

	updatedList := make([]string, 0)
	for _, user := range existingList {
		if friendMap[user] {
			updatedList = append(updatedList, user)
			delete(friendMap, user)
		}
	}

	for user := range friendMap {
		updatedList = append(updatedList, user)
	}

	updatedCloseFriendsStr := strings.Join(updatedList, ",")
	updateQuery := `UPDATE post_PrivateViews SET close_friends = ? WHERE user_id = ?`
	_, err = sqlite.DB.Exec(updateQuery, updatedCloseFriendsStr, userId)
	if err != nil {
		log.Printf("Error updating close friends for user_id %d: %v", userId, err)
		http.Error(w, "failed to update close friends", http.StatusInternalServerError)
		return
	}

	response := map[string]string{"message": "Close friends updated successfully"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func GetFollowstatus(w http.ResponseWriter, r *http.Request) {
	userId, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "problem in getting user id", http.StatusUnauthorized)
		return
	}

	query := `
		SELECT id, followed_id, status 
		FROM followers 
		WHERE follower_id = ?
	`
	rows, err := sqlite.DB.Query(query, userId)
	if err != nil {
		http.Error(w, "failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	follows := []struct {
		ID         int    `json:"id"`
		FollowedID int    `json:"followed_id"`
		Status     string `json:"status"`
	}{}

	for rows.Next() {
		var follow struct {
			ID         int    `json:"id"`
			FollowedID int    `json:"followed_id"`
			Status     string `json:"status"`
		}
		if err := rows.Scan(&follow.ID, &follow.FollowedID, &follow.Status); err != nil {
			continue
		}
		follows = append(follows, follow)
	}

	// Always return an array, even if empty
	if follows == nil {
		follows = []struct {
			ID         int    `json:"id"`
			FollowedID int    `json:"followed_id"`
			Status     string `json:"status"`
		}{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(follows)
}

func FollowRequestHandler(w http.ResponseWriter, r *http.Request) {
    userId, err := util.GetUserID(r, w)
    if err != nil {
        http.Error(w, "problem in getting user id", http.StatusUnauthorized)
        return
    }

    query := `
    SELECT followers.id, followers.follower_id, users.username, COALESCE(users.avatar, '') as avatar
    FROM followers 
    JOIN users ON followers.follower_id = users.id
    WHERE followers.followed_id = ? 
      AND followers.status = ?
    `

    rows, err := sqlite.DB.Query(query, userId, "pending")
    if err != nil {
        http.Error(w, "failed to query database", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    followRequests := []models.FollowRequest{}

    for rows.Next() {
        var followRequest models.FollowRequest
        if err := rows.Scan(&followRequest.ID, &followRequest.FollowedID, &followRequest.Username, &followRequest.Avatar); err != nil {
            log.Printf("Error scanning row: %v", err)
            continue // Skip this row but continue processing others
        }
        followRequests = append(followRequests, followRequest)
    }

    if err := rows.Err(); err != nil {
        http.Error(w, "error during row iteration", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(followRequests)
}

// Add this function to check follow status
func GetFollowStatus(w http.ResponseWriter, r *http.Request, followerID, followedID uint) (string, error) {
	var status string
	err := sqlite.DB.QueryRow(`
		SELECT status FROM followers 
		WHERE follower_id = ? AND followed_id = ?
	`, followerID, followedID).Scan(&status)
	
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return status, nil
}

// Add this function to get follow requests
func GetFollowRequests(w http.ResponseWriter, r *http.Request) {
	userID, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := sqlite.DB.Query(`
		SELECT f.id, u.username, u.avatar
		FROM followers f
		JOIN users u ON f.follower_id = u.id
		WHERE f.followed_id = ? AND f.status = 'pending'
	`, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var requests []models.FollowRequest
	for rows.Next() {
		var req models.FollowRequest
		var avatar sql.NullString
		if err := rows.Scan(&req.ID, &req.Username, &avatar); err != nil {
			continue
		}
		if avatar.Valid {
			req.Avatar = avatar.String
		}
		requests = append(requests, req)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

// Add this function to check if users are friends
func AreFriends(userID1, userID2 uint64) (bool, error) {
	var count int
	err := sqlite.DB.QueryRow(`
		SELECT COUNT(*) FROM followers 
		WHERE follower_id = ? AND followed_id = ? 
		AND status = 'accept'
		AND EXISTS (
			SELECT 1 FROM followers 
			WHERE follower_id = ? AND followed_id = ? 
			AND status = 'accept'
		)
	`, userID1, userID2, userID2, userID1).Scan(&count)
	
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func GetFollowing(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	if userId == "current" {
		var err error
		currentUserID, err := util.GetUserID(r, w)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userId = fmt.Sprintf("%d", currentUserID)
	}

	query := `
		SELECT 
			f.id, 
			f.followed_id, 
			u.username, 
			COALESCE(u.avatar, '') as avatar, 
			f.status,
			f.created_at
		FROM followers f
		JOIN users u ON u.id = f.followed_id
		WHERE f.follower_id = ? AND f.status = 'accept'
	`

	rows, err := sqlite.DB.Query(query, userId)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var following []models.Follow
	for rows.Next() {
		var follow models.Follow
		if err := rows.Scan(
			&follow.ID,
			&follow.FollowedID,
			&follow.Username,
			&follow.Avatar,
			&follow.Status,
			&follow.CreatedAt,
		); err != nil {
			log.Printf("Error scanning follow row: %v", err)
			continue
		}
		following = append(following, follow)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(following)
}
