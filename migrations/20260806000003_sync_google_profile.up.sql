ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_key;
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
