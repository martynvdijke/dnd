## ADDED Requirements

### Requirement: Encounter difficulty calculator
The system SHALL display a difficulty label (easy/medium/hard/deadly) on the encounter builder based on total monster XP vs party level and size.

#### Scenario: Difficulty shown on encounter
- **WHEN** the encounter builder shows an encounter with monsters and a party level entered
- **THEN** the system SHALL compute and display "Easy", "Medium", "Hard", or "Deadly" based on DMG XP thresholds

#### Scenario: Party level and size input
- **WHEN** viewing an encounter
- **THEN** the system SHALL show input fields for average party level and number of party members
- **WHEN** either value changes
- **THEN** the difficulty label SHALL recalculate immediately

#### Scenario: XP thresholds match DMG
- **WHEN** computing difficulty
- **THEN** the system SHALL use the DMG XP thresholds per character level (Easy: 25×level, Medium: 50×level, Hard: 75×level, Deadly: 100×level for level 1; scaling by the standard DMG table)

#### Scenario: Adjusted XP for multiple monsters
- **WHEN** computing difficulty
- **THEN** the system SHALL apply the DMG encounter multiplier based on number of monsters (2: ×1.5, 3-6: ×2, 7-10: ×2.5, 11-14: ×3, 15+: ×4)

#### Scenario: Budget meters shown
- **WHEN** difficulty is displayed
- **THEN** the system SHALL show a visual progress bar from Easy through Deadly thresholds with the current adjusted XP position marked
