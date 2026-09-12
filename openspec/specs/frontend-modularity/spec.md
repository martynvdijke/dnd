# frontend-modularity Specification

## Purpose
TBD - created by archiving change split-ts-monolith. Update Purpose after archive.
## Requirements
### Requirement: Thin orchestrator entry points

After the split, `ts/app.ts` and `ts/admin.ts` SHALL each be fewer than 800 lines and act as thin orchestrators that import domain modules for side effects. Domain logic SHALL live in extracted modules, not in the orchestrators.

#### Scenario: Line count targets met

- **WHEN** lines are counted with `wc -l ts/app.ts ts/admin.ts` after the split
- **THEN** each file is under 800 lines

#### Scenario: Orchestrator imports domain modules

- **WHEN** `ts/app.ts` (or its barrel `ts/init.ts`) is inspected
- **THEN** it imports each extracted domain module for side effects and contains no domain implementation beyond wiring and shared state

### Requirement: Domain modules self-register via expose pattern

Each extracted domain module SHALL self-register its public functions via `expose()` from `ts/lib/expose.ts` on import, following the established precedent (`ts/party.ts`, `ts/combat-tracker.ts`, `ts/characters/*`, `ts/fab.ts`). The module SHALL be imported for side effects by the orchestrator/init entry point.

#### Scenario: Module self-registers on import

- **WHEN** an extracted module (e.g. shops, wiki, one-shot) is imported
- **THEN** its public functions are available on `window` via `expose()` and callable from existing call sites without additional wiring

#### Scenario: No new module framework introduced

- **WHEN** the extraction is reviewed
- **THEN** no new state-management library, DI container, or alternative module system has been introduced; the `expose()` + `init.ts` side-effect import pattern is reused

### Requirement: Behavior preserved byte-identically per phase

Each extraction phase SHALL preserve existing behavior without observable changes. Verification after each phase SHALL include `npm run typecheck`, `npm run test:unit`, and `task test:e2e` all passing.

#### Scenario: Typecheck passes after each phase

- **WHEN** a domain cluster has been extracted and `npm run typecheck` is run
- **THEN** it passes (subject to existing `// @ts-nocheck` pragmas retained on moved code)

#### Scenario: Unit and e2e suites pass after each phase

- **WHEN** `npm run test:unit` and `task test:e2e` are run after a phase
- **THEN** all existing tests pass with no new failures attributed to the move

#### Scenario: Vite build still produces 5 bundles

- **WHEN** `npm run build:vite` is run after the split
- **THEN** the 5 IIFE bundles (app, admin, pwa, setup, login) are produced successfully

### Requirement: Duplication between app.ts and characters modules reconciled

Where a function exists both in `ts/app.ts` and in `ts/characters/*`, the extraction SHALL reconcile the duplication by keeping the `ts/characters/*` version as canonical and removing the `app.ts` copy, updating call sites to use the canonical version.

#### Scenario: Duplicate function deduplicated

- **WHEN** an `app.ts` seam containing a function also defined in `ts/characters/*` is extracted
- **THEN** only one definition remains (in `ts/characters/*`) and all call sites resolve to it without runtime error

### Requirement: Shared helpers extracted for cross-cutting patterns

The split SHALL extract `refreshChar()` (replacing the `setCurrentChar(await api('GET', /api/characters/${id}))` pattern) and `renderError()` (replacing the `catch(e){toast(e.message,true)}` pattern) as shared helpers used by extracted modules.

#### Scenario: Character refresh uses shared helper

- **WHEN** a domain module needs to re-fetch and set the current character
- **THEN** it calls `refreshChar(id)` instead of inlining `setCurrentChar(await api('GET', ...))`

#### Scenario: Error handling uses shared helper

- **WHEN** a domain module handles an async error for user display
- **THEN** it calls `renderError(e)` instead of inlining `catch(e){toast(e.message,true)}`

### Requirement: Existing nocheck pragmas retained on moved code

Moved code that carried `// @ts-nocheck` in its original file SHALL retain the pragma in its new location. This change SHALL NOT remove `@ts-nocheck` pragmas — that is handled by `restore-ts-type-safety`.

#### Scenario: Moved code keeps pragma

- **WHEN** a seam is extracted from a file that had `// @ts-nocheck`
- **THEN** the new module file also carries `// @ts-nocheck` at the top and `npm run typecheck` behavior is unchanged for that code
