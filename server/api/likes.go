package api

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "social-network/models"
    "social-network/pkg/db/sqlite"
    "social-network/util"
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

    // Use a transaction to ensure data consistency
    tx, err := sqlite.DB.Begin()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer tx.Rollback()

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

    // Get total likes
    err := sqlite.DB.QueryRow(`
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