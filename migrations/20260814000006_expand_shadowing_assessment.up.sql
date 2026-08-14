ALTER TABLE shadowing_attempts
    RENAME COLUMN azure_response_json TO provider_response_json;

ALTER TABLE shadowing_attempts
    ADD COLUMN provider VARCHAR(20) NOT NULL DEFAULT 'azure',
    ADD COLUMN prosody_score NUMERIC(5,2) CHECK (prosody_score BETWEEN 0 AND 100),
    ADD COLUMN locale VARCHAR(20),
    ADD COLUMN duration_ms BIGINT CHECK (duration_ms IS NULL OR duration_ms >= 0),
    ADD COLUMN reference_text TEXT;

UPDATE shadowing_attempts attempts
SET reference_text = captions.content
FROM video_captions captions
WHERE captions.id = attempts.caption_id AND attempts.reference_text IS NULL;

ALTER TABLE shadowing_attempts
    ALTER COLUMN reference_text SET NOT NULL;

CREATE TABLE shadowing_word_results (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    attempt_id BIGINT NOT NULL REFERENCES shadowing_attempts(id) ON DELETE CASCADE,
    word_order INT NOT NULL CHECK (word_order >= 0),
    word TEXT NOT NULL CHECK (length(trim(word)) > 0),
    accuracy_score NUMERIC(5,2) CHECK (accuracy_score BETWEEN 0 AND 100),
    error_type VARCHAR(30),
    phonemes_json JSONB,
    CONSTRAINT uq_shadowing_attempt_word_order UNIQUE (attempt_id, word_order)
);

CREATE INDEX idx_shadowing_attempts_user_created
    ON shadowing_attempts(user_id, created_at DESC);
