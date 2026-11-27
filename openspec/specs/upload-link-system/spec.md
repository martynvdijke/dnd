# upload-link-system Specification

## Purpose
Many-to-many upload-to-entity linking via `upload_links` table with API endpoints for CRUD.

## Requirements

### Requirement: Many-to-many upload-to-entity linking

The system SHALL support linking an upload to multiple entities via an `upload_links` join table.

#### Scenario: Create upload link
- **WHEN** user POSTs to `/api/upload-links` with `{upload_id, entity_type, entity_id, field_name}`
- **THEN** a link is created in the upload_links table

#### Scenario: Delete upload link
- **WHEN** user DELETEs `/api/upload-links/:id`
- **THEN** the corresponding link row is removed

#### Scenario: List uploads for entity
- **WHEN** user GETs `/api/uploads?entity_type=X&entity_id=Y`
- **THEN** all uploads linked to that entity (via upload_links OR direct owner_type/owner_id) are returned

#### Scenario: Duplicate link prevention
- **WHEN** user tries to create a duplicate `(upload_id, entity_type, entity_id)` link
- **THEN** the duplicate SHALL be rejected (UNIQUE constraint)

### Requirement: Direct ownership fallback

The existing `owner_type`/`owner_id` fields on the uploads table SHALL continue to work for single-ownership scenarios.

#### Scenario: Upload with owner_type still works
- **WHEN** user uploads a file with `owner_type=npc&owner_id=5`
- **THEN** the upload is saved with those owner fields as before

#### Scenario: Upload with owner_type shown in link-based galleries
- **WHEN** user GETs `/api/uploads?entity_type=npc&entity_id=5`
- **THEN** both linked uploads (via upload_links) and directly-owned uploads (via owner_type/owner_id) are included

### Requirement: Copy URL from gallery

The media gallery SHALL provide a "Copy URL" button for each file.

#### Scenario: Copy URL button
- **WHEN** user clicks "Copy URL" on a file in the gallery
- **THEN** the file's full URL is copied to the clipboard and a brief "Copied!" notification appears

### Requirement: Delete upload from gallery

The media gallery SHALL provide a delete button for each file (for users with permission).

#### Scenario: Delete file from gallery
- **WHEN** user clicks delete on a file
- **THEN** the file is removed from the uploads table and its linked entities (via upload_links), and the media file is deleted from disk
