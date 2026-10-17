## 1. Document the convention

- [x] 1.1 Write `docs/adr-data-access.md`: when to use `ent` vs raw SQL (entity CRUD → ent; projections/FTS/bulk → parameterized raw SQL through helpers), with a one-line rationale per case and examples drawn from `handlers/characters_crud.go`.
- [x] 1.2 Reference the ADR from `AGENTS.md` (one line under a backend section).

## 2. Central dynamic-identifier guard

- [x] 2.1 Add a shared helper (e.g. `handlers/sqlident.go`) that validates a SQL identifier / whitelisted `json_extract` path and returns a safe fragment or an error.
- [x] 2.2 Migrate `handlers/compendium_admin_schemas.go` (`fieldNameRe`) and any equivalent per-handler guards to the shared helper.
- [x] 2.3 Ensure handlers receiving an invalid identifier return `400` and execute no SQL. Add table-driven tests (valid identifier, quotes, `;`, `--`, `(`).

## 3. Parameterization and schema truth

- [x] 3.1 Audit raw SQL sites for value interpolation; convert any interpolated caller value to a bind parameter.
- [x] 3.2 Add a schema-consistency test comparing the migrated schema against the ent-managed schema signature; fail on unregistered raw DDL.
- [x] 3.3 Confirm every raw DDL statement routes through the migration registry in `db/migrations/`.

## 4. Regression guard + verification

- [x] 4.1 Add a CI lint (script under `scripts/ci/`, referenced from `ci.yaml`) that fails on string-concatenated SQL containing caller values in `handlers/**`.
- [x] 4.2 Run `task ci` and confirm all coverage floors hold (Go total ≥20%, `handlers/` ≥40%, `middleware/` ≥40%, vitest ≥35% statements/lines/functions and ≥30% branches).
- [x] 4.3 Confirm no HTTP contract changed (route snapshot + e2e green).
