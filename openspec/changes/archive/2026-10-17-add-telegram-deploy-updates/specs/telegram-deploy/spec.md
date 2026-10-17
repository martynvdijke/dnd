# telegram-deploy

## ADDED Requirements

### Requirement: Bot reports its running build version

`/status` SHALL include the running application version when one is available,
so operators can tell a stale deployment apart from a broken one.

#### Scenario: Linked user requests status with a known build
- **GIVEN** the application version is set and the sender is linked
- **WHEN** the user sends `/status`
- **THEN** the reply contains `Build: <version>`

#### Scenario: Status without a known build
- **GIVEN** no application version is set
- **WHEN** the user sends `/status`
- **THEN** the reply omits the `Build:` line and still replies with the linked status

### Requirement: Deployments can fetch published releases

`docker-compose.yml` SHALL reference the published image so `docker compose pull`
retrieves a newer release instead of being a no-op.

#### Scenario: Operator updates a deployment
- **GIVEN** a running compose deployment on an older image
- **WHEN** the operator runs `docker compose pull && docker compose up -d`
- **THEN** the running container uses the pulled release image
