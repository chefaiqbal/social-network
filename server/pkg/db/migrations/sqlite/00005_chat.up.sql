PRAGMA foreign_keys = ON;

--- Chat Tables

CREATE TABLE chat_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,          
    sender_id INTEGER REFERENCES users(id),           
    recipient_id INTEGER REFERENCES users(id),            
    content TEXT NOT NULL,        
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (recipient_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE group_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,        
    group_id INTEGER REFERENCES groups(id),           
    user_id INTEGER REFERENCES users(id),            
    content TEXT NOT NULL,            
    media TEXT, 
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_chat_messages_users ON chat_messages(sender_id, recipient_id);
CREATE INDEX idx_group_messages_users ON group_messages(group_id, user_id);

