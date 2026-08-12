ALTER TABLE user_vocabularies
    ADD COLUMN word VARCHAR(150),
    ADD COLUMN phonetic_or_pinyin VARCHAR(150),
    ADD COLUMN part_of_speech VARCHAR(50),
    ADD COLUMN meaning TEXT,
    ADD COLUMN example_1_sentence TEXT,
    ADD COLUMN example_1_translation TEXT,
    ADD COLUMN example_2_sentence TEXT,
    ADD COLUMN example_2_translation TEXT;

UPDATE user_vocabularies uv
SET word = d.word,
    phonetic_or_pinyin = d.phonetic_or_pinyin,
    part_of_speech = d.part_of_speech,
    meaning = d.meaning,
    example_1_sentence = d.example_1_sentence,
    example_1_translation = d.example_1_translation,
    example_2_sentence = d.example_2_sentence,
    example_2_translation = d.example_2_translation
FROM dictionary d
WHERE d.id = uv.dictionary_id;

ALTER TABLE user_vocabularies
    ALTER COLUMN word SET NOT NULL,
    ALTER COLUMN meaning SET NOT NULL,
    ADD CONSTRAINT user_vocabularies_word_check CHECK (length(trim(word)) > 0),
    ADD CONSTRAINT user_vocabularies_meaning_check CHECK (length(trim(meaning)) > 0),
    DROP CONSTRAINT uq_user_vocab_set_dict,
    DROP CONSTRAINT fk_user_vocabularies_dictionary,
    DROP COLUMN dictionary_id,
    DROP COLUMN caption_id;

CREATE INDEX idx_user_vocabularies_user_learned_created
    ON user_vocabularies(user_id, is_learned, created_at DESC);

DROP TABLE dictionary;
