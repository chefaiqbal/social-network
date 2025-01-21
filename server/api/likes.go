package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
	"strconv"
	"strings"
)

func LikeHandler(w http.ResponseWriter, r *http.Request) {
    cookie, _ := r.Cookie("AccessToken")
    userID := int(util.UserSession[cookie.Value])

    var like models.Likes
    if err := json.NewDecoder(r.Body).Decode(&like); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    like.UserID = userID
    postID := like.PostID

    currentUserID, err := util.GetUserID(r, w)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    // Use a transaction to ensure data consistency
    tx, err := sqlite.DB.Begin()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer tx.Rollback()

    var postExist bool
    err = tx.QueryRow(`SELECT 1 FROM posts WHERE id = ?`, postID).Scan(&postExist);
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if !postExist {
        http.Error(w, "Post not found", http.StatusNotFound)
        return
    }

    var postPrivacy int
    var postAuthorID int
    err = tx.QueryRow(`
        SELECT privacy, author FROM posts WHERE id = ?`, postID,
    ).Scan(&postPrivacy, &postAuthorID)
    if err != nil {
        log.Printf("Database error checking post privacy: %v", err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    // Privacy checks
    if currentUserID != uint64(postAuthorID) {
        switch postPrivacy {
        case 2: 
            var isFollower bool
            err = sqlite.DB.QueryRow(`
                SELECT EXISTS (
                    SELECT 1 FROM followers
                    WHERE follower_id = ? AND followed_id = ?
                )`, currentUserID, postAuthorID).Scan(&isFollower)
            if err != nil {
                log.Printf("Error checking follower status: %v", err)
                http.Error(w, "Error checking follower status", http.StatusInternalServerError)
                return
            }
            if !isFollower {
                http.Error(w, "You must follow the author to like this post.", http.StatusForbidden)
                return
            }

        case 3: 
            var closeFriends sql.NullString
            err = sqlite.DB.QueryRow(`
                SELECT close_friends FROM post_PrivateViews WHERE user_id = ?`, postAuthorID).Scan(&closeFriends)
            if err != nil {
                log.Printf("Error checking close friends list: %v", err)
                http.Error(w, "Error checking close friends list", http.StatusInternalServerError)
                return
            }
            if !closeFriends.Valid || closeFriends.String == "" {
                http.Error(w, "You are not in the author's close friends list.", http.StatusForbidden)
                return
            }

            closeFriendsList := strings.Split(closeFriends.String, ",")
            isCloseFriend := false
            for _, friend := range closeFriendsList {
                friend = strings.TrimSpace(friend)
                if strconv.Itoa(int(currentUserID)) == friend {
                    isCloseFriend = true
                    break
                }
            }

            if !isCloseFriend {
                http.Error(w, "You are not in the author's close friends list.", http.StatusForbidden)
                return
            }

        default:
            http.Error(w, "This post's privacy settings do not allow you to like it.", http.StatusForbidden)
            return
        }
    }

    var existingLike models.Likes
    query := `SELECT id, is_like FROM likes WHERE user_id = ? AND post_id = ?`
    err = tx.QueryRow(query, like.UserID, like.PostID).Scan(&existingLike.ID, &existingLike.Like)

    var newLikeState bool
    if err == nil {
        // Update existing like
        newLikeState = !existingLike.Like
        _, err = tx.Exec(`UPDATE likes SET is_like = ? WHERE id = ?`, newLikeState, existingLike.ID)
    } else {
        // Create new like
        newLikeState = true
        _, err = tx.Exec(`INSERT INTO likes (user_id, post_id, is_like) VALUES (?, ?, ?)`,
            like.UserID, like.PostID, newLikeState)
    }

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Get updated like count within the transaction
    var likeCount int
    err = tx.QueryRow(`SELECT COUNT(*) FROM likes WHERE post_id = ? AND is_like = true`,
        like.PostID).Scan(&likeCount)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if err = tx.Commit(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Broadcast update
    update := LikeUpdate{
        PostID:    like.PostID,
        LikeCount: likeCount,
        UserLiked: newLikeState,
        UserID:    userID,
    }
    go broadcastLikeUpdate(update)

    json.NewEncoder(w).Encode(update)
}

func GetPostLikes(w http.ResponseWriter, r *http.Request) {
    postID := r.URL.Query().Get("post_id")
    if postID == "" {
        http.Error(w, "Post ID is required", http.StatusBadRequest)
        return
    }

    cookie, _ := r.Cookie("AccessToken")
    userID := int(util.UserSession[cookie.Value])

    var response struct {
        LikeCount int  `json:"like_count"`
        UserLiked bool `json:"user_liked"`
    }

    var err error
    var postExists bool

    err = sqlite.DB.QueryRow(`SELECT 1 FROM posts WHERE id = ?`, postID).Scan(&postExists)
    if err != nil {
        http.Error(w, "Post not found", http.StatusNotFound)
        return
    }

    if !postExists {
        http.Error(w, "Post not found", http.StatusNotFound)
        return
    }

    var postPrivacy int
    var postAuthorID int
    err = sqlite.DB.QueryRow(`
        SELECT privacy, author FROM posts WHERE id = ?`, postID,
    ).Scan(&postPrivacy, &postAuthorID)
    if err != nil {
        log.Printf("Database error checking post privacy: %v", err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    // Privacy checks
    if userID != postAuthorID {
        switch postPrivacy {
        case 2: 
            var isFollower bool
            err = sqlite.DB.QueryRow(`
                SELECT EXISTS (
                    SELECT 1 FROM followers
                    WHERE follower_id = ? AND followed_id = ?
                )`, userID, postAuthorID).Scan(&isFollower)
            if err != nil {
                log.Printf("Error checking follower status: %v", err)
                http.Error(w, "Error checking follower status", http.StatusInternalServerError)
                return
            }
            if !isFollower {
                http.Error(w, "You must follow the author to like this post.", http.StatusForbidden)
                return
            }

        case 3: 
            var closeFriends sql.NullString
            err = sqlite.DB.QueryRow(`
                SELECT close_friends FROM post_PrivateViews WHERE user_id = ?`, postAuthorID).Scan(&closeFriends)
            if err != nil {
                log.Printf("Error checking close friends list: %v", err)
                http.Error(w, "Error checking close friends list", http.StatusInternalServerError)
                return
            }
            if !closeFriends.Valid || closeFriends.String == "" {
                http.Error(w, "You are not in the author's close friends list.", http.StatusForbidden)
                return
            }

            closeFriendsList := strings.Split(closeFriends.String, ",")
            isCloseFriend := false
            for _, friend := range closeFriendsList {
                friend = strings.TrimSpace(friend)
                if strconv.Itoa(int(userID)) == friend {
                    isCloseFriend = true
                    break
                }
            }

            if !isCloseFriend {
                http.Error(w, "You are not in the author's close friends list.", http.StatusForbidden)
                return
            }

        default:
            http.Error(w, "This post's privacy settings do not allow you to like it.", http.StatusForbidden)
            return
        }
    }


    // Get total likes
    err = sqlite.DB.QueryRow(`
        SELECT 
            (SELECT COUNT(*) FROM likes WHERE post_id = ? AND is_like = true),
            COALESCE((SELECT is_like FROM likes WHERE post_id = ? AND user_id = ?), false)
        FROM likes 
        WHERE post_id = ? 
        LIMIT 1
    `, postID, postID, userID, postID).Scan(&response.LikeCount, &response.UserLiked)

    if err != nil && err != sql.ErrNoRows {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // If no rows found, set default values
    if err == sql.ErrNoRows {
        response.LikeCount = 0
        response.UserLiked = false
    }

    json.NewEncoder(w).Encode(response)
} 