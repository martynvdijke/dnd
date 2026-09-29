# oneshot-prep-run-ui Specification

## Purpose
TBD - created by archiving change wire-oneshot-prep. Update Purpose after archive.
## Requirements
### Requirement: One-shot prep and run entry points

The one-shot detail view SHALL expose entry points to the prep dashboard, DM screen, clue board, pregen manager, session-flow view, and the run (pacing) flow. Activating an entry point SHALL render the target surface inside the one-shot view, and the target surface SHALL offer a way back to the adventure detail.

#### Scenario: Open the prep dashboard

- **WHEN** the DM activates the Prep Dashboard action on an adventure
- **THEN** the prep dashboard fragment for that adventure is rendered

#### Scenario: Open the DM screen

- **WHEN** the DM activates the DM Screen action on an adventure
- **THEN** the one-shot DM screen for that adventure is rendered

#### Scenario: Open the clue board

- **WHEN** the DM activates the Clues action on an adventure
- **THEN** the clue board for that adventure is rendered

#### Scenario: Open the pregens

- **WHEN** the DM activates the Pregens action
- **THEN** the pregenerated-character list is rendered

#### Scenario: Open the session flow

- **WHEN** the DM activates the Session Flow action on an adventure
- **THEN** the print-friendly session flow for that adventure is rendered

#### Scenario: Start a run

- **WHEN** the DM activates the Run action for an adventure that has no active pacing session
- **THEN** a pacing session is started and the live pacing dashboard is rendered

#### Scenario: Resume a run

- **WHEN** the DM activates the Run action for an adventure that already has a running or paused pacing session
- **THEN** the existing session is reused and the live pacing dashboard is rendered

### Requirement: Prep dashboard links resolve to working surfaces

Every action on the prep dashboard SHALL target an existing route or control, and the generators action SHALL open the DM tools modal.

#### Scenario: Add clues from the prep dashboard

- **WHEN** the DM activates the Add clues action while the clue list is empty
- **THEN** the clue board for the adventure is rendered

#### Scenario: Open generators from the prep dashboard

- **WHEN** the DM activates the Generators action
- **THEN** the DM tools modal opens with the generator tools available

### Requirement: Live pacing clock

Pacing elapsed time SHALL be derived from persisted timestamps, not only from browser-side counting. While a session is running, its elapsed time SHALL advance; pausing SHALL freeze the elapsed value; resuming SHALL continue from the frozen value; advancing to the next scene SHALL keep the total elapsed time and record per-scene durations; re-rendering or reloading the dashboard SHALL display the accumulated elapsed time.

#### Scenario: Elapsed advances while running

- **WHEN** a pacing session has been running for a period
- **THEN** the dashboard shows an elapsed time greater than the value at start

#### Scenario: Pause freezes elapsed

- **WHEN** the DM pauses a running session
- **THEN** the dashboard shows the elapsed time accumulated up to the pause and stops advancing

#### Scenario: Resume continues from the frozen value

- **WHEN** the DM resumes a paused session
- **THEN** the elapsed time continues from the paused value rather than resetting

#### Scenario: Reload preserves elapsed

- **WHEN** the dashboard is re-rendered or the page is reloaded during a running session
- **THEN** the displayed elapsed time reflects the time since the session started, not zero

#### Scenario: Next scene records the completed scene duration

- **WHEN** the DM advances to the next scene
- **THEN** the previous scene's timing is completed with its actual duration and the total elapsed time is unchanged

### Requirement: Adventure-scoped pacing lookup

`GET /api/oneshot-adventures/:id/pacing` SHALL return the adventure's latest pacing session, preferring a running or paused session over a completed one, and SHALL return 404 when the adventure has no session.

#### Scenario: Active session is returned

- **WHEN** the adventure has a running or paused pacing session
- **THEN** the response contains that session with its status and elapsed time

#### Scenario: No session returns 404

- **WHEN** the adventure has never had a pacing session
- **THEN** the response status is 404
