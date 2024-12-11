-- Remove the type column from notifications
CREATE TABLE notifications_backup (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    to_user_id INTEGER REFERENCES users(id),             
    content TEXT NOT NULL,    
    from_user_id INTEGER REFERENCES users(id),        
    read BOOLEAN DEFAULT FALSE,            
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    group_id INTEGER REFERENCES groups(id)
);

INSERT INTO notifications_backup 
SELECT id, to_user_id, content, from_user_id, read, created_at, group_id 
FROM notifications;

DROP TABLE notifications;

ALTER TABLE notifications_backup RENAME TO notifications; 