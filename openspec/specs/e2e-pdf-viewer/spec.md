# e2e-pdf-viewer Specification

## Purpose
Playwright e2e tests for PDF viewer modal, navigation, zoom, download.

## Requirements

### Requirement: PDF viewer modal e2e coverage
The Playwright e2e tests SHALL verify the PDF viewer modal can be opened and navigated.

#### Scenario: Open PDF viewer from media gallery
- **WHEN** user clicks a PDF file in the media gallery
- **THEN** the PDF viewer modal SHALL open with the PDF displayed

#### Scenario: PDF viewer navigation controls
- **WHEN** user clicks the next/previous page buttons in the PDF viewer
- **THEN** the displayed page SHALL change accordingly

#### Scenario: PDF viewer zoom controls
- **WHEN** user clicks zoom in/out buttons in the PDF viewer
- **THEN** the PDF display zoom level SHALL change

#### Scenario: Close PDF viewer
- **WHEN** user clicks the close button on the PDF viewer modal
- **THEN** the modal SHALL close and be removed from the DOM

#### Scenario: Keyboard navigation in PDF viewer
- **WHEN** user presses left/right arrow keys while PDF viewer is open
- **THEN** the page SHALL navigate accordingly
