ALTER TABLE vocab_sets
    ADD COLUMN source_language_id INT REFERENCES languages(id) ON DELETE RESTRICT,
    ADD COLUMN target_language_id INT REFERENCES languages(id) ON DELETE RESTRICT;

-- Preserve old data. A legacy set containing multiple language pairs is split,
-- retaining the original set for its oldest pair and creating one copy per other pair.
DO $$
DECLARE
    current_set RECORD;
    language_pair RECORD;
    destination_set_id BIGINT;
    pair_number INT;
BEGIN
    FOR current_set IN SELECT * FROM vocab_sets ORDER BY id LOOP
        pair_number := 0;
        FOR language_pair IN
            SELECT d.source_language_id, d.target_language_id, MIN(uv.created_at) AS first_saved_at
            FROM user_vocabularies uv
            JOIN dictionary d ON d.id = uv.dictionary_id
            WHERE uv.vocab_set_id = current_set.id
            GROUP BY d.source_language_id, d.target_language_id
            ORDER BY first_saved_at, d.source_language_id, d.target_language_id
        LOOP
            pair_number := pair_number + 1;
            IF pair_number = 1 THEN
                destination_set_id := current_set.id;
                UPDATE vocab_sets
                SET source_language_id = language_pair.source_language_id,
                    target_language_id = language_pair.target_language_id
                WHERE id = current_set.id;
            ELSE
                INSERT INTO vocab_sets (user_id, title, source_language_id, target_language_id, created_at, updated_at)
                VALUES (current_set.user_id, current_set.title, language_pair.source_language_id,
                        language_pair.target_language_id, current_set.created_at, current_set.updated_at)
                RETURNING id INTO destination_set_id;

                UPDATE user_vocabularies uv
                SET vocab_set_id = destination_set_id
                FROM dictionary d
                WHERE uv.vocab_set_id = current_set.id
                  AND d.id = uv.dictionary_id
                  AND d.source_language_id = language_pair.source_language_id
                  AND d.target_language_id = language_pair.target_language_id;
            END IF;
        END LOOP;

        IF pair_number = 0 THEN
            UPDATE vocab_sets s
            SET source_language_id = u.target_language_id,
                target_language_id = u.native_language_id
            FROM users u
            WHERE s.id = current_set.id AND u.id = s.user_id
              AND u.target_language_id IS NOT NULL
              AND u.native_language_id IS NOT NULL
              AND u.target_language_id <> u.native_language_id;
        END IF;
    END LOOP;
END $$;

-- Empty legacy sets created before language onboarding can remain unscoped. They are
-- intentionally hidden; all newly created sets always receive a language pair.
ALTER TABLE vocab_sets
    ADD CONSTRAINT chk_vocab_sets_different_languages CHECK (source_language_id <> target_language_id);

CREATE INDEX idx_vocab_sets_user_languages_created
    ON vocab_sets(user_id, source_language_id, target_language_id, created_at DESC);
