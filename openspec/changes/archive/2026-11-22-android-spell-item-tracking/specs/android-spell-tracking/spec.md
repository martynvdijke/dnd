# android-spell-tracking Specification

## Purpose

Spell tracking in the Android app reflects real character spellcasting state from the server instead of an always-empty placeholder tab.

## ADDED Requirements

### Requirement: Spellcasting state renders from server data

The app SHALL parse the `spellcasting` and `spells` sub-resources returned by `GET /api/characters/:id` and SHALL render spell slots (used/total per level), prepared spells, and known spells in the character sheet SPELLS tab.

#### Scenario: Character with spellcasting shows real slots

- WHEN a character with spellcasting and expended slots is opened
- THEN the SPELLS tab shows each slot level's used/total from the server response and lists prepared and known spells

#### Scenario: Character without spellcasting shows accurate empty state

- WHEN a character without the spellcasting sub-resource is opened
- THEN the SPELLS tab shows an explanatory empty state stating the character has no spellcasting

### Requirement: Cast a spell via the real spell identifier

The app SHALL expend a spell slot by calling `POST /api/characters/:id/cast-spell` with a spell identifier and slot level taken from the parsed spells sub-resource, and SHALL update the displayed slots from the server response.

#### Scenario: Cast a prepared spell

- WHEN the player casts a prepared spell at its slot level
- THEN the app posts that spell's real identifier to `/cast-spell` and the displayed slots reflect the server's updated usage

#### Scenario: Cast failure is surfaced

- WHEN the server rejects a cast attempt
- THEN the app displays the error and leaves the displayed slots unchanged
