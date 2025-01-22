package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"social-network/util"

	//"social-network/models"
	"social-network/pkg/db/sqlite"
)
func UserProfile(w http.ResponseWriter, r *http.Request) {
    // Get the user ID from the URL parameter
    userIDStr := r.PathValue("userID")
    var targetUserID int64
    var err error

    // Check if we're requesting the current user's profile
    if userIDStr == "current" {
        currentUserID, err := util.GetUserID(r, w)
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        targetUserID = int64(currentUserID)
    } else {
        targetUserID, err = strconv.ParseInt(userIDStr, 10, 64)
        if err != nil {
            http.Error(w, "Invalid user ID", http.StatusBadRequest)
            return
        }
    }

    // Get the logged-in user's ID to check follow status
    loggedInUserID, err := util.GetUserID(r, w)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Query user profile information
    query := `
        SELECT 
            u.id, u.username, u.first_name, u.last_name, u.email, 
            u.date_of_birth, u.about_me, u.avatar, u.is_private, 
            u.nickname,
            CASE WHEN f.status = 'accept' THEN true ELSE false END as is_following,
            CASE WHEN f.status = 'pending' THEN true ELSE false END as is_pending,
            u.created_at
        FROM users u
        LEFT JOIN followers f ON f.followed_id = u.id AND f.follower_id = ?
        WHERE u.id = ?
    `

    var profile struct {
        ID          int64  `json:"id"`
        Username    string `json:"username"`
        FirstName   string `json:"first_name"`
        LastName    string `json:"last_name"`
        Email       string `json:"email"`
        DateOfBirth string `json:"date_of_birth"`
        AboutMe     string `json:"about_me"`
        Avatar      string `json:"avatar"`
        IsPrivate   bool   `json:"is_private"`
        IsFollowing bool   `json:"is_following"`
        IsPending   bool   `json:"is_pending"`
        CreatedAt   string `json:"created_at"`
        Posts       []Post `json:"posts,omitempty"`
        Followers   int    `json:"followers_count"`
        Following   int    `json:"following_count"`
        Nickname    string `json:"nickname"`
    }

    // Get profile details
    var avatar sql.NullString
    var aboutMe sql.NullString
    var nickname sql.NullString
    err = sqlite.DB.QueryRow(query, loggedInUserID, targetUserID).Scan(
        &profile.ID, &profile.Username, &profile.FirstName, &profile.LastName,
        &profile.Email, &profile.DateOfBirth, &aboutMe, &avatar,
        &profile.IsPrivate, &nickname, &profile.IsFollowing, &profile.IsPending,
        &profile.CreatedAt,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "User not found", http.StatusNotFound)
            return
        }
        http.Error(w, "Database error", http.StatusInternalServerError)
        log.Printf("Database error: %v", err)
        return
    }

    profile.AboutMe = aboutMe.String
    profile.Avatar = avatar.String
    profile.Nickname = nickname.String

    // Privacy check for profile
    if profile.IsPrivate && !profile.IsFollowing && profile.ID != int64(loggedInUserID) {
        http.Error(w, "User profile is private", http.StatusForbidden)
        return
    }

    rows, err := sqlite.DB.Query(`
        SELECT id, title, content, created_at, privacy, author
        FROM posts
        WHERE author = ?
        ORDER BY created_at DESC
        LIMIT 10
    `, targetUserID)
    if err != nil {
        log.Printf("Error fetching posts: %v", err)
    } else {
        defer rows.Close()
        for rows.Next() {
            var post Post
            var privacy int
            var authorID int64
            err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &privacy, &authorID)
            if err != nil {
                log.Printf("Error scanning post: %v", err)
                continue
            }

            if int64(loggedInUserID) != authorID {
                switch privacy {
                case 1: 
                case 2: 
                    var isFollower bool
                    err = sqlite.DB.QueryRow(`
                        SELECT EXISTS (
                            SELECT 1 FROM followers
                            WHERE follower_id = ? AND followed_id = ? AND status = 'accept'
                        )
                    `, loggedInUserID, authorID).Scan(&isFollower)
                    if err != nil || !isFollower {
                        continue
                    }
                case 3: 
					var userName string
					err = sqlite.DB.QueryRow(`SELECT username FROM users WHERE id = ? `, loggedInUserID).Scan(&userName)
					if err != nil {
						if err == sql.ErrNoRows {
							continue
						}
						log.Printf("Error checking close friends list: %v", err)
						continue
					}

                    var closeFriends sql.NullString
                    err = sqlite.DB.QueryRow(`
                        SELECT close_friends FROM post_PrivateViews WHERE user_id = ?`, authorID,
                    ).Scan(&closeFriends)
                    if err != nil {
                        if err == sql.ErrNoRows || !closeFriends.Valid {
                            continue
                        }
                        log.Printf("Error checking close friends list: %v", err)
                        continue
                    }

                    closeFriendsList := strings.Split(closeFriends.String, ",")
                    isCloseFriend := false
                    for _, friend := range closeFriendsList {
                        if friend ==  userName{
                            isCloseFriend = true
                            break
                        }
                    }

                    if !isCloseFriend {
                        continue 
                    }
                default:
                    continue 
                }
            }

            profile.Posts = append(profile.Posts, post)
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(profile)
}


