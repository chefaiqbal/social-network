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

	//"time"

	"social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
)

// subject to change might be changed to websockets
func RequestFollowUser(w http.ResponseWriter, r *http.Request) {
	log.Println("RequestFollowUser called")
	var follow models.Follow

	if err := json.NewDecoder(r.Body).Decode(&follow); err != nil {
		http.Error(w, "Error reading json", http.StatusBadRequest)
		return
	}

	if followedExists, err := models.DoesUserExist(follow.FollowedID, sqlite.DB); !followedExists {
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			log.Printf("Error checking user existance: %v", err)
			return
		}
		http.Error(w, "User you are trying to follow does not exists", http.StatusBadRequest)
		return
	}

	if followerExists, err := models.DoesUserExist(follow.FollowerID, sqlite.DB); !followerExists {
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			log.Printf("Error checking user existance: %v", err)
			return
		}
		http.Error(w, "User does not exists", http.StatusBadRequest)
		return
	}

	// by default when follow request is created it will always will be peding status
	if _, err := sqlite.DB.Exec("INSERT INTO followers (follower_id, followed_id, status) VALUES (?, ?, ?)", follow.FollowerID, follow.FollowedID, "pending"); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error following user: %v", err)
		return
	}

	w.Write([]byte("Request sent"))
}

func AcceptOrRejectRequest(w http.ResponseWriter, r *http.Request) {
	// get the request if from the url
	requestIdInString := r.PathValue("requestID")
	requestId, err := strconv.Atoi(requestIdInString)
	if err != nil {
		http.Error(w, "Invalid Id", http.StatusBadRequest)
		return
	}

	var resp models.Follow

	// will only send status
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		http.Error(w, "Error reading json", http.StatusBadRequest)
		return
	}

	if !strings.EqualFold(resp.Status, "accept") && !strings.EqualFold(resp.Status, "reject") {
		http.Error(w, "Invalid status type", http.StatusBadRequest)
		return
	}

	// conver the status to always be lower case
	normalizedStatus := strings.ToLower(resp.Status)

	if _, err := sqlite.DB.Exec("UPDATE followers SET status = ? WHERE id = ? AND status = 'pending'", normalizedStatus, requestId); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error updating status: %v", err)
		return
	}

	successMessage := fmt.Sprintf("Successfully %ved user", resp.Status)
	w.Write([]byte(successMessage))
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

	query := `SELECT id, followed_id FROM followers WHERE follower_id = ? AND status = ?`
	rows, err := sqlite.DB.Query(query, userId, "pending")
	if err != nil {
		http.Error(w, "failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var pendingfollows []models.Follow

	for rows.Next() {
		var follow models.Follow

		err := rows.Scan(&follow.ID, &follow.FollowedID)
		if err != nil {
			http.Error(w, "failed to scan row", http.StatusInternalServerError)
			return
		}

		pendingfollows = append(pendingfollows, follow)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "failed to iterate rows", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pendingfollows)
}

func FollowRequestHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "problem in getting user id", http.StatusUnauthorized)
		return
	}

	query := `
    SELECT followers.id, followers.followed_id, users.username, users.avatar
    FROM followers 
    JOIN users ON followers.followed_id = users.id
    WHERE followers.follower_id = ? 
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
			log.Println("Failed to scan row:", err)
			http.Error(w, "failed to scan row", http.StatusInternalServerError)
			return
		}
		followRequests = append(followRequests, followRequest)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "error during row iteration", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(followRequests); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
