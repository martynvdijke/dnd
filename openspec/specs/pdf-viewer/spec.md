# pdf-viewer Specification

## Purpose
In-browser PDF viewing with page navigation, zoom controls, thumbnail preview, and download. Opens as a modal overlay when clicking on a PDF attachment or media gallery item.

## Requirements

### Requirement: PDF viewer opens as a full-screen modal
The system SHALL provide an in-browser PDF viewer displayed as a full-screen Bootstrap modal overlay.

#### Scenario: Opening viewer with a valid PDF URL
- **WHEN** the user triggers `openPdfViewer(url, title)` with a valid PDF URL
- **THEN** the viewer modal SHALL open
- **AND** display a loading state while the PDF is being fetched
- **AND** render the first page once loaded
- **AND** show the document title (if provided) in the modal header

#### Scenario: Opening viewer with an invalid URL
- **WHEN** the user triggers `openPdfViewer(url, title)` with a URL that returns a non-PDF response (e.g., 404, non-PDF content)
- **THEN** the modal SHALL display an error message
- **AND** SHALL NOT attempt to render pages

#### Scenario: Closing the viewer
- **WHEN** the user clicks the modal close button, the backdrop, or presses Escape
- **THEN** the modal SHALL close
- **AND** the PDF document object SHALL be released from memory

### Requirement: Page navigation
The viewer SHALL allow navigating between pages of the PDF.

#### Scenario: Navigating to next page
- **WHEN** the user clicks the "Next" button or presses the Right Arrow key
- **AND** the current page is not the last page
- **THEN** the viewer SHALL render the next page
- **AND** update the page number indicator

#### Scenario: Navigating to previous page
- **WHEN** the user clicks the "Previous" button or presses the Left Arrow key
- **AND** the current page is not the first page
- **THEN** the viewer SHALL render the previous page
- **AND** update the page number indicator

#### Scenario: Navigating past the last page
- **WHEN** the user clicks "Next" on the last page
- **THEN** the viewer SHALL NOT advance past the last page
- **AND** the "Next" button SHALL appear disabled

#### Scenario: Navigating before the first page
- **WHEN** the user clicks "Previous" on the first page
- **THEN** the viewer SHALL NOT go before the first page
- **AND** the "Previous" button SHALL appear disabled

### Requirement: Zoom controls
The viewer SHALL provide zoom in, zoom out, and fit-to-width controls.

#### Scenario: Zoom in
- **WHEN** the user clicks the "Zoom In" button
- **THEN** the viewer SHALL increase the scale factor by 25%
- **AND** re-render the current page at the new scale

#### Scenario: Zoom out
- **WHEN** the user clicks the "Zoom Out" button
- **THEN** the viewer SHALL decrease the scale factor by 25%
- **AND** re-render the current page at the new scale
- **AND** SHALL NOT zoom below 25% of original size

#### Scenario: Fit to width
- **WHEN** the user clicks the "Fit to Width" button
- **THEN** the viewer SHALL calculate the scale that makes the page width match the canvas width
- **AND** re-render the current page at that scale

### Requirement: Download button
The viewer SHALL provide a button to download the PDF file.

#### Scenario: Download
- **WHEN** the user clicks the "Download" button
- **THEN** the browser SHALL download the PDF file with its original filename (if known) or `document.pdf`

### Requirement: Page number display
The viewer SHALL show the current page number and total page count.

#### Scenario: Page indicator update
- **WHEN** the current page changes (via navigation or initial load)
- **THEN** the page indicator SHALL show "Page X of Y"

### Requirement: PDF.js is loaded lazily
The PDF.js library SHALL NOT be loaded until the first time a PDF is opened, to avoid upfront bandwidth and parsing cost.

#### Scenario: First viewer open loads PDF.js
- **WHEN** `openPdfViewer(url, title)` is called for the first time in a session
- **THEN** the system SHALL dynamically load the PDF.js script from CDN
- **AND** SHALL wait for it to load before fetching the PDF
- **AND** SHALL cache the loaded script for subsequent opens

#### Scenario: Subsequent opens use cached PDF.js
- **WHEN** `openPdfViewer(url, title)` is called after a previous successful open
- **THEN** the system SHALL NOT re-load the PDF.js script
- **AND** SHALL proceed directly to fetching and rendering the PDF
