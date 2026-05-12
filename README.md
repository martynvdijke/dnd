# dnd (villum)

A full-stack web application for Dungeons & Dragons and TTRPG character management.

## Compendium JSON Seed Data

The `data/` directory contains JSON files that seed the compendium database.
To customize for any TTRPG system:

1. Edit the JSON files in `data/` to match your system
2. Each entry supports `system` (e.g., "dnd5e", "pf2e", "generic") and `source` fields
3. Restart the app on a fresh database, or add a `.force` file in `data/` to force reload

## D&D 5e API Fallback

The app can fetch compendium data from the [D&D 5e API](https://www.dnd5eapi.co) as a fallback:
- `GET /api/compendium/api/:category?q=searchterm`
- Supports: equipment, spells, races, classes, monsters, features, magic-items
