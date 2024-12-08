# 🌟 Social Network Application

A modern social network application built with Go and Next.js, featuring real-time communication and a microservices architecture.

## 🚀 Features

- **User Authentication** 🔐
  - Registration and Login
  - Session Management
  - Secure Password Handling

- **Social Interactions** 👥
  - Follow/Unfollow Users
  - User Profiles
  - Groups Management

- **Content Management** 📝
  - Post Creation
  - Likes and Comments
  - Media Upload Support

- **Real-time Features** ⚡
  - Live Notifications
  - WebSocket Communication
  - Instant Messaging

## 🛠️ Technology Stack

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Next.js](https://img.shields.io/badge/Next.js-000000?style=for-the-badge&logo=next.js&logoColor=white)
![SQLite](https://img.shields.io/badge/SQLite-003B57?style=for-the-badge&logo=sqlite&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![WebSocket](https://img.shields.io/badge/WebSocket-010101?style=for-the-badge&logo=socket.io&logoColor=white)
![GitHub Copilot](https://img.shields.io/badge/githubcopilot-%23026AA7.svg?style=for-the-badge&logo=githubcopilot&logoColor=white)

## 🏗 Architecture

```
.
├── server/                         # Backend Go server
│   ├── api/                       # API handlers
│   │   ├── auth.go               # Authentication handlers
│   │   ├── chat.go               # Chat functionality
│   │   ├── chat_socket.go        # WebSocket handlers for chat
│   │   ├── comment.go            # Comment handlers
│   │   ├── follow.go             # Follow/Unfollow functionality
│   │   ├── group.go              # Group management
│   │   ├── group_chat.go         # Group chat functionality
│   │   ├── likes.go              # Post likes handlers
│   │   ├── notifications.go      # Notification system
│   │   ├── post.go               # Post management
│   │   ├── socket.go             # WebSocket core functionality
│   │   └── user.go               # User management
│   │
│   ├── middleware/               # HTTP middleware
│   │   └── cors.go              # CORS configuration
│   │
│   ├── models/                   # Data models
│   │   ├── chat.go              # Chat message models
│   │   ├── comment.go           # Comment models
│   │   ├── follow.go            # Follow relationship models
│   │   ├── group.go             # Group and event models
│   │   ├── likes.go             # Like models
│   │   ├── socket.go            # WebSocket models
│   │   └── user.go              # User models
│   │
│   ├── pkg/                      # Shared packages
│   │   ├── db/                  # Database operations
│   │   │   ├── migrations/     # SQLite migrations
│   │   │   └── sqlite/        # SQLite connection
│   │   └── websocket/         # WebSocket handling
│   │
│   ├── util/                     # Utility functions
│   │   ├── session.go           # Session management
│   │   └── validation.go        # Input validation
│   │
│   ├── main.go                   # Application entry point
│   └── go.mod                    # Go module definition
│
├── frontend/                      # Next.js frontend
└── /                       
    └── docker-compose.yml       # Service orchestration
```

### 🔧 Key Components

1. **API Layer** (`/api`)
   - RESTful endpoints for core functionality
   - WebSocket handlers for real-time features
   - Authentication and authorization
   - Group and event management

2. **Models** (`/models`)
   - Structured data types
   - Database schema representations
   - WebSocket message definitions

3. **Middleware** (`/middleware`)
   - CORS configuration
   - Authentication checks
   - Request logging

4. **Database** (`/pkg/db`)
   - SQLite connection management
   - Database migrations
   - Query optimization

5. **WebSocket** (`/pkg/websocket`)
   - Real-time communication
   - Chat functionality
   - Live notifications

6. **Utilities** (`/util`)
   - Session management
   - Input validation
   - Helper functions

### 🔄 Data Flow

1. **Authentication Flow**
   - User registration and login
   - Session management
   - JWT token handling

2. **Real-time Communication**
   - WebSocket connections
   - Chat messaging
   - Live notifications

3. **Social Interactions**
   - Follow/Unfollow system
   - Group management
   - Post interactions

4. **Event Management**
   - Group event creation
   - RSVP handling
   - Event notifications

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.22 or later
- Node.js 20 or later

### Installation

1. Clone the repository:
```bash
git clone https://github.com/chefaiqbal/social-network.git
```

2. Build and run the Docker containers:
```bash
cd social-network
chmod +x build.sh
./build.sh
```

The application will be available at:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080

## 📦 Project Structure

### Backend Components

- **Models**
  - User Management
  - Post Handling
  - Group Management
  - Notifications
  - WebSocket Connections

- **Database**
  - SQLite with Migrations
  - Efficient Query Handling
  - Data Persistence

- **WebSocket**
  - Real-time Communication
  - Live Updates
  - Connection Management

### Frontend Components

- **Next.js Application**
  - Modern React Framework
  - Server-Side Rendering
  - Optimized Performance

## 🔧 Development

### Running in Development Mode

The project uses Docker Compose for development:
```bash
docker-compose up --build
```

### Hot Reloading

- Backend uses Air for hot reloading
- Frontend has built-in Next.js hot reloading

## 📝 API Documentation

Refer to `api.md` for detailed API documentation.

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 👥 Authors

- Abdulla Aljuffairi - [@xoabdulla](https://github.com/xoabdulla)
- Amir Iqbal - [@chefaiqbal](https://github.com/chefaiqbal)
- Fatema Mohamed - [@famohamed96](https://github.com/famohamed96)
- Hussain Helal - [@hlhol](https://github.com/hlhol)

