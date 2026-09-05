package migrations

var migration038SQL = `
CREATE TABLE IF NOT EXISTS race_colors (
    race_name TEXT PRIMARY KEY,
    color TEXT NOT NULL DEFAULT '#6c757d'
);

INSERT OR IGNORE INTO race_colors (race_name, color) VALUES
    ('Human', '#4a90d9'),
    ('Elf', '#43b581'),
    ('Dwarf', '#f04747'),
    ('Halfling', '#faa61a'),
    ('Dragonborn', '#7289da'),
    ('Gnome', '#e91e63'),
    ('Half-Elf', '#9b59b6'),
    ('Half-Orc', '#e67e22'),
    ('Tiefling', '#c0392b'),
    ('Aarakocra', '#1abc9c'),
    ('Aasimar', '#f1c40f'),
    ('Bugbear', '#2c3e50'),
    ('Centaur', '#8e44ad'),
    ('Changeling', '#d35400'),
    ('Deep Gnome', '#7f8c8d'),
    ('Duergar', '#95a5a6'),
    ('Eladrin', '#3498db'),
    ('Fairy', '#e91e63'),
    ('Firbolg', '#27ae60'),
    ('Genasi', '#2980b9'),
    ('Gith', '#2c3e50'),
    ('Goblin', '#e74c3c'),
    ('Goliath', '#bdc3c7'),
    ('Harengon', '#f48fb1'),
    ('Kenku', '#795548'),
    ('Kobold', '#ff5722'),
    ('Leonin', '#ff9800'),
    ('Lizardfolk', '#4caf50'),
    ('Minotaur', '#3f51b5'),
    ('Orc', '#d32f2f'),
    ('Satyr', '#9c27b0'),
    ('Sea Elf', '#00bcd4'),
    ('Shadar-Kai', '#607d8b'),
    ('Shifter', '#ff6f00'),
    ('Tabaxi', '#ffc107'),
    ('Tortle', '#8d6e63'),
    ('Triton', '#00acc1'),
    ('Verdan', '#43a047'),
    ('Warforged', '#546e7a'),
    ('Yuan-Ti', '#1b5e20');
`
