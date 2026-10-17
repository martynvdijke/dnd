## Context

The CI workflow `.github/workflows/ci.yaml` (lines ~15-18, bumped 09-2026 via `close-coverage-holes`) enforces Go `handlers/ >= 40%`, `middleware/ >= 40%`, total >= 20% (measured at the time: 72.1% / 86.9%), and delegates to `scripts/ci/test-go.sh` which already checks 40%/40%/20%. The shared scripts are the enforcement truth.

Two written records contradict that truth:
- `openspec/specs/local-quality-gates/spec.md` requirement "Coverage thresholds are enforced locally identically to CI" still states `handlers >= 30%, middleware >= 25%`.
- `AGENTS.md:37` states `handlers/ >=30%, middleware/ >=25%; vitest >=20%` (and `vitest.config.ts` enforces 20%).

Separately, vitest actual coverage (`coverage/coverage-summary.json`) is 43.55% lines / 41.85% statements / 39.83% functions / 35.74% branches — 15-23 points above the 20% floor — so a substantial regression could land without any gate firing. The weakest modules (`ts/dice.ts` 7.72% lines, `ts/search.ts` 16.16%, `ts/navigation.ts` 18.36%, `ts/pwa.ts` 18.05%, `ts/compendium.ts` 27.74%) are not addressed here.

Stakeholders: contributors running `task ci` / pre-push hooks, reviewers checking spec/CI alignment.

## Goals / Non-Goals

**Goals:**
- Make documentation the single consistent record of the enforced gates: AGENTS.md, `local-quality-gates` spec, and `critical-path-test-coverage` spec all state exactly what CI enforces.
- Ratchet vitest floors upward from 20% to a conservative value strictly below measured overall (35% statements/lines/functions, 30% branches) so future regressions fail early, without chasing 100% or per-module thresholds.
- Establish a rule that floors ratchet only upward (never down) as coverage rises, and that doc/spec drift fails review.

**Non-Goals:**
- Adding tests to raise coverage (no module-level gap closure; no per-file thresholds).
- Changing CI mechanics, enforcement scripts' structure, or total Go floor (20%).
- Introducing per-module vitest floors or branch-level tightening beyond the conservative ratchet.

## Decisions

**Reconcile to CI truth, not to the stale spec.**
- Why: `ci.yaml` + `scripts/ci/test-go.sh` are the enforced gates and already correct (40%/40%/20%). Changing them would lower the bar. Changing the written record to match the enforcement is the minimal, correct fix.
- Alternative considered: lower CI to match the stale 30%/25% — rejected as a regression.

**Ratchet vitest floors to 35/30/35/35 instead of leaving at 20% or jumping to measured overall.**
- Why: Measured overall is ~43.55% lines / 41.85% statements / 39.83% functions / 35.74% branches. Setting floors to 35% (lines/statements/functions) and 30% (branches) keeps headroom (~4-8 points) so normal variance does not flake CI, while still failing a real regression well before the old 20% floor. Floor equals the `vitest.config.ts` `coverage.thresholds` plus CI/local script check on `coverage/coverage-summary.json`.
- Alternative considered: set floors to measured values (43/41/39/35) — rejected as too tight, would flake on unrelated changes. Alternative considered: leave at 20% — rejected as the status quo that permits silent regression.
- Branches ratchets less (30% vs 35%) because measured branches (35.74%) leaves less headroom; squeezing to 35% would be at the measured value.

**Add a consistency/ratcheting requirement in specs rather than only updating numbers.**
- Why: The defect is not the numbers but the absence of a rule that keeps them consistent. A requirement that AGENTS.md + specs SHALL match `ci.yaml`/scripts, and that floors ratchet only upward, makes drift a review failure rather than a silent reoccurrence.
- Alternative considered: a script that parses ci.yaml and asserts doc equality — over-engineered for now; a spec requirement enforced at review is cheaper and sufficient. Upgrade to automated doc/CI diff check if drift recurs.

## Risks / Trade-offs

- [Ratcheted floor too close to measured branches] → branches floor set to 30% (5.7 points below 35.74%) to avoid flakiness; statements/lines/functions at 35% leave 4-8 points of margin.
- [Existing modules below the new floor on their own] → floor is global overall, not per-module; weakest modules remain below the global floor but are not newly gated — no immediate breakage.
- [Future coverage dip flakes CI] → headroom chosen conservatively; if a legitimate refactor lowers coverage slightly, the ratcheting rule explicitly allows justified discussion, but default is upward only.
- [Manual consistency requirement is review-dependent] → risk of drift returning; mitigated by making it a spec requirement that fails review, with optional future automation.

## Migration Plan

1. Update `AGENTS.md:37` to `handlers/ >=40%, middleware/ >=40%, vitest >=35% (branches >=30%)`.
2. Update `openspec/specs/local-quality-gates/spec.md` (archived spec) via delta — but the delta itself captures the fix; archive applies it.
3. Ratchet `vitest.config.ts` thresholds from 20 to 35/30/35/35 (and ensure `scripts/ci/test-vitest.sh` surfaces the failure).
4. Run `task ci` (full parity suite including `npx vitest run --coverage` and `scripts/ci/test-go.sh`) to confirm green.
5. No rollback risk: thresholds only tightened modestly below measured; revert is a one-line config change.

## Open Questions

- None — floors are fixed at 35/30/35/35 for this change. Next ratchet point is deferred until overall coverage materially exceeds these floors (e.g. >50%).
