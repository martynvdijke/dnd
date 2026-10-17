# fix: Telegram admin character visibility

## Why

The Telegram bot's character queries use `editableCharacterSQL`, a predicate
intended to mirror the web app's `canEditCharacter`. It grants access to a
character's owner, the owner of a campaign the character belongs to, and a
campaign member with the `dm` role — but it omitted the web app's **admin
bypass**. The web `canEditCharacter` returns `true` for administrators first, so
an admin sees every character in `ListCharacters`; the bot, using the narrower
predicate, reports "You have no characters yet. Use /create to make one."

This was observed live: an OIDC login provisioned a duplicate administrator
account (`martynvandijke-2`) that owns no characters, and the Telegram identity
ended up linked to it. The web showed the characters (admin sees all); the bot
showed none.

## What Changes

- Add an administrator bypass to `editableCharacterSQL` in `telegram/claims.go`,
  so the bot's edit predicate matches the web app's `canEditCharacter`.
- This flows through every caller of the predicate: `/characters`, `/claim`,
  `/sheet`, `/stats`, and the character actions (HP, rests, casting), all of
  which already gate on the same predicate.
- Adjust `TestHPRejectNonEditable`, which seeded its "intruder" as an admin and
  therefore relied on the previous (incorrect) absence of the admin bypass.

## Capabilities

### Modified Capabilities
- `telegram-character-commands`: administrators may see, claim, and open the
  sheet/stats of any character, matching the web app; non-administrators keep
  the existing edit-based restriction.

## Impact

- Backend only; no schema or API changes.
- Files: `telegram/claims.go`, `telegram/characters.go`,
  `telegram/telegram_test.go`, `telegram/actions_test.go`.
- No change for non-admin users.
