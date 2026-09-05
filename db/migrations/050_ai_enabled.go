package migrations

var migration050SQL = `
-- Master switch for the LLM-backed AI features (text/image generation and
-- the AI endpoint test). Default enabled ('1') so existing behavior is
-- preserved; an admin can flip it off site-wide from the admin settings.
INSERT OR IGNORE INTO app_settings (key, value) VALUES ('ai_enabled', '1');
`
