-- Remove triggers
DROP TRIGGER IF EXISTS prevent_duplicate_follows;
DROP TRIGGER IF EXISTS create_follow_notification;
DROP TRIGGER IF EXISTS follow_request_response_notification;

-- Remove indexes
DROP INDEX IF EXISTS idx_followers_status;
DROP INDEX IF EXISTS idx_followers_users; 