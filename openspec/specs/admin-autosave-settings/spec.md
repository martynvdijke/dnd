# admin-autosave-settings Specification

## Purpose
TBD - created by archiving change spell-picker-filters-admin-autosave. Update Purpose after archive.
## Requirements
### Requirement: Site-wide auto-save interval
The system MUST provide an administrator-controlled auto-save interval stored in the generic application settings store and shared by all users.

#### Scenario: Admin reads the setting
- **WHEN** an administrator requests the auto-save setting
- **THEN** the system returns the normalized site-wide interval

#### Scenario: Admin updates the setting
- **WHEN** an administrator saves an interval between 5 and 300 seconds
- **THEN** the system persists it and returns the saved value

#### Scenario: Invalid interval is normalized
- **WHEN** an administrator submits an interval below 5, above 300, malformed, or absent
- **THEN** the system clamps it when numeric or uses the 12-second default when invalid/missing

### Requirement: Clients use the admin interval
Character sheets MUST use the site-wide server interval for auto-save and MUST NOT expose a separate user-facing auto-save settings control.

#### Scenario: Character sheet loads the interval
- **WHEN** a user opens a character sheet
- **THEN** its auto-save scheduler uses the interval returned by the application settings endpoint

#### Scenario: No setting exists
- **WHEN** the server has no auto-save interval configured
- **THEN** the client uses 12 seconds

#### Scenario: Dedicated control is absent
- **WHEN** a user views the character sheet controls
- **THEN** no dedicated auto-save interval button or modal is displayed
