# ai-structured-drafting Specification

## Purpose
TBD - created by archiving change ai-adventure-generation. Update Purpose after archive.
## Requirements
### Requirement: Conversational drafting sessions

The system SHALL provide a resumable, multi-turn AI drafting session for the
supported entity types (one-shot adventure, NPC, location, encounter, faction,
campaign, quest, item). Each session SHALL persist its message history, and
every model call SHALL include the full prior history so the assistant has
conversational memory.

#### Scenario: Start a drafting session
- **WHEN** a DM starts a draft for an entity type with an opening message
- **THEN** a session is created, the assistant replies, and the session id is returned

#### Scenario: Continue a session
- **WHEN** the DM sends a follow-up message to an existing session
- **THEN** the assistant receives the entire prior conversation plus the new message

#### Scenario: Reconnect to a session
- **WHEN** the DM reopens the drafting modal for the same entity type
- **THEN** the prior conversation and any draft are restored

#### Scenario: Own only your sessions
- **WHEN** a user requests a session they do not own
- **THEN** the system responds as if it does not exist

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

### Requirement: Commit a draft to an entity

The system SHALL create the entity from an approved draft. Committing a
one-shot draft SHALL also create its acts, scenes, NPCs, locations, encounters
and clues. The created entity's `user_id` SHALL come from the session, and a
campaign-scoped entity SHALL only be created for a campaign the caller can
access.

#### Scenario: Create a one-shot from a draft
- **WHEN** the DM approves and commits a one-shot draft
- **THEN** the adventure and its acts, scenes, NPCs, locations, encounters and clues are created

#### Scenario: Commit before ready is rejected
- **WHEN** the DM tries to commit while the assistant is still asking questions
- **THEN** the system rejects the request and no entity is created

#### Scenario: Parent-scoped types require a parent
- **WHEN** a quest or item draft is committed without a parent id
- **THEN** the system rejects the request with a clear error

### Requirement: AI drafting is owner-scoped

Draft sessions SHALL be visible and mutable only by their owner (administrators
may view any session). Deleting or discarding a session SHALL remove it.

#### Scenario: Foreign sessions are invisible
- **WHEN** a user requests, mutates or deletes a drafting session owned by someone else
- **THEN** the system responds as if the session does not exist

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
