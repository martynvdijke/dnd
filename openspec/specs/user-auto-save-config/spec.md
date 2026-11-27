# user-auto-save-config Specification

## Purpose
User-facing setting to configure auto-save interval (persisted in user preferences).

## Requirements

### Requirement: Configurable auto-save interval

The system SHALL allow users to configure the auto-save interval. The interval SHALL be persisted in `localStorage` under the key `villum-autosave-interval`. The default value SHALL be 12 seconds. Valid values SHALL be integers between 5 and 300 seconds. The interval SHALL be readable and writable at runtime without a page reload.

#### Scenario: Default auto-save interval

- **WHEN** no `villum-autosave-interval` value exists in localStorage
- **THEN** the auto-save SHALL default to 12 seconds

#### Scenario: Custom auto-save interval from localStorage

- **WHEN** `villum-autosave-interval` is set to `30` in localStorage
- **THEN** the auto-save SHALL fire every 30 seconds

#### Scenario: Invalid interval falls back to default

- **WHEN** `villum-autosave-interval` is set to a non-numeric value or a value outside 5-300
- **THEN** the system SHALL fall back to the default 12-second interval

### Requirement: Settings UI for auto-save configuration

The system SHALL expose a settings UI (accessible from a settings/preferences dialog) where users can view and change their auto-save interval. The UI SHALL include a range slider or number input with the label "Auto-save interval (seconds)" and display the current value. Changes SHALL take effect immediately for the current session and be persisted to localStorage.

#### Scenario: Auto-save interval setting displayed

- **WHEN** the user opens the settings/preferences dialog
- **THEN** they SHALL see an "Auto-save interval" control showing the current value in seconds

#### Scenario: Changing auto-save interval takes effect immediately

- **WHEN** the user changes the auto-save interval to 20 seconds
- **THEN** the auto-save timer SHALL immediately use the new 20-second interval without requiring a page reload

#### Scenario: Auto-save interval persisted

- **WHEN** the user changes the auto-save interval and reloads the page
- **THEN** the system SHALL read `villum-autosave-interval` from localStorage and use the configured value
