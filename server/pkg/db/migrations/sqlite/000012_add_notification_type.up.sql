-- Add type column to notifications table
ALTER TABLE notifications ADD COLUMN type TEXT;

-- Update existing notifications with a default type
UPDATE notifications SET type = 'general' WHERE type IS NULL; 