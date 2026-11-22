# Design: android-spell-item-tracking

## Context

The Android app's character sheet tabs are fed by `Character.fromJson` (mobile/lib/src/models/character.dart), which parses only top-level fields and silently drops every sub-resource the server already sends (spellcasting, spells, inventory, currency, features, classes — server payload per mobile/ models/character.go and handlers/characters_load.go). Result: SPELLS always reads "No spellcasting", INVENTORY and FEATURES are always empty, and the currency edit dialog pre-fills zeros so saving it overwrites real coins — a data-loss path. Adjacent correctness defects compound this: sessions never restore at boot, login fires twice, token provisioning failures are silent, the compendium parser crashes on the server's integer ids, and the attack roller fabricates results on API errors using an implicit first-character attacker. The server contract (mobile/README.md) already covers every endpoint this change needs; no server work is required. The recent audit (2026-11) root-caused each defect with file:line evidence.

## Goals

- Spell tracking that reflects real server state: slots (used/total), prepared and known spells, casting via real spell ids.
- Item + currency tracking: render, add/edit/equip/delete, and edits that can never zero out untouched coins.
- Correctness spine: boot session restore, single login, visible token failures, compendium integer ids, roller with chosen attacker and server-only results, removal of fabricated sheet values.

## Non-Goals

- Phase C parity features (dice view, settings/logout screen, party view, one-shots, world, i18n) — deferred to a future change.
- Server changes — the frozen contract already supports everything.
- The Android CI `continue-on-error` masking and the integration_test APK build poisoning — separate small change; documented in the Android analysis.
- Full canEdit role gating across every mutation — surfaced errors now; complete gating deferred with the parity work.

## Decisions

- **Parse sub-resources in `Character.fromJson` (single payload).** The server already returns the full nested payload on `GET /api/characters/:id`; parsing it in the existing model keeps one source of truth and zero extra round-trips. Alternative: per-tab lazy fetches — rejected (more requests, more failure states, no offline benefit since the app is online-first).
- **Cast spells with real ids from the parsed spells list.** The current unreachable cast path sends `spell_id: 0`, which 404s (handlers/spells_cast.go). Wiring the parsed spell ids fixes it without server changes.
- **Currency edit pre-fills server values and submits the complete map.** Prevents the zero-overwrite data loss in the currency dialog. The coins section becomes display-from-server, edit-what-you-see.
- **Boot restore in `AuthController.build` + single login path.** Read secure storage once at bootstrap, validate, enter authenticated state; the double repo.login/ctrl.login path is collapsed to one call. Restores failing validation fall back to onboarding — a visible state, not an error swallowed into "unauthenticated".
- **Token provisioning: bounded retry + surfaced errors.** `ensureApiToken` (api_client.dart) currently ignores non-2xx; it gets a small retry with an actionable error, ending the silent-401 cascade on every mutation.
- **Compendium ids as integers.** The server sends int64 ids (handlers/compendium.go); the model parses int; detail navigation stays in-app.
- **Roller targeting + truthfulness.** Attacker defaults to the open character (not `characters.first`), user-selectable; results come from the server response only; errors surface; hardcoded proficiency booleans, the literal "Proficiency +2", and the dead condition chip are removed in favor of server data or explicit empty states.
- **Test strategy.** The 94 existing tests stay green; new tests cover each capability (fromJson sub-resources, spell cast flow, inventory CRUD, currency preservation, boot restore, single login, token failure surfacing, compendium integer ids, roller error path); the two tolerance tests that currently encode the bugs (character_sheet_test's missing-spellcasting expectation, compendium_test's string-id fake) are updated to the corrected behavior.

## Risks / Trade-offs

- [Currency edit completeness] A save still submits the full map; if the sheet stale-loads between open and save, concurrent changes could be lost. → Mitigation: refetch character before opening the dialog; server values are pre-filled, so the common path is correct; full conflict resolution is out of scope.
- [Cast-spell contract assumptions] The README freezes the cast-spell endpoint; if the server rejects a slot-level/identifier combination the app must surface it. → Mitigation: errors from cast attempts are displayed, never swallowed; no local fabrication.
- [Session restore failure modes] Corrupt storage or a revoked session could trap a user between states. → Mitigation: restore failures fall back to onboarding with the stored server URL pre-filled; one re-login recovers.
- [Scope creep] "and all" (user's request) could balloon into full parity. → Mitigation: this change is scoped to tracking + the correctness spine; Phase C parity is a separate proposal the user explicitly deferred.
