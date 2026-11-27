# pdf-attachment-links Specification

## Purpose
PDF attachments in the media gallery and entity detail views are clickable and open the PDF viewer. External PDF URLs are also supported.

## Requirements

### Requirement: Media gallery PDF items are clickable
PDF items in the media gallery SHALL open the PDF viewer when clicked.

#### Scenario: Clicking a PDF icon in media gallery
- **WHEN** the user clicks on a PDF media gallery item (the PDF icon or card)
- **THEN** the system SHALL call `openPdfViewer(url, title)` with the PDF's URL and filename
- **AND** SHALL NOT trigger a file download or navigation away from the current page

#### Scenario: Non-PDF media items remain unchanged
- **WHEN** the user clicks on a non-PDF media gallery item (image)
- **THEN** the system SHALL behave as before (no PDF viewer, current behavior preserved)

### Requirement: PDF viewer function is globally accessible
The system SHALL provide a global `openPdfViewer(url, title?)` function callable from HTMX templates and inline scripts.

#### Scenario: Function exists on window
- **WHEN** any script or template calls `window.openPdfViewer(url, title)`
- **THEN** the function SHALL exist and open the PDF viewer as specified

#### Scenario: Calling with external URL
- **WHEN** `openPdfViewer` is called with an external URL (e.g., `https://example.com/doc.pdf`)
- **THEN** the viewer SHALL fetch and render the PDF from that URL
- **AND** SHALL respect CORS restrictions (if the external server blocks cross-origin requests, the viewer SHALL show an appropriate error)
