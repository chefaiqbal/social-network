-- Revert posts table
ALTER TABLE posts RENAME TO posts_temp;

CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    media TEXT,
    privacy INTEGER CHECK (privacy IN (1, 2, 3)) DEFAULT 1,
    author INTEGER REFERENCES users(id),
    group_id INTEGER REFERENCES groups(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO posts (id, title, content, privacy, author, group_id, created_at)
SELECT id, title, content, privacy, author, group_id, created_at
FROM posts_temp;

DROP TABLE posts_temp;

-- Revert users table
ALTER TABLE users RENAME TO users_temp;

CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    username TEXT UNIQUE NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    date_of_birth DATE NOT NULL,
    about_me TEXT,
    avatar TEXT,
    is_private BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO users (id, email, password, username, first_name, last_name, date_of_birth, about_me, is_private, created_at)
SELECT id, email, password, username, first_name, last_name, date_of_birth, about_me, is_private, created_at
FROM users_temp;

DROP TABLE users_temp; 