## Why

The codebase carries significant structural tech debt — the ent-vs-raw-SQL split, request-context/rows.Close discipline, coverage-gate documentation drift, and the Go↔JS dice bundle seam — but a repo-wide grep finds zero `TODO`/`FIXME`/`HACK` markers (outside generated code) and only three `ponytail:` deferral comments. The debt is invisible to grep and there is no registry, ledger, or ADR index a newcomer can read to discover what is known, deferred, or tracked elsewhere.

## What Changes

- Introduce a single tracked tech-debt registry document at `docs/tech-debt.md` with a fixed entry schema: ID, area, symptom, evidence (file:line or counts), impact, owner, status, and linked OpenSpec change/PR.
- Define a lightweight marker convention: any deliberate simplification or deferral marked with a `ponytail:` comment, or any `TODO`/`FIXME`/`HACK`-style deferral, MUST either reference a registry entry ID or include its ceiling and upgrade path inline.
- Seed the registry with four known structural items — ent-vs-raw-SQL split, request-context/rows.Close discipline, coverage-gate documentation drift, Go↔JS dice bundle seam — each linked to its corresponding OpenSpec change.
- Add a short convention section to `AGENTS.md` documenting the registry and marker rule.
- (Optional, only if cheap) Add a grep-based check that every `ponytail:` comment references a registry ID.
- Explicitly doc-only: no new tooling, no dashboard, no automation, no database, and no CI enforcement beyond the optional grep check.

## Capabilities

### New Capabilities
- `tech-debt-registry`: a discoverable ledger of known structural tech debt and the convention that keeps deferral markers linked to it.

### Modified Capabilities
<!-- none: no existing spec's requirements change -->

## Impact

- New file `docs/tech-debt.md` (registry); small edit to `AGENTS.md` (convention section); optional one-line grep check in CI/prek config.
- No runtime, API, or data-model change. No new dependencies.
