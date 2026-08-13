# Villum — TTRPG Campaign & Character Manager

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?style=flat&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/Gin-1.12-00ADD8?style=flat&logo=go" alt="Gin">
  <img src="https://img.shields.io/badge/SQLite3_FTS5-003B57?style=flat&logo=sqlite" alt="SQLite FTS5">
  <img src="https://img.shields.io/badge/version-2.13.0-blue" alt="Version">
  <img src="https://img.shields.io/badge/docker-ready-2496ED?style=flat&logo=docker" alt="Docker">
</p>

A full-stack web application for Dungeons & Dragons and TTRPG campaign management. Features comprehensive character management, campaign tracking, combat tools, world-building, and AI-assisted content generation.

## Features

### Character Management
- **Full Character CRUD** — Create and manage characters with all D&D 5e stats (abilities, skills, HP, AC, speed, hit dice)
- **Multi-Class Support** — Characters can have multiple classes with individual levels
- **Spellcasting** — Full spell slot tracking, spell preparation, and known spells management
- **Inventory** — Equipment, weapons, armor, and item management with currency tracking
- **Features & Traits** — Class features, racial traits, and feats with descriptions
- **Proficiencies** — Armor, weapon, tool, skill, and saving throw proficiencies
- **HP Auto-Calc** — Automatic HP calculation based on class, level, and CON modifier
- **Rest System** — Short and long rest with resource recovery
- **Level Up** — Level advancement with automatic HP and feature progression
- **Exhaustion Tracking** — Track and update exhaustion levels
- **Conditions/Ailments** — Condition management with tick system for duration tracking
- **Concentration Checks** — Automatic concentration save handling
- **Character Resources** — Class-specific resources (ki points, rage, etc.) with rest recovery
- **Character Export/Import** — Export characters as JSON, import from JSON
- **Character Comparison** — Side-by-side comparison of multiple characters
- **Character Graph** — Visual relationship graph for characters

### Campaign Management
- **Campaign CRUD** — Create and manage multiple campaigns
- **Campaign Members** — Add/remove members with role-based permissions (DM, player)
- **Campaign Dashboard** — Overview of campaign status, members, and recent activity
- **Campaign Wiki** — Editable wiki pages per campaign
- **Campaign Maps** — Upload maps with fog of war, pins, and active map setting
- **Campaign Recaps** — Write and AI-generate session recaps with email sending
- **Campaign NPCs** — Link NPCs to campaigns with custom relationships
- **Party Inventory** — Shared campaign item tracking
- **Factions & Reputation** — Track factions, reputation scores, and relationships
- **Campaign One-Shots** — Link one-shot adventures to campaigns with reordering
- **Campaign Graph** — Relationship mapping for campaign entities

### Combat & Encounters
- **Combat Tracker** — Initiative order, turn management, HP tracking, status effects
- **Encounter Builder** — Build encounters with monsters from compendium/library
- **Encounter XP Calculation** — Auto-calculate XP difficulty thresholds
- **Combat Log** — Detailed combat event logging with stats
- **Monster Library** — Custom monster entries with full stat blocks

### World-Building
- **Locations** — CRUD for locations with character and NPC associations
- **NPCs** — Full NPC management with interaction logging
- **Quests** — Quest tracking with status updates
- **Journal** — Character journal entries for session notes
- **Timeline** — Campaign timeline events
- **Conditions Library** — Manage custom conditions and ailments
- **Companions** — Animal companions, familiars, and mounts
- **Feats** — Custom feat creation and management
- **Downtime Activities** — Track character downtime with day advancement

