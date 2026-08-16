ALTER TABLE users
    ADD COLUMN IF NOT EXISTS premium_until TIMESTAMPTZ NULL;

CREATE TABLE user_subscriptions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    platform VARCHAR(20) NOT NULL CHECK (platform IN ('ios', 'android')),
    product_id VARCHAR(100) NOT NULL,
    original_transaction_id TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'expired', 'in_grace_period', 'cancelled')),
    starts_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    auto_renew BOOLEAN NOT NULL DEFAULT TRUE,
    -- Store event time is distinct from database write time and prevents stale
    -- webhook deliveries from overwriting a newer state.
    last_event_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uq_user_active_subscription
    ON user_subscriptions (user_id)
    WHERE status IN ('active', 'in_grace_period');
CREATE INDEX idx_user_subscriptions_original_transaction_id
    ON user_subscriptions (original_transaction_id);
CREATE INDEX idx_user_subscriptions_user_status_expires
    ON user_subscriptions (user_id, status, expires_at);

CREATE TABLE iap_receipts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    platform VARCHAR(20) NOT NULL CHECK (platform IN ('ios', 'android')),
    product_id VARCHAR(100) NOT NULL,
    transaction_id TEXT NOT NULL UNIQUE,
    original_transaction_id TEXT NOT NULL,
    purchase_time TIMESTAMPTZ NOT NULL,
    expires_time TIMESTAMPTZ NOT NULL,
    raw_payload JSONB NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_iap_receipts_user_created_at
    ON iap_receipts (user_id, created_at DESC);

-- Durable inbox for at-least-once Apple/Google webhook delivery.
CREATE TABLE iap_webhook_events (
    provider VARCHAR(20) NOT NULL CHECK (provider IN ('apple', 'google')),
    event_id TEXT NOT NULL,
    event_time TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (provider, event_id)
);

