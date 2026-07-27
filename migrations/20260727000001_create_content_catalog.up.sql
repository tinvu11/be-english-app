CREATE TABLE topics (
    id BIGSERIAL PRIMARY KEY, name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE, icon TEXT, description TEXT,
    sort_order INT NOT NULL DEFAULT 0, is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE levels (
    id SERIAL PRIMARY KEY, code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL, sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE channels (
    id BIGSERIAL PRIMARY KEY, channel_id VARCHAR(255) NOT NULL UNIQUE,
    channel_name VARCHAR(255) NOT NULL, thumbnail_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE videos (
    id BIGSERIAL PRIMARY KEY, video_id VARCHAR(255) NOT NULL UNIQUE,
    channel_id BIGINT NOT NULL REFERENCES channels(id) ON DELETE RESTRICT,
    title VARCHAR(255) NOT NULL, description TEXT, thumbnail_url TEXT,
    view_count BIGINT NOT NULL DEFAULT 0 CHECK (view_count >= 0),
    duration INT NOT NULL DEFAULT 0 CHECK (duration >= 0),
    sort_order INT NOT NULL DEFAULT 0, is_active BOOLEAN NOT NULL DEFAULT TRUE,
    published_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE video_topics (
    video_id BIGINT NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    topic_id BIGINT NOT NULL REFERENCES topics(id) ON DELETE RESTRICT,
    PRIMARY KEY (video_id, topic_id)
);
CREATE TABLE video_levels (
    video_id BIGINT NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    level_id INT NOT NULL REFERENCES levels(id) ON DELETE RESTRICT,
    PRIMARY KEY (video_id, level_id)
);
CREATE INDEX videos_channel_id_idx ON videos(channel_id);
CREATE INDEX video_topics_topic_id_idx ON video_topics(topic_id);
CREATE INDEX video_levels_level_id_idx ON video_levels(level_id);
