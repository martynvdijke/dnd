## Why

Villum (`dnd`) is a TTRPG campaign and character manager used during live D&D sessions around a physical table. Players and Dungeon Masters need a native Flutter Android pocket companion for tracking character HP, spell slots, inventory, conditions, initiative order, interactive battlemap tokens, compendium lookups, and AI drafting without crowding the table with laptops.

## What Changes

- Add a Flutter Android application under `mobile/` (`xyz.vandijke.dnd`) targeting the existing Go `/api/*` and `/api/ws` endpoints.
- Authenticate against the existing three-layer API auth: `POST /api/login` (with `GET /api/check-setup` for first-run admin creation) establishing a session cookie, a locally derived `X-CSRF-Token`, and a `vlt_*` bearer token provisioned via `POST /api/tokens`. There is no self-service registration endpoint.
- Implement Interactive Character Sheet tabs (`/api/characters/:id`, `/hp`, `/currency`, `/inventory`, `/spells`, `/spellcasting`, `/cast-spell`, `/classes`, `/features`) with 1-tap HP/temp-HP adjustment, death saves, spell slot expenditure, conditions, and Short/Long Rest recovery.
- Implement Live Initiative & Combat Tracker (`/api/combat/attack`, `/api/combat`, `/api/ws`) and Pinch-to-Zoom Battlemap (`/api/campaigns/:id/battlemap`, `/api/battlemap-tokens/:id`, `/api/campaigns/:id/battlemap/sync`) with draggable player tokens.
- Implement DM Pocket Screen & Compendium Search (`/api/campaigns/:id/dm-screen`, `/api/compendium/search`, `/api/generate/*`) and conversational AI Draft Assistant (`/api/ai/draft`). The DM-only endpoints are gated on the authenticated user's role.
- Add a non-blocking `.github/workflows/flutter-android.yaml` workflow and integrate APK building/publishing into `.releaserc.yaml` and `.github/workflows/release.yaml`.

### Non-goals

- Replacing the desktop DM web workspace.
- Blocking backend or Docker releases if the Android build fails.

## Capabilities

### New Capabilities

- `flutter-android-app`: Native Flutter Android TTRPG companion for Villum character sheets, combat/initiative, battlemaps, compendium search, and AI drafting.
- `android-ci-release`: Non-blocking Flutter Android CI workflow and semantic-release GitHub Release APK attachment.

### Modified Capabilities

- None.

## Impact

- New `mobile/` Flutter project directory.
- New `.github/workflows/flutter-android.yaml` workflow.
- Updates to `.releaserc.yaml` and `.github/workflows/release.yaml`.
