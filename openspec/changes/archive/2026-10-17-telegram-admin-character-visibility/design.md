# Design

## Decisions

- **Add the bypass to the shared predicate, not the call sites.** All six
  queries (`listCharacters`, `loadCharacterSummary`, `resolveCharacterArg`,
  `characterEditable`, `claimedCharacter`, `claimCandidates`) already share
  `editableCharacterSQL`. Adding an `EXISTS (SELECT 1 FROM users u WHERE u.id = ?
  AND u.role = 'admin')` clause fixes every path at once and keeps them
  consistent, at the cost of one extra `?` per call.
- **Mirror the web, not invent a policy.** `handlers.canEditCharacter` checks
  `role == "admin"` first and grants full edit rights. The bot already claims to
  mirror that function in its comment; this change makes the comment true. The
  web admin listing (`ListCharacters`) returns every character, so the bot's
  listing now agrees with what an admin sees in the app.
- **Keep non-admins untouched.** For the normal case — owner, campaign owner, or
  campaign DM — behavior is identical. The spec's "characters the user is
  allowed to see / edit" wording is unchanged for them.

## Alternatives considered

- **Fix the data (merge the duplicate OIDC account).** Rejected by the user;
  the account split is a separate issue and the bot should still behave like the
  web for admins regardless.
- **Special-case admins in each command.** Rejected — duplicates policy and
  risks drift from the web.

## Risks

- Administrators can now edit any character through the bot (HP, rests,
  casting). This is intended and matches the web app, but it is a privilege
  increase for bot users with the admin role.
- `TestHPRejectNonEditable` had to be corrected: it seeded a non-owner as an
  admin, so the new bypass makes that user legitimately editable. The test now
  uses a regular user to represent the intruder, which is what it meant to test.

## Testing

- New `TestAdminSeesAllCharacters` in `telegram/telegram_test.go` covers: admin
  sees and opens another user's character; the owner still sees it; a stranger
  still gets "not found".
- `go test ./telegram/...` green; full `task ci` / prek pre-push before merge.
