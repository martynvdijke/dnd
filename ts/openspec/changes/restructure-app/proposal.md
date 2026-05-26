## Why

The D&D app currently has a flat role system (user/admin) and a data model where one-shots, campaigns, parties, NPCs, and items are tightly coupled or missing critical features. Users need a tiered role system (normal/dm/admin), a restructured data model that properly separates concerns, and rich one-shot editing capabilities including inline durations, items with uploads, NPC↔item linking, monster libraries, and a visual act/scene tree.

## What Changes

1. **Role System** (normal | dm | admin) with DM-only one-shot access and admin system config
2. **Party** becomes a first-class entity (separate from Campaign) with renamable name, file uploads, and Factions moved here
3. **One-Shot** gains Items with file uploads, Shops (moved from Campaign), Monsters (both inline typed and from library), and linked Player Characters
4. **NPCs** become two-tier: simple (name + backstory) or full (character-sheet-like stats)
5. **NPC↔Item linking** within one-shots
6. **Campaign** loses calendar (removed), factions (→Party), shops (→OneShot)
7. **Calendar** removed entirely from the app
8. **Inline-editable durations** on one-shot acts/scenes
9. **Visual act/scene tree** with drag-reorder UX
10. **File uploads** on Party and Items (one-shot)

## Capabilities

### New Capabilities
- `role-system`: Role-based access control with normal, dm, admin tiers
- `party-management`: First-class Party entity with factions, uploads, renaming
- `one-shot-items`: Items within one-shots with file uploads
- `one-shot-shops`: Shops moved from Campaign to OneShot
- `one-shot-monsters`: Monster library + inline monster typing per act/scene
- `one-shot-pcs`: Linking player characters to one-shots
- `npc-tiers`: Two-tier NPC system (simple backstory or full character sheet)
- `npc-item-linking`: Linking NPCs to items within one-shots
- `campaign-cleanup`: Removing calendar, factions, shops from Campaign
- `one-shot-editing`: Inline-editable durations + visual act/scene tree with drag-reorder
- `file-uploads`: File upload attachment system for Party and Items

### Modified Capabilities
- *(none - no existing specs)*

## Impact

- **Data model**: Ent schema changes across User, Campaign, Party, OneShotAdventure, NPC, Item, Upload entities
- **API**: New routes for party CRUD, one-shot item/shop/monster/PC management, role management
- **UI**: New DM dashboard, one-shot editor overhaul with tree component, party management pages
- **Auth**: Middleware changes for role-based access control
- **Tests**: New test suites for all capabilities
