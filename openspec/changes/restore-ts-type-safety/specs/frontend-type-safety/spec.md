## ADDED Requirements

### Requirement: Generic typed API client

The `api()` helper in `ts/lib/api.ts` SHALL be generic `api<T>(method: string, path: string, body?: unknown): Promise<T>` and shared API response types SHALL be defined for common endpoints. Call sites SHALL specify a type argument rather than receiving `any`.

#### Scenario: Caller receives typed response

- **WHEN** a caller invokes `api<Character>('GET', '/api/characters/123')`
- **THEN** the returned value is typed as `Character` and property access is type-checked

#### Scenario: Request body is not any

- **WHEN** a caller passes a body to `api<T>()`
- **THEN** the body parameter is typed as `unknown` (not `any`) so untyped payloads are flagged by the compiler

#### Scenario: No any propagation from api

- **WHEN** the codebase is searched for `Promise<any>` originating from the api helper
- **THEN** no such signature remains; all api call sites that need a type provide an explicit type argument or handle `unknown`

### Requirement: Typed window bridge

The `expose()` helper and its `window` bridge SHALL use a typed `Window` augmentation instead of `(window as any)` casts, so that exposed names are type-checked.

#### Scenario: Exposed names are type-checked

- **WHEN** a module calls `expose('myFeature', value)` and another reads `window.myFeature`
- **THEN** both the expose call and the window access are type-checked against the augmented `Window` interface

#### Scenario: No window as any casts remain in bridge code

- **WHEN** the bridge code in `ts/lib/expose.ts` and its consumers are inspected
- **THEN** no `(window as any)` casts remain for names covered by the augmentation

### Requirement: Safe DOM lookup helper

A central DOM lookup helper (`$id()`/`getEl()` or equivalent) SHALL be provided that returns a properly typed element (or `null`) without requiring a non-null assertion (`!`). New and migrated code SHALL use this helper instead of `document.getElementById(...)!`.

#### Scenario: Helper returns typed nullable element

- **WHEN** a caller invokes `$id<HTMLInputElement>('my-input')`
- **THEN** the return is `HTMLInputElement | null` and the compiler requires null handling

#### Scenario: Non-null assertions reduced

- **WHEN** migrated files are inspected for `getElementById(... )!`
- **THEN** such assertions are replaced by the helper (or justified with a throwing variant that has a clear error message)

### Requirement: Incremental removal of ts-nocheck

Each in-scope file that currently carries `// @ts-nocheck` SHALL have the pragma removed and all resulting type errors fixed without adding new `@ts-ignore`/`@ts-expect-error`/`@ts-nocheck` suppressions. Removal SHALL proceed leaf-first.

#### Scenario: Leaf file detyped cleanly

- **WHEN** `// @ts-nocheck` is removed from a leaf file such as `ts/characters/stats.ts`
- **THEN** `npm run typecheck` passes for that file with no new suppressions and behavior is unchanged

#### Scenario: No new suppressions added

- **WHEN** a file is migrated to strict mode
- **THEN** no `@ts-ignore`, `@ts-expect-error`, or `@ts-nocheck` is added to silence errors; errors are fixed by adding proper types

#### Scenario: app.ts and admin.ts remain out of scope

- **WHEN** this change is complete
- **THEN** `ts/app.ts` and `ts/admin.ts` still carry `// @ts-nocheck` (removal is deferred to the monolith-split follow-up)

### Requirement: CI guard against new ts-nocheck

A CI guard SHALL fail the build if a new `// @ts-nocheck` pragma appears outside the current allowlist (which at completion is exactly `ts/app.ts` and `ts/admin.ts`).

#### Scenario: New nocheck pragma fails CI

- **WHEN** a file outside the allowlist contains `// @ts-nocheck` and CI runs
- **THEN** the guard reports the offending file and the build fails

#### Scenario: Allowlist shrinks as files migrate

- **WHEN** a file is successfully migrated and its pragma removed
- **THEN** the guard's allowlist is updated to no longer permit that file, so re-adding the pragma would fail CI
