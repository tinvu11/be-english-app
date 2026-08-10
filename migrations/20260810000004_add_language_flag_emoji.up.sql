ALTER TABLE languages
    ADD COLUMN flag_emoji VARCHAR(16) NOT NULL DEFAULT '🌐';

UPDATE languages
SET flag_emoji = CASE code
    WHEN 'vi' THEN '🇻🇳'
    WHEN 'en' THEN '🇬🇧'
    WHEN 'zh' THEN '🇨🇳'
    WHEN 'ja' THEN '🇯🇵'
    WHEN 'ko' THEN '🇰🇷'
    ELSE '🌐'
END;

ALTER TABLE languages
    ALTER COLUMN flag_emoji DROP DEFAULT;
