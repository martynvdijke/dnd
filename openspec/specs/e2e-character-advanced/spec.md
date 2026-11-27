# e2e-character-advanced Specification

## Purpose
Playwright e2e tests for downtime activities, level planner, resources, combat log, companions.

## Requirements

### Requirement: Downtime activities e2e coverage
The Playwright e2e tests SHALL verify downtime activity management on character sheets.

#### Scenario: View downtime activities on character
- **WHEN** user opens a character's downtime section
- **THEN** existing downtime activities SHALL be listed

#### Scenario: Create downtime activity
- **WHEN** user creates a new downtime activity with type and description
- **THEN** the activity SHALL appear in the character's downtime list

#### Scenario: Edit downtime activity
- **WHEN** user edits a downtime activity
- **THEN** the activity SHALL be updated

#### Scenario: Delete downtime activity
- **WHEN** user deletes a downtime activity
- **THEN** the activity SHALL be removed

#### Scenario: Advance downtime day
- **WHEN** user advances a downtime activity by one day
- **THEN** the day counter SHALL increment

### Requirement: Level-up planner e2e coverage
The Playwright e2e tests SHALL verify the level-up planner feature on character sheets.

#### Scenario: View level-up plan
- **WHEN** user opens the level-up planner for a character
- **THEN** the current plan and available suggestions SHALL be shown

#### Scenario: Save level-up plan
- **WHEN** user selects level-up choices and saves
- **THEN** the plan SHALL be saved to the character

#### Scenario: Delete level-up plan
- **WHEN** user removes a level-up plan
- **THEN** the plan SHALL be cleared

### Requirement: Character resources e2e coverage
The Playwright e2e tests SHALL verify character resource tracking (e.g., spell slots, class features).

#### Scenario: View character resources
- **WHEN** user opens the resources section on a character sheet
- **THEN** existing resources SHALL be displayed

#### Scenario: Track resource usage
- **WHEN** user updates resource current value
- **THEN** the resource SHALL reflect the change

#### Scenario: Recover resources on rest
- **WHEN** user triggers a rest that recovers resources
- **THEN** resources SHALL be restored to their maximum values

### Requirement: Combat log e2e coverage
The Playwright e2e tests SHALL verify the combat log feature.

#### Scenario: View combat log
- **WHEN** user navigates to the combat log section
- **THEN** recent combat log entries SHALL be displayed

#### Scenario: Log combat action
- **WHEN** user logs a new combat action
- **THEN** the action SHALL appear in the combat log

#### Scenario: View combat log stats
- **WHEN** user opens the combat log stats view
- **THEN** aggregated combat statistics SHALL be shown

### Requirement: Companions e2e coverage
The Playwright e2e tests SHALL verify companion/NPC follower management.

#### Scenario: Create companion
- **WHEN** user creates a companion with name and stats
- **THEN** the companion SHALL appear in the companions list

#### Scenario: Edit companion
- **WHEN** user edits a companion's stats or details
- **THEN** the companion SHALL be updated

#### Scenario: Delete companion
- **WHEN** user deletes a companion
- **THEN** the companion SHALL be removed
