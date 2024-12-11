# Social Network API Documentation

## Base URL
`http://localhost:8080`

## Authentication Endpoints

### Register User
- **URL**: `/register`
- **Method**: `POST`
- **Body**:
```json
{
"email": "string",
"password": "string",
"first_name": "string",
"last_name": "string",
"date_of_birth": "YYYY-MM-DD",
"avatar": "string (optional)",
"username": "string",
"about_me": "string (optional)"
}
```

### Login
- **URL**: `/login`
- **Method**: `POST`
- **Body**:
```json
{
"email": "string",
"password": "string"
}
```
- **Response**: Sets session cookie

### Logout
- **URL**: `/logout`
- **Method**: `POST`
- **Auth Required**: Yes

### User ID Operations
- **URL**: `/userID`
- **Method**: `POST`
- **Description**: Get user ID by name

- **URL**: `/userIDBY`
- **Method**: `GET`
- **Description**: Get user ID of current user

- **URL**: `/userName`
- **Method**: `GET`
- **Description**: Get username

## Posts Endpoints

### Create Post
- **URL**: `/posts`
- **Method**: `POST`
- **Auth Required**: Yes
- **Body**:
```json
{
"title": "string",
"content": "string",
"media": "string (optional)",
"privacy": 0
}
```

### View Post
- **URL**: `/posts/{id}`
- **Method**: `GET`
- **Auth Required**: Yes

### Get All Posts
- **URL**: `/posts`
- **Method**: `GET`
- **Auth Required**: Yes

### Get User Posts
- **URL**: `/posts/user/{id}`
- **Method**: `GET`
- **Auth Required**: Yes

## Comments

### Create Comment
- **URL**: `/comments`
- **Method**: `POST`
- **Auth Required**: Yes

### Get Comments
- **URL**: `/comments/{postID}`
- **Method**: `GET`
- **Auth Required**: Yes

### Get Comment Count
- **URL**: `/comments/{postID}/count`
- **Method**: `GET`
- **Auth Required**: Yes

## Groups

### View Groups
- **URL**: `/groups`
- **Method**: `GET`
- **Auth Required**: Yes

### Create Group
- **URL**: `/groups`
- **Method**: `POST`
- **Auth Required**: Yes

### Group Posts
- **URL**: `/groups/{id}/posts`
- **Method**: `POST/GET`
- **Auth Required**: Yes

### Group Post Comments
- **URL**: `/groups/{groupId}/posts/{postId}/comments`
- **Method**: `POST/GET`
- **Auth Required**: Yes

### Group Management
- **URL**: `/groups/invitation`
- **Method**: `POST/GET`
- **Auth Required**: Yes

- **URL**: `/groups/invitation/accept`k
- **Method**: `POST`
- **Auth Required**: Yes

- **URL**: `/groups/invitation/decline`
- **Method**: `POST`
- **Auth Required**: Yes

- **URL**: `/groups/JoinRequest`
- **Method**: `POST`
- **Auth Required**: Yes

### Group Member Operations
- **URL**: `/groups/accept`
- **Method**: `POST`
- **Auth Required**: Yes

- **URL**: `/groups/reject`
- **Method**: `POST`
- **Auth Required**: Yes

- **URL**: `/groups/leave`
- **Method**: `POST`
- **Auth Required**: Yes

- **URL**: `/groups/myGroup`
- **Method**: `GET`
- **Auth Required**: Yes

- **URL**: `/groups/Members`
- **Method**: `POST`
- **Auth Required**: Yes

## Events

### Create Event
- **URL**: `/event/create`
- **Method**: `POST`
- **Auth Required**: Yes

### Get Group Events
- **URL**: `/event/getGroupEvents/{id}`
- **Method**: `GET`
- **Auth Required**: Yes

### RSVP to Event
- **URL**: `/event/rsvp`
- **Method**: `POST`
- **Auth Required**: Yes

### Get Event RSVPs
- **URL**: `/event/rsvps/{id}`
- **Method**: `GET`

## Follow System

### Follow Operations
- **URL**: `/follow`
- **Method**: `POST`
- **Auth Required**: Yes

- **URL**: `/Unfollow`
- **Method**: `POST`
- **Auth Required**: Yes

- **URL**: `/follow/request/{id}`
- **Method**: `PATCH`
- **Auth Required**: Yes

### Get Followers/Following
- **URL**: `/followers`
- **Method**: `GET`
- **Auth Required**: Yes

- **URL**: `/following/{userId}`
- **Method**: `GET`
- **Auth Required**: Yes

### Close Friend
- **URL**: `/CloseFriend`
- **Method**: `POST`
- **Auth Required**: Yes

## User Profile & Settings

### Get User Profile
- **URL**: `/user/{userID}`
- **Method**: `GET`
- **Auth Required**: Yes

### Update Privacy Settings
- **URL**: `/user/privacy`
- **Method**: `POST`
- **Auth Required**: Yes

## Chat & Messages

### WebSocket Endpoints
- **URL**: `/ws`
- **URL**: `/ws/chat`
- **URL**: `/ws/likes`
- **URL**: `/ws/group-chat`
- **Auth Required**: Yes

### Get Chat Users
- **URL**: `/chat/users`
- **Method**: `GET`
- **Auth Required**: Yes

### Get Messages
- **URL**: `/messages`
- **Method**: `GET`
- **Auth Required**: Yes

### Get Group Chat Messages
- **URL**: `/groups/messages`
- **Method**: `GET`
- **Auth Required**: Yes

## Notifications

### Get Notifications
- **URL**: `/notifications`
- **Method**: `GET`
- **Auth Required**: Yes

### Mark Notification Read
- **URL**: `/notifications/{id}/read`
- **Method**: `POST`
- **Auth Required**: Yes

### Clear Notifications
- **URL**: `/notifications/{id}/clear`
- **Method**: `POST`
- **Auth Required**: Yes

- **URL**: `/notifications/clear-all`
- **Method**: `POST`
- **Auth Required**: Yes

## Likes

### Like Operations
- **URL**: `/likes`
- **Method**: `POST/GET`
- **Auth Required**: Yes

## User Discovery

### Get Suggested Users
- **URL**: `/users/suggested`
- **Method**: `GET`
- **Auth Required**: Yes

### Get All Users
- **URL**: `/AllUsers`
- **Method**: `GET`
- **Auth Required**: Yes
