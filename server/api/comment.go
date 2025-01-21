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
	"fmt"
)

const (
	MAX_COMMENT_LENGTH = 500
)

func CreateComment(w http.ResponseWriter, r *http.Request) { // done
    currentUserID, err := util.GetUserID(r, w)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var commentInput struct {
        Content string `json:"content"`
        Media   string `json:"media"` // Base64 string from frontend
        PostID  int    `json:"post_id"`
        GroupID int    `json:"group_id,omitempty"`
    }

    body, _ := io.ReadAll(r.Body)
    r.Body = io.NopCloser(bytes.NewBuffer(body))
    //log.Printf("Raw request body: %s", string(body))

    if err := json.NewDecoder(r.Body).Decode(&commentInput); err != nil {
       // log.Printf("JSON decode error: %v", err)
        http.Error(w, "Invalid JSON data", http.StatusBadRequest)
        return
    }

    var postExists bool
    err = sqlite.DB.QueryRow(`
        SELECT EXISTS (SELECT 1 FROM posts WHERE id = ?)`, commentInput.PostID,
    ).Scan(&postExists)

    if err != nil {
        //log.Printf("Database error checking post existence: %v", err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    if !postExists {
        http.Error(w, "Post not found", http.StatusBadRequest)
        return
    }

    var postPrivacy int
    var postAuthorID uint64
    err = sqlite.DB.QueryRow(`
        SELECT privacy, author FROM posts WHERE id = ?`, commentInput.PostID,
    ).Scan(&postPrivacy, &postAuthorID)
    if err != nil {
        //log.Printf("Database error checking post privacy: %v", err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    if currentUserID != postAuthorID { 
        switch postPrivacy {
        case 1:

        case 2: 
            var isFollower bool
            err = sqlite.DB.QueryRow(`
                SELECT EXISTS (
                    SELECT 1 FROM followers
                    WHERE follower_id = ? AND followed_id = ?
                )`, currentUserID, postAuthorID).Scan(&isFollower)
            if err != nil {
                http.Error(w, "Error checking follower status", http.StatusInternalServerError)
                //log.Printf("Error checking follower status: %v", err)
                return
            }
            if !isFollower {
                http.Error(w, "You must follow the author to comment on this post.", http.StatusForbidden)
                return
            }
        case 3: 
            var currentUserUsername string
            err = sqlite.DB.QueryRow(`
                SELECT username FROM users WHERE id = ?`, currentUserID).Scan(&currentUserUsername)
            if err != nil {
                http.Error(w, "Error fetching current user's username", http.StatusInternalServerError)
                //log.Printf("Error fetching current user's username: %v", err)
                return
            }

            var closeFriends sql.NullString
            err = sqlite.DB.QueryRow(`
                SELECT close_friends FROM post_PrivateViews WHERE user_id = ?`, postAuthorID).Scan(&closeFriends)
            if err != nil {
                if err == sql.ErrNoRows {
                    http.Error(w, "You are not in the author's close friends list.", http.StatusForbidden)
                    return
                }
                http.Error(w, "Error checking close friends list", http.StatusInternalServerError)
                //log.Printf("Error checking close friends list: %v", err)
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
        default:
            http.Error(w, "This post's privacy settings do not allow you to comment", http.StatusForbidden)
            return
        }
    }

    // Check if user is a member or creator of the group (if applicable)
    if commentInput.GroupID != 0 {
        var isMemberOrCreator bool
        err = sqlite.DB.QueryRow(`
            SELECT EXISTS (
                SELECT 1 FROM group_members 
                WHERE group_id = ? 
                AND user_id = ? 
                AND (status = 'member' OR status = 'creator')
            )`, commentInput.GroupID, currentUserID).Scan(&isMemberOrCreator)

        if err != nil {
            //log.Printf("Database error checking membership: %v", err)
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }

        if !isMemberOrCreator {
            http.Error(w, "Only group members can comment on posts", http.StatusForbidden)
            return
        }
    }

    if strings.TrimSpace(commentInput.Content) == "" && commentInput.Media == "" {
        http.Error(w, "Comment must have either content or media", http.StatusBadRequest)
        return
    }

    // Add length validation
    if len(strings.TrimSpace(commentInput.Content)) > MAX_COMMENT_LENGTH {
        http.Error(w, fmt.Sprintf("Comment exceeds maximum length of %d characters", MAX_COMMENT_LENGTH), http.StatusBadRequest)
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
            //log.Printf("Base64 decode error: %v", err)
            http.Error(w, "Invalid media encoding", http.StatusBadRequest)
            return
        }
    }

    // Insert comment with media
    result, err := sqlite.DB.Exec(`
        INSERT INTO comments (content, media, media_type, author, post_id, created_at) 
        VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
        commentInput.Content,
        mediaBytes,
        mediaType,
        currentUserID,
        commentInput.PostID,
    )
    if err != nil {
        //log.Printf("Database insert error: %v", err)
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
        WHERE c.id = ?`, commentID,
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
	currentUserID, err := util.GetUserID(r, w) 
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	postID, err := strconv.Atoi(postIDString)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var postAuthor int
	var postPrivacy int
	err = sqlite.DB.QueryRow(`
		SELECT author, privacy FROM posts WHERE id = ?`, postID,
	).Scan(&postAuthor, &postPrivacy)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post does not exist", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve post information", http.StatusInternalServerError)
		log.Printf("Error querying post privacy: %v", err)
		return
	}

	if int64(currentUserID) != int64(postAuthor) {
		switch postPrivacy {
        case 1:

		case 2: 
			var isFollower bool
			err = sqlite.DB.QueryRow(`
				SELECT EXISTS (
					SELECT 1 FROM followers WHERE follower_id = ? AND followed_id = ?
				)`, currentUserID, postAuthor,
			).Scan(&isFollower)
			if err != nil || !isFollower {
				http.Error(w, "You must follow the author to view this post.", http.StatusForbidden)
				return
			}
		case 3: 
			var currentUserUsername string
			err = sqlite.DB.QueryRow(`
				SELECT username FROM users WHERE id = ?`, currentUserID,
			).Scan(&currentUserUsername)
			if err != nil {
				http.Error(w, "Error fetching current user's username", http.StatusInternalServerError)
				log.Printf("Error fetching current user's username: %v", err)
				return
			}

			var closeFriends sql.NullString
			err = sqlite.DB.QueryRow(`
				SELECT close_friends FROM post_PrivateViews WHERE user_id = ?`, postAuthor,
			).Scan(&closeFriends)
			if err != nil {
				if err == sql.ErrNoRows || !closeFriends.Valid {
					http.Error(w, "You are not in the author's close friends list.", http.StatusForbidden)
					return
				}
				http.Error(w, "Error checking close friends list", http.StatusInternalServerError)
				log.Printf("Error checking close friends list: %v", err)
				return
			}

			closeFriendsList := strings.Split(closeFriends.String, ",")
			isCloseFriend := false
			for _, friend := range closeFriendsList {
				if strings.TrimSpace(friend) == currentUserUsername {
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
		http.Error(w, "Error fetching comments", http.StatusInternalServerError)
		log.Printf("Error querying comments: %v", err)
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
    postIDString := r.PathValue("postID")
    currentUserID, err := util.GetUserID(r, w)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    postID, err := strconv.Atoi(postIDString)
    if err != nil {
        http.Error(w, "Invalid post ID", http.StatusBadRequest)
        return
    }

    var postAuthor int
    var postPrivacy int
    err = sqlite.DB.QueryRow(`
        SELECT author, privacy FROM posts WHERE id = ?`, postID,
    ).Scan(&postAuthor, &postPrivacy)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Post does not exist", http.StatusNotFound)
            return
        }
        http.Error(w, "Failed to retrieve post information", http.StatusInternalServerError)
        log.Printf("Error querying post privacy: %v", err)
        return
    }

    if int64(currentUserID) != int64(postAuthor) {
        switch postPrivacy {
        case 1:
            
        case 2: 
            var isFollower bool
            err = sqlite.DB.QueryRow(`
                SELECT EXISTS (
                    SELECT 1 FROM followers WHERE follower_id = ? AND followed_id = ?
                )`, currentUserID, postAuthor,
            ).Scan(&isFollower)
            if err != nil || !isFollower {
                http.Error(w, "You must follow the author to view this post.", http.StatusForbidden)
                return
            }
        case 3: 
            var currentUserUsername string
            err = sqlite.DB.QueryRow(`
                SELECT username FROM users WHERE id = ?`, currentUserID,
            ).Scan(&currentUserUsername)
            if err != nil {
                http.Error(w, "Error fetching current user's username", http.StatusInternalServerError)
                log.Printf("Error fetching current user's username: %v", err)
                return
            }

            var closeFriends sql.NullString
            err = sqlite.DB.QueryRow(`
                SELECT close_friends FROM post_PrivateViews WHERE user_id = ?`, postAuthor,
            ).Scan(&closeFriends)
            if err != nil {
                if err == sql.ErrNoRows || !closeFriends.Valid {
                    http.Error(w, "You are not in the author's close friends list.", http.StatusForbidden)
                    return
                }
                http.Error(w, "Error checking close friends list", http.StatusInternalServerError)
                log.Printf("Error checking close friends list: %v", err)
                return
            }

            closeFriendsList := strings.Split(closeFriends.String, ",")
            isCloseFriend := false
            for _, friend := range closeFriendsList {
                if strings.TrimSpace(friend) == currentUserUsername {
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

    var count int
    err = sqlite.DB.QueryRow(`
        SELECT COUNT(*) FROM comments WHERE post_id = ?`, postID,
    ).Scan(&count)
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        log.Printf("Error querying comment count: %v", err)
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

    var postExists bool
    err = sqlite.DB.QueryRow(`
        SELECT EXISTS (
            SELECT 1 FROM group_posts 
            WHERE id = ? AND group_id = ?
        )`,
        postID,
        groupID,
    ).Scan(&postExists)
 
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
 
    if !postExists {
        http.Error(w, "Post does not belong to this group", http.StatusBadRequest)
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

    currentUserID, err := util.GetUserID(r, w)
    if err != nil {
        log.Printf("ynable to fetch current user ID: %v", err)
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    membershipQuery := `
    SELECT COUNT(1) 
    FROM group_members 
    WHERE group_id = ? AND user_id = ? AND status IN ('creator', 'member')`

var isMember int
err = sqlite.DB.QueryRow(membershipQuery, groupID, currentUserID).Scan(&isMember)
if err != nil || isMember == 0 {
    log.Printf("User %d is not a member or creator of group %d", currentUserID, groupID)
    http.Error(w, "You need to join the group to view comments", http.StatusForbidden)
    return
}

    postQuery := `SELECT COUNT(1) FROM group_posts WHERE id = ? AND group_id = ?`
    var postExists int
    err = sqlite.DB.QueryRow(postQuery, postID, groupID).Scan(&postExists)
    if err != nil || postExists == 0 {
        log.Printf("Post %d does not exist in group %d", postID, groupID)
        http.Error(w, "Post not found", http.StatusNotFound)
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

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(comments)
}
