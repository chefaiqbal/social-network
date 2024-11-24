-- Up migration
CREATE TABLE IF NOT EXISTS followers (
    follower_id INTEGER,
    following_id INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_id, following_id),
    FOREIGN KEY (follower_id) REFERENCES users(id),
    FOREIGN KEY (following_id) REFERENCES users(id)
);

-- Down migration
DROP TABLE IF EXISTS followers; 