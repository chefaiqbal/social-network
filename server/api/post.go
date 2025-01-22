package api

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io"
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
	var postInput struct {
			Title          string   `json:"title"`
			Content        string   `json:"content"`
			Media          string   `json:"media"`      // Base64 string from frontend
			Privacy        int      `json:"privacy"`
			GroupID        *int64   `json:"group_id,omitempty"`
			SelectedFriends []string `json:"selectedFriends"`
	}

	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	if err := json.NewDecoder(r.Body).Decode(&postInput); err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, "Error reading data", http.StatusBadRequest)
		return
	}

    const (
        MAX_TITLE_LENGTH   = 200
        MAX_CONTENT_LENGTH = 1000
    )

    if len(strings.TrimSpace(postInput.Title)) > MAX_TITLE_LENGTH {
        http.Error(w, "Title exceeds maximum length of 200 characters", http.StatusBadRequest)
        return
    }

    if len(strings.TrimSpace(postInput.Content)) > MAX_CONTENT_LENGTH {
        http.Error(w, "Content exceeds maximum length of 1000 characters", http.StatusBadRequest)
        return
    }

    if strings.TrimSpace(postInput.Title) == "" && 
       strings.TrimSpace(postInput.Content) == "" && 
       postInput.Media == "" {
        http.Error(w, "Post must have either title, content, or media", http.StatusBadRequest)
        return
    }

	userID, err := util.GetUserID(r, w)
	if err != nil {
		log.Printf("Auth error: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if postInput.Privacy != 1 && postInput.Privacy != 2 && postInput.Privacy != 3 {
		log.Printf("Invalid privacy value: %d", postInput.Privacy)
		http.Error(w, "Invalid privacy type", http.StatusBadRequest)
		return
	}

	var mediaBytes []byte
	var mediaType string

	// Process media if provided
	if postInput.Media != "" {
		parts := strings.Split(postInput.Media, ";base64,")
		if len(parts) != 2 {
			log.Printf("Invalid media format: %s", postInput.Media[:100]) 
			http.Error(w, "Invalid media format", http.StatusBadRequest)
			return
		}

		mediaType = strings.TrimPrefix(parts[0], "data:")
		if strings.Contains(mediaType, "gif") {
			mediaType = "image/gif"
		}

		mediaBytes, err = base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			log.Printf("Base64 decode error: %v", err)
			http.Error(w, "Invalid media encoding", http.StatusBadRequest)
			return
		}
	}

	// If this is a close friends post, store the friends list
	if postInput.Privacy == 3 && len(postInput.SelectedFriends) > 0 {
		// First, delete existing close friends for this user
		_, err = sqlite.DB.Exec(`
			DELETE FROM post_PrivateViews WHERE user_id = ?`,
			userID)
		if err != nil {
			log.Printf("Error deleting existing close friends: %v", err)
			http.Error(w, "Error updating close friends list", http.StatusInternalServerError)
			return
		}

		// Then insert the new list
		friendsList := strings.Join(postInput.SelectedFriends, ",")
		_, err = sqlite.DB.Exec(`
			INSERT INTO post_PrivateViews (user_id, close_friends)
			VALUES (?, ?)`,
			userID, friendsList)
		if err != nil {
			log.Printf("Error inserting close friends: %v", err)
			http.Error(w, "Error saving close friends list", http.StatusInternalServerError)
			return
		}
	}

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
		log.Printf("Database insert error: %v", err)
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	postID, _ := result.LastInsertId()

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

	currentUserID, err := util.GetUserID(r, w)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Query to fetch the post and author details
    var post struct {
        ID        int64     `json:"id"`
        Title     string    `json:"title"`
        Content   string    `json:"content"`
        Media     []byte    `json:"media,omitempty"`
        MediaType string    `json:"media_type,omitempty"`
        Privacy   int       `json:"privacy"`
        Author    int64     `json:"author"`
        CreatedAt time.Time `json:"created_at"`
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
        log.Printf("Error fetching post: %v", err)
        return
    }

    if int64(currentUserID) != post.Author { //for privacey valdition
        switch post.Privacy {
        case 2: 
            var isFollower bool
            err = sqlite.DB.QueryRow(`
                SELECT EXISTS (
                    SELECT 1 FROM followers
                    WHERE follower_id = ? AND followed_id = ?  AND status = 'accept'
                )`,
                currentUserID, post.Author,
            ).Scan(&isFollower)
            if err != nil {
                http.Error(w, "Error checking follower status", http.StatusInternalServerError)
                log.Printf("Error checking follower status: %v", err)
                return
            }
            if !isFollower {
                http.Error(w, "You must follow the author to view this post.", http.StatusForbidden)
                return
            }
        case 3: 
            var currentUserUsername string
            err = sqlite.DB.QueryRow(`
                SELECT username FROM users WHERE id = ?`,
                currentUserID,
            ).Scan(&currentUserUsername)
            if err != nil {
                http.Error(w, "Error fetching current user's username", http.StatusInternalServerError)
                log.Printf("Error fetching current user's username: %v", err)
                return
            }

            var closeFriends sql.NullString
            err = sqlite.DB.QueryRow(`
                SELECT close_friends FROM post_PrivateViews WHERE user_id = ?`,
                post.Author,
            ).Scan(&closeFriends)
            if err != nil {
                if err == sql.ErrNoRows {
                    http.Error(w, "You are not in the author's close friends list.", http.StatusForbidden)
                    return
                }
                http.Error(w, "Error checking close friends list", http.StatusInternalServerError)
                log.Printf("Error checking close friends list: %v", err)
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
                if friend == currentUserUsername {
                    isCloseFriend = true
                    break
                }
            }

            if !isCloseFriend {
                http.Error(w, "You are not in the author's close friends list.", http.StatusForbidden)
                return
            }
        }
    }

    response := m.PostResponse{
        ID:         post.ID,
        Title:      post.Title,
        Content:    post.Content,
        Privacy:    post.Privacy,
        Author:     post.Author,
        CreatedAt:  post.CreatedAt,
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
    cookie, _ := r.Cookie("AccessToken")
    userID := util.UserSession[cookie.Value]

    // Get current user's username
    var currentUsername string
    err := sqlite.DB.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&currentUsername)
    if err != nil {
        http.Error(w, "Error getting current user", http.StatusInternalServerError)
        return
    }

    rows, err := sqlite.DB.Query(`
        SELECT 
            p.id, 
            p.title, 
            p.content, 
            p.media, 
            p.media_type,
            p.privacy, 
            p.author, 
            p.created_at,
            u.username as author_name,
            u.avatar as author_avatar,
            (SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = true) as like_count,
            EXISTS(SELECT 1 FROM likes WHERE post_id = p.id AND user_id = ? AND is_like = true) as user_liked,
            cf.close_friends,
            p.group_id
        FROM posts p
        JOIN users u ON p.author = u.id
        LEFT JOIN post_PrivateViews cf ON p.author = cf.user_id
        ORDER BY p.created_at DESC
    `, userID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var posts []m.PostResponse
    for rows.Next() {
        var post struct {
            ID            int64
            Title         string
            Content       string
            Media         []byte
            MediaType     sql.NullString
            Privacy       int
            Author        int64
            CreatedAt     time.Time
            GroupID       sql.NullInt64
            Username      string
            Avatar        sql.NullString
            CloseFriends  sql.NullString
            LikeCount     int
            UserLiked     sql.NullBool
        }

        if err := rows.Scan(
            &post.ID, &post.Title, &post.Content, &post.Media, &post.MediaType,
            &post.Privacy, &post.Author, &post.CreatedAt, &post.Username, &post.Avatar,
            &post.LikeCount, &post.UserLiked, &post.CloseFriends, &post.GroupID,
        ); err != nil {
            http.Error(w, "Error reading posts", http.StatusInternalServerError)
            log.Printf("Error scanning posts: %v", err)
            return
        }

        // Check privacy settings
        if post.Privacy == 3 && post.Author != int64(userID) {
            if post.CloseFriends.Valid {
                closeFriends := strings.Split(post.CloseFriends.String, ",")
                found := false
                for _, friend := range closeFriends {
                    if strings.TrimSpace(friend) == currentUsername {
                        found = true
                        break
                    }
                }
                if !found {
                    continue 
                }
            } else {
                continue
            }
        }

        if post.Privacy == 2 && post.Author != int64(userID) {
            var isFollower bool
            err := sqlite.DB.QueryRow(`
                SELECT EXISTS (
                    SELECT 1 FROM followers WHERE follower_id = ? AND followed_id = ? AND status = 'accept'
                )`, userID, post.Author).Scan(&isFollower)
            if err != nil || !isFollower {
                continue 
            }
        }

        response := m.PostResponse{
            ID:         post.ID,
            Title:      post.Title,
            Content:    post.Content,
            Privacy:    post.Privacy,
            Author:     post.Author,
            CreatedAt:  post.CreatedAt,
            AuthorName: post.Username,
            LikeCount:  post.LikeCount,
            UserLiked:  post.UserLiked.Valid && post.UserLiked.Bool,
        }

        // Handle media with proper type preservation
        if len(post.Media) > 0 && post.MediaType.Valid {
            mediaType := post.MediaType.String
            response.MediaBase64 = "data:" + mediaType + ";base64," +
                base64.StdEncoding.EncodeToString(post.Media)
            response.MediaType = mediaType
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
        log.Printf("Error iterating posts: %v", err)
        return
    }

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


