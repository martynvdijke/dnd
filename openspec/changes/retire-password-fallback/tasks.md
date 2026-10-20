## 1. Verify add-authelia-oidc end-to-end (prerequisite)

- [x] 1.1 Re-run and verify `add-authelia-oidc` operator tasks 5.1–5.5: Authelia client registration (client_id dnd, redirect https://dnd.vandijke.xyz/api/auth/oidc/callback, PKCE S256, scopes openid email profile groups), NPM bypass, login/logout, session/token checks, public vs private route checks. — verified against the live deployment 2026-10-17, recorded in `docs/oidc-rollout.md`.
- [x] 1.2 Confirm `add-authelia-oidc` is archived; do not proceed to §2 until archived. — archived as `openspec/changes/archive/2026-10-17-add-authelia-oidc`.

## 2. Flip defaults and break-glass

- [x] 2.1 Change default so `OIDC_ENABLED=true` when not explicitly set. — already the effective default in `handlers/oidc.go` (`enabled()` treats unset as on) and `docker-compose.yml` (`${OIDC_ENABLED:-true}`); verified, no code change needed.
- [x] 2.2 Disable password login by default; gate it behind documented break-glass env flag (e.g. `ALLOW_PASSWORD_FALLBACK=true`), wire through config and `docker-compose.yml` / `.env.example`. — `middleware/authpolicy.go` (`PasswordFallbackEnabled`, `PasswordFallbackEnvVar`), `handlers/auth.go` 403 gate before both login branches, `docker-compose.yml`, `README.md`, `docs/oidc-rollout.md`. No tracked `.env.example` exists.
- [x] 2.3 Ensure break-glass restores password login alongside OIDC for recovery (no code deploy needed beyond flag + restart). — flag is read per request, so a restart flips it without a rebuild; e2e phase 3 in `tests/password-retirement.spec.ts` demonstrates the restore.

## 3. Session cutover

- [x] 3.1 Expire/migrate existing password-issued sessions (`auth_method=password`) so users must re-authenticate via Authelia; verify OIDC sessions (`auth_method=oidc`) remain valid. — lazy rejection in `middleware.AuthRequired`: the row is deleted and the request answered with the standard expired-session 401 while break-glass is off; OIDC sessions pass through. Covered by `middleware/auth_test.go` and e2e phase 2.

## 4. Login UI

- [x] 4.1 Update login page to show Authelia as the only option by default; show password fields only when break-glass is enabled. — `GET /api/auth/oidc/status` now reports `password_enabled`; `ts/lib/auth-status.ts` + `applyAuthPolicy()` in `ts/login.ts` render OIDC-only, break-glass, or the "no way to sign in" notice. `static/setup.html`, `static/reset-password.html` and `ts/setup.ts` / `ts/reset.ts` follow the same policy, and `GET /reset-password` redirects to `/login` while retired.
- [x] 4.2 Add/adjust unit and e2e coverage for OIDC-only vs break-glass UI states. — `ts/login.test.ts`, `ts/lib/auth-status.test.ts`, `ts/setup.test.ts`, `ts/reset.test.ts`, `handlers/auth_login_test.go`, `handlers/password_reset_test.go`, `tests/password-retirement.spec.ts`, `tests/oidc-login.spec.ts`.

## 5. Rollback and docs

- [x] 5.1 Document rollback path (enable break-glass flag; if needed `OIDC_ENABLED=false`) in operator docs and release notes. — `docs/oidc-rollout.md` "Rollback" section, `README.md` recovery paragraph, `AGENTS.md` Authentication section.
- [x] 5.2 Test rollback before cutover: demonstrate that enabling break-glass restores password login. — e2e phase 3 restarts the same server/DB with the flag back on and logs in again.
- [x] 5.3 Update `AGENTS.md` and docs to reflect OIDC as the default login and break-glass as recovery only; note that API/TRMNL token auth is unaffected. — done in `AGENTS.md`, `README.md`, `docs/oidc-rollout.md`.

## 6. Validation

- [x] 6.1 Run `openspec validate retire-password-fallback` and fix until passing. — `openspec validate retire-password-fallback --strict` reports the change is valid.
- [x] 6.2 Run `task ci` (or at minimum `task test` + relevant e2e) to verify no regressions. — `task ci` exited 0 (go tests, vitest 487, Playwright chromium incl. `password-retirement` and `oidc-login`, docker smoke).
