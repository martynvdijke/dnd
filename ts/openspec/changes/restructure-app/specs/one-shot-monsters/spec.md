## ADDED Requirements

### Requirement: Monster library
The system SHALL have a `monster_library` table where DM users can store reusable monster definitions. Columns: `id`, `user_id`, `name`, `size`, `type`, `alignment`, `ac`, `hp`, `speed`, `str`, `dex`, `con`, `int`, `wis`, `cha`, `cr`, `xp`, `abilities` (JSON), `actions` (JSON), `description`, `created_at`.

#### Scenario: Create monster in library
- **WHEN** a DM sends POST to `/api/monster-library` with monster data
- **THEN** the monster SHALL be created in the library

#### Scenario: List monster library
- **WHEN** a DM sends GET to `/api/monster-library`
- **THEN** all monsters in the library SHALL be returned

#### Scenario: Delete monster from library
- **WHEN** a DM sends DELETE to `/api/monster-library/:id`
- **THEN** the monster SHALL be removed

### Requirement: Inline monsters per act/scene
Acts and scenes SHALL support inline monster definitions (not from library). A `one_shot_scene_monsters` table SHALL link monsters to scenes with columns: `id`, `scene_id`, `monster_library_id` (optional), `name`, `ac`, `hp`, `initiative`, `quantity`, `notes` (inline stats when no library_id).

#### Scenario: Add library monster to scene
- **WHEN** a DM sends POST to `/api/oneshot-scenes/:id/monsters` with `monster_library_id` and quantity
- **THEN** a monster reference SHALL be added to the scene

#### Scenario: Add inline monster to scene
- **WHEN** a DM sends POST to `/api/oneshot-scenes/:id/monsters` with inline stats (name, ac, hp)
- **THEN** a monster SHALL be added to the scene without a library reference

#### Scenario: List monsters in scene
- **WHEN** a DM sends GET to `/api/oneshot-scenes/:id/monsters`
- **THEN** all monsters (library + inline) for the scene SHALL be returned

#### Scenario: Remove monster from scene
- **WHEN** a DM sends DELETE to `/api/oneshot-scene-monsters/:id`
- **THEN** the monster SHALL be removed from the scene
