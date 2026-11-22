# android-session-restore Specification

## Purpose

A signed-in user stays signed in across app restarts, login executes exactly once, and API token provisioning failures are visible and recoverable.

## ADDED Requirements

### Requirement: Session restores at app launch

The app SHALL read the persisted session from secure storage during bootstrap and, when valid, enter the authenticated state without requiring the user to log in again.

#### Scenario: Relaunch keeps the user signed in

- WHEN the app is killed and relaunched with a previously saved session
- THEN the app starts signed in with the stored server URL

#### Scenario: Missing or invalid session falls back to onboarding

- WHEN no session is stored or the stored session fails validation
- THEN the app presents the onboarding screen

### Requirement: Login executes exactly once per user action

A login action SHALL perform a single authentication call; the app SHALL NOT fire duplicate login requests for one submission.

#### Scenario: Password sign-in performs one request

- WHEN the user submits the password sign-in form
- THEN exactly one login request is sent and its outcome (success or error) is displayed

### Requirement: API token provisioning surfaces failures

When API token provisioning fails, the app SHALL show an actionable error and offer retry; it SHALL NOT silently ignore non-success responses leaving mutations to fail with unexplained 401s.

#### Scenario: Token failure shows retry

- WHEN the token endpoint returns an error during a mutation attempt
- THEN the user sees an error message with a retry action instead of a silent failure
