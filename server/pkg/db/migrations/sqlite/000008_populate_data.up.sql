-- Populate Users
INSERT INTO users (email, password, username, first_name, last_name, date_of_birth, about_me, avatar, is_private, created_at) VALUES
-- Password: 12345678 (bcrypt hashed with cost 10)
('john@example.com', '$2a$10$YEiWyGMTqWkH6rpZJaQFBuLkQc0SCX9CYWVdKXnhWkzQGUJJcXFJi', 'john_doe', 'John', 'Doe', '1990-01-01', 'Software Developer', 'https://api.dicebear.com/7.x/avataaars/svg?seed=john', false, datetime('now')),
('jane@example.com', '$2a$10$YEiWyGMTqWkH6rpZJaQFBuLkQc0SCX9CYWVdKXnhWkzQGUJJcXFJi', 'jane_smith', 'Jane', 'Smith', '1992-05-15', 'UI/UX Designer', 'https://api.dicebear.com/7.x/avataaars/svg?seed=jane', false, datetime('now')),
('bob@example.com', '$2a$10$YEiWyGMTqWkH6rpZJaQFBuLkQc0SCX9CYWVdKXnhWkzQGUJJcXFJi', 'bob_wilson', 'Bob', 'Wilson', '1988-08-20', 'Gamer and Tech Enthusiast', 'https://api.dicebear.com/7.x/avataaars/svg?seed=bob', false, datetime('now')),
('alice@example.com', '$2a$10$YEiWyGMTqWkH6rpZJaQFBuLkQc0SCX9CYWVdKXnhWkzQGUJJcXFJi', 'alice_brown', 'Alice', 'Brown', '1995-03-10', 'Photography Lover', 'https://api.dicebear.com/7.x/avataaars/svg?seed=alice', false, datetime('now')),
('charlie@example.com', '$2a$10$YEiWyGMTqWkH6rpZJaQFBuLkQc0SCX9CYWVdKXnhWkzQGUJJcXFJi', 'charlie_davis', 'Charlie', 'Davis', '1993-11-25', 'Music Producer', 'https://api.dicebear.com/7.x/avataaars/svg?seed=charlie', false, datetime('now'));

-- Populate Groups
INSERT INTO groups (title, description, creator_id, created_at) VALUES
('Tech Enthusiasts', 'A group for technology lovers', 1, datetime('now')),
('Photography Club', 'Share your best shots!', 4, datetime('now')),
('Gaming Squad', 'For all gamers out there', 3, datetime('now')),
('Music Producers', 'Collaborate on music projects', 5, datetime('now'));

-- Add members to groups
INSERT INTO group_members (group_id, user_id, status, created_at) VALUES
(1, 1, 'creator', datetime('now')),
(1, 2, 'member', datetime('now')),
(1, 3, 'member', datetime('now')),
(2, 4, 'creator', datetime('now')),
(2, 1, 'member', datetime('now')),
(3, 3, 'creator', datetime('now')),
(3, 5, 'member', datetime('now')),
(4, 5, 'creator', datetime('now')),
(4, 2, 'member', datetime('now'));

-- Create some followers
INSERT INTO followers (follower_id, followed_id, status, created_at) VALUES
(2, 1, 'accept', datetime('now')),
(3, 1, 'accept', datetime('now')),
(4, 1, 'accept', datetime('now')),
(1, 2, 'accept', datetime('now')),
(3, 2, 'pending', datetime('now')),
(1, 3, 'accept', datetime('now')),
(2, 4, 'accept', datetime('now')),
(5, 4, 'pending', datetime('now'));

-- Create some posts
INSERT INTO posts (title, content, privacy, author, created_at) VALUES
('First Day at Work', 'Started my new job today! Really excited about the opportunities ahead.', 1, 1, datetime('now', '-2 days')),
('Photography Tips', 'Here are some basic tips for better photography: 1. Rule of thirds 2. Good lighting 3. Steady hands', 1, 4, datetime('now', '-1 day')),
('Gaming Session Tonight', 'Who''s up for some multiplayer gaming tonight?', 1, 3, datetime('now', '-12 hours')),
('New Project Launch', 'Just launched my latest project! Check it out and let me know what you think.', 1, 2, datetime('now', '-6 hours')),
('Music Production Setup', 'Finally completed my home studio setup. Ready to create some awesome tracks!', 1, 5, datetime('now', '-3 hours'));

-- Add some comments
INSERT INTO comments (content, post_id, author, created_at) VALUES
('Congratulations on the new job!', 1, 2, datetime('now', '-1 day')),
('Great tips! Very helpful', 2, 1, datetime('now', '-12 hours')),
('Count me in for gaming!', 3, 5, datetime('now', '-6 hours')),
('The project looks amazing!', 4, 3, datetime('now', '-2 hours')),
('Awesome setup! What DAW do you use?', 5, 4, datetime('now', '-1 hour'));

-- Add some likes
INSERT INTO likes (post_id, user_id, is_like) VALUES
(1, 2, true),
(1, 3, true),
(2, 1, true),
(2, 4, true),
(3, 5, true),
(4, 1, true),
(5, 2, true);

-- Add some group posts
INSERT INTO posts (title, content, privacy, author, group_id, created_at) VALUES
('Tech News Discussion', 'What do you think about the latest smartphone releases?', 1, 1, 1, datetime('now', '-1 day')),
('Best Camera Settings', 'Let''s discuss the optimal settings for night photography.', 1, 4, 2, datetime('now', '-12 hours')),
('Gaming Tournament', 'Organizing a tournament this weekend. Who''s in?', 1, 3, 3, datetime('now', '-6 hours'));

-- Add some notifications
INSERT INTO notifications (to_user_id, content, from_user_id, read, created_at) VALUES
(2, 'liked your post', 1, false, datetime('now', '-1 hour')),
(1, 'commented on your post', 3, false, datetime('now', '-2 hours')),
(4, 'started following you', 2, false, datetime('now', '-3 hours')),
(3, 'invited you to join a group', 1, false, datetime('now', '-4 hours')); 