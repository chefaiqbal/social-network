package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	m "social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
)

func CreateComment(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var comment m.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(comment.Content) == "" {
		http.Error(w, "Comment can't be empty", http.StatusBadRequest)
		return
	}

	comment.Author = uint(currentUserID)

	if _, err := sqlite.DB.Exec("INSERT INTO comments (content, media, author, post_id) VALUES (?, ?, ?, ?)", comment.Content, comment.Media, comment.Author, comment.Post_ID); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("insert error: %v", err)
		return
	}

	w.Write([]byte("Comment created"))
}

func GetComments(w http.ResponseWriter, r *http.Request) {
	postIDString := r.PathValue("postID")

	postID, err := strconv.Atoi(postIDString)
	if err != nil {
		http.Error(w, "Invalid number", http.StatusBadRequest)
		return
	}

	query := `
		SELECT 
			c.id,
			c.content,
			c.media,
			c.post_id,
			c.author,
			c.created_at,
			u.username as author_name,
			u.avatar as author_avatar
		FROM comments c
		JOIN users u ON c.author = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at DESC
	`

	rows, err := sqlite.DB.Query(query, postID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post does not exist", http.StatusBadRequest)
			return
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type CommentResponse struct {
		ID           int       `json:"id"`
		Content      string    `json:"content"`
		Media        string    `json:"media,omitempty"`
		PostID       int       `json:"post_id"`
		Author       int       `json:"author"`
		CreatedAt    time.Time `json:"created_at"`
		AuthorName   string    `json:"author_name"`
		AuthorAvatar string    `json:"author_avatar,omitempty"`
	}

	var comments []CommentResponse
	for rows.Next() {
		var comment CommentResponse
		var media, avatar sql.NullString
		if err := rows.Scan(
			&comment.ID,
			&comment.Content,
			&media,
			&comment.PostID,
			&comment.Author,
			&comment.CreatedAt,
			&comment.AuthorName,
			&avatar,
		); err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			log.Printf("Error: %v", err)
			return
		}

		if media.Valid {
			comment.Media = media.String
		}
		if avatar.Valid {
			comment.AuthorAvatar = avatar.String
		}

		comments = append(comments, comment)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(comments); err != nil {
		http.Error(w, "Error sending data", http.StatusInternalServerError)
		return
	}
}

func GetCommentCount(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.PathValue("postID")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var count int
	err = sqlite.DB.QueryRow("SELECT COUNT(*) FROM comments WHERE post_id = ?", postID).Scan(&count)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}
