## ADDED Requirements

### Requirement: Shops moved from Campaign to One-Shot
The existing `shops` table SHALL gain an optional `oneshot_adventure_id` column. Shops SHALL be creatable under one-shot adventures instead of campaigns.

#### Scenario: Create shop in one-shot
- **WHEN** a DM sends POST to `/api/oneshot-adventures/:id/shops` with shop data
- **THEN** the shop SHALL be created with `oneshot_adventure_id` set

#### Scenario: List one-shot shops
- **WHEN** a DM sends GET to `/api/oneshot-adventures/:id/shops`
- **THEN** all shops for that adventure SHALL be returned

#### Scenario: Update one-shot shop
- **WHEN** a DM sends PUT to `/api/shops/:id` with updated fields
- **THEN** the shop SHALL be updated

#### Scenario: Delete one-shot shop
- **WHEN** a DM sends DELETE to `/api/shops/:id`
- **THEN** the shop SHALL be deleted

## MODIFIED Requirements

### Requirement: Shop creation under campaign (existing)
The existing campaign shop creation SHALL still work via `campaign_id`, but the UI SHALL prefer the one-shot shop interface for DM users.

#### Scenario: Campaign shop still functional
- **WHEN** a user sends POST to `/api/shops` with `campaign_id` set
- **THEN** the shop SHALL still be created under the campaign

## REMOVED Requirements

### Requirement: Campaign shop creation (deprecated)
**Reason**: Shops are moving to one-shot adventures for DM-centric workflow
**Migration**: Existing campaign shops keep their campaign_id. New shops should use oneshot_adventure_id. UI will guide DM users to one-shot shop creation.
