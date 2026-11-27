# gcal-color-label-filter Specification

## Purpose
Filtering public events by Google Calendar native color labels (colorId), with a configurable filter mode combining or replacing text-tag filtering.

## Requirements

### Requirement: Events can be filtered by Google Calendar color labels

The system SHALL support filtering events by their Google Calendar color label (`colorId`). The admin SHALL configure a comma-separated list of color labels (either numeric color IDs or GCal color names). When color-label filtering is active, only events whose `colorId` matches a configured label SHALL be included. An empty color-label list SHALL act as a pass-through (all events pass the color filter).

#### Scenario: Filter by color ID
- **WHEN** color labels are configured as "1,5"
- **AND** an event has `colorId` "1"
- **THEN** the event is included in results

#### Scenario: Event with non-matching color is excluded
- **WHEN** color labels are configured as "5"
- **AND** an event has `colorId` "1"
- **THEN** the event is excluded from results

#### Scenario: Event with no color is excluded when color filter is active
- **WHEN** color labels are configured as "1"
- **AND** an event has no `colorId` (empty/unset)
- **THEN** the event is excluded from results

#### Scenario: Empty color-label list is a pass-through
- **WHEN** color-label filtering is active
- **AND** no color labels are configured (empty list)
- **THEN** all events pass the color filter regardless of their `colorId`

#### Scenario: Color names are resolved to color IDs
- **WHEN** color labels are configured as "lavender,blue"
- **AND** the GCal Colors API maps "lavender" to ID "1" and "blue" to ID "5"
- **THEN** events with `colorId` "1" or "5" are included

---

### Requirement: Filter mode is configurable

The system SHALL support a `FilterMode` setting with values `text`, `color`, or `both`. The default SHALL be `text`, preserving existing text-tag-only behavior. In `both` mode, an event SHALL be included only if it passes both the text-tag filter and the color-label filter. If either filter's configured list is empty, that filter SHALL be treated as a pass-through.

#### Scenario: Text-only mode (default)
- **WHEN** `FilterMode` is "text"
- **THEN** only the text-tag filter is applied
- **AND** color labels are ignored

#### Scenario: Color-only mode
- **WHEN** `FilterMode` is "color"
- **THEN** only the color-label filter is applied
- **AND** text tags are ignored

#### Scenario: Both mode with AND logic
- **WHEN** `FilterMode` is "both"
- **AND** text tags include "dnd"
- **AND** color labels include "1"
- **AND** an event has title "Weekly DnD" and `colorId` "1"
- **THEN** the event is included

#### Scenario: Both mode excludes event failing either filter
- **WHEN** `FilterMode` is "both"
- **AND** text tags include "dnd"
- **AND** color labels include "1"
- **AND** an event has title "Weekly DnD" and `colorId` "5"
- **THEN** the event is excluded

#### Scenario: Both mode with empty text tags
- **WHEN** `FilterMode` is "both"
- **AND** no text tags are configured
- **AND** color labels include "1"
- **THEN** the text filter is a pass-through
- **AND** only the color filter determines inclusion

---

### Requirement: Event color is exposed for rendering

The system SHALL include each event's Google Calendar `colorId` in the `Event` struct and make it available to templates. The events list SHALL render a color swatch (colored dot) before the event title when a `colorId` is present.

#### Scenario: Event with color renders swatch
- **WHEN** an event has `colorId` "1"
- **THEN** the list entry displays a colored dot matching GCal color 1 before the title

#### Scenario: Event without color renders no swatch
- **WHEN** an event has no `colorId`
- **THEN** no color swatch is rendered

---

### Requirement: Color labels and filter mode are configurable via admin settings

The system SHALL store `ColorLabels` and `FilterMode` in the `events_settings` table. An admin SHALL be able to update these via the admin UI without restarting the server. The admin UI SHALL provide a filter-mode selector and a color-labels input field with help text listing the standard GCal color names.

#### Scenario: Admin sets color labels
- **WHEN** an admin enters "lavender,blue" in the color labels field and saves
- **THEN** the value is persisted to the database
- **AND** the next event fetch resolves names to IDs and applies the color filter

#### Scenario: Admin changes filter mode
- **WHEN** an admin changes the filter mode from "text" to "color" and saves
- **THEN** subsequent event fetches use color-label filtering only

#### Scenario: Default filter mode is text
- **WHEN** no filter mode has been configured (fresh deployment or migration)
- **THEN** the filter mode defaults to "text"
- **AND** existing text-tag behavior is unchanged
