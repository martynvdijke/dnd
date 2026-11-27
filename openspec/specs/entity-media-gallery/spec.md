# entity-media-gallery Specification

## Purpose
A reusable HTMX-based media gallery component for viewing, uploading, linking, and managing files across all entity types.

## Requirements

### Requirement: Reusable media gallery component

The system SHALL provide a single HTMX-rendered media gallery template that can be used for campaigns, one-shots, characters, and NPCs.

#### Scenario: Gallery shows linked uploads
- **WHEN** user navigates to a campaign/one-shot's media tab
- **THEN** all uploads linked to that entity are shown in a responsive image grid with thumbnails

#### Scenario: Upload file from gallery
- **WHEN** user drags-and-drops or clicks to browse in the gallery upload area
- **THEN** the file is uploaded and linked to the entity, appearing in the gallery immediately

#### Scenario: Gallery shows PDF with icon
- **WHEN** user uploads a PDF file
- **THEN** the gallery shows a PDF file icon instead of an image thumbnail, with the filename visible

#### Scenario: Campaign media tab
- **WHEN** user is viewing a campaign
- **THEN** there is a "Media" tab showing the media gallery for that campaign

#### Scenario: One-Shot media section
- **WHEN** user is viewing a one-shot adventure detail
- **THEN** there is a "Media" section showing the media gallery for that one-shot

### Requirement: File upload constraints

The upload system SHALL support images (jpg, png, gif, webp, avif, svg, bmp, tiff) and PDF files, with a maximum file size of 10MB.

#### Scenario: Upload valid image
- **WHEN** user uploads a valid image file under 10MB
- **THEN** the file is accepted, processed (resized, thumbnailed), and stored

#### Scenario: Upload valid PDF
- **WHEN** user uploads a PDF file under 10MB
- **THEN** the file is accepted and stored as-is (no image processing)

#### Scenario: Upload exceeds size limit
- **WHEN** user uploads a file larger than 10MB
- **THEN** the system returns a 413 Payload Too Large or 400 Bad Request error

#### Scenario: Upload unsupported file type
- **WHEN** user uploads a file with an unsupported extension (.exe, .zip, .docx)
- **THEN** the system returns a 400 Bad Request error

### Requirement: Character portrait upload enhancement

The existing character portrait upload SHALL set `owner_type=character` and `owner_id=<character_id>` to properly track the upload.

#### Scenario: Character portrait upload tracks owner
- **WHEN** user uploads a portrait from the character sheet
- **THEN** the upload record includes owner_type=character and the character's ID
