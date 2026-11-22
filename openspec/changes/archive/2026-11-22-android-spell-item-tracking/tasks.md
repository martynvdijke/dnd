# Tasks: android-spell-item-tracking

Spell tracking, item tracking, and correctness fixes for the Flutter Android app. No server-side changes; the server contract is frozen per `mobile/README.md`.

## 1. Character data foundation

- [x] 1.1 Parse all sub-resources in `Character.fromJson` (spellcasting, spells, inventory, currency, features, classes) into typed model fields. — `mobile/lib/src/models/character.dart:101-135`
- [x] 1.2 Currency dialog pre-fills all coin fields from server state and submits the complete currency map; editing one coin never zeroes the others. — `character_sheet_screen.dart:918-985`
- [x] 1.3 Token provisioning retries once and raises a visible, actionable error on non-2xx instead of silently leaving every mutation to fail. — `mobile/lib/src/data/api_client.dart:74-94`

## 2. Spell tracking

- [x] 2.1 SPELLS tab renders real spellcasting from the parsed model: slot rows with used/total per level, prepared spells, known spells; characters without spellcasting show an accurate empty state. — `mobile/lib/features/character_sheet/character_sheet_screen.dart` (`_SpellsTab`, `_SlotRow`)
- [x] 2.2 Cast a spell posts `/api/characters/:id/cast-spell` with the real parsed spell id and chosen slot level; success reflects the server response, failure surfaces the error. — `character_sheet_screen.dart` (`_SpellsTab._cast`, `_castAtSlotLevel`, `_chooseSlotLevel`)
- [x] 2.3 Death saves and exhaustion wire to their existing endpoints and persist real state. — `mobile/lib/data/api/character_api.dart` (`deathSave`, `setExhaustion`), sheet COMBAT tab

## 3. Item tracking

- [x] 3.1 INVENTORY tab renders server items (name, quantity, equipped) and server currency from the parsed model. — `character_sheet_screen.dart` (`_InventoryTab`)
- [x] 3.2 Replace the add-item snackbar stub with full CRUD: add, edit, equip/unequip, delete. — `character_sheet_screen.dart` (`_editItem`, `_setEquipped`), `character_api.dart` (inventory methods)

## 4. Correctness

- [x] 4.1 Boot session restore: read the persisted session from secure storage at startup and enter the authenticated state without re-login; onboarding remains the failure fallback. — `mobile/lib/src/providers.dart:47-49`
- [x] 4.2 Onboarding performs exactly one login call per user action (remove the repo.login + controller.login double-fire). — `onboarding_screen.dart:210-233`
- [x] 4.3 Compendium result ids parse as server int64; real results no longer crash; update the string-id test fake. — `mobile/lib/src/models/compendium.dart:22-24`, `compendium_test.dart:22`
- [x] 4.4 Attack roller uses a selected attacker (default the current character, not `characters.first`), posts to `/api/combat/attack`, renders only server-returned outcomes, and surfaces API errors. — `combat_screen.dart:298-355`
- [x] 4.5 Sheet truthfulness: remove hardcoded proficiency booleans, the literal "Proficiency +2", and the dead condition chip; displayed values come from server data. — `character_sheet_screen.dart` (`_CoreTab` save rows + subtitle, COMBAT conditions)

## 5. Tests and verification

- [x] 5.1 Update tolerance tests: the sheet test no longer accepts missing spellcasting; the compendium fake uses int ids. — `character_sheet_test.dart:170`, `compendium_test.dart:22`
- [x] 5.2 Add tests per capability: sub-resource parsing, spell cast flow, inventory CRUD, currency no-overwrite, boot restore, single login, compendium int ids, roller targeting.
- [x] 5.3 `cd mobile && flutter analyze` reports no issues and `flutter test` is green (all 94 existing tests plus new). — 121/121 green; analyze clean
- [x] 5.4 `openspec validate --change android-spell-item-tracking` passes; `task ci` green before push; PR checks green before merge.
