# oneshot-editing-reliability Specification

## Purpose
TBD - created by archiving change wire-oneshot-prep. Update Purpose after archive.
## Requirements
### Requirement: Act notes and act fields save independently

Saving the inline act notes SHALL update only the act's notes. Editing an act through the act form SHALL update its title, description, estimated minutes, number, and sort order. A notes-only save SHALL NOT modify any other act field.

#### Scenario: Notes-only save preserves the other act fields

- **WHEN** the DM edits the notes of an act that has a title, description, and estimated minutes and saves
- **THEN** the notes are updated and the title, description, and estimated minutes remain unchanged

#### Scenario: Act form updates act fields

- **WHEN** the DM submits the act edit form with new values
- **THEN** the act's title, description, estimated minutes, number, and sort order reflect the submitted values

### Requirement: Act DM notes can be added from act details

The act details surface SHALL allow adding a DM note with a title and content, and the new note SHALL appear in the act's DM notes list.

#### Scenario: Add an act DM note

- **WHEN** the DM adds a DM note from the act details surface
- **THEN** the note is persisted for that act and rendered in the DM notes list

### Requirement: Scene dialog reordering

The scene dialog list SHALL support drag reordering, and the new order SHALL be persisted so it survives a re-render.

#### Scenario: Drag a dialog to a new position

- **WHEN** the DM drags a scene dialog to a different position
- **THEN** the persisted dialog order matches the new arrangement and re-opening the scene shows that order

### Requirement: Pregen create and edit from the pregens surface

The pregens surface SHALL let a user create a pregenerated character and edit an existing one; both actions SHALL persist the field values entered in the form.

#### Scenario: Create a pregen

- **WHEN** the user fills in the new-pregen form and saves
- **THEN** the character appears in the pregen list with the entered values

#### Scenario: Edit a pregen

- **WHEN** the user edits a pregen's values and saves
- **THEN** the pregen list shows the updated values

### Requirement: Red-herring flag editable from the clue form

The clue create and edit forms SHALL include a red-herring control, and the saved clue SHALL reflect the selected flag.

#### Scenario: Mark a new clue as a red herring

- **WHEN** the DM creates a clue with the red-herring control checked
- **THEN** the clue is stored and displayed as a red herring

#### Scenario: Toggle the flag on an existing clue

- **WHEN** the DM clears the red-herring control on a clue marked as a red herring and saves
- **THEN** the clue is no longer displayed as a red herring

### Requirement: Compendium monster unlink works

Unlinking a compendium monster from a one-shot act or scene SHALL remove the linked snapshot through the registered unlink route.

#### Scenario: Unlink a compendium monster

- **WHEN** the DM activates unlink on a compendium monster in the one-shot monsters section
- **THEN** the request succeeds and the monster no longer appears in the section

### Requirement: Compendium equipment import into one-shot items

Importing compendium equipment into a one-shot SHALL create a one-shot item from the compendium entry (or equipment record) with the requested quantity and optional act, and SHALL return the new item id without error.

#### Scenario: Import equipment at adventure level

- **WHEN** the DM imports a compendium equipment entry into an adventure
- **THEN** a one-shot item is created for the adventure with the equipment's name and the requested quantity

#### Scenario: Import equipment into an act

- **WHEN** the DM imports a compendium equipment entry with an act id
- **THEN** the created item is associated with that act

### Requirement: Search results open the selected one-shot

Selecting a one-shot adventure in global search SHALL open that adventure's detail view, not just the one-shot list.

#### Scenario: Open an adventure from search

- **WHEN** the user selects an adventure result in global search
- **THEN** the one-shot view opens with that adventure's detail loaded
