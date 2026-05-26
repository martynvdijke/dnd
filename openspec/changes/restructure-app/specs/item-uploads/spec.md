## ADDED Requirements

### Requirement: Upload files attached to items
The system SHALL support file uploads attached to one-shot items via polymorphic owner_type/owner_id. The uploads table gains `owner_type` and `owner_id` columns.

#### Scenario: Upload file to item
- **WHEN** a DM uploads a file with owner_type="item" and owner_id=<item_id>
- **THEN** the system stores the file and creates an upload record linked to that item

#### Scenario: List uploads for item
- **WHEN** a DM makes a GET to `/oneshot-items/:id/uploads`
- **THEN** the system returns uploads with owner_type="item" and owner_id=<item_id>

#### Scenario: Delete item upload
- **WHEN** a DM makes a DELETE to `/uploads/:id`
- **THEN** the system deletes the upload record and its file
