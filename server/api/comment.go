package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	m "social-network/models"
	"social-network/pkg/db/sqlite"
)

func CreateComment(w http.ResponseWriter, r *http.Request) {
	var comment m.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(comment.Content) == "" {
		http.Error(w, "Comment can't be empty", http.StatusBadRequest)
		return
	}

	if _, err := sqlite.DB.Exec("INSERT INTO comments (content, media, author, post_id) VALUES (?, ?, ?, ?)", comment.Content, comment.Media, comment.Author, comment.Post_ID); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("insert error: %v", err)
		return
	}

	w.Write([]byte("Comment created"))
}

func GetComments(w http.ResponseWriter, r *http.Request) {
	var comments []m.Comment
	postIDString := r.PathValue("postID")

	postID, err := strconv.Atoi(postIDString)
	if err != nil {
		http.Error(w, "Invalid number", http.StatusBadRequest)
		return
	}
	rows, err := sqlite.DB.Query("SELECT * FROM comments WHERE post_id = ?", postID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post does not exist", http.StatusBadRequest)
			return
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		var comment m.Comment
		if err := rows.Scan(&comment.ID, &comment.Content, &comment.Media, &comment.Post_ID, &comment.Author, &comment.CreatedAt); err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			log.Printf("Error: %v", err)
			return
		}

		comments = append(comments, comment)
	}

	if err := json.NewEncoder(w).Encode(&comments); err != nil {
		http.Error(w, "Error sending data", http.StatusInternalServerError)
		return
	}
}
