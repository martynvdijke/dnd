## 1. Inventory

- [x] 1.1 `telegram/tracking.go`: load the claimed character's inventory ordered by category then name, render it with quantity/equipped/attunement markers, and attach inline controls
- [x] 1.2 Callback actions equip/unequip an item (parameterized `UPDATE inventory SET is_equipped = ... WHERE id = ? AND character_id = ?`)
- [x] 1.3 Quantity `+1`/`−1` buttons with a floor of zero
- [x] 1.4 Remove an item behind a confirm step (`inv:rm:<id>` → confirm/cancel)
- [x] 1.5 `/additem <name> [qty]` inserts a `gear` item defaulting quantity to 1; missing name is refused with usage

## 2. Spells

- [x] 2.1 `/spells` (alias `/spellbook`) lists the claimed character's spells grouped by level with prepared markers
- [x] 2.2 Prepare/unprepare buttons for spells with level > 0 that are not `always_prepared`; cantrips and always-prepared spells render without toggles
- [x] 2.3 `/prepare <spell>` and `/unprepare <spell>` set the `prepared` flag by name for the claimed character; unknown spell is refused

## 3. Conditions, features and currency

- [x] 3.1 `/conditions` lists active conditions with remove controls; `/condition <name>` inserts a trimmed, non-empty, length-capped name
- [x] 3.2 `/features` lists the claimed character's features read-only
- [x] 3.3 `/money` shows the currency row; `/money +N|-N <pp|gp|ep|sp|cp>` adjusts one denomination clamped at zero, creating the row when absent

## 4. Commands, callbacks and routing

- [x] 4.1 Register `/inventory` (`/inv`, `/bag`), `/additem`, `/spells` (`/spellbook`), `/prepare`, `/unprepare`, `/conditions`, `/condition`, `/features` and `/money` in the command registry with usage and descriptions
- [x] 4.2 Add `cbInvPrefix`, `cbSpellPrefix`, `cbCondPrefix` and the remove-confirm action to `keyboard.go` and route them in `handleCallbackData`
- [x] 4.3 Every command and callback resolves the character through `requireActionCharacter` and re-validates edit rights on each action; all SQL is parameterized
- [x] 4.4 Point the `/sheet` other-content section at the tracking commands

## 5. Tests

- [x] 5.1 `telegram` unit tests: inventory listing, equip/unequip, quantity floor and increment, remove-confirm, add-item usage, spellbook listing, prepare/unprepare (including cantrip/always-prepared exclusions), unknown-spell refusal, condition add/remove, feature listing, currency show/add/clamp/invalid-denomination
- [x] 5.2 Authorization tests: unclaimed user guided and non-owned character refused for both listing and mutation
- [x] 5.3 Coverage floors (Go total ≥20%, `handlers/` ≥40%, `middleware/` ≥40%) still met

## 6. Verification and delivery

- [x] 6.1 `openspec validate telegram-player-tracking`
- [x] 6.2 `task ci` green (Go + vitest + e2e + coverage gates); PR opened and merged per the workflow contract
- [x] 6.3 Post-merge: archive the change in a follow-up PR (`openspec archive telegram-player-tracking`, `git add -f`), matching the repo convention
