# trmnl-eink-plugin Specification

## Purpose
TBD - created by archiving change add-trmnl-stats-display. Update Purpose after archive.
## Requirements
### Requirement: TRMNL Liquid templates
The system SHALL ship TRMNL plugin templates covering all four TRMNL layout sizes, modeled on the sandwitches `trmnl/` plugin.

#### Scenario: All layout sizes provided
- **WHEN** the TRMNL plugin is installed
- **THEN** `full.liquid`, `half_horizontal.liquid`, `half_vertical.liquid`, and `quadrant.liquid` SHALL be provided
- **AND** each template SHALL reference polled data via `IDX_0` / `IDX_1` indices matching the configured polling URLs
- **AND** each template SHALL render a `title_bar` with the DnD icon and app title

#### Scenario: Character stats layout
- **WHEN** the display mode is set to character stats
- **THEN** the template SHALL render the character name, level, race, class, and HP as the primary values
- **AND** the six ability scores with +/− modifiers SHALL be rendered as large, high-contrast values
- **AND** AC, initiative, and XP SHALL be shown as supporting stats

#### Scenario: Campaign progress layout
- **WHEN** the display mode is set to campaign progress
- **THEN** the template SHALL render session count, total XP, and total gold as the primary values
- **AND** quest and rest breakdowns SHALL be summarized in text form
- **AND** top NPCs and dice roll summary SHALL be listed

#### Scenario: E-ink readable styling
- **WHEN** any TRMNL template renders
- **THEN** text SHALL be high-contrast black-on-white with no gradients, shadows, or color-dependence
- **AND** values SHALL use large `value--large` / `value--medium` classes and labels for readability from across the table

### Requirement: TRMNL plugin settings
The system SHALL ship a `trmnl/settings.yml` configured for polling the DnD TRMNL endpoints, with custom fields for display mode and character selection.

#### Scenario: Polling configuration
- **WHEN** the plugin settings are read by TRMNL
- **THEN** `polling_url` SHALL list the character-stats and campaign-stats endpoint URLs in the order matching `IDX_0` / `IDX_1` template references
- **AND** `refresh_interval` SHALL be a sensible default (e.g., 1440 minutes)
- **AND** `polling_verb` SHALL be `get`

#### Scenario: Display mode custom field
- **WHEN** a user configures the plugin
- **THEN** a `display_mode` custom field SHALL let them choose between character stats and campaign progress
- **AND** the field SHALL be referenced in templates via `trmnl.plugin_settings.custom_fields_values.display_mode`

#### Scenario: Character selector custom field
- **WHEN** a user configures the plugin
- **THEN** a `character_id` custom field SHALL let them select which character to display
- **AND** the token custom field SHALL let them enter the TRMNL access token
- **AND** the instance URL custom field SHALL let them override the polling base URL

#### Scenario: Instance URL placeholder
- **WHEN** the instance URL custom field is left empty
- **THEN** templates SHALL fall back to a sensible default base URL for asset references (icon/title bar)
