## MODIFIED Requirements

### Requirement: One-shot generators in the DM tools modal

The DM tools modal SHALL expose the five one-shot generators — adventure hook, dungeon dressing, tavern, urban encounter, and road encounter — calling the existing generator endpoints and displaying the generated result in the modal. For generators that produce campaign entities (tavern, adventure hook, NPC name, and treasure hoard), the modal SHALL provide a 1-click Save to Campaign button (`data-testid="dm-tools-save-btn"`) that persists the generated result into the active campaign's locations, quests, NPCs, or party inventory.

#### Scenario: Generate an adventure hook

- **WHEN** the DM activates the adventure hook generator
- **THEN** the modal displays a generated hook from the adventure-hook endpoint

#### Scenario: Save a generated tavern or hook to the campaign

- **WHEN** the DM generates a tavern or adventure hook and clicks the Save to Campaign button
- **THEN** the generated tavern is created via `POST /api/locations` (or the hook via `POST /api/quests`) and a confirmation toast is shown

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
