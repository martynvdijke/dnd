## Context

`ent` is a first-class dependency (`entgo.io/ent v0.14.6`) whose generated output (`ent/`, ~370k LOC) and migrations define the schema. In practice, handlers use the underlying `*sql.DB` (`db.DB`) far more than the ent client (`db.Client`) — roughly 1,134 vs 208 call sites — and frequently in the same function. Both paths target the same SQLite database (`modernc.org/sqlite`). There is no documented rule distinguishing them, and at least one handler carries a bespoke regex (`fieldNameRe`) purely to make dynamic `json_extract` paths safe.

This is the largest un-captured structural risk in the repo: the schema is modeled one way and accessed another, so ent's type-safety, hooks, and validation apply to a minority of traffic.

## Goals / Non-Goals

**Goals:**
- Make the existing hybrid explicit and safe, without a rewrite.
- Reduce the injection surface created by dynamic identifiers and JSON paths.
- Make schema drift detectable by a test.
- Give every future handler a written default instead of a per-author guess.

**Non-Goals:**
- Migrating existing raw-SQL call sites to ent (explicitly out of scope).
- Removing ent, or removing raw SQL.
- Changing any API contract, response shape, or runtime behavior.
- Performance work.

## Decisions

### Decision: Document and guard the seam, not migrate

The hybrid is load-bearing: read-heavy projections, FTS, and bulk paths are already written in raw SQL and work. A migration to ent would touch ~1,134 call sites across 23 files with high regression risk and no user-visible payoff. The cheaper, higher-value move is to accept the hybrid, write down the rule, and remove the sharp edges (unparameterized/dynamic SQL, schema drift). This is the "smallest change in the right place" option.

### Decision: One shared dynamic-identifier helper

Rather than repeat the `fieldNameRe`-style guard per handler, one helper validates identifiers and whitelisted `json_extract` paths. This shrinks both the injection surface and the number of places a future change must touch.

### Decision: Detect drift with a test, not a policy

Schema truth is asserted by a test comparing the migrated schema to the ent-managed schema, so an unregistered raw DDL statement fails CI instead of silently diverging. Tests are enforceable; policy documents are not.

### Decision: Guard new code, grandfather old code

The regression guard targets *new* interpolated SQL. Existing parameterized sites are not rewritten. This keeps the diff reviewable and avoids a mass-touch that would obscure real changes.

## Risks / Trade-offs

- **Documentation can still be ignored.** Mitigation: pair the ADR with the mechanical regression guard so the rule has teeth.
- **Helper too strict may reject legitimate identifiers.** Mitigation: whitelist the known-good identifier grammar and JSON path shapes; return 400 rather than silently degrading.
- **Schema-consistency test may be brittle** across `modernc.org/sqlite` versions. Mitigation: compare a normalized schema signature, not raw formatting.
- **Grandfathering leaves old risk.** Accepted deliberately: existing sites are parameterized today; the change prevents new regressions without a mass rewrite.
