## ADDED Requirements

### Requirement: Request context propagates to database work

Handler database and downstream calls SHALL receive the request context (`c.Request.Context()` or a context derived from it) so that client cancellation and deadlines propagate. `context.TODO()` and `context.Background()` SHALL NOT be used to initiate request-scoped database work in `handlers/**`.

#### Scenario: Client cancellation cancels the query

- **WHEN** a client disconnects while a slow query is in flight
- **THEN** the query's context is cancelled and the handler stops without continuing derived work

#### Scenario: No detached context in request paths

- **WHEN** `handlers/**` is inspected after the change
- **THEN** no request-scoped DB call is initiated with `context.TODO()` or `context.Background()`

#### Scenario: Background context remains valid for genuinely detached work

- **WHEN** a background job or startup task legitimately outlives any request
- **THEN** it uses an explicitly-owned context (not a request context) and this is documented at the call site

### Requirement: SQL rows are always closed

Every `*sql.Rows` returned by `Query` SHALL have `defer rows.Close()` registered immediately after the query succeeds, before any subsequent error check that can return. Any existing non-deferred `rows.Close()` call SHALL be converted to a deferred close with no behavior change on the success path.

#### Scenario: Early-return path closes rows

- **WHEN** a query succeeds but a later scan or validation fails and the handler returns early
- **THEN** `rows.Close()` has already been deferred and executes on return, releasing the connection

#### Scenario: Iteration helper closes rows

- **WHEN** a shared row-iteration helper is used
- **THEN** the helper owns the `Close` and the `rows.Err()` result is checked before returning

#### Scenario: No leaked connection under repeated failures

- **WHEN** the failing path is exercised repeatedly in a test
- **THEN** no connection remains open and subsequent queries succeed

### Requirement: No behavior change and coverage-safe

This change SHALL NOT alter any HTTP contract, response body, or status code. Coverage floors SHALL remain green (Go total ≥20%, `handlers/` ≥40%, `middleware/` ≥40%, vitest ≥35% statements/lines/functions and ≥30% branches).

#### Scenario: Contract unchanged

- **WHEN** the full test suite and e2e run after the change
- **THEN** no status code, JSON shape, or route has changed

#### Scenario: Coverage floors hold

- **WHEN** `task ci` runs after the change
- **THEN** all configured coverage floors remain satisfied
