## Why

`ent` defines the schema, migrations, and generated model layer (~370k generated LOC), yet runtime data access is dominated by hand-written SQL: `db.DB.` appears ~1,134 times vs `db.Client.` ~208 in `handlers/`, and the two are interleaved in the same functions (e.g. `handlers/characters_crud.go:78,95,220,340,514`). Schema truth and access path have drifted apart, ent hooks and validation are bypassed on most traffic, and dynamic SQL forced ad-hoc injection guards (`fieldNameRe`, `handlers/compendium_admin_schemas.go:20`). No document says when to use which, so every new handler re-decides.

## What Changes

- **Record the data-access convention** (ADR): `ent` is the canonical schema and write model for the entities it models; raw parameterized SQL is permitted for read-heavy projections, FTS/search, and bulk operations — but only through shared helpers.
- **Guard the seam**: centralize dynamic identifier / `json_extract` path construction behind one validated helper; remove ad-hoc regex guards.
- **Enforce parameterization**: never build SQL by concatenating caller-supplied values.
- **Single schema truth**: every raw DDL/migration stays registered through the ent/atlas path so `ent/migrate/schema.go` and `SELECT sql FROM sqlite_master` cannot diverge silently.
- **Regression guard**: a test/lint that flags newly concatenated SQL and duplicate hand-written column lists.
- **Not a migration**: this change does **not** convert existing raw-SQL call sites to `ent`.

## Capabilities

### New Capabilities
- `data-access-consistency`: the documented rule for `ent` vs raw SQL, and the guards that keep raw SQL safe and schema-consistent.

### Modified Capabilities
<!-- none: no existing spec's requirements change -->

## Impact

- `handlers/**` (every file mixing `db.DB` and `db.Client`), `db/**`, `ent/migrate/**` (schema truth), a new ADR under `docs/`, and a CI lint step.
- No API or runtime-behavior change; this is documentation plus guardrails.
