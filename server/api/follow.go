package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"social-network/models"
	"social-network/pkg/db/sqlite"
)

// subject to change might be changed to websockets
func RequestFollowUser(w http.ResponseWriter, r *http.Request) {
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
	userIdString := r.PathValue("userID")
	userId, err := strconv.Atoi(userIdString)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	var followers []models.Follow

	rows, err := sqlite.DB.Query("SELECT * FROM followers WHERE followed_id = ? AND status = 'accept'", userId)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No followers found", http.StatusBadRequest)
			return
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error getting followers: %v", err)
		return
	}

	for rows.Next() {
		var follower models.Follow
		if err := rows.Scan(&follower.ID, &follower.FollowerID, &follower.FollowedID, &follower.Status, &follower.CreatedAt); err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			log.Printf("Error scanning follower: %v", err)
			return
		}

		followers = append(followers, follower)
	}

	if err := json.NewEncoder(w).Encode(&followers); err != nil {
		http.Error(w, "Error sending data", http.StatusInternalServerError)
	}
}

