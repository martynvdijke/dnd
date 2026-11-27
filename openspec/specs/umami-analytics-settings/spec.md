# umami-analytics-settings Specification

## Purpose
Admin configuration for Umami self-hosted analytics — enables/disables tracking, sets tracker URL and website ID, controls admin page tracking and data sharing.

## Requirements

### Requirement: Admin can configure Umami analytics settings
The admin panel SHALL provide an "Analytics" tab with a form to configure Umami self-hosted analytics. Settings MUST be persisted to the database and take effect on the next page load.

#### Scenario: Admin sees analytics tab
- **WHEN** an admin user navigates to `/admin`
- **THEN** an "Analytics" tab SHALL be visible in the admin navigation
- **AND** clicking the tab SHALL display the analytics settings form

#### Scenario: Default settings on first visit
- **WHEN** no Umami settings have been saved yet
- **THEN** the form SHALL show default values: enabled=false, tracker_hostname="", website_id="", share_data=false, enable_admin_tracking=false

#### Scenario: Admin saves valid Umami settings
- **WHEN** an admin fills in the form with a tracker_hostname, website_id, and sets enabled=true
- **AND** clicks "Save Settings"
- **THEN** the settings SHALL be persisted to the database
- **AND** a success toast SHALL be displayed

#### Scenario: Admin saves with empty tracker_hostname
- **WHEN** an admin tries to save with enabled=true but tracker_hostname is empty
- **THEN** the request SHALL still succeed (empty fields are allowed — they just won't produce a script tag)
- **AND** the tracking script SHALL NOT be injected until both hostname and website_id are set

#### Scenario: Admin loads saved settings
- **WHEN** an admin navigates to the Analytics tab after previously saving settings
- **THEN** the form SHALL be pre-populated with the saved values
- **AND** the password/token field SHALL show a placeholder indicating a token is set (if one was previously saved)

### Requirement: Tracking script injected into frontend pages
The server SHALL conditionally inject the Umami tracking script into `app.html`, `login.html`, and `admin.html` (subject to `enable_admin_tracking`) based on saved settings.

#### Scenario: Tracking enabled and configured
- **WHEN** Umami is enabled, tracker_hostname and website_id are non-empty
- **AND** a user visits the main app page (`/`)
- **THEN** the HTML `<head>` SHALL include `<script async defer src="{tracker_hostname}/script.js" data-website-id="{website_id}"></script>`

#### Scenario: Tracking disabled
- **WHEN** Umami is disabled in settings
- **THEN** no tracking script SHALL be injected into any page

#### Scenario: Admin tracking disabled
- **WHEN** Umami is enabled but `enable_admin_tracking` is false
- **AND** a user visits `/admin`
- **THEN** the tracking script SHALL NOT be injected into the admin page
- **AND** the main app page AND login page SHALL still receive the script

#### Scenario: Admin tracking enabled
- **WHEN** Umami is enabled and `enable_admin_tracking` is true
- **AND** a user visits `/admin`
- **THEN** the tracking script SHALL be injected into the admin page

### Requirement: Umami tracking respects data sharing preference
The tracking script injection SHALL include the `data-share-data` attribute when the share_data setting is enabled.

#### Scenario: Data sharing enabled
- **WHEN** `share_data` is true in settings
- **THEN** the injected script tag SHALL include `data-share-data="true"`

#### Scenario: Data sharing disabled
- **WHEN** `share_data` is false in settings
- **THEN** the injected script tag SHALL NOT include the `data-share-data` attribute

### Requirement: API endpoints follow existing conventions
The system SHALL expose `GET /api/admin/umami-settings` and `POST /api/admin/umami-settings` endpoints, following the same pattern as email settings.

#### Scenario: GET returns current settings
- **WHEN** `GET /api/admin/umami-settings` is called by an authenticated admin
- **THEN** the response SHALL contain `{ enabled, tracker_hostname, website_id, share_data, enable_admin_tracking }`
- **AND** if no settings exist in the database, defaults SHALL be returned

#### Scenario: POST saves settings
- **WHEN** `POST /api/admin/umami-settings` is called with valid JSON body
- **THEN** the settings SHALL be upserted (INSERT OR REPLACE)
- **AND** the response SHALL be `{"status": "ok"}`
