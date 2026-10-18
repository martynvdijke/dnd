# telegram-auto-post-scheduling Specification

## Purpose
TBD - created by archiving change add-telegram-digest-and-cadence. Update Purpose after archive.
## Requirements
### Requirement: Automatic post cadence

Campaign Telegram settings SHALL accept `auto_post_cadence_minutes` as an integer from 0 to 1440, where 0 means no pacing and preserves the existing behaviour. When the value is greater than 0, the scheduler SHALL skip automatic recap posting for that campaign until at least that many minutes have elapsed since the campaign's last automatic post time, which is stored in UTC. Manual and command-triggered deliveries and retries SHALL NOT be paced.

#### Scenario: Paced campaign is skipped before the window elapses

- **WHEN** a campaign has a cadence of 60 minutes and its last automatic post was less than 60 minutes ago
- **THEN** the scheduler posts no automatic recap for that campaign on this tick

#### Scenario: Paced campaign posts after the window elapses

- **WHEN** a campaign has a cadence of 60 minutes and its last automatic post was at least 60 minutes ago
- **THEN** the scheduler posts the pending automatic recap and records the new automatic post time

#### Scenario: Zero cadence keeps the existing unpaced behaviour

- **WHEN** a campaign has a cadence of 0
- **THEN** the scheduler posts every pending automatic recap without a pacing delay, and the automatic post time is still recorded

#### Scenario: Manual delivery is not paced

- **WHEN** a DM triggers a recap delivery by command or the retry path runs
- **THEN** the delivery is sent even if the cadence window has not elapsed

#### Scenario: Out-of-range cadence is rejected

- **WHEN** a request sets `auto_post_cadence_minutes` below 0 or above 1440
- **THEN** the request is rejected with 400 and the stored value is unchanged

### Requirement: Quiet hours

Campaign Telegram settings SHALL accept `quiet_start_hour` and `quiet_end_hour` as optional integers from 0 to 23 that must be both set or both cleared. While the current UTC hour is inside the quiet window, automatic recap posting SHALL be suppressed. All quiet-hour evaluation SHALL use UTC; no local time zone or daylight-saving conversion SHALL be applied.

#### Scenario: Automatic post inside the quiet window is suppressed

- **WHEN** a campaign has quiet hours from 22 to 7 and the current UTC hour is 23
- **THEN** the scheduler posts no automatic recap

#### Scenario: Automatic post outside the quiet window proceeds

- **WHEN** a campaign has quiet hours from 22 to 7 and the current UTC hour is 12
- **THEN** the scheduler posts the pending automatic recap

#### Scenario: Boundary hours are respected

- **WHEN** a campaign has quiet hours from 22 to 7
- **THEN** the start hour 22 is inside the quiet window and is suppressed, while the end hour 7 is outside it and posting proceeds

#### Scenario: Unset quiet hours suppress nothing

- **WHEN** a campaign has no quiet start or end hour set
- **THEN** automatic recap posting is never suppressed by quiet hours

#### Scenario: Half-set quiet hours are rejected

- **WHEN** a request sets only one of `quiet_start_hour` or `quiet_end_hour`
- **THEN** the request is rejected with 400 and the stored values are unchanged

### Requirement: Campaign Telegram scheduling settings API

`GET /api/campaigns/:id/telegram` SHALL return `auto_post_cadence_minutes`, `quiet_start_hour`, `quiet_end_hour`, and `auto_post_last_at` together with the existing fields. `PUT /api/campaigns/:id/telegram` SHALL accept the first three, where a null quiet hour clears both quiet hours. The endpoint SHALL remain DM-only, and the existing `auto_save_media` field SHALL keep working unchanged.

#### Scenario: Settings read returns the scheduling fields

- **WHEN** a DM reads a campaign's Telegram settings
- **THEN** the response includes the cadence, both quiet hours, and the last automatic post time alongside the existing fields

#### Scenario: Settings write updates scheduling fields

- **WHEN** a DM puts a cadence and a quiet-hour pair
- **THEN** the values are persisted and returned on the next read

#### Scenario: Null quiet hours clear the window

- **WHEN** a DM puts null quiet hour values after a window was set
- **THEN** both quiet hours are cleared and no quiet suppression applies

#### Scenario: Non-DM access is refused

- **WHEN** a non-DM member requests the campaign's Telegram settings
- **THEN** the request is refused

#### Scenario: Auto-save keeps working

- **WHEN** a DM updates the campaign's Telegram settings as before
- **THEN** `auto_save_media` round-trips with unchanged behaviour
