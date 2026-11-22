## Why

The Android app drops every character sub-resource the server sends: `Character.fromJson` (mobile/lib/src/models/character.dart) parses none of the spellcasting, spells, inventory, currency, features, or classes payloads, so the Spells tab always shows "No spellcasting", Inventory and Features are always empty, and the currency edit dialog **overwrites real coins with zeros on save** (data loss). The user has asked for spell tracking, item tracking, and general correctness ("Allow speel tracking item tracking en al in android app"), and the audit root-caused these to the parser plus a set of adjacent defects: no session restore at boot, a compendium crash on integer ids, a double login, silent token failures, fabricated combat results, and hardcoded sheet values.

## What Changes

**Phase A — tracking (user's named priority):**
- Parse all character sub-resources in `Character.fromJson` (spellcasting, spells, inventory, currency, features, classes) — server already sends them (models/character.go, handlers/characters_load.go).
- Spells tab: live slots/prepared/known; cast flow wired to `POST /api/characters/{id}/cast-spell` with real spell ids (current cast path is unreachable and would send `spell_id: 0` → 404).
- Inventory tab: full CRUD (list/add/edit/equip/delete) replacing the "Add item" snackbar stub; currency rendered from server data and never overwritten with zeros.
- Death saves and exhaustion wired to their existing (currently dead) API methods.

**Phase B — correctness:**
- Session restore at boot (read secure storage, start authenticated instead of unauthenticated).
- Single login call (remove the double `repo.login` + `ctrl.login`).
- Token provisioning retries and surfaces actionable errors instead of ignoring non-2xx.
- Compendium ids parsed as integers (no crash on real search results; server sends int64).
- Attack roller: attacker is chosen (not always `characters.first`), no fabricated results on API error.
- Sheet truthfulness: remove hardcoded saving-throw proficiency booleans and the literal "Proficiency +2"; remove the dead "Tap to add…" condition chip (or wire it).

**BREAKING:** none — app-internal changes only; the server contract (mobile/README.md frozen interface) already defines every endpoint used.

## Capabilities

### New Capabilities
- `android-spell-tracking`: spell slots/prepared/known tracking and the cast flow in the Android app
- `android-inventory-tracking`: inventory item CRUD and currency tracking in the Android app
- `android-session-restore`: authenticated cold start — session restore, single login, token retry with visible errors

### Modified Capabilities
- `flutter-android-app`: character sheet must render real server sub-resources and never fabricate data (saving throws, "Proficiency +2", condition chip); combat roller must target a chosen attacker and show only real API outcomes; compendium search must parse integer ids without crashing.

## Impact

- **Code:** mobile/lib — models (character, compendium), character sheet screens/tabs, combat screen, compendium feature, core API client (token retry), app providers (boot restore), onboarding (single login). File:line targets are enumerated in tasks.md.
- **Server:** no endpoint changes; mobile/README.md frozen interface already covers spellcasting/cast-spell/inventory/currency/spells/features/classes.
- **Tests:** mobile/test must keep its 94 tests passing; new unit/widget tests per capability; existing tests that tolerate the broken behavior (character_sheet_test missing-spellcasting tolerance, compendium fake string ids) are updated to assert the fixed behavior.
- **Risks:** currency fix is display-only (no data migration); cast-spell must use real spell ids (server 404s on `spell_id: 0`); `canEdit` gating must be respected for player-visible sheets to avoid 403s.
- **Out of scope (deferred Phase C):** dice view, settings/logout screen, party/one-shots/world/i18n parity — to be proposed separately.
