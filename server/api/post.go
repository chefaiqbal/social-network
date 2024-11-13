package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"encoding/base64"

	m "social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
)


func CreatePost(w http.ResponseWriter, r *http.Request) {
	var postInput struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		Media    string `json:"media"`      // Base64 string from frontend
		Privacy  int    `json:"privacy"`
		GroupID  *int64 `json:"group_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&postInput); err != nil {
		http.Error(w, "Error reading data", http.StatusBadRequest)
		return
	}

	// Get the current user's ID
	userID, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Validate privacy value
	if postInput.Privacy != 1 && postInput.Privacy != 2 && postInput.Privacy != 3 {
		http.Error(w, "Invalid privacy type", http.StatusBadRequest)
		return
	}

	var mediaBytes []byte
	var mediaType string

	// Process media if provided
	if postInput.Media != "" {
		// Split the base64 string to get the media type
		parts := strings.Split(postInput.Media, ";base64,")
		if len(parts) != 2 {
			http.Error(w, "Invalid media format", http.StatusBadRequest)
			return
		}

		mediaType = strings.TrimPrefix(parts[0], "data:")
		mediaBytes, err = base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			http.Error(w, "Invalid media encoding", http.StatusBadRequest)
			return
		}
	}

	// Insert post into the database
	result, err := sqlite.DB.Exec(
		`INSERT INTO posts (title, content, media, media_type, privacy, author, group_id, created_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		postInput.Title,
		postInput.Content,
		mediaBytes,
		mediaType,
		postInput.Privacy,
		userID,
		postInput.GroupID,
		time.Now(),
	)
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		log.Printf("create post error: %v", err)
		return
	}

	postID, _ := result.LastInsertId()

	// Return the created post
	response := m.PostResponse{
		ID: postID,
		Title: postInput.Title,
		Content: postInput.Content,
		MediaBase64: postInput.Media,
		MediaType: mediaType,
		Privacy: postInput.Privacy,
		Author: int64(userID),
		GroupID: postInput.GroupID,
		CreatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}


func ViewPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var post struct {
		ID        int64          `json:"id"`
		Title     string         `json:"title"`
		Content   string         `json:"content"`
		Media     []byte         `json:"media,omitempty"`
		MediaType string         `json:"media_type,omitempty"`
		Privacy   int            `json:"privacy"`
		Author    int64          `json:"author"`
		CreatedAt time.Time      `json:"created_at"`
	}
	var authorName string
	var avatar sql.NullString

	err = sqlite.DB.QueryRow(`
		SELECT p.id, p.title, p.content, p.media, p.media_type, p.privacy, p.author, p.created_at,
			   u.username, u.avatar
		FROM posts p
		JOIN users u ON p.author = u.id
			WHERE p.id = ?`, 
		id,
	).Scan(
		&post.ID, &post.Title, &post.Content, &post.Media, &post.MediaType,
		&post.Privacy, &post.Author, &post.CreatedAt, &authorName, &avatar,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post does not exist", http.StatusNotFound)
			return
		}
		http.Error(w, "Error fetching post", http.StatusInternalServerError)
		log.Printf("view post: %v", err)
		return
	}

	response := m.PostResponse{
		ID:        post.ID,
		Title:     post.Title,
		Content:   post.Content,
		Privacy:   post.Privacy,
		Author:    post.Author,
		CreatedAt: post.CreatedAt,
		AuthorName: authorName,
	}

	if len(post.Media) > 0 {
		response.MediaBase64 = "data:" + post.MediaType + ";base64," + 
			base64.StdEncoding.EncodeToString(post.Media)
		response.MediaType = post.MediaType
	}

	if avatar.Valid {
		response.AuthorAvatar = avatar.String
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
        SELECT p.id, p.title, p.content, p.media, p.media_type, p.privacy, p.author, p.created_at, p.group_id,
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

    var posts []m.PostResponse
    for rows.Next() {
        var post struct {
            ID        int64
            Title     string
            Content   string
            Media     []byte
            MediaType sql.NullString
            Privacy   int
            Author    int64
            CreatedAt time.Time
            GroupID   sql.NullInt64
            Username  string
            Avatar    sql.NullString
        }

        if err := rows.Scan(
            &post.ID, &post.Title, &post.Content, &post.Media, &post.MediaType,
            &post.Privacy, &post.Author, &post.CreatedAt, &post.GroupID,
            &post.Username, &post.Avatar,
        ); err != nil {
            http.Error(w, "Error reading posts", http.StatusInternalServerError)
            log.Printf("scan post: %v", err)
            return
        }

        response := m.PostResponse{
            ID:         post.ID,
            Title:      post.Title,
            Content:    post.Content,
            Privacy:    post.Privacy,
            Author:     post.Author,
            CreatedAt:  post.CreatedAt,
            AuthorName: post.Username,
        }

        if len(post.Media) > 0 && post.MediaType.Valid {
            response.MediaBase64 = "data:" + post.MediaType.String + ";base64," + 
                base64.StdEncoding.EncodeToString(post.Media)
            response.MediaType = post.MediaType.String
        }

        if post.GroupID.Valid {
            response.GroupID = &post.GroupID.Int64
        }

        if post.Avatar.Valid {
            response.AuthorAvatar = post.Avatar.String
        }

        posts = append(posts, response)
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
        SELECT p.id, p.title, p.content, p.media, p.media_type, p.privacy, p.author, p.created_at,
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

    var posts []m.PostResponse
    for rows.Next() {
        var post struct {
            ID        int64
            Title     string
            Content   string
            Media     []byte
            MediaType sql.NullString
            Privacy   int
            Author    int64
            CreatedAt time.Time
            Username  string
            Avatar    sql.NullString
        }

        if err := rows.Scan(
            &post.ID, &post.Title, &post.Content, &post.Media, &post.MediaType,
            &post.Privacy, &post.Author, &post.CreatedAt, &post.Username, &post.Avatar,
        ); err != nil {
            http.Error(w, "Error scanning posts", http.StatusInternalServerError)
            return
        }

        response := m.PostResponse{
            ID:         post.ID,
            Title:      post.Title,
            Content:    post.Content,
            Privacy:    post.Privacy,
            Author:     post.Author,
            CreatedAt:  post.CreatedAt,
            AuthorName: post.Username,
        }

        // Handle media
        if len(post.Media) > 0 && post.MediaType.Valid {
            response.MediaBase64 = "data:" + post.MediaType.String + ";base64," + 
                base64.StdEncoding.EncodeToString(post.Media)
            response.MediaType = post.MediaType.String
        }

        if post.Avatar.Valid {
            response.AuthorAvatar = post.Avatar.String
        }

        posts = append(posts, response)
    }

    if err = rows.Err(); err != nil {
        http.Error(w, "Error iterating posts", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(posts)
}


