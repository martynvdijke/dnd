package migrations

// migration070SQL adds the ink capture inbox: handwritten pages that are
// transcribed and later promoted into a recap (DM) or a journal entry (player).
// Pages reference ordinary uploads so the existing media pipeline does the
// resizing, thumbnailing, and deduplication.
var migration070SQL = `
CREATE TABLE IF NOT EXISTS ink_captures (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL,
    character_id INTEGER,
    owner_user_id INTEGER NOT NULL,
    session_key TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT 'share' CHECK(source IN ('share','canvas')),
    status TEXT NOT NULL DEFAULT 'uploaded' CHECK(status IN ('queued','uploaded','transcribed','reviewed','promoted','failed')),
    ocr_source TEXT NOT NULL DEFAULT 'none' CHECK(ocr_source IN ('none','boox','vision','manual')),
    journal_entry_id INTEGER,
    recap_id INTEGER,
    error TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS ink_pages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    capture_id INTEGER NOT NULL REFERENCES ink_captures(id) ON DELETE CASCADE,
    page_index INTEGER NOT NULL DEFAULT 0,
    upload_id INTEGER,
    ocr_text TEXT NOT NULL DEFAULT '',
    ocr_source TEXT NOT NULL DEFAULT 'none' CHECK(ocr_source IN ('none','boox','vision','manual')),
    sha256 TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_ink_captures_campaign ON ink_captures(campaign_id);
CREATE INDEX IF NOT EXISTS idx_ink_captures_character ON ink_captures(character_id);
CREATE INDEX IF NOT EXISTS idx_ink_captures_owner ON ink_captures(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_ink_pages_capture ON ink_pages(capture_id);
`
