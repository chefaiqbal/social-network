-- Update posts table to use BLOB for media
ALTER TABLE posts RENAME TO posts_old;

CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    media BLOB,
    media_type TEXT,  -- Store the mime type (e.g., 'image/jpeg', 'image/png')
    privacy INTEGER CHECK (privacy IN (1, 2, 3)) DEFAULT 1,
    author INTEGER REFERENCES users(id),
    group_id INTEGER REFERENCES groups(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Copy data from old table to new table
INSERT INTO posts (id, title, content, privacy, author, group_id, created_at)
SELECT id, title, content, privacy, author, group_id, created_at
FROM posts_old;

-- Drop the old table
DROP TABLE posts_old;

-- Update users table to use BLOB for avatar
ALTER TABLE users RENAME TO users_old;

CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    username TEXT UNIQUE NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    date_of_birth DATE NOT NULL,
    about_me TEXT,
    avatar BLOB,
    avatar_type TEXT,  -- Store the mime type
    is_private BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    nickname TEXT
);

-- Copy data from old table to new table
INSERT INTO users (id, email, password, username, first_name, last_name, date_of_birth, about_me, is_private, created_at)
SELECT id, email, password, username, first_name, last_name, date_of_birth, about_me, is_private, created_at
FROM users_old;

-- Drop the old table
DROP TABLE users_old; 