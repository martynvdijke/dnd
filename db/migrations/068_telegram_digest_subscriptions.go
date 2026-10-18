package migrations

var migration068SQL = `
-- Weekly digest opt-ins are raw-owned (there is no ent model for this table),
-- so the table and its index belong in the versioned migration chain next to
-- the other Telegram tables. created_at records the opt-in time and
-- last_sent_at is the delivery claim written by the digest scanner.
CREATE TABLE IF NOT EXISTS telegram_digest_subscriptions (user_id INTEGER NOT NULL, campaign_id INTEGER NOT NULL, frequency TEXT NOT NULL DEFAULT 'weekly', last_sent_at TEXT, created_at TEXT NOT NULL, PRIMARY KEY(user_id, campaign_id));
CREATE INDEX IF NOT EXISTS idx_telegram_digest_subscriptions_campaign ON telegram_digest_subscriptions(campaign_id);
`
