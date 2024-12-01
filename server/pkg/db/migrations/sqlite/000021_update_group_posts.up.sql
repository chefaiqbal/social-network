-- Update posts table to handle media properly
ALTER TABLE posts RENAME TO posts_old;

CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    media BLOB,
    media_type TEXT,
    privacy INTEGER CHECK (privacy IN (1, 2, 3)) DEFAULT 1,
    author INTEGER REFERENCES users(id),
    group_id INTEGER REFERENCES groups(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (author) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

-- Copy existing data
INSERT INTO posts (id, title, content, privacy, author, group_id, created_at)
SELECT id, title, content, privacy, author, group_id, created_at
FROM posts_old;

-- Drop old table
DROP TABLE posts_old;

-- Add indexes for better performance
CREATE INDEX idx_posts_group ON posts(group_id);
CREATE INDEX idx_posts_author ON posts(author); 