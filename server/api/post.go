package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

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
	post.CreatedAt = time.Now()

	result, err := sqlite.DB.Exec(
		"INSERT INTO posts (title, content, media, privacy, author, group_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		post.Title, post.Content, post.Media, post.Privacy, post.Author, post.GroupID, post.CreatedAt,
	)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("create post: %v", err)
		return
	}

	id, _ := result.LastInsertId()
	post.ID = id

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

func ViewPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var post m.Post
	var authorName string
	var avatar sql.NullString

	if err := sqlite.DB.QueryRow(`
		SELECT p.id, p.title, p.content, p.media, p.privacy, p.author, p.created_at,
			   u.username, u.avatar
		FROM posts p
		JOIN users u ON p.author = u.id
		WHERE p.id = ?`, 
		id,
	).Scan(&post.ID, &post.Title, &post.Content, &post.Media, &post.Privacy, 
		   &post.Author, &post.CreatedAt, &authorName, &avatar); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post does not exist", http.StatusNotFound)
			return
		}
		http.Error(w, "Error fetching post", http.StatusInternalServerError)
		log.Printf("view post: %v", err)
		return
	}

	type PostWithAuthor struct {
		m.Post
		AuthorName   string `json:"author_name"`
		AuthorAvatar string `json:"author_avatar,omitempty"`
	}

	postWithAuthor := PostWithAuthor{
		Post:       post,
		AuthorName: authorName,
	}
	if avatar.Valid {
		postWithAuthor.AuthorAvatar = avatar.String
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(postWithAuthor)
}

func GetPosts(w http.ResponseWriter, r *http.Request) {
	rows, err := sqlite.DB.Query(`
		SELECT p.id, p.title, p.content, p.media, p.privacy, p.author, p.created_at, p.group_id,
			   u.username, u.avatar
		FROM posts p
		JOIN users u ON p.author = u.id
		ORDER BY p.created_at DESC
	`)
	if err != nil {
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		log.Printf("get posts: %v", err)
		return
	}
	defer rows.Close()

	type PostWithAuthor struct {
		m.Post
		AuthorName   string `json:"author_name"`
		AuthorAvatar string `json:"author_avatar,omitempty"`
	}

	var posts []PostWithAuthor
	for rows.Next() {
		var post PostWithAuthor
		var media sql.NullString
		var avatar sql.NullString
		var groupID sql.NullInt64

		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&media,
			&post.Privacy,
			&post.Author,
			&post.CreatedAt,
			&groupID,
			&post.AuthorName,
			&avatar,
		); err != nil {
			http.Error(w, "Error reading posts", http.StatusInternalServerError)
			log.Printf("scan post: %v", err)
			return
		}

		if media.Valid {
			post.Media = media.String
		}
		if groupID.Valid {
			post.GroupID = &groupID.Int64
		}
		if avatar.Valid {
			post.AuthorAvatar = avatar.String
		}

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, "Error iterating posts", http.StatusInternalServerError)
		log.Printf("iterate posts: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}
