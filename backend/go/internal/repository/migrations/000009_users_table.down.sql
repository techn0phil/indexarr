-- Drop indexes first
DROP INDEX IF EXISTS idx_users_enabled;
DROP INDEX IF EXISTS idx_users_username;

-- Drop users table
DROP TABLE IF EXISTS users;
