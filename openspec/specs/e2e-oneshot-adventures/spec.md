# e2e-oneshot-adventures Specification

## Purpose
Playwright e2e tests for one-shot/act/scene CRUD, session pacing, clue tracker, prep checklist, DM notes.

## Requirements

### Requirement: Act and scene CRUD e2e coverage
The Playwright e2e tests SHALL verify the creation and management of acts and scenes within one-shot adventures.

#### Scenario: Create act within one-shot
- **WHEN** user opens a one-shot adventure and clicks "Add Act"
- **THEN** a new act SHALL be created and appear in the adventure structure

#### Scenario: Edit act details
- **WHEN** user edits an act's title, description, or order
- **THEN** the act SHALL be updated

#### Scenario: Delete act
- **WHEN** user deletes an act
- **THEN** the act and its associated scenes SHALL be removed

#### Scenario: Create scene within act
- **WHEN** user opens an act and clicks "Add Scene"
- **THEN** a new scene SHALL be created

#### Scenario: Edit scene details
- **WHEN** user edits a scene's title, description, or notes
- **THEN** the scene SHALL be updated

#### Scenario: Delete scene
- **WHEN** user deletes a scene
- **THEN** the scene SHALL be removed from the act

### Requirement: Session pacing e2e coverage
The Playwright e2e tests SHALL verify live session pacing controls.

#### Scenario: Start pacing session
- **WHEN** user clicks "Start Session" on a one-shot adventure
- **THEN** a pacing session SHALL begin with current scene displayed

#### Scenario: Pause and resume pacing
- **WHEN** user clicks "Pause" during a pacing session
- **THEN** the timer SHALL pause
- **WHEN** user clicks "Resume"
- **THEN** the timer SHALL resume

#### Scenario: Advance to next scene
- **WHEN** user clicks "Next Scene" during pacing
- **THEN** the pacing SHALL advance to the next scene

#### Scenario: Complete pacing session
- **WHEN** user clicks "Complete Session" at the end
- **THEN** the pacing session SHALL end with a summary

### Requirement: Clue tracker e2e coverage
The Playwright e2e tests SHALL verify the clue/mystery tracker.

#### Scenario: Create a clue
- **WHEN** user navigates to the clue tracker and adds a clue
- **THEN** the clue SHALL appear in the tracker

#### Scenario: Reveal a clue
- **WHEN** user reveals a hidden clue
- **THEN** the clue SHALL be marked as revealed

#### Scenario: Hide a clue
- **WHEN** user hides a revealed clue
- **THEN** the clue SHALL return to hidden state

#### Scenario: Link NPC to clue
- **WHEN** user links an NPC to a clue
- **THEN** the NPC association SHALL be saved

#### Scenario: Link location to clue
- **WHEN** user links a location to a clue
- **THEN** the location association SHALL be saved

### Requirement: Prep checklist e2e coverage
The Playwright e2e tests SHALL verify the prep checklist (DM preparation items).

#### Scenario: Create prep checklist item
- **WHEN** user adds a prep checklist item with description
- **THEN** the item SHALL appear in the checklist

#### Scenario: Update prep checklist item
- **WHEN** user edits a checklist item
- **THEN** the item SHALL be updated

#### Scenario: Delete prep checklist item
- **WHEN** user deletes a checklist item
- **THEN** the item SHALL be removed

### Requirement: DM notes e2e coverage
The Playwright e2e tests SHALL verify DM quick-reference notes.

#### Scenario: Create DM note
- **WHEN** user adds a DM note with title and content
- **THEN** the note SHALL appear in the DM screen section

#### Scenario: Update DM note
- **WHEN** user edits a DM note
- **THEN** the note SHALL be updated

#### Scenario: Delete DM note
- **WHEN** user deletes a DM note
- **THEN** the note SHALL be removed
