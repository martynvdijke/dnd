# events-share-link-ux Specification

## Purpose
Admin UI for copying and sharing the public events URL, including inline QR code generation and PNG download, with optional per-campaign link support.

## Requirements

### Requirement: Admin Events tab shows the public link

The system SHALL display the full public events URL in the admin Events tab. The URL SHALL be constructed from a configured base URL (`BASE_URL` env var) if set, otherwise from the request host. The URL SHALL be shown in a read-only field.

#### Scenario: Public link displayed with BASE_URL set
- **WHEN** the `BASE_URL` environment variable is set to "https://dnd.example.com"
- **THEN** the admin Events tab shows "https://dnd.example.com/events"

#### Scenario: Public link displayed from request host
- **WHEN** no `BASE_URL` is set
- **AND** the admin page is accessed at "https://192.168.1.5/events-settings"
- **THEN** the admin Events tab shows "https://192.168.1.5/events"

#### Scenario: Public link for a campaign
- **WHEN** a campaign with slug "lost-mines" exists
- **THEN** the admin shows the campaign public link "https://{base}/events/c/lost-mines"

---

### Requirement: Copy link button

The admin Events tab SHALL include a "Copy link" button that copies the public events URL to the clipboard. A visual confirmation ("Copied!") SHALL appear for at least 2 seconds after a successful copy.

#### Scenario: Copy link succeeds
- **WHEN** the admin clicks "Copy link"
- **THEN** the public URL is copied to the clipboard
- **AND** a "Copied!" confirmation is displayed for 2 seconds

#### Scenario: Copy link fallback on unsupported API
- **WHEN** the browser does not support the Clipboard API
- **AND** the admin clicks "Copy link"
- **THEN** the system falls back to a legacy copy method
- **AND** the URL is still copied to the clipboard

---

### Requirement: QR code display

The admin Events tab SHALL display an inline QR code encoding the public events URL. The QR code SHALL be rendered as an image fetched from a server-side endpoint.

#### Scenario: QR code is displayed
- **WHEN** the admin views the Events tab
- **THEN** a QR code image is visible encoding the public events URL

#### Scenario: QR code for a campaign
- **WHEN** a campaign with slug "lost-mines" exists
- **THEN** a QR code encoding "https://{base}/events/c/lost-mines" is displayed for that campaign

---

### Requirement: QR code download

The admin Events tab SHALL include a "Download QR" button that saves the QR code as a PNG file.

#### Scenario: Download QR as PNG
- **WHEN** the admin clicks "Download QR"
- **THEN** a PNG file containing the QR code is downloaded
- **AND** the filename includes "events-qr" or the campaign slug

---

### Requirement: Server-side QR generation endpoint

The system SHALL provide an admin endpoint `GET /admin/events/qr?slug={slug}` that returns a PNG image encoding the public events URL for the global page (no slug) or a specific campaign (with slug). The endpoint SHALL require admin authentication.

#### Scenario: Global QR endpoint
- **WHEN** an authenticated admin requests `GET /admin/events/qr`
- **THEN** the response has `Content-Type: image/png`
- **AND** the response body is a valid PNG encoding the global `/events` URL

#### Scenario: Campaign QR endpoint
- **WHEN** an authenticated admin requests `GET /admin/events/qr?slug=lost-mines`
- **THEN** the response has `Content-Type: image/png`
- **AND** the response body is a valid PNG encoding the `/events/c/lost-mines` URL

#### Scenario: Unauthenticated request is rejected
- **WHEN** an unauthenticated user requests `GET /admin/events/qr`
- **THEN** the response is HTTP 401 or 403
