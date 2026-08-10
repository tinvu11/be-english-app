CREATE TABLE vocab_sets (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL CHECK (length(trim(title)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_vocab_sets_id_user UNIQUE (id, user_id)
);

CREATE INDEX idx_vocab_sets_user_created
    ON vocab_sets(user_id, created_at DESC);

CREATE TABLE user_vocabularies (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vocab_set_id BIGINT,
    word VARCHAR(150) NOT NULL CHECK (length(trim(word)) > 0),
    phonetic_or_pinyin VARCHAR(150),
    part_of_speech VARCHAR(50),
    meaning TEXT NOT NULL CHECK (length(trim(meaning)) > 0),
    example_1_sentence TEXT,
    example_1_translation TEXT,
    example_2_sentence TEXT,
    example_2_translation TEXT,
    is_learned BOOLEAN NOT NULL DEFAULT FALSE,
    learned_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user_vocabularies_set_owner
        FOREIGN KEY (vocab_set_id, user_id)
        REFERENCES vocab_sets(id, user_id) ON DELETE CASCADE
);

CREATE INDEX idx_user_vocabularies_user_created
    ON user_vocabularies(user_id, created_at DESC);

CREATE INDEX idx_user_vocabularies_user_learned_created
    ON user_vocabularies(user_id, is_learned, created_at DESC);

CREATE INDEX idx_user_vocabularies_set_learned_created
    ON user_vocabularies(vocab_set_id, is_learned, created_at DESC);
