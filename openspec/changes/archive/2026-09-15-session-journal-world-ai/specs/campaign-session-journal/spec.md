## ADDED Requirements

### Requirement: Session journal view

The system SHALL provide a campaign-scoped Session Journal view that lists the campaign's session write-ups and lets the DM open one in an editor.

#### Scenario: List session write-ups
- **WHEN** the DM opens the Session Journal for a campaign
- **THEN** the campaign's write-ups are listed with title and session date

#### Scenario: Empty state
- **WHEN** the campaign has no write-ups
- **THEN** the view shows an empty state with a clear action to create the first write-up

### Requirement: Rich-text session write-up

The system SHALL allow the DM to create and edit a session write-up with a rich-text editor supporting formatted text (headings, lists, emphasis, and similar), storing the rendered HTML as the write-up content. Each write-up SHALL have a title and a session date or date range.

#### Scenario: Create a write-up
- **WHEN** the DM fills in a title and content and saves
- **THEN** the write-up is persisted and appears in the list

#### Scenario: Formatting is preserved
- **WHEN** the DM saves content containing headings and lists and reopens it
- **THEN** the formatting is preserved

#### Scenario: Edit marks the write-up edited
- **WHEN** the DM changes and re-saves an existing write-up
- **THEN** the write-up is marked as edited

### Requirement: Session write-up lifecycle

The system SHALL allow the DM to update and delete a session write-up, and SHALL record a word count for each saved write-up.

#### Scenario: Update a write-up
- **WHEN** the DM saves changes to a write-up
- **THEN** the stored content and word count are updated

#### Scenario: Delete a write-up
- **WHEN** the DM deletes a write-up
- **THEN** it is removed from the list and no longer returned

### Requirement: Session write-up links

A session write-up SHALL be linkable to timeline events and places, and those links SHALL be visible on the write-up and on the linked entities.

#### Scenario: Link places and events
- **WHEN** the DM attaches a place and a timeline event to a write-up
- **THEN** the write-up shows both links and the place/event show a backlink to the write-up

#### Scenario: Remove a link
- **WHEN** the DM removes a link from a write-up
- **THEN** the link no longer appears on either side

### Requirement: Session journal access control

The campaign owner (DM) SHALL be able to create, edit, and delete write-ups. Non-owner campaign members SHALL be able to read write-ups that have been published to them, and SHALL NOT be able to modify them.

#### Scenario: Member reads published write-up
- **WHEN** a non-owner member opens the Session Journal
- **THEN** published write-ups are readable in a non-editable form

#### Scenario: Member cannot edit
- **WHEN** a non-owner member attempts to modify a write-up
- **THEN** the request is denied

#### Scenario: Non-member denied
- **WHEN** a user who is not a member of the campaign requests its write-ups
- **THEN** access is denied
