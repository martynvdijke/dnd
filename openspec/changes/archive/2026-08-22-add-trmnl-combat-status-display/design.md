## Context

The TRMNL plugin (established in `add-trmnl-stats-display`, restructured for CI/release in `add-trmnl-character-roster`) polls token-guarded JSON endpoints under `/api/trmnl/` and renders Liquid templates on TRMNL's servers. The plugin lives in `trmnl/src/` with `settings.yml`, four layout templates, a CI lint job, and a release publish job. Since the API-token change (#87) the TRMNL read endpoints are public and unauthenticated — there is no `trmnl_token` setting or `trmnlTokenValid` helper; this endpoint follows that shipped pattern.

Combat state lives in the `combat_entries` table: one row per combatant per campaign with `initiative_roll` + `initiative_mod`, `hp_current`/`hp_max`, `ac`, an `is_active` flag marking the current turn, `turn_order` as tiebreaker, and `condition_ids` (CSV of tokens). The web combat tracker orders by `initiative_roll DESC, turn_order`. There is no conditions library table — the only condition "library" is the built-in `ConditionTypes` list (16 standard 5e conditions served via `/api/conditions/types`); the web UI does not currently write `condition_ids`, so resolution is defensive.

## Goals / Non-Goals

**Goals:**
- Serve the live initiative ladder as TRMNL-pollable JSON, ordered exactly like the web tracker
- Render it across all four TRMNL layouts with the active combatant unmistakable at table distance
- Resolve condition names server-side so templates stay dumb
- Degrade to a clear idle screen when nothing is running

**Non-Goals:**
- Round/turn counters — `combat_entries` has no round column; adding one is out of scope
- Write support (advancing turns from the display) — TRMNL is read-only by design here
- Push/websocket updates — TRMNL polls on its own schedule; e-ink refresh cadence makes push pointless
- Per-user visibility rules — the site-wide TRMNL token model already accepts that the table screen shows what the DM sees

## Decisions

### 1. Single endpoint `GET /api/trmnl/combat?campaign_id=...`
Public and read-only, matching the shipped TRMNL endpoints. Returns `{ campaign_id, entries: [...] }` with entries sorted `initiative_roll DESC, turn_order ASC`. Each entry: `id, name, type, initiative_total, hp_current, hp_max, ac, is_active, conditions[]`.

**Why:** One poll, one template block set — matches the roster endpoint's shape and keeps `settings.yml` simple. `campaign_id` is required (combat is campaign-scoped); omitting it returns an empty entry list rather than an error so the idle screen renders.

**Alternative considered:** Reusing `ListCombatEntries` directly. Rejected — it's authed, returns raw `condition_ids`, and lacks the computed `initiative_total`; a thin dedicated handler avoids leaking web-shaped payloads onto a public route.

### 2. Condition resolution against the built-in condition types
The handler parses the `condition_ids` CSV and keeps tokens matching an entry of the built-in `ConditionTypes` list (case-insensitive), returning canonical library names; unknown tokens are skipped silently.

**Why:** E-ink templates have no lookup ability; names are what the table needs ("paralyzed" beats "3"). There is no conditions library table to join against — inventing one is out of scope, and since nothing currently writes `condition_ids`, defensive name matching covers whatever convention later lands without breaking the payload.

### 3. Active-turn emphasis is data-driven, styling is template-side
`is_active` is already maintained by the web tracker. Templates render the active row inverted/bolded per TRMNL's monochrome classes.

**Why:** No new state to compute or keep warm; the display can never disagree with the tracker.

### 4. New `combat` display mode wired like `roster`
Add a third `polling_url` entry to `settings.yml`, a `campaign_id` custom field, and `combat` blocks in all four layouts branching on `display_mode`.

**Why:** Follows the established multi-mode plugin pattern exactly; no restructuring, no new plugin.

### 5. Idle state is explicit
Zero entries → payload still 200 with `entries: []`; templates render a centered "No active combat" block.

**Why:** A blank e-ink screen reads as "broken"; an explicit message reads as "between fights".

## Risks / Trade-offs

- **HP visible on a shared screen** → Accepted: mirrors the existing party view; the token model already assumes table-trusted content.
- **Stale data between polls** → TRMNL's refresh interval governs freshness; document in settings comments that 5 min suits combat nights, daily otherwise.
- **Large encounters overflow small layouts** → Quadrant/half templates cap the rendered list and show "+N more"; full renders all.
- **Condition CSV drift** → Tokens that don't match a known condition name are skipped silently rather than erroring the whole payload.

## Migration Plan

1. Deploy backend (new public route + tests) — additive, no schema change.
2. Land `trmnl/src/` template/settings updates; `trmnlp lint` gates them in CI.
3. User adds `campaign_id` custom field and selects the combat mode/template in TRMNL.
4. Rollback: remove the route and the template blocks; nothing else depends on them.

## Open Questions

- Whether the title bar should carry the campaign name (payload includes it; template choice).
