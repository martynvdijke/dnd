# OIDC rollout runbook (Authelia)

Operator steps for the `add-authelia-oidc` change. The code for sections 1-4 of that
change is already shipped and merged; what remains are the environment-specific steps in
its section 5, which need the live Authelia instance, Nginx Proxy Manager and the
production host.

This runbook is written so the steps can be executed and the outcome judged without
reading the Go code. Config reference: see the "OIDC SSO (Authelia)" section in the
[project README](../README.md).

## Order matters

Do **5.1 then 5.2 before enabling anything**, then 5.3 and 5.5. In particular the Nginx
Proxy Manager forward-auth bypass must be in place *before* users start logging in:
forward-auth intercepts `/api/auth/oidc/callback` and turns the authorization-code
exchange into a redirect loop.

Do not remove password login as part of this rollout. Password login stays enabled as the
fallback until OIDC has been verified end to end; removing it is a follow-up change
(`retire-password-fallback`).

## 1. Register the Authelia client (task 5.1)

The effective client settings:

| Setting | Value |
| --- | --- |
| `client_id` | `dnd` |
| `redirect_uri` | `https://dnd.vandijke.xyz/api/auth/oidc/callback` |
| scopes | `openid email profile groups` |
| PKCE | S256 required |
| token endpoint auth | `client_secret_post` (the app also accepts `client_secret_basic`) |
| consent | explicit |
| policy | `one_factor` |

Shape of the Authelia configuration (field names follow your Authelia version):

```yaml
identity_providers:
  oidc:
    clients:
      - id: dnd
        description: Villum
        secret: '<client secret>'
        redirect_uris:
          - https://dnd.vandijke.xyz/api/auth/oidc/callback
        scopes:
          - openid
          - email
          - profile
          - groups
        grant_types:
          - authorization_code
        response_types:
          - code
        require_pkce: true
        token_endpoint_auth_method: client_secret_post
        consent_mode: explicit
        authorization_policy: one_factor
```

Keep the secret out of git. The app reads `OIDC_CLIENT_SECRET_FILE` first (a mounted file,
preferred) and only falls back to the `OIDC_CLIENT_SECRET` environment variable.

## 2. Bypass forward-auth in Nginx Proxy Manager (task 5.2)

In the NPM proxy host for `dnd.vandijke.xyz`, disable forward-auth for the host, or at
minimum exempt `/api/auth/oidc/`.

Verify the bypass took effect by confirming the login endpoint redirects to Authelia and
not to NPM's own login page:

```bash
curl -sI https://dnd.vandijke.xyz/api/auth/oidc/login | grep -i '^location'
```

Expected: a `location` pointing at the Authelia authorize endpoint. If you are sent to the
NPM login page instead, the bypass is not active.

## 3. Configure and enable (task 5.4)

`OIDC_ENABLED` defaults to `true`, but OIDC only becomes active once the configuration is
complete - the login button appears only when every required value is present. A
minimal `.env` for the compose stack:

```dotenv
OIDC_ENABLED=true
OIDC_ISSUER_URL=https://authelia.vandijke.xyz
OIDC_CLIENT_ID=dnd
OIDC_CLIENT_SECRET_FILE=/run/secrets/oidc_client_secret
OIDC_REDIRECT_URL=https://dnd.vandijke.xyz/api/auth/oidc/callback
OIDC_SCOPES=openid email profile groups
```

Then apply it and confirm the app agrees the configuration is complete:

```bash
docker compose pull && docker compose up -d
curl -s https://dnd.vandijke.xyz/api/auth/oidc/status
```

## 4. Verify (tasks 5.3 and 5.5)

Public routes must stay anonymous - these are expected to return `200` without any
session:

```bash
for path in / /healthz /api/share/<token>; do
  printf '%s -> ' "$path"
  curl -s -o /dev/null -w '%{http_code}\n' "https://dnd.vandijke.xyz$path"
done
```

Private routes must not be reachable anonymously. The deployment answers `401` for API
reads, API mutations *and* the HTMX web endpoints, so there is no transparent `302` to
the OIDC login - the client-side code reacts to the `401` by sending the browser to
`GET /api/auth/oidc/login`.

Then verify the interactive flows in a browser:

1. `GET /api/auth/oidc/login` redirects to Authelia; authenticating returns you to Villum
   with a session. The account is linked by its **verified** email on first SSO login, and
   membership of the `admins` group grants the admin role.
2. Existing API tokens and sessions keep working - the rollout does not invalidate them.
3. `POST /api/logout` clears the session cookie and redirects to the Authelia logout URL.

## Verification recorded against the live deployment

Probed 2026-10-17 against `https://dnd.vandijke.xyz`, then running **v2.103.0**, with
`GET /api/auth/oidc/status` reporting `{"enabled":true}`. All probes below are
anonymous - no credentials were used - so the interactive login still has to be
finished by hand (task 5.3).

| Probe | Result | Meaning |
| --- | --- | --- |
| `GET /` | `200` | public read stays anonymous |
| `GET /healthz` | `200` (`version: 2.103.0`) | public read stays anonymous |
| `GET /api/share/<invalid-token>` | `404` | the share route is reachable anonymously (bad token, not `401`) |
| `GET /api/user/me`, `/api/characters`, `/api/campaigns`, `/api/tokens` | `401` | private reads need a session |
| `GET /htmx/media-gallery?...` | `401` | private web endpoints answer `401`, not a transparent `302` |
| `POST /api/characters`, `/api/telegram/media`, `/api/tokens` | `401` | private mutations are refused anonymously |
| `POST /api/login` with bad credentials | `401 {"error":"invalid credentials"}` | **password login is still enabled** |
| `GET /login` | `200`, contains "Login with Authelia" | the SSO entry point is rendered |

The handshake is correct as far as Authelia. `GET /api/auth/oidc/login` returns `302`
to

```
https://authelia.vandijke.xyz/api/oidc/authorization?client_id=dnd
  &code_challenge=...&code_challenge_method=S256&nonce=...
  &redirect_uri=https%3A%2F%2Fdnd.vandijke.xyz%2Fapi%2Fauth%2Foidc%2Fcallback
  &response_type=code&scope=openid+email+profile+groups&state=...
```

and sets short-lived `oidc_state`, `oidc_nonce` and `oidc_verifier` cookies
(`Max-Age=300`, `SameSite=Lax`). Requesting that authorize URL answers `303` to
`https://authelia.vandijke.xyz/?flow=openid_connect&flow_id=...`, so **Authelia
accepts the registered client**. `GET /api/auth/oidc/callback` without a code
redirects to `/login?error=oidc_expired`, and `GET /api/auth/oidc/logout` redirects to
`https://authelia.vandijke.xyz/logout`.

Authelia's discovery document at
`https://authelia.vandijke.xyz/.well-known/openid-configuration` matches the registered
values: `issuer https://authelia.vandijke.xyz`,
`code_challenge_methods_supported: ["S256"]`, and
`token_endpoint_auth_methods_supported` including `client_secret_basic` and
`client_secret_post`.

The interactive login was confirmed by the operator on 2026-10-17: signing in through
Authelia completes and returns to Villum with a session. The client's `consent` and
`policy: one_factor` settings live in the Authelia portal and cannot be observed from
outside.

## Rollback

Set `OIDC_ENABLED=false` and restart the stack:

```bash
# in .env
OIDC_ENABLED=false
docker compose up -d
```

Password login returns immediately, and the Authelia button disappears because the
configuration is no longer complete. Existing sessions are unaffected: rolling the toggle
does not invalidate any session, so nobody is logged out by the rollback.
