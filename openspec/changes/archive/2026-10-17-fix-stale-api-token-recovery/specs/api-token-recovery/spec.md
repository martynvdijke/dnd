# api-token-recovery

## ADDED Requirements

### Requirement: Transparent recovery from an invalidated API token

When a mutating request fails with `401` while carrying an API token, the client MUST
drop the stored token, provision a replacement, and retry the request once. Read
requests, requests without a token, and non-401 failures MUST NOT trigger recovery.

#### Scenario: A mutation with a stale token self-heals
- **GIVEN** the client holds an API token that is no longer valid for the session
- **WHEN** a `POST`/`PUT`/`DELETE` request returns `401`
- **THEN** the client re-provisions a token and retries the request once
- **AND** the retried request is sent with the new token

#### Scenario: Read requests never trigger recovery
- **GIVEN** the client holds an API token
- **WHEN** a `GET` request returns `401`
- **THEN** the error is surfaced to the caller and no token is re-provisioned

#### Scenario: Recovery failure keeps the original error
- **GIVEN** the client holds an invalid API token
- **WHEN** re-provisioning does not yield a working token
- **THEN** the original `401` error is surfaced after the single retry

### Requirement: Provisioning must not invalidate other devices

Bootstrap token provisioning MUST create a new token rather than rotate an
existing one, so that provisioning on one device cannot invalidate the token
stored on another.

#### Scenario: Two devices bootstrap independently
- **GIVEN** no token is stored locally
- **WHEN** the app bootstraps
- **THEN** it creates a new token and stores its secret
- **AND** any token already stored on other devices remains valid
