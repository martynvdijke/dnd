# telegram-character-commands

## ADDED Requirements

### Requirement: Administrator character visibility

An administrator SHALL be able to see, claim, and open the sheet and stats of
any character, mirroring the web app's administrator edit bypass
(`handlers.canEditCharacter`). Non-administrators SHALL remain restricted to
characters they own or may edit through campaign ownership or a campaign `dm`
role.

#### Scenario: Administrator lists every character
- **GIVEN** a linked user whose account has the administrative role
- **WHEN** the user sends `/characters`
- **THEN** the listing includes characters owned by other users

#### Scenario: Administrator opens another user's character
- **GIVEN** a linked administrator and a character owned by a different user
- **WHEN** the administrator sends `/sheet <id>` or `/stats <id>` for that character
- **THEN** the bot renders the sheet or stats for that character

#### Scenario: Non-administrator is unaffected
- **GIVEN** a linked non-administrator who neither owns a character nor may edit it
- **WHEN** the user sends `/characters`, `/sheet <id>`, or `/stats <id>` for that character
- **THEN** the character does not appear in the listing and the sheet/stats request returns the not-found response
