## ADDED Requirements

### Requirement: Documented data-access convention

The project SHALL document a single convention stating when to use the `ent` client versus raw `database/sql`, covering at minimum: entity writes/reads, read-heavy projections, full-text search, and bulk operations. The convention SHALL be recorded as an ADR under `docs/` and referenced from `AGENTS.md`.

#### Scenario: Contributor finds the convention

- **WHEN** a contributor looks up how to query the database
- **THEN** the ADR states which access path to use for entity CRUD, projections, FTS/search, and bulk work, with a one-line rationale per case

#### Scenario: Convention matches the code

- **WHEN** the ADR is compared against `handlers/`
- **THEN** the dominant existing patterns are described as allowed (not as violations), so the document reflects reality rather than an aspirational rule nobody follows

### Requirement: Raw SQL is always parameterized

All raw SQL SHALL pass caller-supplied values as bind parameters. SQL text SHALL NOT be assembled by concatenating or `fmt.Sprintf`-ing caller-supplied values into the statement.

#### Scenario: Value is bound, not interpolated

- **WHEN** a handler queries by a user-supplied value
- **THEN** the value is passed as a `?`/named parameter and never embedded in the SQL string

#### Scenario: Regression guard fails on interpolation

- **WHEN** the SQL-safety check runs over `handlers/**` and a statement interpolates a caller value
- **THEN** the check exits non-zero and names the offending file and line

### Requirement: Dynamic identifiers and JSON paths validated centrally

Dynamic SQL identifiers (column names, sort/filter fields) and `json_extract` paths SHALL be produced by a single shared validation helper, rather than per-handler regexes or string concatenation. The helper SHALL reject anything that is not a simple identifier or a whitelisted JSON path.

#### Scenario: Valid identifier accepted

- **WHEN** a whitelisted sort field such as `name` is passed to the helper
- **THEN** the helper returns it as a safe SQL fragment

#### Scenario: Malformed identifier rejected

- **WHEN** a value containing quotes, semicolons, `--`, or `(` is passed to the helper
- **THEN** the helper rejects it and the handler returns `400` without executing SQL

#### Scenario: Ad-hoc guards removed

- **WHEN** the existing per-handler dynamic-path guards (e.g. `fieldNameRe` in `handlers/compendium_admin_schemas.go`) are inspected after the change
- **THEN** they delegate to the shared helper instead of duplicating the validation rule

### Requirement: Single schema source of truth

The ent-managed schema SHALL remain the single source of truth. Raw DDL executed outside ent migrations SHALL be registered through the existing migration registry so that the live schema and `ent/migrate/schema.go` cannot diverge without a test failure.

#### Scenario: Fresh DB schema is reproducible

- **WHEN** migrations run on an empty database
- **THEN** the resulting schema matches the ent-managed schema and the schema-consistency test passes

#### Scenario: Unregistered DDL is detected

- **WHEN** a raw DDL statement is added without a corresponding migration registry entry
- **THEN** the schema-consistency test fails and identifies the unregistered statement

### Requirement: No behavior change

This change SHALL NOT alter any HTTP contract, response body, status code, or data-access semantics. It SHALL keep coverage floors green (Go total ≥20%, `handlers/` ≥40%, `middleware/` ≥40%, vitest ≥20% all metrics).

#### Scenario: Contract unchanged

- **WHEN** the full test suite and e2e run after the change
- **THEN** no status code, JSON shape, or route has changed

#### Scenario: Coverage floors hold

- **WHEN** `task ci` runs after the change
- **THEN** all configured coverage floors remain satisfied
