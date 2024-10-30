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
	if post.Privay != 1 && post.Privay != 2 && post.Privay != 3 {
		http.Error(w, "invalid privacy type", http.StatusBadRequest)
		return
	}

	if _, err := sqlite.DB.Exec("INSERT INTO posts (title, content, media, privacy, author, group_id) VALUES (?, ?, ?, ?, ?, ?)", post.Title, post.Content, post.Media, post.Privay, post.Author, nil); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("create post: %v", err)
		return
	}

	w.Write([]byte("Post created successfully"))
}

// the handler that contains the logic for viewing the post 
// ! need to add the logic to private
func ViewPost(w http.ResponseWriter, r *http.Request) {

	// get the id from the path
	idString := r.PathValue("id")

	// convert the id into number
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Invalid number", http.StatusBadRequest)
		return
	}

	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized user", http.StatusUnauthorized)
		return
	}

	var userID uint64
	if err := sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID); err != nil {
		http.Error(w, "Unauthorized user", http.StatusUnauthorized)
		return
	}

	var post m.Post
	if err := sqlite.DB.QueryRow("SELECT id, title, content, media, privacy, author, created_at FROM posts WHERE id = ?", id).Scan(&post.ID, &post.Title, &post.Content, &post.Media, &post.Privay, &post.Author, &post.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post does not exists", http.StatusBadRequest)
			return
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}

	if post.Privay == 1 { // ? if it's public 
		
		if err := json.NewEncoder(w).Encode(&post); err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

	} else if post.Privay == 2 { //? if it's partial private 
		query := "SELECT follower_id FROM followers WHERE followed_id = ? AND status = 'accept' AND follower_id = ?"
		var followerID uint64
		if err := sqlite.DB.QueryRow(query, post.Author, userID).Scan(&followerID); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Post is not visible to you", http.StatusForbidden)
				return
			}
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			log.Printf("Error: %v", err)
			return
		}

		if err := json.NewEncoder(w).Encode(&post); err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
		
	} else if post.Privay == 3 {  //? if it's partial private
		// ldk how it would look in front-end 
	}
}


func GetPosts(w http.ResponseWriter, r *http.Request) {
    var posts []m.Post

    row, err := sqlite.DB.Query("SELECT id, title, content, media, privacy, author, created_at FROM posts WHERE privacy = 1 AND group_id IS NULL")
    if err != nil {
        http.Error(w, "Something went wrong", http.StatusInternalServerError)
        log.Printf("Error: %v", err)
        return
    }
    defer row.Close() 

	// go through all the posts
    for row.Next() {
        var post m.Post

		// get individual post and copy the values into the variable
        if err := row.Scan(&post.ID, &post.Title, &post.Content, &post.Media, &post.Privay, &post.Author, &post.CreatedAt); err != nil {
            http.Error(w, "Error getting post", http.StatusInternalServerError)
            log.Printf("Error scanning: %v", err)
            return
        }

		// append the post to the slice
        posts = append(posts, post)
    }

		// get individual post and copy the values into the variable
		if err := row.Err(); err != nil {
        http.Error(w, "Something went wrong", http.StatusInternalServerError)
        log.Printf("Row iteration error: %v", err)
        return
    }

    w.Header().Set("Content-Type", "application/json")
	// send the array of posts to the frontend
    if err := json.NewEncoder(w).Encode(posts); err != nil {
        http.Error(w, "Something went wrong", http.StatusInternalServerError)
        log.Printf("Error encoding response: %v", err)
        return
    }
}
