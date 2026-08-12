CREATE TABLE dictionary (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    word VARCHAR(150) NOT NULL CHECK (length(trim(word)) > 0),
    source_language_id INT REFERENCES languages(id) ON DELETE RESTRICT,
    target_language_id INT REFERENCES languages(id) ON DELETE RESTRICT,
    phonetic_or_pinyin VARCHAR(150),
    part_of_speech VARCHAR(50),
    meaning TEXT NOT NULL CHECK (length(trim(meaning)) > 0),
    example_1_sentence TEXT,
    example_1_translation TEXT,
    example_2_sentence TEXT,
    example_2_translation TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_dictionary_word_lang
        UNIQUE NULLS NOT DISTINCT (word, source_language_id, target_language_id)
);

CREATE INDEX idx_dictionary_search
    ON dictionary(word, source_language_id, target_language_id);

INSERT INTO dictionary (
    word,
    source_language_id,
    target_language_id,
    phonetic_or_pinyin,
    part_of_speech,
    meaning,
    example_1_sentence,
    example_1_translation,
    example_2_sentence,
    example_2_translation,
    created_at
)
SELECT DISTINCT ON (uv.word, u.target_language_id, u.native_language_id)
    uv.word,
    u.target_language_id,
    u.native_language_id,
    uv.phonetic_or_pinyin,
    uv.part_of_speech,
    uv.meaning,
    uv.example_1_sentence,
    uv.example_1_translation,
    uv.example_2_sentence,
    uv.example_2_translation,
    uv.created_at
FROM user_vocabularies uv
JOIN users u ON u.id = uv.user_id
ORDER BY uv.word, u.target_language_id, u.native_language_id, uv.created_at, uv.id;

ALTER TABLE user_vocabularies
    ADD COLUMN dictionary_id BIGINT,
    ADD COLUMN caption_id BIGINT REFERENCES video_captions(id) ON DELETE SET NULL;

UPDATE user_vocabularies uv
SET dictionary_id = d.id
FROM users u, dictionary d
WHERE u.id = uv.user_id
  AND d.word = uv.word
  AND d.source_language_id IS NOT DISTINCT FROM u.target_language_id
  AND d.target_language_id IS NOT DISTINCT FROM u.native_language_id;

-- The old schema allowed the same word to be saved repeatedly in one set.
-- Merge those rows before adding the new ownership-aware uniqueness constraint.
WITH ranked AS (
    SELECT id,
           FIRST_VALUE(id) OVER (
               PARTITION BY user_id, vocab_set_id, dictionary_id
               ORDER BY created_at, id
           ) AS keeper_id
    FROM user_vocabularies
), merged AS (
    SELECT ranked.keeper_id,
           bool_or(uv.is_learned) AS is_learned,
           max(uv.learned_at) AS learned_at,
           min(uv.created_at) AS created_at,
           max(uv.updated_at) AS updated_at
    FROM ranked
    JOIN user_vocabularies uv ON uv.id = ranked.id
    GROUP BY ranked.keeper_id
)
UPDATE user_vocabularies keeper
SET is_learned = merged.is_learned,
    learned_at = merged.learned_at,
    created_at = merged.created_at,
    updated_at = merged.updated_at
FROM merged
WHERE keeper.id = merged.keeper_id;

WITH ranked AS (
    SELECT id,
           row_number() OVER (
               PARTITION BY user_id, vocab_set_id, dictionary_id
               ORDER BY created_at, id
           ) AS duplicate_position
    FROM user_vocabularies
)
DELETE FROM user_vocabularies uv
USING ranked
WHERE uv.id = ranked.id
  AND ranked.duplicate_position > 1;

ALTER TABLE user_vocabularies
    ALTER COLUMN dictionary_id SET NOT NULL,
    ADD CONSTRAINT fk_user_vocabularies_dictionary
        FOREIGN KEY (dictionary_id) REFERENCES dictionary(id) ON DELETE CASCADE,
    ADD CONSTRAINT uq_user_vocab_set_dict
        UNIQUE NULLS NOT DISTINCT (user_id, vocab_set_id, dictionary_id),
    DROP COLUMN word,
    DROP COLUMN phonetic_or_pinyin,
    DROP COLUMN part_of_speech,
    DROP COLUMN meaning,
    DROP COLUMN example_1_sentence,
    DROP COLUMN example_1_translation,
    DROP COLUMN example_2_sentence,
    DROP COLUMN example_2_translation;

DROP INDEX IF EXISTS idx_user_vocabularies_user_learned_created;
