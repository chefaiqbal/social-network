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

