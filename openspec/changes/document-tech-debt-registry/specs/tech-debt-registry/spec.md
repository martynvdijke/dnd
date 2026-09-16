## ADDED Requirements

### Requirement: Tech-debt registry document exists and is discoverable
The system SHALL provide a single tech-debt registry document at `docs/tech-debt.md` that serves as the discoverable ledger of known structural debt.

#### Scenario: Newcomer discovers known debt
- **WHEN** a newcomer looks for known tech debt starting from `AGENTS.md` or the `docs/` directory
- **THEN** they can locate `docs/tech-debt.md` and enumerate all currently tracked items without searching the codebase for markers

#### Scenario: Registry is the single ledger
- **WHEN** the repository contains known structural debt items
- **THEN** those items are recorded in `docs/tech-debt.md` rather than only in scattered code comments or external trackers

### Requirement: Registry entry schema
Each entry in the registry SHALL include the following fields: ID, area, symptom, evidence (file:line references or quantitative counts), impact, owner, status, and linked OpenSpec change or PR.

#### Scenario: Entry contains all required fields
- **WHEN** a reader inspects any registry entry
- **THEN** the entry displays an ID, area, symptom, evidence, impact, owner, status, and a link to the corresponding OpenSpec change or PR

#### Scenario: Evidence is verifiable
- **WHEN** a reader follows an entry's evidence field
- **THEN** they find concrete file:line references or counts that allow them to verify the debt exists in the codebase

### Requirement: Registry seeded with known structural items
The registry SHALL be seeded with the four known structural items: ent-vs-raw-SQL data-access split, request-context / rows.Close discipline, coverage-gate documentation drift, and Go↔JS dice bundle seam, each linked to its corresponding OpenSpec change (`unify-data-access-seam`, `propagate-request-context`, `reconcile-coverage-gates`, `harden-dice-bundle-seam`).

#### Scenario: All four seed items present
- **WHEN** a reader opens `docs/tech-debt.md` after this change is applied
- **THEN** they find four seeded entries whose areas match the four structural items and whose linked-change fields point to the four named OpenSpec changes

#### Scenario: Seed entries use the entry schema
- **WHEN** a reader inspects any of the four seeded entries
- **THEN** that entry conforms to the entry schema defined above (ID, area, symptom, evidence, impact, owner, status, linked change/PR)

### Requirement: Deferral marker convention
Any deliberate simplification or deferral marked with a `ponytail:` comment, or any `TODO`/`FIXME`/`HACK`-style deferral introduced in the codebase, SHALL either reference a registry entry ID or include its ceiling and upgrade path inline.

#### Scenario: Ponytail comment references registry entry
- **WHEN** an author adds a `ponytail:` comment to mark a deliberate simplification
- **THEN** the comment either contains a registry entry ID (e.g., `TD-001`) or states its ceiling and upgrade path inline

#### Scenario: TODO-style deferral references registry entry
- **WHEN** an author introduces a `TODO` or `FIXME` or `HACK` marker to defer work
- **THEN** that marker either references a registry entry ID or includes its ceiling and upgrade path inline

#### Scenario: Existing ponytail comments brought into conformance
- **WHEN** the registry is created and the three existing `ponytail:` markers at `ts/init.ts:109`, `ts/characters/sheet.ts:59`, and `ts/app.ts:92` are evaluated
- **THEN** each is either linked to a registry entry ID or updated to include its ceiling and upgrade path inline, or an entry covering it exists in the registry

### Requirement: Registry links to OpenSpec changes without duplication
The registry SHALL link to OpenSpec changes for detailed plans and track disposition via its status field; it SHALL NOT duplicate the full change proposals, designs, or task lists.

#### Scenario: Registry entry links rather than duplicates
- **WHEN** a reader follows a registry entry's linked OpenSpec change
- **THEN** they find the full proposal, design, and tasks in the change directory, while the registry entry itself remains a summary row with a link

#### Scenario: Registry status reflects change lifecycle
- **WHEN** a linked OpenSpec change is archived or superseded
- **THEN** the corresponding registry entry's status is updated to reflect the new disposition (e.g., resolved, superseded) rather than retaining a stale active status
