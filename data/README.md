# Compendium Seed Data

## License & Attribution

The JSON files in this directory contain **game rules content** derived from the **Systems Reference Document (SRD) 5.1** for Dungeons & Dragons.

### SRD 5.1 License

This work includes material taken from the System Reference Document 5.1 ("SRD 5.1") by Wizards of the Coast LLC, available at https://dnd.wizards.com/resources/systems-reference-document. The SRD 5.1 is licensed under the **Creative Commons Attribution 4.0 International License** (CC-BY-4.0).

### Attribution

In accordance with CC-BY-4.0, you must include the following attribution when distributing or using this content:

> This work contains material from the System Reference Document 5.1 ("SRD 5.1") by Wizards of the Coast LLC, used under the CC-BY-4.0 license. The SRD 5.1 is available at https://dnd.wizards.com/resources/systems-reference-document.

### Trademark Notice

**Dungeons & Dragons**, **D&D**, **Wizards of the Coast**, and their respective logos are registered trademarks of Wizards of the Coast LLC. This project is not affiliated with, endorsed by, or sponsored by Wizards of the Coast.

The use of game rules, terminology, and mechanics names (such as spell names, class names, race names, etc.) from the SRD is permitted under the CC-BY-4.0 license. No copyright or trademarked character names, setting details, or proprietary lore are included.

## How Seeding Works

On every startup, the app runs `db.Seed()` which:

1. Checks each category for a JSON file in `data/` (e.g., `data/monsters.json`)
2. If a JSON file exists AND the corresponding DB table is empty → loads from JSON
3. If a JSON file doesn't exist or fails to load → falls back to built-in Go struct data
4. Skips seeding if the table already has rows (prevents re-seeding on restarts)

To force re-seeding, create an empty file named `.force` in the `data/` directory — this clears all data and reloads from JSON on next restart.

## Supported JSON Files

Place these files in the `data/` directory:

| File | DB Table | Fields |
|------|----------|--------|
| `monsters.json` | `compendium_monsters` | name, type, size, ac, hp, str, dex, con, int, wis, cha, cr, source, is_full, saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities, condition_immunities, senses, languages, special_abilities, actions, legendary_actions, description |
| `races.json` | `compendium_races` | name, description, speed, size, ability_bonuses, traits, languages, source_page, system, source |
| `classes.json` | `compendium_classes` | name, description, hit_die, primary_ability, saving_throws, proficiencies, spellcasting_ability, source_page, system, source |
| `spells.json` | `compendium_spells` | name, level, school, casting_time, range, components, duration, description, higher_levels, classes, source_page, system, source |
| `backgrounds.json` | `compendium_backgrounds` | name, description, feature_name, feature_description, proficiencies, source_page, system, source |
| `equipment.json` | `compendium_equipment` | name, category, cost, weight, description, source_page, system, source |

### Monster JSON Format (`monsters.json`)

```json
[
  {
    "name": "Goblin",
    "type": "humanoid",
    "size": "Small",
    "ac": 15,
    "hp": 7,
    "str": 8,
    "dex": 14,
    "con": 10,
    "int": 10,
    "wis": 8,
    "cha": 8,
    "cr": "1/4",
    "source": "SRD",
    "is_full": 1,
    "saves": "",
    "skills": "Stealth +6",
    "damage_vulnerabilities": "",
    "damage_resistances": "",
    "damage_immunities": "",
    "condition_immunities": "",
    "senses": "darkvision 60 ft.",
    "languages": "Common, Goblin",
    "special_abilities": "Nimble Escape: The goblin can take the Disengage or Hide action as a bonus action on each of its turns.",
    "actions": "Scimitar. Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 5 (1d6+2) slashing.\nShortbow. Ranged Weapon Attack: +4 to hit, range 80/320 ft., one target. Hit: 5 (1d6+2) piercing.",
    "legendary_actions": "",
    "description": "Goblins are small, black-hearted humanoids that dwell in caves and dark forests."
  }
]
```

**Field notes:**
- `cr` (challenge rating): string — e.g. `"0"`, `"1/8"`, `"1/4"`, `"1/2"`, `"1"`, `"10"`, etc.
- `is_full`: `1` for fully detailed monster (has actions/abilities), `0` for minimal entry
- `saves`/`skills`: comma-separated with values, e.g. `"Stealth +6, Perception +3"`
- Damage/condition fields: comma-separated lists, e.g. `"fire, poison"` or `"bludgeoning, piercing, slashing (from nonmagical attacks)"`
- `special_abilities`, `actions`, `legendary_actions`: plain text, newlines for formatting
- All string fields default to `""`, numeric fields default to `10` (stats) / `10` (AC) / `1` (HP)

### Other JSON Formats

All other JSON files follow a similar `[{...}, {...}]` array structure. Each entry supports optional `system` (default `"dnd5e"`) and `source` (default `"srd"` or `"custom"`) fields for multi-system compatibility.

## Docker Integration

### Important: `data/` is NOT included in the Docker image

The `Dockerfile` only copies the binary and `static/` directory. The `data/` JSON files are **not** included. This means when running via Docker, JSON seeding will silently fall back to built-in Go data.

### Option A: Mount `data/` as a volume (recommended for customization)

```yaml
services:
  villum:
    image: ghcr.io/martynvdijke/dnd:latest
    volumes:
      - ./data:/app/data           # <-- mount your JSON files
      - villum-data:/db
      - villum-media:/app/media
```

Then place your custom JSON files in `./data/` on the host. On container start, the app will find them and seed from JSON.

### Option B: Extend the Docker image with custom JSON

```dockerfile
FROM ghcr.io/martynvdijke/dnd:latest
COPY monsters.json /app/data/monsters.json
```

### Option C: Built-in Go data (default, no JSON needed)

If you don't provide JSON files, the app uses Go structs compiled into the binary. For monsters, the built-in data includes ~18 common monsters (Goblin, Orc, Zombie, Dragon, Beholder, etc.). The `data/` JSON files ship with a more extensive set (~147 monsters).

## Seeding Behavior

| Scenario | Behavior |
|----------|----------|
| First start, JSON file exists | Seeds from JSON |
| First start, no JSON file | Seeds from built-in Go data |
| Restart, data already exists | Skips (no re-seed) |
| Restart with `data/.force` file | Clears table, re-seeds from JSON |
| JSON added after first start | Ignored (table already has data, unless `.force` exists) |

To add new monsters after the initial seed, either:
- Delete the DB and restart (loses all data!)
- Or use the in-app monster import via the compendium API import modal
