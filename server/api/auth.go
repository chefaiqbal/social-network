package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	m "social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"

	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user m.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Error reading data", http.StatusBadRequest)
		return
	}

	// check if all the required fields are provided
	if strings.TrimSpace(user.Email) == "" || strings.TrimSpace(user.Username) == "" || strings.TrimSpace(user.Password) == "" || strings.TrimSpace(user.FirstName) == "" || strings.TrimSpace(user.LastName) == "" || strings.TrimSpace(user.AboutMe) == "" || strings.TrimSpace(user.Avatar) == "" || user.DateOfBirth.IsZero() {
		http.Error(w, "Please populate all required fields", http.StatusBadRequest)
		return
	}

	var id int

	// check if the username or email already exists
	err := sqlite.DB.QueryRow("SELECT id FROM users WHERE email = ? OR username = ?", user.Email, user.Username).Scan(&id)

	if err == nil {
		http.Error(w, "User already exists", http.StatusBadRequest)
		return
	} else if err != sql.ErrNoRows {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("select err: %v", err)
		return
	}

	hashedpassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusBadRequest)
		log.Printf("hash error: %v", err)
		return
	}

	res, err := sqlite.DB.Exec("INSERT INTO users (username, email, password, first_name, last_name, avatar, about_me, date_of_birth) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", user.Username, user.Email, string(hashedpassword), user.FirstName, user.LastName, user.Avatar, user.AboutMe, user.DateOfBirth)
	if err != nil {
		http.Error(w, "Something went wrong, please try again later", http.StatusInternalServerError)
		log.Printf("Hash error: %v", err)
		return
	}

	// get the last inserted id from the database
	userID, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("get id: %v", err)
		return
	}

	// generate the session for the user
	if err := util.GenerateSession(w, &user); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}

	userResponse := m.UserResponse{
		ID:       userID,
		Username: user.Username,
	}

	if err := json.NewEncoder(w).Encode(&userResponse); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)

		log.Printf("sending back: %v", err)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user m.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Error reading data", http.StatusBadRequest)
		return
	}

	// both can't be empty one has to be populated
	if strings.TrimSpace(user.Username) == "" && strings.TrimSpace(user.Email) == "" {
		http.Error(w, "Provide a valid identifier", http.StatusBadRequest)
		return
	}

	// get the user from the database
	var username, password string
	if err := sqlite.DB.QueryRow("SELECT username, password FROM users WHERE username = ? OR email = ?", user.Username, user.Email).Scan(&username, &password); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User does not exist", http.StatusBadRequest)
			return
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Login: %v", err)
		return
	}

	// compare the passed password with the existing one
	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(user.Password)); err != nil {
		http.Error(w, "Username or Password incorrect", http.StatusBadRequest)
		return
	}

	// generate the session for the user
	util.GenerateSession(w, &user)

	w.Write([]byte("Login successfull"))
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	util.DestroySession(w, r)
	if _, err := w.Write([]byte("User logged out successfully")); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Logout: %v", err)
	}

	w.Write([]byte("User logged out successfully"))
}
