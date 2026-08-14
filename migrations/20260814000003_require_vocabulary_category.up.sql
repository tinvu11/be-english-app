CREATE TEMP TABLE orphan_vocabulary_categories (
    user_id UUID NOT NULL,
    source_language_id INT,
    target_language_id INT,
    vocab_set_id BIGINT NOT NULL
) ON COMMIT DROP;

-- Reuse an existing "Chưa phân loại" set for the same language pair when possible.
INSERT INTO orphan_vocabulary_categories (user_id, source_language_id, target_language_id, vocab_set_id)
SELECT DISTINCT ON (uv.user_id, d.source_language_id, d.target_language_id)
       uv.user_id, d.source_language_id, d.target_language_id, s.id
FROM user_vocabularies uv
JOIN dictionary d ON d.id = uv.dictionary_id
JOIN vocab_sets s ON s.user_id = uv.user_id
    AND s.source_language_id IS NOT DISTINCT FROM d.source_language_id
    AND s.target_language_id IS NOT DISTINCT FROM d.target_language_id
    AND s.title = 'Chưa phân loại'
WHERE uv.vocab_set_id IS NULL
ORDER BY uv.user_id, d.source_language_id, d.target_language_id, s.id;

-- Create one fallback category per user and language pair that still needs one.
WITH missing AS (
    SELECT DISTINCT uv.user_id, d.source_language_id, d.target_language_id
    FROM user_vocabularies uv
    JOIN dictionary d ON d.id = uv.dictionary_id
    LEFT JOIN orphan_vocabulary_categories mapped
        ON mapped.user_id = uv.user_id
        AND mapped.source_language_id IS NOT DISTINCT FROM d.source_language_id
        AND mapped.target_language_id IS NOT DISTINCT FROM d.target_language_id
    WHERE uv.vocab_set_id IS NULL AND mapped.user_id IS NULL
), inserted AS (
    INSERT INTO vocab_sets (user_id, title, source_language_id, target_language_id)
    SELECT user_id, 'Chưa phân loại', source_language_id, target_language_id
    FROM missing
    RETURNING id, user_id, source_language_id, target_language_id
)
INSERT INTO orphan_vocabulary_categories (user_id, source_language_id, target_language_id, vocab_set_id)
SELECT user_id, source_language_id, target_language_id, id
FROM inserted;

UPDATE user_vocabularies uv
SET vocab_set_id = mapped.vocab_set_id
FROM dictionary d, orphan_vocabulary_categories mapped
WHERE uv.vocab_set_id IS NULL
  AND d.id = uv.dictionary_id
  AND mapped.user_id = uv.user_id
  AND mapped.source_language_id IS NOT DISTINCT FROM d.source_language_id
  AND mapped.target_language_id IS NOT DISTINCT FROM d.target_language_id;

ALTER TABLE user_vocabularies
    ALTER COLUMN vocab_set_id SET NOT NULL;
