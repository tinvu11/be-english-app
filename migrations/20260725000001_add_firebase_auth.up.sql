ALTER TABLE users
    ADD COLUMN firebase_uid VARCHAR(128);

CREATE UNIQUE INDEX users_firebase_uid_key
    ON users (firebase_uid)
    WHERE firebase_uid IS NOT NULL;

ALTER TABLE users
    ALTER COLUMN password_hash DROP NOT NULL;
