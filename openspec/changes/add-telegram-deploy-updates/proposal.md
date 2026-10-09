# fix: make bot deployments updatable and version-visible

## Why

Operators reported the Telegram bot silently ignoring `/login`, `/status`, and
claim-button taps while other commands worked. Investigation showed current
`main` cannot do that (every branch of `runLogin`/`runStatus` replies, and
`CallbackQuery` is subscribed), so the running container was a stale build.
Two gaps made that failure mode hard to see and fix:

- `docker-compose.yml` had `build: .` but no `image:`, so `docker compose pull`
  was a no-op and `docker compose up -d` reused whatever image was already on
  the host.
- The bot never reported its own build, so a stale deployment looked identical
  to a broken one.

## What Changes

- `docker-compose.yml`: add `image: martynvandijke/dnd:latest` so
  `docker compose pull && docker compose up -d` fetches the published release
  (`build: .` is retained for local development).
- `README.md`: document the update command and how to verify the running build.
- `telegram`: expose the running build in `/status` (`Build: <version>`),
  injected from `handlers.AppVersion` at startup.
- Add `telegram/status_test.go` covering presence/absence of the build line.

## Capabilities

### Modified Capabilities
- `telegram-bot`: `/status` now reports the running build version.

## Impact

- No API or schema changes.
- `docker-compose.yml` gains an `image:` reference; local builds are unaffected
  (`docker compose up -d --build` still works).