### One-Shot Adventure Builder
- **Full Adventure CRUD** — Create one-shot adventures with acts and scenes
- **Scene Management** — Scene durations, dialog ordering, and reordering
- **Pacing System** — Start/pause/resume pacing sessions, track time per scene
- **Clue/Mystery Tracker** — Create clues with dependencies, NPC/location links, reveal/hide
- **Prep Checklist** — Track preparation tasks per adventure
- **DM Screen** — Quick reference for adventure details during sessions
- **DM Notes** — Private notes per adventure
- **Linked Entities** — Link NPCs, locations, encounters, characters, items, monsters, shops per adventure
- **NPC-Item Links** — Track which NPCs carry which items
- **Act-Level Monsters** — Assign monsters to acts and scenes

### Shops & Economy
- **Shop Management** — Create shops with item listings and pricing
- **Buy/Sell System** — Characters can buy from and sell to shops
- **Transaction Log** — Complete purchase/sale history

### Random Generators
- NPC generation
- Name generation
- Encounter generation
- Loot generation
- Random character generation
- Adventure hook generation
- Dungeon dressing
- Tavern generation
- Urban and road encounters
- Weather generation

### AI Integration
- **Text Generation** — AI-powered content generation for recaps, descriptions, etc.
- **Image Generation** — AI image generation for characters, locations, etc.
- **AI Endpoint Management** — Configure multiple AI providers with API keys
- **Campaign Recap Generation** — AI-generated session recaps from notes

### Dice Rolling
- **Full Dice Engine** — Roll any dice notation (3d6, 2d20+5, etc.)
- **Check Rolls** — Ability checks with automatic success/failure
- **Roll History** — View past dice roll results
- **Initiative Rolling** — Automatic initiative for combat

### Compendium System
- **Dynamic Schemas** — Define custom compendium types with flexible field schemas
- **Compendium Entries** — Full CRUD with FTS5 full-text search
- **Bulk Operations** — Batch delete, batch update compendium entries
- **Import/Export** — CSV/JSON import with field mapping, auto field type detection, export to JSON
- **Import Logs** — Track import history with rollback support
- **D&D 5e API Fallback** — Fetch equipment, spells, races, classes, monsters, features, magic items
- **Preloaded Compendium** — JSON seed data in `data/` directory
- **Compendium Schemas** — Built-in schemas for Races, Classes, Spells, Equipment, Monsters, Feats, Backgrounds

### Admin Panel
- **User Management** — Create, edit, delete users with password reset
- **Email Settings** — SMTP configuration, test emails, campaign highlight emails
- **AI Endpoint Management** — Full CRUD for AI provider endpoints with testing
- **OTel Settings** — OpenTelemetry configuration
- **Umami Analytics** — Analytics settings
- **Application Logs** — Central log viewer with level controls
- **Shop Management** — Global shop and shop item management
- **Compendium Schema Management** — Define schemas, manage entries, bulk operations

### Observability
- **OpenTelemetry Tracing** — Request-level tracing to OTLP backends
- **Prometheus Metrics** — `/metrics/prometheus` endpoint
- **Structured Logging** — slog-based logging with OTel log export
- **Health Check** — `/healthz` endpoint

### Additional Features
- **Media Uploads** — Image upload with cropping support, upload links
- **Share Links** — Create shareable links for characters and entities
- **WebSocket** — Real-time updates for collaborative sessions
- **Backup System** — Configurable automated backups with scheduler
- **Pregenerated Characters** — Create and manage pregens with party balance checker
- **Level-Up Planner** — Plan future levels with suggestions
- **Character Print View** — Print-optimized character sheet view
- **Homebrew Content Manager** — Custom homebrew content with type-based organization
- **Session Plans** — Plan upcoming sessions with agenda and notes
- **Quick Reference** — Rules quick reference for common lookups
- **Party View** — See all party members' status at a glance

## Quick Start

### Docker (Recommended)

```bash
docker compose up -d
```

