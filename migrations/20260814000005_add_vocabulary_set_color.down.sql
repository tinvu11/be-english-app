ALTER TABLE vocab_sets
    DROP CONSTRAINT IF EXISTS chk_vocab_sets_color_hex,
    DROP COLUMN IF EXISTS color_hex;
