## ADDED Requirements

### Requirement: OIDC is the default login

Once `add-authelia-oidc` is verified and archived, the service SHALL default to OIDC login with `OIDC_ENABLED=true` when no explicit value is configured.

#### Scenario: Default OIDC on

- WHEN the service starts with no explicit `OIDC_ENABLED` value after cutover
- THEN OIDC login is enabled and `/api/auth/oidc/login` is available

### Requirement: Password login disabled by default behind break-glass override

The service SHALL disable password login by default and SHALL only re-enable it when an explicit documented break-glass environment flag is set (e.g. `ALLOW_PASSWORD_FALLBACK=true`).

#### Scenario: Password disabled by default

- WHEN the break-glass flag is not set (or set to false) after cutover
- THEN password login endpoints reject requests or are hidden and no password session can be created

#### Scenario: Break-glass re-enables password

- WHEN the operator sets the documented break-glass flag to true
- THEN password login is re-enabled alongside OIDC for recovery

### Requirement: Password-issued sessions expire on cutover

On cutover the service SHALL expire or migrate existing password-issued sessions so that affected users MUST re-authenticate via Authelia; OIDC-issued sessions SHALL remain valid.

#### Scenario: Password sessions invalidated

- WHEN a user presents a session cookie with `auth_method=password` created before cutover and the break-glass flag is not set
- THEN the service treats the session as invalid and requires OIDC re-authentication

#### Scenario: OIDC sessions unaffected

- WHEN a user presents a valid OIDC-issued session (`auth_method=oidc`)
- THEN the service accepts it normally across cutover

### Requirement: Login UI presents Authelia only unless break-glass is set

The login UI SHALL present Authelia/OIDC as the only login option by default and SHALL only show password fields when the break-glass override is enabled.

#### Scenario: Default login UI is OIDC only

- WHEN the break-glass flag is not set
- THEN the login page shows "Login with Authelia" and does not show password inputs

#### Scenario: Break-glass login UI shows password fallback

- WHEN the break-glass flag is set to true
- THEN the login page shows both Authelia and password login options

### Requirement: Rollback path is documented and tested before cutover

The service SHALL document the rollback path for restoring password login (setting the break-glass flag and, if needed, `OIDC_ENABLED=false`) and the rollback SHALL be tested successfully before cutover is considered complete.

#### Scenario: Rollback documented

- WHEN an operator reads the runbook/docs for this change
- THEN the steps to re-enable password login and disable OIDC are explicitly documented

#### Scenario: Rollback tested before cutover

- WHEN cutover verification is performed
- THEN the operator has demonstrated that enabling the break-glass flag restores password login

### Requirement: Sequencing dependency on add-authelia-oidc

This change SHALL NOT ship (defaults SHALL NOT flip) before `add-authelia-oidc` is verified end-to-end (its operator tasks 5.1–5.5) and archived.

#### Scenario: Blocked until OIDC verified

- WHEN `add-authelia-oidc` tasks 5.1–5.5 have not been verified and the change has not been archived
- THEN this change is considered not ready to deploy and the default flip is not applied

#### Scenario: Ready after OIDC archived

- WHEN `add-authelia-oidc` is verified and archived
- THEN this change may proceed to flip defaults
