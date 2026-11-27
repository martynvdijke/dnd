# test-data-factory Specification

## Purpose
Reusable helper package for creating test fixtures in a composable way.

## Requirements

### Requirement: Test data factory package

There SHALL be a `handlers/testutil/` package providing composable helper functions for test fixture creation.

#### Scenario: CreateCharacter returns complete seeded character
- **WHEN** `testutil.CreateCharacter(t, db)` is called with default attributes
- **THEN** the returned character SHALL have: non-empty name, race="Human", class="Fighter", level=1, all abilities=10, hp=12, ac=10, speed=30

#### Scenario: CreateCharacter accepts override attributes
- **WHEN** `testutil.CreateCharacter(t, db, testutil.WithName("Aragorn"), testutil.WithClass("Ranger"), testutil.WithLevel(5))` is called
- **THEN** the returned character SHALL have name="Aragorn", class="Ranger", level=5

#### Scenario: CreateCampaign returns campaign with owner
- **WHEN** `testutil.CreateCampaign(t, db, userID)` is called
- **THEN** the returned campaign SHALL have non-empty name and the given user as owner

#### Scenario: SeedSpells populates compendium
- **WHEN** `testutil.SeedSpells(t, db, 10)` is called
- **THEN** exactly 10 spells SHALL exist in the compendium_spells table

#### Scenario: NewTestRequest creates valid HTTP request
- **WHEN** `testutil.NewTestRequest("POST", "/api/characters", bodyJSON)` is called
- **THEN** the returned request SHALL have Content-Type: application/json and the body serialized

#### Scenario: WithTestAuth adds auth context
- **WHEN** `testutil.WithTestAuth(req, 1, "admin")` is called on a request
- **THEN** the request SHALL have a valid session cookie and X-CSRF-Token header
