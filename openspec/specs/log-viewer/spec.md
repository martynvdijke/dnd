# log-viewer Specification

## Purpose
Admin UI log viewer with search, dynamic source filtering, detail expansion, and export.

## Requirements

### Requirement: Log viewer has plain-text search
The admin log viewer SHALL provide a text search input that filters log entries client-side by message content (case-insensitive substring match).

#### Scenario: Search filters visible entries
- **WHEN** the user types a search query into the search input
- **THEN** only log entries whose `message` field contains the query string (case-insensitive) SHALL be displayed
- **AND** entries that don't match SHALL be hidden
- **AND** the entry count display SHALL update to reflect filtered count

### Requirement: Log source filter is dynamic
The source filter dropdown in the admin log viewer SHALL be populated dynamically from the log data, rather than using a hardcoded list. A new API endpoint `GET /api/admin/log-sources` SHALL return distinct source values from the ring buffer.

#### Scenario: Log sources are fetched on tab activation
- **WHEN** the user navigates to the Logs tab in the admin panel
- **THEN** the system SHALL fetch `GET /api/admin/log-sources` (or extract from loaded log entries)
- **AND** populate the source filter dropdown with these values plus an "All" option

#### Scenario: GET /api/admin/log-sources endpoint
- **WHEN** `GET /api/admin/log-sources` is called
- **THEN** it SHALL return a JSON array of unique source strings currently in the LogBuffer
- **AND** return an empty array if no entries exist

### Requirement: Log entry details can be expanded
The admin log viewer SHALL allow the user to click on a log entry row to expand and view its full attributes in a readable format.

#### Scenario: Clicking a log row expands details
- **WHEN** the user clicks on a log entry row
- **THEN** a details section SHALL expand below the row showing all attributes as key-value pairs
- **AND** clicking again SHALL collapse the details
- **AND** long attribute values SHALL wrap or be scrollable

### Requirement: Log entries can be exported
The admin log viewer SHALL provide a button to export currently visible (filtered) log entries as JSON or CSV.

#### Scenario: Export as JSON
- **WHEN** the user clicks the Export button with JSON format selected
- **THEN** the system SHALL download a `.json` file containing the currently visible log entries as a JSON array

#### Scenario: Export as CSV
- **WHEN** the user clicks the Export button with CSV format selected
- **THEN** the system SHALL download a `.csv` file containing the currently visible log entries with columns: level, timestamp, source, message, attributes

### Requirement: Log viewer auto-refresh is toggleable
The log viewer SHALL provide a toggle to enable/disable the 5-second auto-refresh interval.

#### Scenario: Auto-refresh can be turned off
- **WHEN** the user clicks the auto-refresh toggle
- **THEN** the auto-refresh interval SHALL be cleared
- **AND** the display SHALL update to show "Auto-refresh: off"

#### Scenario: Auto-refresh can be turned back on
- **WHEN** the user clicks the auto-refresh toggle while it is off
- **THEN** the auto-refresh interval SHALL be re-established
- **AND** the display SHALL update to show "Auto-refresh: on"
