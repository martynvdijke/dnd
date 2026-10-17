## 1. Flutter Project Setup (`mobile/`)

- [x] 1.1 Scaffold Flutter Android project in `mobile/` (`xyz.vandijke.dnd`) with dark tabletop/parchment themes, `flutter_riverpod`, `dio`, `web_socket_channel`, `go_router`, and `flutter_secure_storage`.
- [x] 1.2 Implement server URL onboarding and three-layer API authentication (`/api/check-setup`, `/api/login` incl. first-run admin setup, `/api/tokens` bearer provisioning, CSRF derivation).

## 2. Character Sheet, Combat, Battlemap & DM Tools

- [x] 2.1 Implement Character Sheet screen (`/api/characters/:id`) with HP/temp-HP stepper (`/hp`), Spell Slot tracker (`/spells`, `/spellcasting`), Inventory & Currency (`/inventory`, `/currency`), and Short/Long Rest actions.
- [x] 2.2 Implement Combat & Attack Roller (`/api/combat/attack`) and real-time `/api/ws` initiative tracker (`/api/combat?campaign_id=`).
- [x] 2.3 Implement Pinch-to-Zoom Battlemap (`InteractiveViewer`) syncing tokens via `/api/campaigns/:id/battlemap`, `/api/battlemap-tokens/:id`, and `/api/campaigns/:id/battlemap/sync`.
- [x] 2.4 Implement Compendium Search (`/api/compendium/search`), DM Screen generators (`/api/generate/*`), and AI Draft Assistant (`/api/ai/draft`).
- [x] 2.5 Add widget and unit tests in `mobile/test/`.

## 3. Non-Blocking CI & Release Integration

- [x] 3.1 Create `.github/workflows/flutter-android.yaml` with `continue-on-error: true` and SHA-pinned actions.
- [x] 3.2 Update `.releaserc.yaml` to bump `mobile/pubspec.yaml` and include it in `@semantic-release/git` `assets`.
- [x] 3.3 Update `.github/workflows/release.yaml` with a non-blocking `android` job that uploads `dnd-${VERSION}-android.apk` to the GitHub Release.
- [x] 3.4 Verify workflows pass `actionlint`, `zizmor`, and `pinact run --check`.
