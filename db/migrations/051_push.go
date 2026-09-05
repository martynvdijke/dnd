package migrations

var migration051SQL = `
-- Web Push subscriptions: one row per browser subscription endpoint.
-- A user may have several (one per device/browser profile). The endpoint is
-- the push service URL and is unique; re-subscribing upserts in place.
CREATE TABLE IF NOT EXISTS push_subscriptions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	endpoint TEXT NOT NULL UNIQUE,
	p256dh TEXT NOT NULL DEFAULT '',
	auth TEXT NOT NULL DEFAULT '',
	expiration_time INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_push_subscriptions_user ON push_subscriptions(user_id);

-- Per-campaign mute: a muted user receives no pushes for that campaign
-- (recap fan-out and session reminders both honor it).
CREATE TABLE IF NOT EXISTS push_mutes (
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	UNIQUE(user_id, campaign_id)
);
CREATE INDEX IF NOT EXISTS idx_push_mutes_campaign ON push_mutes(campaign_id);

-- Dedup marker for session reminders, keyed by source + event id so each
-- event reminds at most once regardless of how many members are subscribed.
CREATE TABLE IF NOT EXISTS push_reminder_log (
	source TEXT NOT NULL,
	event_key TEXT NOT NULL,
	sent_at TEXT NOT NULL DEFAULT (datetime('now')),
	UNIQUE(source, event_key)
);
`
