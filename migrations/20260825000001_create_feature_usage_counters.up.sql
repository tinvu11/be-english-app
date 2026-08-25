CREATE TABLE feature_usage_counters (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    feature_key VARCHAR(100) NOT NULL,
    window_start TIMESTAMPTZ NOT NULL,
    used_count INTEGER NOT NULL DEFAULT 0 CHECK (used_count >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, feature_key, window_start)
);

CREATE INDEX idx_feature_usage_counters_window
    ON feature_usage_counters (feature_key, window_start);
