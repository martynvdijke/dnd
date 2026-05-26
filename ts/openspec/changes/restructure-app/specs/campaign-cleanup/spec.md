## REMOVED Requirements

### Requirement: Calendar events in campaigns
**Reason**: Calendar functionality removed entirely. No replacement.
**Migration**: Existing calendar events will be archived to a backup table. Campaign data unaffected otherwise.

### Requirement: Calendar CRUD routes
**Reason**: Calendar removed entirely.
**Migration**: Routes `/api/calendar/*` will be removed from the router. Existing calendar data backed up.

### Requirement: Campaign factions
**Reason**: Factions moved to Party entity.
**Migration**: All existing factions with `campaign_id` will be migrated to have `party_id` set. A migration script will create parties from existing campaign party_names where needed.

### Requirement: Campaign shops
**Reason**: Shops moved to One-Shot adventure.
**Migration**: Existing shops keep their `campaign_id`. New shops created via DM UI will use `oneshot_adventure_id`. UI will be updated to guide DM users to one-shot shop creation.

## MODIFIED Requirements

### Requirement: Campaign model cleanup
The Campaign entity SHALL lose its `party_name` field, `calendar_events` route, `factions` route, and `shops` route. Campaign keeps: `id`, `user_id`, `name`, `description`, `dm_notes`, `created_at`, and edges to members, wiki, encounters, maps, recaps, combat.

#### Scenario: Campaign creation without party_name
- **WHEN** a user creates a campaign without `party_name`
- **THEN** the campaign SHALL be created successfully (party_name is no longer required)

#### Scenario: Campaign GET no longer returns factions
- **WHEN** a user gets a campaign
- **THEN** the response SHALL NOT include faction data (factions live under Party)
