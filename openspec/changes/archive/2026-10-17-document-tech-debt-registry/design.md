## Context

Villum's known tech debt is structural: the ent-vs-raw-SQL access split, request-context/rows.Close handling, coverage-gate documentation drift, and the Go↔JS dice bundle seam are each being captured as separate OpenSpec changes (`unify-data-access-seam`, `propagate-request-context`, `reconcile-coverage-gates`, `harden-dice-bundle-seam`). Outside generated code a grep finds effectively zero `TODO`/`FIXME`/`HACK` markers; the only deliberate-deferral markers are three `ponytail:` comments in TypeScript (`ts/init.ts:109`, `ts/characters/sheet.ts:59`, `ts/app.ts:92`). There is no single document a newcomer can open to find what debt is known, why it was deferred, or where the plan lives. Scattering more markers does not solve this because structural debt has no natural anchor file — the symptom is a pattern across many files, not a line to annotate.

## Goals / Non-Goals

**Goals:**
- Give every known structural debt item one discoverable home with enough fields (ID, area, symptom, evidence, impact, owner, status, linked change/PR) that a reader can verify it and find the plan.
- Keep the authoring cost near zero: a markdown table plus a short convention in `AGENTS.md`.
- Seed the registry with the four known items so it is useful on day one.

**Non-Goals:**
- Becoming a bug tracker or replacing GitHub Issues / OpenSpec changes — the registry links to them and tracks disposition, it does not duplicate them.
- New tooling, dashboard, database, automation, or required CI enforcement beyond an optional grep that checks `ponytail:` comments reference an entry ID.
- Migrating existing raw SQL or other debt; that work stays in its own changes.

## Decisions

**Decision: Single markdown file at `docs/tech-debt.md` with entries in a table.**
Rationale: Lowest ceremony that is still greppable, diffable, and reviewable in PRs. Markdown tables render on GitHub, require no schema migration, and match existing `docs/` ADR patterns.
Alternatives considered:
- YAML/JSON ledger — adds a parser, schema file, and contributor learning curve for no query need (fewer than ~20 entries expected).
- Per-area ADR files — scatters the debt again and duplicates the problem this change solves.
- GitHub Project / external board — not versioned with the code, invisible in offline review, and would duplicate OpenSpec as the plan store.

**Decision: Fixed entry schema (ID, area, symptom, evidence, impact, owner, status, linked change/PR) — fields-in-a-table, not freeform.**
Rationale: Evidence (file:line or counts) is what makes structural debt verifiable; without it entries rot into folklore. Status + linked change/PR keeps the registry from diverging from OpenSpec. Owner prevents orphan items.
Alternative: Freeform prose entries — cheaper to write, but omits evidence and breaks the "verify the debt" requirement.

**Decision: Marker convention — `ponytail:` (and any TODO-style) deferrals MUST reference a registry ID or state ceiling + upgrade path inline.**
Rationale: Greppable markers alone failed here: structural debt has no single file to mark, so markers would either be absent (current state) or scattered without context. Tying markers to the registry gives each simplification a home without forcing every structural issue to have a code annotation. Allowing the inline `ceiling + upgrade path` alternative keeps trivial local simplifications cheap.
Alternative: Mandate a registry entry for every `ponytail:` — heavier than needed for local, low-impact simplifications.

**Decision: Seed the four known items on creation, each linked to its change.**
Rationale: An empty registry invites neglect. Seeding proves the schema and gives immediate value.

## Risks / Trade-offs

- **Registry drifts stale** (entries not updated when changes land) → Mitigation: status field + link to change; update registry in the same PR that archives or supersedes a change. Keep entry count small so drift is visible in review.
- **Convention ignored** (authors add `ponytail:` or `TODO` without ID or ceiling) → Mitigation: optional lightweight grep check (`ponytail:` references `TD-`); cheap to add, not required for v1. Documented in `AGENTS.md` as the primary enforcement.
- **Registry mistaken for issue tracker** → Mitigation: design explicitly states non-goal and spec requires entries link out rather than duplicate change content.
- **Evidence lines go stale** (file:line shifts) → Mitigation: evidence may use counts or directory-level references where line stability is low; review updates evidence when touching the area.

## Migration Plan

1. Create `docs/tech-debt.md` with table header and four seeded rows referencing `unify-data-access-seam`, `propagate-request-context`, `reconcile-coverage-gates`, `harden-dice-bundle-seam`.
2. Add one short section to `AGENTS.md` describing the registry location and the `ponytail:`/TODO marker rule.
3. Bring the three existing `ponytail:` comments into conformance (reference an entry ID or add inline ceiling + upgrade path).
4. Optionally add a grep-based check that flags `ponytail:` lines lacking a `TD-` reference.
5. Rollback: delete `docs/tech-debt.md` and revert the `AGENTS.md` section; no runtime effect.

## Open Questions

- None — scope is intentionally doc-only. Whether to enforce the grep check in CI vs prek is left to the implementer; either satisfies the optional task.
