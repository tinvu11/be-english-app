DROP INDEX IF EXISTS idx_vocab_sets_user_languages_created;

ALTER TABLE vocab_sets
    DROP CONSTRAINT IF EXISTS chk_vocab_sets_different_languages,
    DROP COLUMN IF EXISTS target_language_id,
    DROP COLUMN IF EXISTS source_language_id;
