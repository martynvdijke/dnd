# Tech Debt Registry

Single ledger of known **structural** tech debt in Villum. Structural debt has no
natural anchor file — the symptom is a pattern across many files — so it lives
here instead of in scattered code comments.

This registry is a summary, not a tracker: it links to the OpenSpec change (or PR)
that carries the full proposal, design, and tasks. It does not duplicate them.

## How to add or update an entry

1. Add one row per item using the schema below. Use the next free `TD-NNN` id.
2. **Evidence is required** and must be verifiable: a `path:line` reference or a
   concrete count. Without evidence an entry rots into folklore.
3. Link the entry to the OpenSpec change or PR that plans the fix.
4. Update `status` when the linked change is archived or superseded — do it in the
   same PR, so the registry does not drift stale.

| Field | Meaning |
| --- | --- |
| ID | `TD-NNN`, stable, never reused |
| Area | Short subsystem label (e.g. data access, request lifecycle, dice engine) |
| Symptom | One line describing what is wrong |
| Evidence | `path:line` references or quantitative counts that let a reader verify it |
| Impact | Why it matters / what breaks if ignored |
| Owner | Team or area accountable for disposition |
| Status | `active`, `in-review`, `resolved`, or `superseded` |
| Linked change/PR | OpenSpec change or PR carrying the plan |

## Registry

| ID | Area | Symptom | Evidence | Impact | Owner | Status | Linked change/PR |
| --- | --- | --- | --- | --- | --- | --- | --- |
| TD-001 | Data access | `ent` and raw SQL are both used, often in the same function | `db.DB.` ~1,134 call sites vs `db.Client.` ~208; mixed in `handlers/characters_crud.go:78,95,220,340,514`; 23 files import `database/sql`, 17 import `ent` | Two access paths plus ad-hoc identifier guards make schema drift and injection regressions easy | backend | active | `unify-data-access-seam` |
| TD-002 | Request lifecycle | DB calls use detached contexts and some `*sql.Rows` are not deferred-closed | ~22 `context.TODO()`/`context.Background()` sites under `handlers/**`; `rows.Close()` appears 188 times, ~29 not deferred | Client cancellation does not abort in-flight queries; early returns can leak connections | backend | active | `propagate-request-context` |
| TD-003 | Quality gates | Coverage floors documented in repo docs contradicted CI enforcement | `AGENTS.md` and `openspec/specs/local-quality-gates/spec.md` said handlers ≥30%, middleware ≥25%, vitest ≥20%, while `.github/workflows/ci.yaml:15-18` and `scripts/ci/test-go.sh` enforce 40%, 40%, 20% | Contributors cannot trust the documented floors; vitest floors sat far below measured coverage | ci | in-review | `reconcile-coverage-gates` |
| TD-004 | Dice engine | Embedded Go bundle can be present but stale relative to Go expectations | `dice/bundle_stub.go` embeds gitignored `dice/bundle.js`; `dice/engine.go` loads it via goja; dep `@dice-roller/rpg-dice-roller` ranged at `^5.5.1` | A stale bundle silently breaks the `__diceRoll`/`__diceRoller` contract instead of failing fast | dice | active | `harden-dice-bundle-seam` |
| TD-005 | Edit UI | Several entities remain create/delete-only or lack any edit screen in the UI | NPC edit form + `PUT /htmx/npcs/:id` added (`handlers/htmx_campaign.go`, `handlers/templates/npcs_form.html`); `OneShotActNPC` gained `PUT /oneshot-acts/:id/npcs/:nid` plus act-details panel wiring (`ts/app/oneshot.ts`); `CraftingRecipe` gained `PUT /crafting/recipes/:id` + edit modal (`ts/app.ts`); `DowntimeActivity` gained a full TS tab (`ts/characters/downtime.ts`). `FactionReputation` is already editable via the `POST /faction-reputation` upsert + edit modal, so no PUT added (YAGNI) | Users can now edit these records instead of delete+recreate | frontend | resolved | `edit-affordances-3` (follow-up to PR #135) |

## Marker convention

Any deliberate simplification or deferral marked with a `ponytail:` comment, or
introduced with a `TODO`/`FIXME`/`HACK` marker, **must** either reference a
registry entry id (e.g. `TD-002`) or state its ceiling and upgrade path inline
(`Ceiling: …; upgrade: …`). Trivial local simplifications may use the inline form;
anything structural gets an entry in the table above.
