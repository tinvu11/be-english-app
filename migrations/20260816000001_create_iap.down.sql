DROP TABLE IF EXISTS iap_webhook_events;
DROP TABLE IF EXISTS iap_receipts;
DROP TABLE IF EXISTS user_subscriptions;
ALTER TABLE users DROP COLUMN IF EXISTS premium_until;

