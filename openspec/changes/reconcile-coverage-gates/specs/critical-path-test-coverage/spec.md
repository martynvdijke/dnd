## ADDED Requirements

### Requirement: Coverage gates and documentation remain consistent with CI
The coverage floor values documented in `AGENTS.md` and in `openspec/specs/local-quality-gates/spec.md` and `openspec/specs/critical-path-test-coverage/spec.md` SHALL remain consistent with the thresholds enforced in `.github/workflows/ci.yaml` and the shared gate scripts (`scripts/ci/test-go.sh`, `scripts/ci/test-vitest.sh` / `vitest.config.ts`), which are handlers >=40%, middleware >=40%, total >=20% for Go and statements >=35%, branches >=30%, functions >=35%, lines >=35% for vitest.

#### Scenario: Spec/doc/CI consistency check
- **WHEN** the coverage thresholds in AGENTS.md, the specs, `ci.yaml`, and the gate scripts are compared
- **THEN** all sources report the same Go thresholds (handlers >=40%, middleware >=40%, total >=20%) and the same vitest thresholds (statements >=35%, branches >=30%, functions >=35%, lines >=35%)

### Requirement: Coverage floors ratchet upward never downward
Coverage floors for vitest (statements, branches, functions, lines) and for Go (`handlers/`, `middleware/`) SHALL only be raised over time as measured coverage rises, never lowered, except with explicit justification approved in the change proposal. When overall measured vitest coverage exceeds the current floor by a sustained margin, the next change that touches coverage gates SHALL propose a higher floor still strictly below the new measured value.

#### Scenario: Measured coverage rises and floor follows
- **WHEN** measured overall vitest coverage has risen and remained above the current floor
- **THEN** a subsequent change SHALL raise the floor to a higher value that is still below the new measured coverage, and SHALL NOT lower any existing floor

#### Scenario: Attempted floor reduction is rejected
- **WHEN** a change proposes lowering any coverage floor without explicit justification and approval
- **THEN** review SHALL reject the change as violating the ratcheting rule
