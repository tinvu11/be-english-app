DROP INDEX IF EXISTS users_firebase_uid_key;

ALTER TABLE users
    DROP COLUMN IF EXISTS firebase_uid;

UPDATE users
SET password_hash = ''
WHERE password_hash IS NULL;

ALTER TABLE users
    ALTER COLUMN password_hash SET NOT NULL;
