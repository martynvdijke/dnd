## MODIFIED Requirements

### Requirement: Steer before drafting

The assistant SHALL ask clarifying questions and offer suggestions while the
draft is not approved, and SHALL return the structured draft only once the DM
approves or asks it to generate. The reply SHALL be a JSON object carrying a
status ("chatting" or "ready"), a user-facing message, and the draft (null
while chatting). The assistant SHALL receive a detailed output contract for the
requested entity type — field rules, allowed values, and size limits — and
SHALL keep the draft within those limits.

#### Scenario: Questions come first
- **WHEN** the DM gives a short opening prompt
- **THEN** the assistant asks clarifying questions and does not emit a final draft

#### Scenario: Approval produces a draft
- **WHEN** the DM approves the plan
- **THEN** the assistant returns status "ready" with a complete structured draft

#### Scenario: Non-JSON replies degrade gracefully
- **WHEN** the model returns prose instead of the expected JSON
- **THEN** the reply is shown as a chat message and no draft is offered

#### Scenario: One-shot contract is included
- **WHEN** a one-shot drafting session starts
- **THEN** the system prompt includes the one-shot field contract, allowed values, and size limits

## ADDED Requirements

### Requirement: Draft budget, timeout and failure reporting

The system SHALL resolve the token budget for draft replies from the endpoint's
`max_tokens` setting, clamped to a supported range, and SHALL use a generous
default when the setting is unset. Draft requests SHALL be allowed a longer
timeout than free-text generation. The system SHALL treat an empty reply or a
`finish_reason` of "length" as a failure, return an actionable error, and SHALL
NOT store the broken assistant turn.

#### Scenario: Endpoint budget is honored
- **WHEN** an enabled text endpoint has `max_tokens` set
- **THEN** draft requests use that value clamped to the supported range

#### Scenario: Default budget when unset
- **WHEN** the endpoint has no `max_tokens` value
- **THEN** draft requests use the default budget

#### Scenario: Empty reply is reported
- **WHEN** the provider returns an empty assistant message
- **THEN** the system responds with an actionable error and stores no assistant turn

#### Scenario: Truncated reply is reported
- **WHEN** the provider stops because it hit the token limit
- **THEN** the system responds with an actionable error mentioning the token budget and leaves the conversation retryable

### Requirement: Revise an existing draft with AI

The system SHALL provide a stateless revision endpoint that accepts an entity
type, the current draft JSON and an instruction, and returns the standard
`{"status","message","draft"}` envelope. The assistant SHALL change only what
the instruction asks for and return the complete revised object, keeping the
rest identical. The endpoint SHALL use the same token budget, timeout and
empty/truncated failure reporting as draft sessions, and SHALL resolve the
first enabled text endpoint when none is named.

#### Scenario: Only the requested part changes
- **WHEN** the DM asks to rename one NPC in the current draft
- **THEN** the returned draft differs from the input only in that NPC

#### Scenario: A revision can ask a question
- **WHEN** the instruction is ambiguous
- **THEN** the response carries status "chatting" with a message and no draft

#### Scenario: Truncated revision is reported
- **WHEN** the provider stops because it hit the token limit
- **THEN** the system responds with an actionable error instead of a broken draft
