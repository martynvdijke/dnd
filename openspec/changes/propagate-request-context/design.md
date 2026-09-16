## Context

Handlers run on Gin and have a request context available at `c.Request.Context()`. Two anti-patterns recur:

1. **Detached contexts** — ~22 `context.TODO()` / `context.Background()` uses initiate DB work, so cancellation and deadlines never reach the driver.
2. **Non-deferred row closes** — ~188 `rows.Close()` calls, ~159 of them deferred. The remainder sit on paths that may return early, risking a held connection under error load.

The existing `backend-error-resilience` spec covers swallowed errors, dead assignments, and magic numbers, but says nothing about context propagation or row lifecycle. This change adds that coverage.

## Goals / Non-Goals

**Goals:**
- Request cancellation and deadlines reach all request-scoped DB work.
- No `*sql.Rows` can leak on an early return.
- Add regression tests for both.

**Non-Goals:**
- Introducing a new context framework or request-scoped dependency injection.
- Rewriting data-access style (that is `unify-data-access-seam`).
- Changing any response contract.

## Decisions

### Decision: Use the request context directly

`c.Request.Context()` already carries cancellation and deadline. Threading it through is a one-argument change per call site; no new abstraction is warranted. Derived contexts are used only when a handler adds a timeout.

### Decision: Defer immediately after a successful Query

The rule is mechanical: `rows, err := ...; if err != nil {...}; defer rows.Close()`. This is the standard library idiom and needs no helper. A small iteration helper is introduced only where several handlers repeat the same scan loop — not as a framework.

### Decision: Detect regressions by lint, not convention

A grep-based CI check forbids new `context.TODO()/Background()` in handler DB paths. Cheap, and consistent with the repo's existing `lint:e2e` grep guard.

### Decision: Legitimate detached work is explicit

Startup and scheduler tasks that outlive requests keep an owned context, with a comment. The rule is "no request-scoped work on a detached context", not "never use Background".

## Risks / Trade-offs

- **Cancellation may surface latent errors** previously hidden by detached contexts (queries that were silently completing). Mitigation: run the full suite and e2e; treat new context-cancelled errors as the correct behavior.
- **Mechanical churn across many files.** Mitigation: scope to the flagged sites; keep the diff limited to context arguments and `defer` placement.
- **Lint false positives** on background jobs. Mitigation: allowlist with a documented comment, mirroring the e2e lint's escape hatch.
