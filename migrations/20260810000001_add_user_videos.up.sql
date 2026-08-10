ALTER TABLE videos
    ALTER COLUMN channel_id DROP NOT NULL,
    ALTER COLUMN level_id DROP NOT NULL;

CREATE TABLE user_videos (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    video_id BIGINT NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_user_videos_user_video UNIQUE (user_id, video_id)
);

CREATE INDEX idx_user_videos_user_created
    ON user_videos(user_id, created_at DESC);

CREATE INDEX idx_user_videos_video
    ON user_videos(video_id);
