# dm-tooling Specification

## Purpose
TBD - created by archiving change add-dm-tooling. Update Purpose after archive.
## Requirements
### Requirement: Treasure hoard generation

The system SHALL provide `GET /api/generate/treasure` that returns a DMG-style treasure hoard for a tier. The `tier` query parameter SHALL accept `1`–`4` (default `1`) mapping to CR bands 0–4, 5–10, 11–16, and 17+. The response SHALL include `tier`, `coins` with `cp`, `sp`, `ep`, `gp`, `pp`, `gems`, `art_objects`, `magic_items`, and `total_gp_value`.

#### Scenario: Hoard for a tier

- **WHEN** a client requests `/api/generate/treasure?tier=2`
- **THEN** the response has `tier` equal to `2`, a `coins` object with all five denominations, and arrays for `gems`, `art_objects`, and `magic_items`

#### Scenario: Unknown tier falls back

- **WHEN** a client requests `/api/generate/treasure?tier=99`
- **THEN** the response is a valid hoard with `tier` equal to `1`

### Requirement: Biome-aware weather

The system SHALL accept a `biome` query parameter on `GET /api/generate/weather` and SHALL include the resolved `biome` in the response. Supported biomes SHALL include `temperate`, `arctic`, `desert`, `forest`, `mountain`, `swamp`, `coast`, and `underdark`. An unknown biome SHALL fall back to `temperate`. Biome SHALL constrain temperature and precipitation (for example, `desert` SHALL NOT produce snow and `arctic` SHALL favour snow).

#### Scenario: Desert never snows

- **WHEN** a client requests `/api/generate/weather?biome=desert` repeatedly
- **THEN** no response has a precipitation of `Snow` or `Heavy Snow`

#### Scenario: Unknown biome falls back

- **WHEN** a client requests `/api/generate/weather?biome=volcano`
- **THEN** the response `biome` is `temperate`

### Requirement: Name cultures

The system SHALL support `orc`, `dragonborn`, `tiefling`, and `gnome` in addition to the existing `dwarf`, `elf`, `halfling`, and default human cultures on `GET /api/generate/name`. The response SHALL include `name`, `first`, `last`, and `race`.

#### Scenario: Orc name

- **WHEN** a client requests `/api/generate/name?race=orc`
- **THEN** the response has a non-empty `name` and `race` equal to `orc`

### Requirement: Adventuring-day XP budget

The system SHALL provide `GET /api/encounters/daily-budget` accepting a comma-separated `levels` query parameter. The response SHALL include `levels`, `per_level` (each with `level` and `budget`), `party_budget`, and `suggested_encounters`. Levels outside 1–20 SHALL be ignored.

#### Scenario: Party budget

- **WHEN** a client requests `/api/encounters/daily-budget?levels=1,1,1,1`
- **THEN** `party_budget` equals four times the level-1 daily budget and `suggested_encounters` is a non-empty range

#### Scenario: No valid levels

- **WHEN** a client requests `/api/encounters/daily-budget?levels=99`
- **THEN** `party_budget` is `0` and `per_level` is empty

### Requirement: One-shot generators in the DM tools modal

The DM tools modal SHALL expose the five one-shot generators — adventure hook, dungeon dressing, tavern, urban encounter, and road encounter — calling the existing generator endpoints and displaying the generated result in the modal.

#### Scenario: Generate an adventure hook

- **WHEN** the DM activates the adventure hook generator
- **THEN** the modal displays a generated hook from the adventure-hook endpoint

#### Scenario: Generate dungeon dressing

- **WHEN** the DM activates the dungeon dressing generator
- **THEN** the modal displays generated dressing entries

#### Scenario: Generate a tavern

- **WHEN** the DM activates the tavern generator
- **THEN** the modal displays a generated tavern

#### Scenario: Generate an urban encounter

- **WHEN** the DM activates the urban encounter generator
- **THEN** the modal displays a generated urban encounter

#### Scenario: Generate a road encounter

- **WHEN** the DM activates the road encounter generator
- **THEN** the modal displays a generated road encounter
