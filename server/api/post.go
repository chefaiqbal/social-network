package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	m "social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
)

func CreatePost(w http.ResponseWriter, r *http.Request) {
	var post m.Post

	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "Error reading data", http.StatusBadRequest)
		return
	}

	// check if the passed privacy is within the allowed range
	if post.Privacy != 1 && post.Privacy != 2 && post.Privacy != 3 {
		http.Error(w, "invalid privacy type", http.StatusBadRequest)
		return
	}

	// Get the user ID from the session
	cookie, err := r.Cookie("AccessToken")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := util.UserSession[cookie.Value]
	post.Author = userID

	if _, err := sqlite.DB.Exec(
		"INSERT INTO posts (title, content, media, privacy, author, group_id) VALUES (?, ?, ?, ?, ?, ?)",
		post.Title, post.Content, post.Media, post.Privacy, post.Author, post.GroupID,
	); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("create post: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Post created successfully"})
}

func ViewPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var post m.Post
	if err := sqlite.DB.QueryRow(
		"SELECT id, title, content, media, privacy, author, created_at FROM posts WHERE id = ?", 
		id,
	).Scan(&post.ID, &post.Title, &post.Content, &post.Media, &post.Privacy, &post.Author, &post.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post does not exist", http.StatusNotFound)
			return
		}
		http.Error(w, "Error fetching post", http.StatusInternalServerError)
		log.Printf("view post: %v", err)
		return
	}

	// Get the user ID from the session
	cookie, err := r.Cookie("AccessToken")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := util.UserSession[cookie.Value]

	// Check privacy settings
	switch post.Privacy {
	case 1: // Public
		json.NewEncoder(w).Encode(post)
	case 2: // Private - only followers can see
		var followerID uint64
		err := sqlite.DB.QueryRow(
			"SELECT follower_id FROM followers WHERE followed_id = ? AND status = 'accept' AND follower_id = ?",
			post.Author, userID,
		).Scan(&followerID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Not authorized to view this post", http.StatusForbidden)
				return
			}
			http.Error(w, "Error checking follow status", http.StatusInternalServerError)
			log.Printf("check follower: %v", err)
			return
		}
		json.NewEncoder(w).Encode(post)
	case 3: // Almost private - only selected users can see
		var viewerID int
		err := sqlite.DB.QueryRow(
			"SELECT user_id FROM post_PrivateViews WHERE post_id = ? AND user_id = ?",
			post.ID, userID,
		).Scan(&viewerID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Not authorized to view this post", http.StatusForbidden)
				return
			}
			http.Error(w, "Error checking view permission", http.StatusInternalServerError)
			log.Printf("check viewer: %v", err)
			return
		}
		json.NewEncoder(w).Encode(post)
	default:
		http.Error(w, "Invalid privacy setting", http.StatusBadRequest)
	}
}

func GetPosts(w http.ResponseWriter, r *http.Request) {
	rows, err := sqlite.DB.Query("SELECT id, title, content, media, privacy, author, created_at FROM posts ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		log.Printf("get posts: %v", err)
		return
	}
	defer rows.Close()

	var posts []m.Post
	for rows.Next() {
		var post m.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Media, &post.Privacy, &post.Author, &post.CreatedAt); err != nil {
			http.Error(w, "Error reading posts", http.StatusInternalServerError)
			log.Printf("scan post: %v", err)
			return
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, "Error iterating posts", http.StatusInternalServerError)
		log.Printf("iterate posts: %v", err)
		return
	}

	json.NewEncoder(w).Encode(posts)
}
