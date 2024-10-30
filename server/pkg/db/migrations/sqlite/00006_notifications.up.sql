PRAGMA foreign_keys = ON;



CREATE TABLE notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    to_user_id INTEGER REFERENCES users(id),             
    content TEXT NOT NULL,    
    from_user_id INTEGER REFERENCES users(id),        
    read BOOLEAN DEFAULT FALSE,            
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    group_id INTEGER REFERENCES groups(id) -- Optional
);

CREATE TABLE likes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id  INTEGER REFERENCES posts(id),
    comment_id INTEGER REFERENCES comments(id),
    user_id    INTEGER REFERENCES users(id),
    is_like    BOOLEAN NOT NULL,
    CHECK (post_id IS NOT NULL OR comment_id IS NOT NULL)
);
