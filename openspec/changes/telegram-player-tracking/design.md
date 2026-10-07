## Context

The bot already reads character data: `telegram/characters.go` builds summaries and renders the sheet/stats, and its "Other" section prints counts of `character_conditions`, `character_features`, `spells` and `inventory` plus currency. It already mutates character state too — `telegram/actions.go` (`runHP`, `runRest`, `runCast`) goes through `requireActionCharacter(c)`, which resolves the claim and re-checks edit rights via `editableCharacterSQL` (`admin OR owner OR campaign owner OR campaign dm`). Every one of those tables is already present in the schema (`inventory`, `spells`, `character_spellcasting`, `character_conditions`, `character_features`, `character_currency`), so tracking needs no migration.

What is missing is a command surface and inline controls for listing and changing those rows, plus callback routing. The codebase already has a callback pattern in `telegram/keyboard.go` (`handleCallbackData` dispatching on prefixes like `claim:`, `sheet:`, `campaign:`) and a shared reply helper (`sendReply`, which chunks at 4096 characters).

## Goals / Non-Goals

**Goals:**

- List and manage the claimed character's inventory, spellbook preparation, conditions, features and currency from the bot.
- Make every mutation authorization-checked against the character and parameterized.
- Keep replies readable on small screens and resilient to large data sets.
- Reuse existing helpers and callback patterns rather than introduce a new subsystem.

**Non-Goals:**

- Editing item stats, spell definitions, or feature text.
- Enforcing attunement limits or class/level spell-preparation legality beyond the existing `prepared`/`always_prepared` flags.
- Multiclass spell slot bookkeeping beyond what `runCast` already does.
- Tracking party-level items (`party_items` already has the `/items` campaign command) or quests/journal.

## Decisions

1. **Authorization reuses `requireActionCharacter`.** Every tracking command resolves the claimed character and refuses with the same "not found"/claim-unavailable behaviour as `/hp` and `/cast`. Inline callbacks carry the character ID and re-validate it, so a stale button cannot act on a character the user no longer owns. Alternative: a new authz helper — rejected; one path is safer and already tested.

2. **Inline callbacks carry action + row id, dispatched by new prefixes.** `keyboard.go` gains `cbInvPrefix` (`inv:`), `cbSpellPrefix` (`spl:`), `cbCondPrefix` (`cond:`) and a remove-confirm action, all routed in the existing `handleCallbackData` switch. Toggles re-render the list in place via the reply keyboard. Callback data stays well under Telegram's 64-byte limit because it holds only the action and a numeric id.

3. **Quantity is ±1 buttons; removal is confirmed.** Equip/unequip and prepare/unprepare are simple toggles. Quantity uses `−1`/`+1` buttons bounded at zero (removing at zero is refused, use remove). Removal is a two-step confirm (`inv:rm:<id>` → confirm/cancel) because it is destructive and not undoable. Alternative: a typed-argument interface — rejected as clunky compared with buttons.

4. **Spells with levels 0 or `always_prepared` are not toggleable.** Cantrips never need preparation and always-prepared spells (domain/oath) cannot be unprepared, so the list shows them with a fixed marker and no buttons. The `character_spellcasting` table is read only for display; preparation mutates `spells.prepared`.

5. **Conditions are free text, trimmed, capped in length.** `/condition <name>` inserts a trimmed name (rejecting empty and over-long input); `/conditions` lists with remove buttons. No canonical condition list is enforced — D&D conditions include arbitrary named effects in this tool.

6. **Currency is clamped in the write.** `/money` renders the `character_currency` row; `/money +N|-N <denom>` computes the new amount and clamps at zero in the same statement/transaction, creating the row if absent. Alternative: clamp in Go — equivalent, but doing it in the mutation keeps it atomic with the write.

7. **Large lists and message limits.** Replies reuse `sendReply` chunking. Inventory and spellbook are ordered deterministically (category then name; spell level then name) and capped per page where necessary, so a large sheet stays usable.

8. **Sheet cross-reference.** The "Other" section of `renderSheet` names the tracking commands so `/sheet` remains the discovery point, satisfying the modified `telegram-character-commands` requirement.

## Risks / Trade-offs

- **Callback races / stale buttons.** A button may reference a row that was removed or a character the user lost rights to. Mitigated by re-resolving and re-authorizing on every callback and replying gracefully when the row is gone.
- **Message size for large inventories.** Mitigated by deterministic ordering, `sendReply` chunking, and a per-message cap; users with hundreds of rows may still see multiple messages.
- **Callback-data length** is a hard 64 bytes; numeric-only payloads keep it safe, but future expansions must not embed names.
- **Destructive removal** could be tapped by mistake; mitigated by the confirm step.
