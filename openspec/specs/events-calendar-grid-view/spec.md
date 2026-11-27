# events-calendar-grid-view Specification

## Purpose
A month-grid calendar view of upcoming events with month navigation, and per-event detail pages showing full event information with optional map links.

## Requirements

### Requirement: Month-grid calendar view

The system SHALL provide a month-grid calendar view of events on the events page, toggled via a `?view=grid` query parameter. The grid SHALL display the current month as a 7-column calendar with events placed on their start-date cells. The default view SHALL remain the list view (`?view=list` or no parameter).

#### Scenario: Grid view loads
- **WHEN** a user visits `/events?view=grid`
- **THEN** the page displays a month-grid calendar for the current month
- **AND** events are shown on their corresponding start-date cells

#### Scenario: Default view is list
- **WHEN** a user visits `/events` with no view parameter
- **THEN** the page displays the list view (existing behavior)

#### Scenario: View switcher is available
- **WHEN** a user views the events page
- **THEN** a toggle control is visible allowing switching between "List" and "Calendar" views

#### Scenario: Grid cell shows event titles
- **WHEN** a date cell has events
- **THEN** up to 3 event titles are displayed in the cell
- **AND** if more than 3 events exist, a "+N more" indicator is shown

#### Scenario: Empty date cell
- **WHEN** a date cell has no events
- **THEN** the cell shows only the day number

---

### Requirement: Month navigation via HTMX

The system SHALL provide month navigation (previous, next, today) on the grid view that updates the grid without a full page reload. A new endpoint `GET /htmx/events/grid?month=YYYY-MM` SHALL return the grid HTML for the requested month.

#### Scenario: Navigate to next month
- **WHEN** a user clicks the "Next" button on the grid view
- **THEN** the grid updates to show the next month via HTMX
- **AND** the page does not fully reload

#### Scenario: Navigate to previous month
- **WHEN** a user clicks the "Previous" button on the grid view
- **THEN** the grid updates to show the previous month via HTMX

#### Scenario: Navigate to current month
- **WHEN** a user clicks the "Today" button on the grid view
- **THEN** the grid updates to show the current month via HTMX

#### Scenario: Grid endpoint returns HTML
- **WHEN** a client requests `GET /htmx/events/grid?month=2026-07`
- **THEN** the response is HTML containing the July 2026 calendar grid
- **AND** events for July 2026 are rendered in their date cells

---

### Requirement: Event detail page

The system SHALL provide a public event detail page at `GET /events/{id}` showing the full information for a single event: title, start and end date/time, location, and description. When the event has a location, the page SHALL include a link to view the location on Google Maps.

#### Scenario: Event detail page loads
- **WHEN** a user visits `/events/{id}` for an existing event
- **THEN** the page displays the event's title, date/time, location, and description

#### Scenario: Event with location shows map link
- **WHEN** an event has a location "123 Main St, Springfield"
- **THEN** the detail page includes a "View on Google Maps" link
- **AND** the link URL encodes the location as a search query

#### Scenario: Event without location shows no map link
- **WHEN** an event has no location
- **THEN** the detail page does not display a map link

#### Scenario: Event not found returns 404
- **WHEN** a user visits `/events/{id}` for a non-existent event ID
- **THEN** the server returns HTTP 404 with a friendly "Event not found" message

#### Scenario: Back to events link
- **WHEN** a user views an event detail page
- **THEN** a "Back to events" link is visible
- **AND** clicking it returns to the events page

---

### Requirement: Event entries are clickable

In both the list and grid views, event titles SHALL be links to the event detail page at `/events/{id}`.

#### Scenario: List view entry is clickable
- **WHEN** a user clicks an event title in the list view
- **THEN** the browser navigates to that event's detail page

#### Scenario: Grid view entry is clickable
- **WHEN** a user clicks an event title in a grid cell
- **THEN** the browser navigates to that event's detail page
