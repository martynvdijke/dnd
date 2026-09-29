## Why
One-shot adventures are organised as Act → Scene trees, and encounters are built separately as `EncounterTemplate` rows. The join table `oneshot_adventure_encounters` already carries a nullable `act_id` (migration 029) and `OneShotAct` has an `encounters` edge, but every handler ignores it: `LinkOneShotEncounter` always inserts `act_id=NULL` and the only routes are adventure-scoped. DMs can therefore build encounters but cannot attach them to the act they belong to, and the act tree shows no encounter context.

## What Changes
- Accept an optional `act_id` when linking an encounter; add act-scoped JSON routes `GET/POST/DELETE /api/oneshot-acts/:id/encounters`.
- Return `act_id` from encounter-link queries and load per-act encounter names so the act tree can render them.
- Add an HTMX encounter picker per act (`/htmx/oneshot-acts/:id/encounters`) showing linked encounters plus a select of the adventure's campaign encounters.
- Expose the existing `OneShotScene.encounter_id` field in the scene form (select), and persist it in the HTMX scene create/update handlers.

## Capabilities
### New Capabilities
- `oneshot-act-encounters`
### Modified Capabilities
- none

## Impact
Backend: `handlers/oneshot_links.go`, `handlers/oneshot_mapping.go`, `handlers/oneshot_htmx_acts.go`, `handlers/routes_oneshot.go`, `handlers/htmx.go`. Templates: new `handlers/templates/oneshot_act_encounters.html`, edits to `oneshot_detail.html` and `oneshot_scene_form.html`. No schema migration required (`act_id` already exists). No breaking API change — `act_id` is optional and adventure-level links keep working unchanged.
