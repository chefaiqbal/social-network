-- Drop and recreate followers table with proper constraints
DROP TABLE IF EXISTS followers;

CREATE TABLE followers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    follower_id INTEGER NOT NULL,
    followed_id INTEGER NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'accept', 'reject')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (follower_id) REFERENCES users(id),
    FOREIGN KEY (followed_id) REFERENCES users(id),
    UNIQUE(follower_id, followed_id)
);

-- Create index for better performance
CREATE INDEX idx_followers_status ON followers(follower_id, followed_id, status); 