# file-picker Specification

## Purpose
A reusable frontend component for browsing, uploading, and selecting images from the existing upload system.

## Requirements

### Requirement: File picker component

The system SHALL provide a reusable file picker component that allows users to browse previously uploaded images and select one, or upload a new image, returning the selected file's URL.

#### Scenario: Open file picker
- **WHEN** user clicks a "Browse Images" or "Choose Image" button
- **THEN** a modal overlay opens showing a grid of previously uploaded images

#### Scenario: Select existing image
- **WHEN** user clicks on an image in the file picker grid
- **THEN** the image is visually selected with a highlight border and a "Select" button becomes active

#### Scenario: Confirm selection
- **WHEN** user clicks "Select" after choosing an image
- **THEN** the modal closes and the selected image URL is returned via a callback/promise

#### Scenario: Upload new image from picker
- **WHEN** user drags-and-drops a file onto the picker or clicks an upload area
- **THEN** the file is uploaded via the existing `/api/upload` endpoint and appears in the grid immediately

#### Scenario: Close without selecting
- **WHEN** user clicks the close button or clicks outside the modal
- **THEN** the picker closes without returning a selection

#### Scenario: Empty gallery state
- **WHEN** no images have been uploaded yet
- **THEN** the picker shows an empty state message with a prominent upload area

#### Scenario: Keyboard navigation
- **WHEN** user presses Escape while the picker is open
- **THEN** the picker closes without selecting

### Requirement: File picker integration with character portrait

The file picker SHALL be usable from the character portrait upload flow in the SPA character sheet.

#### Scenario: Pick portrait from character sheet
- **WHEN** user clicks the portrait area on the character sheet
- **THEN** the file picker opens filtered to the user's uploads
- **WHEN** user selects an image
- **THEN** the character's portrait_url is updated and the portrait displays the new image

### Requirement: File picker integration with NPC portrait

The file picker SHALL be usable from the NPC form for setting an NPC portrait.

#### Scenario: Pick NPC portrait from form
- **WHEN** user clicks "Choose Portrait" on the NPC create/edit form
- **THEN** the file picker opens filtered to the user's uploads
- **WHEN** user selects an image
- **THEN** the NPC portrait_url field is populated and a preview thumbnail is shown
