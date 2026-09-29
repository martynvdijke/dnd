## ADDED Requirements

### Requirement: One-shot generators in the DM tools modal

The DM tools modal SHALL expose the five one-shot generators — adventure hook, dungeon dressing, tavern, urban encounter, and road encounter — calling the existing generator endpoints and displaying the generated result in the modal.

#### Scenario: Generate an adventure hook

- **WHEN** the DM activates the adventure hook generator
- **THEN** the modal displays a generated hook from the adventure-hook endpoint

#### Scenario: Generate dungeon dressing

- **WHEN** the DM activates the dungeon dressing generator
- **THEN** the modal displays generated dressing entries

#### Scenario: Generate a tavern

- **WHEN** the DM activates the tavern generator
- **THEN** the modal displays a generated tavern

#### Scenario: Generate an urban encounter

- **WHEN** the DM activates the urban encounter generator
- **THEN** the modal displays a generated urban encounter

#### Scenario: Generate a road encounter

- **WHEN** the DM activates the road encounter generator
- **THEN** the modal displays a generated road encounter
