# pwa-support Specification

## Purpose
Progressive Web App support — service worker for offline character access, install prompt, standalone display mode.

## Requirements

### Requirement: Service Worker Registration
The system SHALL register a service worker on app load. The current service worker unregistration code in `app.html` SHALL be removed.

#### Scenario: Service worker registers on first load
- **WHEN** a user loads the app for the first time in a browser that supports service workers
- **THEN** the service worker (`/sw.js`) SHALL be registered with scope `/`

#### Scenario: Service worker unregistration code is removed
- **WHEN** the app loads
- **THEN** the existing code that unregisters service workers (currently on line 211 of `app.html`) SHALL NOT execute

#### Scenario: Service worker updates on new version
- **WHEN** a new version of the service worker is deployed
- **THEN** the new service worker SHALL activate after the user refreshes the page, replacing the old worker

### Requirement: App Shell Caching
The service worker SHALL cache the app shell (HTML, CSS, JavaScript) using a cache-first strategy. Cached assets SHALL be served from cache when the network is unavailable.

#### Scenario: App shell cached on first load
- **WHEN** a user loads the app with network connectivity
- **THEN** the app shell files (HTML, CSS, JS bundles) SHALL be cached by the service worker

#### Scenario: App shell served from cache offline
- **WHEN** a user loads the app without network connectivity
- **THEN** the app shell SHALL load from the service worker cache

#### Scenario: Cache updates on new deployment
- **WHEN** a new version of the app is deployed and the user refreshes
- **THEN** the service worker SHALL fetch and cache the updated app shell files

### Requirement: Character Data Offline Access
The service worker SHALL cache character sheet API responses for the logged-in user's characters. Cached character data SHALL be viewable offline in read-only mode.

#### Scenario: Character data cached on view
- **WHEN** a user views a character sheet while online
- **THEN** the API response for that character's data SHALL be cached by the service worker

#### Scenario: Character sheet viewable offline
- **WHEN** a user navigates to a previously viewed character sheet while offline
- **THEN** the character sheet SHALL display using cached data

#### Scenario: Offline character sheet shows sync status
- **WHEN** a character sheet is displayed from cached data while offline
- **THEN** a "Last synced: [timestamp]" indicator SHALL be visible on the sheet

#### Scenario: Edits disabled while offline
- **WHEN** a user views a character sheet while offline
- **THEN** edit actions (modify stats, change inventory, etc.) SHALL be disabled and a "Read-only — offline" message SHALL be displayed

### Requirement: PWA Web Manifest
The system SHALL serve a Web App Manifest (`manifest.json`) that enables app installation on supported browsers.

#### Scenario: Manifest is linked in HTML
- **WHEN** the app HTML is loaded
- **THEN** a `<link rel="manifest" href="/manifest.json">` tag SHALL be present in the document head

#### Scenario: Manifest contains required fields
- **WHEN** the manifest is fetched
- **THEN** it SHALL contain: `name` ("Villum"), `short_name` ("Villum"), `start_url` ("/app"), `display` ("standalone"), `background_color`, `theme_color`, and at least two icon sizes (192x192 and 512x512)

#### Scenario: Go handler serves manifest with correct content type
- **WHEN** a browser requests `/manifest.json`
- **THEN** the Go server SHALL respond with the manifest file and `Content-Type: application/manifest+json` header

### Requirement: PWA Install Prompt
The system SHALL capture the browser's `beforeinstallprompt` event and present a custom install prompt to the user.

#### Scenario: Install prompt captured
- **WHEN** the browser fires the `beforeinstallprompt` event (user meets installability criteria)
- **THEN** the system SHALL store the event and display an "Install App" button in the More menu or as a dismissible banner

#### Scenario: User installs the app
- **WHEN** the user taps the "Install App" button
- **THEN** the browser's native install prompt SHALL be triggered, and upon confirmation, the app SHALL be installed

#### Scenario: Install prompt dismissed
- **WHEN** the user dismisses the install prompt
- **THEN** the prompt SHALL NOT reappear for at least 30 days (tracked in `localStorage`)

### Requirement: Standalone Display Mode
When launched from the home screen or app drawer, the app SHALL open in standalone display mode without browser chrome.

#### Scenario: App launches in standalone mode
- **WHEN** a user opens the installed PWA from their home screen or app drawer
- **THEN** the app SHALL display without browser address bar, navigation buttons, or other browser chrome

#### Scenario: Standalone mode status bar theming
- **WHEN** the app is running in standalone mode
- **THEN** the status bar SHALL use the `theme_color` specified in the manifest
