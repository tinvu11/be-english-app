WITH ranked_attempts AS (
    SELECT id,
           ROW_NUMBER() OVER (
               PARTITION BY user_id, caption_id
               ORDER BY pronunciation_score DESC NULLS LAST, created_at DESC, id DESC
           ) AS result_rank
    FROM shadowing_attempts
)
DELETE FROM shadowing_attempts attempts
USING ranked_attempts ranked
WHERE attempts.id = ranked.id AND ranked.result_rank > 1;

ALTER TABLE shadowing_attempts
    ADD CONSTRAINT uq_shadowing_user_caption UNIQUE (user_id, caption_id);

