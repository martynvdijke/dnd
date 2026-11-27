# campaign-character-roster Specification

## Purpose
TBD - created by archiving change campaign-management-roster. Update Purpose after archive.
## Requirements
### Requirement: Members can attach characters to a campaign

Any member of a campaign SHALL be able to attach a character to the campaign roster by calling `POST /api/campaigns/:id/characters` with a `character_id`. The caller SHALL be a member of the campaign (owner, or present in `campaign_members`); otherwise the request SHALL be rejected with `403`. The target character SHALL be owned by the caller or by another member of the same campaign; otherwise the request SHALL be rejected. On success a membership SHALL be created for this campaign without altering the character's memberships in other campaigns, and the character SHALL appear in `GET /api/campaigns/:id/characters`. Attaching a character that is already in this campaign's roster SHALL be an idempotent success.

#### Scenario: Member adds their own character

- **WHEN** a campaign member calls `POST /api/campaigns/:id/characters` with the `character_id` of a character they own
- **THEN** a membership for this campaign is created and the endpoint returns success

#### Scenario: Member adds an external character owned by another member

- **WHEN** a campaign member calls `POST /api/campaigns/:id/characters` with the `character_id` of a character owned by a different member of the same campaign
- **THEN** the character is attached to the campaign roster even though the caller does not own it

#### Scenario: Non-member is rejected

- **WHEN** a user who is not a member of the campaign calls `POST /api/campaigns/:id/characters`
- **THEN** the endpoint returns `403` and no character is attached

#### Scenario: Character owned by a stranger is rejected

- **WHEN** a campaign member tries to attach a character owned by a user who is not a member of the campaign
- **THEN** the endpoint returns an error and the character is not attached

#### Scenario: Character already in another campaign can also join this one

- **WHEN** a member attaches a character that already belongs to a different campaign
- **THEN** a membership for this campaign is added and the character's other memberships are unchanged

#### Scenario: Duplicate attach is idempotent

- **WHEN** a member attaches a character already in this campaign's roster
- **THEN** the endpoint succeeds and creates no duplicate membership

### Requirement: Members can detach characters from a campaign

Any member of a campaign SHALL be able to detach a character from the roster by calling `DELETE /api/campaigns/:id/characters/:characterId`. Detaching SHALL remove only this campaign's membership; the character's memberships in other campaigns SHALL be unaffected. If the character is not in this campaign's roster, the operation SHALL be an idempotent no-op success. The detached character SHALL no longer appear in `GET /api/campaigns/:id/characters` and, when it has no remaining memberships, SHALL surface under the "Unassigned" grouping in the selection flow.

#### Scenario: Member removes a roster character

- **WHEN** a campaign member calls `DELETE /api/campaigns/:id/characters/:characterId` for a character currently in this campaign's roster
- **THEN** the membership is removed and the character no longer appears in the campaign's character list

#### Scenario: Detaching does not affect other campaigns

- **WHEN** a member detaches a character that also belongs to another campaign
- **THEN** the other campaign's membership and roster are unchanged

#### Scenario: Non-member removal is rejected

- **WHEN** a user who is not a member of the campaign calls `DELETE /api/campaigns/:id/characters/:characterId`
- **THEN** the endpoint returns `403` and the character's memberships are unchanged

#### Scenario: Detaching a character not in the roster is a no-op

- **WHEN** a member calls `DELETE /api/campaigns/:id/characters/:characterId` for a character with no membership in this campaign
- **THEN** the endpoint returns success and no other membership is altered

### Requirement: Roster candidates are discoverable

`GET /api/campaigns/:id/character-candidates` SHALL return the characters a member may add to the roster: characters owned by the caller and characters owned by other members of the campaign, including characters that already belong to other campaigns. Each candidate SHALL be annotated with `owned` (true when owned by the caller), `owner_username`, `character_type`, and `in_roster` (true when already attached to this campaign). Non-members SHALL receive `403`.

#### Scenario: Candidates include own and external characters

- **WHEN** a campaign member requests the candidate list
- **THEN** the response contains their own characters (marked `owned: true`) and characters owned by other campaign members (marked `owned: false` with the owner's username)

#### Scenario: Characters in another campaign are included

- **WHEN** a candidate belongs to a different campaign
- **THEN** that character is present in the candidate list with `in_roster: false` for this campaign

#### Scenario: Attached characters are flagged

- **WHEN** a character in the candidate list is already attached to this campaign
- **THEN** it is returned with `in_roster: true` so the UI can pre-check it
