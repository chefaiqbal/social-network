package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DateOfBirth string    `json:"date_of_birth"`
	AboutMe     string    `json:"about_me"`
	Avatar      string    `json:"avatar"`
}

func GetUserIDBY(w http.ResponseWriter, r *http.Request) {
    username, err := util.GetUsernameFromSession(r)  
    if err != nil {
        http.Error(w, "Unauthorized: no session cookie", http.StatusUnauthorized)
        return
    }

    var userID uint64
    err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
    if err != nil {
        http.Error(w, "Unauthorized: user not found", http.StatusUnauthorized)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]uint64{"userID": userID})  
}


// func for endpoint :
func GetUserIDED(w http.ResponseWriter, r *http.Request) {

    var requestData struct {
        Username string `json:"username"`
    }
    err := json.NewDecoder(r.Body).Decode(&requestData)
    if err != nil || requestData.Username == "" {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    var userID uint64
    err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", requestData.Username).Scan(&userID)
    if err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    response := map[string]uint64{"userID": userID}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error reading request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Email == "" || req.Password == "" || req.Username == "" {
		http.Error(w, "Email, password and username are required", http.StatusBadRequest)
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error processing password", http.StatusInternalServerError)
		return
	}

	// Parse date of birth
	dateOfBirth, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		http.Error(w, "Invalid date format for date of birth", http.StatusBadRequest)
		return
	}

	// Create user model
	user := models.User{
		Email:       req.Email,
		Password:    string(hashedPassword),
		Username:    req.Username,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DateOfBirth: dateOfBirth,
		AboutMe:     req.AboutMe,
		Avatar:      req.Avatar,
		CreatedAt:   time.Now(),
	}

	// Insert into database
	result, err := sqlite.DB.Exec(`
		INSERT INTO users (email, password, username, first_name, last_name, date_of_birth, about_me, avatar, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.Email, user.Password, user.Username, user.FirstName, user.LastName, user.DateOfBirth, user.AboutMe, user.Avatar, user.CreatedAt)
	
	if err != nil {
		http.Error(w, "Error creating user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Error getting user ID", http.StatusInternalServerError)
		return
	}

	// Create session
	sessionToken := util.GenerateSessionToken()
	util.UserSession[sessionToken] = uint(id)

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "AccessToken",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.UserResponse{
		ID:       id,
		Username: user.Username,
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var user models.User
	err := sqlite.DB.QueryRow("SELECT id, email, password, username FROM users WHERE email = ?", req.Email).
		Scan(&user.ID, &user.Email, &user.Password, &user.Username)
	
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	sessionToken := util.GenerateSessionToken()
	util.UserSession[sessionToken] = user.ID

	http.SetCookie(w, &http.Cookie{
		Name:     "AccessToken",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.UserResponse{
			ID:       int64(user.ID),
			Username: user.Username,
	})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("AccessToken")
	if err != nil {
		http.Error(w, "No session to logout from", http.StatusBadRequest)
		return
	}

	// Delete the session
	delete(util.UserSession, cookie.Value)

	// Expire the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "AccessToken",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	})

	if _, err := w.Write([]byte("User logged out successfully")); err != nil {
		log.Printf("Logout: %v", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
}
