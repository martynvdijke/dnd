package migrations

var migration069SQL = `
-- NOTE: campaign_telegram_settings is raw-managed (no ent model), so these
-- columns belong here and NOT in ApplySafeAlters (db/migrations/safe_alters.go).
-- safe_alters.go only re-adds columns on ent-owned tables whose schema
-- ent.Schema.Create would otherwise rebuild and drop.

ALTER TABLE campaign_telegram_settings ADD COLUMN auto_post_cadence_minutes INTEGER NOT NULL DEFAULT 0;
ALTER TABLE campaign_telegram_settings ADD COLUMN quiet_start_hour INTEGER;
ALTER TABLE campaign_telegram_settings ADD COLUMN quiet_end_hour INTEGER;
ALTER TABLE campaign_telegram_settings ADD COLUMN auto_post_last_at TEXT;
`
