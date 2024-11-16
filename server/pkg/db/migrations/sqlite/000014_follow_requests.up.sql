-- Add indexes and constraints to followers table
DROP INDEX IF EXISTS idx_followers_status;
DROP INDEX IF EXISTS idx_followers_users;

CREATE INDEX idx_followers_status ON followers(follower_id, followed_id, status);
CREATE INDEX idx_followers_users ON followers(follower_id, followed_id);

-- Add trigger to prevent duplicate follow requests
CREATE TRIGGER IF NOT EXISTS prevent_duplicate_follows
BEFORE INSERT ON followers
FOR EACH ROW
WHEN EXISTS (
    SELECT 1 FROM followers 
    WHERE follower_id = NEW.follower_id 
    AND followed_id = NEW.followed_id
    AND status != 'reject'
)
BEGIN
    SELECT RAISE(ABORT, 'Follow request already exists');
END;

-- Add trigger to create notification when follow request is created
CREATE TRIGGER IF NOT EXISTS create_follow_notification
AFTER INSERT ON followers
WHEN NEW.status = 'pending'
BEGIN
    INSERT INTO notifications (
        to_user_id,
        from_user_id,
        content,
        type,
        read,
        created_at
    )
    SELECT
        NEW.followed_id,
        NEW.follower_id,
        (SELECT username || ' wants to follow you' FROM users WHERE id = NEW.follower_id),
        'follow_request',
        false,
        DATETIME('now')
    WHERE EXISTS (
        SELECT 1 FROM users 
        WHERE id = NEW.followed_id 
        AND is_private = true
    );
END;

-- Add trigger to create notification when follow request is accepted/rejected
CREATE TRIGGER IF NOT EXISTS follow_request_response_notification
AFTER UPDATE ON followers
WHEN OLD.status = 'pending' AND (NEW.status = 'accept' OR NEW.status = 'reject')
BEGIN
    INSERT INTO notifications (
        to_user_id,
        from_user_id,
        content,
        type,
        read,
        created_at
    )
    SELECT
        NEW.follower_id,
        NEW.followed_id,
        (SELECT username || ' ' || NEW.status || 'ed your follow request' 
         FROM users WHERE id = NEW.followed_id),
        CASE NEW.status
            WHEN 'accept' THEN 'follow_accept'
            WHEN 'reject' THEN 'follow_reject'
        END,
        false,
        DATETIME('now');
END; 