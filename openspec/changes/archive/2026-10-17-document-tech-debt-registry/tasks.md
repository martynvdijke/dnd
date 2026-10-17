## 1. Registry document

- [x] 1.1 Create `docs/tech-debt.md` with a brief header (purpose, how to add/update entries) and a markdown table whose columns match the entry schema: ID, area, symptom, evidence, impact, owner, status, linked OpenSpec change/PR
- [x] 1.2 Seed the table with four entries for the known structural items — ent-vs-raw-SQL data-access split (`unify-data-access-seam`), request-context/rows.Close discipline (`propagate-request-context`), coverage-gate documentation drift (`reconcile-coverage-gates`), Go↔JS dice bundle seam (`harden-dice-bundle-seam`) — each with concrete evidence (file:line or counts) and a link to its change
- [x] 1.3 Verify the registry is discoverable from `AGENTS.md` and `docs/` (e.g., referenced in `AGENTS.md`; no orphan file)

## 2. Marker convention

- [x] 2.1 Add a short convention section to `AGENTS.md` stating that any `ponytail:` comment or any `TODO`/`FIXME`/`HACK`-style deferral MUST either reference a registry entry ID or include its ceiling and upgrade path inline
- [x] 2.2 Bring the three existing `ponytail:` markers (`ts/init.ts:109`, `ts/characters/sheet.ts:59`, `ts/app.ts:92`) into conformance — each now carries an inline `Ceiling: …; upgrade: …` line

## 3. Optional grep check (only if cheap)

- [x] 3.1 Add a lightweight grep-based check (CI or prek) that flags `ponytail:` comments lacking a registry ID reference — **skipped by design**: the convention allows the inline `Ceiling:/upgrade:` alternative, so a `TD-`-only grep would flag conforming comments. No new tooling added; revisit only if the convention is narrowed to require IDs.

## 4. Validation

- [x] 4.1 Run `openspec validate document-tech-debt-registry` and fix until it passes
