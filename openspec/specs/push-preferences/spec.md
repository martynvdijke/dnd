# push-preferences Specification

## Purpose
TBD - created by archiving change add-web-push-notifications. Update Purpose after archive.
## Requirements
### Requirement: Per-campaign mute preference
A user SHALL be able to mute and unmute push notifications per campaign from the campaign member UI. Muted campaigns SHALL receive no pushes of any type for that user.

#### Scenario: User mutes a campaign
- **WHEN** the user toggles mute on for a campaign
- **THEN** subsequent recap and reminder fan-outs skip that user

#### Scenario: User unmutes a campaign
- **WHEN** the user toggles mute off
- **THEN** future notifications reach them again without re-subscribing

#### Scenario: Mute state persists across sessions
- **WHEN** a user mutes a campaign and logs in later from another device
- **THEN** the mute preference still applies
