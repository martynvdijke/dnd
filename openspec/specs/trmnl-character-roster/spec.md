# trmnl-character-roster Specification

## Purpose
TBD - created by archiving change add-trmnl-character-roster. Update Purpose after archive.
## Requirements
### Requirement: Character roster polling endpoint
The system SHALL expose a public read-only JSON endpoint returning the full party roster — every character with their core stats — for TRMNL polling.

#### Scenario: Request roster
- **WHEN** a client polls `/api/trmnl/characters` with a valid token
- **THEN** the response SHALL be a JSON array of all characters ordered by name
- **AND** each entry SHALL include `id`, `name`, `race`, `class`, `subclass`, `level`, `xp`, `hp_current`, `hp_max`, `ac`, `initiative`, the six ability scores (`str`, `dex`, `con`, `int`, `wis`, `cha`) with their modifiers, and `character_type`
- **AND** the response SHALL be valid JSON suitable for Liquid templating

#### Scenario: Token required
- **WHEN** a request is made to `/api/trmnl/characters` without a valid `token` query parameter
- **THEN** the endpoint SHALL return `401 Unauthorized`
- **AND** no character data SHALL be returned

#### Scenario: Read-only
- **WHEN** a client polls `/api/trmnl/characters`
- **THEN** no records SHALL be created, updated, or deleted
- **AND** the endpoint SHALL only support `GET`

### Requirement: Roster display mode in Liquid templates
The TRMNL plugin SHALL render the party roster on e-ink displays when `display_mode` is set to `roster`, without breaking the existing character and campaign modes.

#### Scenario: Roster layout
- **WHEN** `display_mode` is `roster`
- **THEN** the template SHALL iterate the polled roster (`IDX_2`) and render each character as a row
- **AND** each row SHALL show the character name, race/class/subclass, level, and HP in high-contrast e-ink styling
- **AND** the `title_bar` SHALL still render

#### Scenario: Backward compatibility
- **WHEN** `display_mode` is `character` or `campaign`
- **THEN** the existing `IDX_0` / `IDX_1` template blocks SHALL render unchanged

#### Scenario: All layout sizes
- **WHEN** any of `full.liquid`, `half_horizontal.liquid`, `half_vertical.liquid`, or `quadrant.liquid` renders in roster mode
- **THEN** the roster block SHALL be present in all four templates

### Requirement: TRMNL plugin settings for roster
The system SHALL ship a `settings.yml` that configures TRMNL to poll the roster endpoint, with a roster display mode option.

#### Scenario: Polling configuration
- **WHEN** the plugin settings are read by TRMNL
- **THEN** `polling_url` SHALL include the character-stats, campaign-stats, and roster endpoints in that order
- **AND** `refresh_interval` SHALL remain a sensible default (e.g., 1440 minutes)
- **AND** `polling_verb` SHALL be `get`

#### Scenario: Roster custom field option
- **WHEN** a user configures the plugin
- **THEN** the `display_mode` custom field SHALL offer `roster` alongside `character` and `campaign`
- **AND** the field SHALL be referenced in templates via `trmnl.plugin_settings.custom_fields_values.display_mode`
