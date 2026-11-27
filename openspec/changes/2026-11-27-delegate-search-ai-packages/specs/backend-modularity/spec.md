## ADDED Requirements

### Requirement: Handler delegation to the extracted ai and search packages

The system SHALL have exactly one implementation of the AI provider transport and of the campaign/compendium retrieval pipelines, located in packages `ai` and `search`. Handlers SHALL delegate to those packages instead of carrying private copies, and the delegating wrappers SHALL preserve the existing HTTP contracts (paths, query parameters, status codes and JSON shapes).

#### Scenario: AI generation handler delegates to the ai package
- **WHEN** a caller requests text generation (`POST /api/ai/text`) or the drafting assistant generates a reply
- **THEN** the provider request is built and executed by `ai.GenerateText` / `ai.GenerateChat`, and the handler only maps the response (or the `ai.AIGenError` status/message) onto the HTTP response exactly as before

#### Scenario: Copilot context retrieval delegates to the search package
- **WHEN** the copilot chat or transcript-summarise handler needs campaign context
- **THEN** the sources come from `search.RetrieveCampaignContext` with the request context, and the handler response keeps its existing `entity_type`, `entity_id`, `title`, `url`, `snippet` fields

#### Scenario: Compendium search endpoint delegates and keeps its contract
- **WHEN** a client calls `GET /api/compendium/search` with `q` and optional `type`/`class`/`level`/`school`/`category`/`cr`/`monster_type` filters
- **THEN** the results are produced by `search.SearchCompendium`, an empty `q` still yields 400, and each hit reports the correct `type`, `id`, `name` and type-specific fields (including schema-based entries, whose `type` is the schema type name)

#### Scenario: AI enablement read delegates to the ai package
- **WHEN** any handler checks whether AI features are enabled
- **THEN** the check is `ai.IsEnabled` against the `ai_enabled` app setting, with a missing or unreadable setting still meaning enabled
