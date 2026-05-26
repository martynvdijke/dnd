## ADDED Requirements

### Requirement: Shops associated with one-shots
The system SHALL support shops that belong to a one-shot adventure. Shops gain a nullable `oneshot_adventure_id` foreign key alongside the existing `campaign_id`. Shop CRUD SHALL be gated behind DMRequired middleware.

#### Scenario: Create shop for one-shot
- **WHEN** a DM makes a POST to `/oneshot-adventures/:id/shops` with shop details
- **THEN** the system creates the shop linked to the one-shot and returns 201

#### Scenario: List shops for one-shot
- **WHEN** a DM makes a GET to `/oneshot-adventures/:id/shops`
- **THEN** the system returns shops belonging to that one-shot

#### Scenario: Add item to one-shot shop
- **WHEN** a DM makes a POST to `/oneshot-adventures/:id/shops/:shopId/items`
- **THEN** the system creates a shop item and returns 201

#### Scenario: Remove shop from one-shot
- **WHEN** a DM makes a DELETE to `/oneshot-adventures/:id/shops/:shopId`
- **THEN** the system deletes the shop and returns 200
