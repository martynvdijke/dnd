# offline-resilience Specification

## Purpose
Service worker caching strategy for CDN assets, graceful degradation when offline, cached fallback for static resources.

## Requirements

### Requirement: CDN assets SHALL be cached by the service worker on first access
The service worker SHALL intercept requests to known CDN URLs (Bootstrap, Font Awesome, Google Fonts, Leaflet, HTMX) and serve them from a cache named `villum-cdn-v1` using a cache-first strategy. On cache miss, the service worker SHALL fetch from network, populate the cache, and return the response.

#### Scenario: First visit loads CDN assets
- **WHEN** a user visits the app for the first time
- **THEN** CDN assets are fetched from network and stored in `villum-cdn-v1` cache
- **THEN** subsequent visits serve CDN assets from cache without network requests

#### Scenario: Offline navigation loads cached assets
- **WHEN** a user navigates the app while offline
- **THEN** Bootstrap, Font Awesome, and other cached CDN assets are served from the service worker cache
- **THEN** the page renders with full styling and icons

### Requirement: Google Fonts SHALL be self-hosted as a fallback
The app SHALL bundle the used Google Font subsets (Playfair Display, Vollkorn) as self-hosted `woff2` files in the `static/fonts/` directory. When online, the app SHALL use Google Fonts CDN as primary. When offline, the service worker SHALL fall back to the self-hosted fonts.

#### Scenario: Online uses Google Fonts CDN
- **WHEN** the user is online
- **THEN** the page loads fonts from `fonts.googleapis.com`
- **THEN** fonts render with optimal loading from Google CDN

#### Scenario: Offline falls back to self-hosted fonts
- **WHEN** the user is offline
- **THEN** the service worker serves self-hosted font files from `static/fonts/`
- **THEN** all text renders with the correct typeface

### Requirement: Static app pages SHALL be available offline
The service worker SHALL precache the app shell (app.html, style.css, app.js) on installation using a `villum-static-v1` cache. The app shell SHALL be served from cache when the user navigates to `/` while offline.

#### Scenario: App shell loads offline
- **WHEN** a user navigates to the app root while offline
- **THEN** the app shell (HTML, CSS, JS) is served from the service worker cache
- **THEN** the user sees the app UI with a "You are offline" indicator

### Requirement: Offline indicator SHALL be shown to users
When the browser reports `navigator.onLine === false` or the service worker detects a failed fetch, the app SHALL display a non-intrusive banner: "You are offline — some features may be limited."

#### Scenario: Offline indicator appears on disconnect
- **WHEN** the browser goes offline
- **THEN** a banner appears at the top of the page with the offline message
- **THEN** the banner disappears automatically when connectivity returns
