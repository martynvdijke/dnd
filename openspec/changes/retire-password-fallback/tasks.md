## 1. Verify add-authelia-oidc end-to-end (prerequisite)

- [ ] 1.1 Re-run and verify `add-authelia-oidc` operator tasks 5.1–5.5: Authelia client registration (client_id dnd, redirect https://dnd.vandijke.xyz/api/auth/oidc/callback, PKCE S256, scopes openid email profile groups), NPM bypass, login/logout, session/token checks, public vs private route checks.
- [ ] 1.2 Confirm `add-authelia-oidc` is archived; do not proceed to §2 until archived.

## 2. Flip defaults and break-glass

- [ ] 2.1 Change default so `OIDC_ENABLED=true` when not explicitly set.
- [ ] 2.2 Disable password login by default; gate it behind documented break-glass env flag (e.g. `ALLOW_PASSWORD_FALLBACK=true`), wire through config and `docker-compose.yml` / `.env.example`.
- [ ] 2.3 Ensure break-glass restores password login alongside OIDC for recovery (no code deploy needed beyond flag + restart).

## 3. Session cutover

- [ ] 3.1 Expire/migrate existing password-issued sessions (`auth_method=password`) so users must re-authenticate via Authelia; verify OIDC sessions (`auth_method=oidc`) remain valid.

## 4. Login UI

- [ ] 4.1 Update login page to show Authelia as the only option by default; show password fields only when break-glass is enabled.
- [ ] 4.2 Add/adjust unit and e2e coverage for OIDC-only vs break-glass UI states.

## 5. Rollback and docs

- [ ] 5.1 Document rollback path (enable break-glass flag; if needed `OIDC_ENABLED=false`) in operator docs and release notes.
- [ ] 5.2 Test rollback before cutover: demonstrate that enabling break-glass restores password login.
- [ ] 5.3 Update `AGENTS.md` and docs to reflect OIDC as the default login and break-glass as recovery only; note that API/TRMNL token auth is unaffected.

## 6. Validation

- [ ] 6.1 Run `openspec validate retire-password-fallback` and fix until passing.
- [ ] 6.2 Run `task ci` (or at minimum `task test` + relevant e2e) to verify no regressions.
