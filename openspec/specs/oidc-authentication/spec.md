# oidc-authentication Specification

## Purpose
TBD - created by archiving change add-authelia-oidc. Update Purpose after archive.
## Requirements
### Requirement: OIDC login via Authelia

The service SHALL support OIDC Authorization Code flow with PKCE S256 against Authelia issuer https://authelia.vandijke.xyz (scopes `openid email profile groups`, client auth `client_secret_post` or `client_secret_basic`), verifying ID tokens via RS256/JWKS discovery and requiring `state`+`nonce`+`PKCE` checks. Handlers `/api/auth/oidc/login`, `/api/auth/oidc/callback`, `/api/auth/oidc/logout` reuse the existing session cookie.

#### Scenario: Successful OIDC login

- WHEN an unauthenticated user clicks "Login with Authelia" and completes Authelia authentication
- THEN DND (Villum) validates state/nonce/PKCE, verifies the ID token via Authelia JWKS (RS256), creates a session with the same `session` cookie as password auth, and redirects to the app

#### Scenario: Invalid state or nonce rejected

- WHEN the callback presents an invalid or mismatched state/nonce/PKCE verifier
- THEN the service rejects the callback and does not create a session

#### Scenario: Discovery via issuer

- WHEN OIDC is enabled with OIDC_ISSUER_URL=https://authelia.vandijke.xyz
- THEN the service discovers authorization/token/JWKS endpoints from `.well-known/openid-configuration` and validates tokens against that JWKS

### Requirement: User auto-provision/link by email + groups->admin

The service SHALL auto-provision or link users by verified email on OIDC login and sync the admin flag from the `groups` claim containing `admins` (LEDit additionally maps groups->roles per add-user-roles; traces links via family-logins email key).

#### Scenario: Existing user link by email

- WHEN an OIDC login returns a verified email matching an existing admin_users record
- THEN the service links `oidc_sub` to that user and sets `auth_method=oidc`

#### Scenario: First-time user auto-provision

- WHEN an OIDC login returns a verified email with no existing user
- THEN the service creates a new user keyed by email, stores `oidc_sub`, and grants admin only if `groups` contains `admins`

#### Scenario: Unverified email rejected

- WHEN the ID token has `email_verified=false` or no email claim
- THEN the service rejects the login and does not provision a user

#### Scenario: Groups to admin sync

- WHEN a returning OIDC user logs in and the `groups` claim now includes (or no longer includes) `admins`
- THEN the service updates the user's admin flag to match

### Requirement: Logout + session handling

The service SHALL extend AuthMiddleware to accept OIDC sessions (same `session` cookie with `auth_method=oidc|password` + `oidc_sub`) and SHALL implement logout that clears the local session then redirects to Authelia logout.

#### Scenario: Authenticated request via OIDC session

- WHEN a request carries a valid OIDC-issued session cookie
- THEN AuthMiddleware treats it as authenticated (same as password session)

#### Scenario: Logout clears session and redirects

- WHEN an OIDC-authenticated user hits `/api/auth/oidc/logout`
- THEN the service clears the local session cookie and redirects to Authelia `https://authelia.vandijke.xyz/logout` (with post-logout redirect back to the app)

#### Scenario: OIDC session reuse

- WHEN an OIDC session is active
- THEN subsequent API calls using the session cookie succeed without re-authenticating to Authelia until session expiry

### Requirement: Public reads stay anonymous

Public GET routes SHALL remain accessible without authentication; OIDC is for SSO login only, not to gate public reads.

#### Scenario: Anonymous public read succeeds

- WHEN an anonymous user performs GET on a public route (e.g. `/`, `GET /healthz`, `GET /api/share/:token`)
- THEN the service returns 200 without redirect to OIDC login

#### Scenario: Anonymous private mutation requires login

- WHEN an anonymous user hits a private mutation (e.g. POST /api/*, admin pages) without a session
- THEN the service redirects to OIDC login (web 302 to /api/auth/oidc/login) or returns 401 for API clients

### Requirement: OIDC secret handling and break-glass recovery

The service SHALL never commit OIDC client secrets to git; the client secret SHALL be supplied through `OIDC_CLIENT_SECRET_FILE` or an environment variable wired through `docker-compose.yml`. Password login SHALL require the explicit break-glass flag `ALLOW_PASSWORD_FALLBACK`, because `retire-password-fallback` retires password login as the default and keeps it only as a documented recovery path.

#### Scenario: Secret not in repo

- **WHEN** inspecting git history and compose files
- **THEN** no client secret value appears; only OIDC_CLIENT_SECRET_FILE / env references are present

#### Scenario: Password fallback requires break-glass

- **WHEN** an operator needs password login after the cutover
- **THEN** setting `ALLOW_PASSWORD_FALLBACK=true` and restarting the deployment restores password login without a code change

#### Scenario: OIDC disabled toggle

- **WHEN** `OIDC_ENABLED=false`
- **THEN** the OIDC routes report the flow as disabled and password login is available only while the break-glass flag is set
