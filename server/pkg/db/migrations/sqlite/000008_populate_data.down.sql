-- Clear all data in reverse order to avoid foreign key constraints
DELETE FROM notifications;
DELETE FROM likes;
DELETE FROM comments;
DELETE FROM posts;
DELETE FROM followers;
DELETE FROM group_members;
DELETE FROM groups;
DELETE FROM users; 