CREATE TABLE languages (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code VARCHAR(10) NOT NULL UNIQUE,
    name VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE levels (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code VARCHAR(20) NOT NULL,
    language_id INT NOT NULL REFERENCES languages(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_level_code_language UNIQUE (code, language_id),
    CONSTRAINT uq_levels_id_language UNIQUE (id, language_id)
);

CREATE TABLE level_translations (
    level_id INT NOT NULL REFERENCES levels(id) ON DELETE CASCADE,
    language_id INT NOT NULL REFERENCES languages(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    PRIMARY KEY (level_id, language_id)
);

CREATE TABLE topics (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    slug VARCHAR(100) NOT NULL UNIQUE,
    icon_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE topic_translations (
    topic_id INT NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    language_id INT NOT NULL REFERENCES languages(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    PRIMARY KEY (topic_id, language_id)
);

CREATE TABLE channels (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    channel_youtube_id VARCHAR(100) NOT NULL UNIQUE,
    channel_name VARCHAR(150) NOT NULL,
    avatar_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE users
    ADD COLUMN avatar_url TEXT,
    ADD COLUMN native_language_id INT REFERENCES languages(id) ON DELETE RESTRICT,
    ADD COLUMN target_language_id INT REFERENCES languages(id) ON DELETE RESTRICT,
    ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE,
    ADD CONSTRAINT chk_user_languages CHECK (
        native_language_id IS NULL
        OR target_language_id IS NULL
        OR native_language_id <> target_language_id
    );

CREATE INDEX idx_users_native_language ON users(native_language_id);
CREATE INDEX idx_users_target_language ON users(target_language_id);

CREATE TABLE videos (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    youtube_id VARCHAR(50) NOT NULL UNIQUE,
    thumbnail_url TEXT,
    duration_seconds INT NOT NULL DEFAULT 0 CHECK (duration_seconds >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'published',
    language_id INT NOT NULL REFERENCES languages(id) ON DELETE RESTRICT,
    level_id INT NOT NULL,
    channel_id INT NOT NULL REFERENCES channels(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_video_status CHECK (status IN ('draft', 'published', 'archived')),
    CONSTRAINT fk_videos_level_language
        FOREIGN KEY (level_id, language_id) REFERENCES levels(id, language_id)
);

CREATE INDEX idx_videos_lang_level ON videos(language_id, level_id, status);
CREATE INDEX idx_videos_channel ON videos(channel_id);

CREATE TABLE video_topics (
    video_id BIGINT NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    topic_id INT NOT NULL REFERENCES topics(id) ON DELETE RESTRICT,
    PRIMARY KEY (video_id, topic_id)
);

CREATE INDEX idx_video_topics_topic_id ON video_topics(topic_id);

CREATE TABLE video_captions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    video_id BIGINT NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    sentence_order INT NOT NULL CHECK (sentence_order >= 0),
    start_time_ms BIGINT NOT NULL CHECK (start_time_ms >= 0),
    end_time_ms BIGINT NOT NULL,
    content TEXT NOT NULL CHECK (length(trim(content)) > 0),
    pinyin_or_furigana TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_video_sentence_order UNIQUE (video_id, sentence_order),
    CONSTRAINT uq_video_captions_id_video UNIQUE (id, video_id),
    CONSTRAINT chk_caption_time CHECK (end_time_ms >= start_time_ms)
);

CREATE INDEX idx_video_captions_video_order ON video_captions(video_id, sentence_order);

CREATE TABLE caption_translations (
    caption_id BIGINT NOT NULL REFERENCES video_captions(id) ON DELETE CASCADE,
    language_id INT NOT NULL REFERENCES languages(id) ON DELETE CASCADE,
    translated_text TEXT NOT NULL CHECK (length(trim(translated_text)) > 0),
    PRIMARY KEY (caption_id, language_id)
);

CREATE TABLE user_watch_later (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    video_id BIGINT NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_user_watch_later UNIQUE (user_id, video_id)
);

CREATE INDEX idx_watch_later_user_created ON user_watch_later(user_id, created_at DESC);

CREATE TABLE user_watch_history (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    video_id BIGINT NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    last_position_seconds INT NOT NULL DEFAULT 0 CHECK (last_position_seconds >= 0),
    last_watched_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_user_watch_history UNIQUE (user_id, video_id)
);

CREATE INDEX idx_watch_history_user ON user_watch_history(user_id, last_watched_at DESC);

CREATE TABLE dictation_progress (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    video_id BIGINT NOT NULL,
    caption_id BIGINT NOT NULL,
    is_completed BOOLEAN NOT NULL DEFAULT TRUE,
    completed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, caption_id),
    CONSTRAINT fk_dictation_caption_video
        FOREIGN KEY (caption_id, video_id)
        REFERENCES video_captions(id, video_id) ON DELETE CASCADE
);

CREATE INDEX idx_dictation_user_video ON dictation_progress(user_id, video_id);

CREATE TABLE shadowing_attempts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    caption_id BIGINT NOT NULL REFERENCES video_captions(id) ON DELETE CASCADE,
    audio_url TEXT,
    accuracy_score NUMERIC(5,2) CHECK (accuracy_score BETWEEN 0 AND 100),
    fluency_score NUMERIC(5,2) CHECK (fluency_score BETWEEN 0 AND 100),
    completeness_score NUMERIC(5,2) CHECK (completeness_score BETWEEN 0 AND 100),
    pronunciation_score NUMERIC(5,2) CHECK (pronunciation_score BETWEEN 0 AND 100),
    azure_response_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_shadowing_user_caption ON shadowing_attempts(user_id, caption_id);
