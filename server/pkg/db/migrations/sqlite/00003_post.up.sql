PRAGMA foreign_keys = ON;

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
    media TEXT, -- Optional
    post_id INTEGER REFERENCES posts(id),
    author INTEGER REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE post_PrivateViews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER REFERENCES posts(id),            
    user_id INTEGER REFERENCES users(id)
);

