## Why

The enforced coverage gates in `.github/workflows/ci.yaml` (Go `handlers/ >= 40%`, `middleware/ >= 40%`, total >= 20%) diverge from the written record in `openspec/specs/local-quality-gates/spec.md` and `AGENTS.md`, which still state `handlers/ >= 30%` / `middleware/ >= 25%`. The enforced gates are correct; the documentation is stale, so drift can silently persist. Separately, the vitest floor of 20% on all metrics is ~15-23 points below the measured overall of 43.55% and allows large regressions before failing.

## What Changes

- Reconcile `openspec/specs/local-quality-gates/spec.md` requirement "Coverage thresholds are enforced locally identically to CI" from `handlers >= 30%, middleware >= 25%, vitest >= 20% all metrics` to the CI-enforced truth: `handlers >= 40%, middleware >= 40%, total >= 20%` and a ratcheted vitest floor of `statements/lines/functions >= 35%, branches >= 30%` (strictly below the measured 43.55% / 41.85% / 39.83% / 35.74%).
- Reconcile `AGENTS.md:37` coverage floor line to the same values.
- Ratchet `vitest.config.ts` (and/or `scripts/ci/test-vitest.sh`) threshold config from 20% to the new floors so local and CI enforcement move together.
- Add a consistency/ratcheting requirement to `critical-path-test-coverage` (and `local-quality-gates`) that AGENTS.md and the specs SHALL state the CI-enforced values and that floors move only upward as measured coverage rises.

## Capabilities

### New Capabilities

<!-- No new capabilities — this change reconciles existing gates -->

### Modified Capabilities

- `local-quality-gates`: Coverage-threshold requirement values are updated to CI truth and ratcheted vitest floors; added documentation-consistency rule.
- `critical-path-test-coverage`: Added requirement that AGENTS.md/specs stay consistent with `ci.yaml` and vitest floors ratchet upward (never down).

## Impact

- Docs: `AGENTS.md` line 37.
- Specs: `openspec/specs/local-quality-gates/spec.md`, `openspec/specs/critical-path-test-coverage/spec.md` (and their change deltas under `openspec/changes/reconcile-coverage-gates/specs/`).
- Config/scripts: `vitest.config.ts` and/or `scripts/ci/test-vitest.sh` (threshold bump), `scripts/ci/test-go.sh` already correct — no change.
- CI: `ci.yaml` is the source of truth and is not changed; local gates are aligned to it.
