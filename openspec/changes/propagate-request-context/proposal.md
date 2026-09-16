## Why

Handlers frequently detach database work from the request lifecycle: `context.TODO()` / `context.Background()` appears ~22 times in `handlers/**`, so DB calls and downstream work ignore client cancellation and deadlines. Separately, `rows.Close()` is called ~188 times but only ~159 are deferred, leaving early-return error paths that can leak a connection. Both are bounded, mechanical correctness issues that get worse as handlers grow.

## What Changes

- Pass `c.Request.Context()` (or a derived context) into every DB and downstream call in handlers, replacing `context.TODO()` / `context.Background()`.
- Ensure every `rows` returned by `Query`/`QueryRow` has `defer rows.Close()` immediately after a successful query, before any error check that returns.
- Replace the non-deferred `rows.Close()` sites with deferred closes (or a shared iteration helper).
- Add tests for cancellation propagation and for the early-return path that previously skipped `Close`.
- No API contract change; behavior identical except that abandoned requests now release DB resources and stop work.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `backend-error-resilience`: add requirements for request-context propagation and SQL row lifecycle (the existing spec covers swallowed errors and dead assignments, not context/rows).

## Impact

- `handlers/**` (all files using raw SQL and `context.TODO/Background`), `middleware/**`.
- CI gains a lint/tests asserting no new `context.TODO/Background` in handler DB paths and correct `rows.Close` deferral.
