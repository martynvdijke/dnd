## Context

Current codebase: Go + Gin backend with raw SQL (no Ent ORM used for custom queries), in-memory session auth, HTMX + TypeScript frontend. User model has `role` field (user/admin only). OneShotAdventure is tied to User and Campaign. Campaign has party_name, calendar, factions, shops embedded. NPCs are flat with 6 stats + HP. Uploads table exists but is detached (no FK relations). Tests use `main_test.go` with `buildRouter()` and testClient helper.

## Goals / Non-Goals

**Goals:**
- Role system: normal | dm | admin with middleware enforcement
- Party becomes standalone entity with factions, uploads, rename
- One-shot gains items, shops, monsters (library + inline), linked PCs
- NPCs become two-tier (simple backstory / full sheet)
- NPC↔Item linking within one-shots
- Campaign loses calendar, factions, shops
- Calendar removed entirely
- Inline-editable durations + visual act/scene tree
- File uploads on Party and Items
- Tests for all capabilities

**Non-Goals:**
- Calendar reimplementation (removed entirely)
- Full migration from raw SQL to Ent ORM
- Frontend framework migration (stays HTMX + TypeScript)

## Decisions

1. **New tables for all new entities** - No migration complexity: create `parties`, `one_shot_items`, `one_shot_monsters`, `one_shot_pcs`, `npc_backgrounds` tables. Keep existing tables with new optional columns or migration.
2. **Role enforcement via middleware** - New `DMRequired()` middleware alongside existing `AdminRequired()`. DM can see all users' characters for PC linking.
3. **Uploads become FK-linked** - Add `owner_type` (VARCHAR: "party"|"item"|"npc") and `owner_id` columns to `uploads` table.
4. **Shop ownership** - Existing `shops` table gets optional `oneshot_adventure_id` FK. Migration sets `campaign_id` to null for moved shops.
5. **Monster system** - New `monster_library` table (DM-owned, reusable) + `one_shot_scene_monsters` table for scene-specific inline monsters.
6. **Inline editing** - HTMX `hx-trigger="blur"` on duration/description fields with PUT to update endpoint.
7. **Act/scene tree** - Table-based display with number ordering. Drag-reorder via SortableJS + POST reorder endpoint.
8. **NPC tiers** - Add `npc_type` field to `npcs` table ("simple"|"full"). Simple uses existing fields + backstory. Full gets additional character-sheet fields or FK to character template.

## Risks / Trade-offs

- **Data migration risk**: Several tables change FK relationships (shops, factions). Rollback requires backup restore.
- **Test isolation**: Existing tests use shared test DB. New tests need proper cleanup.
- **Calendar removal**: Breaking change for any existing calendar data. Mitigation: archive to backup table.
- **Role expansion**: Existing users default to "user" role. DM assignment needs admin UI addition.
