# Social Network Project Task List


## Backend Implementation Status

### Authentication & User Management
- [x] User registration with required fields:
  - [x] Email
  - [x] Password
  - [x] First Name
  - [x] Last Name
  - [x] Date of Birth
  - [x] Avatar/Image (Optional)
  - [x] Nickname (Optional)
  - [x] About Me (Optional)
- [x] Login functionality
- [x] Session management with cookies
- [x] Logout functionality
- [ ] Prevent duplicate email registration
- [ ] Multi-browser session handling

### Profile System
- [x] Basic profile data structure
- [x] Public/Private profile toggle
- [ ] Display all user information (except password)
- [ ] Display all user posts
- [ ] Display followers and following users
- [ ] Profile privacy enforcement
  - [ ] Public profile visible to all
  - [ ] Private profile visible only to followers

### Posts System
- [x] Basic post creation
- [x] Post privacy settings
  - [x] Public (all users can see)
  - [x] Private (only followers can see)
  - [x] Almost private (selected followers only)
- [x] Comments on posts
- [ ] Media support validation
  - [ ] JPEG support
  - [ ] PNG support
  - [ ] GIF support
- [ ] Selected users for almost private posts

### Follow System
- [x] Follow request functionality
- [x] Accept/Reject follow requests
- [x] Automatic following for public profiles
- [ ] Unfollow functionality
- [ ] Follow request notifications

### Groups
- [x] Group creation (title & description)
- [x] Group membership management
- [x] Group posts
- [x] Group invitations
- [x] Join request handling
- [ ] Group events
  - [ ] Title
  - [ ] Description
  - [ ] Day/Time
  - [ ] Going/Not Going options
  - [ ] Event response tracking
- [ ] Group chat room
- [ ] Member invitation by existing members

### Chat System
- [x] Basic WebSocket setup
- [x] Private messaging structure
- [x] Message persistence in database
- [ ] Follow-based chat restrictions
- [ ] Emoji support
- [ ] Group chat functionality
- [ ] Real-time message delivery

### Notifications
- [x] Database structure for notifications
- [x] WebSocket setup for notifications
- [ ] Notification visibility on all pages
- [ ] Follow request notifications
- [ ] Group invitation notifications
- [ ] Group join request notifications
- [ ] Group event notifications

### Database
- [x] User table
- [x] Posts table
- [x] Comments table
- [x] Groups table
- [x] Chat messages table
- [x] Notifications table
- [x] Followers table
- [x] Likes table
- [x] Database migrations
- [x] Migration folder structure

## Frontend Implementation (Next.js)
- [ ] Authentication pages
  - [ ] Login page
  - [ ] Registration page with all required fields
- [ ] Profile pages
  - [ ] View profile
  - [ ] Toggle privacy setting
  - [ ] Display posts/followers
- [ ] Post management
  - [ ] Create post with privacy options
  - [ ] Media upload
  - [ ] Comment interface
- [ ] Group features
  - [ ] Group creation
  - [ ] Group management
  - [ ] Event creation
- [ ] Chat interface
  - [ ] Private chat
  - [ ] Group chat
  - [ ] Emoji support
- [ ] Notification system
  - [ ] Global notification access
  - [ ] Real-time updates

## Docker
- [ ] Backend container
- [ ] Frontend container
- [ ] Container verification
  - [ ] Non-zero container sizes
  - [ ] Application accessibility