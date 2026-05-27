## ADDED Requirements

### Requirement: Passive Investigation and Insight display
The system SHALL display passive Investigation and Insight scores alongside the existing passive Perception on the stats tab.

#### Scenario: Passive scores shown in stats section
- **WHEN** the stats tab is rendered
- **THEN** the system SHALL show passive Perception, passive Investigation (10 + INT mod + proficiency if proficient in Investigation), and passive Insight (10 + WIS mod + proficiency if proficient in Insight)

#### Scenario: Passive scores calculated from character data
- **WHEN** computing passive Investigation
- **THEN** the formula SHALL be 10 + INT modifier + proficiency bonus (if proficient in Investigation)
- **WHEN** computing passive Insight
- **THEN** the formula SHALL be 10 + WIS modifier + proficiency bonus (if proficient in Insight)

#### Scenario: Bonus displayed for each passive score
- **WHEN** displaying passive Investigation
- **THEN** the system SHALL also show the base modifier breakdown (e.g., "10 + 2 (INT) + 3 (proficient)")
- **WHEN** displaying passive Insight
- **THEN** the system SHALL also show the base modifier breakdown (e.g., "10 + 1 (WIS) + 0 (not proficient)")
