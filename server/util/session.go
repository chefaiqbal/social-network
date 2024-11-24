package util

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"social-network/pkg/db/sqlite"
)

// UserSession stores the mapping of session tokens to user IDs
var UserSession = make(map[string]uint)

// GenerateSessionToken creates a random session token
func GenerateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// GetUsernameFromSession retrieves the username from the session
func GetUsernameFromSession(r *http.Request) (string, error) {
	cookie, err := r.Cookie("AccessToken")
	if err != nil {
		return "", err
	}

	userID := UserSession[cookie.Value]
	var username string
	err = sqlite.DB.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	if err != nil {
		return "", err
	}

	return username, nil
}

func GetUserID(r  *http.Request, w http.ResponseWriter) (uint64, error) {
	username, err := GetUsernameFromSession(r)
    if err != nil {
        http.Error(w, "Unauthorized: no session cookie", http.StatusUnauthorized)
        return 0, err
    }

    var userID uint64
    err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
    if err != nil {
        http.Error(w, "Unauthorized: user not found", http.StatusUnauthorized)
        return 0, err
    }

	return userID, nil

}

