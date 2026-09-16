# ADR: Data access — ent for canonical schema and entity CRUD, raw parameterized SQL for projections / FTS / bulk

Status: Accepted
Date: 2026-09-16
Deciders: backend

## Context

Villum's persistence is SQLite via `modernc.org/sqlite`. Two access paths hit the same database:

- **ent** (`entgo.io/ent v0.14.6`) — ~336k lines of generated code under `ent/` (2363 lines in the canonical `ent/migrate/schema.go` alone). Import path `villum/ent/migrate` exposes `migrate.Tables` (`[]*schema.Table`, 65 tables) as the single generated source of truth for the schema.
- **raw `*sql.DB`** (`db.DB` in `db/db.go`) — the majority of handler traffic. A repo-wide count shows `db.DB.` on the order of ~1,134 call sites vs `db.Client.` (the ent client) around 208. 23 files import `database/sql`, 17 import `villum/ent` (counts per `docs/tech-debt.md` TD-001; current measurement 48 vs 18 file-level imports is the same shape). The two are interleaved in the same functions.

Canonical example is `handlers/characters_crud.go:53-120`, function `ListCharacters`:

- **Search branch** (`q != ""`, lines ~78-94) uses raw parameterized SQL against the FTS5 index because ent has no FTS support:

  ```go
  rows, err := db.DB.Query(`
      SELECT c.id, c.user_id, c.name, c.race, c.class, c.level, c.hp_max, c.hp_current, COALESCE(c.portrait_url,''), COALESCE(c.character_type,'player')
      FROM characters c JOIN characters_fts fts ON c.id = fts.rowid
      WHERE characters_fts MATCH ? ORDER BY c.updated_at DESC`, query)
  defer rows.Close()
  ```

  The `characters_fts MATCH ?` bind parameter is the correct way to pass FTS queries — ent cannot express this.

- **Non-search branch** (lines ~95-105) uses ent:

  ```go
  entChars, err := db.Client.Character.Query().Order(ent.Desc(character.FieldUpdatedAt)).All(c.Request.Context())
  ```

Both paths are legitimate and must coexist. A full migration of the ~1,134 raw sites to ent is explicitly out of scope and would carry high regression risk for no user-visible payoff.

Two additional constraints ground this decision:

- `db/db.go:169` sets `DB.SetMaxOpenConns(4)`. Holder-of-connection bugs are real — any raw `*sql.Rows` must be closed promptly (deferred) or the 4-connection pool stalls.
- New helper `handlers/sqlident.go` (added by this same change) centralizes the ad-hoc `fieldNameRe`-style guards: `validIdentifier(string) bool`, `jsonPathForField(string) (string, error)`, `orderDirection(string) string`. Dynamic identifiers and JSON paths must come from these or be compile-time constants, and values must be bound parameters where SQLite allows it (SQLite accepts a bind parameter as the `json_extract` path argument and in `LIMIT`).

Without a written rule, every new handler re-decides the seam, schema truth drifts, and injection surface grows.

## Decision

We accept the hybrid and guard its seam. The smallest change in the right place is to document the rule and make drift mechanically detectable, not to rewrite the handlers.

- ent is the canonical schema and the default path for entity CRUD.
- Raw parameterized SQL remains permitted for the cases ent cannot model well: read-heavy projections (multi-table joins into view models, e.g. `ListAllCharacters`), FTS5 / search, and bulk or cross-entity operations.

## Consequences

- New entity tables/columns go through ent first; `ent/migrate/schema.go:Tables` grows and `Client.Schema.Create` generates the DDL. Readers can trust ent as schema truth without grepping `db/migrations/`.
- Read-heavy and search paths stay in raw SQL where they already work and are easier to tune.
- The `handlers/sqlident.go` helpers become the single place to review identifier safety.
- `db/schema_consistency_test.go` makes unregistered raw DDL a CI failure instead of silent drift.
- Connection discipline under `SetMaxOpenConns(4)` is explicit: close `Rows` promptly.

## Rules

Normative language (SHALL/MUST) for new and modified code:

1. **Schema truth — ent.** ent SHALL be the single source of truth for schema and for entity CRUD. Every new entity field/table that ent can express MUST be added via the ent schema; `ent/migrate/schema.go` and `migrate.Tables` are canonical.
2. **Raw SQL — permitted scope.** Raw SQL is permitted for projections, FTS5/search, and bulk operations. Outside that scope, prefer ent.
3. **Parameterization.** Raw SQL MUST be parameterized for all values (`?` placeholders, `Query`/`Exec` bind args). Code MUST NOT interpolate client-supplied strings into SQL. String concatenation that embeds a caller value is a defect.
4. **Identifiers and ordering.** Identifiers that cannot be bound (table/column names, `json_extract` paths, `ASC`/`DESC`) MUST be compile-time constants or validated by `handlers/sqlident.go` (`validIdentifier`, `jsonPathForField`, `orderDirection`). Invalid identifiers MUST yield HTTP 400 and execute no SQL. Where SQLite permits it, pass `json_extract` paths and `LIMIT` as bound parameters.
5. **Connection hygiene.** Raw `*sql.Rows` MUST be closed promptly (deferred `rows.Close()` immediately after error check) — `DB.SetMaxOpenConns(4)` (`db/db.go:169`) makes leaks pool-blocking.
6. **Schema additions.** New schema SHALL be added through ent (then ent migrate generates DDL) or, when ent cannot express it (FTS5 virtual tables, `entity_search_index`, migration-only backfills), through `db/migrations/NNN_*.go` and the `migrations.Registry`. Unregistered raw DDL (ad-hoc `CREATE TABLE`/`CREATE VIRTUAL TABLE` outside those two paths) is a defect and MUST be caught by `db/schema_consistency_test.go`.

## Non-Goals

- No mass migration of the ~1,134 raw-SQL sites to ent.
- No ORM replacement and no removal of raw SQL.
- No change to HTTP contracts, response shapes, or frontend behavior.
- No performance work beyond the existing `SetMaxOpenConns(4)` discipline.

## References

- `ent/migrate/schema.go` — generated canonical schema, `var Tables = []*schema.Table{...}` (65 tables), `import villum/ent/migrate`.
- `db/db.go:169` — `DB.SetMaxOpenConns(4)`; `db/db.go:231-242` — ent client shares the same `*sql.DB`.
- `handlers/characters_crud.go:53-120` — `ListCharacters` hybrid (FTS `MATCH ?` vs `db.Client.Character.Query()`).
- `handlers/sqlident.go` — `validIdentifier`, `jsonPathForField`, `orderDirection`.
- `db/migrations/` + `db/search_index.go:398-417` (`EnsureSearchIndex`) — where non-ent DDL (FTS5 `characters_fts`, `compendium_entries_fts`, `entity_search_index`, `entity_links`, `npc_item_links`, etc.) is registered and triggers are re-created after ent's `Schema.Create`.
- `db/schema_consistency_test.go` — mechanical guard for this ADR.
- Tech debt: `docs/tech-debt.md` TD-001.
- OpenSpec change: `unify-data-access-seam` (`openspec/changes/unify-data-access-seam/`).
