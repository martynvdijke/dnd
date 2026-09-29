## Context
`oneshot_adventure_encounters` links an `EncounterTemplate` to a `OneShotAdventure`, optionally to a `OneShotAct` via nullable `act_id`. The act edge is `edge.From("act").Ref("encounters").Field("act_id").Unique()` — `Unique` marks child→one-parent, so a single act may link many encounters (the FK is not DB-unique; confirmed in `ent/migrate/schema.go`). Adventure-level links are rows with `act_id NULL`.

## Goals / Non-Goals
- Goals: DMs can link/unlink encounters per act; linked encounters are visible on the act tree and manageable from an act-scoped modal; scenes can reference an encounter.
- Non-goals: editing encounter contents from the one-shot UI (campaign encounter builder already does that); auto-creating encounters; scene→location picking (out of scope, existing JSON field only).

## Decisions
- **Optional `act_id` on the existing link endpoint** rather than a parallel table/path: one insert statement, `INSERT OR IGNORE` keeps idempotency. Adventure-level rows (no `act_id`) remain valid.
- **Server-rendered HTMX fragment** for the act modal instead of new JS: the pattern already exists (`HtmxActDetails`, act NPC lists) and avoids new global JS.
- **Candidate encounters**: campaign-scoped (`encounter_templates WHERE campaign_id=?`) when the adventure has a campaign, otherwise the owner's encounters — the same rule `ListEncounters` already uses.
- **Scene encounter field**: reuse the existing `SetNillableEncounterID` semantics; add a `<select>` to the scene form and parse `encounter_id` in the HTMX scene handlers.

## Risks / Trade-offs
- HTMX re-render of the modal after link/unlink keeps state consistent but must target the fragment container, not the whole section (avoid closing the modal).
- Per-act encounter names require an extra lookup map in `loadAdventureDetail`; done once per adventure, no N+1.
