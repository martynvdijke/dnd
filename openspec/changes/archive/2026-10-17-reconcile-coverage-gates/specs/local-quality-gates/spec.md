## MODIFIED Requirements

### Requirement: Coverage thresholds are enforced locally identically to CI
The Go coverage gates (total >=20%, handlers >=40%, middleware >=40%, measured with `-coverpkg=./handlers/...,./middleware/... -covermode=atomic`) and the vitest coverage gates (>=35% statements, >=30% branches, >=35% functions, >=35% lines) SHALL be evaluated by the shared scripts on every local parity run.

#### Scenario: Coverage below threshold locally
- **WHEN** a local parity run measures coverage below any configured threshold
- **THEN** the corresponding script exits non-zero and reports which threshold failed

#### Scenario: Documented thresholds match CI-enforced values
- **WHEN** coverage thresholds are inspected in `AGENTS.md`, `openspec/specs/local-quality-gates/spec.md`, and `.github/workflows/ci.yaml` / gate scripts
- **THEN** the documented Go values equal handlers >=40%, middleware >=40%, total >=20% and the documented vitest values equal statements >=35%, branches >=30%, functions >=35%, lines >=35%, matching CI enforcement

## ADDED Requirements

### Requirement: Documentation stays consistent with CI-enforced gates
AGENTS.md and `openspec/specs/local-quality-gates/spec.md` SHALL state the same coverage floor values that are enforced in `.github/workflows/ci.yaml` and the shared gate scripts (`scripts/ci/test-go.sh`, `scripts/ci/test-vitest.sh` / `vitest.config.ts`). Any change to an enforced gate threshold SHALL be accompanied by an update to AGENTS.md and the specs in the same change.

#### Scenario: CI threshold bump without doc update is rejected
- **WHEN** a change modifies an enforced threshold in `ci.yaml` or a gate script without updating AGENTS.md and the spec to the same value
- **THEN** review SHALL reject the change as inconsistent with this requirement

#### Scenario: Doc-only threshold edit without enforcement change is rejected
- **WHEN** AGENTS.md or the spec is edited to state a different threshold than what CI and the gate scripts enforce
- **THEN** review SHALL reject the change and require the enforcement and documentation to be reconciled
