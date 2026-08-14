ALTER TABLE vocab_sets
    ADD COLUMN color_hex VARCHAR(7) NOT NULL DEFAULT '#3B82F6',
    ADD CONSTRAINT chk_vocab_sets_color_hex
        CHECK (color_hex ~ '^#[0-9A-F]{6}$');
