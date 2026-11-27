# character-campaign-membership Specification

## Purpose
TBD - created by archiving change character-multi-campaign. Update Purpose after archive.
## Requirements
### Requirement: Membership is many-to-many

A character SHALL be able to belong to zero or more campaigns, and a campaign SHALL be able to contain zero or more characters. Membership SHALL be stored in a `campaign_characters` join table and SHALL be the single source of truth; `characters.campaign_id` SHALL no longer exist. Character API payloads SHALL expose `campaigns` as a list of `{id, name}` entries for every campaign the character belongs to.

#### Scenario: Character belongs to two campaigns

- **WHEN** a character is attached to campaigns A and B
- **THEN** its payload lists both campaigns and it appears in both campaigns' rosters

#### Scenario: Character belongs to no campaign

- **WHEN** a character has no memberships
- **THEN** its `campaigns` list is empty and the selection flow surfaces it under the "Unassigned" grouping

#### Scenario: Membership is unique per pair

- **WHEN** the same character is attached to the same campaign twice
- **THEN** only one membership row exists and the second request succeeds without error

#### Scenario: Backfill preserves existing assignments

- **WHEN** a database upgrade runs against a character with `campaign_id` set
- **THEN** a membership row exists for that campaign after migration and the character remains on its roster

### Requirement: Membership editing from the character side

A character owner (or an admin) SHALL be able to set a character's memberships by submitting `campaign_ids` when creating or updating the character. The caller SHALL be a member (owner counts as member) of every campaign added; otherwise the request SHALL be rejected with `403`. Removing a membership in a campaign the caller is not a member of SHALL be rejected with `403` unless the caller is an admin. Shared, read-only characters SHALL NOT expose membership editing.

#### Scenario: Owner assigns a new character to campaigns

- **WHEN** an owner creates a character with `campaign_ids` listing campaigns they belong to
- **THEN** the character is created and a membership exists for each listed campaign

#### Scenario: Owner updates memberships

- **WHEN** an owner updates a character with a different `campaign_ids` set
- **THEN** memberships are added and removed to match the submitted set

#### Scenario: Adding to a campaign the caller cannot access

- **WHEN** an owner submits a `campaign_ids` entry for a campaign they are not a member of
- **THEN** the request is rejected with `403` and memberships are unchanged

#### Scenario: Removing a membership the caller cannot manage

- **WHEN** a non-admin owner submits `campaign_ids` that omits a membership in a campaign they are not a member of
- **THEN** the request is rejected with `403` and that membership is preserved

#### Scenario: Shared character membership editing is blocked

- **WHEN** a user who does not own the character attempts to change its memberships
- **THEN** the request is rejected and the UI does not offer membership controls

### Requirement: Campaign-scoped reads follow memberships

Every campaign-scoped read of character data (party grouping, campaign dashboard, DM and ownership checks, timelines, recaps, copilot context, email highlights, public shares, table screen, one-shot pacing, and the campaign graph) SHALL derive membership from the `campaign_characters` join table rather than a character column.

#### Scenario: Party view groups a multi-campaign character per campaign

- **WHEN** a character belongs to two campaigns
- **THEN** it appears in the party view group of each campaign

#### Scenario: Dashboard scoping follows memberships

- **WHEN** a campaign dashboard lists characters, conditions, downtime, or journals
- **THEN** only data belonging to characters that are members of that campaign is included

#### Scenario: DM checks follow memberships

- **WHEN** a user is a DM of campaign A and a character belongs to campaigns A and B
- **THEN** the user's DM permissions over that character derive only from campaign A's membership

### Requirement: Deleting a campaign removes only memberships

Deleting a campaign SHALL delete its memberships and SHALL NOT delete or alter characters or their memberships in other campaigns.

#### Scenario: Delete a campaign with shared characters

- **WHEN** a campaign containing a character that also belongs to another campaign is deleted
- **THEN** the character remains, its membership in the deleted campaign is removed, and its other membership is unchanged
