## 1. Inventory

- [x] 1.1 List all `context.TODO()` / `context.Background()` sites in `handlers/**` and classify each as request-scoped (must use request context) or genuinely detached (owned context + comment).
- [x] 1.2 List all `rows.Close()` calls and flag the ones that are not deferred.

## 2. Context propagation

- [x] 2.1 Replace request-scoped `context.TODO()` / `context.Background()` with `c.Request.Context()` (or a derived context) across `handlers/**`.
- [x] 2.2 For genuinely detached work (startup/schedulers), keep an owned context and add a comment explaining why it is not request-scoped.

## 3. Row lifecycle

- [x] 3.1 Move each non-deferred `rows.Close()` to `defer rows.Close()` immediately after a successful query.
- [x] 3.2 Where a scan loop is duplicated across handlers, extract one small iteration helper that owns `Close` and checks `rows.Err()`.

## 4. Tests + guard

- [x] 4.1 Add a test proving client cancellation cancels an in-flight handler query.
- [x] 4.2 Add a test exercising the early-return path and asserting no connection leaks.
- [x] 4.3 Add a CI lint (under `scripts/ci/`, referenced from `ci.yaml`) forbidding new `context.TODO()/Background()` in request-scoped handler DB paths, with a documented allowlist comment.

## 5. Verification

- [x] 5.1 Run `task ci`; confirm all coverage floors hold (Go total ≥20%, `handlers/` ≥40%, `middleware/` ≥40%, vitest ≥35% statements/lines/functions, ≥30% branches).
- [x] 5.2 Confirm no HTTP contract changed (route snapshot + e2e green).
