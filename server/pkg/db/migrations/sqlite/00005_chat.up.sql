PRAGMA foreign_keys = ON;

--- Chat Tables

CREATE TABLE chat_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,          
    sender_id INTEGER REFERENCES users(id),           
    recipient_id INTEGER REFERENCES users(id),            
    content TEXT NOT NULL,        
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP 
);

CREATE TABLE group_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,        
    group_id INTEGER REFERENCES groups(id),           
    user_id INTEGER REFERENCES users(id),            
    content TEXT NOT NULL,            
    media TEXT, 
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP 
);

