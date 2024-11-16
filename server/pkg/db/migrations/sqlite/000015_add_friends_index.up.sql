-- Add index for faster friend lookups
CREATE INDEX IF NOT EXISTS idx_friends ON followers(follower_id, followed_id, status); 