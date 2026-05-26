## ADDED Requirements

### Requirement: Inline monsters per act/scene
The system SHALL support monsters defined inline within a one-shot act or scene. Inline monsters have: name, ac, hp, str/dex/con/int/wis/cha, cr, source, and an is_full toggle for full-stat mode.

#### Scenario: Add inline monster to act
- **WHEN** a DM makes a POST to `/oneshot-acts/:id/monsters` with monster name, ac, hp, cr
- **THEN** the system creates the monster linked to that act

#### Scenario: Add inline monster to scene
- **WHEN** a DM makes a POST to `/oneshot-scenes/:id/monsters` with monster details
- **THEN** the system creates the monster linked to that scene

#### Scenario: Update inline monster
- **WHEN** a DM makes a PUT to `/oneshot-monsters/:id` with updated stats
- **THEN** the system updates the monster

#### Scenario: Remove inline monster
- **WHEN** a DM makes a DELETE to `/oneshot-monsters/:id`
- **THEN** the system removes the monster

### Requirement: Monster library
The system SHALL support a monster library where DMs can save reusable monster stat blocks. Library monsters have the same stats as inline monsters plus description and created_at.

#### Scenario: Create library monster
- **WHEN** a DM makes a POST to `/monster-library` with monster name, ac, hp, str, dex, con, int, wis, cha, cr
- **THEN** the system creates a library monster entry

#### Scenario: Quick-add library monster to act
- **WHEN** a DM makes a POST to `/oneshot-acts/:id/monsters` with library_monster_id=<id>
- **THEN** the system copies the library monster stats into an inline monster linked to that act

#### Scenario: List library monsters
- **WHEN** a DM makes a GET to `/monster-library`
- **THEN** the system returns all library monsters for that user

### Requirement: Full-stat monster mode
Monsters (both inline and library) SHALL support a full-stat mode toggle. When is_full is true, additional fields appear: saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities, condition_immunities, senses, languages, challenge_rating_xp, special_abilities, actions, legendary_actions.

#### Scenario: Create full-stat monster
- **WHEN** a DM creates a monster with is_full=true and provides full stat block
- **THEN** the system stores all full-stat fields

#### Scenario: Render full-stat monster card
- **WHEN** a DM views a full-stat monster in the one-shot UI
- **THEN** the system displays the complete stat block card
