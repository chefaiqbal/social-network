CREATE TABLE IF NOT EXISTS group_posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    media BLOB,
    media_type TEXT,
    author INTEGER REFERENCES users(id),
    group_id INTEGER REFERENCES groups(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (author) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

CREATE INDEX idx_group_posts_group ON group_posts(group_id);
CREATE INDEX idx_group_posts_author ON group_posts(author);

-- Move existing group posts from posts table to group_posts
INSERT INTO group_posts (title, content, media, media_type, author, group_id, created_at)
SELECT title, content, media, media_type, author, group_id, created_at
FROM posts
WHERE group_id IS NOT NULL;

-- Remove group posts from posts table
DELETE FROM posts WHERE group_id IS NOT NULL; 