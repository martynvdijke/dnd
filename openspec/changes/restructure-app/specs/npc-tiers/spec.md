## ADDED Requirements

### Requirement: Simple NPC tier
NPCs SHALL support a "simple" mode with only name, race, class, description, backstory, and notes (no full stat block). Simple NPCs have the is_full field set to false.

#### Scenario: Create simple NPC
- **WHEN** a user creates an NPC with is_full=false, name, and backstory
- **THEN** the system creates a simple NPC (stats default to 10)

#### Scenario: Simple NPC display
- **WHEN** a user views a simple NPC
- **THEN** the system shows name, race, class, description, and backstory (no stat block)

### Requirement: Full NPC tier
NPCs SHALL support a "full" mode with complete character-sheet-like stats. Full NPCs have is_full set to true and include: str, dex, con, int, wis, cha, hp_max, hp_current, ac, speed, skills, saves, features, actions.

#### Scenario: Create full NPC
- **WHEN** a user creates an NPC with is_full=true and full stat fields
- **THEN** the system creates a full NPC with all stat fields

#### Scenario: Full NPC display
- **WHEN** a user views a full NPC
- **THEN** the system shows the complete stat block with skills, saves, features, actions

### Requirement: NPC schema migration
The existing NPC schema SHALL gain an `is_full` boolean field (default false) and additional full-stat fields. Existing NPCs default to simple mode.

#### Scenario: Existing NPC shows as simple
- **WHEN** a user views an NPC created before the schema change
- **THEN** the system displays the NPC in simple mode (is_full defaults to false)
