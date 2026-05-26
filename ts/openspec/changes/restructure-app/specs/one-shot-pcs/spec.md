## ADDED Requirements

### Requirement: Link player characters to one-shots
A DM SHALL be able to link existing player characters (from any user) to a one-shot adventure. A `one_shot_pcs` table SHALL store links: `id`, `adventure_id`, `character_id`, `notes`, `created_at`.

#### Scenario: Link PC to one-shot
- **WHEN** a DM sends POST to `/api/oneshot-adventures/:id/pcs` with `character_id`
- **THEN** the character SHALL be linked to the adventure

#### Scenario: List linked PCs
- **WHEN** a DM sends GET to `/api/oneshot-adventures/:id/pcs`
- **THEN** all linked characters with their full data SHALL be returned

#### Scenario: Unlink PC from one-shot
- **WHEN** a DM sends DELETE to `/api/oneshot-adventures/:id/pcs/:cid`
- **THEN** the character SHALL be unlinked from the adventure
