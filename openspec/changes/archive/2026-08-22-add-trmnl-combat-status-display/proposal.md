## Why

The TRMNL plugin already ships two display modes (character/campaign stats and a party roster), but neither helps the DM *run a fight at the table*. During combat everyone leans over the DM's laptop or asks "whose turn is it?" A dedicated e-ink screen showing the live initiative order — who's active, who's hurt, who's condition-locked — keeps the table moving without touching the web app.

## What Changes

- Add a public read-only polling endpoint `GET /api/trmnl/combat` (token-guarded) returning the current combat tracker state for a campaign: entries sorted by initiative order with name, type, initiative total, HP current/max, AC, active-turn flag, and resolved condition names
- Add a `combat` display mode to the TRMNL plugin: new `polling_url` entry and Liquid blocks that render the initiative ladder in a compact, high-contrast layout across all four screen sizes, with the active combatant visually distinct
- Render an explicit idle state ("No active combat") when the campaign has no combat entries, so a stale screen is never ambiguous
- Add unit tests for the new endpoint following the existing TRMNL handler test pattern

## Capabilities

### New Capabilities

- `trmnl-combat-status`: Public read-only JSON endpoint exposing the live combat tracker state, plus a combat display mode in the TRMNL Liquid templates and settings

### Modified Capabilities

- *(none)*

## Impact

- **Frontend**: New `combat` blocks in all four Liquid templates under `trmnl/src/`; new `polling_url` entry and `campaign_id` handling in `settings.yml`
- **Backend**: New `GetTRMNLCombatStatus` handler registered in the public route group; reuses `trmnlTokenValid`; read-only query over `combat_entries` plus condition-name resolution
- **Database**: None (read-only over existing `combat_entries`; conditions resolved from the existing conditions library)
- **Dependencies**: None
- **CI/CD**: Covered by the existing `trmnlp lint` job — no workflow changes
