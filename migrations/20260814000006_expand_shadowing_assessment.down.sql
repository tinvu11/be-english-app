DROP INDEX IF EXISTS idx_shadowing_attempts_user_created;
DROP TABLE IF EXISTS shadowing_word_results;

ALTER TABLE shadowing_attempts
    DROP COLUMN IF EXISTS reference_text,
    DROP COLUMN IF EXISTS duration_ms,
    DROP COLUMN IF EXISTS locale,
    DROP COLUMN IF EXISTS prosody_score,
    DROP COLUMN IF EXISTS provider;

ALTER TABLE shadowing_attempts
    RENAME COLUMN provider_response_json TO azure_response_json;
