## ADDED Requirements

### Requirement: Faction reputation visualization
The system SHALL display character reputation with factions as visual bars with tier labels instead of raw numbers.

#### Scenario: Reputation shown as bar with label
- **WHEN** viewing a faction in the factions view
- **THEN** the system SHALL show each character's reputation as a horizontal bar, with the reputation value mapped to a tier label

#### Scenario: Reputation tier labels
- **WHEN** a character's reputation with a faction is:
  - -50 or less: "Hostile" (red)
  - -49 to -1: "Unfriendly" (orange)
  - 0 to 9: "Neutral" (gray)
  - 10 to 24: "Friendly" (green)
  - 25+: "Allied" (gold)

#### Scenario: Reputation value shown on hover
- **WHEN** hovering over a reputation bar
- **THEN** the system SHALL show a tooltip with the exact reputation value

#### Scenario: Reputation bar fills proportionally
- **WHEN** displaying the reputation bar
- **THEN** the fill width SHALL be proportional within the visible range (-50 to +50 scale), with the tier color filling from left
