## Context

Villum is a Go + TypeScript DnD campaign folio with an Ent ORM backend, HTMX-driven frontend, and SQLite database. Current architecture has 55+ schema entities and 50+ API handlers. The existing UI is a single-page app using Bootstrap 5 with tabbed character sheets and top-level navigation for campaign tools.

The gaps identified all share a common pattern: the data model already has the foundation (e.g., STR → carry capacity, attunement bool on items, spell prepared flag, exhaustion as a condition type), but the UI doesn't surface these mechanics or provide the workflow glue. The additions are additive — no existing views break, no schema migrations require data backfills.

**Key architectural constraints:**
- SQLite database (single-file, no concurrent writers from app side)
- Ent ORM for schema/query generation
- HTMX for dynamic content loading (spells, features, etc.)
- TypeScript compiled to `static/js/app.js` via Vite
- No client-side state management library (plain TS with DOM manipulation)
- CSS custom properties theme via `data-theme` attribute

## Goals / Non-Goals

**Goals:**
- Players can see encumbrance status and manage carry capacity from inventory
- Players can track attunement limits (max 3) and see count in inventory
- Players can manage exhaustion levels (1-6) and see effects
- Players can batch-prepare/unprepare spells during a long rest
- Players can see a quick "Loadout" of all equipped items
- Party members can share inventory and split coins
- DMs can calculate encounter difficulty with party level input
- DMs can see faction reputations as visual bars with tier labels
- DMs can plan sessions with notes, encounters, and player goals
- DMs can generate treasure from loot tables
- DMs have a campaign dashboard showing recent activity

**Non-Goals:**
- Real-time sync or WebSocket-based collaboration (existing WS stays as-is)
- CRDT or offline-first architecture (service worker intentionally unregistered)
- Full stat-block rendering for monsters (out of scope — existing NPC view covers it)
- Dice roll automation integrated into combat tracker (existing dice engine is standalone)
- Print/PDF character sheet redesign (existing print view is functional)
- PWA or mobile app (responsive Bootstrap is sufficient for now)

## Decisions

### 1. Encumbrance calculated client-side from existing character data
- **Decision**: Compute encumbrance state in TypeScript from `character.str`, total inventory weight, and coin count (50 coins = 1 lb), rather than adding server-side endpoints.
- **Rationale**: The formula (STR × 15 = carry capacity, STR × 30 = push/drag) is simple and deterministic. No server round-trip needed. The data already arrives in the character API response including inventory items and currency.
- **Alternatives considered**: Server-side computed field → unnecessary overhead for a trivial calculation that changes on every inventory edit anyway.

### 2. Spell preparation as a modal workflow, not a new page
- **Decision**: A "Prepare Spells" button on the spells tab opens a modal showing all known spells with checkboxes for prepared status. On confirm, it batch-updates via a single API call.
- **Rationale**: Spell preparation happens once per long rest. A modal avoids context-switching and keeps the UX contained within the spells tab. The existing `Spell` entity already has a `prepared` boolean — this is just batch-set on it.
- **Alternatives considered**: Dedicated "Preparation" tab → too much navigation overhead for a twice-per-session action.

### 3. Party inventory uses a new `party_items` table and existing `character_currency`
- **Decision**: Create a lightweight `PartyItem` join table (party_id, item_name, quantity, notes) for shared loot. For coin-split, add a `party_share` field to `character_currency` or a mini party treasury table. This avoids coupling to the existing per-character `inventory` table which has weapon/armor-specific fields.
- **Rationale**: Party loot is typically untyped ("10 gp", "a rusty sword", "a gemstone") and doesn't need weapon properties or AC bonuses. Using a separate simple table avoids bloating the structured inventory model.
- **Alternatives considered**: Reusing InventoryItem with nullable fields → messier queries, harder to distinguish "my gear" from "party loot."

### 4. Encounter difficulty as a client-side calculator
- **Decision**: Implement difficulty thresholds (easy/medium/hard/deadly) based on DMG XP thresholds as a pure function in TypeScript, taking monster XP values and party level/composition.
- **Rationale**: The encounter builder already has `xp_budget` and `total_xp` fields. A client-side helper function adds the label with no server changes.
- **Alternatives considered**: Server-side endpoint → unnecessary for a lookup-table calculation that changes infrequently.

### 5. Campaign dashboard as a re-aggregation of existing data
- **Decision**: The dashboard is a new view (`showView('dashboard')`) that queries existing endpoints — recent recaps, upcoming calendar events, pending combat entries, recent dice rolls — and renders them in a Bootstrap card grid. No new API endpoint needed except optionally a dedicated `/api/campaign/dashboard` that aggregates.
- **Rationale**: Avoids N+1 client-side queries. A single aggregation endpoint is more efficient than 5 separate API calls on page load.
- **Alternatives considered**: Pure client-side aggregation → 5+ API calls on view load, slower UX.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Client-side encumbrance may drift if inventory edits happen from multiple browser tabs | Low risk — single-user app pattern. Future: reload on visibility change. |
| Batch spell prep API call could be slow for high-level wizards (50+ spells) | Use a single `PUT /api/characters/:id/spells/prepare` with array of spell IDs. Single query, no N+1. |
| Party inventory without real-time sync means stale views for other party members | Acceptable for now — party members typically play together and can refresh. |
| Campaign dashboard aggregation endpoint could become a query performance concern | SQLite with proper indexes. If slow, add dedicated dashboard materialization later. |
| Adding new schema entities requires `ent` code generation and migration | Run `go generate ./ent` and test migration with existing data backup. |
| Modal-based spell prep might feel heavy on mobile | Modal uses `modal-dialog-scrollable` class (already in base template). Stays within viewport. |
