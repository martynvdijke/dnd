package migrations

var migration057SQL = `
CREATE INDEX IF NOT EXISTS idx_timeline_linked ON campaign_timeline_events(linked_entity_type, linked_entity_id);
`
