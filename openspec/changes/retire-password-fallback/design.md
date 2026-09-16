## Context

`add-authelia-oidc` introduced OIDC login against Authelia at `https://authelia.vandijke.xyz` as an additive capability, keeping password/session login ON until OIDC is verified end-to-end. That change shipped code (handlers, middleware, session columns `oidc_sub`/`auth_method`, `OIDC_ENABLED=false` default, `OIDC_CLIENT_SECRET_FILE`, groups→admin sync) but its operator verification tasks 5.1–5.5 remain unchecked — OIDC is not yet proven in production. The intended end state — OIDC as the primary and only default login with password retired — has no change capturing defaults, session cutover, UI, or rollback.

Stakeholders: DND (Villum) operators (homelab), end users who currently use password login. Constraint: TRMNL/API token auth is out of scope and unaffected.

## Goals / Non-Goals

**Goals:**
- Close the auth loop opened by `add-authelia-oidc` by retiring password as the default.
- Make OIDC the default (`OIDC_ENABLED=true` by default) once verified.
- Keep a documented, tested break-glass override for password recovery.
- Expire/migrate password sessions on cutover so users re-auth via Authelia.
- Update login UI to reflect the new default; document and test rollback.

**Non-Goals:**
- Re-specifying the OIDC protocol itself (Authorization Code + PKCE S256, JWKS/RS256, `state`/`nonce`, email-verified linking, groups→admin) — that lives in `oidc-authentication` inside `add-authelia-oidc`.
- Deleting the password code in one step (code may be removed in a later clean-up once break-glass is proven unnecessary).
- Removing or changing API token auth (TRMNL/API bearer tokens remain as-is).
- Gating public reads behind login.

## Decisions

- **Separate subtractive change rather than folding into `add-authelia-oidc`**: Additive-then-subtractive keeps each change independently revertible. Folding retirement into the same change would couple "prove OIDC works" with "remove the fallback that lets you recover if it doesn't" — a single revert would lose both. Two changes: the first is safe to ship (fallback stays), the second flips defaults only after verification. Alternative (single change that adds OIDC and removes password together) was rejected — it raises blast radius and prevents incremental rollout.

- **Break-glass env flag (e.g. `ALLOW_PASSWORD_FALLBACK=true`) instead of immediate code deletion**: Operators need a recovery path if Authelia/issuer is down or misconfigured post-cutover. A flag is the smallest reversible mechanism; deletion can follow in a later change once the flag has never been needed. Alternative (delete password code now) rejected — it removes rollback without operational proof.

- **Defaults flip (`OIDC_ENABLED=true` by default)**: Makes OIDC primary without operator action on new deploys; existing deployments that still need password set the flag explicitly. Alternative (keep `OIDC_ENABLED=false` default forever) rejected — it leaves the intended state opt-in indefinitely.

- **Session cutover: expire/migrate password sessions**: Password sessions (`auth_method=password`) issued before cutover are invalidated so users re-auth via Authelia; OIDC sessions continue. Reuses existing `auth_method` column from `add-authelia-oidc` — no new session store.

- **Login UI conditional on break-glass**: Single code path — Authelia-only by default, password fields only when break-glass is set. Avoids a separate "password disabled" page.

## Risks / Trade-offs

- **Lockout before OIDC is proven** → Mitigation: sequencing dependency — this change MUST NOT ship before `add-authelia-oidc` tasks 5.1–5.5 are verified and the change is archived. Pre-cutover checklist includes live login/logout, NPM bypass, and public/private route checks.
- **Authelia outage post-cutover blocks all logins** → Mitigation: break-glass flag restores password login without code deploy; rollback path is documented and must be tested before cutover.
- **Users with active password sessions confused by sudden re-auth** → Mitigation: session expiry is deliberate and communicated in release notes; OIDC sessions are not affected.
- **Flag sprawl / forgotten break-glass left on** → Mitigation: docs mark break-glass as temporary recovery only; future change can remove password code entirely once flag is observably unused.

## Migration Plan

1. Verify `add-authelia-oidc` operator tasks 5.1–5.5 (Authelia client registration, NPM bypass, login/logout, public/private route checks) and archive that change.
2. Flip defaults (`OIDC_ENABLED=true` by default, password disabled unless `ALLOW_PASSWORD_FALLBACK=true`).
3. Invalidate/migrate existing password-issued sessions.
4. Update login UI to be Authelia-only unless break-glass is set.
5. Document rollback (re-enable break-glass; if needed `OIDC_ENABLED=false`) and test it before declaring cutover complete.
6. Update `AGENTS.md` and operator docs to reflect OIDC-default.

Rollback: set break-glass flag (and if needed `OIDC_ENABLED=false`) and restart; password login resumes with same session cookie shape. No data migration to undo beyond session re-auth.

## Open Questions

- Exact break-glass flag name (`ALLOW_PASSWORD_FALLBACK` vs `PASSWORD_FALLBACK_ENABLED`) — to be finalized in implementation; spec uses example name.
- Session invalidation strategy: bulk expiry of `auth_method=password` rows vs lazy rejection on next request — implementation may choose either as long as password sessions no longer authenticate when break-glass is off.
