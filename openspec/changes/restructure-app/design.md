## Context

The app uses Go + Gin backend, raw SQL (no Ent ORM in handlers — Ent is used only for schema generation/migration), HTMX frontend with embedded Go templates, and vanilla TypeScript (app.ts). SQLite database. Current role system is `user`/`admin` only. One-shots are user-owned via raw SQL. Uploads are detached (no owner relationships). Campaign has calendar, factions, shops.

## Goals / Non-Goals

**Goals:**
- Role system: `normal | dm | admin` with middleware enforcement
- Party as first-class entity (separated from Campaign), with uploads, factions, rename
- One-shots become DM-only, gain shops, items (with uploads), monsters, linked PCs
- NPC tier system: simple (backstory only) or full (character-sheet-like)
- NPC↔Item linking within one-shots
- Remove calendar from Campaign entirely
- Move factions from Campaign to Party
- Move shops from Campaign to OneShot
- Inline-editable durations on acts/scenes (HTMX)
- SortableJS visual act/scene tree with drag-reorder
- File uploads attached to items, party, NPCs

**Non-Goals:**
- No Ent ORM changes — all new tables use raw SQL via existing pattern in models/oneshot.go
- No React/Vue — HTMX + vanilla TS stays
- No migration of existing one-shots — schema additions are backward-compatible
- No normal user access to one-shots (DM-only enforced)

## Decisions

### 1. Raw SQL for new tables (not Ent)
All current one-shot data uses raw SQL in handlers/oneshot.go. New tables (Party, OneShotItems, Monsters, NPCItemLinks, OneShotPCs) will follow the same raw-SQL pattern. The Upload schema stays in Ent but gets owner_type/owner_id columns.

### 2. Polymorphic file uploads
Add `owner_type` (string: "item", "party", "npc") and `owner_id` (int64) columns to the existing `uploads` table. The existing Upload Ent schema gets new fields. Upload handler checks ownership and scopes access.

### 3. DMRequired() middleware
New middleware function (pattern from AdminRequired). Checks `role` context value is "dm" or "admin". Applied to one-shot routes. Admin inherits DM privileges.

### 4. Party as first-class entity
New `parties` table: id, user_id, name, description, created_at. Factions get a `party_id` foreign key (nullable, existing `campaign_id` kept nullable for migration period). Files uploaded to party via polymorphic uploads.

### 5. One-Shot Items
New `oneshot_items` table: id, oneshot_adventure_id, name, description, category, quantity, weight, price_gp, is_magical, attunement, notes, created_at. Supports file uploads via polymorphic owner_type="item".

### 6. One-Shot Shops
Shops gain an `oneshot_adventure_id` nullable foreign key (alongside existing `campaign_id`). Shop CRUD moves to DM-only group. UI routing shows shops under one-shot view.

### 7. Monsters
Two-tier: inline monsters `oneshot_monsters` (id, adventure_id, act_id, scene_id, name, ac, hp, str/dex/con/int/wis/cha, cr, source) and library monsters `monster_library` (id, user_id, name, same stats, description, created_at). Both have is_simple flag vs full-stat mode.

### 8. NPC↔Item Links
New join table `npc_item_links`: id, npc_id, oneshot_adventure_id, item_id, relationship_type ("owns", "knows", "wields", etc), notes. Ensures NPCs link to items within a one-shot context.

### 9. Linked Player Characters
New join table `oneshot_player_characters`: id, oneshot_adventure_id, character_id, role, notes. DM can browse all characters across all users and link them.

### 10. Inline Editing
HTMX pattern: acts/scenes have `hx-trigger="blur"` on duration fields, `hx-patch` to partial update endpoints. New routes: PATCH /oneshot-acts/:id/duration, PATCH /oneshot-scenes/:id/duration.

### 11. Act/Scene Tree UI
SortableJS used for drag-reorder. HTMX fragments for the tree: GET /oneshot-adventures/:id/tree returns a reorderable list of acts with nested scenes. Each item has inline controls.

### 12. Campaign Cleanup
- `campaign_calendar_events` table: no removal (data preserved), but UI routes removed
- `calendar` fields removed from Campaign Go structs/templates
- Factions: UI views no longer show under Campaign, moved to Party UI
- Shops: UI views no longer show under Campaign, moved to OneShot

## Risks / Trade-offs

- **Polymorphic uploads**: Simple to implement but no FK constraints. Mitigation: application-level validation in upload handlers.
- **Raw SQL for new tables**: No Ent-generated CRUD. Acceptable since one-shot code already uses raw SQL. Mitigation: keep SQL patterns consistent in models/ package.
- **SortableJS dependency**: Adds 12KB JS bundle. Small enough to inline or load from CDN.
- **Data migration**: Existing campaigns with factions/shops need manual migration. Mitigation: backward-compatible schema (old FK columns kept nullable).
