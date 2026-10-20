INSERT INTO languages (code, name, flag_emoji, is_learnable, is_active)
VALUES
    ('vi', 'Tiếng Việt', '🇻🇳', FALSE, TRUE),
    ('en', 'English', '🇬🇧', TRUE, TRUE)
ON CONFLICT (code) DO UPDATE
SET is_active = TRUE,
    is_learnable = CASE
        WHEN EXCLUDED.code = 'en' THEN TRUE
        ELSE languages.is_learnable
    END;
