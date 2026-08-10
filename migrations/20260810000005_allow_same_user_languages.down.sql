ALTER TABLE users
    ADD CONSTRAINT chk_user_languages CHECK (
        native_language_id IS NULL
        OR target_language_id IS NULL
        OR native_language_id <> target_language_id
    );
