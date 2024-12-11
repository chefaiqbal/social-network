-- Move group posts back to posts table
INSERT INTO posts (title, content, media, media_type, author, group_id, created_at)
SELECT title, content, media, media_type, author, group_id, created_at
FROM group_posts;

-- Drop the group_posts table
DROP INDEX IF EXISTS idx_group_posts_group;
DROP INDEX IF EXISTS idx_group_posts_author;
DROP TABLE IF EXISTS group_posts; 