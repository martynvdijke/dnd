## Why

The app's data model has evolved organically and now needs restructuring to support DM-first workflows, proper separation of concerns between campaigns and parties, and richer one-shot editing capabilities including item uploads, NPC-item links, monster libraries, and player character linking.

## What Changes

- **Role System**: Introduce `normal | dm | admin` tier. DM created/manages one-shots. Normal users have zero one-shot access. Admin handles system config.
- **Party**: First-class entity separated from Campaign. Has name (renamable), file uploads, factions.
- **Campaign**: Loses calendar (removed entirely), factions (moved to Party), shops (moved to OneShot).
- **OneShot Editing**: Full inline editing - durations, acts/scenes, monster quick-add from library, item attachment with uploads, NPC linking to items, linked player characters.
- **NPC Tiers**: Simple (backstory only) or full (character-sheet-like stats/abilities).
- **Monster Library**: Both inline typed monsters per act/scene and reusable library monsters.
- **NPC↔Item Links**: Many-to-many relationship between NPCs and items within a one-shot.
- **Shops**: Moved from Campaign to OneShot/Adventure view.
- **Linked Player Characters**: DM can link player characters to one-shots.
- **File Uploads**: Items, party, and NPCs support file uploads.
- **Calendar**: Removed entirely.
- **Tree UI**: Visual act/scene tree with drag-reorder (SortableJS).

## Capabilities

### New Capabilities
- `role-system`: User role tiers (normal, dm, admin) with middleware enforcement
- `party-management`: First-class Party entity with uploads, factions, rename
- `one-shot-items`: Items within one-shots with file uploads and NPC linking
- `one-shot-shops`: Shops moved from campaign to one-shot context
- `one-shot-monsters`: Monster library + inline monsters per act/scene
- `one-shot-linked-pcs`: DM linking of player characters to one-shots
- `npc-tiers`: Simple vs full NPC character sheets
- `item-uploads`: File uploads attached to items
- `party-uploads`: File uploads attached to parties
- `one-shot-tree-ui`: Act/scene tree with inline editing and drag-reorder
- `campaign-cleanup`: Remove calendar, factions, shops from campaign

### Modified Capabilities
<!-- No existing specs to modify -->

## Impact

- **Models**: New tables for Party, ShopItem, Monster, MonsterLibrary, NPCItemLink, OneShotPC, UploadItem. Calendar removed. Faction/shop fields moved.
- **Auth**: New middleware `DMRequired()`, role field on User.
- **Routes**: One-shot editing routes become DM-only. New party CRUD routes.
- **Templates**: HTMX fragments for tree UI, inline editing, file upload modals.
- **JS**: SortableJS added for drag-reorder.
