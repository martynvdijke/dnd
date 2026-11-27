# image-crop-resize Specification

## Purpose
Server-side crop-and-resize endpoint for uploaded images.

## Requirements

### Requirement: Server-side crop endpoint

The system SHALL provide a POST endpoint to crop and resize an uploaded image on the server.

#### Scenario: Crop image with coordinates
- **WHEN** user POSTs to `/api/uploads/:id/crop` with `{x, y, width, height, targetWidth, targetHeight}`
- **THEN** the server extracts the sub-image region, resizes to target dimensions, saves as a new variant, and updates the upload's `resized_url`

#### Scenario: Crop validates bounds
- **WHEN** user sends crop coordinates outside the image bounds (e.g., x + width > image width)
- **THEN** the server returns a 400 Bad Request error

#### Scenario: Crop non-image returns error
- **WHEN** user POSTs to crop endpoint for a PDF upload
- **THEN** the server returns a 400 Bad Request error (crop is for images only)
