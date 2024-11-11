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


func CreatePost(w http.ResponseWriter, r *http.Request) {
	var post m.Post

	// Decode JSON data into the post struct
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "Error reading data", http.StatusBadRequest)
		return
	}

	// Validate privacy value
	if post.Privacy != 1 && post.Privacy != 2 && post.Privacy != 3 {
		http.Error(w, "Invalid privacy type", http.StatusBadRequest)
		return
	}

	// Set the CreatedAt timestamp
	post.CreatedAt = time.Now()

	// Check if media is empty and set to NULL if so
	if post.Media.String == "" {
		post.Media = sql.NullString{String: "", Valid: false} // NULL value
	} else {
		post.Media.Valid = true
	}

	// Insert post into the database
	result, err := sqlite.DB.Exec(
		"INSERT INTO posts (title, content, media, privacy, author, group_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		post.Title, post.Content, post.Media, post.Privacy, post.Author, post.GroupID, post.CreatedAt,
	)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("create post: %v", err)
		return
	}

	// Retrieve and assign the new post ID
	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to retrieve post ID", http.StatusInternalServerError)
		return
	}
	post.ID = id

	// Respond with the created post
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
    userID, err := util.GetUserID(r, w)
    if err != nil {
        http.Error(w, "Problem in getting user ID", http.StatusUnauthorized)
        return
    }

    var closeFriendsStr string
    query := `SELECT close_friends FROM post_PrivateViews WHERE user_id = ?`
    err = sqlite.DB.QueryRow(query, userID).Scan(&closeFriendsStr)
    if err != nil && err != sql.ErrNoRows {
        http.Error(w, "Error fetching close friends", http.StatusInternalServerError)
        log.Printf("get close friends: %v", err)
        return
    }

    closeFriendsArray := []string{}
    if closeFriendsStr != "" {
        closeFriendsArray = strings.Split(closeFriendsStr, ",")
    }

	log.Printf("close friends: %v", closeFriendsArray)

    rows, err := sqlite.DB.Query(`
        SELECT p.id, p.title, p.content, p.media, p.privacy, p.author, p.created_at, p.group_id,
               u.username, u.avatar
        FROM posts p
        JOIN users u ON p.author = u.id
        LEFT JOIN followers f ON f.followed_id = p.author AND f.follower_id = ? AND f.status = 'active'
        WHERE 
            p.privacy = 1 OR 
            p.author = ? OR 
            (p.privacy = 2 AND f.follower_id IS NOT NULL) OR
            (p.privacy = 3 AND u.username IN (?))
        ORDER BY p.created_at DESC
    `, userID, userID, strings.Join(closeFriendsArray, ","))
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
            post.Media = media
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

    // Step 4: Return the filtered posts
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(posts)
}

func GetUserPosts(w http.ResponseWriter, r *http.Request) {
    userIdString := r.PathValue("id")
    var targetUserID int64

    // Check if we're requesting the current user's posts
    if userIdString == "current" {
        // Get the current user's ID from the session
        currentUserID, err := util.GetUserID(r, w)
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        targetUserID = int64(currentUserID)
    } else {
        // Convert id to number
        var err error
        targetUserID, err = strconv.ParseInt(userIdString, 10, 64)
        if err != nil {
            http.Error(w, "Invalid user ID", http.StatusBadRequest)
            return
        }
    }

    // Query to get posts based on privacy settings
    rows, err := sqlite.DB.Query(`
        SELECT p.id, p.title, p.content, p.media, p.privacy, p.author, p.created_at,
               u.username as author_name, u.avatar as author_avatar
        FROM posts p
        JOIN users u ON p.author = u.id
        LEFT JOIN followers f ON f.followed_id = p.author AND f.follower_id = ?
        WHERE p.author = ? AND (
            p.privacy = 1 OR
            p.author = ? OR
            (p.privacy = 2 AND f.status = 'accept')
        )
        ORDER BY p.created_at DESC
    `, targetUserID, targetUserID, targetUserID)

    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
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
        var media, avatar sql.NullString
        if err := rows.Scan(
            &post.ID, &post.Title, &post.Content, &media, &post.Privacy,
            &post.Author, &post.CreatedAt, &post.AuthorName, &avatar,
        ); err != nil {
            http.Error(w, "Error scanning posts", http.StatusInternalServerError)
            return
        }

        if media.Valid {
            post.Media = media
        }
        if avatar.Valid {
            post.AuthorAvatar = avatar.String
        }

        posts = append(posts, post)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(posts)
}


