-- Add some initial chat messages between John and Jane
INSERT INTO chat_messages (sender_id, recipient_id, content, created_at) VALUES
(1, 2, 'Hey Jane, how are you?', datetime('now', '-1 hour')),
(2, 1, 'Hi John! I''m good, thanks for asking!', datetime('now', '-55 minutes')),
(1, 2, 'Great! Would you like to collaborate on a project?', datetime('now', '-50 minutes')),
(2, 1, 'Sure, that sounds interesting!', datetime('now', '-45 minutes')); 