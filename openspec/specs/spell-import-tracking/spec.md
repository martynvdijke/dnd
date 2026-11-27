# spell-import-tracking Specification

## Purpose
Visual indicators for spell import source and status (SRD, homebrew, custom import).

## Requirements

### Requirement: Import source badges on spell entries

The system SHALL display visual badges on each spell entry indicating the import source and status.

#### Scenario: SRD source badge
- **WHEN** a spell has `system = "dnd5e"` AND `source = "srd"`
- **THEN** the system SHALL display an "SRD" badge with a book icon in a muted/info style

#### Scenario: Homebrew source badge
- **WHEN** a spell has `source = "homebrew"` OR a non-empty `publisher` field
- **THEN** the system SHALL display a "Homebrew" badge or the publisher name in a distinct style (e.g., purple/custom)

#### Scenario: Custom import badge
- **WHEN** a spell has a non-empty `publisher` field that is not "homebrew" or "srd"
- **THEN** the system SHALL display the publisher name as a badge

#### Scenario: System-agnostic fallback
- **WHEN** a spell has `system` other than "dnd5e" (e.g., "pf2e")
- **THEN** the system SHALL display a badge with the system name

### Requirement: Filter by import source

The system SHALL allow filtering the spell list by import source (SRD, homebrew, custom, all).

#### Scenario: Filter by SRD only
- **WHEN** a user selects "SRD" from the source filter
- **THEN** the system SHALL display only spells with `system = "dnd5e"` AND `source = "srd"`

#### Scenario: Filter by homebrew only
- **WHEN** a user selects "Homebrew" from the source filter
- **THEN** the system SHALL display only spells with `source = "homebrew"` OR non-empty `publisher`

### Requirement: Count by source

The system SHALL display a summary count of spells by import source in the filter area.

#### Scenario: Source counts visible
- **WHEN** the spell browser loads
- **THEN** the system SHALL show a count like "245 SRD · 12 Homebrew · 3 Imported" in the filter header area
