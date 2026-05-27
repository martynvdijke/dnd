## ADDED Requirements

### Requirement: Exhaustion level display
The system SHALL display the character's current exhaustion level (0-6) on the combat or stats tab with the associated mechanical effects.

#### Scenario: Exhaustion shown as level track
- **WHEN** a character has `exhaustion_level = 0`
- **THEN** the system SHALL show "Exhaustion: None"
- **WHEN** a character has `exhaustion_level` between 1 and 6
- **THEN** the system SHALL show "Exhaustion: Level N" with the corresponding effect text

#### Scenario: Exhaustion effects displayed per level
- **WHEN** displaying exhaustion level 1
- **THEN** the system SHALL show "Disadvantage on ability checks"
- **WHEN** displaying exhaustion level 2
- **THEN** the system SHALL show "Speed halved"
- **WHEN** displaying exhaustion level 3
- **THEN** the system SHALL show "Disadvantage on attack rolls and saving throws"
- **WHEN** displaying exhaustion level 4
- **THEN** the system SHALL show "Hit point maximum halved"
- **WHEN** displaying exhaustion level 5
- **THEN** the system SHALL show "Speed reduced to 0"
- **WHEN** displaying exhaustion level 6
- **THEN** the system SHALL show "Death"

#### Scenario: Adjust exhaustion level from UI
- **WHEN** a user clicks an up/down button next to the exhaustion display
- **THEN** the system SHALL send a POST to update the character's exhaustion_level, AND the display SHALL update with the new level and effects

#### Scenario: Long rest resets exhaustion
- **WHEN** a long rest is logged via the rest log
- **THEN** the system SHALL reduce exhaustion_level by 1 (minimum 0) AND update the display
