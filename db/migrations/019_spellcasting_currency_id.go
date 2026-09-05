package migrations

var migration019SQL = `
-- Add id column to character_spellcasting (currently character_id is PK, no id column)
CREATE TABLE IF NOT EXISTS character_spellcasting_v2 (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    ability TEXT NOT NULL DEFAULT '',
    save_dc INTEGER NOT NULL DEFAULT 10,
    attack_bonus INTEGER NOT NULL DEFAULT 0,
    slots_1_max INTEGER NOT NULL DEFAULT 0,
    slots_1_used INTEGER NOT NULL DEFAULT 0,
    slots_2_max INTEGER NOT NULL DEFAULT 0,
    slots_2_used INTEGER NOT NULL DEFAULT 0,
    slots_3_max INTEGER NOT NULL DEFAULT 0,
    slots_3_used INTEGER NOT NULL DEFAULT 0,
    slots_4_max INTEGER NOT NULL DEFAULT 0,
    slots_4_used INTEGER NOT NULL DEFAULT 0,
    slots_5_max INTEGER NOT NULL DEFAULT 0,
    slots_5_used INTEGER NOT NULL DEFAULT 0,
    slots_6_max INTEGER NOT NULL DEFAULT 0,
    slots_6_used INTEGER NOT NULL DEFAULT 0,
    slots_7_max INTEGER NOT NULL DEFAULT 0,
    slots_7_used INTEGER NOT NULL DEFAULT 0,
    slots_8_max INTEGER NOT NULL DEFAULT 0,
    slots_8_used INTEGER NOT NULL DEFAULT 0,
    slots_9_max INTEGER NOT NULL DEFAULT 0,
    slots_9_used INTEGER NOT NULL DEFAULT 0,
    UNIQUE(character_id)
);
INSERT OR IGNORE INTO character_spellcasting_v2 (character_id, ability, save_dc, attack_bonus,
    slots_1_max, slots_1_used, slots_2_max, slots_2_used, slots_3_max, slots_3_used,
    slots_4_max, slots_4_used, slots_5_max, slots_5_used, slots_6_max, slots_6_used,
    slots_7_max, slots_7_used, slots_8_max, slots_8_used, slots_9_max, slots_9_used)
SELECT character_id, ability, save_dc, attack_bonus,
    slots_1_max, slots_1_used, slots_2_max, slots_2_used, slots_3_max, slots_3_used,
    slots_4_max, slots_4_used, slots_5_max, slots_5_used, slots_6_max, slots_6_used,
    slots_7_max, slots_7_used, slots_8_max, slots_8_used, slots_9_max, slots_9_used
FROM character_spellcasting;
DROP TABLE IF EXISTS character_spellcasting;
ALTER TABLE character_spellcasting_v2 RENAME TO character_spellcasting;

-- Add id column to character_currency (currently character_id is PK, no id column)
CREATE TABLE IF NOT EXISTS character_currency_v2 (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    cp INTEGER NOT NULL DEFAULT 0,
    sp INTEGER NOT NULL DEFAULT 0,
    ep INTEGER NOT NULL DEFAULT 0,
    gp INTEGER NOT NULL DEFAULT 0,
    pp INTEGER NOT NULL DEFAULT 0
);
INSERT OR IGNORE INTO character_currency_v2 (character_id, cp, sp, ep, gp, pp)
SELECT character_id, cp, sp, ep, gp, pp FROM character_currency;
DROP TABLE IF EXISTS character_currency;
ALTER TABLE character_currency_v2 RENAME TO character_currency;
`
