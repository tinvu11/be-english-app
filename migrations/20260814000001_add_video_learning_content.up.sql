CREATE TABLE video_summaries (
    video_id BIGINT NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    language_id INT NOT NULL REFERENCES languages(id) ON DELETE CASCADE,
    content TEXT NOT NULL CHECK (length(trim(content)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (video_id, language_id)
);

CREATE TABLE video_quizzes (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    video_id BIGINT NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    display_order INT NOT NULL CHECK (display_order >= 1),
    question TEXT NOT NULL CHECK (length(trim(question)) > 0),
    correct_option INT NOT NULL CHECK (correct_option BETWEEN 0 AND 3),
    explanation TEXT NOT NULL CHECK (length(trim(explanation)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (video_id, display_order)
);

CREATE TABLE video_quiz_options (
    quiz_id BIGINT NOT NULL REFERENCES video_quizzes(id) ON DELETE CASCADE,
    display_order INT NOT NULL CHECK (display_order BETWEEN 0 AND 3),
    content TEXT NOT NULL CHECK (length(trim(content)) > 0),
    PRIMARY KEY (quiz_id, display_order)
);

CREATE TABLE video_dictionary_entries (
    video_id BIGINT NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    target_language_id INT NOT NULL REFERENCES languages(id) ON DELETE CASCADE,
    dictionary_id BIGINT NOT NULL REFERENCES dictionary(id) ON DELETE CASCADE,
    display_order INT NOT NULL CHECK (display_order >= 1),
    PRIMARY KEY (video_id, target_language_id, dictionary_id),
    UNIQUE (video_id, target_language_id, display_order)
);

CREATE INDEX idx_video_dictionary_entries_dictionary
    ON video_dictionary_entries(dictionary_id);

CREATE FUNCTION invalidate_video_learning_content() RETURNS trigger AS $$
DECLARE
    affected_video_id BIGINT;
BEGIN
    affected_video_id := COALESCE(NEW.video_id, OLD.video_id);
    DELETE FROM video_summaries WHERE video_id = affected_video_id;
    DELETE FROM video_quizzes WHERE video_id = affected_video_id;
    DELETE FROM video_dictionary_entries WHERE video_id = affected_video_id;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_invalidate_video_learning_content
    AFTER INSERT OR UPDATE OR DELETE ON video_captions
    FOR EACH ROW EXECUTE FUNCTION invalidate_video_learning_content();
