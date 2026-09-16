## Why

OIDC login via Authelia was added additively in `add-authelia-oidc` but password login was deliberately left ON until OIDC is verified end-to-end (operator tasks 5.1–5.5 remain unchecked). Without a closing change the auth loop never converges: OIDC ships as an unverified secondary path and the intended end state — OIDC as the sole primary login with password retired — has no captured scope, defaults, or cutover plan.

## What Changes

- Make OIDC the default login: `OIDC_ENABLED=true` by default once `add-authelia-oidc` is verified and archived.
- Disable password login by default; retain it behind an explicit documented break-glass env flag (e.g. `ALLOW_PASSWORD_FALLBACK=true`) for recovery only.
- Expire/migrate existing password-issued sessions on cutover so affected users re-authenticate via Authelia.
- Update login UI to present Authelia as the only option unless the break-glass override is set.
- Document the rollback path (re-enable break-glass / set `OIDC_ENABLED=false`) and require it be tested before cutover.
- State sequencing dependency: this change SHALL NOT ship before `add-authelia-oidc` is verified end-to-end and archived.
- **BREAKING**: Default login flips from password to OIDC; deployments without Authelia must opt in via break-glass until migrated.

## Capabilities

### New Capabilities

<!-- No new capabilities — OIDC protocol lives in `oidc-authentication` (add-authelia-oidc). -->

### Modified Capabilities

- `api-authentication`: Retire password login as the default; define OIDC-default, break-glass override, session cutover, login UI, and rollback/sequencing requirements.

## Impact

Backend auth config defaults (`OIDC_ENABLED`, break-glass flag), session handling (expiry/migration of password sessions), frontend login page, `docker-compose.yml` / `.env.example`, and operator docs / `AGENTS.md`. No change to TRMNL/API token auth. Revertible via break-glass flag.
