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

DROP INDEX IF EXISTS idx_posts_group;
DROP INDEX IF EXISTS idx_posts_author; 