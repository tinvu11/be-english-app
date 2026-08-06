DROP TABLE IF EXISTS shadowing_attempts;
DROP TABLE IF EXISTS dictation_progress;
DROP TABLE IF EXISTS user_watch_history;
DROP TABLE IF EXISTS user_watch_later;
DROP TABLE IF EXISTS caption_translations;
DROP TABLE IF EXISTS video_captions;
DROP TABLE IF EXISTS video_topics;
DROP TABLE IF EXISTS videos;

DROP INDEX IF EXISTS idx_users_target_language;
DROP INDEX IF EXISTS idx_users_native_language;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS chk_user_languages,
    DROP COLUMN IF EXISTS is_active,
    DROP COLUMN IF EXISTS target_language_id,
    DROP COLUMN IF EXISTS native_language_id,
    DROP COLUMN IF EXISTS avatar_url;

DROP TABLE IF EXISTS channels;
DROP TABLE IF EXISTS topic_translations;
DROP TABLE IF EXISTS topics;
DROP TABLE IF EXISTS level_translations;
DROP TABLE IF EXISTS levels;
DROP TABLE IF EXISTS languages;
