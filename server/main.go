package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"social-network/api"
	"social-network/middleware"
	"social-network/pkg/db/sqlite"
	"social-network/util"
)

// authMiddleware checks the existence of the cookie on each handler
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("AccessToken")
		if err != nil {
			if err == http.ErrNoCookie || cookie.Value == "" {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				http.Error(w, "Unauthenticated user", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Something went wrong", http.StatusUnauthorized)
			return
		}

		cookieValue := cookie.Value

		if _, ok := util.UserSession[cookieValue]; !ok {
			http.Error(w, "Unauthorized user", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func checkAuth(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("AccessToken")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Authenticated request"))
}

func main() {
	log.Println("Initializing database...")
	
	dbPath := "./social-network.db"
	log.Printf("Using database at: %s", dbPath)
	
	// Add version flag
	var targetVersion uint
	flag.UintVar(&targetVersion, "version", 0, "Specify migration version to use (0 means latest)")
	flag.Parse()

	args := flag.Args()
	
	if len(args) > 0 && args[0] == "rollback" {
		// Handle rollback command
		if err := sqlite.OpenDB(dbPath, nil); err != nil {
			log.Fatal(err)
		}
		if err := sqlite.RollbackMigrations(1); err != nil {
			log.Fatal(err)
		}
		return
	}

	// Normal server start
	opts := &sqlite.DBOptions{
		ApplyMigrations: true,
	}
	
	// If version specified, set it in options
	if targetVersion > 0 {
		opts.TargetVersion = &targetVersion
	}

	if err := sqlite.OpenDB(dbPath, opts); err != nil {
		log.Fatal(err)
	}
	defer sqlite.DB.Close()

	// Handle command line arguments
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "flush":
			if err := sqlite.ClearDatabase(); err != nil {
				log.Fatalf("Error flushing database: %v", err)
			}
			log.Println("Database flushed successfully")
		case "rollback":
			if err := sqlite.RollbackMigrations(1); err != nil { // Changed to pass required steps argument
				log.Fatalf("Error rolling back migrations: %v", err)
			}
			return
		case "cleanup":
			if err := sqlite.CleanupMigrations(); err != nil {
				log.Fatalf("Error cleaning up migrations: %v", err)
			}
			log.Println("Migration cleanup completed")
			return
		}
	}

	mux := http.NewServeMux()

	// Add CORS middleware
	handler := middleware.CORS(mux)

	//NOTE: GO VERSION 1.22+ WILL BE USED IN THIS PROJECT IF YOU DON'T HAVE THAT PLEASE UPDATE YOUR GO
	mux.HandleFunc("POST /register", api.RegisterHandler)
	mux.HandleFunc("POST /login", api.LoginHandler)
	mux.HandleFunc("POST /logout", api.LogoutHandler)
	mux.HandleFunc("POST /userID", api.GetUserIDED)  // BY NAME
	mux.HandleFunc("GET /userIDBY", api.GetUserIDBY) // BY itself
	mux.HandleFunc("GET /userName", func(w http.ResponseWriter, r *http.Request) {
		api.GetUername(r, w)
	}) // Get username
	mux.Handle("GET /checkAuth", authMiddleware(http.HandlerFunc(checkAuth)))

	mux.Handle("POST /posts", authMiddleware(http.HandlerFunc(api.CreatePost)))
	mux.Handle("GET /posts/{id}", authMiddleware(http.HandlerFunc(api.ViewPost)))
	mux.Handle("GET /posts", authMiddleware(http.HandlerFunc(api.GetPosts)))

	mux.Handle("POST /comments", authMiddleware(http.HandlerFunc(api.CreateComment)))
	mux.Handle("GET /comments/{postID}", authMiddleware(http.HandlerFunc(api.GetComments)))

	mux.Handle("GET /groups", authMiddleware(http.HandlerFunc(api.VeiwGorups)))
	mux.Handle("POST /groups", authMiddleware(http.HandlerFunc(api.CreateGroup)))
	mux.Handle("POST /groups/{id}/posts", authMiddleware(http.HandlerFunc(api.CreateGroupPost)))
	mux.Handle("GET /groups/{id}/posts", middleware.GroupAuthMiddleware(api.GetGroupPost))
	mux.Handle("POST /groups/{groupId}/posts/{postId}/comments", authMiddleware(http.HandlerFunc(api.CreateGroupPostComment)))
	mux.Handle("GET /groups/{groupId}/posts/{postId}/comments", authMiddleware(http.HandlerFunc(api.GetGroupPostComments)))
	mux.Handle("POST /groups/invitation", authMiddleware(http.HandlerFunc(api.GroupInvitation)))
	mux.Handle("GET /groups/invitation", authMiddleware(http.HandlerFunc(api.GetGroupInvitations)))
	mux.Handle("POST /groups/invitation/accept", authMiddleware(http.HandlerFunc(api.InvitationAccept)))
	mux.Handle("POST /groups/invitation/decline", authMiddleware(http.HandlerFunc(api.InvitationDecline)))
	mux.Handle("POST /groups/JoinRequest", authMiddleware(http.HandlerFunc(api.GroupJoinRequest)))

	mux.Handle("POST /groups/accept", authMiddleware(http.HandlerFunc(api.GroupAccept)))
	mux.Handle("POST /groups/reject", authMiddleware(http.HandlerFunc(api.GroupReject)))
	mux.Handle("POST /groups/leave", authMiddleware(http.HandlerFunc(api.GroupLeave)))
	mux.Handle("GET /groups/myGroup", authMiddleware(http.HandlerFunc(api.MyGroups)))
	mux.Handle("DELETE /groups/{id}", authMiddleware(middleware.GroupAuthMiddleware(api.DelGroup)))
	mux.Handle("GET /groups/{id}", middleware.GroupAuthMiddleware(api.GetGroupName))

	mux.Handle("POST /event/create", middleware.GroupAuthMiddleware(api.CreateEvent))
	mux.Handle("GET  /event/getGroupEvents/{id}", middleware.GroupAuthMiddleware(api.GetGroupEvents))
	mux.Handle("POST /event/rsvp", authMiddleware(http.HandlerFunc(api.RSVPEvent)))
	mux.Handle("GET /event/rsvps/{id}", http.HandlerFunc(api.GetRSVPs))
	mux.Handle("POST /groups/pendingUsers", authMiddleware(http.HandlerFunc(api.GetPendingUsers)))
	mux.Handle("POST /groups/getnonmembers", (http.HandlerFunc(api.GetnonMembers)))
	mux.Handle("POST /groups/getMembers", authMiddleware(http.HandlerFunc(api.GetMembers)))
	mux.Handle("POST /groups/ismember", (http.HandlerFunc(api.IsMember)))
	mux.Handle("POST /follow", authMiddleware(http.HandlerFunc(api.RequestFollowUser)))
	mux.Handle("POST /Unfollow", authMiddleware(http.HandlerFunc(api.UnfollowUser)))
	mux.Handle("PATCH /follow/request/{id}", authMiddleware(http.HandlerFunc(api.AcceptOrRejectRequest)))
	mux.Handle("PATCH /follow/requestF/{id}", authMiddleware(http.HandlerFunc(api.HandelAcceptOrRejectRequest)))
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
	mux.Handle("GET /following", authMiddleware(http.HandlerFunc(api.GetFollowingBYIt)))

	mux.Handle("GET /comments/{postID}/count", authMiddleware(http.HandlerFunc(api.GetCommentCount)))

	mux.Handle("GET /groups/messages", authMiddleware(http.HandlerFunc(api.GetGroupChatMessages)))

	mux.Handle("/ws/group-chat", authMiddleware(http.HandlerFunc(api.GroupChatHandler)))

	fmt.Println("Server running on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
