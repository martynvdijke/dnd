## ADDED Requirements

### Requirement: Campaign-scoped retrieval
The copilot SHALL assemble context only from entities the requesting user can already see within the given campaign, using the existing search index and permission rules. Unshared knowledge entries SHALL NOT be retrievable by non-owner members.

#### Scenario: Member retrieval respects visibility
- **WHEN** a non-owner member asks a question in a campaign
- **THEN** only entities visible to that member (including only `shared` knowledge) are used as context

#### Scenario: Owner sees full context
- **WHEN** the campaign owner asks a question
- **THEN** the full campaign content is available as context

### Requirement: Cited answers
The copilot SHALL answer using the retrieved campaign context and SHALL return the source entities used, so each answer is verifiable. When the context does not contain an answer, the copilot SHALL say so rather than inventing one.

#### Scenario: Answer with sources
- **WHEN** a user asks who owes the party a favour and such an NPC exists
- **THEN** the answer names the NPC and returns that NPC as a cited source

#### Scenario: Insufficient context
- **WHEN** the campaign contains no relevant content
- **THEN** the copilot reports that it could not find an answer and returns no fabricated source

### Requirement: Conversation persistence
The system SHALL persist copilot conversations and messages scoped to the campaign and their owner, so a conversation can be resumed.

#### Scenario: Resume a conversation
- **WHEN** a user reopens a saved conversation
- **THEN** prior messages are returned in order

#### Scenario: Conversations are isolated
- **WHEN** another campaign member requests a conversation they do not own
- **THEN** access is denied

### Requirement: Prep assistant
The copilot SHALL generate session preparation or a recap from the campaign's own content via the retrieval path.

#### Scenario: Generate prep
- **WHEN** the DM asks for prep for the next session
- **THEN** a prep summary is returned citing the campaign entities it draws on

### Requirement: Transcript ingestion
The system SHALL accept a session transcript, store it as searchable campaign content, and allow it to be summarized into a recap.

#### Scenario: Transcript becomes searchable
- **WHEN** a user submits a session transcript
- **THEN** the content is stored, indexed by the search triggers, and later retrievable

#### Scenario: Transcript summarized
- **WHEN** a user asks for a recap of an ingested transcript
- **THEN** a summary is returned using the AI provider path

### Requirement: AI availability
When AI generation is disabled or no text endpoint is configured, the copilot SHALL return a clear error and SHALL NOT write partial conversation state.

#### Scenario: AI disabled
- **WHEN** AI is disabled and a copilot request is made
- **THEN** the request fails with a clear error and no message is persisted
