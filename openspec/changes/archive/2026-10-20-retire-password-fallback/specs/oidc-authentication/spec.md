## RENAMED Requirements

- FROM: `### Requirement: Password fallback until verified + no secret in git`
- TO: `### Requirement: OIDC secret handling and break-glass recovery`

## MODIFIED Requirements

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
