# public-sharing Specification

## Purpose
Creating, viewing, and revoking public share links for notes, journal entries, campaign maps, and uploaded files, with ownership enforcement and public read-only access (JSON + HTML).

## Requirements

### Requirement: Create share link for content entities
A user SHALL be able to create a public share link for any of the following entity types: `note` (character_notes), `journal` (journal), `map` (campaign_maps), `upload` (uploads), in addition to the existing `character` and `party` types. Creation SHALL require authentication and SHALL verify the requester owns the underlying entity or is a member of its campaign, or is an admin. Share links SHALL carry a cryptographically random 128-bit token, an optional expiration, and SHALL be revocable.

#### Scenario: Owner shares a note
- **WHEN** an authenticated user creates a share link for a note whose `character_id` belongs to one of their characters
- **THEN** the system returns the token and a share URL

#### Scenario: Non-owner cannot share
- **WHEN** an authenticated user attempts to create a share link for a note, journal entry, map, or upload they do not own and are not a member of its campaign
- **THEN** the system returns HTTP 403 and creates no link

#### Scenario: Campaign member shares a map
- **WHEN** an authenticated user who is a `campaign_members` row for a campaign creates a share link for a map belonging to that campaign
- **THEN** the system returns the token and a share URL

#### Scenario: Upload owned by empty owner type
- **WHEN** an authenticated non-admin user attempts to create a share link for an upload whose `owner_type` is empty
- **THEN** the system returns HTTP 403

#### Scenario: Upload owned via linked entity
- **WHEN** an authenticated user creates a share link for an upload whose `owner_type`/`owner_id` resolves to an entity they own (character, npc, campaign, oneshot adventure, or item)
- **THEN** the system returns the token and a share URL

#### Scenario: Admin shares anything
- **WHEN** an admin user creates a share link for any supported entity type
- **THEN** the system returns the token and a share URL regardless of ownership

### Requirement: Public JSON access to shared content
Any unauthenticated client holding a valid, unexpired share token SHALL be able to fetch a JSON representation of the shared entity via `GET /api/share/:token`. For `note` the payload SHALL include id, title, content, category, visibility, created_at, character name, and owner name. For `journal` it SHALL include id, title, entry, entry_date, created_at, character name, and owner name. For `map` it SHALL include id, campaign name, name, image_url, width, height, grid_size, and pins (excluding hidden pins). For `upload` it SHALL include id, url, resized_url, thumbnail_url, and ext.

#### Scenario: Fetch shared note
- **WHEN** a client requests `GET /api/share/:token` with a valid token for a note
- **THEN** the system returns HTTP 200 with the note's title, content, category, character name, and owner name

#### Scenario: Fetch shared map with pins
- **WHEN** a client requests `GET /api/share/:token` with a valid token for a map that has non-hidden pins
- **THEN** the system returns HTTP 200 with the map metadata and the non-hidden pins

#### Scenario: Fetch shared map hides hidden pins
- **WHEN** a client requests `GET /api/share/:token` with a valid token for a map that has hidden pins
- **THEN** the response pin list SHALL NOT include pins flagged `is_hidden`

#### Scenario: Unknown or expired token
- **WHEN** a client requests `GET /api/share/:token` with an unknown token or a token past its `expires_at`
- **THEN** the system returns HTTP 404 for unknown tokens and HTTP 410 for expired tokens

### Requirement: Public HTML view pages for shared content
Any unauthenticated client holding a valid share token SHALL be able to view a human-readable HTML page at `GET /share/:token` for `note`, `journal`, `map`, and `upload` entity types. Pages SHALL be server-rendered, SHALL escape all entity content to prevent XSS, SHALL include a `noindex` robots meta tag, and SHALL NOT require a session. Unknown tokens SHALL return HTTP 404; expired tokens HTTP 410.

#### Scenario: View shared note page
- **WHEN** a client opens `GET /share/:token` for a note token
- **THEN** the system returns HTTP 200 with an HTML page showing the note title and content

#### Scenario: View shared map page
- **WHEN** a client opens `GET /share/:token` for a map token
- **THEN** the system returns HTTP 200 with an HTML page showing the map image, its name, and its non-hidden pins rendered over the image at their x/y positions

#### Scenario: Shared pages are not indexed
- **WHEN** a client fetches any `GET /share/:token` HTML page
- **THEN** the page SHALL contain a `<meta name="robots" content="noindex">` tag

#### Scenario: Malicious content is escaped
- **WHEN** a shared note or journal entry contains HTML or script tags in its content
- **THEN** the rendered page displays the content as plain text and SHALL NOT execute it

### Requirement: Shared uploads render inline
The public HTML page for an `upload` share link SHALL render image uploads inline as an `<img>`, PDF uploads via an embedded viewer, and other types as a download link, all referencing the media URL. Upload media SHALL remain served by the existing public `/media` static route (content-addressed filenames).

#### Scenario: View shared image
- **WHEN** a client opens `GET /share/:token` for an image upload token
- **THEN** the page contains an `<img>` element whose src is the upload's public media URL

#### Scenario: View shared PDF
- **WHEN** a client opens `GET /share/:token` for a PDF upload token
- **THEN** the page embeds the PDF via an inline viewer referencing the upload's public media URL

### Requirement: Manage share links
An authenticated user SHALL be able to list their own share links (token, entity type, entity id, created_at, expires_at, and a human-readable label for the shared entity) via `GET /share`, and revoke any of their own links via `DELETE /share/:token`. Admins SHALL see and revoke all links. The UI SHALL provide a management dialog with list and revoke actions.

#### Scenario: List own share links
- **WHEN** an authenticated user calls `GET /share`
- **THEN** the system returns only links created by that user, including entity labels

#### Scenario: Revoke own share link
- **WHEN** an authenticated user calls `DELETE /share/:token` for a link they created
- **THEN** the link is deleted and subsequent fetches of that token return HTTP 404

#### Scenario: Non-owner cannot revoke
- **WHEN** an authenticated user calls `DELETE /share/:token` for a link created by another non-admin user
- **THEN** the system returns HTTP 403

### Requirement: Share button in content UIs
The notes list, journal list, campaign maps view, and uploads gallery SHALL each provide a "Share" action on individual items that creates a share link for that item and presents the resulting URL for copying. The URL SHALL point to the public HTML page.

#### Scenario: Share a note from the notes list
- **WHEN** a user clicks Share on a note in the notes list
- **THEN** a share link is created and a dialog shows the `GET /share/:token` URL with a copy button

#### Scenario: Share an upload from the gallery
- **WHEN** a user clicks Share on an upload in the uploads gallery
- **THEN** a share link is created and a dialog shows the `GET /share/:token` URL with a copy button
