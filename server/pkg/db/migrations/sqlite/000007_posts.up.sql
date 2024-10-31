CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    media TEXT,
    privacy INTEGER CHECK (privacy IN (1, 2, 3)) DEFAULT 1,
    author INTEGER REFERENCES users(id),
    group_id INTEGER REFERENCES groups(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS post_PrivateViews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER REFERENCES posts(id),            
    user_id INTEGER REFERENCES users(id)
); 