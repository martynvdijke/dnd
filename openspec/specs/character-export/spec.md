# character-export Specification

## Purpose
TBD - created by archiving change add-content-ecosystem. Update Purpose after archive.
## Requirements
### Requirement: Printable character sheet
The system SHALL render a character as a self-contained HTML document suitable for printing, including identity, ability scores, combat stats, skills, features, spells, inventory and currency where present.

#### Scenario: HTML sheet requested
- **WHEN** a user requests `GET /api/characters/:id/print?format=html`
- **THEN** the response is `text/html` containing the character's name and core stats

#### Scenario: Default format unchanged
- **WHEN** a user requests `GET /api/characters/:id/print` without a format
- **THEN** the response remains `text/plain` as before

### Requirement: Export PDF action
The character sheet SHALL offer an export action that opens the printable HTML sheet and invokes the browser print dialog, allowing the user to save it as PDF without a server-side PDF dependency.

#### Scenario: Export PDF from the sheet
- **WHEN** a user activates the Export PDF action on a character sheet
- **THEN** the printable HTML sheet opens and the browser print dialog is invoked
