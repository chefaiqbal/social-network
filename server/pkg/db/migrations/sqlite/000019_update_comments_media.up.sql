-- First create a temporary table with the new structure
CREATE TABLE new_comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content TEXT NOT NULL,
    media BLOB,
    media_type TEXT,
    post_id INTEGER NOT NULL,
    author INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (author) REFERENCES users(id)
);

-- Copy existing data
INSERT INTO new_comments (id, content, post_id, author, created_at)
SELECT id, content, post_id, author, created_at
FROM comments;

-- Drop the old table
DROP TABLE comments;

-- Rename the new table
ALTER TABLE new_comments RENAME TO comments;

-- Create indexes for better performance
CREATE INDEX idx_comments_post ON comments(post_id);
CREATE INDEX idx_comments_author ON comments(author); 