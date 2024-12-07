CREATE TABLE IF NOT EXISTS group_post_comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content TEXT NOT NULL,
    media BLOB,
    media_type TEXT,
    post_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    author INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (author) REFERENCES users(id)
);

-- Create indexes for better performance
CREATE INDEX idx_group_comments_post ON group_post_comments(post_id);
CREATE INDEX idx_group_comments_group ON group_post_comments(group_id);
CREATE INDEX idx_group_comments_author ON group_post_comments(author); 