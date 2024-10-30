PRAGMA foreign_keys = ON;

CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL, 
    password TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    date_of_birth DATE NOT NULL,
    avatar TEXT, -- Optional
    about_me TEXT, -- Optional
    is_private BOOLEAN DEFAULT FALSE, -- Public or Private profile
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    media TEXT, -- Path to image or GIF
    privacy INTEGER CHECK (privacy IN (0, 1, 2)) DEFAULT 0, -- 0: public, 1: private, 2: almost private
    author INTEGER REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    group_id INTEGER REFERENCES groups(id) -- Optional
);

CREATE TABLE comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content TEXT NOT NULL,
    media TEXT, -- optional
    post_id INTEGER REFERENCES posts(id),
    author INTEGER REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    creator_id INTEGER REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE group_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER REFERENCES groups(id),
    user_id INTEGER REFERENCES users(id),
    status TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

--- Chat Tables

CREATE TABLE chat_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,          
    sender_id INTEGER REFERENCES users(id),           
    recipient_id INTEGER REFERENCES users(id),            
    content TEXT NOT NULL,        
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP 
);

CREATE TABLE group_chat_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,          
    group_id INTEGER REFERENCES groups(id),            
    user_id INTEGER REFERENCES users(id),             
    content TEXT NOT NULL,            
    media TEXT, 
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP 
);

CREATE TABLE group_messages ( -- Fixed from "group_messsages"
    id INTEGER PRIMARY KEY AUTOINCREMENT,        
    group_id INTEGER REFERENCES groups(id),           
    user_id INTEGER REFERENCES users(id),            
    content TEXT NOT NULL,            
    media TEXT, 
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP 
);

CREATE TABLE group_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER REFERENCES groups(id),              
    creator_id INTEGER REFERENCES users(id),                  
    title TEXT NOT NULL,                     
    description TEXT NOT NULL,                    
    event_date DATETIME NOT NULL,                
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP 
);

CREATE TABLE group_event_RSVP (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER REFERENCES group_events(id), -- Fixed table reference from GroupEvents to group_events                 
    user_id INTEGER REFERENCES users(id),               
    rsvp_status TEXT NOT NULL CHECK (rsvp_status IN ('going', 'not going')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP 
);

CREATE TABLE post_PrivateViews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER REFERENCES posts(id),            
    user_id INTEGER REFERENCES users(id)
);

CREATE TABLE notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    to_user_id INTEGER REFERENCES users(id),             
    content TEXT NOT NULL,    
    from_user_id INTEGER REFERENCES users(id),        
    read BOOLEAN DEFAULT FALSE,            
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    group_id INTEGER REFERENCES groups(id) -- Optional
);
