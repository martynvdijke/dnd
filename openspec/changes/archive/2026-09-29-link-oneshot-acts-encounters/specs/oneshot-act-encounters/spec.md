## ADDED Requirements

### Requirement: Link encounters to a one-shot act
The service SHALL allow an encounter (`EncounterTemplate`) to be linked to a specific `OneShotAct` within a one-shot adventure, in addition to the existing adventure-level links.

#### Scenario: Link an encounter to an act

- WHEN an authenticated DM posts an encounter id to the act-scoped link endpoint
- THEN a row is stored in `oneshot_adventure_encounters` with the act's `adventure_id` and the given `act_id`
- AND the link is idempotent (re-linking the same pair does not error or duplicate)

#### Scenario: List encouners for an act

- WHEN the DM requests the encounters for an act
- THEN only links whose `act_id` equals that act are returned, each with its encounter name

#### Scenario: Unlink an encounter from an act

- WHEN the DM deletes an act-scoped encounter link
- THEN that row is removed and the encounter is no longer listed for the act

#### Scenario: Adventure-level links unaffected

- WHEN an encounter is linked without an `act_id`
- THEN the row keeps `act_id NULL` and continues to appear in the adventure-level encounter list

### Requirement: Act encounter picker in the one-shot UI
The one-shot act tree SHALL expose an act-scoped encounter picker that lists linked encounters and offers the adventure's candidate encounters.

#### Scenario: Open the picker

- WHEN the DM opens the encounters picker for an act
- THEN the fragment lists the act's currently linked encounters with unlink controls and a select of candidate encounters for the adventure's campaign (or the owner's encounters when standalone)

#### Scenario: Link from the picker

- WHEN the DM selects an encounter and submits the picker form
- THEN the encounter is linked to the act and the fragment re-renders showing it as linked

### Requirement: Scene encounter reference
The scene edit form SHALL let the DM choose an encounter for a scene.

#### Scenario: Set a scene encounter

- WHEN the DM selects an encounter in the scene form and saves
- THEN the scene's `encounter_id` is persisted and the encounter name badge is shown on the act tree
