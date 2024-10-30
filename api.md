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
"username": "string (optional)",
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

## Profile Endpoints

### Get Profile
- **URL**: `/profile/{userID}`
- **Method**: `GET`
- **Auth Required**: Yes

### Toggle Profile Privacy
- **URL**: `/profile/privacy`
- **Method**: `PUT`
- **Auth Required**: Yes
- **Body**:
```json
{
"is_private": boolean
}
```

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
"privacy": 0, // 0: public, 1: private, 2: almost private
"allowed_users": ["userID"] // required if privacy is 2
}
```

### Create Comment
- **URL**: `/posts/{postID}/comments`
- **Method**: `POST`
- **Auth Required**: Yes
- **Body**:
```json
{
"content": "string",
"media": "string (optional)"
}
```

## Follow System

### Send Follow Request
- **URL**: `/follow/request`
- **Method**: `POST`
- **Auth Required**: Yes
- **Body**:
```json
{
"followed_id": "integer"
}
```

## Groups

### Create Group
- **URL**: `/groups`
- **Method**: `POST`
- **Auth Required**: Yes
- **Body**:
```json
{
"title": "string",
"description": "string"
}
```

### Create Group Event
- **URL**: `/groups/{groupID}/events`
- **Method**: `POST`
- **Auth Required**: Yes
- **Body**:
```json
{
"title": "string",
"description": "string",
"event_date": "datetime",
"options": ["going", "not_going"]
}
```

### Join Group Request
- **URL**: `/groups/{groupID}/join`
- **Method**: `POST`
- **Auth Required**: Yes

## WebSocket Connection
- **URL**: `/ws`
- **Protocol**: WebSocket
- **Auth Required**: Yes

### Message Types:
1. Chat Message:{
"type": "chat",
"content": {
"recipient_id": "integer",
"message": "string"
}
}


2. Group Chat Message:{
"type": "groupChat",
"content": {
"group_id": "integer",
"message": "string"
}
}


3. Notification:
{
"type": "notification",
"content": {
"to_user_id": "integer",
"content": "string",
"notification_type": "string" // follow_request, group_invite, event_created
}
}


## Media Upload
- **URL**: `/upload`
- **Method**: `POST`
- **Auth Required**: Yes
- **Content-Type**: `multipart/form-data`
- **Accepted Formats**: JPEG, PNG, GIF
