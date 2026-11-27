# e2e-campaign-features Specification

## Purpose
Playwright e2e tests for campaign dashboard, wiki, recaps, calendar/timeline, faction reputation, conditions.

## Requirements

### Requirement: Campaign dashboard e2e coverage
The Playwright e2e tests SHALL verify the campaign dashboard view.

#### Scenario: View campaign dashboard
- **WHEN** user navigates to a campaign's dashboard
- **THEN** the dashboard SHALL display campaign overview, recent activity, and key stats

### Requirement: Campaign wiki e2e coverage
The Playwright e2e tests SHALL verify the campaign wiki page CRUD.

#### Scenario: Create wiki page
- **WHEN** user creates a wiki page with title and content in a campaign
- **THEN** the wiki page SHALL be created and viewable

#### Scenario: Edit wiki page
- **WHEN** user edits a wiki page's content
- **THEN** the wiki page SHALL be updated

#### Scenario: Delete wiki page
- **WHEN** user deletes a wiki page
- **THEN** the wiki page SHALL be removed

#### Scenario: View wiki page
- **WHEN** user opens a wiki page from the list
- **THEN** the wiki page content SHALL be displayed

### Requirement: Campaign recaps e2e coverage
The Playwright e2e tests SHALL verify campaign recap management.

#### Scenario: Create campaign recap
- **WHEN** user creates a recap with session summary and highlights
- **THEN** the recap SHALL appear in the campaign recaps list

#### Scenario: Edit campaign recap
- **WHEN** user edits an existing recap
- **THEN** the recap SHALL be updated

#### Scenario: Delete campaign recap
- **WHEN** user deletes a recap
- **THEN** the recap SHALL be removed

#### Scenario: Generate recap with AI
- **WHEN** user clicks "Generate Recap" on a campaign
- **THEN** an AI-generated recap SHALL be created (or graceful fallback shown)

### Requirement: Calendar and timeline e2e coverage
The Playwright e2e tests SHALL verify the in-game calendar and timeline features.

#### Scenario: View calendar
- **WHEN** user navigates to the calendar section
- **THEN** the calendar SHALL display existing events

#### Scenario: Create calendar event
- **WHEN** user creates a calendar event with title and date
- **THEN** the event SHALL appear on the calendar

#### Scenario: Edit calendar event
- **WHEN** user edits a calendar event
- **THEN** the event SHALL be updated

#### Scenario: Delete calendar event
- **WHEN** user deletes a calendar event
- **THEN** the event SHALL be removed

#### Scenario: View timeline
- **WHEN** user navigates to the timeline section
- **THEN** the timeline SHALL display chronological events

#### Scenario: Create timeline event
- **WHEN** user creates a timeline event
- **THEN** the event SHALL appear in the timeline

#### Scenario: Delete timeline event
- **WHEN** user deletes a timeline event
- **THEN** the event SHALL be removed from the timeline

### Requirement: Faction reputation e2e coverage
The Playwright e2e tests SHALL verify faction reputation management.

#### Scenario: View faction reputations
- **WHEN** user navigates to the factions section
- **THEN** existing factions and reputation scores SHALL be displayed

#### Scenario: Set faction reputation
- **WHEN** user sets a reputation score for a faction
- **THEN** the score SHALL be saved and displayed

#### Scenario: Delete faction reputation
- **WHEN** user deletes a faction reputation entry
- **THEN** the entry SHALL be removed

### Requirement: Conditions e2e coverage
The Playwright e2e tests SHALL verify conditions tracking on the combat tracker.

#### Scenario: View conditions on combat entry
- **WHEN** user opens the combat tracker
- **THEN** conditions attached to combat entries SHALL be visible

#### Scenario: Add condition to character
- **WHEN** user adds a condition to a character in combat
- **THEN** the condition SHALL appear on that character's entry