Open **[http://localhost:6280](http://localhost:6280)** in your browser.

### Manual Setup

```bash
# Install dependencies
go mod download

# Build with FTS5 support
CGO_ENABLED=1 go build -tags fts5 -o villum-server .

# Run
./villum-server
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `6270` | HTTP listen port |
| `DB_PATH` | `villum.db` | SQLite database path |
| `MEDIA_PATH` | `{DB_DIR}/media` | Media upload directory |
| `DOCKER` | `false` | Docker-specific paths (auto in Dockerfile) |

## Project Structure

```
dnd/
├── main.go                    # Application entry point & route setup
├── main_test.go               # Tests
├── handlers/                  # HTTP handlers organized by domain
│   ├── auth.go                # Authentication
│   ├── characters.go          # Character CRUD
│   ├── spells.go              # Spell management
│   ├── combat.go              # Combat tracker
│   ├── campaigns.go           # Campaign management
│   ├── compendium.go          # Compendium & API fallback
│   ├── npcs.go                # NPC management
│   ├── locations.go           # Location management
│   ├── shops.go               # Shop & economy
│   ├── dice.go                # Dice rolling
│   ├── generators.go          # Random generators
│   ├── ai.go                  # AI integration
│   ├── backup.go              # Backup system
│   └── ...                    # Many more handler files
├── db/                        # Database initialization & queries
├── models/                    # Data structures
├── middleware/                 # Auth, CSRF, security, logging
├── dice/                      # Dice engine
├── crypto/                    # Crypto utilities
├── static/                    # Frontend assets (embedded via EmbedFS)
├── ts/                        # TypeScript source
├── data/                      # Compendium seed data (JSON)
├── scripts/                   # Utility scripts
├── Dockerfile                 # Multi-stage Docker build (Node + Go + Alpine)
├── docker-compose.yml         # Docker Compose configuration
└── go.mod / go.sum            # Go module dependencies
```

## Compendium JSON Seed Data

The `data/` directory contains JSON files that seed the compendium database. To customize for any TTRPG system:

1. Edit the JSON files in `data/` to match your system
2. Each entry supports `system` (e.g., "dnd5e", "pf2e", "generic") and `source` fields
3. Restart the app on a fresh database, or add a `.force` file in `data/` to force reload

### D&D 5e API Fallback

The app can fetch compendium data from the [D&D 5e API](https://www.dnd5eapi.co) as a fallback:
- `GET /api/compendium/api/:category?q=searchterm`
- Supports: equipment, spells, races, classes, monsters, features, magic-items

## TRMNL e-ink Display Plugin

[TRMNL](https://usetrmnl.com) is a dedicated e-ink display that polls JSON endpoints and renders Liquid templates on-device — no browser needed. Villum ships a plugin (in `trmnl/`) that shows character stats or campaign progress at a glance.

### Endpoints

Both endpoints are **read-only** and require the TRMNL access token as a query parameter:

- `GET /api/trmnl/character-stats?token=<token>&character_id=<id>` — character name, race, class, level, XP, HP, AC, initiative, and the six ability scores with modifiers
- `GET /api/trmnl/campaign-stats?token=<token>&character_id=<id>` — session count, total XP/gold earned, quest and rest breakdowns, top NPCs, and dice roll summary

Responses are `401` without a valid token and `404` for unknown `character_id`.

### Token setup

1. Log in to Villum as an admin and open **Admin → TRMNL** tab
2. The current token is displayed there — use **Regenerate** to issue a new one (the old token immediately stops working)
3. The token is stored site-wide in `app_settings` and only readable by admins

### Installing the plugin

1. In the TRMNL web app, create a new plugin and upload the files from `trmnl/` (all four layouts: `full`, `half_horizontal`, `half_vertical`, `quadrant`) and `settings.yml`
2. Set the custom fields:
   - `url` — your Villum instance base URL (e.g., `https://dnd.example.com`)
   - `token` — the TRMNL access token from the admin panel
   - `character_id` — the character to display
   - `display_mode` — `character` for character stats, `campaign` for campaign progress
3. The plugin polls daily by default (`refresh_interval: 1440` minutes); adjust in `settings.yml` if needed

## License

MIT
