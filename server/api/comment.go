package api

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"bytes"
	//"time"
	"encoding/base64"
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

    var commentInput struct {
        Content string `json:"content"`
        Media   string `json:"media"` // Base64 string from frontend
        PostID  int    `json:"post_id"`
        GroupID int    `json:"group_id"`
    }

    // Log the raw request body for debugging
    body, _ := io.ReadAll(r.Body)
    r.Body = io.NopCloser(bytes.NewBuffer(body))
    log.Printf("Raw request body: %s", string(body))

    if err := json.NewDecoder(r.Body).Decode(&commentInput); err != nil {
        log.Printf("JSON decode error: %v", err)
        http.Error(w, "Invalid JSON data", http.StatusBadRequest)
        return
    }

    // Check if user is member or creator of the group
    var isMemberOrCreator bool
    err = sqlite.DB.QueryRow(`
        SELECT EXISTS (
            SELECT 1 FROM group_members 
            WHERE group_id = ? 
            AND user_id = ? 
            AND (status = 'member' OR status = 'creator')
        )`,
        commentInput.GroupID,
        currentUserID,
    ).Scan(&isMemberOrCreator)

    if err != nil {
        log.Printf("Database error checking membership: %v", err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    if !isMemberOrCreator {
        http.Error(w, "Only group members can comment on posts", http.StatusForbidden)
        return
    }

    // Log the parsed input
    log.Printf("Parsed comment input: %+v", commentInput)

    if strings.TrimSpace(commentInput.Content) == "" && commentInput.Media == "" {
        http.Error(w, "Comment must have either content or media", http.StatusBadRequest)
        return
    }

    var mediaBytes []byte
    var mediaType string

    // Process media if provided
    if commentInput.Media != "" {
        parts := strings.Split(commentInput.Media, ";base64,")
        if len(parts) != 2 {
            log.Printf("Invalid media format: %s", commentInput.Media[:100])
            http.Error(w, "Invalid media format", http.StatusBadRequest)
            return
        }

        mediaType = strings.TrimPrefix(parts[0], "data:")
        mediaBytes, err = base64.StdEncoding.DecodeString(parts[1])
        if err != nil {
            log.Printf("Base64 decode error: %v", err)
            http.Error(w, "Invalid media encoding", http.StatusBadRequest)
            return
        }
    }

    // Log the values before insert
    log.Printf("Inserting comment: content=%s, mediaType=%s, author=%d, postID=%d",
        commentInput.Content, mediaType, currentUserID, commentInput.PostID)

    // Insert comment with media
    result, err := sqlite.DB.Exec(
        `INSERT INTO comments (content, media, media_type, author, post_id, created_at) 
         VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
        commentInput.Content,
        mediaBytes,
        mediaType,
        currentUserID,
        commentInput.PostID,
    )
    if err != nil {
        log.Printf("Database insert error: %v", err)
        http.Error(w, "Failed to create comment", http.StatusInternalServerError)
        return
    }

    commentID, _ := result.LastInsertId()

    // Fetch the complete comment data including author information
    var comment m.CommentResponse
    err = sqlite.DB.QueryRow(`
        SELECT 
            c.id, c.content, c.media, c.media_type, c.post_id, c.author, c.created_at,
            u.username as author_name, u.avatar as author_avatar
        FROM comments c
        JOIN users u ON c.author = u.id
        WHERE c.id = ?`,
        commentID,
    ).Scan(
        &comment.ID,
        &comment.Content,
        &mediaBytes,
        &mediaType,
        &comment.PostID,
        &comment.Author,
        &comment.CreatedAt,
        &comment.AuthorName,
        &comment.AuthorAvatar,
    )

    if err != nil {
        log.Printf("Error fetching created comment: %v", err)
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "message": "Comment created",
            "id":     commentID,
        })
        return
    }

    // Convert media bytes to base64 if present
    if len(mediaBytes) > 0 {
        comment.MediaBase64 = "data:" + mediaType + ";base64," +
            base64.StdEncoding.EncodeToString(mediaBytes)
        comment.MediaType = mediaType
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(comment)
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
			c.media_type,
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

	var comments []m.CommentResponse
	for rows.Next() {
		var comment m.CommentResponse
		var mediaBytes []byte
		var mediaType sql.NullString
		var avatar sql.NullString

		if err := rows.Scan(
			&comment.ID,
			&comment.Content,
			&mediaBytes,
			&mediaType,
			&comment.PostID,
			&comment.Author,
			&comment.CreatedAt,
			&comment.AuthorName,
			&avatar,
		); err != nil {
			log.Printf("Error scanning comment: %v", err)
			continue
		}

		// Convert media bytes to base64 if present
		if len(mediaBytes) > 0 && mediaType.Valid {
			comment.MediaBase64 = "data:" + mediaType.String + ";base64," + 
				base64.StdEncoding.EncodeToString(mediaBytes)
			comment.MediaType = mediaType.String
		}

		if avatar.Valid {
			comment.AuthorAvatar = avatar.String
		}

		comments = append(comments, comment)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
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

func CreateGroupPostComment(w http.ResponseWriter, r *http.Request) {
    currentUserID, err := util.GetUserID(r, w)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Get groupId and postId from URL parameters
    groupIDStr := r.PathValue("groupId")
    postIDStr := r.PathValue("postId")

    groupID, err := strconv.Atoi(groupIDStr)
    if err != nil {
        http.Error(w, "Invalid group ID", http.StatusBadRequest)
        return
    }

    postID, err := strconv.Atoi(postIDStr)
    if err != nil {
        http.Error(w, "Invalid post ID", http.StatusBadRequest)
        return
    }

    var commentInput struct {
        Content string `json:"content"`
        Media   string `json:"media"`
    }

    if err := json.NewDecoder(r.Body).Decode(&commentInput); err != nil {
        http.Error(w, "Invalid JSON data", http.StatusBadRequest)
        return
    }

    // Verify user is a member of the group
    var isMember bool
    err = sqlite.DB.QueryRow(`
        SELECT EXISTS (
            SELECT 1 FROM group_members 
            WHERE group_id = ? AND user_id = ? 
            AND (status = 'member' OR status = 'creator')
        )`,
        groupID,
        currentUserID,
    ).Scan(&isMember)

    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    if !isMember {
        http.Error(w, "Only group members can comment", http.StatusForbidden)
        return
    }

    var mediaBytes []byte
    var mediaType string

    if commentInput.Media != "" {
        parts := strings.Split(commentInput.Media, ";base64,")
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

    // Insert into group_post_comments table
    result, err := sqlite.DB.Exec(
        `INSERT INTO group_post_comments (content, media, media_type, post_id, group_id, author, created_at) 
         VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
        commentInput.Content,
        mediaBytes,
        mediaType,
        postID,
        groupID,
        currentUserID,
    )
    if err != nil {
        log.Printf("Error inserting comment: %v", err)
        http.Error(w, "Failed to create comment", http.StatusInternalServerError)
        return
    }

    commentID, _ := result.LastInsertId()

    // Fetch the created comment with user information
    var comment m.CommentResponse
    err = sqlite.DB.QueryRow(`
        SELECT 
            c.id, c.content, c.media, c.media_type, c.post_id, c.author, c.created_at,
            u.username as author_name, u.avatar as author_avatar
        FROM group_post_comments c
        JOIN users u ON c.author = u.id
        WHERE c.id = ?`,
        commentID,
    ).Scan(
        &comment.ID,
        &comment.Content,
        &mediaBytes,
        &mediaType,
        &comment.PostID,
        &comment.Author,
        &comment.CreatedAt,
        &comment.AuthorName,
        &comment.AuthorAvatar,
    )

    if err != nil {
        log.Printf("Error fetching created comment: %v", err)
        http.Error(w, "Error fetching created comment", http.StatusInternalServerError)
        return
    }

    if len(mediaBytes) > 0 {
        comment.MediaBase64 = "data:" + mediaType + ";base64," +
            base64.StdEncoding.EncodeToString(mediaBytes)
        comment.MediaType = mediaType
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(comment)
}

func GetGroupPostComments(w http.ResponseWriter, r *http.Request) {
    postIDStr := r.PathValue("postId")
    groupIDStr := r.PathValue("groupId")

    postID, err := strconv.Atoi(postIDStr)
    if err != nil {
        log.Printf("Invalid post ID: %v", err)
        http.Error(w, "Invalid post ID", http.StatusBadRequest)
        return
    }

    groupID, err := strconv.Atoi(groupIDStr)
    if err != nil {
        log.Printf("Invalid group ID: %v", err)
        http.Error(w, "Invalid group ID", http.StatusBadRequest)
        return
    }

    log.Printf("Fetching comments for group %d and post %d", groupID, postID)

    query := `
        SELECT 
            c.id,
            c.content,
            c.media,
            c.media_type,
            c.post_id,
            c.author,
            c.created_at,
            u.username as author_name,
            u.avatar as author_avatar
        FROM group_post_comments c
        JOIN users u ON c.author = u.id
        WHERE c.post_id = ? AND c.group_id = ?
        ORDER BY c.created_at DESC
    `

    rows, err := sqlite.DB.Query(query, postID, groupID)
    if err != nil {
        log.Printf("Database error: %v", err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var comments []m.CommentResponse
    for rows.Next() {
        var comment m.CommentResponse
        var mediaBytes []byte
        var mediaType sql.NullString
        var avatar sql.NullString

        err := rows.Scan(
            &comment.ID,
            &comment.Content,
            &mediaBytes,
            &mediaType,
            &comment.PostID,
            &comment.Author,
            &comment.CreatedAt,
            &comment.AuthorName,
            &avatar,
        )
        if err != nil {
            log.Printf("Error scanning comment: %v", err)
            continue
        }

        if len(mediaBytes) > 0 && mediaType.Valid {
            comment.MediaBase64 = "data:" + mediaType.String + ";base64," +
                base64.StdEncoding.EncodeToString(mediaBytes)
            comment.MediaType = mediaType.String
        }

        if avatar.Valid {
            comment.AuthorAvatar = avatar.String
        }

        comments = append(comments, comment)
    }

    log.Printf("Found %d comments", len(comments))

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(comments)
}
