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
| TD-001 | Data access | `ent` and raw SQL are both used, often in the same function | 2026-11-27 re-measure: `db.DB` 1,540 references in `handlers/*.go` (1,149 non-test) vs `db.Client` 236, still mixed in e.g. `handlers/characters_crud.go`; `docs/adr-data-access.md` sanctions the hybrid (ent for entity CRUD, parameterized raw SQL for projections/FTS/bulk); guards landed: `handlers/sqlident.go`, `scripts/ci/check-concatenated-sql.sh`, `db/schema_consistency_test.go` | No outstanding work — the hybrid is the accepted architecture and its regression risk is covered by the identifier guard, the SQL-concatenation check and the schema consistency test | backend | superseded | `2026-10-17-unify-data-access-seam` |
| TD-002 | Request lifecycle | DB calls used detached contexts and some `*sql.Rows` were not deferred-closed (raw-SQL `Query`/`QueryRow`/`Exec` → `*Context` conversion was **not** in scope; it is tracked as TD-008) | 2026-11-27: 0 `context.TODO()`/`context.Background()` in committed `handlers/*.go` (excluding tests); all `rows.Close()` sites are deferred after deferring the last 3 non-deferred ones (`handlers/battlemap.go`, `handlers/combat.go`, `handlers/ink.go`) | Client cancellation now reaches handler DB work and early returns cannot leak a `*sql.Rows` | backend | resolved | `2026-10-17-propagate-request-context` |
| TD-003 | Quality gates | Coverage floors documented in repo docs contradicted CI enforcement | 2026-11-27: all sources agree — `scripts/ci/test-go.sh` and CI enforce Go total ≥20%, `handlers/` ≥40%, `middleware/` ≥40%; `vitest.config.ts` sets statements/lines/functions 35 and branches 30 (35/30/35/35); `AGENTS.md` and `openspec/specs/local-quality-gates/spec.md` state the same | Documented floors and enforced floors are identical, so coverage failures are predictable from the docs | ci | resolved | `2026-10-17-reconcile-coverage-gates` |
| TD-004 | Dice engine | Embedded Go bundle could be present but stale relative to Go expectations | `package.json` pins `@dice-roller/rpg-dice-roller` at exact `5.5.1`; `npm run build:dice` bakes a `globalThis.__diceBuildStamp='<version>'` banner into `dice/bundle.js`; `dice/engine.go:51,121` `verifyBuildStamp()` fails fast at engine init on a mismatch; `dice/contract_test.go` covers the `__diceRoll`/`__diceRoller` contract; Taskfile/CI/Dockerfile rebuild the bundle | A stale or mismatched bundle now fails at startup instead of silently changing dice results | dice | resolved | `2026-10-17-harden-dice-bundle-seam` |
| TD-005 | Edit UI | Several entities remain create/delete-only or lack any edit screen in the UI | NPC edit form + `PUT /htmx/npcs/:id` added (`handlers/htmx_campaign.go`, `handlers/templates/npcs_form.html`); `OneShotActNPC` gained `PUT /oneshot-acts/:id/npcs/:nid` plus act-details panel wiring (`ts/app/oneshot.ts`); `CraftingRecipe` gained `PUT /crafting/recipes/:id` + edit modal (`ts/app.ts`); `DowntimeActivity` gained a full TS tab (`ts/characters/downtime.ts`). `FactionReputation` is already editable via the `POST /faction-reputation` upsert + edit modal, so no PUT added (YAGNI) | Users can now edit these records instead of delete+recreate | frontend | resolved | `edit-affordances-3` (follow-up to PR #135) |
| TD-006 | Data access | Campaign-scoped character reads still rely on hand-written SQL joins instead of ent traversal | 2026-11-27: collapsed behind `handlers/campaign_character_access.go` — `campaignCharacterJoinSQL` (the `characters c` → `campaign_characters cm` visibility join), `campaignCharactersIn`, `isCampaignDMForCharacter`, `firstCampaignIDForCharacter`, `campaignIDsForCharacter`; 32 of the 35 joins removed from 15 handler files (`combat_automation`, `dashboard`, `dice`, `dm_screen`, `email`, `hp_authority`, `htmx_character`, `knowledge`, `notes`, `oneshot_htmx_pacing`, `recaps`, `share`, `spells_cast`, `table_screen`, `ws`), each now a call to a helper or a constant fragment; 3 intentionally remain because their semantics are unique — `party.go:228` (`campaign_id AND character_id` membership guard), `transfer.go:416,478` (dynamic IN-list / dynamic where), `transfer_import.go:325` (INSERT, not a read); `search/campaign.go` keeps its own context-aware joins because `search` cannot import `handlers` | The campaign visibility predicate now has one definition, so schema drift and injection regressions have a single choke point; the raw-SQL/ent split itself is accepted architecture (TD-001) | backend | resolved | `2026-11-27-campaign-join-helpers` |
| TD-007 | Search/AI | Four handlers still held duplicate implementations of logic already extracted into `search/` and `ai/` (`handlers/ai.go` provider transport, `copilot.go` retrieval, `compendium.go` recall, `settings.go` setting read); the packages themselves were already consumed by `handlers/search.go`, `handlers/ai_search.go`, `handlers/ai_draft_lore.go` and `telegram/*` | 2026-11-27: all five `ponytail: TD-007` markers removed; `generateText`/`generateChat` wrap `ai.GenerateText`/`ai.GenerateChat`, `aiEnabled` wraps `ai.IsEnabled`, `retrieveCampaignContext` wraps `search.RetrieveCampaignContext`, `SearchCompendium` wraps `search.SearchCompendium` | None — each behaviour now has exactly one implementation; provider and retrieval fixes land once | backend | resolved | `2026-11-27-delegate-search-ai-packages` |
| TD-008 | Request lifecycle | ~1,448 raw `db.DB` `Query`/`QueryRow`/`Exec` calls in handlers are not context-aware, so request cancellation/deadline does not reach raw SQL (ent calls are covered) | 2026-11-27: `db.DB` 1,540 references in `handlers/*.go` (1,149 non-test); 547 `.Query(` + 402 `.QueryRow(` + 499 `.Exec(` vs only 35 `*Context` variants; `scripts/ci/check-detached-context.sh` guards against **new** detached contexts | Cancellation does not abort a long-running raw query — low practical impact on single-node SQLite, but the two access paths stay inconsistent until converted | backend | active (parked: convert-on-touch; full sweep est. 2-4 days, high review noise) | `2026-10-17-propagate-request-context` |

## Marker convention

Any deliberate simplification or deferral marked with a `ponytail:` comment, or
introduced with a `TODO`/`FIXME`/`HACK` marker, **must** either reference a
registry entry id (e.g. `TD-002`) or state its ceiling and upgrade path inline
(`Ceiling: …; upgrade: …`). Trivial local simplifications may use the inline form;
anything structural gets an entry in the table above.
