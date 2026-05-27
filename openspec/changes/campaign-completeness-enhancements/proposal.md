## Why

Villum has a deep data model for DnD campaign management but has critical gaps in both player-facing and DM-facing workflows that surface during actual gameplay. Players can't see encumbrance, track attunement limits, or batch-prepare spells after a rest. DMs lack encounter difficulty calculators, visualized faction reputations, and a session planning workflow. These gaps force users to track essential game info outside the app, undermining its value as a complete campaign companion.

## What Changes

**Player-facing additions:**
- Encumbrance system: STR-based carry capacity, coin weight, encumbered/heavily-encumbered states shown on inventory tab
- Attunement tracker: max-3 attuned items counter with visual indicator in inventory
- Exhaustion tracking: dedicated 6-level exhaustion UI on combat/stats tab with effects
- Spell preparation workflow: batch prepare/unprepare modal triggered on long rest
- Equipped items summary: collapsible "Loadout" panel showing currently equipped gear
- Party inventory & treasury: shared party loot pool with coin-split and item assignment
- Passive scores display: passive Investigation and Insight alongside existing Perception
- Magic item identification: identified/unidentified state toggle on magical items

**DM-facing additions:**
- Encounter difficulty calculator: XP thresholds → easy/medium/hard/deadly labels with party-level scaling
- Faction reputation visualization: reputation bars with tier labels (hostile/unfriendly/neutral/friendly/allied)
- Campaign session planner: DM-side session prep view (planned encounters, notes, player goals)
- Treasure & loot generator: hoard and individual treasure tables based on CR/level
- Campaign dashboard: single overview showing recent activity, upcoming events, pending items

**Breaking changes:** None. All additions are backward-compatible — new fields, new views, no schema migrations that break existing data.

## Capabilities

### New Capabilities
- `player-encumbrance`: STR-based carry capacity, coin weight calculation, encumbrance states
- `player-attunement`: Attunement slot tracker (max 3) with inventory integration
- `player-exhaustion`: 6-level exhaustion tracking with effect descriptions
- `player-spell-prep-workflow`: Batch spell preparation/unpreparation on rest
- `player-equipped-loadout`: Quick-view panel of all equipped gear
- `party-inventory`: Shared party loot pool, coin-split, item assignment
- `player-passive-scores`: Passive Investigation and Insight alongside Perception
- `player-magic-id`: Magic item identified/unidentified state management
- `dm-encounter-difficulty`: XP-based difficulty calculator with party scaling
- `dm-faction-reputation-viz`: Visual reputation bars with tier labels
- `dm-session-planner`: DM-side session preparation and notes view
- `dm-treasure-generator`: Loot tables and treasure hoard generation
- `dm-campaign-dashboard`: Consolidated campaign activity overview

### Modified Capabilities
- *(None — existing capabilities don't change behavior, they get additions)*

## Impact

- **Ent schema**: New fields on Character (exhaustion_level, encumbrance) and InventoryItem (is_identified). New tables for party_inventory, session_plans.
- **Handlers**: New or extended endpoints for encumbrance calc, spell batch-prep, party inventory, session planner, treasure gen, campaign dashboard.
- **Frontend (app.ts)**: New render functions for loadout panel, spell prep modal, exhaustion UI, reputation bars, campaign dashboard. Extended inventory, spells, faction views.
- **CSS (style.css)**: New styles for loadout card, exhaustion track, reputation bars, dashboard grid.
- **Tests**: New E2E tests for encumbrance display, attunement limit, spell prep flow, treasure gen, difficulty calculator.
