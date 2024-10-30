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
