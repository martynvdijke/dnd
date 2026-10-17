# backend-error-resilience Specification

## Purpose
TBD - created by archiving change harden-go-error-handling. Update Purpose after archive.
## Requirements
### Requirement: Dice engine failure degrades gracefully

The system SHALL NOT panic if the dice engine fails to initialize (e.g., corrupt or missing bundle). The init error SHALL be stored and surfaced via structured logging, and every dice endpoint SHALL return `503 Service Unavailable` with a JSON error when the engine is unavailable, while all non-dice endpoints SHALL continue to serve normally.

#### Scenario: Corrupt dice bundle does not crash the server
- **WHEN** the dice bundle is corrupt or missing at startup and any non-dice endpoint is called
- **THEN** the server starts successfully and the non-dice endpoint returns its normal response

#### Scenario: Dice endpoint returns 503 when engine unavailable
- **WHEN** the dice engine failed to init and a dice endpoint is called
- **THEN** the endpoint returns 503 with a JSON error indicating the dice engine is unavailable

### Requirement: Corrupt session timestamps are treated as invalid

When loading a session row, if `CreatedAt` or `ExpiresAt` fails to parse, the system SHALL treat the session as invalid: delete the corrupt row and return no session (forcing re-authentication), and log a warning with the session ID. The system SHALL NOT silently use a zero-time value.

#### Scenario: Corrupt session timestamp invalidates the session
- **WHEN** a session row with an unparseable `CreatedAt` or `ExpiresAt` is loaded
- **THEN** the row is deleted, no session is returned, and the caller is treated as unauthenticated

#### Scenario: Valid session timestamps still work
- **WHEN** a session row has valid timestamps
- **THEN** the session is loaded normally with correct time values

### Requirement: Token write failures are not swallowed

The system SHALL check the error returned by token persistence (`TokenDB.Exec` in `middleware/apitoken.go:114` and equivalent sites) and, on failure, SHALL log the error and return a 500 to the caller instead of silently continuing.

#### Scenario: Token write failure returns 500
- **WHEN** the token store fails to persist a token
- **THEN** the request returns 500 and the error is logged

### Requirement: Malformed numeric inputs return 400

Every `ParseInt`/`Atoi` site that currently discards the error (`handlers/htmx.go:1525`, `handlers/upload.go:79`, `handlers/upload_links.go:90`, `handlers/compendium_linking.go:222,262`, and any equivalent sites) SHALL validate the parse result and, on failure, SHALL return `400 Bad Request` with an error describing the invalid field, instead of silently using `0`.

#### Scenario: Malformed OwnerID returns 400
- **WHEN** `OwnerID` is supplied as a non-numeric or empty value to `upload.go:79` or `upload_links.go:90`
- **THEN** the handler returns 400 and does not create or link with `OwnerID 0`

#### Scenario: Malformed quantity returns 400
- **WHEN** `quantity` is supplied as a non-numeric value to `compendium_linking.go:222` or `:262`
- **THEN** the handler returns 400 and does not silently use quantity 0

#### Scenario: Malformed HTMX ID returns 400
- **WHEN** a path or query ID at `handlers/htmx.go:1525` is non-numeric
- **THEN** the handler returns 400 with an invalid-ID error

### Requirement: Dead assignments are resolved

Each dead assignment (`_ =`) site SHALL be audited and either wired into the intended logic or deleted if the computation is truly dead, with the decision documented. Specifically: `handlers/dashboard.go:98`, `handlers/htmx.go:924`, `handlers/oneshot.go:1975`, `handlers/oneshot.go:3217`, `handlers/conditions.go:208`, `handlers/oneshot_import.go:84`. No `_ =` dead assignment SHALL remain that masks missing logic.

#### Scenario: Dashboard dead assignment resolved
- **WHEN** `handlers/dashboard.go:98` is inspected after the change
- **THEN** `dash.TotalMembers` is either used in the response/template or the dead computation is removed

#### Scenario: HTMX authorization dead assignment resolved
- **WHEN** `handlers/htmx.go:924` is inspected after the change
- **THEN** `n.UserID` is either checked against the caller (returning 403 on mismatch) or the assignment is removed with a comment explaining why no check is needed

#### Scenario: Oneshot dead assignments resolved
- **WHEN** `handlers/oneshot.go:1975` and `:3217` are inspected
- **THEN** `sceneCount` is used or removed and the `_ = s` no-op loop is deleted or wired to its intended side effect

#### Scenario: Conditions and import dead assignments resolved
- **WHEN** `handlers/conditions.go:208` and `handlers/oneshot_import.go:84` are inspected
- **THEN** `durStr` and `isFull` are either wired into their respective logic or the dead code is removed

### Requirement: Magic numbers are named constants (env where ops-tunable)

Inlined magic numbers SHALL be extracted to named constants (and env-overridable where ops-tunable): `handlers/ws.go` 4096 / 60s / 30s / 10s, `handlers/ai.go` 30s/60s/120s/60s timeouts with deduplicated `http.Client` construction, `middleware/auth.go` 15m backoff, `db/db.go` 60m/168h intervals, `handlers/cleanup.go` 4h, `handlers/backup.go` 1h, and paging `.Limit(50/30/10/20)` caps in `handlers/campaign.go`. Ops-tunable values SHALL read `os.Getenv` with the const as default.

#### Scenario: WebSocket limits use named constants
- **WHEN** `handlers/ws.go` is inspected
- **THEN** read limit, read deadline, ping interval, and write deadline are defined as named constants and used consistently

#### Scenario: AI client timeouts are deduplicated
- **WHEN** `handlers/ai.go` is inspected
- **THEN** per-handler timeouts are named constants and `http.Client` construction is via a single factory function

#### Scenario: Backup and cleanup intervals are named
- **WHEN** `db/db.go`, `handlers/cleanup.go`, and `handlers/backup.go` are inspected
- **THEN** sleep/tick intervals are named constants, not inline durations

### Requirement: No API contract changes and coverage-safe

The hardening SHALL NOT change any API contract except: (a) previously-silent malformed inputs now correctly return `400`, and (b) dice endpoints return `503` when the engine is unavailable. All other status codes and JSON shapes SHALL remain identical, and coverage floors SHALL stay green (Go total ≥20%, `handlers/` ≥30%, `middleware/` ≥25%).

#### Scenario: Coverage floors hold after hardening
- **WHEN** `go test -coverprofile` is run after the change
- **THEN** total, `handlers/`, and `middleware/` coverages remain at or above their floors

#### Scenario: Non-dice APIs unaffected
- **WHEN** any non-dice endpoint is called before and after the change with the same inputs
- **THEN** the response status and JSON shape are identical

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