// Add this struct for the posts
type Post struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	CreatedAt     time.Time `json:"created_at"`
	LikesCount    int       `json:"likes_count"`
	CommentsCount int       `json:"comments_count"`
}

func GetSuggestedUsers(w http.ResponseWriter, r *http.Request) {
	// Get current user ID from session
	userID, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Query for users that the current user is not following
	rows, err := sqlite.DB.Query(`
		SELECT DISTINCT u.id, u.username, u.avatar 
		FROM users u 
		WHERE u.id NOT IN (
			SELECT followed_id 
			FROM followers 
			WHERE follower_id = ? AND status = 'accept'
		) 
		AND u.id != ?
		AND u.id IN (
			SELECT id FROM users WHERE id != 1
		)
		LIMIT 5
	`, userID, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Error querying suggested users: %v", err)
		return
	}
	defer rows.Close()

	var suggestedUsers []struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Avatar   string `json:"avatar,omitempty"`
		Online   bool   `json:"online"`
	}

	// Get current online users
	onlineUsers := make(map[uint64]bool)
	for _, id := range GetOnlineUsers(socketManager) {
		onlineUsers[id] = true
	}

	for rows.Next() {
		var user struct {
			ID       uint   `json:"id"`
			Username string `json:"username"`
			Avatar   string `json:"avatar,omitempty"`
			Online   bool   `json:"online"`
		}
		var avatar sql.NullString
		if err := rows.Scan(&user.ID, &user.Username, &avatar); err != nil {
			http.Error(w, "Error scanning users", http.StatusInternalServerError)
			log.Printf("Error scanning user row: %v", err)
			return
		}
		if avatar.Valid {
			user.Avatar = avatar.String
		}
		// Check if user is online
		user.Online = onlineUsers[uint64(user.ID)]
		suggestedUsers = append(suggestedUsers, user)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, "Error iterating users", http.StatusInternalServerError)
		log.Printf("Error iterating users: %v", err)
		return
	}

	// If no suggested users found, return empty array instead of null
	if suggestedUsers == nil {
		suggestedUsers = make([]struct {
			ID       uint   `json:"id"`
			Username string `json:"username"`
			Avatar   string `json:"avatar,omitempty"`
			Online   bool   `json:"online"`
		}, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(suggestedUsers); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	// Get current user ID from session
	userID, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Modified query to show all users except:
	// 1. The current user
	// 2. Show follow status for each user
	query := `
		SELECT DISTINCT 
			u.id, 
			u.username, 
			u.avatar, 
			u.is_private, 
			u.about_me, 
			u.first_name, 
			u.last_name,
			CASE 
				WHEN f.status = 'accept' THEN true
				ELSE false
			END as is_following,
			CASE 
				WHEN f.status = 'pending' THEN true
				ELSE false
			END as is_pending,
			CASE 
				WHEN EXISTS (
					SELECT 1 FROM followers f2 
					WHERE f2.follower_id = u.id 
					AND f2.followed_id = ? 
					AND f2.status = 'accept'
				) THEN true
				ELSE false
			END as follows_you
		FROM users u 
		LEFT JOIN followers f ON (f.follower_id = ? AND f.followed_id = u.id)
		WHERE u.id != ? 
		ORDER BY 
			CASE WHEN f.status = 'pending' THEN 0 ELSE 1 END,
			u.username
	`
	rows, err := sqlite.DB.Query(query, userID, userID, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Error querying users: %v", err)
		return
	}
	defer rows.Close()

	type User struct {
		ID          uint   `json:"id"`
		Username    string `json:"username"`
		Avatar      string `json:"avatar,omitempty"`
		IsPrivate   bool   `json:"is_private"`
		AboutMe     string `json:"about_me"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		IsFollowing bool   `json:"is_following"`
		IsPending   bool   `json:"is_pending"`
		FollowsYou  bool   `json:"follows_you"`
	}
	var users []User

	for rows.Next() {
		var user User
		var avatar, aboutMe sql.NullString
		if err := rows.Scan(
			&user.ID, 
			&user.Username, 
			&avatar, 
			&user.IsPrivate, 
			&aboutMe, 
			&user.FirstName, 
			&user.LastName, 
			&user.IsFollowing,
			&user.IsPending,
			&user.FollowsYou,
		); err != nil {
			http.Error(w, "Error scanning users", http.StatusInternalServerError)
			log.Printf("Error scanning user row: %v", err)
			return
		}
		if avatar.Valid {
			user.Avatar = avatar.String
		}
		if aboutMe.Valid {
			user.AboutMe = aboutMe.String
		}
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// Add a new endpoint to update privacy settings
func UpdatePrivacySettings(w http.ResponseWriter, r *http.Request) {
	userID, err := util.GetUserID(r, w)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var settings struct {
		IsPrivate bool `json:"is_private"`
	}

	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err = sqlite.DB.Exec("UPDATE users SET is_private = ? WHERE id = ?", settings.IsPrivate, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Error updating privacy settings: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"is_private": settings.IsPrivate})
}

func GetUername (r *http.Request, w http.ResponseWriter) {
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"username": username})
}