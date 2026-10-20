## Context

`add-authelia-oidc` introduced OIDC login against Authelia at `https://authelia.vandijke.xyz` as an additive capability, keeping password/session login ON until OIDC is verified end-to-end. That change shipped code (handlers, middleware, session columns `oidc_sub`/`auth_method`, `OIDC_CLIENT_SECRET_FILE`, groups→admin sync) and already defaults OIDC on — `handlers/oidc.go` treats `OIDC_ENABLED` as enabled when unset and `docker-compose.yml` passes `${OIDC_ENABLED:-true}`, so task 2.1 is verification-only. Its operator verification tasks 5.1–5.5 were then verified against the live deployment on 2026-10-17 (recorded in `docs/oidc-rollout.md`) and the change is archived, so OIDC is proven in production. The intended end state — OIDC as the primary and only default login with password retired — has no change capturing defaults, session cutover, UI, or rollback.

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

- **Break-glass flag name finalized as `ALLOW_PASSWORD_FALLBACK`**: read per request through `middleware.PasswordFallbackEnabled()`, which treats only `true`/`1` (trimmed, case-insensitive) as enabled. This matches the example name in the spec delta, so the spec, `docker-compose.yml` and the code agree on one spelling.

- **One gate covers both login branches**: `HandleLogin` refuses with `403 {"error": "password login is disabled"}` after binding and *before* the setup/first-admin branch and the normal branch. Gating setup too is deliberate — it would otherwise mint a password session that middleware rejects on the very next request. A distinct `403` rather than an indistinguishable `401` leaks nothing (the login page already discloses the retirement) and lets an operator diagnose a misconfigured deployment.

- **Session cutover is lazy rejection in `AuthRequired`**: a bearer of a session whose `auth_method == "password"` is deleted and answered with `401 {"error": "session expired"}` while break-glass is off; OIDC sessions pass through untouched. Rejected alternative: a blanket `DELETE FROM auth_sessions WHERE auth_method='password'` at startup — it cannot be undone by flipping the flag back and would destroy break-glass sessions created after the cutover.

- **The e2e suite runs with break-glass on**: `tests/fixtures.ts` spawns the server with `ALLOW_PASSWORD_FALLBACK=true` because `tests/helpers.ts` `login()` signs in with a password. This keeps CI green without weakening the shipped default, and exercises the recovery path on every run.

- **A fresh deployment with no admin bootstraps through break-glass**: first-run setup creates an email-less admin that OIDC email linking can never claim, so the setup path is gated alongside normal login and the bootstrap procedure is documented as a break-glass case.

## Risks / Trade-offs

- **Lockout before OIDC is proven** → Mitigation: sequencing dependency — this change MUST NOT ship before `add-authelia-oidc` tasks 5.1–5.5 are verified and the change is archived. Pre-cutover checklist includes live login/logout, NPM bypass, and public/private route checks.
- **Authelia outage post-cutover blocks all logins** → Mitigation: break-glass flag restores password login without code deploy; rollback path is documented and must be tested before cutover.
- **Users with active password sessions confused by sudden re-auth** → Mitigation: session expiry is deliberate and communicated in release notes; OIDC sessions are not affected.
- **Flag sprawl / forgotten break-glass left on** → Mitigation: docs mark break-glass as temporary recovery only; future change can remove password code entirely once flag is observably unused.

## Migration Plan

1. Verify `add-authelia-oidc` operator tasks 5.1–5.5 (Authelia client registration, NPM bypass, login/logout, public/private route checks) and archive that change. **Done 2026-10-17**: the tasks were confirmed against the live deployment and the change is archived (`openspec/changes/archive/2026-10-17-add-authelia-oidc`, capability `oidc-authentication`).
2. Flip defaults (`OIDC_ENABLED=true` by default, password disabled unless `ALLOW_PASSWORD_FALLBACK=true`).
3. Invalidate/migrate existing password-issued sessions.
4. Update login UI to be Authelia-only unless break-glass is set.
5. Document rollback (re-enable break-glass; if needed `OIDC_ENABLED=false`) and test it before declaring cutover complete.
6. Update `AGENTS.md` and operator docs to reflect OIDC-default.

Rollback: set break-glass flag (and if needed `OIDC_ENABLED=false`) and restart; password login resumes with same session cookie shape. No data migration to undo beyond session re-auth.

## Resolved Questions

- Break-glass flag name: `ALLOW_PASSWORD_FALLBACK` (see Decisions).
- Session invalidation strategy: lazy rejection in `AuthRequired` (see Decisions), chosen over a bulk startup `DELETE` because flipping the flag back must be able to restore password login.
