## ADDED Requirements

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
while chatting).

#### Scenario: Questions come first
- **WHEN** the DM gives a short opening prompt
- **THEN** the assistant asks clarifying questions and does not emit a final draft

#### Scenario: Approval produces a draft
- **WHEN** the DM approves the plan
- **THEN** the assistant returns status "ready" with a complete structured draft

#### Scenario: Non-JSON replies degrade gracefully
- **WHEN** the model returns prose instead of the expected JSON
- **THEN** the reply is shown as a chat message and no draft is offered

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
