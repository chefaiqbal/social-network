package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"social-network/api"
	"social-network/pkg/db/sqlite"
	"social-network/util"
	"social-network/middleware"
)

// authMiddleware checks the existence of the cookie on each handler
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get the cookie from the browser
		cookie, err := r.Cookie("AccessToken")
		if err != nil {
			// check if the cookie exists from the browser
			if err == http.ErrNoCookie {
				http.Error(w, "Unauthenticated user", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Something went wrong", http.StatusUnauthorized)
			return
		}

		// get the value of the cookie
		cookieValue := cookie.Value

		// check if the cookie exists in the already active sessions
		if _, ok := util.UserSession[cookieValue]; !ok {
			http.Error(w, "Unauthorized user", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Open the database connection
	err := sqlite.OpenDB("./social-network.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer sqlite.DB.Close()

	var arg string

	// check if an argument is passed
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}

	// check case insesitive
	if strings.EqualFold(arg, "flush") {
		// remove all data from the database
		if err := sqlite.ClearDatabase(); err != nil {
			log.Fatalf("Error flushing database: %v", err)
		}
	} else if strings.EqualFold(arg, "rollback") {
		// roll back the migrations
		if err := sqlite.RollbackMigrations(); err != nil {
			log.Fatalf("Error rolling back: %v", err)
		}

		// exit after rolling back
		return
	}

	mux := http.NewServeMux()

	// Add CORS middleware
	handler := middleware.CORS(mux)


	//NOTE: GO VERSION 1.22+ WILL BE USED IN THIS PROJECT IF YOU DON'T HAVE THAT PLEASE UPDATE YOUR GO
	mux.HandleFunc("POST /register", api.RegisterHandler)
	mux.HandleFunc("POST /login", api.LoginHandler)
	mux.HandleFunc("POST /logout", api.LogoutHandler)
	mux.HandleFunc("POST /userID", api.GetUserIDED) // BY NAME 
	mux.HandleFunc("GET /userIDBY", api.GetUserIDBY) // BY itself
	mux.HandleFunc("GET /userName", func(w http.ResponseWriter, r *http.Request) {
		api.GetUername(r, w)
	})  // Get username  



	mux.Handle("POST /posts", authMiddleware(http.HandlerFunc(api.CreatePost)))
	mux.Handle("GET /posts/{id}", authMiddleware(http.HandlerFunc(api.ViewPost)))
	mux.Handle("GET /posts", authMiddleware(http.HandlerFunc(api.GetPosts)))

	mux.Handle("POST /comments", authMiddleware(http.HandlerFunc(api.CreateComment)))
	mux.Handle("GET /comments/{postID}", authMiddleware(http.HandlerFunc(api.GetComments)))

	mux.Handle("GET /groups", authMiddleware(http.HandlerFunc(api.VeiwGorups)))
	mux.Handle("POST /groups", authMiddleware(http.HandlerFunc(api.CreateGroup)))
	mux.Handle("POST /groups/{id}/posts", authMiddleware(http.HandlerFunc(api.CreateGroupPost)))
	mux.Handle("GET /groups/{id}/posts", authMiddleware(http.HandlerFunc(api.GetGroupPost)))
	mux.Handle("POST /groups/invitation", authMiddleware(http.HandlerFunc(api.GroupInvitation)))
	mux.Handle("POST /groups/accept", authMiddleware(http.HandlerFunc(api.GroupAccept)))
	mux.Handle("POST /groups/reject", authMiddleware(http.HandlerFunc(api.GroupReject)))
	mux.Handle("POST /groups/leave", authMiddleware(http.HandlerFunc(api.GroupLeave)))
	mux.Handle("GET /groups/myGroup", authMiddleware(http.HandlerFunc(api.MyGroups)))
	mux.Handle("POST /groups/Members", authMiddleware(http.HandlerFunc(api.Members)))
	mux.Handle("DELETE /groups/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		api.DelGroup(r, w)
	}))  

	mux.Handle("POST /event/create", authMiddleware(http.HandlerFunc(api.CreateEvent)))
	mux.Handle("GET  /event/getGroupEvents/{id}", authMiddleware(http.HandlerFunc(api.GetGroupEvents)))
	mux.Handle("POST /event/rsvp",authMiddleware( http.HandlerFunc(api.RSVPEvent)))
	mux.Handle("GET /event/rsvps/{id}", http.HandlerFunc(api.GetRSVPs))


	mux.Handle("POST /follow", authMiddleware(http.HandlerFunc(api.RequestFollowUser)))
	mux.Handle("PATCH /follow/request/{id}", authMiddleware(http.HandlerFunc(api.AcceptOrRejectRequest)))
	mux.Handle("GET /followers", authMiddleware(http.HandlerFunc(api.GetFollowers)))
	mux.Handle("POST /CloseFriend", authMiddleware(http.HandlerFunc(api.CloseFriend)))
	mux.Handle("GET /followStatus", authMiddleware(http.HandlerFunc(api.GetFollowstatus)))
	mux.Handle("GET /followRequest", authMiddleware(http.HandlerFunc(api.FollowRequestHandler)))

	mux.Handle("GET /user/{userID}", authMiddleware(http.HandlerFunc(api.UserProfile)))

	mux.Handle("/ws", authMiddleware(http.HandlerFunc(api.WebSocketHandler)))

	mux.Handle("GET /users/suggested", authMiddleware(http.HandlerFunc(api.GetSuggestedUsers)))
	mux.Handle("GET /AllUsers", authMiddleware(http.HandlerFunc(api.GetAllUsers)))

	mux.Handle("GET /chat/users", authMiddleware(http.HandlerFunc(api.GetChatUsers)))
	mux.Handle("GET /messages", authMiddleware(http.HandlerFunc(api.GetChatMessages)))

	mux.Handle("GET /posts/user/{id}", authMiddleware(http.HandlerFunc(api.GetUserPosts)))

	mux.Handle("POST /user/privacy", authMiddleware(http.HandlerFunc(api.UpdatePrivacySettings)))

	mux.Handle("GET /notifications", authMiddleware(http.HandlerFunc(api.GetNotifications)))

	mux.Handle("POST /notifications/{id}/read", authMiddleware(http.HandlerFunc(api.MarkNotificationRead)))

	mux.Handle("POST /notifications/{id}/clear", authMiddleware(http.HandlerFunc(api.ClearNotification)))
	mux.Handle("POST /notifications/clear-all", authMiddleware(http.HandlerFunc(api.ClearAllNotifications)))

	mux.Handle("/ws/chat", authMiddleware(http.HandlerFunc(api.ChatWebSocketHandler)))

	mux.Handle("POST /likes", authMiddleware(http.HandlerFunc(api.LikeHandler)))
	mux.Handle("GET /likes", authMiddleware(http.HandlerFunc(api.GetPostLikes)))

	mux.Handle("/ws/likes", authMiddleware(http.HandlerFunc(api.LikeWebSocketHandler)))

	mux.Handle("GET /following/{userId}", authMiddleware(http.HandlerFunc(api.GetFollowing)))

	mux.Handle("GET /comments/{postID}/count", authMiddleware(http.HandlerFunc(api.GetCommentCount)))

	fmt.Println("Server running on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}